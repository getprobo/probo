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
	"go.probo.inc/probo/pkg/coredata"
)

func herokuRegistration() *Registration {
	return &Registration{
		Provider:    coredata.ConnectorProviderHeroku,
		DisplayName: "Heroku",
		Endpoints: Endpoints{
			Auth:  "https://id.heroku.com/oauth/authorize",
			Token: "https://id.heroku.com/oauth/token",
			// The Platform API lives on api.heroku.com, a different host
			// from the id.heroku.com OAuth endpoints above; it carries no
			// version segment (the version is negotiated through the Accept
			// media type instead).
			APIBase: "https://api.heroku.com",
		},
		OAuth2: &OAuth2Spec{
			Scopes: []string{"read"},
		},
		// Heroku requires the versioned Accept header; a plain ProbeURL GET
		// (Accept: application/json) returns 400 and would read as connected,
		// so probe via a closure that sends application/vnd.heroku+json.
		Probe: ProbeOver(probeHeroku),
		SetOrganizationSettings: func(c *coredata.Connector, teamID string) error {
			return c.SetSettings(&coredata.HerokuConnectorSettings{TeamID: teamID})
		},
	}
}
