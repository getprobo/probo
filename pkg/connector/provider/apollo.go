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

func apolloRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderApollo,
		DisplayName:    "Apollo.io",
		SupportsAPIKey: true,
		// Apollo's REST API authenticates with a master API key in the
		// x-api-key header; it rejects Authorization: Bearer (and, since
		// Sept 2024, query/body key params). APIKeyHeader makes the
		// APIKeyConnection send x-api-key instead of Bearer. There is no
		// OAuth2 flow needed: the customer supplies a master key, which is
		// bound to one Apollo account, so there is nothing to pick
		// (Pattern 3): no settings struct, no picker.
		APIKeyHeader: "x-api-key",
		// ProbeURL lets the connection-status check confirm the key with a
		// lightweight GET; the transport attaches x-api-key, and a missing,
		// dead, or non-master key returns 401/403.
		ProbeURL: "https://api.apollo.io/api/v1/users/search?page=1&per_page=1",
		//
		// No NewNameResolver: Apollo exposes no stable account-name
		// endpoint reachable with the master key, so the source keeps its
		// generic name.
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			return drivers.NewApolloDriver(c), nil
		},
	}
}
