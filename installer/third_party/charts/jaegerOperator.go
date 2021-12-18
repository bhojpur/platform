// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package charts

import (
	"embed"
)

// Imported from https://github.com/jaegertracing/helm-charts/tree/main/charts/jaeger-operator

//go:embed jaeger-operator/*
var jaegerOperator embed.FS

func JaegerOperator() *Chart {
	return &Chart{
		Name:     "jaeger-operator",
		Content:  &jaegerOperator,
		Location: "jaeger-operator/",
		AdditionalFiles: []string{
			"jaeger-operator/crd.yaml",
		},
	}
}
