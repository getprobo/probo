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

package auth

import (
	"context"

	"go.probo.inc/probo/pkg/coredata"
)

type (
	ctxKey struct{ name string }
)

var (
	trustIdentityContextKey = &ctxKey{name: "trust_identity"}
	trustSessionContextKey  = &ctxKey{name: "trust_session"}
	trustCompliancePageKey  = &ctxKey{name: "trust_compliance_page"}
	trustMembershipKey      = &ctxKey{name: "trust_membership"}
)

func TrustSessionFromContext(ctx context.Context) *coredata.Session {
	session, _ := ctx.Value(trustSessionContextKey).(*coredata.Session)
	return session
}

func TrustContextWithSession(ctx context.Context, session *coredata.Session) context.Context {
	return context.WithValue(ctx, trustSessionContextKey, session)
}

func TrustIdentityFromContext(ctx context.Context) *coredata.Identity {
	identity, _ := ctx.Value(trustIdentityContextKey).(*coredata.Identity)
	return identity
}

func TrustContextWithIdentity(ctx context.Context, identity *coredata.Identity) context.Context {
	return context.WithValue(ctx, trustIdentityContextKey, identity)
}

func TrustCompliancePageFromContext(ctx context.Context) *coredata.TrustCenter {
	page, _ := ctx.Value(trustCompliancePageKey).(*coredata.TrustCenter)
	return page
}

func TrustContextWithCompliancePage(ctx context.Context, page *coredata.TrustCenter) context.Context {
	return context.WithValue(ctx, trustCompliancePageKey, page)
}

func TrustMembershipFromContext(ctx context.Context) *coredata.TrustCenterAccess {
	membership, _ := ctx.Value(trustMembershipKey).(*coredata.TrustCenterAccess)
	return membership
}

func TrustContextWithMembership(ctx context.Context, membership *coredata.TrustCenterAccess) context.Context {
	return context.WithValue(ctx, trustMembershipKey, membership)
}
