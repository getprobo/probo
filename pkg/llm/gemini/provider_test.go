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

package gemini_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/llm"
	"go.probo.inc/probo/pkg/llm/gemini"
	"golang.org/x/oauth2"
)

func TestChatCompletion_TextAndSystemInstruction(t *testing.T) {
	t.Parallel()

	var captured map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.True(t, strings.HasSuffix(r.URL.Path, "/models/gemini-3.7-flash:generateContent"))
		assert.Equal(t, "test-key", r.Header.Get("x-goog-api-key"))
		// The key must not be in the URL, where it would land in access logs.
		assert.NotContains(t, r.URL.RawQuery, "test-key")

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &captured))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"modelVersion": "gemini-3.7-flash",
			"candidates": [{"content": {"role":"model","parts":[{"text":"bonjour"}]}, "finishReason":"STOP"}],
			"usageMetadata": {"promptTokenCount": 9, "candidatesTokenCount": 4}
		}`))
	}))
	defer srv.Close()

	p := gemini.NewProvider("test-key", gemini.WithBaseURL(srv.URL))

	resp, err := p.ChatCompletion(context.Background(), &llm.ChatCompletionRequest{
		Model: "gemini-3.7-flash",
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Parts: []llm.Part{llm.TextPart{Text: "be terse"}}},
			{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hello"}}},
		},
	})
	require.NoError(t, err)

	require.Len(t, resp.Message.Parts, 1)
	assert.Equal(t, "bonjour", resp.Message.Parts[0].(llm.TextPart).Text)
	assert.Equal(t, llm.FinishReasonStop, resp.FinishReason)
	assert.Equal(t, 9, resp.Usage.InputTokens)
	assert.Equal(t, 4, resp.Usage.OutputTokens)

	// The system turn is hoisted out of contents, not left inline as a role.
	require.Contains(t, captured, "systemInstruction")
	contents, ok := captured["contents"].([]any)
	require.True(t, ok)
	assert.Len(t, contents, 1)
}

func TestChatCompletion_AssistantRoleIsModel(t *testing.T) {
	t.Parallel()

	var captured map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(body, &captured))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"ok"}]},"finishReason":"STOP"}]}`))
	}))
	defer srv.Close()

	p := gemini.NewProvider("k", gemini.WithBaseURL(srv.URL))

	_, err := p.ChatCompletion(context.Background(), &llm.ChatCompletionRequest{
		Model: "gemini-3.7-flash",
		Messages: []llm.Message{
			{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hi"}}},
			{Role: llm.RoleAssistant, Parts: []llm.Part{llm.TextPart{Text: "hey"}}},
			{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "again"}}},
		},
	})
	require.NoError(t, err)

	contents, ok := captured["contents"].([]any)
	require.True(t, ok)
	require.Len(t, contents, 3)

	second, ok := contents[1].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "model", second["role"])
}

func TestChatCompletion_FunctionCall(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"candidates": [{
				"content": {"role":"model","parts":[
					{"functionCall": {"name":"get_weather","args":{"city":"Paris"}}}
				]},
				"finishReason": "STOP"
			}]
		}`))
	}))
	defer srv.Close()

	p := gemini.NewProvider("k", gemini.WithBaseURL(srv.URL))

	resp, err := p.ChatCompletion(context.Background(), &llm.ChatCompletionRequest{
		Model:    "gemini-3.7-flash",
		Messages: []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "weather?"}}}},
		Tools: []llm.Tool{{
			Name:       "get_weather",
			Parameters: json.RawMessage(`{"type":"object"}`),
		}},
	})
	require.NoError(t, err)

	require.Len(t, resp.Message.ToolCalls, 1)
	assert.Equal(t, "get_weather", resp.Message.ToolCalls[0].Function.Name)
	assert.JSONEq(t, `{"city":"Paris"}`, resp.Message.ToolCalls[0].Function.Arguments)

	// Gemini reports STOP even when the turn ends in a call, so the finish
	// reason has to be derived from the presence of one.
	assert.Equal(t, llm.FinishReasonToolCalls, resp.FinishReason)
}

func TestChatCompletion_ToolResultBecomesFunctionResponse(t *testing.T) {
	t.Parallel()

	var captured map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		require.NoError(t, json.Unmarshal(body, &captured))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"ok"}]},"finishReason":"STOP"}]}`))
	}))
	defer srv.Close()

	p := gemini.NewProvider("k", gemini.WithBaseURL(srv.URL))

	_, err := p.ChatCompletion(context.Background(), &llm.ChatCompletionRequest{
		Model: "gemini-3.7-flash",
		Messages: []llm.Message{
			{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "weather?"}}},
			{
				Role:       llm.RoleTool,
				ToolCallID: "get_weather",
				Parts:      []llm.Part{llm.TextPart{Text: "sunny"}},
			},
		},
	})
	require.NoError(t, err)

	contents, ok := captured["contents"].([]any)
	require.True(t, ok)
	require.Len(t, contents, 2)

	toolTurn, ok := contents[1].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "user", toolTurn["role"])

	parts, ok := toolTurn["parts"].([]any)
	require.True(t, ok)
	require.Len(t, parts, 1)

	first, ok := parts[0].(map[string]any)
	require.True(t, ok)

	fnResp, ok := first["functionResponse"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "get_weather", fnResp["name"])

	// A bare string is not a valid functionResponse payload, so it is wrapped.
	response, ok := fnResp["response"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "sunny", response["output"])
}

