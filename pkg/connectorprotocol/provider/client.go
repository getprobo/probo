// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package provider

import (
	"context"
	"fmt"
	"net/http"

	"go.probo.inc/probo/pkg/cloud"
	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/identityfederation"
)

// OpenSession opens authenticated access to a workload identity connector.
// An empty accountID, or the initial account's own id, keeps that account.
// Every consumer must go through this helper so an organization install
// cannot leave one of them on the hub.
func OpenSession(
	ctx context.Context,
	federation *identityfederation.Issuer,
	registries *Registry,
	dbConnector *coredata.Connector,
	accountID string,
) (cloud.Session, error) {
	reg, ok := registries.Get(dbConnector.Provider)
	if !ok || reg.WorkloadIdentity == nil {
		return nil, fmt.Errorf(
			"cannot open cloud session: provider %s offers no workload identity path",
			dbConnector.Provider,
		)
	}

	// The initial account's own id is empty too: AWS would assume the member
	// role instead of the configured RoleARN.
	if accountID != "" {
		initialID, _, err := reg.ResolveInitialAccount(dbConnector)
		if err != nil {
			return nil, fmt.Errorf("cannot resolve initial account: %w", err)
		}

		if accountID == initialID {
			accountID = ""
		}
	}

	session, err := reg.WorkloadIdentity.NewSession(ctx, federation, dbConnector, accountID)
	if err != nil {
		return nil, fmt.Errorf("cannot open cloud session for %s connector: %w", dbConnector.Provider, err)
	}

	return session, nil
}

func oauthClient(
	ctx context.Context,
	connectorRegistry *connectorprotocol.Registry,
	conn *connectorprotocol.OAuth2Connection,
	prvdr coredata.ConnectorProvider,
) (*http.Client, error) {
	refreshCfg := connectorRegistry.GetOAuth2RefreshConfig(string(prvdr))
	if refreshCfg != nil {
		return conn.RefreshableClient(ctx, *refreshCfg)
	}

	return conn.Client(ctx)
}

// HTTPClient returns an HTTP client for the given connection. For OAuth2
// connections it refreshes the token. For API-key connections it overlays a
// Probo-held key onto a copy when the provider is managed, so the loaded row
// is never mutated.
func HTTPClient(
	ctx context.Context,
	connectorRegistry *connectorprotocol.Registry,
	registries *Registry,
	prvdr coredata.ConnectorProvider,
	conn connectorprotocol.HTTPConnection,
) (*http.Client, error) {
	if err := connectorRegistry.ConfigureConnection(string(prvdr), conn); err != nil {
		return nil, err
	}

	if oauth2Conn := connectorprotocol.OAuth2ConnectionOf(conn); oauth2Conn != nil {
		return oauthClient(ctx, connectorRegistry, oauth2Conn, prvdr)
	}

	apiKeyConn, ok := conn.(*connectorprotocol.APIKeyConnection)
	if !ok {
		return conn.Client(ctx)
	}

	key, err := registries.APIKeyFor(prvdr, apiKeyConn)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve API key for %s connector: %w", prvdr, err)
	}

	prepared := *apiKeyConn
	prepared.APIKey = key

	return prepared.Client(ctx)
}
