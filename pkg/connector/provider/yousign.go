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

func yousignRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderYousign,
		DisplayName:    "Yousign",
		SupportsAPIKey: true,
		// Yousign authenticates with an API key as Authorization: Bearer. The
		// key is bound to one organization, so GET /v3/users returns everyone
		// with nothing to pick (Pattern 3). The connector targets Yousign
		// production; the sandbox runs on a separate host and is not a reviewed
		// environment.
		//
		// ProbeURL lets the connection-status check confirm the key with a
		// lightweight GET; the transport attaches the Bearer credential and a
		// dead key returns 401/403.
		//
		// No NewNameResolver: Yousign v3 exposes no organization-name endpoint,
		// so the source keeps its generic name.
		ProbeURL: "https://api.yousign.app/v3/users?limit=1",
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			return drivers.NewYousignDriver(c), nil
		},
	}
}
