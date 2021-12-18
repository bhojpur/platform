// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package cluster

// Valid characters for affinities are alphanumeric, -, _, . and one / as a subdomain prefix
const (
	AffinityLabelMeta                = "bhojpur.net/workload_meta"
	AffinityLabelIDE                 = "bhojpur.net/workload_ide"
	AffinityLabelApplicationServices = "bhojpur.net/workload_application_services"
	AffinityLabelApplicationRegular  = "bhojpur.net/workload_application_regular"
	AffinityLabelApplicationHeadless = "bhojpur.net/workload_application_headless"
)

var AffinityList = []string{
	AffinityLabelMeta,
	AffinityLabelIDE,
	AffinityLabelApplicationServices,
	AffinityLabelApplicationRegular,
	AffinityLabelApplicationHeadless,
}
