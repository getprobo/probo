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
	"net/http"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

func pagerdutyRegistration() *Registration {
	// PagerDuty Scoped OAuth requires PKCE (RFC 7636). The customer
	// subdomain surfaces as a callback query parameter (or
	// occasionally in the token response body) and is persisted on
	// PagerDutyConnectorSettings by the OAuth callback handler.
	return &Registration{
		Provider:     coredata.ConnectorProviderPagerDuty,
		DisplayName:  "PagerDuty",
		AuthURL:      "https://identity.pagerduty.com/oauth/authorize",
		TokenURL:     "https://identity.pagerduty.com/oauth/token",
		ProbeURL:     "https://api.pagerduty.com/users/me",
		OAuth2Scopes: []string{"users.read"},
		RequiresPKCE: true,
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			// PagerDuty's REST API uses the regional api.pagerduty.com host;
			// the driver does not consume the per-tenant subdomain.
			return drivers.NewPagerDutyDriver(c), nil
		},
		NewNameResolver: func(ctx context.Context, _ *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.PagerDutyConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read pagerduty connector settings", log.Error(err))
				return nil
			}

			return drivers.NewPagerDutyNameResolver(s.Subdomain)
		},
	}
}
