// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

interface Window {
    bhojpur: {
        service: import('../bhojpur-service').BhojpurService
        appsService?: import('../apps-frontend-service').AppsFrontendService
    }
}
