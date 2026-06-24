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
