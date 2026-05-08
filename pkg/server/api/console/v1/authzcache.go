// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package console_v1

import (
	"context"
	"net/http"
	"sync"

	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/authz"
)

// authzCache is a request-scoped dry-run authorization cache. It is
// keyed on (subject identity, resource id, action) and holds boolean
// allow decisions. Used by field-level resolvers (e.g. the
// permission(action: ...) resolver and the lastProbeError /
// scope.identifier OWNER/ADMIN gating) to render 25-row pages
// without 75 sequential pg.WithTx round-trips.
//
// CRITICAL: this cache is dry-run only. Mutation resolvers MUST call
// r.authorize(...) directly so the Authorizer records the audit-log
// entry; reading from the cache there would silently drop the audit
// trail.
type authzCache struct {
	mu      sync.Mutex
	results map[string]bool
}

type authzCacheCtxKey struct{}

// newAuthzCache returns a fresh empty cache. Used per HTTP request.
func newAuthzCache() *authzCache {
	return &authzCache{results: make(map[string]bool)}
}

// AuthzCacheMiddleware injects a fresh authzCache into the request
// context. The cache lives for the duration of the HTTP request and
// is GC'd afterwards.
func AuthzCacheMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), authzCacheCtxKey{}, newAuthzCache())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func authzCacheFromContext(ctx context.Context) *authzCache {
	cache, _ := ctx.Value(authzCacheCtxKey{}).(*authzCache)
	return cache
}

// authorizeCached returns the cached dry-run decision for (subject,
// resourceID, action). On cache miss it delegates to the supplied
// authorize closure (which MUST be a dry-run authorize to avoid
// emitting audit-log entries) and stores the result. Returns
// (allowed bool) -- the underlying error from authorize is treated
// as "not allowed" for the field-gating use case.
func (r *Resolver) authorizeCached(
	ctx context.Context,
	resourceID gid.GID,
	action string,
) bool {
	cache := authzCacheFromContext(ctx)
	subject := subjectKey(ctx)
	key := subject + "|" + resourceID.String() + "|" + action

	if cache != nil {
		cache.mu.Lock()
		if v, ok := cache.results[key]; ok {
			cache.mu.Unlock()
			return v
		}
		cache.mu.Unlock()
	}

	allowed := r.authorize(ctx, resourceID, action, authz.WithDryRun()) == nil

	if cache != nil {
		cache.mu.Lock()
		cache.results[key] = allowed
		cache.mu.Unlock()
	}

	return allowed
}

// subjectKey extracts a stable identity for the current session. The
// authn middleware exposes the membership / API-token identity; we
// fall back to "anonymous" when no session is present (the
// underlying authorize call will short-circuit to "not allowed" and
// the cache hit is harmless).
func subjectKey(ctx context.Context) string {
	if identity := authn.IdentityFromContext(ctx); identity != nil {
		return "identity:" + identity.ID.String()
	}
	if apiKey := authn.APIKeyFromContext(ctx); apiKey != nil {
		return "api-key:" + apiKey.ID.String()
	}
	return "anonymous"
}
