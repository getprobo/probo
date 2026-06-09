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

func clerkRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderClerk,
		DisplayName:    "Clerk",
		SupportsAPIKey: true,
		// Clerk's Backend API authenticates with a server-side secret key
		// (sk_...) presented as Authorization: Bearer, the default
		// APIKeyConnection scheme. There is no third-party OAuth2 flow for
		// account-listing: Clerk's OAuth is an end-user IdP (scoped consent
		// to a single user's profile), not a partner grant over the Backend
		// API. The secret key is bound to one Clerk instance, so there is
		// nothing to pick (Pattern 3): no settings struct, no picker, no
		// SetOrganizationSettings.
		//
		// ProbeURL is intentionally empty: the connection probe runs only
		// for OAuth2 connections, so it would be dead config for an API-key
		// provider; a dead key surfaces on the first ListAccounts instead.
		//
		// No NewNameResolver: the Backend API exposes no instance/application
		// name endpoint reachable with a secret key, so the source keeps its
		// generic name (the source-name worker degrades gracefully).
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			return drivers.NewClerkDriver(c), nil
		},
	}
}
