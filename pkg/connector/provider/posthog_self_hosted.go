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

// posthogSelfHostedRegistration is customer-hosted PostHog: API-key plus an
// operator-supplied instance URL (Metabase/Grafana style). It shares the
// PostHog driver and name resolver, pointed at the instance's BaseURL. OAuth
// is deliberately not offered here — a single static authorization URL cannot
// serve arbitrary per-customer instances, so self-hosted OAuth is a separate
// future effort. Cloud PostHog (POSTHOG) owns the OAuth path.
func posthogSelfHostedRegistration() *Registration {
	return &Registration{
		Provider:    coredata.ConnectorProviderPostHogSelfHosted,
		DisplayName: "PostHog (Self-Hosted)",

		SupportsAPIKey: true,
		ExtraSettings: []ExtraSetting{
			{Key: "instanceUrl", Label: "Instance URL", Required: true},
		},

		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.PostHogConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read posthog self-hosted connector settings: %w", err)
			}

			// Never fall back to the cloud gateway for a self-hosted
			// connector — the instance URL is required at creation time.
			if s.BaseURL == "" {
				return nil, fmt.Errorf("cannot create posthog self-hosted driver: instance URL is required")
			}

			return drivers.NewPostHogDriver(c, s.BaseURL), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.PostHogConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read posthog self-hosted connector settings", log.Error(err))
				return nil
			}

			if s.BaseURL == "" {
				return nil
			}

			return drivers.NewPostHogNameResolver(c, s.BaseURL)
		},
	}
}
