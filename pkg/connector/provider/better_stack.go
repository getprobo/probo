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
	"strings"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

// betterStackRegistration wires the Better Stack Uptime access-review
// connector. Better Stack has no third-party OAuth app for listing team
// members (its OAuth is an end-user MCP sign-in), so the connector is
// API-key only: the operator supplies a Bearer API token plus the team
// name that scopes the team-members listing.
func betterStackRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderBetterStack,
		DisplayName:    "Better Stack",
		SupportsAPIKey: true,
		ProbeURL:       "https://betterstack.com/api/v2/team-members",
		ExtraSettings: []ExtraSetting{
			{Key: "teamName", Label: "Team Name", Required: true},
		},
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.BetterStackConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read better stack connector settings: %w", err)
			}

			teamName := strings.TrimSpace(s.TeamName)
			if teamName == "" {
				return nil, fmt.Errorf("cannot create better stack driver: team_name is required")
			}

			return drivers.NewBetterStackDriver(c, teamName), nil
		},
		NewNameResolver: func(ctx context.Context, _ *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.BetterStackConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read better stack connector settings", log.Error(err))
				return nil
			}

			return drivers.NewBetterStackNameResolver(strings.TrimSpace(s.TeamName))
		},
	}
}
