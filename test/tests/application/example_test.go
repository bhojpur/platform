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

	agent "github.com/bhojpur/platform/test/pkg/agent/application/api"
	"github.com/bhojpur/platform/test/pkg/integration"
)

func TestApplicationInstrumentation(t *testing.T) {
	f := features.New("instrumentation").
		WithLabel("component", "server").
		Assess("it can instrument an application", func(_ context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			t.Cleanup(func() {
				api.Done(t)
			})

			nfo, stopWs, err := integration.LaunchApplicationFromContextURL(ctx, "github.com/bhojpur/platform", username, api)
			if err != nil {
				t.Fatal(err)
			}
			defer stopWs(true)

			rsa, closer, err := integration.Instrument(integration.ComponentApplication, "application", cfg.Namespace(), cfg.Client(), integration.WithInstanceID(nfo.LatestInstance.ID))
			if err != nil {
				t.Fatal(err)
			}
			defer rsa.Close()
			integration.DeferCloser(t, closer)

			var ls agent.ListDirResponse
			err = rsa.Call("ApplicationAgent.ListDir", &agent.ListDirRequest{
				Dir: "/application/platform",
			}, &ls)
			if err != nil {
				t.Fatal(err)
			}
			for _, f := range ls.Files {
				t.Log(f)
			}

			return ctx
		}).
		Feature()

	testEnv.Test(t, f)
}

func TestLaunchApplicationDirectly(t *testing.T) {
	f := features.New("application").
		WithLabel("component", "server").
		Assess("it can run application tasks", func(_ context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			t.Cleanup(func() {
				api.Done(t)
			})

			nfo, err := integration.LaunchApplicationDirectly(ctx, api)
			if err != nil {
				t.Fatal(err)
			}

			err = integration.DeleteApplication(ctx, api, nfo.Req.Id)
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Feature()

	testEnv.Test(t, f)
}
