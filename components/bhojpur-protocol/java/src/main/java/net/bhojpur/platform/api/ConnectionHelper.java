// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package net.bhojpur.platform.api;

import java.io.IOException;
import java.net.URI;
import java.util.Arrays;
import java.util.Collection;
import java.util.List;
import java.util.Map;

import javax.websocket.ClientEndpointConfig;
import javax.websocket.ContainerProvider;
import javax.websocket.DeploymentException;
import javax.websocket.Session;
import javax.websocket.WebSocketContainer;

import org.eclipse.lsp4j.jsonrpc.Launcher;
import org.eclipse.lsp4j.websocket.WebSocketEndpoint;

public class ConnectionHelper {

    private Session session;

    public BhojpurClient connect(final String uri, final String origin, final String token)
            throws DeploymentException, IOException {
        final BhojpurClientImpl bhojpurClient = new BhojpurClientImpl();

        final WebSocketEndpoint<BhojpurServer> webSocketEndpoint = new WebSocketEndpoint<BhojpurServer>() {
            @Override
            protected void configure(final Launcher.Builder<BhojpurServer> builder) {
                builder.setLocalService(bhojpurClient).setRemoteInterface(BhojpurServer.class);
            }

            @Override
            protected void connect(final Collection<Object> localServices, final BhojpurServer remoteProxy) {
                localServices.forEach(s -> ((BhojpurClient) s).connect(remoteProxy));
            }
        };

        final ClientEndpointConfig.Configurator configurator = new ClientEndpointConfig.Configurator() {
            @Override
            public void beforeRequest(final Map<String, List<String>> headers) {
                headers.put("Origin", Arrays.asList(origin));
                headers.put("Authorization", Arrays.asList("Bearer " + token));
            }
        };
        final ClientEndpointConfig clientEndpointConfig = ClientEndpointConfig.Builder.create()
                .configurator(configurator).build();
        final WebSocketContainer webSocketContainer = ContainerProvider.getWebSocketContainer();
        this.session = webSocketContainer.connectToServer(webSocketEndpoint, clientEndpointConfig, URI.create(uri));
        return bhojpurClient;
    }

    public void close() throws IOException {
        if (this.session != null && this.session.isOpen()) {
            this.session.close();
        }
    }
}
