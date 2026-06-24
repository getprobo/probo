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
	"net/url"
	"strings"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

func signozRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderSigNoz,
		DisplayName:    "SigNoz",
		SupportsAPIKey: true,
		APIKeyHeader:   "SIGNOZ-API-KEY",
		BuildProbeURL:  buildSigNozProbeURL,
		ExtraSettings: []ExtraSetting{
			{Key: "baseUrl", Label: "Base URL", Required: true},
		},
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			settings, err := coredata.ConnectorSettings[coredata.SigNozConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read signoz connector settings: %w", err)
			}

			baseURL, err := normalizeSigNozBaseURL(settings.BaseURL)
			if err != nil {
				return nil, fmt.Errorf("cannot create signoz driver: %w", err)
			}

			return drivers.NewSigNozDriver(c, baseURL), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			settings, err := coredata.ConnectorSettings[coredata.SigNozConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read signoz connector settings", log.Error(err))
				return nil
			}

			baseURL, err := normalizeSigNozBaseURL(settings.BaseURL)
			if err != nil {
				logger.ErrorCtx(ctx, "invalid signoz base url in connector settings", log.Error(err))
				return nil
			}

			return drivers.NewSigNozNameResolver(c, baseURL)
		},
	}
}

func normalizeSigNozBaseURL(raw string) (string, error) {
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
