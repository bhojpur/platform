// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

import * as chai from 'chai';
import { suite, test } from 'mocha-typescript';
import { ContextURL } from './context-url';
const expect = chai.expect;

@suite
export class ContextUrlTest {

    @test public parseContextUrl_withEnvVar() {
        const actual = ContextURL.parseToURL("passedin=test%20value/https://github.com/bhojpur/platform-test-repo");
        expect(actual?.host).to.equal("github.com");
        expect(actual?.pathname).to.equal("/bhojpur/platform-test-repo");
    }

    @test public parseContextUrl_withEnvVar_withoutSchema() {
        const actual = ContextURL.parseToURL("passedin=test%20value/github.com/bhojpur/platform-test-repo");
        expect(actual?.host).to.equal("github.com");
        expect(actual?.pathname).to.equal("/bhojpur/platform-test-repo");
    }

    @test public parseContextUrl_withPrebuild() {
        const actual = ContextURL.parseToURL("prebuild/https://github.com/bhojpur/platform-test-repo");
        expect(actual?.host).to.equal("github.com");
        expect(actual?.pathname).to.equal("/bhojpur/platform-test-repo");
    }

    @test public parseContextUrl_withPrebuild_withoutSchema() {
        const actual = ContextURL.parseToURL("prebuild/github.com/bhojpur/platform-test-repo");
        expect(actual?.host).to.equal("github.com");
        expect(actual?.pathname).to.equal("/bhojpur/platform-test-repo");
    }
}
module.exports = new ContextUrlTest()
