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
		MarshalSettings: func(in *SettingsInput) (json.RawMessage, error) {
			if in == nil || in.SentryOrganizationSlug == nil || *in.SentryOrganizationSlug == "" {
				return nil, fmt.Errorf("cannot create sentry connector: sentryOrganizationSlug is required")
			}

			return json.Marshal(&coredata.SentryConnectorSettings{OrganizationSlug: *in.SentryOrganizationSlug})
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
