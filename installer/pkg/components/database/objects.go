// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package database

import (
	"github.com/bhojpur/platform/installer/pkg/common"
	"github.com/bhojpur/platform/installer/pkg/components/database/cloudsql"
	"github.com/bhojpur/platform/installer/pkg/components/database/external"
	"github.com/bhojpur/platform/installer/pkg/components/database/incluster"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
)

func cloudSqlEnabled(cfg *common.RenderContext) bool {
	return !pointer.BoolDeref(cfg.Config.Database.InCluster, false) && cfg.Config.Database.CloudSQL != nil
}

func externalEnabled(cfg *common.RenderContext) bool {
	return !pointer.BoolDeref(cfg.Config.Database.InCluster, false) && cfg.Config.Database.External != nil
}

func inClusterEnabled(cfg *common.RenderContext) bool {
	return pointer.BoolDeref(cfg.Config.Database.InCluster, false)
}

var Objects = common.CompositeRenderFunc(
	common.CompositeRenderFunc(func(cfg *common.RenderContext) ([]runtime.Object, error) {
		if inClusterEnabled(cfg) {
			return incluster.Objects(cfg)
		}
		if cloudSqlEnabled(cfg) {
			return cloudsql.Objects(cfg)
		}
		if externalEnabled(cfg) {
			return external.Objects(cfg)
		}
		return nil, nil
	}),
)

var Helm = common.CompositeHelmFunc(
	common.CompositeHelmFunc(func(cfg *common.RenderContext) ([]string, error) {
		if inClusterEnabled(cfg) {
			return incluster.Helm(cfg)
		}

		return nil, nil
	}),
)