func TestChatCompletion_NoCandidatesIsContentFilter(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"candidates":[]}`))
	}))
	defer srv.Close()

	p := gemini.NewProvider("k", gemini.WithBaseURL(srv.URL))

	_, err := p.ChatCompletion(context.Background(), &llm.ChatCompletionRequest{
		Model:    "gemini-3.7-flash",
		Messages: []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hi"}}}},
	})

	var filtered *llm.ErrContentFilter
	require.ErrorAs(t, err, &filtered)
}

func TestChatCompletion_Authentication(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":{"code":403,"message":"API key not valid"}}`))
	}))
	defer srv.Close()

	p := gemini.NewProvider("bad", gemini.WithBaseURL(srv.URL))

	_, err := p.ChatCompletion(context.Background(), &llm.ChatCompletionRequest{
		Model:    "gemini-3.7-flash",
		Messages: []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hi"}}}},
	})

	var auth *llm.ErrAuthentication
	require.ErrorAs(t, err, &auth)
}

func TestChatCompletionStream_UsesSSEAndAccumulates(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Without alt=sse the endpoint returns one JSON array at the end,
		// which is not a stream.
		assert.Equal(t, "sse", r.URL.Query().Get("alt"))
		assert.True(t, strings.HasSuffix(r.URL.Path, ":streamGenerateContent"))

		w.Header().Set("Content-Type", "text/event-stream")

		for _, chunk := range []string{
			`{"candidates":[{"content":{"parts":[{"text":"bon"}]}}]}`,
			`{"candidates":[{"content":{"parts":[{"text":"jour"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":6,"candidatesTokenCount":2}}`,
		} {
			_, _ = io.WriteString(w, "data: "+chunk+"\n\n")
		}
	}))
	defer srv.Close()

	p := gemini.NewProvider("k", gemini.WithBaseURL(srv.URL))

	stream, err := p.ChatCompletionStream(context.Background(), &llm.ChatCompletionRequest{
		Model:    "gemini-3.7-flash",
		Messages: []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hello"}}}},
	})
	require.NoError(t, err)
	defer stream.Close()

	var (
		text  string
		usage *llm.Usage
	)

	for stream.Next() {
		event := stream.Event()
		text += event.Delta.Content

		if event.Usage != nil {
			usage = event.Usage
		}
	}
	require.NoError(t, stream.Err())

	assert.Equal(t, "bonjour", text)
	require.NotNil(t, usage)
	assert.Equal(t, 6, usage.InputTokens)
}

func TestChatCompletion_VertexModeUsesProjectPathAndBearerAuth(t *testing.T) {
	t.Parallel()

	var (
		gotPath string
		gotAuth string
		gotKey  string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotKey = r.Header.Get("x-goog-api-key")

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"ok"}]},"finishReason":"STOP"}]}`))
	}))
	defer srv.Close()

	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: "ya29.test",
		TokenType:   "Bearer",
	})

	p := gemini.NewProvider(
		"",
		gemini.WithBaseURL(srv.URL),
		gemini.WithVertex("benedoc-dev", "us-east4", ts),
	)

	_, err := p.ChatCompletion(context.Background(), &llm.ChatCompletionRequest{
		Model:    "gemini-3.7-flash",
		Messages: []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hi"}}}},
	})
	require.NoError(t, err)

	// Vertex nests the model under a project and location rather than
	// serving /models/{model} like the Generative Language API.
	assert.Equal(
		t,
		"/v1/projects/benedoc-dev/locations/us-east4/publishers/google/models/gemini-3.7-flash:generateContent",
		gotPath,
	)

	// ADC, not an API key: the workload's own identity is the credential.
	assert.Equal(t, "Bearer ya29.test", gotAuth)
	assert.Empty(t, gotKey)
}

func TestChatCompletion_RejectsUnsupportedOptions(t *testing.T) {
	t.Parallel()

	p := gemini.NewProvider("k", gemini.WithBaseURL("https://example.invalid"))

	_, err := p.ChatCompletion(context.Background(), &llm.ChatCompletionRequest{
		Model:            "gemini-3.7-flash",
		Messages:         []llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hi"}}}},
		FrequencyPenalty: new(0.5),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "frequency penalty")
}
