// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package components

import (
	"github.com/bhojpur/platform/installer/pkg/common"
	agentsmith "github.com/bhojpur/platform/installer/pkg/components/agent-smith"
	"github.com/bhojpur/platform/installer/pkg/components/blobserve"
	"github.com/bhojpur/platform/installer/pkg/components/cluster"
	contentservice "github.com/bhojpur/platform/installer/pkg/components/content-service"
	"github.com/bhojpur/platform/installer/pkg/components/dashboard"
	"github.com/bhojpur/platform/installer/pkg/components/database"
	dockerregistry "github.com/bhojpur/platform/installer/pkg/components/docker-registry"
	"github.com/bhojpur/platform/installer/pkg/components/bhojpur"
	ide_proxy "github.com/bhojpur/platform/installer/pkg/components/ide-proxy"
	imagebuildermk3 "github.com/bhojpur/platform/installer/pkg/components/image-builder-mk3"
	jaegeroperator "github.com/bhojpur/platform/installer/pkg/components/jaeger-operator"
	"github.com/bhojpur/platform/installer/pkg/components/migrations"
	"github.com/bhojpur/platform/installer/pkg/components/minio"
	openvsxproxy "github.com/bhojpur/platform/installer/pkg/components/openvsx-proxy"
	"github.com/bhojpur/platform/installer/pkg/components/proxy"
	"github.com/bhojpur/platform/installer/pkg/components/rabbitmq"
	registryfacade "github.com/bhojpur/platform/installer/pkg/components/registry-facade"
	"github.com/bhojpur/platform/installer/pkg/components/server"
	"github.com/bhojpur/platform/installer/pkg/components/workspace"
	wsdaemon "github.com/bhojpur/platform/installer/pkg/components/ws-daemon"
	wsmanager "github.com/bhojpur/platform/installer/pkg/components/ws-manager"
	wsmanagerbridge "github.com/bhojpur/platform/installer/pkg/components/ws-manager-bridge"
	wsproxy "github.com/bhojpur/platform/installer/pkg/components/ws-proxy"
	wsscheduler "github.com/bhojpur/platform/installer/pkg/components/ws-scheduler"
)

var MetaObjects = common.CompositeRenderFunc(
	contentservice.Objects,
	proxy.Objects,
	dashboard.Objects,
	database.Objects,
	ide_proxy.Objects,
	imagebuildermk3.Objects,
	migrations.Objects,
	minio.Objects,
	openvsxproxy.Objects,
	rabbitmq.Objects,
	server.Objects,
	wsmanagerbridge.Objects,
)

var ApplicationObjects = common.CompositeRenderFunc(
	agentsmith.Objects,
	blobserve.Objects,
	bhojpur.Objects,
	registryfacade.Objects,
	application.Objects,
	wsdaemon.Objects,
	wsmanager.Objects,
	wsproxy.Objects,
	wsscheduler.Objects,
)

var FullObjects = common.CompositeRenderFunc(
	MetaObjects,
	WorkspaceObjects,
)

var MetaHelmDependencies = common.CompositeHelmFunc(
	database.Helm,
	jaegeroperator.Helm,
	minio.Helm,
	rabbitmq.Helm,
)

var ApplicationHelmDependencies = common.CompositeHelmFunc()

var FullHelmDependencies = common.CompositeHelmFunc(
	MetaHelmDependencies,
	ApplicationHelmDependencies,
)

// Anything in the "common" section are included in all installation types

var CommonObjects = common.CompositeRenderFunc(
	dockerregistry.Objects,
	cluster.Objects,
)

var CommonHelmDependencies = common.CompositeHelmFunc(
	dockerregistry.Helm,
)
