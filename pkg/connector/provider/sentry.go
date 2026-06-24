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

func sentryRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderSentry,
		DisplayName:    "Sentry",
		AuthURL:        "https://sentry.io/oauth/authorize/",
		TokenURL:       "https://sentry.io/oauth/token/",
		ProbeURL:       "https://sentry.io/api/0/organizations/",
		OAuth2Scopes:   []string{"org:read", "member:read"},
		SupportsAPIKey: true,
		ExtraSettings: []ExtraSetting{
			{Key: "organizationSlug", Label: "Organization Slug", Required: true},
		},
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.SentryConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read sentry connector settings: %w", err)
			}

			// OrganizationSlug may be empty for OAuth connections; the driver auto-discovers it.
			return drivers.NewSentryDriver(c, s.OrganizationSlug), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.SentryConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read sentry connector settings", log.Error(err))
				return nil
			}

			return drivers.NewSentryNameResolver(c, s.OrganizationSlug)
		},
		SetOrganizationSettings: func(c *coredata.Connector, slug string) error {
			return c.SetSettings(&coredata.SentryConnectorSettings{OrganizationSlug: slug})
		},
	}
}
