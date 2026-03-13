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

package llm_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.probo.inc/probo/pkg/llm"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

type mockProvider struct {
	chatResp   *llm.ChatCompletionResponse
	chatErr    error
	streamResp llm.ChatCompletionStream
	streamErr  error
}

func (m *mockProvider) ChatCompletion(_ context.Context, _ *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
	return m.chatResp, m.chatErr
}

func (m *mockProvider) ChatCompletionStream(_ context.Context, _ *llm.ChatCompletionRequest) (llm.ChatCompletionStream, error) {
	return m.streamResp, m.streamErr
}

type mockStream struct {
	events   []llm.ChatCompletionStreamEvent
	idx      int
	current  llm.ChatCompletionStreamEvent
	err      error
	closeErr error
	closed   bool
}

func (s *mockStream) Next() bool {
	if s.idx >= len(s.events) {
		return false
	}
	s.current = s.events[s.idx]
	s.idx++
	return true
}

func (s *mockStream) Event() llm.ChatCompletionStreamEvent { return s.current }
func (s *mockStream) Err() error                           { return s.err }
func (s *mockStream) Close() error                         { s.closed = true; return s.closeErr }

func newTestClient(provider llm.Provider) (*llm.Client, *tracetest.SpanRecorder) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))

	client := llm.NewClient(provider, "test",
		llm.WithTracerProvider(tp),
	)
	return client, recorder
}

func spanAttrMap(recorder *tracetest.SpanRecorder) map[string]any {
	spans := recorder.Ended()
	if len(spans) == 0 {
		return nil
	}
	m := make(map[string]any)
	for _, a := range spans[0].Attributes() {
		m[string(a.Key)] = a.Value.AsInterface()
	}
	return m
}

//go:fix inline

// ---------------------------------------------------------------------------
// Message.Text
// ---------------------------------------------------------------------------

func TestMessageText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  llm.Message
		want string
	}{
		{
			name: "single text part",
			msg: llm.Message{
				Parts: []llm.Part{llm.TextPart{Text: "hello"}},
			},
			want: "hello",
		},
		{
			name: "multiple text parts concatenated",
			msg: llm.Message{
				Parts: []llm.Part{
					llm.TextPart{Text: "hello"},
					llm.TextPart{Text: " world"},
				},
			},
			want: "hello world",
		},
		{
			name: "image parts skipped",
			msg: llm.Message{
				Parts: []llm.Part{
					llm.TextPart{Text: "before"},
					llm.ImagePart{URL: "http://example.com/img.png"},
					llm.TextPart{Text: "after"},
				},
			},
			want: "beforeafter",
		},
		{
			name: "no parts returns empty string",
			msg:  llm.Message{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.msg.Text())
		})
	}
}

// ---------------------------------------------------------------------------
// Usage.Add
// ---------------------------------------------------------------------------

func TestUsageAdd(t *testing.T) {
	t.Parallel()

	a := llm.Usage{InputTokens: 10, OutputTokens: 5}
	b := llm.Usage{InputTokens: 20, OutputTokens: 15}
	c := a.Add(b)

	assert.Equal(t, 30, c.InputTokens)
	assert.Equal(t, 20, c.OutputTokens)
}

// ---------------------------------------------------------------------------
// Error types
// ---------------------------------------------------------------------------

