// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package application

const (
	Component                    = "application"
	ContainerPort                = 23000
	DefaultApplicationImage      = "bhojpur/platform-full"
	DefaultApplicationImageVersion = "latest"
	ApplicationImage             = "app/dcp"
	ApplicationImageStableVersion = "commit-d8477d484d00967a92686642b33541aed824cb63" // stable version that will be updated manually on demand
	DCPSaaSImage                 = "app/DCP"
  SCMSaaSImage                 = "app/SCM"
  CRMSaaSImage                 = "app/CRM"
  ERPSaaSImage                 = "app/ERP"
  SRMSaaSImage                 = "app/SRM"
  ODESaaSImage                 = "app/ODE"
  QVMSaaSImage                 = "app/QVM"
	DockerUpImage                = "docker-up"
	SupervisorImage              = "supervisor"
	ApplicationkitImage          = "applicationkit"
	SupervisorPort               = 22999
)
