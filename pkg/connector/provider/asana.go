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
