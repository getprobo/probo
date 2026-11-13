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

package openaicompat_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/llm"
	"go.probo.inc/probo/pkg/llm/openaicompat"
)

func TestChatCompletion_Text(t *testing.T) {
	t.Parallel()

	var captured map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &captured))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"model": "local-model",
			"choices": [{"message": {"role": "assistant", "content": "hello"}, "finish_reason": "stop"}],
			"usage": {"prompt_tokens": 11, "completion_tokens": 3}
		}`))
	}))
	defer srv.Close()

	p := openaicompat.NewProvider(srv.URL+"/v1", "")

	resp, err := p.ChatCompletion(context.Background(), &llm.ChatCompletionRequest{
		Model: "local-model",
		Messages: []llm.Message{
			{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hi"}}},
		},
	})
	require.NoError(t, err)

	require.Len(t, resp.Message.Parts, 1)
	assert.Equal(t, "hello", resp.Message.Parts[0].(llm.TextPart).Text)
	assert.Equal(t, llm.FinishReasonStop, resp.FinishReason)
	assert.Equal(t, 11, resp.Usage.InputTokens)
	assert.Equal(t, 3, resp.Usage.OutputTokens)

	// A server implementing only the required fields must not be sent
	// optional ones it may reject.
	assert.NotContains(t, captured, "temperature")
	assert.NotContains(t, captured, "tools")
	assert.NotContains(t, captured, "response_format")
}

func TestChatCompletion_NoAPIKeyOmitsAuthorizationHeader(t *testing.T) {
	t.Parallel()

	var hadAuth bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, hadAuth = r.Header["Authorization"]

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"},"finish_reason":"stop"}]}`))
	}))
	defer srv.Close()

	p := openaicompat.NewProvider(srv.URL+"/v1", "")

	_, err := p.ChatCompletion(context.Background(), &llm.ChatCompletionRequest{
		Model:    "local-model",
		Messages: []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hi"}}}},
	})
	require.NoError(t, err)

	// Some local servers reject an empty bearer token outright, so the header
	// must be absent rather than empty.
	assert.False(t, hadAuth)
}

func TestChatCompletion_ToolCalls(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [{
				"message": {
					"role": "assistant",
					"tool_calls": [{
						"index": 0, "id": "call_1", "type": "function",
						"function": {"name": "get_weather", "arguments": "{\"city\":\"Paris\"}"}
					}]
				},
				"finish_reason": "tool_calls"
			}]
		}`))
	}))
	defer srv.Close()

	p := openaicompat.NewProvider(srv.URL+"/v1", "key")

	resp, err := p.ChatCompletion(context.Background(), &llm.ChatCompletionRequest{
		Model:    "local-model",
		Messages: []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "weather?"}}}},
		Tools: []llm.Tool{{
			Name:       "get_weather",
			Parameters: json.RawMessage(`{"type":"object"}`),
		}},
	})
	require.NoError(t, err)

	require.Len(t, resp.Message.ToolCalls, 1)
	assert.Equal(t, "call_1", resp.Message.ToolCalls[0].ID)
	assert.Equal(t, "get_weather", resp.Message.ToolCalls[0].Function.Name)
	assert.Equal(t, llm.FinishReasonToolCalls, resp.FinishReason)
}

func TestChatCompletion_RateLimit(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "7")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"message":"slow down"}}`))
	}))
	defer srv.Close()

	p := openaicompat.NewProvider(srv.URL+"/v1", "key")

	_, err := p.ChatCompletion(context.Background(), &llm.ChatCompletionRequest{
		Model:    "local-model",
		Messages: []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hi"}}}},
	})

	var rateLimit *llm.ErrRateLimit
	require.ErrorAs(t, err, &rateLimit)
	assert.Equal(t, float64(7), rateLimit.RetryAfter.Seconds())
}

func TestChatCompletion_ContextLengthFromMessage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"This model's maximum context length is 8192 tokens"}}`))
	}))
	defer srv.Close()

	p := openaicompat.NewProvider(srv.URL+"/v1", "key")

	_, err := p.ChatCompletion(context.Background(), &llm.ChatCompletionRequest{
		Model:    "local-model",
		Messages: []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hi"}}}},
	})

	var contextLength *llm.ErrContextLength
	require.ErrorAs(t, err, &contextLength)
}

func TestChatCompletionStream_TextAndUsage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		for _, chunk := range []string{
			`{"model":"local-model","choices":[{"delta":{"content":"he"}}]}`,
			`{"choices":[{"delta":{"content":"llo"},"finish_reason":"stop"}]}`,
			`{"usage":{"prompt_tokens":5,"completion_tokens":2}}`,
			`[DONE]`,
		} {
			_, _ = io.WriteString(w, "data: "+chunk+"\n\n")
		}
	}))
	defer srv.Close()

	p := openaicompat.NewProvider(srv.URL+"/v1", "key")

	stream, err := p.ChatCompletionStream(context.Background(), &llm.ChatCompletionRequest{
		Model:    "local-model",
		Messages: []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hi"}}}},
	})
	require.NoError(t, err)
	defer stream.Close()

	var (
		text   string
		usage  *llm.Usage
		finish *llm.FinishReason
	)

	for stream.Next() {
		event := stream.Event()
		text += event.Delta.Content

		if event.Usage != nil {
			usage = event.Usage
		}

		if event.FinishReason != nil {
			finish = event.FinishReason
		}
	}
	require.NoError(t, stream.Err())

	assert.Equal(t, "hello", text)
	require.NotNil(t, finish)
	assert.Equal(t, llm.FinishReasonStop, *finish)
	require.NotNil(t, usage)
	assert.Equal(t, 5, usage.InputTokens)
	assert.Equal(t, 2, usage.OutputTokens)
}
