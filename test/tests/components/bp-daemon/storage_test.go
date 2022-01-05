// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package bpdaemon

import (
	"context"
	"fmt"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	agent "github.com/bhojpur/platform/test/pkg/agent/daemon/api"
	"github.com/bhojpur/platform/test/pkg/integration"
)

func TestCreateBucket(t *testing.T) {
	f := features.New("DaemonAgent.CreateBucket").
		WithLabel("component", "bp-daemon").
		Assess("it should create a bucket", func(_ context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			rsa, closer, err := integration.Instrument(integration.ComponentApplicationDaemon, "daemon", cfg.Namespace(), cfg.Client(),
				integration.WithApplicationkitLift(false),
				integration.WithContainer("bp-daemon"),
			)
			if err != nil {
				t.Fatal(err)
			}
			integration.DeferCloser(t, closer)

			var resp agent.CreateBucketResponse
			err = rsa.Call("DaemonAgent.CreateBucket", agent.CreateBucketRequest{
				Owner:       fmt.Sprintf("integration-test-%d", time.Now().UnixNano()),
				Application: "test-ws",
			}, &resp)
			if err != nil {
				t.Fatalf("cannot create bucket: %q", err)
			}

			return ctx
		}).
		Feature()

	testEnv.Test(t, f)
}
