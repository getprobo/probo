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

// mongoDBAtlasRegistration describes the MongoDB Atlas access-review
// connector: OAuth2 client credentials from a service account the customer
// creates in their own organization.
//
// No ExtraSettings, and so no settings struct and no picker — a service
// account is scoped to exactly one organization, and the driver resolves its id
// from GET /orgs. The commit body carries the rest of the reasoning.
func mongoDBAtlasRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderMongoDBAtlas,
		DisplayName:      "MongoDB Atlas",
		DocumentationURL: accessReviewDocsURL("mongodb-atlas"),
		Endpoints: Endpoints{
			Token:   "https://cloud.mongodb.com/api/oauth/token",
			APIBase: "https://cloud.mongodb.com/api/atlas/v2",
		},
		// A closure rather than Endpoints.Probe: the check must send Atlas's
		// versioned Accept header.
		Probe:             probeMongoDBAtlas,
		ClientCredentials: &ClientCredentialsConfig{},
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) (drivers.Driver, error) {
			return drivers.NewMongoDBAtlasDriver(c, ep.APIBase), nil
		},
		NewNameResolver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) drivers.NameResolver {
			return drivers.NewMongoDBAtlasNameResolver(c, ep.APIBase)
		},
	}
}
