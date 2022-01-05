// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package integration

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/xerrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	protocol "github.com/bhojpur/platform/bhojpur-protocol"
	wsmanapi "github.com/bhojpur/platform/bp-manager/api"
	"github.com/bhojpur/platform/common-go/namegen"
	csapi "github.com/bhojpur/platform/content-service/api"
	imgbldr "github.com/bhojpur/platform/image-builder/api"
)

const (
	bhojpurBuiltinUserID = "builtin-user-application-probe-0000000"
	perCallTimeout       = 1 * time.Minute
)

type launchApplicationDirectlyOptions struct {
	BaseImage   string
	IdeImage    string
	Mods        []func(*wsmanapi.StartApplicationRequest) error
	WaitForOpts []WaitForApplicationOpt
}

// LaunchApplicationDirectlyOpt configures the behaviour of LaunchApplicationDirectly
type LaunchApplicationDirectlyOpt func(*launchApplicationDirectlyOptions) error

// WithoutApplicationImage prevents the image-builder based base image resolution and sets
// the application image to an empty string.
// Usually callers would then use WithRequestModifier to set the application image themselves.
func WithoutApplicationImage() LaunchApplicationDirectlyOpt {
	return func(lwdo *launchApplicationDirectlyOptions) error {
		lwdo.BaseImage = ""
		return nil
	}
}

// WithBaseImage configures the base image used to start the application. The base image
// will be resolved to a application image using the image builder. If the corresponding
// Application image isn't built yet, it will NOT be built.
func WithBaseImage(baseImage string) LaunchApplicationDirectlyOpt {
	return func(lwdo *launchApplicationDirectlyOptions) error {
		lwdo.BaseImage = baseImage
		return nil
	}
}

// WithIDEImage configures the IDE image used to start the application. Using this option
// as compared to setting the image using a modifier prevents the image ref computation
// based on the server's configuration.
func WithIDEImage(ideImage string) LaunchApplicationDirectlyOpt {
	return func(lwdo *launchApplicationDirectlyOptions) error {
		lwdo.IdeImage = ideImage
		return nil
	}
}

// WithRequestModifier modifies the start application request before it's sent.
func WithRequestModifier(mod func(*wsmanapi.StartApplicationRequest) error) LaunchApplicationDirectlyOpt {
	return func(lwdo *launchApplicationDirectlyOptions) error {
		lwdo.Mods = append(lwdo.Mods, mod)
		return nil
	}
}

// WithWaitApplicationForOpts adds options to the WaitForApplication call that happens as part of LaunchApplicationDirectly
func WithWaitApplicationForOpts(opt ...WaitForApplicationOpt) LaunchApplicationDirectlyOpt {
	return func(lwdo *launchApplicationDirectlyOptions) error {
		lwdo.WaitForOpts = opt
		return nil
	}
}

// LaunchApplicationDirectlyResult is returned by LaunchApplicationDirectly
type LaunchApplicationDirectlyResult struct {
	Req        *wsmanapi.StartApplicationRequest
	IdeURL     string
	LastStatus *wsmanapi.ApplicationStatus
}

