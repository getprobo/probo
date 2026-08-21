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

func googleWorkspaceRegistration() *Registration {
	return &Registration{
		Provider:    coredata.ConnectorProviderGoogleWorkspace,
		DisplayName: "Google Workspace",
		// APIBase is deliberately empty. Unlike every other provider here, the
		// Admin Directory host is not a Probo literal: it lives in the Google
		// SDK's generated basePath, which admin.NewService picks up (see
		// drivers.GoogleWorkspaceDriver). Honouring an override would mean
		// threading option.WithEndpoint into the SDK, and the SDK expects a
		// trailing-slash root — the opposite of the no-trailing-slash APIBase
		// this struct documents. Declaring an APIBase nothing reads would be
		// worse than declaring none: the Probe below would silently keep
		// pointing at production while the SDK went on using its own basePath.
		EndpointOverrideUnsupported: "its data host is the Google SDK's basePath (admin.NewService), not a value in Endpoints",
		Endpoints: Endpoints{
			Auth:  "https://accounts.google.com/o/oauth2/v2/auth",
			Token: "https://oauth2.googleapis.com/token",
			Probe: "https://admin.googleapis.com/admin/directory/v1/users?customer=my_customer&maxResults=1",
		},
		ExtraAuthParams: map[string]string{
			"access_type": "offline",
			"prompt":      "consent",
		},
		SupportsIncrementalAuth: true,
		OAuth2Scopes: []string{
			"https://www.googleapis.com/auth/admin.directory.user.readonly",
			"https://www.googleapis.com/auth/admin.directory.group.member.readonly",
			"https://www.googleapis.com/auth/admin.directory.customer.readonly",
		},
		NewDriver: HTTP(func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, _ Endpoints) (drivers.Driver, error) {
			return drivers.NewGoogleWorkspaceDriver(c), nil
		}),
		NewNameResolver: HTTPNameResolver(func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, _ Endpoints) drivers.NameResolver {
			return drivers.NewGoogleWorkspaceNameResolver(c)
		}),
	}
}
