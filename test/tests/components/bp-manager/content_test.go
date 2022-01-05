// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package bpmanager

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	wsapi "github.com/bhojpur/platform/bp-manager/api"
	content_service_api "github.com/bhojpur/platform/content-service/api"
	agent "github.com/bhojpur/platform/test/pkg/agent/application/api"
	"github.com/bhojpur/platform/test/pkg/integration"
)

// TestBackup tests a basic start/modify/restart cycle
func TestBackup(t *testing.T) {
	f := features.New("backup").
		Assess("it should start an application, create a file and successfully create a backup", func(_ context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			t.Cleanup(func() {
				api.Done(t)
			})

			wsm, err := api.ApplicationManager()
			if err != nil {
				t.Fatal(err)
			}

			ws, err := integration.LaunchApplicationDirectly(ctx, api)
			if err != nil {
				t.Fatal(err)
			}

			rsa, closer, err := integration.Instrument(integration.ComponentApplication, "application", cfg.Namespace(), cfg.Client(),
				integration.WithInstanceID(ws.Req.Id),
				integration.WithContainer("application"),
				integration.WithApplicationkitLift(true),
			)
			if err != nil {
				t.Fatal(err)
			}
			integration.DeferCloser(t, closer)

			var resp agent.WriteFileResponse
			err = rsa.Call("ApplicationAgent.WriteFile", &agent.WriteFileRequest{
				Path:    "/application/foobar.txt",
				Content: []byte("hello world"),
				Mode:    0644,
			}, &resp)
			if err != nil {
				_, _ = wsm.StopApplication(ctx, &wsapi.StopApplicationRequest{Id: ws.Req.Id})
				t.Fatal(err)
			}
			rsa.Close()

			_, err = wsm.StopApplication(ctx, &wsapi.StopApplicationRequest{
				Id: ws.Req.Id,
			})
			if err != nil {
				t.Fatal(err)
			}

			_, err = integration.WaitForApplicationStop(ctx, api, ws.Req.Id)
			if err != nil {
				t.Fatal(err)
			}

			ws, err = integration.LaunchApplicationDirectly(ctx, api,
				integration.WithRequestModifier(func(w *wsapi.StartApplicationRequest) error {
					w.ServicePrefix = ws.Req.ServicePrefix
					w.Metadata.MetaId = ws.Req.Metadata.MetaId
					w.Metadata.Owner = ws.Req.Metadata.Owner
					return nil
				}),
			)
			if err != nil {
				t.Fatal(err)
			}

			rsa, closer, err = integration.Instrument(integration.ComponentApplication, "application", cfg.Namespace(), cfg.Client(),
				integration.WithInstanceID(ws.Req.Id),
			)
			if err != nil {
				t.Fatal(err)
			}
			integration.DeferCloser(t, closer)

			defer func() {
				t.Log("Cleaning up on TestBackup exit")
				sctx, scancel := context.WithTimeout(ctx, 5*time.Second)
				defer scancel()
				_, _ = wsm.StopApplication(sctx, &wsapi.StopApplicationRequest{
					Id: ws.Req.Id,
				})
			}()

			var ls agent.ListDirResponse
			err = rsa.Call("ApplicationAgent.ListDir", &agent.ListDirRequest{
				Dir: "/application",
			}, &ls)
			if err != nil {
				t.Fatal(err)
			}

			rsa.Close()

			var found bool
			for _, f := range ls.Files {
				if filepath.Base(f) == "foobar.txt" {
					found = true
					break
				}
			}
			if !found {
				t.Fatal("did not find foobar.txt from previous application instance")
			}

			return ctx
		}).
		Feature()

	testEnv.Test(t, f)

}

// TestMissingBackup ensures applications fail if they should have a backup but don't have one
func TestMissingBackup(t *testing.T) {
	f := features.New("CreateApplication").
		WithLabel("component", "server").
		Assess("it can run application tasks", func(_ context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			t.Cleanup(func() {
				api.Done(t)
			})

			ws, err := integration.LaunchApplicationDirectly(ctx, api)
			if err != nil {
				t.Fatal(err)
			}

			wsm, err := api.ApplicationManager()
			if err != nil {
				t.Fatal(err)
			}

			_, err = wsm.StopApplication(ctx, &wsapi.StopApplicationRequest{Id: ws.Req.Id})
			if err != nil {
				t.Fatal(err)
			}

			_, err = integration.WaitForApplicationStop(ctx, api, ws.Req.Id)
			if err != nil {
				t.Fatal(err)
			}

			contentSvc, err := api.ContentService()
			if err != nil {
				t.Fatal(err)
			}

			_, err = contentSvc.DeleteApplication(ctx, &content_service_api.DeleteApplicationRequest{
				OwnerId:       ws.Req.Metadata.Owner,
				ApplicationId: ws.Req.Metadata.MetaId,
			})
			if err != nil {
				t.Fatal(err)
			}

			tests := []struct {
				Name string
				FF   []wsapi.ApplicationFeatureFlag
			}{
				{Name: "classic"},
				{Name: "fwb", FF: []wsapi.ApplicationFeatureFlag{wsapi.ApplicationFeatureFlag_FULL_APPLICATION_BACKUP}},
			}
			for _, test := range tests {
				t.Run(test.Name+"_backup_init", func(t *testing.T) {
					testws, err := integration.LaunchApplicationDirectly(ctx, api, integration.WithRequestModifier(func(w *wsapi.StartApplicationRequest) error {
						w.ServicePrefix = ws.Req.ServicePrefix
						w.Metadata.MetaId = ws.Req.Metadata.MetaId
						w.Metadata.Owner = ws.Req.Metadata.Owner
						w.Spec.Initializer = &content_service_api.ApplicationInitializer{
							Spec: &content_service_api.ApplicationInitializer_Backup{
								Backup: &content_service_api.FromBackupInitializer{},
							},
						}
						w.Spec.FeatureFlags = test.FF
						return nil
					}), integration.WithWaitApplicationForOpts(integration.ApplicationCanFail))
					if err != nil {
						t.Fatal(err)
					}

					if testws.LastStatus == nil {
						t.Fatal("did not receive a last status")
						return
					}
					if testws.LastStatus.Conditions.Failed == "" {
						t.Error("restarted application did not fail despite missing backup")
					}
				})
			}
			return ctx
		}).
		Feature()

	testEnv.Test(t, f)
}
