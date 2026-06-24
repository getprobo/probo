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

func herokuRegistration() *Registration {
	return &Registration{
		Provider:    coredata.ConnectorProviderHeroku,
		DisplayName: "Heroku",
		AuthURL:     "https://id.heroku.com/oauth/authorize",
		TokenURL:    "https://id.heroku.com/oauth/token",
		// Heroku requires the versioned Accept header; a plain ProbeURL GET
		// (Accept: application/json) returns 400 and would read as connected,
		// so probe via a closure that sends application/vnd.heroku+json.
		Probe:        probeHeroku,
		OAuth2Scopes: []string{"read"},
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.HerokuConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read heroku connector settings: %w", err)
			}

			// TeamID may be empty or the personal-account slug for a solo
			// Heroku account (no Team); the driver runs in personal mode
			// (app owner + collaborators) in that case.
			return drivers.NewHerokuDriver(c, s.TeamID), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.HerokuConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read heroku connector settings", log.Error(err))
				return nil
			}

			return drivers.NewHerokuNameResolver(c, s.TeamID)
		},
		SetOrganizationSettings: func(c *coredata.Connector, teamID string) error {
			return c.SetSettings(&coredata.HerokuConnectorSettings{TeamID: teamID})
		},
	}
}
