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
