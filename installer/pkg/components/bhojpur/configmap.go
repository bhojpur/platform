// Copyright (c) 2018 Bhojpur consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package bhojpur

import (
	"encoding/json"
	"fmt"

	"github.com/bhojpur/platform/installer/pkg/common"
	"github.com/bhojpur/platform/installer/pkg/config/versions"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Bhojpur struct {
	VersionManifest versions.Manifest `json:"versions"`
}

func configmap(ctx *common.RenderContext) ([]runtime.Object, error) {
	gpcfg := Bhojpur{
		VersionManifest: ctx.VersionManifest,
	}

	fc, err := json.MarshalIndent(gpcfg, "", " ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Bhojpur.NET Platform config: %w", err)
	}

	return []runtime.Object{
		&corev1.ConfigMap{
			TypeMeta: common.TypeMetaConfigmap,
			ObjectMeta: metav1.ObjectMeta{
				Name:      Component,
				Namespace: ctx.Namespace,
				Labels:    common.DefaultLabels(Component),
			},
			Data: map[string]string{
				"config.json": string(fc),
			},
		},
	}, nil
}
