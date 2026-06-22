// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/compliancepage"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

func (r *Resolver) resourceAliasForStorageResource(
	ctx context.Context,
	storageResourceID gid.GID,
) (*string, error) {
	trustCenter := compliancepage.CompliancePageFromContext(ctx)
	scope := coredata.NewScopeFromObjectID(trustCenter.ID)

	alias, err := r.resourceAlias.GetByResourceID(ctx, scope, storageResourceID)
	if err != nil {
		r.logger.ErrorCtx(ctx, "cannot load resource alias", log.Error(err))

		return nil, gqlutils.Internal(ctx)
	}

	return alias, nil
}
