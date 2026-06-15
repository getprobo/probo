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
