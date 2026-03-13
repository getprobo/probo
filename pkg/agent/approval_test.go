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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/llm"
)

func TestBuildToolNameSet(t *testing.T) {
	t.Parallel()

	t.Run(
		"nil input returns nil",
		func(t *testing.T) {
			t.Parallel()

			assert.Nil(t, buildToolNameSet(nil))
		},
	)

	t.Run(
		"empty slice returns nil",
		func(t *testing.T) {
			t.Parallel()

			assert.Nil(t, buildToolNameSet([]string{}))
		},
	)

	t.Run(
		"single name",
		func(t *testing.T) {
			t.Parallel()

			set := buildToolNameSet([]string{"delete"})

			assert.Len(t, set, 1)
			_, ok := set["delete"]
			assert.True(t, ok)
		},
	)

	t.Run(
		"multiple names",
		func(t *testing.T) {
			t.Parallel()

			set := buildToolNameSet([]string{"delete", "update", "create"})

			assert.Len(t, set, 3)
			for _, name := range []string{"delete", "update", "create"} {
				_, ok := set[name]
				assert.True(t, ok, "expected set to contain %q", name)
			}
		},
	)

	t.Run(
		"duplicate names are deduplicated",
		func(t *testing.T) {
			t.Parallel()

			set := buildToolNameSet([]string{"delete", "delete", "update"})

			assert.Len(t, set, 2)
		},
	)
}

func TestApprovalConfig_RequiresApproval(t *testing.T) {
	t.Parallel()

	tc := llm.ToolCall{
		ID:       "tc_1",
		Function: llm.FunctionCall{Name: "delete_user", Arguments: `{}`},
	}

	t.Run(
		"nil config returns false",
		func(t *testing.T) {
			t.Parallel()

			var c *ApprovalConfig
			assert.False(t, c.requiresApproval(context.Background(), tc))
		},
	)

	t.Run(
		"empty config returns false",
		func(t *testing.T) {
			t.Parallel()

			c := &ApprovalConfig{}
			assert.False(t, c.requiresApproval(context.Background(), tc))
		},
	)

	t.Run(
		"tool name in set returns true",
		func(t *testing.T) {
			t.Parallel()

			c := &ApprovalConfig{
				toolNameSet: map[string]struct{}{
					"delete_user": {},
				},
			}

			assert.True(t, c.requiresApproval(context.Background(), tc))
		},
	)

	t.Run(
		"tool name not in set returns false",
		func(t *testing.T) {
			t.Parallel()

			c := &ApprovalConfig{
				toolNameSet: map[string]struct{}{
					"list_users": {},
				},
			}

			assert.False(t, c.requiresApproval(context.Background(), tc))
		},
	)

	t.Run(
		"ShouldApprove takes precedence over tool name set",
		func(t *testing.T) {
			t.Parallel()

			c := &ApprovalConfig{
				toolNameSet: map[string]struct{}{
					"delete_user": {},
				},
				ShouldApprove: func(_ context.Context, _ llm.ToolCall) bool {
					return false
				},
			}

			assert.False(t, c.requiresApproval(context.Background(), tc))
		},
	)

	t.Run(
		"ShouldApprove receives context and tool call",
		func(t *testing.T) {
			t.Parallel()

			type ctxKey struct{}
			ctx := context.WithValue(context.Background(), ctxKey{}, "marker")

			var capturedCtx context.Context
			var capturedTC llm.ToolCall

			c := &ApprovalConfig{
				ShouldApprove: func(ctx context.Context, tc llm.ToolCall) bool {
					capturedCtx = ctx
					capturedTC = tc
					return true
				},
			}

			result := c.requiresApproval(ctx, tc)

			assert.True(t, result)
			assert.Equal(t, "marker", capturedCtx.Value(ctxKey{}))
			assert.Equal(t, "delete_user", capturedTC.Function.Name)
		},
	)

	t.Run(
		"ShouldApprove returning true requires approval",
		func(t *testing.T) {
			t.Parallel()

			c := &ApprovalConfig{
				ShouldApprove: func(_ context.Context, _ llm.ToolCall) bool {
					return true
				},
			}

			assert.True(t, c.requiresApproval(context.Background(), tc))
		},
	)
}
