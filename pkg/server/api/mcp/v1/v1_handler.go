// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package mcp_v1

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.gearno.de/kit/log"
	mcpgenmcp "go.probo.inc/mcpgen/mcp"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/riskmanagement"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/mcp/mcputils"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/server"
)

func NewMux(
	logger *log.Logger,
	proboSvc *probo.Service,
	iamSvc *iam.Service,
	accessReviewSvc *accessreview.Service,
	cookieBannerSvc *cookiebanner.Service,
	riskManagementSvc *riskmanagement.Service,
	baseURL *baseurl.BaseURL,
	tokenSecret string,
) *chi.Mux {
	logger = logger.Named("mcp.v1")

	logger.Info("initializing MCP server")

	resolver := &Resolver{
		proboSvc:       proboSvc,
		iamSvc:         iamSvc,
		accessReview:   accessReviewSvc,
		cookieBanner:   cookieBannerSvc,
		riskManagement: riskManagementSvc,
		logger:         logger,
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

	protectedResource, err := ProtectedResourceConfig(baseURL)
	if err != nil {
		panic(fmt.Errorf("cannot configure MCP protected resource: %w", err))
	}

	r := chi.NewMux()
	r.Use(authn.NewAPIKeyMiddleware(iamSvc, tokenSecret))
	r.Use(authn.NewOAuth2AccessTokenMiddleware(iamSvc))
	r.Handle("/", RequireIdentityHandler(logger, protectedResource, protectedHandler))

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
