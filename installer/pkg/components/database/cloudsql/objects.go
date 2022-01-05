// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package cloudsql

import (
	"github.com/bhojpur/platform/installer/pkg/common"
	dbinit "github.com/bhojpur/platform/installer/pkg/components/database/init"
)

var Objects = common.CompositeRenderFunc(
	deployment,
	dbinit.Objects,
	rolebinding,
	common.DefaultServiceAccount(Component),
	common.GenerateService(Component, map[string]common.ServicePort{
		Component: {
			ContainerPort: Port,
			ServicePort:   Port,
		},
	}),
)
