// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package net.bhojpur.platform.testclient;

import net.bhojpur.platform.api.ConnectionHelper;
import net.bhojpur.platform.api.BhojpurClient;
import net.bhojpur.platform.api.BhojpurServer;
import net.bhojpur.platform.api.entities.SendHeartBeatOptions;
import net.bhojpur.platform.api.entities.User;

public class TestClient {
    public static void main(String[] args) throws Exception {
        String uri = "wss://bhojpur.net/api/v1";
        String token = "CHANGE-ME";
        String origin = "https://CHANGE-ME.bhojpur.net/";

        ConnectionHelper conn = new ConnectionHelper();
        try {
            BhojpurClient bhojpurClient = conn.connect(uri, origin, token);
            BhojpurServer bhojpurServer = bhojpurClient.server();
            User user = bhojpurServer.getLoggedInUser().join();
            System.out.println("logged in user:" + user);

            Void result = bhojpurServer
                    .sendHeartBeat(new SendHeartBeatOptions("CHANGE-ME", false)).join();
            System.out.println("send heart beat:" + result);
        } finally {
            conn.close();
        }
    }
}
