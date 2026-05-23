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
	"fmt"
	"net/http"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

func netlifyRegistration() *Registration {
	// Netlify OAuth flow has no scope granularity, so OAuth2Scopes is empty.
	return &Registration{
		Provider:    coredata.ConnectorProviderNetlify,
		DisplayName: "Netlify",
		AuthURL:     "https://app.netlify.com/authorize",
		TokenURL:    "https://api.netlify.com/oauth/token",
		ProbeURL:    "https://api.netlify.com/api/v1/user",
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.NetlifyConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read netlify connector settings: %w", err)
			}

			if s.AccountSlug == "" {
				return nil, fmt.Errorf("cannot create netlify driver: account_slug is required")
			}

			return drivers.NewNetlifyDriver(c, s.AccountSlug), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.NetlifyConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read netlify connector settings", log.Error(err))
				return nil
			}

			return drivers.NewNetlifyNameResolver(c, s.AccountSlug)
		},
		SetOrganizationSettings: func(c *coredata.Connector, accountSlug string) error {
			return c.SetSettings(&coredata.NetlifyConnectorSettings{AccountSlug: accountSlug})
		},
	}
}
