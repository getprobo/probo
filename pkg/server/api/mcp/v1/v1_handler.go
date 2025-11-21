package mcp_v1

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/authz"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/mcp/mcputils"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/server"
	serverauth "go.probo.inc/probo/pkg/server/auth"
)

func (r *Resolver) ProboService(ctx context.Context, objectID gid.GID) *probo.TenantService {
	serverauth.RequireTenantAccess(ctx, objectID.TenantID())
	return r.proboSvc.WithTenant(objectID.TenantID())
}

func NewMux(logger *log.Logger, proboSvc *probo.Service, authSvc *auth.Service, authzSvc *authz.Service, cfg Config) *chi.Mux {
	logger = logger.Named("mcp.v1")

	logger.Info("initializing MCP server",
		log.String("version", cfg.Version),
		log.String("request_timeout", cfg.RequestTimeout.String()),
	)
	// server.AddReceivingMiddleware(mcputils.LoggingMiddleware(logger))

	resolver := &Resolver{
		proboSvc: proboSvc,
		authSvc:  authSvc,
		authzSvc: authzSvc,
		logger:   logger,
	}

	mcpServer := server.New(resolver)

	// Add panic recovery middleware to handle panics in goroutines spawned by MCP SDK
	mcpServer.AddReceivingMiddleware(mcputils.LoggingMiddleware(logger))
	mcpServer.AddReceivingMiddleware(mcputils.RecoveryMiddleware(logger))

	getServer := func(r *http.Request) *mcp.Server { return mcpServer }
	eventStore := mcp.NewMemoryEventStore(nil)

	handler := mcp.NewStreamableHTTPHandler(
		getServer,
		&mcp.StreamableHTTPOptions{
			Stateless:      false,
			SessionTimeout: 30 * time.Minute,
			EventStore:     eventStore,
			Logger:         nil, // TODO put logger here
		},
	)

	authHandler := WithMCPAuth(logger, authSvc, authzSvc, handler)

	r := chi.NewMux()
	r.Handle("/", authHandler)

	logger.Info("MCP server initialized successfully")

	return r
}
