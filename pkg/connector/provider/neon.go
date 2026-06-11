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

func neonRegistration() *Registration {
	return &Registration{
		Provider:    coredata.ConnectorProviderNeon,
		DisplayName: "Neon",
		// Neon's API authenticates with an API key (napi_...) presented
		// as Authorization: Bearer, the default APIKeyConnection scheme.
		// Neon's OAuth is partner-gated (manual application), so the
		// connector is API-key only. A personal or organization API key
		// can belong to several organizations; the operator supplies the
		// org ID (org-...) of the one to review.
		//
		SupportsAPIKey: true,
		BuildProbeURL:  buildNeonProbeURL,
		ExtraSettings: []ExtraSetting{
			{Key: "organizationId", Label: "Organization ID", Required: true},
		},
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.NeonConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read neon connector settings: %w", err)
			}

			if s.OrganizationID == "" {
				return nil, fmt.Errorf("cannot create neon driver: organization_id is required")
			}

			return drivers.NewNeonDriver(c, s.OrganizationID), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.NeonConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read neon connector settings", log.Error(err))
				return nil
			}

			return drivers.NewNeonNameResolver(c, s.OrganizationID)
		},
	}
}
