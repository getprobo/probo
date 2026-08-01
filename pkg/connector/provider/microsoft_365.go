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

func microsoft365Registration() *Registration {
	return &Registration{
		Provider:    coredata.ConnectorProviderMicrosoft365,
		DisplayName: "Microsoft 365",
		Endpoints: Endpoints{
			Auth:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			Token: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			Probe: "https://graph.microsoft.com/v1.0/organization?$top=1",
			// Microsoft Graph, a different host from the
			// login.microsoftonline.com OAuth endpoints above. The scopes
			// below only look like URLs — they are Graph permission
			// identifiers, not endpoints, and must not be derived from this.
			APIBase: "https://graph.microsoft.com/v1.0",
		},
		ExtraAuthParams: map[string]string{
			"prompt": "consent",
		},
		OAuth2Scopes: []string{
			"openid",
			"profile",
			"offline_access",
			"https://graph.microsoft.com/AuditLog.Read.All",
			"https://graph.microsoft.com/User.Read.All",
			"https://graph.microsoft.com/Directory.Read.All",
			"https://graph.microsoft.com/RoleManagement.Read.Directory",
		},
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, logger *log.Logger, ep Endpoints) (drivers.Driver, error) {
			return drivers.NewMicrosoft365Driver(c, logger.Named("microsoft365"), ep.APIBase), nil
		},
		NewNameResolver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) drivers.NameResolver {
			return drivers.NewMicrosoft365NameResolver(c, ep.APIBase)
		},
	}
}
