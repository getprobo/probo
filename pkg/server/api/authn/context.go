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
	trustCenterKey     = &ctxKey{name: "trust_center"}
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
