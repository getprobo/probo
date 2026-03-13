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

package agent

import (
	"context"
	"fmt"
)

type runContextKey struct{}

func WithRunContext(ctx context.Context, val any) context.Context {
	return context.WithValue(ctx, runContextKey{}, val)
}

func RunContextFrom[C any](ctx context.Context) C {
	val := ctx.Value(runContextKey{})
	if val == nil {
		var zero C
		panic(fmt.Sprintf("agent: no run context found (expected %T)", zero))
	}

	typed, ok := val.(C)
	if !ok {
		var zero C
		panic(fmt.Sprintf("agent: run context type mismatch: stored %T, requested %T", val, zero))
	}

	return typed
}

func TryRunContextFrom[C any](ctx context.Context) (C, bool) {
	val := ctx.Value(runContextKey{})
	if val == nil {
		var zero C
		return zero, false
	}

	typed, ok := val.(C)
	return typed, ok
}
