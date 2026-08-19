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

func mondayRegistration() *Registration {
	return &Registration{
		Provider:    coredata.ConnectorProviderMonday,
		DisplayName: "Monday.com",
		Endpoints: Endpoints{
			Auth:  "https://auth.monday.com/oauth2/authorize",
			Token: "https://auth.monday.com/oauth2/token",
			// Monday.com's data API is a single GraphQL endpoint, so APIBase
			// is that endpoint: the driver, name resolver and probe POST to
			// it verbatim rather than joining a path onto it.
			APIBase: "https://api.monday.com/v2",
		},
		Probe:        HTTPProbe(probeMonday),
		OAuth2Scopes: []string{"users:read", "account:read"},
		NewDriver: HTTP(func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) (drivers.Driver, error) {
			return drivers.NewMondayDriver(c, ep.APIBase), nil
		}),
		NewNameResolver: HTTPNameResolver(func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) drivers.NameResolver {
			return drivers.NewMondayNameResolver(c, ep.APIBase)
		}),
	}
}