func TestErrors(t *testing.T) {
	t.Parallel()

	inner := errors.New("upstream")

	t.Run("ErrRateLimit", func(t *testing.T) {
		t.Parallel()

		t.Run("with retry after", func(t *testing.T) {
			t.Parallel()
			e := &llm.ErrRateLimit{RetryAfter: 30 * time.Second, Err: inner}
			assert.Contains(t, e.Error(), "retry after 30s")
			assert.Contains(t, e.Error(), "upstream")
			assert.ErrorIs(t, e, inner)
		})

		t.Run("without retry after", func(t *testing.T) {
			t.Parallel()
			e := &llm.ErrRateLimit{Err: inner}
			assert.Contains(t, e.Error(), "rate limited")
			assert.NotContains(t, e.Error(), "retry after")
			assert.ErrorIs(t, e, inner)
		})

		t.Run("errors.As", func(t *testing.T) {
			t.Parallel()
			var target *llm.ErrRateLimit
			e := &llm.ErrRateLimit{RetryAfter: 5 * time.Second, Err: inner}
			require.ErrorAs(t, e, &target)
			assert.Equal(t, 5*time.Second, target.RetryAfter)
		})
	})

	t.Run("ErrContextLength", func(t *testing.T) {
		t.Parallel()

		t.Run("with max tokens", func(t *testing.T) {
			t.Parallel()
			e := &llm.ErrContextLength{MaxTokens: 4096, Err: inner}
			assert.Contains(t, e.Error(), "4096")
			assert.ErrorIs(t, e, inner)
		})

		t.Run("without max tokens", func(t *testing.T) {
			t.Parallel()
			e := &llm.ErrContextLength{Err: inner}
			assert.Contains(t, e.Error(), "context length exceeded")
			assert.NotContains(t, e.Error(), "max")
			assert.ErrorIs(t, e, inner)
		})
	})

	t.Run("ErrContentFilter", func(t *testing.T) {
		t.Parallel()
		e := &llm.ErrContentFilter{Err: inner}
		assert.Contains(t, e.Error(), "content filtered")
		assert.ErrorIs(t, e, inner)
	})

	t.Run("ErrAuthentication", func(t *testing.T) {
		t.Parallel()
		e := &llm.ErrAuthentication{Err: inner}
		assert.Contains(t, e.Error(), "authentication failed")
		assert.ErrorIs(t, e, inner)
	})
}

// ---------------------------------------------------------------------------
// Client — ChatCompletion
// ---------------------------------------------------------------------------

func TestChatCompletion(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		provider := &mockProvider{
			chatResp: &llm.ChatCompletionResponse{
				Model: "test-model",
				Message: llm.Message{
					Role:  llm.RoleAssistant,
					Parts: []llm.Part{llm.TextPart{Text: "Hello!"}},
				},
				Usage:        llm.Usage{InputTokens: 10, OutputTokens: 5},
				FinishReason: llm.FinishReasonStop,
			},
		}

		client, recorder := newTestClient(provider)
		temp := 0.7
		resp, err := client.ChatCompletion(context.Background(), &llm.ChatCompletionRequest{
			Model: "test-model",
			Messages: []llm.Message{
				{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "Hi"}}},
			},
			Temperature: &temp,
		})

		require.NoError(t, err)
		assert.Equal(t, "test-model", resp.Model)
		assert.Equal(t, "Hello!", resp.Message.Text())
		assert.Equal(t, 10, resp.Usage.InputTokens)
		assert.Equal(t, 5, resp.Usage.OutputTokens)
		assert.Equal(t, llm.FinishReasonStop, resp.FinishReason)

		spans := recorder.Ended()
		require.Len(t, spans, 1)
		assert.Equal(t, "chat test-model", spans[0].Name())

		attrs := spanAttrMap(recorder)
		assert.Equal(t, "test", attrs["gen_ai.provider.name"])
		assert.Equal(t, "test-model", attrs["gen_ai.request.model"])
		assert.Equal(t, 0.7, attrs["gen_ai.request.temperature"])
		assert.Equal(t, "test-model", attrs["gen_ai.response.model"])
		assert.Equal(t, int64(10), attrs["gen_ai.usage.input_tokens"])
		assert.Equal(t, int64(5), attrs["gen_ai.usage.output_tokens"])
	})

	t.Run("all span attributes", func(t *testing.T) {
		t.Parallel()

		provider := &mockProvider{
			chatResp: &llm.ChatCompletionResponse{
				Model:        "gpt-4",
				FinishReason: llm.FinishReasonStop,
				Usage:        llm.Usage{InputTokens: 50, OutputTokens: 25},
				Message: llm.Message{
					Role:  llm.RoleAssistant,
					Parts: []llm.Part{llm.TextPart{Text: "ok"}},
				},
			},
		}

		client, recorder := newTestClient(provider)
		maxTokens := 1024
		topP := 0.9
		resp, err := client.ChatCompletion(context.Background(), &llm.ChatCompletionRequest{
			Model:         "gpt-4",
			Messages:      []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "test"}}}},
			MaxTokens:     &maxTokens,
			TopP:          &topP,
			StopSequences: []string{"END", "STOP"},
		})

		require.NoError(t, err)
		require.NotNil(t, resp)

		attrs := spanAttrMap(recorder)
		assert.Equal(t, int64(1024), attrs["gen_ai.request.max_tokens"])
		assert.Equal(t, 0.9, attrs["gen_ai.request.top_p"])
		assert.Equal(t, []string{"END", "STOP"}, attrs["gen_ai.request.stop_sequences"])
	})

	t.Run("error sets span status", func(t *testing.T) {
		t.Parallel()

		provider := &mockProvider{
			chatErr: errors.New("provider error"),
		}

		client, recorder := newTestClient(provider)
		_, err := client.ChatCompletion(context.Background(), &llm.ChatCompletionRequest{
			Model:    "test-model",
			Messages: []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "Hi"}}}},
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "provider error")

		spans := recorder.Ended()
		require.Len(t, spans, 1)
		assert.Equal(t, codes.Error, spans[0].Status().Code)
	})
}

