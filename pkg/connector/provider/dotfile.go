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

func dotfileRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderDotfile,
		DisplayName:      "Dotfile",
		DocumentationURL: accessReviewDocsURL("dotfile"),
		SupportsAPIKey:   true,
		// Dotfile authenticates with the API key in the X-DOTFILE-API-KEY
		// header rather than Authorization: Bearer. APIKeyHeader makes the
		// APIKeyConnection send that header and omit Authorization. The key is
		// bound to one workspace, so there is nothing to pick (Pattern 3): no
		// settings struct, no picker.
		APIKeyHeader: "X-DOTFILE-API-KEY",
		Endpoints: Endpoints{
			// Every endpoint the driver calls lives under the same /v1
			// prefix, so the version segment stays in APIBase.
			APIBase: "https://api.dotfile.com/v1",
			// ProbeURL lets the connection-status check confirm the key with a
			// lightweight GET; the transport attaches X-DOTFILE-API-KEY and a dead
			// key returns 401.
			Probe: "https://api.dotfile.com/v1/users?limit=1",
		},
		// No NewNameResolver: the users endpoint carries no workspace name, so
		// the source keeps its generic name.
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) (drivers.Driver, error) {
			return drivers.NewDotfileDriver(c, ep.APIBase), nil
		},
	}
}
