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

func yousignRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderYousign,
		DisplayName:    "Yousign",
		SupportsAPIKey: true,
		// Yousign authenticates with an API key as Authorization: Bearer. The
		// key is bound to one organization, so GET /v3/users returns everyone
		// with nothing to pick (Pattern 3). The connector targets Yousign
		// production; the sandbox runs on a separate host and is not a reviewed
		// environment.
		//
		// ProbeURL lets the connection-status check confirm the key with a
		// lightweight GET; the transport attaches the Bearer credential and a
		// dead key returns 401/403.
		//
		// No NewNameResolver: Yousign v3 exposes no organization-name endpoint,
		// so the source keeps its generic name.
		ProbeURL: "https://api.yousign.app/v3/users?limit=1",
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			return drivers.NewYousignDriver(c), nil
		},
	}
}
