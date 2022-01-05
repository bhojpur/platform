// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package application

import (
	"context"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	wsmanapi "github.com/bhojpur/platform/bp-manager/api"
	"github.com/bhojpur/platform/test/pkg/integration"
)

func TestGhostApplication(t *testing.T) {
	f := features.New("ghost").
		WithLabel("component", "bp-manager").
		Assess("it can start a ghost application", func(_ context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			t.Cleanup(func() {
				api.Done(t)
			})

			// there's nothing specific about ghost that we want to test beyond that they start properly
			ws, err := integration.LaunchApplicationDirectly(ctx, api, integration.WithRequestModifier(func(req *wsmanapi.StartApplicationRequest) error {
				req.Type = wsmanapi.ApplicationType_GHOST
				req.Spec.Envvars = append(req.Spec.Envvars, &wsmanapi.EnvironmentVariable{
					Name:  "BHOJPUR_TASKS",
					Value: `[{ "init": "echo \"some output\" > someFile; sleep 20; exit 0;" }]`,
				})
				return nil
			}))
			if err != nil {
				t.Fatal(err)
			}

			_, err = integration.WaitForApplicationStart(ctx, ws.Req.Id, api)
			if err != nil {
				t.Fatal(err)
			}

			err = integration.DeleteApplication(ctx, api, ws.Req.Id)
			if err != nil {
				t.Fatal(err)
			}
			return ctx
		}).
		Feature()

	testEnv.Test(t, f)
}
