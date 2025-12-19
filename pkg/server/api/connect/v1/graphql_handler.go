// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package connect_v1

import (
	"context"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/connect/v1/schema"
	"go.probo.inc/probo/pkg/server/api/connect/v1/types"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

var (
	ErrForbidden = &gqlerror.Error{
		Message: "You are not authorized to access this resource",
		Extensions: map[string]any{
			"code": "FORBIDDEN",
		},
	}

	ErrUnauthorized = &gqlerror.Error{
		Message: "You are not authorized to access this resource",
		Extensions: map[string]any{
			"code": "UNAUTHORIZED",
		},
	}

	ErrAlreadyAuthenticated = &gqlerror.Error{
		Message: "authentication not allowed for this resource/action",
		Extensions: map[string]any{
			"code": "ALREADY_AUTHENTICATED",
		},
	}
)

func SessionDirective(ctx context.Context, obj any, next graphql.Resolver, required types.SessionRequirement) (any, error) {
	session := SessionFromContext(ctx)

	switch required {
	case types.SessionRequirementOptional:
	case types.SessionRequirementPresent:
		if session == nil {
			return nil, ErrUnauthorized
		}
	case types.SessionRequirementNone:
		if session != nil {
			return nil, ErrAlreadyAuthenticated
		}
	}

	return next(ctx)
}

func IsViewerDirective(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
	identity := UserFromContext(ctx)

	switch node := obj.(type) {
	case *types.Identity:
		if identity.ID != node.ID {
			return nil, ErrForbidden
		}
	case *types.Membership:
		if identity.ID != node.IdentityID {
			return nil, ErrForbidden
		}
	case *types.Session:
		if identity.ID != node.IdentityID {
			return nil, ErrForbidden
		}
	default:
	}

	return next(ctx)
}

func NewGraphQLHandler(svc *iam.Service, logger *log.Logger, cookieConfig securecookie.Config) http.Handler {
	config := schema.Config{
		Resolvers: &Resolver{
			iam:          svc,
			cookieConfig: cookieConfig,
		},
		Directives: schema.DirectiveRoot{
			Session:  SessionDirective,
			IsViewer: IsViewerDirective,
		},
	}

	es := schema.NewExecutableSchema(config)
	gqlh := gqlutils.NewHandler(es, logger)
	return gqlh
}