// LaunchApplicationDirectly starts an application pod by talking directly to bp-manager.
// Whenever possible prefer this function over LaunchApplicationFromContextURL, because
// it has fewer prerequisites.
func LaunchApplicationDirectly(ctx context.Context, api *ComponentAPI, opts ...LaunchApplicationDirectlyOpt) (*LaunchApplicationDirectlyResult, error) {
	options := launchApplicationDirectlyOptions{
		BaseImage: "docker.io/bhojpur/platform-full:latest",
	}
	for _, o := range opts {
		err := o(&options)
		if err != nil {
			return nil, err
		}
	}

	instanceID, err := uuid.NewRandom()
	if err != nil {
		return nil, err

	}
	applicationID, err := namegen.GenerateApplicationID()
	if err != nil {
		return nil, err

	}

	var applicationImage string
	if options.BaseImage != "" {
		applicationImage, err = resolveOrBuildImage(ctx, api, options.BaseImage)
		if err != nil {
			return nil, xerrors.Errorf("cannot resolve base image: %v", err)
		}
	}
	if applicationImage == "" {
		return nil, xerrors.Errorf("cannot start applications without an application image (required by registry-facade resolver)")
	}

	ideImage := options.IdeImage
	if ideImage == "" {
		cfg, err := GetServerIDEConfig(api.namespace, api.client)
		if err != nil {
			return nil, xerrors.Errorf("cannot find server IDE config: %q", err)
		}
		ideImage = cfg.IDEImageAliases.Code
		if ideImage == "" {
			return nil, xerrors.Errorf("cannot start applications without an IDE image (required by registry-facade resolver)")
		}
	}

	req := &wsmanapi.StartApplicationRequest{
		Id:            instanceID.String(),
		ServicePrefix: instanceID.String(),
		Metadata: &wsmanapi.ApplicationMetadata{
			Owner:  bhojpurBuiltinUserID,
			MetaId: applicationID,
		},
		Type: wsmanapi.ApplicationType_REGULAR,
		Spec: &wsmanapi.StartApplicationSpec{
			ApplicationImage:   applicationImage,
			DeprecatedIdeImage: ideImage,
			IdeImage: &wsmanapi.IDEImage{
				WebRef: ideImage,
			},
			CheckoutLocation:    "/",
			ApplicationLocation: "/",
			Timeout:             "30m",
			Initializer: &csapi.ApplicationInitializer{
				Spec: &csapi.ApplicationInitializer_Empty{
					Empty: &csapi.EmptyInitializer{},
				},
			},
			Git: &wsmanapi.GitSpec{
				Username: "integration-test",
				Email:    "integration-test@bhojpur.net",
			},
			Admission: wsmanapi.AdmissionLevel_ADMIT_OWNER_ONLY,
		},
	}
	for _, m := range options.Mods {
		err := m(req)
		if err != nil {
			return nil, err
		}
	}

	sctx, scancel := context.WithTimeout(ctx, perCallTimeout)
	defer scancel()

	wsm, err := api.ApplicationManager()
	if err != nil {
		return nil, xerrors.Errorf("cannot start application: %q", err)
	}

	sresp, err := wsm.StartApplication(sctx, req)
	scancel()
	if err != nil {
		return nil, xerrors.Errorf("cannot start application: %q", err)
	}

	lastStatus, err := WaitForApplicationStart(ctx, instanceID.String(), api, options.WaitForOpts...)
	if err != nil {
		return nil, xerrors.Errorf("cannot start application: %q", err)
	}

	// it.t.Logf("application is running: instanceID=%s", instanceID.String())

	return &LaunchApplicationDirectlyResult{
		Req:        req,
		IdeURL:     sresp.Url,
		LastStatus: lastStatus,
	}, nil
}

// LaunchApplicationFromContextURL force-creates a new application using the Bhojpur.NET Platform server API,
// and waits for the application to start. If any step along the way fails, this function will
// fail the test.
//
// When possible, prefer the less complex LaunchApplicationDirectly.
func LaunchApplicationFromContextURL(ctx context.Context, contextURL string, username string, api *ComponentAPI, serverOpts ...BhojpurServerOpt) (*protocol.ApplicationInfo, func(waitForStop bool), error) {
	var defaultServerOpts []BhojpurServerOpt
	if username != "" {
		defaultServerOpts = []BhojpurServerOpt{WithBhojpurUser(username)}
	}

	server, err := api.BhojpurServer(append(defaultServerOpts, serverOpts...)...)
	if err != nil {
		return nil, nil, xerrors.Errorf("cannot start server: %q", err)
	}

	cctx, ccancel := context.WithTimeout(context.Background(), perCallTimeout)
	defer ccancel()
	resp, err := server.CreateApplication(cctx, &protocol.CreateApplicationOptions{
		ContextURL: contextURL,
		Mode:       "force-new",
	})
	if err != nil {
		return nil, nil, xerrors.Errorf("cannot start application: %q", err)
	}

	nfo, err := server.GetApplication(ctx, resp.CreatedApplicationID)
	if err != nil {
		return nil, nil, xerrors.Errorf("cannot get application: %q", err)
	}
	if nfo.LatestInstance == nil {
		return nil, nil, xerrors.Errorf("CreateApplication did not start the application")
	}

	// GetApplication might receive an instance before we seen the first event
	// from ws-manager, in which case IdeURL is not set
	nfo.LatestInstance.IdeURL = resp.ApplicationURL

	stopWs := func(waitForStop bool) {
		sctx, scancel := context.WithTimeout(ctx, perCallTimeout)
		_ = server.StopApplication(sctx, resp.CreatedApplicationID)
		scancel()
		//if err != nil {
		//it.t.Errorf("cannot stop application: %q", err)
		//}

		if waitForStop {
			_, _ = WaitForApplicationStop(ctx, api, nfo.LatestInstance.ID)
		}
	}
	defer func() {
		if err != nil {
			stopWs(false)
		}
	}()
	// it.t.Logf("created application: applicationID=%s url=%s", resp.CreatedApplicationID, resp.ApplicationURL)

	_, err = WaitForApplicationStart(ctx, nfo.LatestInstance.ID, api)
	if err != nil {
		return nil, nil, xerrors.Errorf("cannot start application: %q", err)
	}

	// it.t.Logf("Application is running: instanceID=%s", nfo.LatestInstance.ID)

	return nfo, stopWs, nil
}

