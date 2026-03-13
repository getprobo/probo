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

package agent_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/agent"
)

type testRunContext struct {
	UserID string
	Locale string
}

func TestWithRunContext(t *testing.T) {
	t.Parallel()

	t.Run(
		"stores value retrievable by RunContextFrom",
		func(t *testing.T) {
			t.Parallel()

			rc := testRunContext{UserID: "u_123", Locale: "en"}
			ctx := agent.WithRunContext(context.Background(), rc)

			got := agent.RunContextFrom[testRunContext](ctx)
			assert.Equal(t, "u_123", got.UserID)
			assert.Equal(t, "en", got.Locale)
		},
	)

	t.Run(
		"overwrites previous run context",
		func(t *testing.T) {
			t.Parallel()

			ctx := agent.WithRunContext(context.Background(), testRunContext{UserID: "first"})
			ctx = agent.WithRunContext(ctx, testRunContext{UserID: "second"})

			got := agent.RunContextFrom[testRunContext](ctx)
			assert.Equal(t, "second", got.UserID)
		},
	)
}

func TestRunContextFrom(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns stored value with correct type",
		func(t *testing.T) {
			t.Parallel()

			ctx := agent.WithRunContext(context.Background(), "hello")

			got := agent.RunContextFrom[string](ctx)
			assert.Equal(t, "hello", got)
		},
	)

	t.Run(
		"panics when no run context is set",
		func(t *testing.T) {
			t.Parallel()

			require.Panics(
				t,
				func() {
					agent.RunContextFrom[string](context.Background())
				},
			)
		},
	)

	t.Run(
		"panic message includes expected type when missing",
		func(t *testing.T) {
			t.Parallel()

			assert.PanicsWithValue(
				t,
				"agent: no run context found (expected string)",
				func() {
					agent.RunContextFrom[string](context.Background())
				},
			)
		},
	)

	t.Run(
		"panics on type mismatch",
		func(t *testing.T) {
			t.Parallel()

			ctx := agent.WithRunContext(context.Background(), 42)

			require.Panics(
				t,
				func() {
					agent.RunContextFrom[string](ctx)
				},
			)
		},
	)

	t.Run(
		"panic message includes both types on mismatch",
		func(t *testing.T) {
			t.Parallel()

			ctx := agent.WithRunContext(context.Background(), 42)

			assert.PanicsWithValue(
				t,
				"agent: run context type mismatch: stored int, requested string",
				func() {
					agent.RunContextFrom[string](ctx)
				},
			)
		},
	)
}

func TestTryRunContextFrom(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns value and true when context is set",
		func(t *testing.T) {
			t.Parallel()

			rc := testRunContext{UserID: "u_456", Locale: "fr"}
			ctx := agent.WithRunContext(context.Background(), rc)

			got, ok := agent.TryRunContextFrom[testRunContext](ctx)

			assert.True(t, ok)
			assert.Equal(t, "u_456", got.UserID)
			assert.Equal(t, "fr", got.Locale)
		},
	)

	t.Run(
		"returns zero value and false when no context is set",
		func(t *testing.T) {
			t.Parallel()

			got, ok := agent.TryRunContextFrom[testRunContext](context.Background())

			assert.False(t, ok)
			assert.Equal(t, testRunContext{}, got)
		},
	)

	t.Run(
		"returns zero value and false on type mismatch",
		func(t *testing.T) {
			t.Parallel()

			ctx := agent.WithRunContext(context.Background(), 42)

			got, ok := agent.TryRunContextFrom[string](ctx)

			assert.False(t, ok)
			assert.Equal(t, "", got)
		},
	)

	t.Run(
		"works with interface types",
		func(t *testing.T) {
			t.Parallel()

			ctx := agent.WithRunContext(context.Background(), "a string value")

			got, ok := agent.TryRunContextFrom[any](ctx)

			assert.True(t, ok)
			assert.Equal(t, "a string value", got)
		},
	)
}
