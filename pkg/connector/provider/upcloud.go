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

func upcloudRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderUpCloud,
		DisplayName:      "UpCloud",
		DocumentationURL: accessReviewDocsURL("upcloud"),
		APIKey:           &APIKeyConfig{},
		Endpoints: Endpoints{
			// Every endpoint the driver calls lives under the same /1.3
			// prefix, so the version segment stays in APIBase.
			APIBase: "https://api.upcloud.com/1.3",
			// UpCloud API tokens ("ucat_...", created under Account > API
			// tokens) authenticate as a standard Bearer token, so the default
			// APIKeyConnection mode applies; no Header/Scheme/BasicAuth
			// override is needed. There is no OAuth2 flow, and account/list
			// already returns the main account plus every sub-account the token
			// reaches, so there is nothing to pick: no settings struct, no
			// picker.
			//
			// ProbeURL lets the connection-status check confirm the token with
			// the same lightweight GET the driver uses; a bad token returns 401.
			// account/list is main-account-only, so it also rejects a
			// sub-account token, which authenticates but sees nothing.
			Probe: "https://api.upcloud.com/1.3/account/list",
		},
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, logger *log.Logger, ep Endpoints) (drivers.Driver, error) {
			return drivers.NewUpCloudDriver(c, logger.Named("upcloud"), ep.APIBase), nil
		},
		// GET /1.3/account names the source after the token's own account.
		NewNameResolver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger, ep Endpoints) drivers.NameResolver {
			return drivers.NewUpCloudNameResolver(c, ep.APIBase)
		},
	}
}
