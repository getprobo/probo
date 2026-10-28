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

// daytonaRegistration wires Daytona. An API key belongs to one organization,
// which the driver reads from the key itself, so there is nothing to pick.
// No NewNameResolver: an API key cannot read the organization name.
func daytonaRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderDaytona,
		DisplayName:      "Daytona",
		DocumentationURL: accessReviewDocsURL("daytona"),
		APIKey: &APIKeyConfig{
			CredentialPage: &CredentialPage{URL: "https://app.daytona.io/dashboard/keys"},
		},
		Endpoints: Endpoints{
			Probe:   "https://app.daytona.io/api/api-keys/current",
			APIBase: "https://app.daytona.io/api",
		},
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) (drivers.Driver, error) {
			return drivers.NewDaytonaDriver(c, ep.APIBase), nil
		},
	}
}
