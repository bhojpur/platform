// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

const REGEX_APPLICATION_ID = /[0-9a-z]{2,16}-[0-9a-z]{2,16}-[0-9a-z]{8}/;
const REGEX_APPLICATION_ID_EXACT = new RegExp(`^${REGEX_APPLICATION_ID.source}$`);
// We need to parse the application id precisely here to get the case '<some-str>-<port>-<wsid>.ws.' right
const REGEX_APPLICATION_ID_FROM_HOSTNAME = new RegExp(`(${REGEX_APPLICATION_ID.source})\.ws`);

const REGEX_APPLICATION_ID_LEGACY = /[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/;
const REGEX_APPLICATION_ID_LEGACY_EXACT = new RegExp(`^${REGEX_APPLICATION_ID_LEGACY.source}$`);
const REGEX_APPLICATION_ID_LEGACY_FROM_HOSTNAME = new RegExp(`(${REGEX_APPLICATION_ID_LEGACY.source})\.ws`);

/**
 * Hostname may be of the form:
 *  - moccasin-ferret-155799b3.ws-ap01.bhojpur.net
 *  - 1234-moccasin-ferret-155799b3.ws-ap01.bhojpur.net
 *  - webview-1234-moccasin-ferret-155799b3.ws-ap01.bhojpur.net (or any other string replacing webview)
 * @param hostname The hostname the request is headed to
 */
export const parseApplicationIdFromHostname = function(hostname: string) {
    const match = REGEX_APPLICATION_ID_FROM_HOSTNAME.exec(hostname);
    if (match && match.length >= 2) {
        return match[1];
    } else {
        const legacyMatch = REGEX_APPLICATION_ID_LEGACY_FROM_HOSTNAME.exec(hostname);
        if (legacyMatch && legacyMatch.length >= 2) {
            return legacyMatch[1];
        }
        return undefined;
    }
};

/** Equalls UUIDv4 (and REGEX_APPLICATION_ID_LEGACY!) */
const REGEX_INSTANCE_ID = /[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/;
const REGEX_INSTANCE_ID_EXACT = new RegExp(`^${REGEX_INSTANCE_ID.source}$`);

/**
 * @param maybeId
 * @returns
 */
export const matchesInstanceIdOrLegacyApplicationIdExactly = function(maybeId: string): boolean {
    return REGEX_INSTANCE_ID_EXACT.test(maybeId)
        || REGEX_APPLICATION_ID_LEGACY_EXACT.test(maybeId);
};

/**
 * @param maybeApplicationId
 * @returns
 */
export const matchesNewApplicationIdExactly = function(maybeApplicationId: string): boolean {
    return REGEX_APPLICATION_ID_EXACT.test(maybeApplicationId);
};