// WaitForApplicationOpt configures a WaitForApplication call
type WaitForApplicationOpt func(*waitForApplicationOpts)

type waitForApplicationOpts struct {
	CanFail bool
}

// ApplicationCanFail doesn't fail the test if the application fails to start
func ApplicationCanFail(o *waitForApplicationOpts) {
	o.CanFail = true
}

// WaitForApplication waits until an application is running. Fails the test if the application
// fails or does not become RUNNING before the context is canceled.
func WaitForApplicationStart(ctx context.Context, instanceID string, api *ComponentAPI, opts ...WaitForApplicationOpt) (lastStatus *wsmanapi.ApplicationStatus, err error) {
	var cfg waitForApplicationOpts
	for _, o := range opts {
		o(&cfg)
	}

	wsman, err := api.ApplicationManager()
	if err != nil {
		return nil, err
	}

	var sub wsmanapi.ApplicationManager_SubscribeClient
	for i := 0; i < 5; i++ {
		sub, err = wsman.Subscribe(ctx, &wsmanapi.SubscribeRequest{})
		if status.Code(err) == codes.NotFound {
			time.Sleep(1 * time.Second)
			continue
		}
		if err != nil {
			return nil, xerrors.Errorf("cannot listen for application updates: %q", err)
		}
		defer func() {
			_ = sub.CloseSend()
		}()
		break
	}

	done := make(chan *wsmanapi.ApplicationStatus)
	errStatus := make(chan error)

	go func() {
		var status *wsmanapi.ApplicationStatus
		defer func() {
			done <- status
			close(done)
		}()
		for {
			resp, err := sub.Recv()
			if err != nil {
				errStatus <- xerrors.Errorf("application update error: %q", err)
				return
			}
			status = resp.GetStatus()
			if status == nil {
				continue
			}
			if status.Id != instanceID {
				continue
			}

			if cfg.CanFail {
				if status.Phase == wsmanapi.ApplicationPhase_STOPPING {
					return
				}
				if status.Phase == wsmanapi.ApplicationPhase_STOPPED {
					return
				}
			} else {
				if status.Conditions.Failed != "" {
					errStatus <- xerrors.Errorf("application instance %s failed: %s", instanceID, status.Conditions.Failed)
					return
				}
				if status.Phase == wsmanapi.ApplicationPhase_STOPPING {
					errStatus <- xerrors.Errorf("application instance %s is stopping", instanceID)
					return
				}
				if status.Phase == wsmanapi.ApplicationPhase_STOPPED {
					errStatus <- xerrors.Errorf("application instance %s has stopped", instanceID)
					return
				}
			}
			if status.Phase != wsmanapi.ApplicationPhase_RUNNING {
				// we're still starting
				continue
			}

			// all is well, the application is running
			return
		}
	}()

	// maybe the application has started in the meantime and we've missed the update
	desc, _ := wsman.DescribeApplication(ctx, &wsmanapi.DescribeApplicationRequest{Id: instanceID})
	if desc != nil {
		switch desc.Status.Phase {
		case wsmanapi.ApplicationPhase_RUNNING:
			return
		case wsmanapi.ApplicationPhase_STOPPING:
			if !cfg.CanFail {
				return nil, xerrors.Errorf("application instance %s is stopping", instanceID)
			}
		case wsmanapi.ApplicationPhase_STOPPED:
			if !cfg.CanFail {
				return nil, xerrors.Errorf("application instance %s has stopped", instanceID)
			}
		}
	}

	select {
	case <-ctx.Done():
		return nil, xerrors.Errorf("cannot wait for application: %q", ctx.Err())
	case s := <-done:
		return s, nil
	case err := <-errStatus:
		return nil, err
	}
}

