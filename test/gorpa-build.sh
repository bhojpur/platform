#!/bin/bash
# Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
# Licensed under the GNU Affero General Public License (AGPL).
# See License-AGPL.txt in the project root for license information.

export CGO_ENABLED=0

mkdir -p bin

for AGENT in pkg/agent/*; do
    echo building agent "$AGENT"
    base=$(basename "$AGENT")
    go build -trimpath -ldflags="-buildid= -w -s" -o bin/bhojpur-integration-test-"${base%_agent}"-agent ./"$AGENT"
done

for COMPONENT in tests/components/*; do
    echo building test "$COMPONENT"
    OUTPUT=$(basename "$COMPONENT")
    go test -trimpath -ldflags="-buildid= -w -s" -c -o bin/"$OUTPUT".test ./"$COMPONENT"
done

go test -trimpath -ldflags="-buildid= -w -s" -o bin/application -c ./tests/application