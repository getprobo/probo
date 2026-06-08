// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"fmt"
	"net/http"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

func docusignRegistration() *Registration {
	return &Registration{
		Provider:          coredata.ConnectorProviderDocuSign,
		DisplayName:       "DocuSign",
		AuthURL:           "https://account.docusign.com/oauth/auth",
		TokenURL:          "https://account.docusign.com/oauth/token",
		TokenEndpointAuth: "basic-form",
		ProbeURL:          "https://account.docusign.com/oauth/userinfo",
		// signature grants the eSignature REST API (the userinfo probe and
		// the account users list). extended rolls the 30-day refresh-token
		// window on every refresh so the connection survives long-term — the
		// review engine persists the rotated token on each poll. Without it
		// the refresh token hard-expires 30 days after the initial consent.
		OAuth2Scopes: []string{"signature", "extended"},
		// DocuSign enables PKCE (S256) on the integration key. The confidential
		// authorization-code grant still authenticates the token exchange with
		// Basic auth (basic-form); PKCE rides along as the documented hardening
		// layer, replaying the verifier in the token request body.
		RequiresPKCE: true,
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.DocuSignConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read docusign connector settings: %w", err)
			}

			if s.AccountID == "" {
				return nil, fmt.Errorf("cannot create docusign driver: account_id is required")
			}

			return drivers.NewDocuSignDriver(c, s.AccountID), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.DocuSignConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read docusign connector settings", log.Error(err))
				return nil
			}

			return drivers.NewDocuSignNameResolver(c, s.AccountID)
		},
		SetOrganizationSettings: func(c *coredata.Connector, accountID string) error {
			return c.SetSettings(&coredata.DocuSignConnectorSettings{AccountID: accountID})
		},
	}
}