// WaitForApplicationStop waits until an application is stopped. Fails the test if the application
// fails or does not stop before the context is canceled.
func WaitForApplicationStop(ctx context.Context, api *ComponentAPI, instanceID string) (lastStatus *wsmanapi.ApplicationStatus, err error) {
	wsman, err := api.ApplicationManager()
	if err != nil {
		return nil, xerrors.Errorf("cannot listen for application updates: %q", err)
	}

	sub, err := wsman.Subscribe(context.Background(), &wsmanapi.SubscribeRequest{})
	if err != nil {
		return nil, xerrors.Errorf("cannot listen for application updates: %q", err)
	}
	defer func() {
		_ = sub.CloseSend()
	}()

	var applicationID string
	done := make(chan struct{})
	errCh := make(chan error)
	go func() {
		defer close(done)
		for {
			resp, err := sub.Recv()
			if err != nil {
				errCh <- xerrors.Errorf("application update error: %q", err)
				return
			}
			status := resp.GetStatus()
			if status == nil {
				continue
			}
			if status.Id != instanceID {
				continue
			}

			applicationID = status.Metadata.MetaId
			if status.Conditions.Failed != "" {
				errCh <- xerrors.Errorf("application instance %s failed: %s", instanceID, status.Conditions.Failed)
				return
			}
			if status.Phase == wsmanapi.ApplicationPhase_STOPPED {
				lastStatus = status
				return
			}
		}
	}()

	// maybe the application has stopped in the meantime and we've missed the update
	desc, _ := wsman.DescribeApplication(context.Background(), &wsmanapi.DescribeApplicationRequest{Id: instanceID})
	if desc != nil {
		switch desc.Status.Phase {
		case wsmanapi.ApplicationPhase_STOPPED:
			// ensure theia service is cleaned up
			lastStatus = desc.Status
		}
	}

	select {
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, xerrors.Errorf("cannot wait for application: %q", ctx.Err())
	case <-done:
	}

	// wait for the Theia service to be properly deleted
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var (
		start       = time.Now()
		serviceGone bool
	)

	// NOTE: this needs to be kept in sync with components/ws-manager/pkg/manager/manager.go:getTheiaServiceName()
	// TODO(rl) expose it?
	theiaName := fmt.Sprintf("ws-%s-theia", strings.TrimSpace(strings.ToLower(applicationID)))
	for time.Since(start) < 1*time.Minute {
		var svc corev1.Service
		err := api.client.Resources().Get(ctx, fmt.Sprintf("ws-%s-theia", applicationID), api.namespace, &svc)
		if errors.IsNotFound(err) {
			serviceGone = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !serviceGone {
		return nil, xerrors.Errorf("application service did not disappear in time (theia)")
	}
	// Wait for the theia endpoints to be properly deleted (i.e. syncing)
	var endpointGone bool
	for time.Since(start) < 1*time.Minute {
		var svc corev1.Endpoints
		err := api.client.Resources().Get(ctx, theiaName, api.namespace, &svc)
		if errors.IsNotFound(err) {
			endpointGone = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !endpointGone {
		return nil, xerrors.Errorf("Theia endpoint:%s did not disappear in time", theiaName)
	}

	return
}

// WaitForApplication waits until the condition function returns true. Fails the test if the condition does
// not become true before the context is canceled.
func WaitForApplication(ctx context.Context, api *ComponentAPI, instanceID string, condition func(status *wsmanapi.ApplicationStatus) bool) (lastStatus *wsmanapi.ApplicationStatus, err error) {
	wsman, err := api.ApplicationManager()
	if err != nil {
		return
	}

	sub, err := wsman.Subscribe(ctx, &wsmanapi.SubscribeRequest{})
	if err != nil {
		return nil, xerrors.Errorf("cannot listen for application updates: %q", err)
	}

	done := make(chan *wsmanapi.ApplicationStatus, 1)
	errCh := make(chan error)

	var once sync.Once
	go func() {
		var status *wsmanapi.ApplicationStatus
		defer func() {
			once.Do(func() {
				done <- status
				close(done)
			})
			_ = sub.CloseSend()
		}()
		for {
			resp, err := sub.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				errCh <- xerrors.Errorf("application update error: %q", err)
				return
			}
			status = resp.GetStatus()
			if status == nil {
				continue
			}
			if status.Id != instanceID {
				continue
			}

			if condition(status) {
				return
			}
		}
	}()

	// maybe the application has started in the meantime and we've missed the update
	desc, err := wsman.DescribeApplication(ctx, &wsmanapi.DescribeApplicationRequest{Id: instanceID})
	if err != nil {
		return nil, xerrors.Errorf("cannot get application: %q", err)
	}
	if condition(desc.Status) {
		once.Do(func() { close(done) })
		return desc.Status, nil
	}

	select {
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, xerrors.Errorf("cannot wait for application: %q", ctx.Err())
	case s := <-done:
		return s, nil
	}
}

func resolveOrBuildImage(ctx context.Context, api *ComponentAPI, baseRef string) (absref string, err error) {
	cl, err := api.ImageBuilder()
	if err != nil {
		return
	}

	reslv, err := cl.ResolveApplicationImage(ctx, &imgbldr.ResolveApplicationImageRequest{
		Source: &imgbldr.BuildSource{
			From: &imgbldr.BuildSource_Ref{
				Ref: &imgbldr.BuildSourceReference{
					Ref: baseRef,
				},
			},
		},
		Auth: &imgbldr.BuildRegistryAuth{
			Mode: &imgbldr.BuildRegistryAuth_Total{
				Total: &imgbldr.BuildRegistryAuthTotal{
					AllowAll: true,
				},
			},
		},
	})
	if err != nil {
		return
	}

	if reslv.Status == imgbldr.BuildStatus_done_success {
		return reslv.Ref, nil
	}

	bld, err := cl.Build(ctx, &imgbldr.BuildRequest{
		Source: &imgbldr.BuildSource{
			From: &imgbldr.BuildSource_Ref{
				Ref: &imgbldr.BuildSourceReference{
					Ref: baseRef,
				},
			},
		},
		Auth: &imgbldr.BuildRegistryAuth{
			Mode: &imgbldr.BuildRegistryAuth_Total{
				Total: &imgbldr.BuildRegistryAuthTotal{
					AllowAll: true,
				},
			},
		},
	})
	if err != nil {
		return
	}

	for {
		resp, err := bld.Recv()
		if err != nil {
			return "", err
		}

		if resp.Status == imgbldr.BuildStatus_done_success {
			break
		} else if resp.Status == imgbldr.BuildStatus_done_failure {
			return "", xerrors.Errorf("cannot build application image: %s", resp.Message)
		}
	}

	return reslv.Ref, nil
}

// DeleteApplication cleans up application started during an integration test
func DeleteApplication(ctx context.Context, api *ComponentAPI, instanceID string) error {
	wm, err := api.ApplicationManager()
	if err != nil {
		return err
	}

	_, err = wm.StopApplication(ctx, &wsmanapi.StopApplicationRequest{
		Id: instanceID,
	})
	if err != nil {
		return err
	}

	if err == nil {
		return nil
	}

	s, ok := status.FromError(err)
	if ok && s.Code() == codes.NotFound {
		return nil
	}

	return err
}
