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

func metabaseRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderMetabase,
		DisplayName:    "Metabase",
		SupportsAPIKey: true,
		APIKeyHeader:   "x-api-key",
		BuildProbeURL:  buildMetabaseProbeURL,
		ExtraSettings: []ExtraSetting{
			{Key: "instanceUrl", Label: "Instance URL", Required: true},
		},
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			settings, err := coredata.ConnectorSettings[coredata.MetabaseConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read metabase connector settings: %w", err)
			}

			instanceURL := strings.TrimSpace(settings.InstanceURL)
			if instanceURL == "" {
				return nil, fmt.Errorf("cannot create metabase driver: instance_url is required")
			}

			if err := validateMetabaseInstanceURL(instanceURL); err != nil {
				return nil, err
			}

			return drivers.NewMetabaseDriver(c, instanceURL), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			settings, err := coredata.ConnectorSettings[coredata.MetabaseConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read metabase connector settings", log.Error(err))
				return nil
			}

			instanceURL := strings.TrimSpace(settings.InstanceURL)
			if instanceURL == "" {
				logger.ErrorCtx(ctx, "missing metabase instance url in connector settings")
				return nil
			}

			if err := validateMetabaseInstanceURL(instanceURL); err != nil {
				logger.ErrorCtx(ctx, "invalid metabase instance url in connector settings", log.Error(err))
				return nil
			}

			return drivers.NewMetabaseNameResolver(c, instanceURL)
		},
	}
}

func validateMetabaseInstanceURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("cannot create metabase driver: instance_url is invalid: %w", err)
	}

	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("cannot create metabase driver: instance_url must be an http(s) URL")
	}

	return nil
}
