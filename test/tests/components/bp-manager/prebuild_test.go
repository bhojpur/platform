// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package wsmanager

import (
	"context"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	wsmanapi "github.com/bhojpur/platform/bp-manager/api"
	"github.com/bhojpur/platform/test/pkg/integration"
)

func TestPrebuildApplicationTaskSuccess(t *testing.T) {
	f := features.New("prebuild").
		WithLabel("component", "bp-manager").
		Assess("it should create a prebuild and succeed the defined tasks", func(_ context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			t.Cleanup(func() {
				api.Done(t)
			})

			ws, err := integration.LaunchApplicationDirectly(ctx, api, integration.WithRequestModifier(func(req *wsmanapi.StartWorkspaceRequest) error {
				req.Type = wsmanapi.ApplicationType_PREBUILD
				req.Spec.Envvars = append(req.Spec.Envvars, &wsmanapi.EnvironmentVariable{
					Name:  "BHOJPUR_TASKS",
					Value: `[{ "init": "echo \"some output\" > someFile; sleep 20; exit 0;" }]`,
				})
				return nil
			}))
			if err != nil {
				t.Fatalf("cannot launch a application: %q", err)
			}

			t.Cleanup(func() {
				_, _ = integration.WaitForApplicationStop(ctx, api, ws.Req.Id)
			})

			return ctx
		}).
		Feature()

	testEnv.Test(t, f)
}

func TestPrebuildApplicationTaskFail(t *testing.T) {
	t.Skip("status never returns HeadlessTaskFailed (exit 1)")

	f := features.New("prebuild").
		WithLabel("component", "server").
		Assess("it should create a prebuild and fail after running the defined tasks", func(_ context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			t.Cleanup(func() {
				api.Done(t)
			})

			ws, err := integration.LaunchApplicationDirectly(ctx, api, integration.WithRequestModifier(func(req *wsmanapi.StartApplicationRequest) error {
				req.Type = wsmanapi.ApplicationType_PREBUILD
				req.Spec.Envvars = append(req.Spec.Envvars, &wsmanapi.EnvironmentVariable{
					Name:  "BHOJPUR_TASKS",
					Value: `[{ "init": "echo \"some output\" > someFile; sleep 20; exit 1;" }]`,
				})
				return nil
			}))
			if err != nil {
				t.Fatalf("cannot start application: %q", err)
			}

			_, err = integration.WaitForApplication(ctx, api, ws.Req.Id, func(status *wsmanapi.ApplicaitonStatus) bool {
				if status.Phase != wsmanapi.ApplicationPhase_STOPPED {
					return false
				}
				if status.Conditions.HeadlessTaskFailed == "" {
					t.Logf("Conditions: %v", status.Conditions)
					t.Fatal("expected HeadlessTaskFailed condition")
				}
				return true
			})
			if err != nil {
				t.Fatalf("cannot start application: %q", err)
			}

			t.Cleanup(func() {
				_, _ = integration.WaitForApplicationStop(ctx, api, ws.Req.Id)
			})

			return ctx
		}).
		Feature()

	testEnv.Test(t, f)
}
