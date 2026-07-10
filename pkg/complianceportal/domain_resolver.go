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

package complianceportal

import (
	"context"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/complianceportal/resolver"
	"go.probo.inc/probo/pkg/coredata"
)

// EffectiveDomainForTrustCenter returns the domain a compliance page is served
// under: the custom domain when it has an active certificate, otherwise the
// default subdomain when its certificate is active. It returns nil when no
// serving domain is available yet.
//
// It is the single certificate-status-aware domain resolver shared by the
// management and visitor sub-packages.
func EffectiveDomainForTrustCenter(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	trustCenter *coredata.TrustCenter,
) (*coredata.CustomDomain, error) {
	return resolver.EffectiveDomainForTrustCenter(ctx, conn, scope, trustCenter)
}

// PublicURLForTrustCenter returns the canonical public URL of a compliance
// page on its dedicated domain.
func PublicURLForTrustCenter(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	trustCenter *coredata.TrustCenter,
	baseDomain string,
) (string, error) {
	return resolver.PublicURLForTrustCenter(ctx, conn, scope, trustCenter, baseDomain)
}
