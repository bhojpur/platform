// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package application

const (
	Component                    = "application"
	ContainerPort                = 23000
	DefaultApplicationImage      = "bhojpur/platform-full"
	DefaultApplicationImageVersion = "latest"
	ApplicationImage             = "platform-cms"
	ApplicationImageStableVersion = "commit-d8477d484d00967a92686642b33541aed824cb63" // stable version that will be updated manually on demand
	dcpSaaSImage                 = "platform-dcp"
  	scmSaaSImage                 = "platform-scm"
  	crmSaaSImage                 = "platform-crm"
  	erpSaaSImage                 = "platform-erp"
  	srmSaaSImage                 = "platform-srm"
  	odeSaaSImage                 = "platform-ode"
  	qvmSaaSImage                 = "platform-qvm"
	DockerUpImage                = "docker-up"
	SupervisorImage              = "supervisor"
	ApplicationkitImage          = "applicationkit"
	SupervisorPort               = 22999
)
