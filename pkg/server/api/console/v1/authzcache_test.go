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
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/authz"
)

const (
	authzTestEntityType uint16 = 7 // arbitrary; mirrors existing entity types
)

// spyAuthorize wraps an authz.AuthorizeFunc with a per-key call counter.
// Tests use it to assert that the request-scoped cache short-circuits
// repeated lookups and to confirm that the underlying authorize is
// dispatched on cache miss.
type spyAuthorize struct {
	mu       sync.Mutex
	calls    map[string]int
	allow    func(resource gid.GID, action string) bool
	totalHit atomic.Int64
}

func newSpyAuthorize(allow func(resource gid.GID, action string) bool) *spyAuthorize {
	return &spyAuthorize{
		calls: make(map[string]int),
		allow: allow,
	}
}

func (s *spyAuthorize) fn() authz.AuthorizeFunc {
	return func(_ context.Context, id gid.GID, action string, _ ...authz.AuthorizeFuncOption) error {
		s.totalHit.Add(1)
		s.mu.Lock()
		s.calls[id.String()+"|"+action]++
		s.mu.Unlock()
		if s.allow == nil || s.allow(id, action) {
			return nil
		}
		return errors.New("not allowed")
	}
}

func (s *spyAuthorize) total() int64 {
	return s.totalHit.Load()
}

// withCache returns a context carrying a fresh authzCache (mimics the
// AuthzCacheMiddleware behaviour for direct unit testing).
func withCache(ctx context.Context) context.Context {
	return context.WithValue(ctx, authzCacheCtxKey{}, newAuthzCache())
}

// withIdentity attaches a synthetic identity so subjectKey returns a
// stable, distinct value per subject id.
func withIdentity(ctx context.Context, id gid.GID) context.Context {
	return authn.ContextWithIdentity(ctx, &coredata.Identity{ID: id})
}

func newGID(t *testing.T) gid.GID {
	t.Helper()
	return gid.New(gid.NewTenantID(), authzTestEntityType)
}

func TestAuthorizeCached_CacheHitDoesNotInvokeAuthorize(t *testing.T) {
	t.Parallel()

	subject := newGID(t)
	resource := newGID(t)
	action := "core:cloud-account:get"

	spy := newSpyAuthorize(func(_ gid.GID, _ string) bool { return true })
	r := &Resolver{authorize: spy.fn()}

	ctx := withCache(withIdentity(context.Background(), subject))

	got1 := r.authorizeCached(ctx, resource, action)
	got2 := r.authorizeCached(ctx, resource, action)

	assert.True(t, got1)
	assert.True(t, got2)
	assert.Equal(t, int64(1), spy.total(), "authorize must be invoked once across the two calls; the second is a cache hit")
}

func TestAuthorizeCached_CacheMissCallsAuthorizeAndStoresResult(t *testing.T) {
	t.Parallel()

	subject := newGID(t)
	resource := newGID(t)
	action := "core:cloud-account:list"

	// Authorizer denies; cache must record the denial as well.
	spy := newSpyAuthorize(func(_ gid.GID, _ string) bool { return false })
	r := &Resolver{authorize: spy.fn()}

	ctx := withCache(withIdentity(context.Background(), subject))

	got := r.authorizeCached(ctx, resource, action)
	assert.False(t, got)

	got2 := r.authorizeCached(ctx, resource, action)
	assert.False(t, got2, "cached denial must be returned without re-invoking authorize")
	assert.Equal(t, int64(1), spy.total(), "denial outcome must also be cached")
}

