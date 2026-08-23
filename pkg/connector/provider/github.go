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

func githubRegistration() *Registration {
	return &Registration{
		Provider:    coredata.ConnectorProviderGitHub,
		DisplayName: "GitHub",
		Endpoints: Endpoints{
			Auth:  "https://github.com/login/oauth/authorize",
			Token: "https://github.com/login/oauth/access_token",
			Probe: "https://api.github.com/user",
			// The REST API carries no version path segment; the version is
			// negotiated through the Accept header instead.
			APIBase: "https://api.github.com",
		},
		// read:audit_log is Enterprise Cloud only, so it cannot be a
		// universal requirement for GitHub.com organizations. The driver
		// treats audit-log access as optional and leaves LastLogin nil when
		// GitHub rejects that request.
		OAuth2Scopes:   []string{"read:org"},
		SupportsAPIKey: true,
		APIKeyExtraSettings: []ExtraSetting{
			{Key: "organization", Label: "Organization", Required: true},
		},
		Probe: probeGitHub,
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger, ep Endpoints) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.GitHubConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read github connector settings: %w", err)
			}

			if s.Organization == "" {
				return nil, fmt.Errorf("cannot create github driver: organization is required")
			}

			return drivers.NewGitHubDriver(c, s.Organization, logger.Named("github"), ep.APIBase), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger, ep Endpoints) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.GitHubConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read github connector settings", log.Error(err))
				return nil
			}

			return drivers.NewGitHubNameResolver(c, s.Organization, ep.APIBase)
		},
		SetOrganizationSettings: func(c *coredata.Connector, org string) error {
			return c.SetSettings(&coredata.GitHubConnectorSettings{Organization: org})
		},
	}
}
