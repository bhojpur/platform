// Copyright (c) 2018 Bhojpur Consulting Private Limited, India All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package server

import (
	"context"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	protocol "github.com/bhojpur/platform/bhojpur-protocol"
	"github.com/bhojpur/platform/test/pkg/integration"
)

func TestServerAccess(t *testing.T) {
	f := features.New("GetLoggedInUser").
		WithLabel("component", "server").
		Assess("it can get a not built-in logged user", func(_ context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			t.Cleanup(func() {
				api.Done(t)
			})

			server, err := api.BhojpurServer(integration.WithBhojpurUser(username))
			if err != nil {
				t.Fatalf("cannot get BhojpurServer: %q", err)
			}

			_, err = server.GetLoggedInUser(ctx)
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Feature()

	testEnv.Test(t, f)
}

func TestStartApplication(t *testing.T) {
	integration.SkipWithoutUsername(t, username)

	f := features.New("CreateApplication").
		WithLabel("component", "server").
		Assess("it can run application tasks", func(_ context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			t.Cleanup(func() {
				api.Done(t)
			})

			server, err := api.BhojpurServer(integration.WithBhojpurUser(username))
			if err != nil {
				t.Fatalf("cannot get BhojpurServer: %q", err)
			}

			resp, err := server.CreateApplication(ctx, &protocol.CreateApplicationOptions{
				ContextURL: "github.com/bhojpur/platform",
				Mode:       "force-new",
			})
			if err != nil {
				t.Fatalf("cannot start application: %q", err)
			}

			t.Cleanup(func() {
				cctx, ccancel := context.WithTimeout(context.Background(), 10*time.Second)
				err := server.StopApplication(cctx, resp.CreatedApplicationID)
				ccancel()
				if err != nil {
					t.Logf("cannot stop application: %q", err)
				}
			})

			t.Logf("created application: applicationID=%s url=%s", resp.CreatedApplicationID, resp.ApplicationURL)

			nfo, err := server.GetApplication(ctx, resp.CreatedApplicationID)
			if err != nil {
				t.Fatalf("cannot get application: %q", err)
			}
			if nfo.LatestInstance == nil {
				t.Fatal("CreateApplication did not start the application")
			}

			_, err = integration.WaitForApplicationStart(ctx, nfo.LatestInstance.ID, api)
			if err != nil {
				t.Fatalf("cannot get application: %q", err)
			}

			t.Logf("application is running: instanceID=%s", nfo.LatestInstance.ID)
			return ctx
		}).
		Feature()

	testEnv.Test(t, f)
}
