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

// elevenLabsRegistration wires the ElevenLabs access-review connector.
//
// ElevenLabs authenticates with a workspace API key in its own xi-api-key
// header — it does not read Authorization — and the key is bound to one
// workspace, so there is nothing to pick (Pattern 3): no settings struct, no
// picker, no SetOrganizationSettings. There is no third-party OAuth2 flow for
// the workspace member list.
//
// No NewNameResolver: nothing ElevenLabs exposes to an API key carries the
// workspace name (GET /v1/user answers with the caller's own profile and
// subscription), so the source keeps its generic name.
func elevenLabsRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderElevenLabs,
		DisplayName:      "ElevenLabs",
		DocumentationURL: accessReviewDocsURL("elevenlabs"),
		APIKey: &APIKeyConfig{
			Auth: APIKeyAuth{Mode: APIKeyAuthHeader, Name: "xi-api-key"},
			// No KeyFormat: ElevenLabs documents no prefix for its keys, and a
			// shape inferred from a sample would reject valid keys the day it
			// mints a different one. probeElevenLabs names a bad key instead,
			// off the provider's own answer.
		},
		Endpoints: Endpoints{
			APIBase: "https://api.elevenlabs.io/v1",
			// No static Probe: ElevenLabs rejects a key with 400 rather than
			// 401, which a plain GET would read as connected, so the check is
			// the probeElevenLabs closure.
		},
		Probe: probeElevenLabs,
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) (drivers.Driver, error) {
			return drivers.NewElevenLabsDriver(c, ep.APIBase), nil
		},
	}
}
