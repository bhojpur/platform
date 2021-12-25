// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package common

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// This file exists to break cyclic-dependency errors

const (
	AppName                     = "bhojpur"
	BlobServeServicePort        = 4000
	CertManagerCAIssuer         = "ca-issuer"
	DockerRegistryURL           = "docker.io"
	DockerRegistryName          = "registry"
	BhojpurContainerRegistry    = "us-west2-docker.pkg.dev/bhojpur/platform/build"
	InClusterDbSecret           = "mysql"
	InClusterMessageQueueName   = "rabbitmq"
	InClusterMessageQueueTLS    = "messagebus-certificates-secret-core"
	KubeRBACProxyRepo           = "quay.io"
	KubeRBACProxyImage          = "brancz/kube-rbac-proxy"
	KubeRBACProxyTag            = "v0.11.0"
	MinioServiceAPIPort         = 9000
	MonitoringChart             = "monitoring"
	ProxyComponent              = "proxy"
	ProxyContainerHTTPPort      = 80
	ProxyContainerHTTPName      = "http"
	ProxyContainerHTTPSPort     = 443
	ProxyContainerHTTPSName     = "https"
	RegistryAuthSecret          = "builtin-registry-auth"
	RegistryTLSCertSecret       = "builtin-registry-certs"
	RegistryFacadeComponent     = "registry-facade"
	RegistryFacadeServicePort   = 30000
	RegistryFacadeTLSCertSecret = "builtin-registry-facade-cert"
	ServerComponent             = "server"
	SystemNodeCritical          = "system-node-critical"
	WSManagerComponent          = "bp-manager"
	WSManagerBridgeComponent    = "bp-manager-bridge"
	WSProxyComponent            = "bp-proxy"
	WSSchedulerComponent        = "bp-scheduler"

	AnnotationConfigChecksum    = "bhojpur.net/checksum_config"
)

var (
	InternalCertDuration = &metav1.Duration{Duration: time.Hour * 24 * 90}
)
