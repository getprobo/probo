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
