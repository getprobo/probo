// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
)

func onePasswordRegistration() *Registration {
	return &Registration{
		Provider:                  coredata.ConnectorProviderOnePassword,
		DisplayName:               "1Password",
		ProbeURL:                  "https://events.1password.com/api/v1/auditevents",
		SupportsAPIKey:            true,
		SupportsClientCredentials: true,
		ExtraSettings: []ExtraSetting{
			{Key: "accountId", Label: "Account ID", Required: true},
			{Key: "region", Label: "Region", Required: true},
		},
		// 1Password has two settings shapes selected by protocol:
		//  - Client-credentials: AccountID + Region (Users API driver).
		//  - API key:            SCIMBridgeURL      (SCIM-bridge driver).
		// MarshalSettings picks the shape based on which input fields
		// are populated. The resolvers ensure that only one path is
		// possible for any given request.
		MarshalSettings: func(in *SettingsInput) (json.RawMessage, error) {
			if in == nil {
				return nil, nil
			}

			if in.OnePasswordAccountID != nil && in.OnePasswordRegion != nil {
				if *in.OnePasswordAccountID == "" || *in.OnePasswordRegion == "" {
					return nil, fmt.Errorf("cannot create 1password connector: onePasswordAccountId and onePasswordRegion must be non-empty")
				}

				return json.Marshal(&coredata.OnePasswordUsersAPISettings{
					AccountID: *in.OnePasswordAccountID,
					Region:    *in.OnePasswordRegion,
				})
			}

			if in.OnePasswordSCIMBridgeURL != nil && *in.OnePasswordSCIMBridgeURL != "" {
				u, err := url.Parse(*in.OnePasswordSCIMBridgeURL)
				if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
					return nil, fmt.Errorf("cannot create 1password connector: onePasswordScimBridgeURL must be an http(s) URL")
				}

				return json.Marshal(&coredata.OnePasswordConnectorSettings{
					SCIMBridgeURL: *in.OnePasswordSCIMBridgeURL,
				})
			}

			return nil, nil
		},
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			// Client credentials grant uses the Users API driver; the
			// authorization-code grant uses the SCIM-bridge driver.
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
