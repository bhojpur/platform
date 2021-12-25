// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

const URL = require('url').URL || window.URL;
import { log } from './logging';

export interface UrlChange {
    (old: URL): Partial<URL>
}
export type UrlUpdate = UrlChange | Partial<URL>;

const baseApplicationIDRegex = "(([a-f][0-9a-f]{7}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})|([0-9a-z]{2,16}-[0-9a-z]{2,16}-[0-9a-z]{8}))";

// this pattern matches v4 UUIDs as well as the new generated application ids (e.g. pink-panda-ns35kd21)
const applicationIDRegex = RegExp(`^${baseApplicationIDRegex}$`);

// this pattern matches URL prefixes of applications
const applicationUrlPrefixRegex = RegExp(`^([0-9]{4,6}-)?${baseApplicationIDRegex}\\.`);

export class BhojpurHostUrl {
    readonly url: URL;

    constructor(urlParam?: string | URL) {
        if (urlParam === undefined || typeof urlParam === 'string') {
            this.url = new URL(urlParam || 'https://bhojpur.net');
            this.url.search = '';
            this.url.hash = '';
            this.url.pathname = '';
        } else if (urlParam instanceof URL) {
            this.url = urlParam;
        } else {
            log.error('Unexpected urlParam', { urlParam });
        }
    }

    static fromApplicationUrl(url: string) {
        return new BhojpurHostUrl(new URL(url));
    }

    withApplicationPrefix(applicationId: string, region: string) {
        return this.withDomainPrefix(`${applicationId}.ws-${region}.`);
    }

    withDomainPrefix(prefix: string): BhojpurHostUrl {
        return this.with(url => ({ host: prefix + url.host }));;
    }

    withoutApplicationPrefix(): BhojpurHostUrl {
        if (!this.url.host.match(applicationUrlPrefixRegex)) {
            // by default, the URL has no Bhojpur.NET Platform application prefix
            return this;
        }

        return this.withoutDomainPrefix(2);
    }

    withoutDomainPrefix(removeSegmentsCount: number): BhojpurHostUrl {
        return this.with(url => ({ host: url.host.split('.').splice(removeSegmentsCount).join('.') }));
    }

    with(urlUpdate: UrlUpdate) {
        const update = typeof urlUpdate === 'function' ? urlUpdate(this.url) : urlUpdate;
        const addSlashToPath = update.pathname && update.pathname.length > 0 && !update.pathname.startsWith('/');
        if (addSlashToPath) {
            update.pathname = '/' + update.pathname;
        }
        const result = Object.assign(new URL(this.toString()), update);
        return new BhojpurHostUrl(result);
    }

    withApi(urlUpdate?: UrlUpdate) {
        const updated = urlUpdate ? this.with(urlUpdate) : this;
        return updated.with(url => ({ pathname: `/api${url.pathname}` }));
    }

    withContext(contextUrl: string) {
        return this.with(url => ({ hash: contextUrl }));
    }

    asWebsocket(): BhojpurHostUrl {
        return this.with(url => ({ protocol: url.protocol === 'https:' ? 'wss:' : 'ws:' }));
    }

    asDashboard(): BhojpurHostUrl {
        return this.with(url => ({ pathname: '/' }));
    }

    asAbout(): BhojpurHostUrl {
        return this.with(url => ({ pathname: '/about' }));
    }

    asLogin(): BhojpurHostUrl {
        return this.with(url => ({ pathname: '/login' }));
    }

    asUpgradeSubscription(): BhojpurHostUrl {
        return this.with(url => ({ pathname: '/plans' }));
    }

    asAccessControl(): BhojpurHostUrl {
        return this.with(url => ({ pathname: '/integrations' }));
    }

    asSettings(): BhojpurHostUrl {
        return this.with(url => ({ pathname: '/settings' }));
    }

    asPreferences(): BhojpurHostUrl {
        return this.with(url => ({ pathname: '/preferences' }));
    }

    asSupportServices(): BhojpurHostUrl {
        return this.with(url => ({ pathname: '/support' }));
    }

    asDocumentation(): BhojpurHostUrl {
        return this.with(url => ({ pathname: '/document' }));
    }

    asGraphQLApi(): BhojpurHostUrl {
        return this.with(url => ({ pathname: '/graphql/' }));
    }

    asStart(applicationId = this.applicationId): BhojpurHostUrl {
        return this.withoutApplicationPrefix().with({
            pathname: '/start/',
            hash: '#' + applicationId
        });
    }

    asApplicationAuth(instanceID: string, redirect?: boolean): BhojpurHostUrl {
        return this.with(url => ({ pathname: `/api/auth/application-cookie/${instanceID}`, search: redirect ? "redirect" : "" }));
    }

    toString() {
        return this.url.toString();
    }

    toStringWoRootSlash() {
        let result = this.toString();
        if (result.endsWith('/')) {
            result = result.slice(0, result.length - 1);
        }
        return result;
    }

    get applicationId(): string | undefined {
        const hostSegs = this.url.host.split(".");
        if (hostSegs.length > 1) {
            const matchResults = hostSegs[0].match(applicationIDRegex);
            if (matchResults) {
                // URL has a Bhojpur.NET Platform application prefix
                // port prefixes are excluded
                return matchResults[0];
            }
        }

        const pathSegs = this.url.pathname.split("/")
        if (pathSegs.length > 3 && pathSegs[1] === "application") {
            return pathSegs[2];
        }

        return undefined;
    }

    get blobServe(): boolean {
        const hostSegments = this.url.host.split(".");
        if (hostSegments[0] === 'blobserve') {
            return true;
        }

        const pathSegments = this.url.pathname.split("/")
        return pathSegments[0] === "blobserve";
    }

    asSorry(message: string) {
        return this.with({ pathname: '/sorry', hash: message });
    }

    asApiLogout(): BhojpurHostUrl {
        return this.withApi(url => ({ pathname: '/logout/' }));
    }
}
