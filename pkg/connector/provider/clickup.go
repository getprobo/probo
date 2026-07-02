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

func clickupRegistration() *Registration {
	// ClickUp OAuth flow has no scope granularity, so OAuth2Scopes is empty.
	return &Registration{
		Provider:    coredata.ConnectorProviderClickUp,
		DisplayName: "ClickUp",
		AuthURL:     "https://app.clickup.com/api",
		TokenURL:    "https://api.clickup.com/api/v2/oauth/token",
		ProbeURL:    "https://api.clickup.com/api/v2/user",
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.ClickUpConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read clickup connector settings: %w", err)
			}

			if s.TeamID == "" {
				return nil, fmt.Errorf("cannot create clickup driver: team_id is required")
			}

			return drivers.NewClickUpDriver(c, s.TeamID), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.ClickUpConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read clickup connector settings", log.Error(err))
				return nil
			}

			return drivers.NewClickUpNameResolver(c, s.TeamID)
		},
		SetOrganizationSettings: func(c *coredata.Connector, teamID string) error {
			return c.SetSettings(&coredata.ClickUpConnectorSettings{TeamID: teamID})
		},
	}
}
