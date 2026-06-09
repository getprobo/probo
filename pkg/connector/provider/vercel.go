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
	"net/url"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

func vercelRegistration() *Registration {
	// Vercel's authorization URL embeds the operator's integration slug
	// as a path segment; the operator supplies it via the
	// `integration-slug` config field. BuildAuthURL constructs the URL
	// with net/url so the slug is escaped. Vercel does not use OAuth
	// scopes — capabilities are pinned on the integration registration
	// in the Vercel dashboard.
	return &Registration{
		Provider:    coredata.ConnectorProviderVercel,
		DisplayName: "Vercel",
		TokenURL:    "https://api.vercel.com/v2/oauth/access_token",
		ProbeURL:    "https://api.vercel.com/v2/user",
		BuildAuthURL: func(slug string) (string, error) {
			u, err := url.JoinPath("https://vercel.com/integrations", url.PathEscape(slug), "new")
			if err != nil {
				return "", fmt.Errorf("cannot build vercel auth URL: %w", err)
			}

			return u, nil
		},
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.VercelConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read vercel connector settings: %w", err)
			}

			if s.TeamID == "" {
				return nil, fmt.Errorf("cannot create vercel driver: team_id is required")
			}

			return drivers.NewVercelDriver(c, s.TeamID), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.VercelConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read vercel connector settings", log.Error(err))
				return nil
			}

			return drivers.NewVercelNameResolver(c, s.TeamID)
		},
	}
}
