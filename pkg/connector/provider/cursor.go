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

const cursorMembersEndpoint = "https://api.cursor.com/teams/members"

func cursorRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderCursor,
		DisplayName:    "Cursor",
		SupportsAPIKey: true,
		// Cursor's Admin API has no third-party OAuth2 flow; it
		// authenticates with a team admin key (key_...) presented as the
		// HTTP Basic auth username with an empty password ("-u <key>:")
		// and rejects Bearer tokens. APIKeyBasicAuth selects that scheme
		// on the APIKeyConnection. The key is bound to a single team, so
		// there is nothing to pick (Pattern 3): no settings struct, no
		// picker, and no SetOrganizationSettings.
		APIKeyBasicAuth: true,
		ProbeURL:        cursorMembersEndpoint,
		// No NewNameResolver: the Admin API exposes no team/organization
		// name endpoint, so the source keeps its generic name (the
		// source-name worker degrades gracefully when no resolver is set).
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			return drivers.NewCursorDriver(c), nil
		},
	}
}
