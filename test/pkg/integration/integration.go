// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	kubectlcp "k8s.io/kubectl/pkg/cmd/cp"
	kubectlexec "k8s.io/kubectl/pkg/cmd/exec"
	"sigs.k8s.io/e2e-framework/klient"

	"github.com/bhojpur/platform/test/pkg/integration/common"
)

type PodExec struct {
	RestConfig *rest.Config
	*kubernetes.Clientset
}

func NewPodExec(config rest.Config, clientset *kubernetes.Clientset) *PodExec {
	config.APIPath = "/api"                                   // Make sure we target /api and not just /
	config.GroupVersion = &schema.GroupVersion{Version: "v1"} // this targets the core api groups so the url path will be /api/v1
	config.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs}
	return &PodExec{
		RestConfig: &config,
		Clientset:  clientset,
	}
}

func (p *PodExec) PodCopyFile(src string, dst string, containername string) (*bytes.Buffer, *bytes.Buffer, *bytes.Buffer, error) {
	ioStreams, in, out, errOut := genericclioptions.NewTestIOStreams()
	copyOptions := kubectlcp.NewCopyOptions(ioStreams)
	copyOptions.Clientset = p.Clientset
	copyOptions.ClientConfig = p.RestConfig
	copyOptions.Container = containername
	err := copyOptions.Run([]string{src, dst})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Could not run copy operation: %v", err)
	}
	return in, out, errOut, nil
}

func (p *PodExec) ExecCmd(command []string, podname string, namespace string, containername string) (*bytes.Buffer, *bytes.Buffer, *bytes.Buffer, error) {
	ioStreams, in, out, errOut := genericclioptions.NewTestIOStreams()
	execOptions := &kubectlexec.ExecOptions{
		StreamOptions: kubectlexec.StreamOptions{
			IOStreams:     ioStreams,
			Namespace:     namespace,
			PodName:       podname,
			ContainerName: containername,
		},

		Command:   command,
		Executor:  &kubectlexec.DefaultRemoteExecutor{},
		PodClient: p.Clientset.CoreV1(),
		Config:    p.RestConfig,
	}
	err := execOptions.Run()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Could not run exec operation: %v", err)
	}
	return in, out, errOut, nil
}

// InstrumentOption configures an Instrument call
type InstrumentOption func(*instrumentOptions) error

type instrumentOptions struct {
	SPO                selectPodOptions
	ApplicationkitLift bool
}

type selectPodOptions struct {
	InstanceID string

	Container string
}

// WithInstanceID provides a hint during pod selection for Instrument.
// When instrumenting ws-daemon, we try to select the daemon on the node where the application is located.
// When instrumenting the application, we select the application based on the instance ID.
// For all other component types, this hint is ignored.
func WithInstanceID(instanceID string) InstrumentOption {
	return func(io *instrumentOptions) error {
		io.SPO.InstanceID = instanceID
		return nil
	}
}

// Container provides a hint during pod selection for Instrument a particular container
func WithContainer(container string) InstrumentOption {
	return func(io *instrumentOptions) error {
		io.SPO.Container = container
		return nil
	}
}

// WithApplicationkitLift executes the agent using `applicationkit lift` thereby lifting it into ring1.
// Only relevant for ComponentApplication and ignored for all other components.
// Defaults to true.
func WithApplicationkitLift(lift bool) InstrumentOption {
	return func(io *instrumentOptions) error {
		io.ApplicationkitLift = lift
		return nil
	}
}

