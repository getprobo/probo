package v1

import (
	"context"
	"fmt"
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
)

type (
	resolver struct {
		proboSvc *probo.Service
		authSvc  *auth.Service
		authzSvc *authz.Service
		logger   *log.Logger
	}
)

func (r *resolver) ProboService(ctx context.Context, tenantID gid.TenantID) *probo.TenantService {
	validateTenantAccess(ctx, tenantID)
	return r.proboSvc.WithTenant(tenantID)
}

func validateTenantAccess(ctx context.Context, tenantID gid.TenantID) {
	mcpCtx := MCPContextFromContext(ctx)
	if mcpCtx == nil {
		panic(fmt.Errorf("authentication context not found"))
	}

	for _, tid := range mcpCtx.TenantIDs {
		if tid == tenantID {
			return
		}
	}

	panic(fmt.Errorf("access denied: user does not have access to tenant %s", tenantID.String()))
}

func NewMux(logger *log.Logger, proboSvc *probo.Service, authSvc *auth.Service, authzSvc *authz.Service, cfg Config) *chi.Mux {
	logger = logger.Named("mcp.v1")

	logger.Info("initializing MCP server",
		log.String("version", cfg.Version),
		log.String("request_timeout", cfg.RequestTimeout.String()),
	)

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "probo",
			Title:   "Probo",
			Version: cfg.Version,
		},
		&mcp.ServerOptions{},
	)

	server.AddReceivingMiddleware(mcputils.LoggingMiddleware(logger))

	resolver := &resolver{
		proboSvc: proboSvc,
		authSvc:  authSvc,
		authzSvc: authzSvc,
		logger:   logger,
	}

	mcp.AddTool(server, ListOrganizationsTool, resolver.ListOrganizations)
	mcp.AddTool(server, ListVendorsTool, resolver.ListVendors)
	mcp.AddTool(server, AddVendorTool, resolver.AddVendor)

	getServer := func(r *http.Request) *mcp.Server { return server }
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
