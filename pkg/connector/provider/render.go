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
