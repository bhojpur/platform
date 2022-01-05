// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package bpmanager

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	gitpod "github.com/bhojpur/platform/bhojpur-protocol"
	wsmanapi "github.com/bhojpur/platform/bp-manager/api"
	supervisor "github.com/bhojpur/platform/supervisor/api"
	agent "github.com/bhojpur/platform/test/pkg/agent/application/api"
	"github.com/bhojpur/platform/test/pkg/integration"
)

func TestRegularApplicationTasks(t *testing.T) {
	tests := []struct {
		Name        string
		Task        gitpod.TasksItems
		LookForFile string
	}{
		{
			Name:        "init",
			Task:        gitpod.TasksItems{Init: "touch /application/init-ran; exit"},
			LookForFile: "init-ran",
		},
		{
			Name:        "before",
			Task:        gitpod.TasksItems{Before: "touch /application/before-ran; exit"},
			LookForFile: "before-ran",
		},
		{
			Name:        "command",
			Task:        gitpod.TasksItems{Command: "touch /application/command-ran; exit"},
			LookForFile: "command-ran",
		},
	}

	f := features.New("bp-manager").
		WithLabel("component", "bp-manager").
		WithLabel("type", "tasks").
		Assess("it can run application tasks", func(_ context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			t.Cleanup(func() {
				api.Done(t)
			})

			for _, test := range tests {
				t.Run(test.Name, func(t *testing.T) {
					// t.Parallel()

					addInitTask := func(swr *wsmanapi.StartApplicationRequest) error {
						tasks, err := json.Marshal([]gitpod.TasksItems{test.Task})
						if err != nil {
							return err
						}
						swr.Spec.Envvars = append(swr.Spec.Envvars, &wsmanapi.EnvironmentVariable{
							Name:  "BHOJPUR_TASKS",
							Value: string(tasks),
						})
						return nil
					}

					nfo, err := integration.LaunchApplicationDirectly(ctx, api, integration.WithRequestModifier(addInitTask))
					if err != nil {
						t.Fatal(err)
					}

					t.Cleanup(func() {
						_ = integration.DeleteApplication(ctx, api, nfo.Req.Id)
					})

					conn, err := api.Supervisor(nfo.Req.Id)
					if err != nil {
						t.Fatal(err)
					}

					tsctx, tscancel := context.WithTimeout(ctx, 60*time.Second)
					defer tscancel()

					statusService := supervisor.NewStatusServiceClient(conn)
					resp, err := statusService.TasksStatus(tsctx, &supervisor.TasksStatusRequest{Observe: false})
					if err != nil {
						t.Fatal(err)
					}

					for {
						status, err := resp.Recv()
						if errors.Is(err, io.EOF) {
							break
						}

						if err != nil {
							t.Fatal(err)
						}
						if len(status.Tasks) != 1 {
							t.Fatalf("expected one task to run, but got %d", len(status.Tasks))
						}
						if status.Tasks[0].State == supervisor.TaskState_closed {
							break
						}
					}

					rsa, closer, err := integration.Instrument(integration.ComponentApplication, "application", cfg.Namespace(), cfg.Client(), integration.WithInstanceID(nfo.Req.Id))
					if err != nil {
						t.Fatalf("unexpected error instrumenting application: %v", err)
					}
					defer rsa.Close()
					integration.DeferCloser(t, closer)

					var ls agent.ListDirResponse
					err = rsa.Call("ApplicationAgent.ListDir", &agent.ListDirRequest{
						Dir: "/application",
					}, &ls)
					if err != nil {
						t.Fatal(err)
					}

					var foundMaker bool
					for _, f := range ls.Files {
						t.Logf("file in application: %s", f)
						if f == test.LookForFile {
							foundMaker = true
							break
						}
					}
					if !foundMaker {
						t.Fatal("task seems to have run, but cannot find the file it should have created")
					}
				})
			}

			return ctx
		}).
		Feature()

	testEnv.Test(t, f)
}
