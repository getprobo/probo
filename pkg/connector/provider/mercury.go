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
	"net/http"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

func mercuryRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderMercury,
		DisplayName:    "Mercury",
		SupportsAPIKey: true,
		// Mercury authenticates with a self-serve API token presented as
		// Authorization: Bearer, the default APIKeyConnection scheme. There
		// is no third-party OAuth2 flow for the Users API. The token is
		// bound to one Mercury organization, so there is nothing to pick
		// (Pattern 3): no settings struct, no picker, no
		// SetOrganizationSettings.
		//
		// ProbeURL lets the connection-status check confirm the token is
		// live with a lightweight GET; the transport attaches the Bearer
		// token and a dead token returns 401/403.
		ProbeURL: "https://api.mercury.com/api/v1/users?limit=1",
		//
		// No NewNameResolver: GET /api/v1/users carries no organization
		// name and a read-only token may lack other scopes, so the source
		// keeps its generic name (the source-name worker degrades
		// gracefully).
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			return drivers.NewMercuryDriver(c), nil
		},
	}
}
