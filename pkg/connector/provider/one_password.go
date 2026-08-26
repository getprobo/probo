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
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
)

func onePasswordRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderOnePassword,
		DisplayName:      "1Password",
		DocumentationURL: accessReviewDocsURL("one-password"),
		// APIBase is deliberately empty: 1Password has no single data host to
		// name. Four host families live under this one registration —
		// the per-connection SCIM bridge URL (customer-hosted, from
		// APIKeyExtraSettings), the three regional Users API hosts resolved
		// from the Region setting (drivers.onePasswordBaseURL), and the
		// events.1password.com probe below, which is yet another host and is
		// therefore not the root any driver joins onto. A single APIBase would
		// have to be wrong for at least two of them, and Register's
		// Probe-host-matches-APIBase check would reject the honest pairing.
		Endpoints: Endpoints{
			Probe: "https://events.1password.com/api/v1/auditevents",
		},
		APIKey: &APIKeyConfig{
			ExtraSettings: []ExtraSetting{
				{Key: "scimBridgeUrl", Label: "SCIM Bridge URL", Required: true},
			},
		},
		ClientCredentials: &ClientCredentialsConfig{
			ExtraSettings: []ExtraSetting{
				{Key: "accountId", Label: "Account ID", Required: true},
				{Key: "region", Label: "Region", Required: true},
			},
		},
		// Two settings shapes, one per connect path, because a different
		// driver sits behind each:
		//  - API key:            SCIMBridgeURL      (SCIM-bridge driver).
		//  - Client credentials: AccountID + Region (Users API driver).
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger, _ Endpoints) (drivers.Driver, error) {
			// The client-credentials grant uses the Users API driver.
			// Everything else is the API-key connection, whose
			// *APIKeyConnection makes GrantType() return "": it uses the
			// SCIM-bridge driver. 1Password declares no AuthURL/TokenURL, so
			// the authorization-code path is unreachable.
			if conn.GrantType() == string(connector.OAuth2GrantTypeClientCredentials) {
				s, err := coredata.ConnectorSettings[coredata.OnePasswordUsersAPISettings](conn)
				if err != nil {
					return nil, fmt.Errorf("cannot read 1password users api settings: %w", err)
				}

				return drivers.NewOnePasswordUsersAPIDriver(c, s.AccountID, s.Region), nil
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
