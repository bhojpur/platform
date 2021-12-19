// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package blobserve

import (
	"github.com/bhojpur/platform/installer/pkg/common"
)

const (
	Component       = "blobserve"
	ContainerPort   = 32224
	ServicePort     = common.BlobServeServicePort
	ServicePortName = "service"
	MaxSizeBytes    = 1024 * 1024 * 1024 // 1 Gibibyte
)
