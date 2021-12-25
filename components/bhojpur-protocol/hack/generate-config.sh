#!/bin/bash

go install -u github.com/a-h/generate/...

schema-generate -p protocol ../data/bhojpur-schema.json > ../go/bhojpur-config-types.go

sed -i 's/json:/yaml:/g' ../go/bhojpur-config-types.go
gofmt -w ../go/bhojpur-config-types.go

gorpa run components:update-license-header
