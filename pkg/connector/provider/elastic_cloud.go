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
	"fmt"
	"net/http"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

// Elastic Cloud has no OAuth flow for its organization API. An API key and
// organization ID identify one organization, so no organization picker is used.
func elasticCloudRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderElasticCloud,
		DisplayName:      "Elastic Cloud",
		DocumentationURL: accessReviewDocsURL("elastic-cloud"),
		APIKey: &APIKeyConfig{
			Auth: APIKeyAuth{Mode: APIKeyAuthScheme, Name: "ApiKey"},
			ExtraSettings: []ExtraSetting{
				{Key: "organization_id", Label: "Organization ID", Required: true},
			},
		},
		BuildProbeURL: buildElasticCloudProbeURL,
		Endpoints: Endpoints{
			APIBase: "https://api.elastic-cloud.com/api/v1",
		},
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger, ep Endpoints) (drivers.Driver, error) {
			settings, err := coredata.ConnectorSettings[coredata.ElasticCloudConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read elastic cloud connector settings: %w", err)
			}

			if settings.OrganizationID == "" {
				return nil, fmt.Errorf("cannot create elastic cloud driver: organization_id is required")
			}

			return drivers.NewElasticCloudDriver(c, settings.OrganizationID, ep.APIBase), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger, ep Endpoints) drivers.NameResolver {
			settings, err := coredata.ConnectorSettings[coredata.ElasticCloudConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read elastic cloud connector settings", log.Error(err))

				return nil
			}

			if settings.OrganizationID == "" {
				logger.ErrorCtx(ctx, "missing elastic cloud organization ID")

				return nil
			}

			return drivers.NewElasticCloudNameResolver(c, settings.OrganizationID, ep.APIBase)
		},
	}
}
