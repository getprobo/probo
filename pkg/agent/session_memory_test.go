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
	"go.probo.inc/probo/pkg/llm"
)

func TestNewMemorySession(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns non-nil session",
		func(t *testing.T) {
			t.Parallel()

			s := agent.NewMemorySession()
			assert.NotNil(t, s)
		},
	)
}

func TestMemorySession_Load(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns nil for unknown session ID",
		func(t *testing.T) {
			t.Parallel()

			s := agent.NewMemorySession()

			msgs, err := s.Load(context.Background(), "unknown")

			require.NoError(t, err)
			assert.Nil(t, msgs)
		},
	)

	t.Run(
		"returns saved messages",
		func(t *testing.T) {
			t.Parallel()

			s := agent.NewMemorySession()
			messages := []llm.Message{
				{
					Role:  llm.RoleUser,
					Parts: []llm.Part{llm.TextPart{Text: "hello"}},
				},
				{
					Role:  llm.RoleAssistant,
					Parts: []llm.Part{llm.TextPart{Text: "hi there"}},
				},
			}

			err := s.Save(context.Background(), "sess-1", messages)
			require.NoError(t, err)

			loaded, err := s.Load(context.Background(), "sess-1")

			require.NoError(t, err)
			require.Len(t, loaded, 2)
			assert.Equal(t, llm.RoleUser, loaded[0].Role)
			assert.Equal(t, "hello", loaded[0].Text())
			assert.Equal(t, llm.RoleAssistant, loaded[1].Role)
			assert.Equal(t, "hi there", loaded[1].Text())
		},
	)

	t.Run(
		"returns a defensive copy of messages",
		func(t *testing.T) {
			t.Parallel()

			s := agent.NewMemorySession()
			messages := []llm.Message{
				{
					Role:  llm.RoleUser,
					Parts: []llm.Part{llm.TextPart{Text: "original"}},
				},
			}

			err := s.Save(context.Background(), "sess-copy", messages)
			require.NoError(t, err)

			loaded, err := s.Load(context.Background(), "sess-copy")
			require.NoError(t, err)

			loaded[0].Parts = []llm.Part{llm.TextPart{Text: "mutated"}}

			reloaded, err := s.Load(context.Background(), "sess-copy")
			require.NoError(t, err)

			assert.Equal(t, "original", reloaded[0].Text())
		},
	)

	t.Run(
		"different session IDs are independent",
		func(t *testing.T) {
			t.Parallel()

			s := agent.NewMemorySession()

			err := s.Save(
				context.Background(),
				"sess-a",
				[]llm.Message{
					{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "alpha"}}},
				},
			)
			require.NoError(t, err)

			err = s.Save(
				context.Background(),
				"sess-b",
				[]llm.Message{
					{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "beta"}}},
				},
			)
			require.NoError(t, err)

			a, err := s.Load(context.Background(), "sess-a")
			require.NoError(t, err)
			require.Len(t, a, 1)
			assert.Equal(t, "alpha", a[0].Text())

			b, err := s.Load(context.Background(), "sess-b")
			require.NoError(t, err)
			require.Len(t, b, 1)
			assert.Equal(t, "beta", b[0].Text())
		},
	)
}

func TestMemorySession_Save(t *testing.T) {
	t.Parallel()

	t.Run(
		"stores a defensive copy of input messages",
		func(t *testing.T) {
			t.Parallel()

			s := agent.NewMemorySession()
			messages := []llm.Message{
				{
					Role:  llm.RoleUser,
					Parts: []llm.Part{llm.TextPart{Text: "before"}},
				},
			}

			err := s.Save(context.Background(), "sess-def", messages)
			require.NoError(t, err)

			messages[0].Parts = []llm.Part{llm.TextPart{Text: "after"}}

			loaded, err := s.Load(context.Background(), "sess-def")
			require.NoError(t, err)

			assert.Equal(t, "before", loaded[0].Text())
		},
	)

	t.Run(
		"overwrites previous messages for same session ID",
		func(t *testing.T) {
			t.Parallel()

			s := agent.NewMemorySession()

			err := s.Save(
				context.Background(),
				"sess-ow",
				[]llm.Message{
					{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "first"}}},
				},
			)
			require.NoError(t, err)

			err = s.Save(
				context.Background(),
				"sess-ow",
				[]llm.Message{
					{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "second"}}},
					{Role: llm.RoleAssistant, Parts: []llm.Part{llm.TextPart{Text: "reply"}}},
				},
			)
			require.NoError(t, err)

			loaded, err := s.Load(context.Background(), "sess-ow")
			require.NoError(t, err)
			require.Len(t, loaded, 2)
			assert.Equal(t, "second", loaded[0].Text())
			assert.Equal(t, "reply", loaded[1].Text())
		},
	)

	t.Run(
		"preserves tool calls and tool call ID",
		func(t *testing.T) {
			t.Parallel()

			s := agent.NewMemorySession()
			messages := []llm.Message{
				{
					Role: llm.RoleAssistant,
					ToolCalls: []llm.ToolCall{
						{
							ID: "call-1",
							Function: llm.FunctionCall{
								Name:      "get_weather",
								Arguments: `{"city":"Paris"}`,
							},
						},
					},
				},
				{
					Role:       llm.RoleTool,
					ToolCallID: "call-1",
					Parts:      []llm.Part{llm.TextPart{Text: "sunny"}},
				},
			}

			err := s.Save(context.Background(), "sess-tc", messages)
			require.NoError(t, err)

			loaded, err := s.Load(context.Background(), "sess-tc")
			require.NoError(t, err)
			require.Len(t, loaded, 2)

			require.Len(t, loaded[0].ToolCalls, 1)
			assert.Equal(t, "call-1", loaded[0].ToolCalls[0].ID)
			assert.Equal(t, "get_weather", loaded[0].ToolCalls[0].Function.Name)
			assert.Equal(t, `{"city":"Paris"}`, loaded[0].ToolCalls[0].Function.Arguments)

			assert.Equal(t, "call-1", loaded[1].ToolCallID)
			assert.Equal(t, "sunny", loaded[1].Text())
		},
	)

	t.Run(
		"handles empty message slice",
		func(t *testing.T) {
			t.Parallel()

			s := agent.NewMemorySession()

			err := s.Save(context.Background(), "sess-empty", []llm.Message{})
			require.NoError(t, err)

			loaded, err := s.Load(context.Background(), "sess-empty")
			require.NoError(t, err)
			assert.Empty(t, loaded)
		},
	)
}
