// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package bpmanager

import (
	"context"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	wsmanager_api "github.com/bhojpur/platform/bp-manager/api"
	"github.com/bhojpur/platform/test/pkg/integration"
)

func TestGetApplications(t *testing.T) {
	f := features.New("applications").
		WithLabel("component", "bp-manager").
		Assess("it should get applications", func(_ context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			t.Cleanup(func() {
				api.Done(t)
			})

			wsman, err := api.ApplicationManager()
			if err != nil {
				t.Fatal(err)
			}

			_, err = wsman.GetApplications(ctx, &wsmanager_api.GetApplicationsRequest{})
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Feature()

	testEnv.Test(t, f)
}
