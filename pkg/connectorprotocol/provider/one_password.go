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

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/coredata"
)

func onePasswordRegistration() *Registration {
	return &Registration{
		Provider:           coredata.ConnectorProviderOnePassword,
		DisplayName:        "1Password",
		InitialAccountFunc: onePasswordInitialAccount,
		DocumentationURL:   accessReviewDocsURL("one-password"),
		// Endpoints is deliberately empty: every 1Password host is chosen per
		// connection. The API-key path reaches the customer's own SCIM bridge,
		// and the client-credentials path one of the regional Users API hosts
		// its Region setting names (drivers.OnePasswordUsersAPIBaseURL), which
		// also serve its token endpoint.
		Probe: probeOnePassword,
		APIKey: &APIKeyConfig{
			ExtraSettings: []ExtraSetting{
				{Key: "scimBridgeUrl", Label: "SCIM Bridge URL", Required: true},
			},
			CredentialPage: &CredentialPage{URL: "https://start.1password.com/integrations/directory"},
		},
		ClientCredentials: &ClientCredentialsConfig{
			ExtraSettings: []ExtraSetting{
				{
					Key:         "accountId",
					Label:       "Account ID",
					Required:    true,
					Description: "In the 1Password web app, open View vault items and copy the 26 characters after /app#/ in the address bar. This is not the client ID.",
				},
				{Key: "region", Label: "Region", Required: true, Control: ExtraSettingControlRegion},
			},
			CredentialPage: &CredentialPage{URL: "https://start.1password.com/integrations/directory"},
			TokenURL:       onePasswordTokenURL,
		},
		// Two settings shapes, one per connect path, because a different
		// driver sits behind each:
		//  - API key:            SCIMBridgeURL      (SCIM-bridge driver).
		//  - Client credentials: AccountID + Region (Users API driver).
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger, _ Endpoints) (drivers.Driver, error) {
			if isOnePasswordUsersAPI(conn) {
				s, err := coredata.ConnectorSettings[coredata.OnePasswordUsersAPISettings](conn)
				if err != nil {
					return nil, fmt.Errorf("cannot read 1password users api settings: %w", err)
				}

				return drivers.NewOnePasswordUsersAPIDriver(c, drivers.OnePasswordUsersAPIBaseURL(s.Region), s.AccountID), nil
			}

			s, err := coredata.ConnectorSettings[coredata.OnePasswordConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read 1password connector settings: %w", err)
			}

			if s.SCIMBridgeURL == "" {
				return nil, fmt.Errorf("cannot create 1password driver: scim_bridge_url is required")
			}

			return drivers.NewOnePasswordDriver(c, s.SCIMBridgeURL), nil
		},
	}
}

// isOnePasswordUsersAPI reports whether the connector is a client-credentials
// one, served by the Users API. Everything else is the API-key connection to a
// SCIM bridge, whose *APIKeyConnection makes GrantType() return "". 1Password
// declares no authorization endpoint, so there is no third kind.
func isOnePasswordUsersAPI(c *coredata.Connector) bool {
	return c.GrantType() == string(connectorprotocol.OAuth2GrantTypeClientCredentials)
}

func onePasswordTokenURL(c *coredata.Connector) (string, error) {
	s, err := coredata.ConnectorSettings[coredata.OnePasswordUsersAPISettings](c)
	if err != nil {
		return "", fmt.Errorf("cannot read 1password users api settings: %w", err)
	}

	return drivers.OnePasswordTokenURL(s.Region), nil
}

func onePasswordInitialAccount(c *coredata.Connector) (string, string, error) {
	users, err := coredata.ConnectorSettings[coredata.OnePasswordUsersAPISettings](c)
	if err != nil {
		return "", "", fmt.Errorf("cannot read connector settings: %w", err)
	}

	if users.AccountID != "" {
		return users.AccountID, users.AccountID, nil
	}

	return initialAccount(
		func(s coredata.OnePasswordConnectorSettings) string {
			return s.SCIMBridgeURL
		},
	)(c)
}
