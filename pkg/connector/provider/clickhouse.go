// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
