// Copyright (c) 2018 Bhojpur Consulting Private Limited, Indi. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/bhojpur/platform/test/pkg/integration"
	"github.com/bhojpur/platform/test/tests/application/common"
)

type ContextTest struct {
	Skip               bool
	Name               string
	ContextURL         string
	ApplicationRoot    string
	ExpectedBranch     string
	ExpectedBranchFunc func(username string) string
}

func TestGitHubContexts(t *testing.T) {
	tests := []ContextTest{
		{
			Name:            "open repository",
			ContextURL:      "github.com/bhojpur/platform",
			ApplicationRoot: "/application/platform",
			ExpectedBranch:  "main",
		},
		{
			Name:            "open branch",
			ContextURL:      "github.com/bhojpur/platform-test-repo/tree/integration-test-1",
			ApplicationRoot: "/application/platform-test-repo",
			ExpectedBranch:  "integration-test-1",
		},
		{
			Name:            "open issue",
			ContextURL:      "github.com/bhojpur/platform-test-repo/issues/88",
			ApplicationRoot: "/application/platform-test-repo",
			ExpectedBranchFunc: func(username string) string {
				return fmt.Sprintf("%s/integration-tests-test-context-88", username)
			},
		},
		{
			Name:            "open tag",
			ContextURL:      "github.com/bhojpur/platform-test-repo/tree/integration-test-context-tag",
			ApplicationRoot: "/application/platform-test-repo",
			ExpectedBranch:  "HEAD",
		},
	}
	runContextTests(t, tests)
}

func TestGitLabContexts(t *testing.T) {
	integration.SkipWithoutUsername(t, username)

	tests := []ContextTest{
		{
			Name:            "open repository",
			ContextURL:      "gitlab.com/bhojpur/bp-test",
			ApplicationRoot: "/application/bp-test",
			ExpectedBranch:  "master",
		},
		{
			Name:            "open branch",
			ContextURL:      "gitlab.com/bhojpur/bp-test/tree/wip",
			ApplicationRoot: "/application/bp-test",
			ExpectedBranch:  "wip",
		},
		{
			Name:            "open issue",
			ContextURL:      "gitlab.com/bhojpur/bp-test/issues/1",
			ApplicationRoot: "/application/bp-test",
			ExpectedBranchFunc: func(username string) string {
				return fmt.Sprintf("%s/write-a-readme-1", username)
			},
		},
		{
			Name:            "open tag",
			ContextURL:      "gitlab.com/bhojpur/bp-test/merge_requests/2",
			ApplicationRoot: "/application/bp-test",
			ExpectedBranch:  "wip2",
		},
	}
	runContextTests(t, tests)
}

func runContextTests(t *testing.T, tests []ContextTest) {
	f := features.New("context").
		WithLabel("component", "server").
		Assess("should run context tests", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			for _, test := range tests {
				t.Run(test.ContextURL, func(t *testing.T) {
					if test.Skip {
						t.SkipNow()
					}

					t.Parallel()

					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
					defer cancel()

					api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
					t.Cleanup(func() {
						api.Done(t)
					})

					if username == "" && test.ExpectedBranchFunc != nil {
						t.Skipf("skipping '%s' because there is not username configured", test.Name)
					}

					nfo, stopWS, err := integration.LaunchApplicationFromContextURL(ctx, test.ContextURL, username, api)
					if err != nil {
						t.Fatal(err)
					}
					defer stopWS(false) // we do not wait for stopped here as it does not matter for this test case and speeds things up

					_, err = integration.WaitForApplicationStart(ctx, nfo.LatestInstance.ID, api)
					if err != nil {
						t.Fatal(err)
					}

					rsa, closer, err := integration.Instrument(integration.ComponentApplication, "application", cfg.Namespace(), cfg.Client(), integration.WithInstanceID(nfo.LatestInstance.ID))
					if err != nil {
						t.Fatal(err)
					}
					defer rsa.Close()
					integration.DeferCloser(t, closer)

					// get actual from application
					git := common.Git(rsa)
					actBranch, err := git.GetBranch(test.ApplicationRoot)
					if err != nil {
						t.Fatal(err)
					}

					expectedBranch := test.ExpectedBranch
					if test.ExpectedBranchFunc != nil {
						expectedBranch = test.ExpectedBranchFunc(username)
					}
					if actBranch != expectedBranch {
						t.Fatalf("expected branch '%s', got '%s'!", test.ExpectedBranch, actBranch)
					}
				})
			}
			return ctx
		}).
		Feature()

	testEnv.Test(t, f)
}
