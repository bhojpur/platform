// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package net.bhojpur.platform.api;

import java.util.concurrent.CompletableFuture;

import org.eclipse.lsp4j.jsonrpc.services.JsonRequest;

import net.bhojpur.platform.api.entities.SendHeartBeatOptions;
import net.bhojpur.platform.api.entities.User;

public interface BhojpurServer {
    @JsonRequest
    CompletableFuture<User> getLoggedInUser();

    @JsonRequest
    CompletableFuture<Void> sendHeartBeat(SendHeartBeatOptions options);
}
