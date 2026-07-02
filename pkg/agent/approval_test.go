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

			var (
				capturedCtx context.Context
				capturedTC  llm.ToolCall
			)

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
