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

func clickhouseRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderClickHouse,
		DisplayName:    "ClickHouse Cloud",
		SupportsAPIKey: true,
		// ClickHouse Cloud's control-plane API authenticates with HTTP Basic
		// auth where the credential is keyId:keySecret. APIKeyBasicAuthUserPass
		// makes the APIKeyConnection base64 the verbatim "keyId:keySecret"
		// the operator pastes (the empty-password APIKeyBasicAuth cannot
		// carry the secret). There is no OAuth2 flow; a key/secret pair is
		// scoped to exactly one organization, which the driver discovers via
		// GET /v1/organizations, so there is nothing to pick or configure
		// (Pattern 3): no settings struct, no picker.
		APIKeyBasicAuthUserPass: true,
		// ProbeURL lets the connection-status check confirm the key/secret
		// with a lightweight GET; the transport attaches the Basic
		// credential and a dead key/secret returns 401/403.
		ProbeURL: "https://api.clickhouse.cloud/v1/organizations",
		//
		// No NewNameResolver: the organization name is available but would
		// duplicate the driver's discovery call; the source keeps its
		// generic name.
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			return drivers.NewClickHouseDriver(c), nil
		},
	}
}