func TestAuthorizeCached_CompositeKey(t *testing.T) {
	t.Parallel()

	subjectA := newGID(t)
	subjectB := newGID(t)
	resourceA := newGID(t)
	resourceB := newGID(t)

	t.Run("different subjects on the same (resource, action) do NOT share a hit", func(t *testing.T) {
		t.Parallel()

		spy := newSpyAuthorize(func(_ gid.GID, _ string) bool { return true })
		r := &Resolver{authorize: spy.fn()}
		action := "core:cloud-account:get"

		ctxA := withCache(withIdentity(context.Background(), subjectA))
		ctxB := withCache(withIdentity(context.Background(), subjectB))

		_ = r.authorizeCached(ctxA, resourceA, action)
		_ = r.authorizeCached(ctxB, resourceA, action)

		assert.Equal(t, int64(2), spy.total(), "different subjects must each trigger an authorize call")
	})

	t.Run("different resources on the same (subject, action) do NOT share a hit", func(t *testing.T) {
		t.Parallel()

		spy := newSpyAuthorize(func(_ gid.GID, _ string) bool { return true })
		r := &Resolver{authorize: spy.fn()}
		action := "core:cloud-account:get"

		ctx := withCache(withIdentity(context.Background(), subjectA))

		_ = r.authorizeCached(ctx, resourceA, action)
		_ = r.authorizeCached(ctx, resourceB, action)

		assert.Equal(t, int64(2), spy.total(), "different resources must each trigger an authorize call")
	})

	t.Run("different actions on the same (subject, resource) do NOT share a hit", func(t *testing.T) {
		t.Parallel()

		spy := newSpyAuthorize(func(_ gid.GID, _ string) bool { return true })
		r := &Resolver{authorize: spy.fn()}

		ctx := withCache(withIdentity(context.Background(), subjectA))

		_ = r.authorizeCached(ctx, resourceA, "core:cloud-account:get")
		_ = r.authorizeCached(ctx, resourceA, "core:cloud-account:list")

		assert.Equal(t, int64(2), spy.total(), "different actions must each trigger an authorize call")
	})
}

func TestAuthzCache_PerRequestIsolation(t *testing.T) {
	t.Parallel()

	a := newAuthzCache()
	b := newAuthzCache()
	require.NotSame(t, a, b, "newAuthzCache must yield a fresh instance every call")

	// Mutate one and assert the other is unaffected (per-request isolation
	// is the whole point of the per-request cache instance).
	a.mu.Lock()
	a.results["k"] = true
	a.mu.Unlock()

	b.mu.Lock()
	_, hit := b.results["k"]
	b.mu.Unlock()
	assert.False(t, hit, "the second cache instance must not share state with the first")
}

func TestAuthzCacheMiddleware_InjectsFreshCachePerRequest(t *testing.T) {
	t.Parallel()

	var firstCache, secondCache *authzCache
	mw := AuthzCacheMiddleware()
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := authzCacheFromContext(r.Context())
		require.NotNil(t, c)
		if firstCache == nil {
			firstCache = c
		} else {
			secondCache = c
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec1 := httptest.NewRecorder()
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, httptest.NewRequest(http.MethodGet, "/", nil))
	handler.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/", nil))

	require.NotNil(t, firstCache)
	require.NotNil(t, secondCache)
	assert.NotSame(t, firstCache, secondCache, "each request must get its own authzCache instance")
}

func TestAuthorizeCached_ConcurrentReadsAreSafe(t *testing.T) {
	t.Parallel()

	subject := newGID(t)
	resource := newGID(t)
	action := "core:cloud-account:get"

	spy := newSpyAuthorize(func(_ gid.GID, _ string) bool { return true })
	r := &Resolver{authorize: spy.fn()}

	ctx := withCache(withIdentity(context.Background(), subject))

	// Pre-warm the cache so the goroutines are exercising the read
	// fast-path under contention.
	require.True(t, r.authorizeCached(ctx, resource, action))

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)
	results := make([]bool, goroutines)
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			results[i] = r.authorizeCached(ctx, resource, action)
		}()
	}
	wg.Wait()

	for i, ok := range results {
		assert.Truef(t, ok, "goroutine %d must observe the cached allow result", i)
	}
	assert.Equal(t, int64(1), spy.total(), "cache pre-warm + 100 concurrent reads must yield exactly one underlying authorize call")
}

// TestAuthorizeCached_NilCacheStillCalls verifies that when no cache
// is present in the context (e.g. a code path that bypassed the
// middleware -- typically mutation paths) the resolver still
// dispatches to the underlying authorize. The test mirrors the
// "mutations bypass the cache" invariant by exercising the
// no-cache-in-context branch.
func TestAuthorizeCached_NilCacheStillCalls(t *testing.T) {
	t.Parallel()

	subject := newGID(t)
	resource := newGID(t)
	action := "core:cloud-account:get"

	spy := newSpyAuthorize(func(_ gid.GID, _ string) bool { return true })
	r := &Resolver{authorize: spy.fn()}

	// Identity present but NO authzCache in the context.
	ctx := withIdentity(context.Background(), subject)

	_ = r.authorizeCached(ctx, resource, action)
	_ = r.authorizeCached(ctx, resource, action)

	assert.Equal(t, int64(2), spy.total(), "without a cache in context, every call must reach authorize")
}
