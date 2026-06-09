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
