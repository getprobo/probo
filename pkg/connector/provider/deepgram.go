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

func deepgramRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderDeepgram,
		DisplayName:      "Deepgram",
		DocumentationURL: accessReviewDocsURL("deepgram"),
		SupportsAPIKey:   true,
		// Deepgram authenticates with an API key under the `Token` scheme
		// (`Authorization: Token <key>`), not Bearer. APIKeyAuthScheme makes
		// the APIKeyConnection use that scheme. There is no third-party
		// OAuth2 flow; the customer supplies an owner/admin key bound to one
		// account, so there is nothing to pick (Pattern 3): no settings
		// struct, no picker.
		APIKeyAuthScheme: "Token",
		// ProbeURL lets the connection-status check confirm the key with a
		// lightweight GET; the transport attaches the `Token` credential and
		// a dead key returns 401/403.
		ProbeURL: "https://api.deepgram.com/v1/projects",
		//
		// No NewNameResolver: an account may span several projects, so there
		// is no single instance name; the source keeps its generic name.
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			return drivers.NewDeepgramDriver(c), nil
		},
	}
}
