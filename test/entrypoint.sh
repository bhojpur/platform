#!/bin/sh
# Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
# Licensed under the GNU Affero General Public License (AGPL).
# See License-AGPL.txt in the project root for license information.

set -ex

export PATH=$PATH:/tests

# shellcheck disable=SC2045
for i in $(find /tests/ -name "*.test" | sort -R); do
    "$i" "$@";
done