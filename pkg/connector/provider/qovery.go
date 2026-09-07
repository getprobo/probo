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
	"regexp"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

// qoveryKeyPattern covers both organization token classes, which ride the
// same Token scheme: a regular API token and a policy token. What it turns
// away is the Console's own JWT, which the API takes only as a Bearer.
var qoveryKeyPattern = regexp.MustCompile(`^(?:qov_|sk-qov-)[\s\S]`)

func qoveryRegistration() *Registration {
	return &Registration{
		Provider:         coredata.ConnectorProviderQovery,
		DisplayName:      "Qovery",
		DocumentationURL: accessReviewDocsURL("qovery"),
		APIKey: &APIKeyConfig{
			Auth: APIKeyAuth{Mode: APIKeyAuthScheme, Name: "Token"},
			ExtraSettings: []ExtraSetting{
				{Key: "organizationId", Label: "Organization ID", Required: true},
			},
			KeyFormat: &KeyFormat{
				Pattern: qoveryKeyPattern,
				Example: "qov_… or sk-qov-…",
			},
		},
		BuildProbeURL: buildQoveryProbeURL,
		Endpoints: Endpoints{
			// Qovery's API is unversioned in the path; the driver joins the
			// resource segments onto this origin.
			APIBase: "https://api.qovery.com",
		},
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger, ep Endpoints) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.QoveryConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read qovery connector settings: %w", err)
			}

			if s.OrganizationID == "" {
				return nil, fmt.Errorf("cannot create qovery driver: organization_id is required")
			}

			return drivers.NewQoveryDriver(c, s.OrganizationID, ep.APIBase), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger, ep Endpoints) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.QoveryConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read qovery connector settings", log.Error(err))
				return nil
			}

			return drivers.NewQoveryNameResolver(c, s.OrganizationID, ep.APIBase)
		},
	}
}
