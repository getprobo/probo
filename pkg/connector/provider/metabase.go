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

func metabaseRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderMetabase,
		DisplayName:    "Metabase",
		SupportsAPIKey: true,
		APIKeyHeader:   "x-api-key",
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
