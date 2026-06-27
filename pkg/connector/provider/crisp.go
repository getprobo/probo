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

func crispRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderCrisp,
		DisplayName:    "Crisp",
		SupportsAPIKey: true,
		// Crisp authenticates with a plugin token presented as HTTP Basic, the
		// credential being the verbatim "identifier:key" pair.
		// APIKeyBasicAuthUserPass base64-encodes it (the empty-password
		// APIKeyBasicAuth cannot carry the key). A plugin token can serve
		// several websites, so the reviewed website is captured via
		// ExtraSettings. Every request also needs the non-auth X-Crisp-Tier
		// header (set by the driver/probe/name resolver), so the probe is a
		// custom closure.
		APIKeyBasicAuthUserPass: true,
		ExtraSettings: []ExtraSetting{
			{Key: "websiteId", Label: "Website ID", Required: true},
		},
		Probe: probeCrisp,
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.CrispConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read crisp connector settings: %w", err)
			}

			if s.WebsiteID == "" {
				return nil, fmt.Errorf("cannot create crisp driver: website_id is required")
			}

			return drivers.NewCrispDriver(c, s.WebsiteID), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.CrispConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read crisp connector settings", log.Error(err))

				return nil
			}

			return drivers.NewCrispNameResolver(c, s.WebsiteID)
		},
	}
}
