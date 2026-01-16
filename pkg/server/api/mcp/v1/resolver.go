//go:generate go run go.probo.inc/mcpgen generate

package mcp_v1

import (
	"context"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/authn"
)

type Resolver struct {
	proboSvc *probo.Service
	iamSvc   *iam.Service
	logger   *log.Logger
}

func (r *Resolver) MustAuthorize(ctx context.Context, entityID gid.GID, action iam.Action) {
	identity := authn.IdentityFromContext(ctx)

	err := r.iamSvc.Authorizer.Authorize(
		ctx,
		iam.AuthorizeParams{
			Principal: identity.ID,
			Resource:  entityID,
			Action:    action,
		},
	)
	if err != nil {
		panic(err)
	}
}
