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

func googleAnalyticsRegistration() *Registration {
	return &Registration{
		Provider:    coredata.ConnectorProviderGoogleAnalytics,
		DisplayName: "Google Analytics",
		AuthURL:     "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:    "https://oauth2.googleapis.com/token",
		ExtraAuthParams: map[string]string{
			"access_type": "offline",
			"prompt":      "consent",
		},
		SupportsIncrementalAuth: true,
		// analytics.readonly is required to LIST accounts and properties (the
		// picker and the probe); analytics.manage.users.readonly is required to
		// read the access bindings. The manage.users scope alone cannot list
		// accounts (it returns 403), so both are requested.
		OAuth2Scopes: []string{
			"https://www.googleapis.com/auth/analytics.readonly",
			"https://www.googleapis.com/auth/analytics.manage.users.readonly",
		},
		ProbeURL: "https://analyticsadmin.googleapis.com/v1alpha/accounts?pageSize=1",
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.GoogleAnalyticsConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read google analytics connector settings: %w", err)
			}

			if s.AccountID == "" {
				return nil, fmt.Errorf("cannot create google analytics driver: account_id is required")
			}

			return drivers.NewGoogleAnalyticsDriver(c, s.AccountID), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.GoogleAnalyticsConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read google analytics connector settings", log.Error(err))

				return nil
			}

			return drivers.NewGoogleAnalyticsNameResolver(c, s.AccountID)
		},
		SetOrganizationSettings: func(c *coredata.Connector, accountID string) error {
			return c.SetSettings(&coredata.GoogleAnalyticsConnectorSettings{AccountID: accountID})
		},
	}
}
