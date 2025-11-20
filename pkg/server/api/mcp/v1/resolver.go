//go:generate go run go.probo.inc/mcpgen generate

package mcp_v1

import (
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/authz"
	"go.probo.inc/probo/pkg/probo"
)

type Resolver struct {
	proboSvc *probo.Service
	authSvc  *auth.Service
	authzSvc *authz.Service
	logger   *log.Logger
}
