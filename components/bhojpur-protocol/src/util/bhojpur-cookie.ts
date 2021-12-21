// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

import * as cookie from 'cookie';

/**
 * This cookie indicates whether the connected client is a Bhojpur.NET Platform user (= "has logged in within the last year") or not.
 * This is used by "bhojpur.net" and "www.bhojpur.net" to display different content/buttons.
 */
export const NAME = "bhojpur-user";
export const VALUE = "true";

/**
 * @param domain The domain of the Bhojpur.NET Platform installation is installed onto
 * @returns
 */
export function options(domain: string): cookie.CookieSerializeOptions {
    // Reference: https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies
    return {
        path: "/",                          // make sure we send the cookie to all sub-pages
        httpOnly: false,
        secure: false,
        maxAge: 60 * 60 * 24 * 365,         // 1 year
        sameSite: "lax",                    // default: true. "Lax" needed to ensure we see cookies from users that neavigate to bhojpur.net from external sites
        domain: `.${domain}`,               // explicilty include subdomains to not only cover "bhojpur.net", but also "www.bhojpur.net" or applications
    };
};

export function generateCookie(domain: string): string {
    return cookie.serialize(NAME, VALUE, options(domain));
};

export function isPresent(cookies: string): boolean {
    // needs to match the old (bhojpur-user=loggedIn) and new (bhojpur-user=true) values to ensure a smooth transition during rollout.
    return !!cookies.match(`${NAME}=`);
};
