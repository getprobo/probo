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

package ratelimit_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.probo.inc/probo/pkg/server/gqlx"
	"go.probo.inc/probo/pkg/server/gqlx/directives/ratelimit"
)

func TestDirective_WritesRateLimitHeaders(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	ctx := ratelimit.WithReport(gqlx.WithHTTPContext(
		context.Background(),
		rec,
		requestFrom("203.0.113.10"),
	))
	signUp := withField(ctx, "Mutation", "signUp", "signUp")
	signIn := withField(ctx, "Mutation", "signIn", "signIn")
	limiter := ratelimit.New(ratelimit.Budget{}, nil)
	next := func(context.Context) (any, error) {
		return "ok", nil
	}

	_, err := limiter.Directive(signUp, nil, next)
	require.NoError(t, err)
	assert.Equal(t, `"signUp";q=10;w=60`, rec.Header().Get("RateLimit-Policy"))
	assert.Equal(t, `"signUp";r=9;t=6`, rec.Header().Get("RateLimit"))
	assert.Empty(t, rec.Header().Get("Retry-After"))

	_, err = limiter.Directive(signIn, nil, next)
	require.NoError(t, err)
	assert.Equal(t, `"signIn";q=10;w=60, "signUp";q=10;w=60`, rec.Header().Get("RateLimit-Policy"))
	assert.Equal(t, `"signIn";r=9;t=6, "signUp";r=9;t=6`, rec.Header().Get("RateLimit"))
	assert.Empty(t, rec.Header().Get("Retry-After"))
}

func TestDirective_SetsRetryAfterWhenRejected(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	ctx := ratelimit.WithReport(fieldContextOn(rec, "203.0.113.10", "Mutation", "signUp", "signUp"))
	limiter := ratelimit.New(fallback(1, 60), nil)
	next := func(context.Context) (any, error) {
		return "ok", nil
	}

	_, err := limiter.Directive(ctx, nil, next)
	require.NoError(t, err)
	assert.Equal(t, `"signUp";r=0;t=60`, rec.Header().Get("RateLimit"))
	assert.Empty(t, rec.Header().Get("Retry-After"))

	_, err = limiter.Directive(ctx, nil, next)
	require.Error(t, err)
	assert.Equal(t, "RATE_LIMITED", errorCode(t, err))
	assert.Equal(t, `"signUp";q=1;w=60`, rec.Header().Get("RateLimit-Policy"))
	assert.Equal(t, `"signUp";r=0;t=60`, rec.Header().Get("RateLimit"))
	assert.Equal(t, "60", rec.Header().Get("Retry-After"))
}

func TestDirective_StopsAfterLimit(t *testing.T) {
	t.Parallel()

	limiter := ratelimit.New(fallback(2, 60), nil)
	ctx := fieldContext("203.0.113.10", "Mutation", "signUp", "signUp")
	calls := 0
	next := func(context.Context) (any, error) {
		calls++

		return "ok", nil
	}

	for range 2 {
		result, err := limiter.Directive(ctx, nil, next)
		require.NoError(t, err)
		assert.Equal(t, "ok", result)
	}

	_, err := limiter.Directive(ctx, nil, next)
	require.Error(t, err)
	assert.Equal(t, "RATE_LIMITED", errorCode(t, err))
	assert.Equal(t, 2, calls)
}

func TestDirective_SeparatesIPAndField(t *testing.T) {
	t.Parallel()

	limiter := ratelimit.New(fallback(1, 60), nil)
	signUp := fieldContext("203.0.113.10", "Mutation", "signUp", "signUp")
	otherIP := fieldContext("203.0.113.11", "Mutation", "signUp", "signUp")
	signIn := fieldContext("203.0.113.10", "Mutation", "signIn", "signIn")
	next := func(context.Context) (any, error) {
		return "ok", nil
	}

	_, err := limiter.Directive(signUp, nil, next)
	require.NoError(t, err)

	_, err = limiter.Directive(signUp, nil, next)
	require.Error(t, err)
	assert.Equal(t, "RATE_LIMITED", errorCode(t, err))

	_, err = limiter.Directive(otherIP, nil, next)
	require.NoError(t, err)

	_, err = limiter.Directive(signIn, nil, next)
	assert.NoError(t, err)
}

func TestDirective_CountsEachAlias(t *testing.T) {
	t.Parallel()

	limiter := ratelimit.New(fallback(2, 60), nil)
	first := fieldContext("203.0.113.10", "Mutation", "signUp", "registered")
	second := fieldContext("203.0.113.10", "Mutation", "signUp", "fresh")
	calls := 0
	next := func(context.Context) (any, error) {
		calls++

		return "ok", nil
	}

	_, err := limiter.Directive(first, nil, next)
	require.NoError(t, err)

	_, err = limiter.Directive(second, nil, next)
	require.NoError(t, err)

	_, err = limiter.Directive(first, nil, next)
	require.Error(t, err)
	assert.Equal(t, "RATE_LIMITED", errorCode(t, err))
	assert.Equal(t, 2, calls)
}

func TestDirective_ReplacesConfiguredBudget(t *testing.T) {
	t.Parallel()

	limiter := ratelimit.New(fallback(2, 60), map[string]ratelimit.Budget{
		"signUp": {Limit: 1, Window: 30},
	})
	signUp := fieldContext("203.0.113.10", "Mutation", "signUp", "signUp")
	signIn := fieldContext("203.0.113.10", "Mutation", "signIn", "signIn")
	next := func(context.Context) (any, error) {
		return "ok", nil
	}

	_, err := limiter.Directive(signUp, nil, next)
	require.NoError(t, err)

	_, err = limiter.Directive(signUp, nil, next)
	require.Error(t, err)
	assert.Equal(t, "RATE_LIMITED", errorCode(t, err))

	for range 2 {
		_, err = limiter.Directive(signIn, nil, next)
		require.NoError(t, err)
	}

	_, err = limiter.Directive(signIn, nil, next)
	require.Error(t, err)
	assert.Equal(t, "RATE_LIMITED", errorCode(t, err))
}

func fallback(limit, window int) ratelimit.Budget {
	return ratelimit.Budget{Limit: limit, Window: window}
}

func fieldContext(ip, object, field, alias string) context.Context {
	return fieldContextOn(httptest.NewRecorder(), ip, object, field, alias)
}

func fieldContextOn(rec *httptest.ResponseRecorder, ip, object, field, alias string) context.Context {
	ctx := gqlx.WithHTTPContext(context.Background(), rec, requestFrom(ip))

	return withField(ctx, object, field, alias)
}

func requestFrom(ip string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "http://example.test/graphql", nil)
	req.RemoteAddr = ip + ":1234"

	return req
}

func withField(ctx context.Context, object, field, alias string) context.Context {
	return graphql.WithFieldContext(ctx, &graphql.FieldContext{
		Object: object,
		Field: graphql.CollectedField{
			Field: &ast.Field{
				Name:  field,
				Alias: alias,
			},
		},
	})
}

func errorCode(t *testing.T, err error) string {
	t.Helper()

	var gqlErr *gqlerror.Error
	require.ErrorAs(t, err, &gqlErr)

	code, _ := gqlErr.Extensions["code"].(string)

	return code
}
