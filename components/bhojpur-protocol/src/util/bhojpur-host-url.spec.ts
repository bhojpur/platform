// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

import * as chai from 'chai';
import { suite, test } from 'mocha-typescript';
import { BhojpurHostUrl } from './bhojpur-host-url';
const expect = chai.expect;

@suite
export class BhojpurHostUrlTest {

    @test public parseApplicationId_pathBased() {
        const actual = BhojpurHostUrl.fromApplicationUrl("https://code.bhojpur.net/application/bc77e03d-c781-4235-bca0-e24087f5e472/").applicationId;
        expect(actual).to.equal("bc77e03d-c781-4235-bca0-e24087f5e472");
    }

    @test public parseApplicationId_hosts_withEnvVarsInjected() {
        const actual = BhojpurHostUrl.fromApplicationUrl("https://gray-grasshopper-nfbitfia.ws-ap01.staging.bhojpur.net/#passedin=test%20value/https://github.com/bhojpur/bhojpur-test-repo").applicationId;
        expect(actual).to.equal("gray-grasshopper-nfbitfia");
    }

    @test public async testWithoutApplicationPrefix() {
        expect(BhojpurHostUrl.fromApplicationUrl("https://3000-moccasin-ferret-155799b3.ws-ap01.staging.bhojpur.net/").withoutApplicationPrefix().toString()).to.equal("https://staging.bhojpur.net/");
    }

    @test public async testWithoutApplicationPrefix2() {
        expect(BhojpurHostUrl.fromApplicationUrl("https://staging.bhojpur.net/").withoutApplicationPrefix().toString()).to.equal("https://staging.bhojpur.net/");
    }

    @test public async testWithoutApplicationPrefix3() {
        expect(BhojpurHostUrl.fromApplicationUrl("https://gray-rook-5523v5d8.ws-dev.my-branch-1234.staging.bhojpur.net/").withoutApplicationPrefix().toString()).to.equal("https://my-branch-1234.staging.bhojpur.net/");
    }

    @test public async testWithoutApplicationPrefix4() {
        expect(BhojpurHostUrl.fromApplicationUrl("https://my-branch-1234.staging.bhojpur.net/").withoutApplicationPrefix().toString()).to.equal("https://my-branch-1234.staging.bhojpur.net/");
    }

    @test public async testWithoutApplicationPrefix5() {
        expect(BhojpurHostUrl.fromApplicationUrl("https://abc-nice-brunch-4224.staging.bhojpur.net/").withoutApplicationPrefix().toString()).to.equal("https://abc-nice-brunch-4224.staging.bhojpur.net/");
    }

    @test public async testWithoutApplicationPrefix6() {
        expect(BhojpurHostUrl.fromApplicationUrl("https://gray-rook-5523v5d8.ws-dev.abc-nice-brunch-4224.staging.bhojpur.net/").withoutApplicationPrefix().toString()).to.equal("https://abc-nice-brunch-4224.staging.bhojpur.net/");
    }
}
module.exports = new BhojpurHostUrlTest()
