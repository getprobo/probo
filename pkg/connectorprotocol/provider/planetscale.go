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

// planetscaleRegistration wires PlanetScale. One grant can cover several
// organizations, so the user picks one.
func planetscaleRegistration() *Registration {
	return &Registration{
		Provider: coredata.ConnectorProviderPlanetScale,
		InitialAccountFunc: initialAccount(
			func(s coredata.PlanetScaleConnectorSettings) string {
				return s.Organization
			},
		),
		DisplayName:      "PlanetScale",
		DocumentationURL: accessReviewDocsURL("planetscale"),
		Endpoints: Endpoints{
			Auth:    "https://app.planetscale.com/oauth/authorize",
			Token:   "https://auth.planetscale.com/oauth/token",
			Probe:   "https://api.planetscale.com/v1/user",
			APIBase: "https://api.planetscale.com/v1",
		},
		OAuth2: &OAuth2Config{
			// PlanetScale ignores this parameter and grants the scopes configured
			// on the OAuth application, which must match this list.
			Scopes: []string{
				"read_organizations",
				"read_user",
				"read_organization",
				"read_members",
			},
			// A refresh without these scopes returns a token with none.
			RefreshScopes: []string{"read_organizations", "read_user"},
			RequiresPKCE:  true,
		},
		// No NewNameResolver: the organization name is the identifier, so the
		// initial account already carries it.
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger, ep Endpoints) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.PlanetScaleConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read planetscale connector settings: %w", err)
			}

			if s.Organization == "" {
				return nil, fmt.Errorf("cannot create planetscale driver: organization is required")
			}

			return drivers.NewPlanetScaleDriver(c, s.Organization, ep.APIBase), nil
		},
		SetOrganizationSettings: func(c *coredata.Connector, organization string) error {
			return c.SetSettings(&coredata.PlanetScaleConnectorSettings{Organization: organization})
		},
	}
}
