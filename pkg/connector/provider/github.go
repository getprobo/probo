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

func githubRegistration() *Registration {
	return &Registration{
		Provider:     coredata.ConnectorProviderGitHub,
		DisplayName:  "GitHub",
		AuthURL:      "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		ProbeURL:     "https://api.github.com/user",
		OAuth2Scopes: []string{"read:org"},
		DiscoveryOAuth2Scopes: []string{
			"repo",
			"security_events",
			"read:enterprise",
			"read:audit_log",
		},
		SupportsIncrementalAuth: true,
		SupportsAPIKey:          true,
		ExtraSettings: []ExtraSetting{
			{Key: "organization", Label: "Organization", Required: true},
		},
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.GitHubConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read github connector settings: %w", err)
			}

			if s.Organization == "" {
				return nil, fmt.Errorf("cannot create github driver: organization is required")
			}

			return drivers.NewGitHubDriver(c, s.Organization, logger.Named("github")), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.GitHubConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read github connector settings", log.Error(err))
				return nil
			}

			return drivers.NewGitHubNameResolver(c, s.Organization)
		},
		SetOrganizationSettings: func(c *coredata.Connector, org string) error {
			return c.SetSettings(&coredata.GitHubConnectorSettings{Organization: org})
		},
	}
}
