// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package application

import (
	"context"
	"net/rpc"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	agent "github.com/bhojpur/platform/test/pkg/agent/application/api"
	"github.com/bhojpur/platform/test/pkg/integration"
	"github.com/bhojpur/platform/test/tests/application/common"
)

//
type GitTest struct {
	Skip            bool
	Name            string
	ContextURL      string
	ApplicationRoot string
	Action          GitFunc
}

type GitFunc func(rsa *rpc.Client, git common.GitClient, applicationRoot string) error

func TestGitActions(t *testing.T) {
	tests := []GitTest{
		{
			Name:            "create, add and commit",
			ContextURL:      "github.com/bhojpur/platform-test-repo/tree/integration-test/commit-and-push",
			ApplicationRoot: "/application/platform-test-repo",
			Action: func(rsa *rpc.Client, git common.GitClient, applicationRoot string) (err error) {

				var resp agent.ExecResponse
				err = rsa.Call("ApplicationAgent.Exec", &agent.ExecRequest{
					Dir:     applicationRoot,
					Command: "bash",
					Args: []string{
						"-c",
						"echo \"another test run...\" >> file_to_commit.txt",
					},
				}, &resp)
				if err != nil {
					return err
				}
				err = git.Add(applicationRoot)
				if err != nil {
					return err
				}
				err = git.Commit(applicationRoot, "automatic test commit", false)
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			Skip:            true,
			Name:            "create, add and commit and PUSH",
			ContextURL:      "github.com/bhojpur/platform-test-repo/tree/integration-test/commit-and-push",
			ApplicationRoot: "/application/platform-test-repo",
			Action: func(rsa *rpc.Client, git common.GitClient, applicationRoot string) (err error) {

				var resp agent.ExecResponse
				err = rsa.Call("ApplicationAgent.Exec", &agent.ExecRequest{
					Dir:     applicationRoot,
					Command: "bash",
					Args: []string{
						"-c",
						"echo \"another test run...\" >> file_to_commit.txt",
					},
				}, &resp)
				if err != nil {
					return err
				}
				err = git.Add(applicationRoot)
				if err != nil {
					return err
				}
				err = git.Commit(applicationRoot, "automatic test commit", false)
				if err != nil {
					return err
				}
				err = git.Push(applicationRoot, false)
				if err != nil {
					return err
				}
				return nil
			},
		},
	}

	f := features.New("GitActions").
		WithLabel("component", "server").
		Assess("it can run git actions", func(_ context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			t.Cleanup(func() {
				api.Done(t)
			})

			for _, test := range tests {
				t.Run(test.ContextURL, func(t *testing.T) {
					if test.Skip {
						t.SkipNow()
					}

					nfo, stopWS, err := integration.LaunchApplicationFromContextURL(ctx, test.ContextURL, username, api)
					if err != nil {
						t.Fatal(err)
					}

					defer stopWS(false)

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

					git := common.Git(rsa)
					err = test.Action(rsa, git, test.ApplicationRoot)
					if err != nil {
						t.Fatal(err)
					}
				})
			}

			return ctx
		}).
		Feature()

	testEnv.Test(t, f)
}
