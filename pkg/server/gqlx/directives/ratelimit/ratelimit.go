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

package ratelimit

import (
	"context"
	"fmt"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"go.probo.inc/probo/pkg/server/api/clientip"
	"go.probo.inc/probo/pkg/server/gqlx"
	"golang.org/x/time/rate"
)

type (
	Budget struct {
		Limit  int `json:"limit"`
		Window int `json:"window"`
	}

	Limiter struct {
		fallback Budget
		budgets  map[string]Budget
		mu       sync.Mutex
		buckets  map[string]*bucket
		sweptAt  time.Time
	}

	bucket struct {
		limiter *rate.Limiter
		window  time.Duration
		last    time.Time
	}

	policy struct {
		limit     int
		window    int
		remaining int
		reset     int
		rejected  bool
	}

	report struct {
		mu       sync.Mutex
		policies map[string]policy
	}
)

type reportKey struct{}

// WithReport collects the rate-limit headers for one HTTP response.
func WithReport(ctx context.Context) context.Context {
	return context.WithValue(ctx, reportKey{}, &report{
		policies: map[string]policy{},
	})
}

// New returns a process-local token-bucket limiter. fallback applies when
// budgets has no entry for the field. budgets is keyed by the field name.
// A zero fallback is 10 requests per 60 seconds.
func New(fallback Budget, budgets map[string]Budget) *Limiter {
	if fallback.Limit < 1 || fallback.Window < 1 {
		fallback = Budget{Limit: 10, Window: 60}
	}

	copied := make(map[string]Budget, len(budgets))
	maps.Copy(copied, budgets)

	return &Limiter{
		fallback: fallback,
		budgets:  copied,
		buckets:  map[string]*bucket{},
	}
}

// Directive rejects the field once this client IP has spent the token bucket
// for the field. A configured budget replaces the global one. It does not
// call next when the bucket is empty.
func (l *Limiter) Directive(
	ctx context.Context,
	obj any,
	next graphql.Resolver,
) (any, error) {
	name := policyName(ctx)

	budget := l.fallback
	if configured, ok := l.budgets[name]; ok {
		budget = configured
	}

	if budget.Limit < 1 || budget.Window < 1 {
		return nil, gqlx.RateLimited(ctx)
	}

	req := gqlx.HTTPRequestFromContext(ctx)

	key := bucketKey(ctx, clientip.Extract(req))
	window := time.Duration(budget.Window) * time.Second
	ok, remaining, reset := l.allow(key, budget.Limit, window)
	writeReport(ctx, name, budget, remaining, reset, !ok)

	if !ok {
		return nil, gqlx.RateLimited(ctx)
	}

	return next(ctx)
}

func (l *Limiter) allow(key string, limit int, window time.Duration) (bool, int, int) {
	now := time.Now()

	l.mu.Lock()
	l.dropIdle(now)

	current := l.buckets[key]
	if current == nil {
		current = &bucket{
			limiter: rate.NewLimiter(rate.Limit(float64(limit)/window.Seconds()), limit),
			window:  window,
		}
		l.buckets[key] = current
	}

	current.last = now
	lim := current.limiter
	l.mu.Unlock()

	ok := lim.AllowN(now, 1)
	remaining, reset := remainingQuota(lim.TokensAt(now), limit, window)

	return ok, remaining, reset
}

func remainingQuota(tokens float64, limit int, window time.Duration) (int, int) {
	if tokens < 0 {
		tokens = 0
	}

	if tokens > float64(limit) {
		tokens = float64(limit)
	}

	remaining := int(math.Floor(tokens))

	deficit := float64(limit) - tokens
	if deficit <= 0 || limit < 1 || window <= 0 {
		return remaining, 0
	}

	seconds := math.Ceil(deficit * window.Seconds() / float64(limit))

	return remaining, int(seconds)
}

// dropIdle removes buckets that have been silent for a full window. By then
// the bucket would be full again, so the next request can start a new one.
// The scan runs at most once a minute.
func (l *Limiter) dropIdle(now time.Time) {
	if l.sweptAt.IsZero() {
		l.sweptAt = now

		return
	}

	if now.Sub(l.sweptAt) < time.Minute {
		return
	}

	l.sweptAt = now

	for key, current := range l.buckets {
		if now.Sub(current.last) >= current.window {
			delete(l.buckets, key)
		}
	}
}

func bucketKey(ctx context.Context, ip string) string {
	object := "unknown"
	field := "unknown"

	if fc := graphql.GetFieldContext(ctx); fc != nil {
		if fc.Object != "" {
			object = fc.Object
		}

		if fc.Field.Name != "" {
			field = fc.Field.Name
		}
	}

	return ip + "\x00" + object + "." + field
}

func policyName(ctx context.Context) string {
	field := "unknown"
	if fc := graphql.GetFieldContext(ctx); fc != nil && fc.Field.Name != "" {
		field = fc.Field.Name
	}

	return field
}

func writeReport(
	ctx context.Context,
	name string,
	budget Budget,
	remaining int,
	reset int,
	rejected bool,
) {
	current, _ := ctx.Value(reportKey{}).(*report)
	if current == nil {
		return
	}

	current.mu.Lock()
	defer current.mu.Unlock()

	current.policies[name] = policy{
		limit:     budget.Limit,
		window:    budget.Window,
		remaining: remaining,
		reset:     reset,
		rejected:  rejected,
	}

	names := slices.Sorted(maps.Keys(current.policies))
	policyHeader := make([]string, 0, len(names))
	limitHeader := make([]string, 0, len(names))
	retryAfter := 0

	for _, policyName := range names {
		item := current.policies[policyName]
		quoted := strconv.Quote(policyName)
		policyHeader = append(policyHeader, fmt.Sprintf("%s;q=%d;w=%d", quoted, item.limit, item.window))
		limitHeader = append(limitHeader, fmt.Sprintf("%s;r=%d;t=%d", quoted, item.remaining, item.reset))

		if item.rejected && item.reset > retryAfter {
			retryAfter = item.reset
		}
	}

	writer := gqlx.HTTPResponseWriterFromContext(ctx)
	header := writer.Header()
	header.Set("RateLimit-Policy", strings.Join(policyHeader, ", "))
	header.Set("RateLimit", strings.Join(limitHeader, ", "))

	if retryAfter > 0 {
		header.Set("Retry-After", strconv.Itoa(retryAfter))

		return
	}

	header.Del("Retry-After")
}
