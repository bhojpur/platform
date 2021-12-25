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
		dcpSaaSImage  Versioned `json:"dcpSaaS"`
		scmSaaSImage  Versioned `json:"scmSaaS"`
		crmSaaSImage  Versioned `json:"crmSaaS"`
		erpSaaSImage  Versioned `json:"erpSaaS"`
		mrpSaaSImage  Versioned `json:"mrpSaaS"`
      		srmSaaSImage  Versioned `json:"srmSaaS"`
		fmsSaaSImage  Versioned `json:"fmsSaaS"`
      		odeSaaSImage  Versioned `json:"odeSaaS"`
		qvmSaaSImage  Versioned `json:"qvmSaaS"`
    } `json:"desktopIdeImages"`
	} `json:"workspace"`
	WSDaemon struct {
		Versioned

		UserNamespaces struct {
			SeccompProfileInstaller Versioned `json:"seccompProfileInstaller"`
			ShiftFSModuleLoader     Versioned `json:"shiftfsModuleLoader"`
		} `json:"userNamespaces"`
	} `json:"wsDaemon"`
	WSManager       Versioned `json:"bpManager"`
	WSManagerBridge Versioned `json:"bpManagerBridge"`
	WSProxy         Versioned `json:"bpProxy"`
	WSScheduler     Versioned `json:"bpScheduler"`
}