// ---------------------------------------------------------------------------
// Client — ChatCompletionStream
// ---------------------------------------------------------------------------

func TestChatCompletionStream(t *testing.T) {
	t.Parallel()

	t.Run("success collects deltas", func(t *testing.T) {
		t.Parallel()

		events := []llm.ChatCompletionStreamEvent{
			{Delta: llm.MessageDelta{Content: "Hello"}},
			{Delta: llm.MessageDelta{Content: " world"}},
			{
				FinishReason: new(llm.FinishReasonStop),
				Usage:        &llm.Usage{InputTokens: 8, OutputTokens: 4},
			},
		}

		client, recorder := newTestClient(&mockProvider{
			streamResp: &mockStream{events: events},
		})
		stream, err := client.ChatCompletionStream(context.Background(), &llm.ChatCompletionRequest{
			Model:    "test-model",
			Messages: []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "Hi"}}}},
		})
		require.NoError(t, err)

		var collected []string
		for stream.Next() {
			e := stream.Event()
			if e.Delta.Content != "" {
				collected = append(collected, e.Delta.Content)
			}
		}
		require.NoError(t, stream.Err())
		require.NoError(t, stream.Close())

		assert.Equal(t, []string{"Hello", " world"}, collected)

		spans := recorder.Ended()
		require.Len(t, spans, 1)
		assert.Equal(t, "chat test-model", spans[0].Name())
	})

	t.Run("provider error sets span status", func(t *testing.T) {
		t.Parallel()

		client, recorder := newTestClient(&mockProvider{
			streamErr: errors.New("stream open failed"),
		})
		_, err := client.ChatCompletionStream(context.Background(), &llm.ChatCompletionRequest{
			Model:    "test-model",
			Messages: []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "Hi"}}}},
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "stream open failed")

		spans := recorder.Ended()
		require.Len(t, spans, 1)
		assert.Equal(t, codes.Error, spans[0].Status().Code)
	})

	t.Run("inner stream error records error on span", func(t *testing.T) {
		t.Parallel()

		ms := &mockStream{
			events: []llm.ChatCompletionStreamEvent{
				{Delta: llm.MessageDelta{Content: "partial"}},
			},
			err: errors.New("connection reset"),
		}

		client, recorder := newTestClient(&mockProvider{streamResp: ms})
		stream, err := client.ChatCompletionStream(context.Background(), &llm.ChatCompletionRequest{
			Model:    "test-model",
			Messages: []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "Hi"}}}},
		})
		require.NoError(t, err)

		for stream.Next() {
		}
		assert.ErrorContains(t, stream.Err(), "connection reset")
		_ = stream.Close()

		spans := recorder.Ended()
		require.Len(t, spans, 1)
		assert.Equal(t, codes.Error, spans[0].Status().Code)
	})

	t.Run("span finalized on close before exhausting", func(t *testing.T) {
		t.Parallel()

		events := []llm.ChatCompletionStreamEvent{
			{Delta: llm.MessageDelta{Content: "a"}},
			{Delta: llm.MessageDelta{Content: "b"}},
			{Delta: llm.MessageDelta{Content: "c"}},
		}

		client, recorder := newTestClient(&mockProvider{
			streamResp: &mockStream{events: events},
		})
		stream, err := client.ChatCompletionStream(context.Background(), &llm.ChatCompletionRequest{
			Model:    "test-model",
			Messages: []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "Hi"}}}},
		})
		require.NoError(t, err)

		// Read only one event, then close early.
		require.True(t, stream.Next())
		require.NoError(t, stream.Close())

		spans := recorder.Ended()
		require.Len(t, spans, 1, "span should be ended by Close even without exhausting stream")
	})

	t.Run("close error traced on span", func(t *testing.T) {
		t.Parallel()

		ms := &mockStream{
			events: []llm.ChatCompletionStreamEvent{
				{Delta: llm.MessageDelta{Content: "partial"}},
			},
			closeErr: errors.New("broken pipe"),
		}

		client, recorder := newTestClient(&mockProvider{streamResp: ms})
		stream, err := client.ChatCompletionStream(context.Background(), &llm.ChatCompletionRequest{
			Model:    "test-model",
			Messages: []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "Hi"}}}},
		})
		require.NoError(t, err)

		require.True(t, stream.Next())
		closeErr := stream.Close()
		require.ErrorContains(t, closeErr, "broken pipe")

		spans := recorder.Ended()
		require.Len(t, spans, 1)
		assert.Equal(t, codes.Error, spans[0].Status().Code)
		assert.Contains(t, spans[0].Status().Description, "broken pipe")
	})

	t.Run("stream span records usage and finish reason", func(t *testing.T) {
		t.Parallel()

		events := []llm.ChatCompletionStreamEvent{
			{Delta: llm.MessageDelta{Content: "done"}},
			{
				FinishReason: new(llm.FinishReasonLength),
				Usage:        &llm.Usage{InputTokens: 100, OutputTokens: 50},
			},
		}

		client, recorder := newTestClient(&mockProvider{
			streamResp: &mockStream{events: events},
		})
		stream, err := client.ChatCompletionStream(context.Background(), &llm.ChatCompletionRequest{
			Model:    "test-model",
			Messages: []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "Hi"}}}},
		})
		require.NoError(t, err)

		for stream.Next() {
		}
		require.NoError(t, stream.Err())
		require.NoError(t, stream.Close())

		spans := recorder.Ended()
		require.Len(t, spans, 1)

		attrs := make(map[string]any)
		for _, a := range spans[0].Attributes() {
			attrs[string(a.Key)] = a.Value.AsInterface()
		}
		assert.Equal(t, int64(100), attrs["gen_ai.usage.input_tokens"])
		assert.Equal(t, int64(50), attrs["gen_ai.usage.output_tokens"])
		assert.Equal(t, []string{"length"}, attrs["gen_ai.response.finish_reasons"])
	})
}

