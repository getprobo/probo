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

func cursorRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderCursor,
		DisplayName:    "Cursor",
		SupportsAPIKey: true,
		// Cursor's Admin API has no third-party OAuth2 flow; it
		// authenticates with a team admin key (key_...) presented as the
		// HTTP Basic auth username with an empty password ("-u <key>:")
		// and rejects Bearer tokens. APIKeyBasicAuth selects that scheme
		// on the APIKeyConnection. The key is bound to a single team, so
		// there is nothing to pick (Pattern 3): no settings struct, no
		// picker, and no SetOrganizationSettings.
		APIKeyBasicAuth: true,
		// ProbeURL is intentionally empty. API-key connectors skip the
		// connection probe entirely (it runs only for OAuth2), so a probe
		// URL would be dead config; a dead key surfaces on the first
		// ListAccounts instead.
		//
		// No NewNameResolver: the Admin API exposes no team/organization
		// name endpoint, so the source keeps its generic name (the
		// source-name worker degrades gracefully when no resolver is set).
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			return drivers.NewCursorDriver(c), nil
		},
	}
}
