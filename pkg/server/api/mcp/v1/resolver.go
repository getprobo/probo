//go:generate go run go.probo.inc/mcpgen generate

package mcp_v1

import (
	"context"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	serverauth "go.probo.inc/probo/pkg/server/auth"
)

type Resolver struct {
	proboSvc *probo.Service
	iamSvc   *iam.Service
	logger   *log.Logger
}

func (r *Resolver) MustBeAuthorized(ctx context.Context, entityID gid.GID, action iam.Action) {
	user := serverauth.UserFromContext(ctx)
	apiKey := serverauth.UserAPIKeyFromContext(ctx)
	if user == nil {
		panic(&iam.TenantAccessError{Message: "authentication required"})
	}

	iamSvc := r.IAMService(ctx, entityID.TenantID())
	err := iamSvc.Authorize(ctx, user, apiKey, entityID, action)
	if err != nil {
		panic(err)
	}
}

func (r *Resolver) IAMService(ctx context.Context, tenantID gid.TenantID) *iam.TenantService {
	return GetTenantIAMService(ctx, r.iamSvc, tenantID)
}

func GetTenantIAMService(ctx context.Context, iamSvc *iam.Service, tenantID gid.TenantID) *iam.TenantService {
	serverauth.RequireTenantAccess(ctx, tenantID)
	return iamSvc.WithTenant(tenantID)
}