// ---------------------------------------------------------------------------
// StreamAccumulator
// ---------------------------------------------------------------------------

func TestStreamAccumulator(t *testing.T) {
	t.Parallel()

	t.Run("text and single tool call", func(t *testing.T) {
		t.Parallel()

		events := []llm.ChatCompletionStreamEvent{
			{Model: "gpt-4o", Delta: llm.MessageDelta{Content: "Hello"}},
			{Delta: llm.MessageDelta{Content: " world"}},
			{Delta: llm.MessageDelta{
				ToolCalls: []llm.ToolCallDelta{
					{Index: 0, ID: "tc_1", Name: "get_weather"},
				},
			}},
			{Delta: llm.MessageDelta{
				ToolCalls: []llm.ToolCallDelta{
					{Index: 0, Arguments: `{"city":`},
				},
			}},
			{Delta: llm.MessageDelta{
				ToolCalls: []llm.ToolCallDelta{
					{Index: 0, Arguments: `"Paris"}`},
				},
			}},
			{
				FinishReason: new(llm.FinishReasonToolCalls),
				Usage:        &llm.Usage{InputTokens: 20, OutputTokens: 15},
			},
		}

		acc := llm.NewStreamAccumulator(&mockStream{events: events})
		for acc.Next() {
		}
		require.NoError(t, acc.Err())

		resp := acc.Response()
		assert.Equal(t, "gpt-4o", resp.Model)
		assert.Equal(t, "Hello world", resp.Message.Text())
		assert.Equal(t, llm.RoleAssistant, resp.Message.Role)
		assert.Equal(t, llm.FinishReasonToolCalls, resp.FinishReason)
		assert.Equal(t, 20, resp.Usage.InputTokens)
		assert.Equal(t, 15, resp.Usage.OutputTokens)

		require.Len(t, resp.Message.ToolCalls, 1)
		tc := resp.Message.ToolCalls[0]
		assert.Equal(t, "tc_1", tc.ID)
		assert.Equal(t, "get_weather", tc.Function.Name)
		assert.Equal(t, `{"city":"Paris"}`, tc.Function.Arguments)
	})

	t.Run("multiple tool calls at different indices", func(t *testing.T) {
		t.Parallel()

		events := []llm.ChatCompletionStreamEvent{
			{Delta: llm.MessageDelta{
				ToolCalls: []llm.ToolCallDelta{
					{Index: 0, ID: "tc_a", Name: "search"},
				},
			}},
			{Delta: llm.MessageDelta{
				ToolCalls: []llm.ToolCallDelta{
					{Index: 0, Arguments: `{"q":"go"}`},
				},
			}},
			{Delta: llm.MessageDelta{
				ToolCalls: []llm.ToolCallDelta{
					{Index: 1, ID: "tc_b", Name: "fetch"},
				},
			}},
			{Delta: llm.MessageDelta{
				ToolCalls: []llm.ToolCallDelta{
					{Index: 1, Arguments: `{"url":"https://example.com"}`},
				},
			}},
			{
				FinishReason: new(llm.FinishReasonToolCalls),
				Usage:        &llm.Usage{InputTokens: 30, OutputTokens: 10},
			},
		}

		acc := llm.NewStreamAccumulator(&mockStream{events: events})
		for acc.Next() {
		}
		require.NoError(t, acc.Err())

		resp := acc.Response()
		require.Len(t, resp.Message.ToolCalls, 2)

		assert.Equal(t, "tc_a", resp.Message.ToolCalls[0].ID)
		assert.Equal(t, "search", resp.Message.ToolCalls[0].Function.Name)
		assert.Equal(t, `{"q":"go"}`, resp.Message.ToolCalls[0].Function.Arguments)

		assert.Equal(t, "tc_b", resp.Message.ToolCalls[1].ID)
		assert.Equal(t, "fetch", resp.Message.ToolCalls[1].Function.Name)
		assert.Equal(t, `{"url":"https://example.com"}`, resp.Message.ToolCalls[1].Function.Arguments)
	})

	t.Run("text only without tool calls", func(t *testing.T) {
		t.Parallel()

		events := []llm.ChatCompletionStreamEvent{
			{Delta: llm.MessageDelta{Content: "Just text."}},
			{
				FinishReason: new(llm.FinishReasonStop),
				Usage:        &llm.Usage{InputTokens: 5, OutputTokens: 3},
			},
		}

		acc := llm.NewStreamAccumulator(&mockStream{events: events})
		for acc.Next() {
		}
		require.NoError(t, acc.Err())

		resp := acc.Response()
		assert.Equal(t, "Just text.", resp.Message.Text())
		assert.Equal(t, llm.FinishReasonStop, resp.FinishReason)
		assert.Empty(t, resp.Message.ToolCalls)
	})

	t.Run("proxies events transparently", func(t *testing.T) {
		t.Parallel()

		events := []llm.ChatCompletionStreamEvent{
			{Delta: llm.MessageDelta{Content: "a"}},
			{Delta: llm.MessageDelta{Content: "b"}},
			{FinishReason: new(llm.FinishReasonStop)},
		}

		acc := llm.NewStreamAccumulator(&mockStream{events: events})
		var seen []string
		for acc.Next() {
			e := acc.Event()
			if e.Delta.Content != "" {
				seen = append(seen, e.Delta.Content)
			}
		}

		assert.Equal(t, []string{"a", "b"}, seen)
	})
}
