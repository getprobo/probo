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

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

func asanaRegistration() *Registration {
	return &Registration{
		Provider:     coredata.ConnectorProviderAsana,
		DisplayName:  "Asana",
		AuthURL:      "https://app.asana.com/-/oauth_authorize",
		TokenURL:     "https://app.asana.com/-/oauth_token",
		ProbeURL:     "https://app.asana.com/api/1.0/users/me",
		OAuth2Scopes: []string{"workspaces:read", "users:read"},
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.AsanaConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read asana connector settings: %w", err)
			}

			if s.WorkspaceGID == "" {
				return nil, fmt.Errorf("cannot create asana driver: workspace_gid is required")
			}

			return drivers.NewAsanaDriver(c, s.WorkspaceGID), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.AsanaConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read asana connector settings", log.Error(err))
				return nil
			}

			return drivers.NewAsanaNameResolver(c, s.WorkspaceGID)
		},
		SetOrganizationSettings: func(c *coredata.Connector, workspaceGID string) error {
			return c.SetSettings(&coredata.AsanaConnectorSettings{WorkspaceGID: workspaceGID})
		},
	}
}
