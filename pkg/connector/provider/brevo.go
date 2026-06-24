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

func brevoRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderBrevo,
		DisplayName:    "Brevo",
		SupportsAPIKey: true,
		// Brevo authenticates with an API key sent in the api-key header
		// rather than Authorization: Bearer. APIKeyHeader makes the
		// APIKeyConnection send api-key instead and omit Authorization. There
		// is no OAuth2 flow needed: the key is bound to one Brevo account, so
		// there is nothing to pick (Pattern 3): no settings struct, no
		// picker.
		APIKeyHeader: "api-key",
		// ProbeURL lets the connection-status check confirm the key with a
		// lightweight GET; the transport attaches the api-key header and a
		// dead key returns 401/403.
		ProbeURL: "https://api.brevo.com/v3/organization/invited/users",
		//
		// No NewNameResolver: the invited-users endpoint carries no account
		// name, so the source keeps its generic name.
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			return drivers.NewBrevoDriver(c), nil
		},
	}
}
