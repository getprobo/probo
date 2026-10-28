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

// elasticCloudRegistration describes the Elastic Cloud access-review
// connector: an organization API key, sent as "Authorization: ApiKey <key>".
// No KeyFormat: keys only gained the essu_ prefix in late 2023, and an older
// key may never expire.
//
// The probe lists API keys, which a Billing admin key is refused, so a key
// too narrow to review the organization reads as not authorized.
func elasticCloudRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderElasticCloud,
		DisplayName:      "Elastic Cloud",
		DocumentationURL: accessReviewDocsURL("elastic-cloud"),
		APIKey: &APIKeyConfig{
			Auth:           APIKeyAuth{Mode: APIKeyAuthScheme, Name: "ApiKey"},
			CredentialPage: &CredentialPage{URL: "https://cloud.elastic.co/organization/keys"},
		},
		Endpoints: Endpoints{
			APIBase: "https://api.elastic-cloud.com/api/v1",
			Probe:   "https://api.elastic-cloud.com/api/v1/users/auth/keys",
		},
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) (drivers.Driver, error) {
			return drivers.NewElasticCloudDriver(c, ep.APIBase), nil
		},
		NewNameResolver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) drivers.NameResolver {
			return drivers.NewElasticCloudNameResolver(c, ep.APIBase)
		},
	}
}
