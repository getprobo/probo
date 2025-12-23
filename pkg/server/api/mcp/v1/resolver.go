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
	user := connect_v1.IdentityFromContext(ctx)
	apiKey := connect_v1.APIKeyFromContext(ctx)
	if user == nil {
		panic(&iam.TenantAccessError{Message: "authentication required"})
	}

	var credentialID *gid.GID
	if apiKey != nil {
		credentialID = &apiKey.ID
	}

	// When API key is used, fall back to legacy system for intersection semantics.
	// The legacy system handles API key role checking properly.
	// TODO: Migrate API key authorization to new system.
	if credentialID != nil {
		err := r.iamSvc.LegacyAccessManagementService.Authorize(ctx, user.ID, credentialID, entityID, action)
		if err != nil {
			panic(err)
		}
		return
	}

	// Map legacy action to new namespaced action
	newAction, ok := probo.MapLegacyAction(entityID.EntityType(), action)
	if !ok {
		// Fall back to legacy system for unmapped actions
		err := r.iamSvc.LegacyAccessManagementService.Authorize(ctx, user.ID, credentialID, entityID, action)
		if err != nil {
			panic(err)
		}
		return
	}

	// Use new authorizer with mapped action
	err := r.iamSvc.Authorizer.Authorize(ctx, iam.AuthorizeParams{
		Principal: user.ID,
		Resource:  entityID,
		Action:    newAction,
	})
	if err != nil {
		panic(err)
	}
}
