// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

/**
* These cookies are set in the IDE frontend. This pattern is relied upon in:
*  - proxy:
*    - to filter it out on port locations
*    - to forward it to the server for authentication
*  - server:
*    - to authenticate access to port locations
*/
export const applicationPortAuthCookieName = function(host: string, applicationId: string): string {
    return host
        .replace(/https?/, '')
        .replace(/[\W_]+/g, "_")
        + `_ws_${applicationId}_port_auth_`;
};
