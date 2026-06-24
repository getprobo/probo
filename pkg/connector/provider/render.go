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

// renderRegistration wires Render as an API-key access-review connector.
// Render exposes no partner OAuth program, so the customer supplies a
// read-scoped API key plus their Workspace ID (Render's owner ID). The key
// authenticates with the default Authorization: Bearer scheme, so no
// APIKeyAuthScheme override is set. There is no picker — the workspace is
// captured up front via ExtraSettings — so SetOrganizationSettings is omitted.
func renderRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderRender,
		DisplayName:    "Render",
		SupportsAPIKey: true,
		BuildProbeURL:  buildRenderProbeURL,
		ExtraSettings: []ExtraSetting{
			{Key: "workspaceId", Label: "Workspace ID", Required: true},
		},
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.RenderConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read render connector settings: %w", err)
			}

			if s.OwnerID == "" {
				return nil, fmt.Errorf("cannot create render driver: owner_id is required")
			}

			return drivers.NewRenderDriver(c, s.OwnerID), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.RenderConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read render connector settings", log.Error(err))
				return nil
			}

			return drivers.NewRenderNameResolver(c, s.OwnerID)
		},
	}
}
