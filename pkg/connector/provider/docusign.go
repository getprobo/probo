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
	"go.probo.inc/probo/pkg/coredata"
)

func docusignRegistration() *Registration {
	return &Registration{
		Provider:    coredata.ConnectorProviderDocuSign,
		DisplayName: "DocuSign",
		// APIBase is deliberately empty. account.docusign.com above is the
		// identity host only; the eSignature data host is per account,
		// discovered at runtime from the `base_uri` field of the userinfo
		// response (drivers.DocuSignDriver.discoverBaseURI). Recording the
		// identity host as APIBase would look right and be wrong: an override
		// would move the userinfo call while every data call still went to the
		// account's real data center, splitting one connector across two
		// environments. Identity records that host instead: NewDriver,
		// NewNameResolver and the org picker all resolve accounts from it, so
		// an override reaches every one of them, not just the health check.
		// Probe and Identity are the same URL here and Register enforces that
		// their hosts stay in agreement.
		Endpoints: Endpoints{
			Auth:     "https://account.docusign.com/oauth/auth",
			Token:    "https://account.docusign.com/oauth/token",
			Probe:    "https://account.docusign.com/oauth/userinfo",
			Identity: "https://account.docusign.com/oauth/userinfo",
		},
		TokenEndpointAuth: "basic-form",
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
		NewDriver: HTTP(func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger, ep Endpoints) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.DocuSignConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read docusign connector settings: %w", err)
			}

			if s.AccountID == "" {
				return nil, fmt.Errorf("cannot create docusign driver: account_id is required")
			}

			return drivers.NewDocuSignDriver(c, s.AccountID, ep.Identity), nil
		}),
		NewNameResolver: HTTPNameResolver(func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger, ep Endpoints) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.DocuSignConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read docusign connector settings", log.Error(err))
				return nil
			}

			return drivers.NewDocuSignNameResolver(c, s.AccountID, ep.Identity)
		}),
		ListOrganizations: HTTPOrganizations(drivers.ListDocuSignOrganizations),
		SetOrganizationSettings: func(c *coredata.Connector, accountID string) error {
			return c.SetSettings(&coredata.DocuSignConnectorSettings{AccountID: accountID})
		},
	}
}
