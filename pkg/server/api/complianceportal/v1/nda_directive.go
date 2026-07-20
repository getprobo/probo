// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package complianceportal_v1

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/complianceportal/visitor"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/complianceportal"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

func newNDADirective(
	logger *log.Logger,
	visitorSvc *visitor.Service,
	esignSvc *esign.Service,
) func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
	return func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
		identity := authn.IdentityFromContext(ctx)
		if identity == nil {
			return next(ctx)
		}

		compliancePage := complianceportal.CompliancePortalFromContext(ctx)
		if compliancePage == nil {
			logger.ErrorCtx(ctx, "cannot get compliance page from context")
			return nil, gqlutils.Internal(ctx)
		}

		membership, err := visitorSvc.GetPortalMembership(ctx, compliancePage.ID, identity.ID)
		if err != nil {
			logger.ErrorCtx(ctx, "cannot get compliance page membership", log.Error(err))
			return nil, gqlutils.Internal(ctx)
		}

		if membership.ElectronicSignatureID == nil {
			return next(ctx)
		}

		scope := coredata.NewScopeFromObjectID(compliancePage.OrganizationID)

		sig, err := esignSvc.GetSignatureByID(ctx, scope, *membership.ElectronicSignatureID)
		if err != nil {
			logger.ErrorCtx(ctx, "cannot get NDA signature", log.Error(err))
			return nil, gqlutils.Internal(ctx)
		}

		// We need full name before user signs NDA
		if identity.FullName == "" {
			return nil, gqlutils.FullNameRequiredf(ctx, "full name is required")
		}

		if sig.Status != coredata.ElectronicSignatureStatusCompleted {
			return nil, gqlutils.NDASignatureRequiredf(ctx, "NDA signature required")
		}

		return next(ctx)
	}
}
