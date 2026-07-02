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
	"net/http"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

func pylonRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderPylon,
		DisplayName:    "Pylon",
		SupportsAPIKey: true,
		// Pylon authenticates with an account API token presented as
		// Authorization: Bearer, the default APIKeyConnection scheme. There
		// is no third-party OAuth2 flow for the Users API. The token is bound
		// to one Pylon organization, so there is nothing to pick (Pattern 3):
		// no settings struct, no picker, no SetOrganizationSettings.
		//
		// ProbeURL lets the connection-status check confirm the token is live
		// with a lightweight GET; the transport attaches the Bearer token and
		// a dead token returns 401/403.
		ProbeURL: "https://api.usepylon.com/users?limit=1",
		//
		// No NewNameResolver: GET /users carries no organization name, so the
		// source keeps its generic name (the source-name worker degrades
		// gracefully).
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			return drivers.NewPylonDriver(c), nil
		},
	}
}
