// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package net.bhojpur.platform.api;

public class BhojpurClientImpl implements BhojpurClient {

    private BhojpurServer server;

    @Override
    public void connect(BhojpurServer server) {
        this.server = server;
    }

    @Override
    public BhojpurServer server() {
        if (this.server == null) {
            throw new IllegalStateException("server is null");
        }
        return this.server;
    }
}