// Instrument builds and uploads an agent to a pod, then connects to its RPC service.
// We first check if there's an executable in the path named `bhojpur-integration-test-<agentName>-agent`.
// If there isn't, we attempt to build `<agentName>_agent/main.go`.
// The binary is copied to the destination pod, started and port-forwarded. Then we
// create an RPC client.
func Instrument(component ComponentType, agentName string, namespace string, client klient.Client, opts ...InstrumentOption) (*rpc.Client, []func() error, error) {
	var closer []func() error

	options := instrumentOptions{
		ApplicationkitLift: true,
	}
	for _, o := range opts {
		err := o(&options)
		if err != nil {
			return nil, closer, err
		}
	}

	expectedBinaryName := fmt.Sprintf("bhojpur-integration-test-%s-agent", agentName)
	agentLoc, _ := exec.LookPath(expectedBinaryName)
	if agentLoc == "" {
		var err error
		agentLoc, err = buildAgent(agentName)
		if err != nil {
			return nil, closer, err
		}
		defer os.Remove(agentLoc)
	}

	podName, containerName, err := selectPod(component, options.SPO, namespace, client)
	if err != nil {
		return nil, closer, err
	}
	tgtFN := filepath.Base(agentLoc)

	clientConfig, err := kubernetes.NewForConfig(client.RESTConfig())
	if err != nil {
		return nil, closer, err
	}
	podExec := NewPodExec(*client.RESTConfig(), clientConfig)
	_, _, _, err = podExec.PodCopyFile(agentLoc, fmt.Sprintf("%s/%s:/home/bhojpur/%s", namespace, podName, tgtFN), containerName)
	if err != nil {
		return nil, closer, err
	}

	localAgentPort, err := getFreePort()
	if err != nil {
		return nil, closer, err
	}

	cmd := []string{filepath.Join("/home/bhojpur/", tgtFN), "-rpc-port", strconv.Itoa(localAgentPort)}
	if options.ApplicationkitLift {
		cmd = append([]string{"/.supervisor/applicationkit", "lift"}, cmd...)
	}

	execErrs := make(chan error, 1)
	go func() {
		defer close(execErrs)
		execErr := executeAgent(cmd, podName, containerName, namespace, client)
		if execErr != nil {
			execErrs <- execErr
		}
	}()
	select {
	case err := <-execErrs:
		if err != nil {
			return nil, closer, err
		}
		return nil, closer, fmt.Errorf("agent stopped unexepectedly")
	case <-time.After(1 * time.Second):
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		if err == nil {
			closer = append(closer, func() error {
				cancel()
				return nil
			})
		} else {
			cancel()
		}
	}()

	fwdReady, fwdErr := common.ForwardPort(ctx, client.RESTConfig(), namespace, podName, strconv.Itoa(localAgentPort))
	select {
	case <-fwdReady:
	case err := <-execErrs:
		if err != nil {
			return nil, closer, err
		}
	case err := <-fwdErr:
		if err != nil {
			return nil, closer, err
		}
	}

	var res *rpc.Client
	var lastError error
	waitErr := wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		res, lastError = rpc.DialHTTP("tcp", fmt.Sprintf("localhost:%d", localAgentPort))
		if lastError != nil {
			return false, nil
		}

		return true, nil
	})
	if waitErr == wait.ErrWaitTimeout {
		return nil, closer, xerrors.Errorf("timed out attempting to connect agent: %v", lastError)
	}
	if waitErr != nil {
		return nil, closer, err
	}

	closer = append(closer, func() error {
		err := res.Call(MethodTestAgentShutdown, new(TestAgentShutdownRequest), new(TestAgentShutdownResponse))
		if err != nil && strings.Contains(err.Error(), "connection is shut down") {
			return nil
		}

		if err != nil {
			return xerrors.Errorf("cannot shutdown agent: %w", err)
		}
		return nil
	})

	return res, closer, nil
}

func getFreePort() (int, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()

	result, err := net.ResolveTCPAddr("tcp", l.Addr().String())
	if err != nil {
		return 0, err
	}

	return result.Port, nil
}

func executeAgent(cmd []string, pod, container string, namespace string, client klient.Client) error {
	args := []string{"exec", pod, fmt.Sprintf("--namespace=%v", namespace)}
	if len(container) > 0 {
		args = append(args, fmt.Sprintf("--container=%s", container))
	}
	args = append(args, "--")
	args = append(args, cmd...)

	command := exec.Command("kubectl", args...)
	out, err := command.CombinedOutput()
	if err != nil {
		return xerrors.Errorf("cannot run kubectl command: %w\n%v", err, string(out))
	}

	return nil
}

func buildAgent(name string) (loc string, err error) {
	defer func() {
		if err != nil {
			err = xerrors.Errorf("cannot build agent: %w", err)
		}
	}()

	_, filename, _, _ := runtime.Caller(0)
	src := path.Join(path.Dir(filename), "..", "agent", name, "main.go")
	if _, err := os.Stat(src); err != nil {
		return "", err
	}

	f, err := os.CreateTemp("", "bhojpur-integration-test-*")
	if err != nil {
		return "", err
	}
	f.Close()

	cmd := exec.Command("go", "build", "-trimpath", "-ldflags='-buildid= -w -s'", "-o", f.Name(), src)
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=0",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", xerrors.Errorf("%w: %s", err, string(out))
	}

	return f.Name(), nil
}

