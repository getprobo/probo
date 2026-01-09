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

package authn

import (
	"context"

	"go.probo.inc/probo/pkg/coredata"
)

type (
	ctxKey struct{ name string }
)

var (
	identityContextKey = &ctxKey{name: "identity"}
	sessionContextKey  = &ctxKey{name: "session"}
	apiKeyContextKey   = &ctxKey{name: "api_key"}
	TrustCenterKey     = &ctxKey{name: "trust_center"}
)

func SessionFromContext(ctx context.Context) *coredata.Session {
	session, _ := ctx.Value(sessionContextKey).(*coredata.Session)
	return session
}

func ContextWithSession(ctx context.Context, session *coredata.Session) context.Context {
	return context.WithValue(ctx, sessionContextKey, session)
}

func IdentityFromContext(ctx context.Context) *coredata.Identity {
	identity, _ := ctx.Value(identityContextKey).(*coredata.Identity)
	return identity
}

func ContextWithIdentity(ctx context.Context, identity *coredata.Identity) context.Context {
	return context.WithValue(ctx, identityContextKey, identity)
}

func APIKeyFromContext(ctx context.Context) *coredata.PersonalAPIKey {
	apiKey, _ := ctx.Value(apiKeyContextKey).(*coredata.PersonalAPIKey)
	return apiKey
}

func ContextWithAPIKey(ctx context.Context, apiKey *coredata.PersonalAPIKey) context.Context {
	return context.WithValue(ctx, apiKeyContextKey, apiKey)
}
