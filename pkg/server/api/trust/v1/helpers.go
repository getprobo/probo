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

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

func (r *mutationResolver) checkNDASignature(ctx context.Context, trustCenter *coredata.TrustCenter, identity *coredata.Identity) error {
	if trustCenter.NonDisclosureAgreementFileID == nil {
		return nil
	}

	trustService := r.TrustService(ctx, trustCenter.ID.TenantID())

	access, err := trustService.TrustCenterAccesses.GetAccess(ctx, trustCenter.ID, identity.EmailAddress)
	if err != nil {
		return gqlutils.Forbiddenf(ctx, "NDA signature required")
	}

	if access.ElectronicSignatureID == nil {
		return gqlutils.Forbiddenf(ctx, "NDA signature required")
	}

	sig, err := r.esign.GetSignatureByID(ctx, *access.ElectronicSignatureID)
	if err != nil {
		return gqlutils.Forbiddenf(ctx, "NDA signature required")
	}

	if sig.Status != coredata.ElectronicSignatureStatusCompleted {
		return gqlutils.Forbiddenf(ctx, "NDA signature required")
	}

	return nil
}
