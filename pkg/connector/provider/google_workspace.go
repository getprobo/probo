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
		AuthURL:     "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:    "https://oauth2.googleapis.com/token",
		ExtraAuthParams: map[string]string{
			"access_type": "offline",
			"prompt":      "consent",
		},
		SupportsIncrementalAuth: true,
		ProbeURL:                "https://admin.googleapis.com/admin/directory/v1/users?customer=my_customer&maxResults=1",
		OAuth2Scopes: []string{
			"https://www.googleapis.com/auth/admin.directory.user.readonly",
			"https://www.googleapis.com/auth/admin.directory.group.member.readonly",
			"https://www.googleapis.com/auth/admin.directory.customer.readonly",
		},
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			return drivers.NewGoogleWorkspaceDriver(c), nil
		},
		NewNameResolver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger) drivers.NameResolver {
			return drivers.NewGoogleWorkspaceNameResolver(c)
		},
	}
}
