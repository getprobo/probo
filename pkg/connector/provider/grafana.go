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
	"net/url"
	"strings"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

func grafanaRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderGrafana,
		DisplayName:    "Grafana",
		SupportsAPIKey: true,
		ExtraSettings: []ExtraSetting{
			{Key: "baseUrl", Label: "Base URL", Required: true},
		},
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.GrafanaConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read grafana connector settings: %w", err)
			}

			baseURL, err := normalizeGrafanaBaseURL(s.BaseURL)
			if err != nil {
				return nil, fmt.Errorf("cannot create grafana driver: %w", err)
			}

			return drivers.NewGrafanaDriver(c, baseURL), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.GrafanaConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read grafana connector settings", log.Error(err))
				return nil
			}

			baseURL, err := normalizeGrafanaBaseURL(s.BaseURL)
			if err != nil {
				logger.ErrorCtx(ctx, "invalid grafana base url in connector settings", log.Error(err))
				return nil
			}

			return drivers.NewGrafanaNameResolver(c, baseURL)
		},
	}
}

func normalizeGrafanaBaseURL(raw string) (string, error) {
	baseURL := strings.TrimSpace(raw)
	if baseURL == "" {
		return "", fmt.Errorf("base_url is required")
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("base_url must be a valid URL: %w", err)
	}

	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", fmt.Errorf("base_url must be an http(s) URL")
	}

	u.Path = strings.TrimRight(u.Path, "/")
	u.RawQuery = ""
	u.Fragment = ""

	return u.String(), nil
}
