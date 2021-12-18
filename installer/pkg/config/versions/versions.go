// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package versions

type Manifest struct {
	Version    string     `json:"version"`
	Components Components `json:"components"`
}

type Versioned struct {
	Version string `json:"version"`
}

type Components struct {
	AgentSmith      Versioned `json:"agentSmith"`
	Blobserve       Versioned `json:"blobserve"`
	CAUpdater       Versioned `json:"caUpdater"`
	ContentService  Versioned `json:"contentService"`
	Dashboard       Versioned `json:"dashboard"`
	DBMigrations    Versioned `json:"dbMigrations"`
	DBSync          Versioned `json:"dbSync"`
	IDEProxy        Versioned `json:"ideProxy"`
	ImageBuilder    Versioned `json:"imageBuilder"`
	ImageBuilderMk3 struct {
		Versioned
		BuilderImage Versioned `json:"builderImage"`
	} `json:"imageBuilderMk3"`
	IntegrationTests Versioned `json:"integrationTests"`
	Kedge            Versioned `json:"kedge"`
	OpenVSXProxy     Versioned `json:"openVSXProxy"`
	PaymentEndpoint  Versioned `json:"paymentEndpoint"`
	Proxy            Versioned `json:"proxy"`
	RegistryFacade   Versioned `json:"registryFacade"`
	Server           Versioned `json:"server"`
	ServiceWaiter    Versioned `json:"serviceWaiter"`
	Platform         struct {
		SaaSImage        Versioned `json:"saasImage"`
		DockerUp         Versioned `json:"dockerUp"`
		Supervisor       Versioned `json:"supervisor"`
		Applicationkit   Versioned `json:"applicationkit"`
		ApplicationImages struct {
			DCPSaaSImage  Versioned `json:"dcpSaaS"`
			SCMSaaSImage  Versioned `json:"scmSaaS"`
			CRMSaaSImage  Versioned `json:"crmSaaS"`
			ERPSaaSImage  Versioned `json:"erpSaaS"`
			MRPSaaSImage  Versioned `json:"mrpSaaS"`
      SRMSaaSImage  Versioned `json:"srmSaaS"`
			FMSSaaSImage  Versioned `json:"fmsSaaS"`
      ODESaaSImage  Versioned `json:"odeSaaS"`
			QVMSaaSImage  Versioned `json:"qvmSaaS"`
    } `json:"desktopIdeImages"`
	} `json:"workspace"`
	WSDaemon struct {
		Versioned

		UserNamespaces struct {
			SeccompProfileInstaller Versioned `json:"seccompProfileInstaller"`
			ShiftFSModuleLoader     Versioned `json:"shiftfsModuleLoader"`
		} `json:"userNamespaces"`
	} `json:"wsDaemon"`
	WSManager       Versioned `json:"wsManager"`
	WSManagerBridge Versioned `json:"wsManagerBridge"`
	WSProxy         Versioned `json:"wsProxy"`
	WSScheduler     Versioned `json:"wsScheduler"`
}
