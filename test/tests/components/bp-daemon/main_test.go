// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package bpdaemon

import (
	"context"
	"os"
	"testing"

	"github.com/bhojpur/platform/test/pkg/integration"
	"sigs.k8s.io/e2e-framework/pkg/env"
)

var (
	testEnv   env.Environment
	username  string
	namespace string
)

func TestMain(m *testing.M) {
	username, namespace, testEnv = integration.Setup(context.Background())
	os.Exit(testEnv.Run(m))
}
