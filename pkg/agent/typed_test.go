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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/llm"
)

type typedMockProvider struct {
	responses []*llm.ChatCompletionResponse
	calls     int
}

func (m *typedMockProvider) ChatCompletion(_ context.Context, _ *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
	if m.calls >= len(m.responses) {
		return nil, errors.New("no more mock responses")
	}
	resp := m.responses[m.calls]
	m.calls++
	return resp, nil
}

func (m *typedMockProvider) ChatCompletionStream(_ context.Context, _ *llm.ChatCompletionRequest) (llm.ChatCompletionStream, error) {
	return nil, errors.New("not implemented")
}

func typedStopResponse(text string) *llm.ChatCompletionResponse {
	return &llm.ChatCompletionResponse{
		Model: "test-model",
		Message: llm.Message{
			Role:  llm.RoleAssistant,
			Parts: []llm.Part{llm.TextPart{Text: text}},
		},
		Usage:        llm.Usage{InputTokens: 10, OutputTokens: 5},
		FinishReason: llm.FinishReasonStop,
	}
}

func TestTypeName(t *testing.T) {
	t.Parallel()

	t.Run(
		"named struct",
		func(t *testing.T) {
			t.Parallel()

			type UserInfo struct {
				Name string `json:"name"`
			}

			assert.Equal(t, "userinfo", typeName[UserInfo]())
		},
	)

	t.Run(
		"pointer to named struct",
		func(t *testing.T) {
			t.Parallel()

			type Invoice struct {
				Amount int `json:"amount"`
			}

			assert.Equal(t, "invoice", typeName[*Invoice]())
		},
	)

	t.Run(
		"anonymous struct returns output",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "output", typeName[struct{ X int }]())
		},
	)

	t.Run(
		"basic string type",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "string", typeName[string]())
		},
	)

	t.Run(
		"basic int type",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "int", typeName[int]())
		},
	)

	t.Run(
		"custom type alias",
		func(t *testing.T) {
			t.Parallel()

			type Status string

			assert.Equal(t, "status", typeName[Status]())
		},
	)

	t.Run(
		"any interface returns output",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "output", typeName[any]())
		},
	)

	t.Run(
		"named interface returns lowercased name",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, "error", typeName[error]())
		},
	)
}

