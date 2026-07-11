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

func dotfileRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderDotfile,
		DisplayName:    "Dotfile",
		SupportsAPIKey: true,
		// Dotfile authenticates with the API key in the X-DOTFILE-API-KEY
		// header rather than Authorization: Bearer. APIKeyHeader makes the
		// APIKeyConnection send that header and omit Authorization. The key is
		// bound to one workspace, so there is nothing to pick (Pattern 3): no
		// settings struct, no picker.
		APIKeyHeader: "X-DOTFILE-API-KEY",
		// ProbeURL lets the connection-status check confirm the key with a
		// lightweight GET; the transport attaches X-DOTFILE-API-KEY and a dead
		// key returns 401.
		ProbeURL: "https://api.dotfile.com/v1/users?limit=1",
		// No NewNameResolver: the users endpoint carries no workspace name, so
		// the source keeps its generic name.
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			return drivers.NewDotfileDriver(c), nil
		},
	}
}
