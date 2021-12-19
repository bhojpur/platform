// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package application

import "github.com/bhojpur/platform/installer/pkg/common"

var Objects = common.CompositeRenderFunc(
	networkpolicy,
	podsecuritypolicies,
	role,
	rolebinding,
	common.DefaultServiceAccount(Component),
)
