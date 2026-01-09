package mcp_v1

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.gearno.de/kit/log"
	mcpgenmcp "go.probo.inc/mcpgen/mcp"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/mcp/mcputils"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/server"
)

func (r *Resolver) ProboService(ctx context.Context, objectID gid.GID) *probo.TenantService {
	return r.proboSvc.WithTenant(objectID.TenantID())
}

func NewMux(logger *log.Logger, proboSvc *probo.Service, iamSvc *iam.Service, tokenSecret string) *chi.Mux {
	logger = logger.Named("mcp.v1")

	logger.Info("initializing MCP server")
	// server.AddReceivingMiddleware(mcputils.LoggingMiddleware(logger))

	resolver := &Resolver{
		proboSvc: proboSvc,
		iamSvc:   iamSvc,
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
			Stateless: true,
			// SessionTimeout: 30 * time.Minute,
			EventStore: eventStore,
			Logger:     nil, // TODO put logger here
		},
	)

	r := chi.NewMux()
	r.Use(authn.NewAPIKeyMiddleware(iamSvc, tokenSecret))
	r.Handle("/", RequireAPIKeyHandler(logger, handler))

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
