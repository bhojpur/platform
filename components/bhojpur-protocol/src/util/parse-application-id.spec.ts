// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

import * as chai from 'chai';
import { suite, test } from 'mocha-typescript';
import { matchesInstanceIdOrLegacyApplicationIdExactly, matchesNewApplicationIdExactly, parseApplicationIdFromHostname } from './parse-application-id';
const expect = chai.expect;

@suite
export class ParseApplicationIdTest {

    @test public parseApplicationIdFromHostname_fromApplicationLocation() {
        const actual = parseApplicationIdFromHostname("moccasin-ferret-155799b3.ws-ap01.bhojpur.net");
        expect(actual).to.equal("moccasin-ferret-155799b3");
    }

    @test public parseApplicationIdFromHostname_fromApplicationPortLocation() {
        const actual = parseApplicationIdFromHostname("3000-moccasin-ferret-155799b3.ws-ap01.bhojpur.net");
        expect(actual).to.equal("moccasin-ferret-155799b3");
    }

    @test public parseApplicationIdFromHostname_fromApplicationPortLocationWithWebviewPrefix() {
        const actual = parseApplicationIdFromHostname("webview-3000-moccasin-ferret-155799b3.ws-ap01.bhojpur.net");
        expect(actual).to.equal("moccasin-ferret-155799b3");
    }

    @test public parseApplicationIdFromHostname_fromApplicationPortLocationWithWebviewPrefixCustomHost() {
        const actual = parseApplicationIdFromHostname("webview-3000-moccasin-ferret-155799b3.ws-ap01.some.subdomain.somehost.com");
        expect(actual).to.equal("moccasin-ferret-155799b3");
    }

    // legacy mode
    @test public parseLegacyApplicationIdFromHostname_fromApplicationLocation() {
        const actual = parseApplicationIdFromHostname("b7e0eaf8-ec73-44ec-81ea-04859263b656.ws-ap01.bhojpur.net");
        expect(actual).to.equal("b7e0eaf8-ec73-44ec-81ea-04859263b656");
    }

    @test public parseLegacyApplicationIdFromHostname_fromApplicationPortLocation() {
        const actual = parseApplicationIdFromHostname("3000-b7e0eaf8-ec73-44ec-81ea-04859263b656.ws-ap01.bhojpur.net");
        expect(actual).to.equal("b7e0eaf8-ec73-44ec-81ea-04859263b656");
    }

    @test public parseLegacyApplicationIdFromHostname_fromApplicationPortLocationWithWebviewPrefix() {
        const actual = parseApplicationIdFromHostname("webview-3000-b7e0eaf8-ec73-44ec-81ea-04859263b656.ws-ap01.bhojpur.net");
        expect(actual).to.equal("b7e0eaf8-ec73-44ec-81ea-04859263b656");
    }

    @test public parseLegacyApplicationIdFromHostname_fromApplicationPortLocationWithWebviewPrefixCustomHost() {
        const actual = parseApplicationIdFromHostname("webview-3000-ca81a50f-09d7-465c-acd9-264a747d5351.ws-ap01.some.subdomain.somehost.com");
        expect(actual).to.equal("ca81a50f-09d7-465c-acd9-264a747d5351");
    }

    // match - instance ID
    @test public matchesInstanceIdOrLegacyApplicationIdExactly_positive() {
        const actual = matchesInstanceIdOrLegacyApplicationIdExactly("b7e0eaf8-ec73-44ec-81ea-04859263b656");
        expect(actual).to.be.true;
    }
    @test public matchesInstanceIdOrLegacyApplicationIdExactly_negative() {
        const actual = matchesInstanceIdOrLegacyApplicationIdExactly("b7e0eaf8-ec73-44ec-81a-04859263b656");
        expect(actual).to.be.false;
    }

    // match - new application ID
    @test public matchesApplicationIdExactly_new_positive() {
        const actual = matchesNewApplicationIdExactly("moccasin-ferret-155799b3");
        expect(actual).to.be.true;
    }
    @test public matchesApplicationIdExactly_new_negative() {
        const actual = matchesNewApplicationIdExactly("moccasin-ferret-15599b3");
        expect(actual).to.be.false;
    }
}
module.exports = new ParseApplicationIdTest()
