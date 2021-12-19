// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

// The cluster package is designed for cluster-level objects that are always installed
// These will mostly be security related

package cluster

import "github.com/bhojpur/platform/installer/pkg/common"

var Objects = common.CompositeRenderFunc(
	certmanager,
	clusterrole,
	podsecuritypolicies,
	resourcequota,
	rolebinding,
	common.DefaultServiceAccount(NobodyComponent),
)
