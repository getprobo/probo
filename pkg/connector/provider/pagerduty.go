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
