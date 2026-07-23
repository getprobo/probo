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
		Provider:       coredata.ConnectorProviderUpCloud,
		DisplayName:    "UpCloud",
		SupportsAPIKey: true,
		// UpCloud's newer API tokens (the "ucat_..." personal access tokens
		// created under People > API access) authenticate as a standard
		// Bearer token, so the default APIKeyConnection mode (Authorization:
		// Bearer <key>) applies; no Header/Scheme/BasicAuth override is
		// needed. There is no OAuth2 flow; account/list already returns the
		// main account plus every sub-account reachable with the token, so
		// there is nothing to pick or configure: no settings struct, no
		// picker.
		//
		// ProbeURL lets the connection-status check confirm the token with
		// the same lightweight GET the driver uses; an invalid token returns
		// 401.
		ProbeURL: "https://api.upcloud.com/1.3/account/list",
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, logger *log.Logger) (drivers.Driver, error) {
			return drivers.NewUpCloudDriver(c, logger), nil
		},
	}
}
