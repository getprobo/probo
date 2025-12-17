//go:generate go run go.probo.inc/mcpgen generate

package mcp_v1

import (
	"context"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	connect_v1 "go.probo.inc/probo/pkg/server/api/connect/v1"
)

type Resolver struct {
	proboSvc *probo.Service
	iamSvc   *iam.Service
	logger   *log.Logger
}

func (r *Resolver) MustBeAuthorized(ctx context.Context, entityID gid.GID, action iam.Action) {
	user := connect_v1.UserFromContext(ctx)
	apiKey := connect_v1.APIKeyFromContext(ctx)
	if user == nil {
		panic(&iam.TenantAccessError{Message: "authentication required"})
	}

	var credentialID *gid.GID
	if apiKey != nil {
		credentialID = &apiKey.ID
	}

	err := r.iamSvc.LegacyAccessManagementService.Authorize(ctx, user.ID, credentialID, entityID, action)
	if err != nil {
		panic(err)
	}
}
