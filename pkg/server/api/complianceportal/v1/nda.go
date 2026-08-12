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
	"errors"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/complianceportal"
	"go.probo.inc/probo/pkg/server/gqlutils"
	"go.probo.inc/probo/pkg/validator"
)

// mapNDASigningValidationError maps esign full-name validation failures to
// FULL_NAME_REQUIRED so the portal gate redirect keeps working. Other
// validation errors become InvalidValidationErrors. Returns nil when err is
// not a ValidationErrors value.
func mapNDASigningValidationError(ctx context.Context, err error) error {
	validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
	if !ok {
		return nil
	}

	if len(validationErrors.ByField("signer_full_name")) > 0 ||
		len(validationErrors.ByField("actor_full_name")) > 0 {
		return gqlutils.FullNameRequiredf(ctx, "full name is required")
	}

	return gqlutils.InvalidValidationErrors(ctx, validationErrors)
}

// requireCompletedNDA enforces portal NDA completion for the signed-in identity.
// No-ops when there is no viewer or the portal membership has no NDA signature.
// Call from protected (non-PUBLIC) export resolvers after authentication.
// Access-request mutations must not call this — requesting is allowed before signing.
func (r *Resolver) requireCompletedNDA(ctx context.Context) error {
	identity := authn.IdentityFromContext(ctx)
	if identity == nil {
		return nil
	}

	compliancePage := complianceportal.CompliancePortalFromContext(ctx)
	if compliancePage == nil {
		r.logger.ErrorCtx(ctx, "cannot get compliance page from context")
		return gqlutils.Internal(ctx)
	}

	membership, err := r.visitor.GetPortalMembership(ctx, compliancePage.ID, identity.ID)
	if err != nil {
		r.logger.ErrorCtx(ctx, "cannot get compliance page membership", log.Error(err))
		return gqlutils.Internal(ctx)
	}

	if membership.ElectronicSignatureID == nil {
		return nil
	}

	scope := coredata.NewScopeFromObjectID(compliancePage.OrganizationID)

	sig, err := r.esign.GetSignatureByID(ctx, scope, *membership.ElectronicSignatureID)
	if err != nil {
		r.logger.ErrorCtx(ctx, "cannot get NDA signature", log.Error(err))
		return gqlutils.Internal(ctx)
	}

	// Full name before NDA so export paths send users to /full-name first.
	// Accept maps esign signer_full_name validation to the same gate code.
	if identity.FullName == "" {
		return gqlutils.FullNameRequiredf(ctx, "full name is required")
	}

	if sig.Status != coredata.ElectronicSignatureStatusCompleted {
		return gqlutils.NDASignatureRequiredf(ctx, "NDA signature required")
	}

	return nil
}
