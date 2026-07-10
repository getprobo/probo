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

package console_v1

import (
	"net/http"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/agentrun"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/complianceportal/management"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/mailman"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/resourcealias"
	"go.probo.inc/probo/pkg/riskmanagement"
	"go.probo.inc/probo/pkg/server/api/authz"
	"go.probo.inc/probo/pkg/server/api/console/v1/dataloader"
	"go.probo.inc/probo/pkg/server/api/console/v1/schema"
	"go.probo.inc/probo/pkg/server/gqlutils"
	"go.probo.inc/probo/pkg/thirdparty"
)

func NewGraphQLHandler(
	iamSvc *iam.Service,
	proboSvc *probo.Service,
	resourceAliasSvc *resourcealias.Service,
	esignSvc *esign.Service,
	managementSvc *management.Service,
	accessReviewSvc *accessreview.Service,
	agentRunSvc *agentrun.Service,
	mailmanSvc *mailman.Service,
	cookieBannerSvc *cookiebanner.Service,
	connectorRegistry *connector.ConnectorRegistry,
	providerRegistry *provider.Registry,
	customDomainCname string,
	tokenSecret string,
	logger *log.Logger,
	thirdPartySvc *thirdparty.Service,
	riskManagementSvc *riskmanagement.Service,
	fileManagerSvc *filemanager.Service,
	baseURL *baseurl.BaseURL,
	limits gqlutils.Limits,
) http.Handler {
	config := schema.Config{
		Resolvers: &Resolver{
			authorize:         dataloader.NewAuthorizeFunc(logger),
			batchAuthorize:    authz.NewBatchAuthorizeFunc(iamSvc, logger),
			probo:             proboSvc,
			resourceAlias:     resourceAliasSvc,
			iam:               iamSvc,
			esign:             esignSvc,
			management:        managementSvc,
			accessReview:      accessReviewSvc,
			agentRun:          agentRunSvc,
			mailman:           mailmanSvc,
			cookieBanner:      cookieBannerSvc,
			connectorRegistry: connectorRegistry,
			providerRegistry:  providerRegistry,
			riskManagement:    riskManagementSvc,
			thirdParty:        thirdPartySvc,
			customDomainCname: customDomainCname,
			tokenSecret:       tokenSecret,
			fileManager:       fileManagerSvc,
			baseURL:           baseURL,
			logger:            logger,
		},
	}

	es := schema.NewExecutableSchema(config)
	gqlh := gqlutils.NewHandler(es, logger, limits)

	return gqlh
}