func TestRunTyped(t *testing.T) {
	t.Parallel()

	t.Run(
		"successful structured output",
		func(t *testing.T) {
			t.Parallel()

			type CompanyInfo struct {
				Name    string `json:"name"`
				Country string `json:"country"`
			}

			provider := &typedMockProvider{
				responses: []*llm.ChatCompletionResponse{
					typedStopResponse(`{"name":"Probo","country":"FR"}`),
				},
			}

			ag := New(
				"assistant",
				llm.NewClient(provider, "test"),
				WithModel("test-model"),
				WithInstructions("Return structured info."),
			)

			result, err := RunTyped[CompanyInfo](
				context.Background(),
				ag,
				[]llm.Message{{
					Role:  llm.RoleUser,
					Parts: []llm.Part{llm.TextPart{Text: "Tell me about Probo"}},
				}},
			)

			require.NoError(t, err)
			assert.Equal(t, "Probo", result.Output.Name)
			assert.Equal(t, "FR", result.Output.Country)
			assert.Equal(t, 1, result.Turns)
			assert.Equal(t, 10, result.Usage.InputTokens)
			assert.Equal(t, 5, result.Usage.OutputTokens)
			assert.Equal(t, "assistant", result.LastAgent.Name())
		},
	)

	t.Run(
		"nested struct output",
		func(t *testing.T) {
			t.Parallel()

			type Address struct {
				City    string `json:"city"`
				Country string `json:"country"`
			}

			type Person struct {
				Name    string  `json:"name"`
				Age     int     `json:"age"`
				Address Address `json:"address"`
			}

			provider := &typedMockProvider{
				responses: []*llm.ChatCompletionResponse{
					typedStopResponse(`{"name":"Alice","age":30,"address":{"city":"Paris","country":"FR"}}`),
				},
			}

			ag := New(
				"assistant",
				llm.NewClient(provider, "test"),
				WithModel("test-model"),
			)

			result, err := RunTyped[Person](
				context.Background(),
				ag,
				[]llm.Message{{
					Role:  llm.RoleUser,
					Parts: []llm.Part{llm.TextPart{Text: "Describe Alice"}},
				}},
			)

			require.NoError(t, err)
			assert.Equal(t, "Alice", result.Output.Name)
			assert.Equal(t, 30, result.Output.Age)
			assert.Equal(t, "Paris", result.Output.Address.City)
			assert.Equal(t, "FR", result.Output.Address.Country)
		},
	)

	t.Run(
		"invalid JSON response",
		func(t *testing.T) {
			t.Parallel()

			type Info struct {
				Name string `json:"name"`
			}

			provider := &typedMockProvider{
				responses: []*llm.ChatCompletionResponse{
					typedStopResponse(`not valid json`),
				},
			}

			ag := New(
				"assistant",
				llm.NewClient(provider, "test"),
				WithModel("test-model"),
			)

			_, err := RunTyped[Info](
				context.Background(),
				ag,
				[]llm.Message{{
					Role:  llm.RoleUser,
					Parts: []llm.Part{llm.TextPart{Text: "test"}},
				}},
			)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot parse typed output")
		},
	)

	t.Run(
		"underlying Run failure propagates",
		func(t *testing.T) {
			t.Parallel()

			type Info struct {
				Name string `json:"name"`
			}

			provider := &typedMockProvider{
				responses: []*llm.ChatCompletionResponse{},
			}

			ag := New(
				"assistant",
				llm.NewClient(provider, "test"),
				WithModel("test-model"),
			)

			_, err := RunTyped[Info](
				context.Background(),
				ag,
				[]llm.Message{{
					Role:  llm.RoleUser,
					Parts: []llm.Part{llm.TextPart{Text: "test"}},
				}},
			)

			require.Error(t, err)
		},
	)

	t.Run(
		"does not mutate original agent",
		func(t *testing.T) {
			t.Parallel()

			type Info struct {
				Name string `json:"name"`
			}

			provider := &typedMockProvider{
				responses: []*llm.ChatCompletionResponse{
					typedStopResponse(`{"name":"test"}`),
					typedStopResponse("plain text"),
				},
			}

			client := llm.NewClient(provider, "test")

			ag := New(
				"assistant",
				client,
				WithModel("test-model"),
			)

			_, err := RunTyped[Info](
				context.Background(),
				ag,
				[]llm.Message{{
					Role:  llm.RoleUser,
					Parts: []llm.Part{llm.TextPart{Text: "structured"}},
				}},
			)
			require.NoError(t, err)

			assert.Nil(t, ag.responseFormat)
		},
	)

	t.Run(
		"slice output",
		func(t *testing.T) {
			t.Parallel()

			type Items struct {
				Names []string `json:"names"`
			}

			provider := &typedMockProvider{
				responses: []*llm.ChatCompletionResponse{
					typedStopResponse(`{"names":["Alice","Bob","Charlie"]}`),
				},
			}

			ag := New(
				"assistant",
				llm.NewClient(provider, "test"),
				WithModel("test-model"),
			)

			result, err := RunTyped[Items](
				context.Background(),
				ag,
				[]llm.Message{{
					Role:  llm.RoleUser,
					Parts: []llm.Part{llm.TextPart{Text: "list names"}},
				}},
			)

			require.NoError(t, err)
			assert.Equal(t, []string{"Alice", "Bob", "Charlie"}, result.Output.Names)
		},
	)

	t.Run(
		"context cancellation",
		func(t *testing.T) {
			t.Parallel()

			type Info struct {
				Name string `json:"name"`
			}

			provider := &typedMockProvider{
				responses: []*llm.ChatCompletionResponse{
					typedStopResponse(`{"name":"test"}`),
				},
			}

			ag := New(
				"assistant",
				llm.NewClient(provider, "test"),
				WithModel("test-model"),
			)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			_, err := RunTyped[Info](
				ctx,
				ag,
				[]llm.Message{{
					Role:  llm.RoleUser,
					Parts: []llm.Part{llm.TextPart{Text: "test"}},
				}},
			)

			require.Error(t, err)
			assert.Equal(t, 0, provider.calls)
		},
	)

	t.Run(
		"with tool call",
		func(t *testing.T) {
			t.Parallel()

			type Summary struct {
				City    string `json:"city"`
				Weather string `json:"weather"`
			}

			type Params struct {
				City string `json:"city"`
			}

			weatherTool := FunctionTool[Params](
				"get_weather",
				"Get weather for a city",
				func(_ context.Context, p Params) (ToolResult, error) {
					return ToolResult{Content: "Sunny, 22°C in " + p.City}, nil
				},
			)

			// Three responses: (1) tool call, (2) free-text summary
			// that triggers promotion to the synthesis turn, (3) the
			// forced structured output produced on the synthesis turn
			// with ToolChoice=none + schema enforced.
			provider := &typedMockProvider{
				responses: []*llm.ChatCompletionResponse{
					{
						Model: "test-model",
						Message: llm.Message{
							Role: llm.RoleAssistant,
							ToolCalls: []llm.ToolCall{{
								ID: "tc_1",
								Function: llm.FunctionCall{
									Name:      "get_weather",
									Arguments: `{"city":"Paris"}`,
								},
							}},
						},
						Usage:        llm.Usage{InputTokens: 10, OutputTokens: 5},
						FinishReason: llm.FinishReasonToolCalls,
					},
					typedStopResponse("Got the weather, ready to respond."),
					typedStopResponse(`{"city":"Paris","weather":"Sunny, 22°C"}`),
				},
			}

			ag := New(
				"assistant",
				llm.NewClient(provider, "test"),
				WithModel("test-model"),
				WithTools(weatherTool),
			)

			result, err := RunTyped[Summary](
				context.Background(),
				ag,
				[]llm.Message{{
					Role:  llm.RoleUser,
					Parts: []llm.Part{llm.TextPart{Text: "What's the weather in Paris?"}},
				}},
			)

			require.NoError(t, err)
			assert.Equal(t, "Paris", result.Output.City)
			assert.Equal(t, "Sunny, 22°C", result.Output.Weather)
			assert.Equal(t, 3, result.Turns)
		},
	)
}
