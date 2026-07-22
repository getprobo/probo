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

func squareRegistration() *Registration {
	return &Registration{
		Provider:    coredata.ConnectorProviderSquare,
		DisplayName: "Square",
		AuthURL:     "https://connect.squareup.com/oauth2/authorize",
		TokenURL:    "https://connect.squareup.com/oauth2/token",
		// EMPLOYEES_READ lists team members; MERCHANT_PROFILE_READ is needed
		// for the merchant-name resolver and the /v2/merchants/me probe.
		// Square's confidential token endpoint accepts client credentials in
		// the form body (the default post-form scheme) and rejects HTTP Basic,
		// so no TokenEndpointAuth override is set.
		OAuth2Scopes: []string{"EMPLOYEES_READ", "MERCHANT_PROFILE_READ"},
		Probe:        probeSquare,
		// SupportsAPIKey enables the Personal Access Token fallback, which
		// authenticates with the same Authorization: Bearer scheme as the OAuth
		// token. A Square token — OAuth or PAT — is always scoped to one
		// merchant, so there is nothing to pick (Pattern 3): no settings
		// struct, no picker, no OAuth-callback capture.
		SupportsAPIKey: true,
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			return drivers.NewSquareDriver(c), nil
		},
		NewNameResolver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger) drivers.NameResolver {
			return drivers.NewSquareNameResolver(c)
		},
	}
}
