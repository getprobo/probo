//go:generate go run go.probo.inc/mcpgen generate

package mcp_v1

import (
	"context"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/authz"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probo"
	serverauth "go.probo.inc/probo/pkg/server/auth"
)

type Resolver struct {
	proboSvc *probo.Service
	authSvc  *auth.Service
	authzSvc *authz.Service
	logger   *log.Logger
}

func (r *Resolver) MustBeAuthorized(ctx context.Context, entityID gid.GID, action authz.Action) {
	user := serverauth.UserFromContext(ctx)
	apiKey := serverauth.UserAPIKeyFromContext(ctx)
	if user == nil {
		panic(&authz.TenantAccessError{Message: "authentication required"})
	}

	authzSvc := r.AuthzService(ctx, entityID.TenantID())
	err := authzSvc.Authorize(ctx, user, apiKey, entityID, action)
	if err != nil {
		panic(err)
	}
}

func (r *Resolver) AuthzService(ctx context.Context, tenantID gid.TenantID) *authz.TenantAuthzService {
	return GetTenantAuthzService(ctx, r.authzSvc, tenantID)
}

func GetTenantAuthzService(ctx context.Context, authzSvc *authz.Service, tenantID gid.TenantID) *authz.TenantAuthzService {
	serverauth.RequireTenantAccess(ctx, tenantID)
	return authzSvc.WithTenant(tenantID)
}
