#!/bin/bash
# Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
# Licensed under the GNU Affero General Public License (AGPL).
# See License-AGPL.txt in the project root for license information.

set -x

go install github.com/golang/mock/mockgen@v1.6.0

mockgen \
    -package=protocol \
    -self_package=github.com/bhojpur/platform/bhojpur-protocol \
    -source=bhojpur-service.go > mock.go_tmp > mock.go_tmp

mv mock.go_tmp mock.go

gorpa run components:update-license-header
