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

package console_v1

import (
	"context"
	"net/url"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
)

// emptyOrganizationsRemediationURL returns the provider page where the user
// can act on an empty organization list, or nil when the provider exposes no
// such page.
//
// GitHub is the only provider with one today: the list of OAuth Apps the
// signed-in user has authorized. From there a member can request — and an
// owner can grant — Probo access to an organization that turned on OAuth App
// access restrictions. Such an organization never appears in GET /user/orgs,
// which is the usual reason the listing comes back empty.
func (r *Resolver) emptyOrganizationsRemediationURL(ctx context.Context, provider coredata.ConnectorProvider) *string {
	if provider != coredata.ConnectorProviderGitHub {
		return nil
	}

	c, err := r.connectorRegistry.Get(provider.String())
	if err != nil {
		return nil
	}

	oauth2Connector, ok := c.(*connector.OAuth2Connector)
	if !ok || oauth2Connector.ClientID == "" {
		return nil
	}

	u, err := url.JoinPath(
		"https://github.com",
		"settings", "connections", "applications",
		url.PathEscape(oauth2Connector.ClientID),
	)
	if err != nil {
		r.logger.ErrorCtx(ctx, "cannot build github oauth application URL", log.Error(err))

		return nil
	}

	return &u
}
