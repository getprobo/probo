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

func openrouterRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderOpenRouter,
		DisplayName:    "OpenRouter",
		SupportsAPIKey: true,
		// OpenRouter authenticates with an organization management
		// (provisioning) API key presented as Authorization: Bearer, the
		// default APIKeyConnection scheme. The key is bound to one
		// organization, so there is nothing to pick (Pattern 3): no settings
		// struct, no picker.
		//
		// Probe confirms the key with a lightweight GET against the members
		// endpoint; the transport attaches the Bearer token. probeOpenRouter
		// rejects 401/403 (revoked/invalid key) and also 404, which a valid
		// but personal (non-organization) key returns — so a key that cannot
		// list members shows as not-connected rather than failing later.
		Probe: probeOpenRouter,
		//
		// No NewNameResolver: the members endpoint carries no organization
		// name, so the source keeps its generic name.
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			return drivers.NewOpenRouterDriver(c), nil
		},
	}
}
