// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package mcp_v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.gearno.de/kit/log"
	mcpgenmcp "go.probo.inc/mcpgen/mcp"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/complianceportal/management"
	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/resourcealias"
	"go.probo.inc/probo/pkg/riskmanagement"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/mcp/mcputils"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/server"
	"go.probo.inc/probo/pkg/thirdparty"
)

func NewMux(
	logger *log.Logger,
	proboSvc *probo.Service,
	managementSvc *management.Service,
	resourceAliasSvc *resourcealias.Service,
	thirdPartySvc *thirdparty.Service,
	iamSvc *iam.Service,
	accessReviewSvc *accessreview.Service,
	cookieBannerSvc *cookiebanner.Service,
	riskManagementSvc *riskmanagement.Service,
	tokenSecret string,
	fileManagerSvc *filemanager.Service,
	baseURL *baseurl.BaseURL,
) *chi.Mux {
	logger = logger.Named("mcp.v1")

	logger.Info("initializing MCP server")

	resolver := &Resolver{
		proboSvc:       proboSvc,
		management:     managementSvc,
		resourceAlias:  resourceAliasSvc,
		thirdPartySvc:  thirdPartySvc,
		iamSvc:         iamSvc,
		accessReview:   accessReviewSvc,
		cookieBanner:   cookieBannerSvc,
		riskManagement: riskManagementSvc,
		logger:         logger,
		fileManager:    fileManagerSvc,
		baseURL:        baseURL,
	}

	mcpServer := server.New(resolver, mcpgenmcp.WithRecoverFunc(mcputils.NewRecoverFunc(logger)))

	mcpServer.AddReceivingMiddleware(mcputils.LoggingMiddleware(logger))

	getServer := func(r *http.Request) *mcp.Server { return mcpServer }
	eventStore := mcp.NewMemoryEventStore(nil)

	handler := mcp.NewStreamableHTTPHandler(
		getServer,
		&mcp.StreamableHTTPOptions{
			Stateless: true,
			// SessionTimeout: 30 * time.Minute,
			EventStore: eventStore,
			Logger:     nil, // TODO put logger here
		},
	)
	protectedHandler := http.NewCrossOriginProtection().Handler(handler)

	r := chi.NewMux()
	r.Use(authn.NewAPIKeyMiddleware(iamSvc, tokenSecret))
	r.Use(authn.NewOAuth2AccessTokenMiddleware(iamSvc))
	r.Use(authn.NewIdentityPresenceMiddleware(baseURL))
	r.Handle("/", protectedHandler)

	logger.Info("MCP server initialized successfully")

	return r
}

func UnwrapOmittable[T any](field mcpgenmcp.Omittable[T]) *T {
	if !field.IsSet() {
		return nil
	}

	value, _ := field.Value()

	return &value
}
