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

func TestHandoffTo(t *testing.T) {
	t.Parallel()

	t.Run(
		"sets the target agent",
		func(t *testing.T) {
			t.Parallel()

			target := agent.New(
				"billing",
				newTestClient(&mockProvider{}),
			)

			h := agent.HandoffTo(target)

			assert.Equal(t, target, h.Agent)
		},
	)

	t.Run(
		"defaults have zero values",
		func(t *testing.T) {
			t.Parallel()

			target := agent.New(
				"support",
				newTestClient(&mockProvider{}),
			)

			h := agent.HandoffTo(target)

			assert.Empty(t, h.ToolName)
			assert.Empty(t, h.ToolDescription)
			assert.Nil(t, h.InputFilter)
			assert.Nil(t, h.OnHandoff)
		},
	)

	t.Run(
		"applies multiple options",
		func(t *testing.T) {
			t.Parallel()

			target := agent.New(
				"escalation",
				newTestClient(&mockProvider{}),
			)

			filter := func(data agent.HandoffInputData) []llm.Message {
				return data.NewItems
			}

			callback := func(_ context.Context) error {
				return nil
			}

			h := agent.HandoffTo(
				target,
				agent.WithHandoffToolName("escalate"),
				agent.WithHandoffToolDescription("Escalate to senior agent"),
				agent.WithHandoffInputFilter(filter),
				agent.WithOnHandoff(callback),
			)

			assert.Equal(t, target, h.Agent)
			assert.Equal(t, "escalate", h.ToolName)
			assert.Equal(t, "Escalate to senior agent", h.ToolDescription)
			assert.NotNil(t, h.InputFilter)
			assert.NotNil(t, h.OnHandoff)
		},
	)
}

func TestWithHandoffToolName(t *testing.T) {
	t.Parallel()

	target := agent.New("billing", newTestClient(&mockProvider{}))
	h := agent.HandoffTo(target, agent.WithHandoffToolName("ask_billing"))

	assert.Equal(t, "ask_billing", h.ToolName)
}

func TestWithHandoffToolDescription(t *testing.T) {
	t.Parallel()

	target := agent.New("billing", newTestClient(&mockProvider{}))
	h := agent.HandoffTo(
		target,
		agent.WithHandoffToolDescription("Route billing questions"),
	)

	assert.Equal(t, "Route billing questions", h.ToolDescription)
}

func TestWithHandoffInputFilter(t *testing.T) {
	t.Parallel()

	t.Run(
		"sets the filter function",
		func(t *testing.T) {
			t.Parallel()

			target := agent.New("specialist", newTestClient(&mockProvider{}))
			h := agent.HandoffTo(
				target,
				agent.WithHandoffInputFilter(func(data agent.HandoffInputData) []llm.Message {
					return data.NewItems
				}),
			)

			assert.NotNil(t, h.InputFilter)
		},
	)

	t.Run(
		"filter receives correct data and returns filtered messages",
		func(t *testing.T) {
			t.Parallel()

			history := []llm.Message{
				userMessage("old message"),
				assistantMessage("old reply"),
			}
			newItems := []llm.Message{
				userMessage("new question"),
			}

			target := agent.New("specialist", newTestClient(&mockProvider{}))
			h := agent.HandoffTo(
				target,
				agent.WithHandoffInputFilter(func(data agent.HandoffInputData) []llm.Message {
					var filtered []llm.Message
					for _, m := range data.NewItems {
						if m.Role == llm.RoleUser {
							filtered = append(filtered, m)
						}
					}
					return filtered
				}),
			)

			result := h.InputFilter(agent.HandoffInputData{
				InputHistory: history,
				NewItems:     newItems,
			})

			require.Len(t, result, 1)
			assert.Equal(t, llm.RoleUser, result[0].Role)
		},
	)

	t.Run(
		"filter can combine history and new items",
		func(t *testing.T) {
			t.Parallel()

			history := []llm.Message{
				userMessage("context"),
			}
			newItems := []llm.Message{
				userMessage("question"),
			}

			target := agent.New("specialist", newTestClient(&mockProvider{}))
			h := agent.HandoffTo(
				target,
				agent.WithHandoffInputFilter(func(data agent.HandoffInputData) []llm.Message {
					all := make([]llm.Message, 0, len(data.InputHistory)+len(data.NewItems))
					all = append(all, data.InputHistory...)
					all = append(all, data.NewItems...)
					return all
				}),
			)

			result := h.InputFilter(agent.HandoffInputData{
				InputHistory: history,
				NewItems:     newItems,
			})

			assert.Len(t, result, 2)
		},
	)

	t.Run(
		"filter can return empty slice",
		func(t *testing.T) {
			t.Parallel()

			target := agent.New("specialist", newTestClient(&mockProvider{}))
			h := agent.HandoffTo(
				target,
				agent.WithHandoffInputFilter(func(_ agent.HandoffInputData) []llm.Message {
					return nil
				}),
			)

			result := h.InputFilter(agent.HandoffInputData{
				InputHistory: []llm.Message{userMessage("hello")},
				NewItems:     []llm.Message{userMessage("world")},
			})

			assert.Empty(t, result)
		},
	)
}

func TestWithOnHandoff(t *testing.T) {
	t.Parallel()

	t.Run(
		"sets the callback",
		func(t *testing.T) {
			t.Parallel()

			target := agent.New("billing", newTestClient(&mockProvider{}))
			h := agent.HandoffTo(
				target,
				agent.WithOnHandoff(func(_ context.Context) error {
					return nil
				}),
			)

			assert.NotNil(t, h.OnHandoff)
		},
	)

	t.Run(
		"callback is invocable",
		func(t *testing.T) {
			t.Parallel()

			var called bool

			target := agent.New("billing", newTestClient(&mockProvider{}))
			h := agent.HandoffTo(
				target,
				agent.WithOnHandoff(func(_ context.Context) error {
					called = true
					return nil
				}),
			)

			err := h.OnHandoff(context.Background())
			require.NoError(t, err)
			assert.True(t, called)
		},
	)

	t.Run(
		"callback propagates errors",
		func(t *testing.T) {
			t.Parallel()

			target := agent.New("billing", newTestClient(&mockProvider{}))
			h := agent.HandoffTo(
				target,
				agent.WithOnHandoff(func(_ context.Context) error {
					return assert.AnError
				}),
			)

			err := h.OnHandoff(context.Background())
			assert.ErrorIs(t, err, assert.AnError)
		},
	)
}

func TestHandoffTo_OptionOrder(t *testing.T) {
	t.Parallel()

	target := agent.New("billing", newTestClient(&mockProvider{}))

	h := agent.HandoffTo(
		target,
		agent.WithHandoffToolName("first_name"),
		agent.WithHandoffToolName("second_name"),
	)

	assert.Equal(t, "second_name", h.ToolName, "last option wins")
}
