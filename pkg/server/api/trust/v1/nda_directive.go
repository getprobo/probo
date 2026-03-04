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

package trust_v1

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/compliancepage"
	"go.probo.inc/probo/pkg/server/gqlutils"
	"go.probo.inc/probo/pkg/trust"
)

func newNDADirective(
	logger *log.Logger,
	trustSvc *trust.Service,
	esignSvc *esign.Service,
) func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
	return func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
		identity := authn.IdentityFromContext(ctx)
		if identity == nil {
			return next(ctx)
		}

		compliancePage := compliancepage.CompliancePageFromContext(ctx)
		if compliancePage == nil {
			logger.ErrorCtx(ctx, "cannot get compliance page from context")
			return nil, gqlutils.Internal(ctx)
		}

		membership, err := trustSvc.GetMembershipByCompliancePageIDAndIdentityID(ctx, compliancePage.ID, identity.ID)
		if err != nil {
			if errors.Is(err, trust.ErrMembershipNotFound) {
				return nil, gqlutils.Unauthenticatedf(ctx, "authentication needed")
			}

			logger.ErrorCtx(ctx, "cannot get compliance page membership", log.Error(err))
			return nil, gqlutils.Internal(ctx)
		}

		if membership.ElectronicSignatureID == nil {
			return next(ctx)
		}

		sig, err := esignSvc.GetSignatureByID(ctx, *membership.ElectronicSignatureID)
		if err != nil {
			logger.ErrorCtx(ctx, "cannot get NDA signature", log.Error(err))
			return nil, gqlutils.Internal(ctx)
		}

		if sig.Status != coredata.ElectronicSignatureStatusCompleted {
			return nil, gqlutils.NDASignatureRequiredf(ctx, "NDA signature required")
		}

		return next(ctx)
	}
}
