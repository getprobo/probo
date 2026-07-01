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

func scalewayRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderScaleway,
		DisplayName:    "Scaleway",
		SupportsAPIKey: true,
		// Scaleway authenticates with the secret key in the X-Auth-Token header
		// rather than Authorization: Bearer. APIKeyHeader makes the
		// APIKeyConnection send that header and omit Authorization. The key is
		// bound to one Organization, but GET /iam/v1alpha1/users requires the
		// organization_id explicitly, so it is captured via ExtraSettings rather
		// than discovered — hence no picker and a BuildProbeURL.
		APIKeyHeader: "X-Auth-Token",
		ExtraSettings: []ExtraSetting{
			{Key: "organizationId", Label: "Organization ID", Required: true},
		},
		BuildProbeURL: buildScalewayProbeURL,
		// No NewNameResolver: Scaleway exposes no read-only endpoint that maps
		// an Organization UUID to its display name, so the source keeps its
		// generic name.
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.ScalewayConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read scaleway connector settings: %w", err)
			}

			if s.OrganizationID == "" {
				return nil, fmt.Errorf("cannot create scaleway driver: organization_id is required")
			}

			return drivers.NewScalewayDriver(c, s.OrganizationID), nil
		},
	}
}