func selectPod(component ComponentType, options selectPodOptions, namespace string, client klient.Client) (string, string, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: "component=" + string(component),
	}

	if component == ComponentApplication && options.InstanceID != "" {
		listOptions.LabelSelector = "component=application,applicationID=" + options.InstanceID
	}

	if component == ComponentApplicationDaemon && options.InstanceID != "" {
		var pods corev1.PodList
		err := client.Resources(namespace).List(context.Background(), &pods, func(opts *metav1.ListOptions) {
			opts.LabelSelector = "component=application,applicationID=" + options.InstanceID
		})
		if err != nil {
			return "", "", xerrors.Errorf("cannot list pods: %w", err)
		}

		if len(pods.Items) == 0 {
			return "", "", xerrors.Errorf("no application pod for instance %s", options.InstanceID)
		}

		listOptions.FieldSelector = "spec.nodeName=" + pods.Items[0].Spec.NodeName
	}

	var pods corev1.PodList
	err := client.Resources(namespace).List(context.Background(), &pods, func(opts *metav1.ListOptions) {
		opts.LabelSelector = listOptions.LabelSelector
		if listOptions.FieldSelector != "" {
			opts.FieldSelector = listOptions.FieldSelector
		}
	})
	if err != nil {
		return "", "", xerrors.Errorf("cannot list pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return "", "", xerrors.Errorf("no pods for %s", component)
	}

	if len(pods.Items) > 1 {
		//t.t.Logf("found multiple pods for %s, choosing %s", component, pod)
	}

	p := pods.Items[0]

	if !isPodRunningReady(p) {
		return "", "", xerrors.Errorf("pods for component %s is not running", component)
	}

	pod := p.Name

	var container string
	if options.Container != "" {
		var found bool
		for _, container := range pods.Items[0].Spec.Containers {
			if container.Name == options.Container {
				found = true
				break
			}
		}

		if !found {
			return "", "", xerrors.Errorf("no container name %s found", options.Container)
		}

		container = options.Container
	}

	return pod, container, nil
}

// ServerConfigPartial is the subset of server config we're using for integration tests.
// Ideally we're using a definition derived from the config interface, someday...
// NOTE: keep in sync with chart/templates/server-configmap.yaml
type ServerConfigPartial struct {
	HostURL             string `json:"hostUrl"`
	ApplicationDefaults struct {
		ApplicationImage string `json:"applicationImage"`
	} `json:"applicationDefaults"`
	Session struct {
		Secret string `json:"secret"`
	} `json:"session"`
}

func GetServerConfig(namespace string, client klient.Client) (*ServerConfigPartial, error) {
	var cm corev1.ConfigMap
	err := client.Resources().Get(context.Background(), "server-config", namespace, &cm)
	if err != nil {
		return nil, err
	}

	key := "config.json"
	configJson, ok := cm.Data[key]
	if !ok {
		return nil, fmt.Errorf("key %s not found", key)
	}

	var config ServerConfigPartial
	err = json.Unmarshal([]byte(configJson), &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling server config: %v", err)
	}
	return &config, nil
}

// ServerIDEConfigPartial is the subset of server IDE config we're using for integration tests.
// NOTE: keep in sync with chart/templates/server-ide-configmap.yaml
type ServerIDEConfigPartial struct {
	IDEVersion      string `json:"ideVersion"`
	IDEImageRepo    string `json:"ideImageRepo"`
	IDEImageAliases struct {
		Code       string `json:"code"`
		CodeLatest string `json:"code-latest"`
	} `json:"ideImageAliases"`
}

func GetServerIDEConfig(namespace string, client klient.Client) (*ServerIDEConfigPartial, error) {
	var cm corev1.ConfigMap
	err := client.Resources().Get(context.Background(), "server-ide-config", namespace, &cm)
	if err != nil {
		return nil, err
	}

	key := "config.json"
	configJson, ok := cm.Data[key]
	if !ok {
		return nil, fmt.Errorf("key %s not found", key)
	}

	var config ServerIDEConfigPartial
	err = json.Unmarshal([]byte(configJson), &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling server IDE config: %v", err)
	}
	return &config, nil
}

// ComponentType denotes a Bhojpur.NET Platform component
type ComponentType string

const (
	// ComponentApplicationDaemon points to the application daemon
	ComponentApplicationDaemon ComponentType = "bp-daemon"
	// ComponentApplicationManager points to the application manager
	ComponentApplicationManager ComponentType = "bp-manager"
	// ComponentContentService points to the content service
	ComponentContentService ComponentType = "content-service"
	// ComponentApplication points to a application
	ComponentApplication ComponentType = "application"
	// ComponentImageBuilder points to the image-builder
	ComponentImageBuilder ComponentType = "image-builder"
	// ComponentImageBuilder points to the image-builder-mk3
	ComponentImageBuilderMK3 ComponentType = "image-builder-mk3"
)

func isPodRunningReady(p corev1.Pod) bool {
	if p.Status.Phase != corev1.PodRunning {
		return false
	}

	return isPodReady(&p.Status)
}
func isPodReady(s *corev1.PodStatus) bool {
	for i := range s.Conditions {
		if s.Conditions[i].Type == corev1.PodReady {
			return s.Conditions[i].Status == corev1.ConditionTrue
		}
	}

	return false
}
