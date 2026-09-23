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

package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/llm"
)

func TestProvider_UsesResponsesEndpoint(t *testing.T) {
	t.Parallel()

	var requestBody map[string]any

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/responses", r.URL.Path)

				err := json.NewDecoder(r.Body).Decode(&requestBody)
				require.NoError(t, err)

				w.Header().Set("Content-Type", "application/json")
				_, err = fmt.Fprint(
					w,
					`{
						"id": "resp_1",
						"object": "response",
						"created_at": 1,
						"status": "completed",
						"model": "gpt-5.6",
						"output": [
							{
								"id": "msg_1",
								"type": "message",
								"status": "completed",
								"role": "assistant",
								"content": [
									{"type": "output_text", "text": "done", "annotations": []}
								]
							},
							{
								"id": "fc_1",
								"type": "function_call",
								"status": "completed",
								"call_id": "call_2",
								"name": "lookup",
								"arguments": "{\"id\":\"2\"}"
							}
						],
						"usage": {
							"input_tokens": 12,
							"output_tokens": 4,
							"total_tokens": 16,
							"input_tokens_details": {"cached_tokens": 0},
							"output_tokens_details": {"reasoning_tokens": 0}
						}
					}`,
				)
				require.NoError(t, err)
			},
		),
	)
	t.Cleanup(server.Close)

	maxTokens := 2048
	temperature := 0.4
	topP := 0.9
	parallelToolCalls := true
	provider := NewProvider(
		"test-key",
		WithBaseURL(server.URL+"/v1"),
		WithMaxRetries(0),
	)

	response, err := provider.ChatCompletion(
		context.Background(),
		&llm.ChatCompletionRequest{
			Model: "gpt-5.6",
			Messages: []llm.Message{
				{
					Role:  llm.RoleSystem,
					Parts: []llm.Part{llm.TextPart{Text: "Follow instructions."}},
				},
				{
					Role: llm.RoleUser,
					Parts: []llm.Part{
						llm.TextPart{Text: "Review this."},
						llm.ImagePart{URL: "https://example.com/image.png"},
						llm.FilePart{
							Data:     "cGRm",
							MimeType: "application/pdf",
							Filename: "report.pdf",
						},
					},
				},
				{
					Role: llm.RoleAssistant,
					ToolCalls: []llm.ToolCall{
						{
							ID: "call_1",
							Function: llm.FunctionCall{
								Name:      "lookup",
								Arguments: `{"id":"1"}`,
							},
						},
					},
				},
				{
					Role:       llm.RoleTool,
					Parts:      []llm.Part{llm.TextPart{Text: `{"name":"Probo"}`}},
					ToolCallID: "call_1",
				},
			},
			MaxTokens:         &maxTokens,
			Temperature:       &temperature,
			TopP:              &topP,
			ParallelToolCalls: &parallelToolCalls,
			Tools: []llm.Tool{
				{
					Name:        "lookup",
					Description: "Look up a record",
					Parameters:  json.RawMessage(`{"type":"object"}`),
				},
			},
			ToolChoice: &llm.ToolChoice{
				Type:     llm.ToolChoiceFunction,
				Function: "lookup",
			},
			ResponseFormat: &llm.ResponseFormat{
				Type: llm.ResponseFormatJSONSchema,
				JSONSchema: &llm.JSONSchema{
					Name:   "result",
					Schema: json.RawMessage(`{"type":"object"}`),
					Strict: true,
				},
			},
			Thinking: &llm.ThinkingConfig{
				Enabled:      true,
				BudgetTokens: 4096,
			},
		},
	)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, "gpt-5.6", response.Model)
	assert.Equal(t, "done", response.Message.Text())
	assert.Equal(t, llm.FinishReasonToolCalls, response.FinishReason)
	assert.Equal(t, llm.Usage{InputTokens: 12, OutputTokens: 4}, response.Usage)
	assert.Equal(
		t,
		[]llm.ToolCall{
			{
				ID: "call_2",
				Function: llm.FunctionCall{
					Name:      "lookup",
					Arguments: `{"id":"2"}`,
				},
			},
		},
		response.Message.ToolCalls,
	)

	assert.Equal(t, "gpt-5.6", requestBody["model"])
	assert.Equal(t, false, requestBody["store"])
	assert.Equal(t, float64(maxTokens), requestBody["max_output_tokens"])
	assert.Equal(t, "medium", requestBody["reasoning"].(map[string]any)["effort"])
	assert.Equal(t, "json_schema", requestBody["text"].(map[string]any)["format"].(map[string]any)["type"])

	input, ok := requestBody["input"].([]any)
	require.True(t, ok)
	require.Len(t, input, 4)
	assert.Equal(t, "system", input[0].(map[string]any)["role"])
	assert.Equal(t, "user", input[1].(map[string]any)["role"])
	assert.Equal(t, "function_call", input[2].(map[string]any)["type"])
	assert.Equal(t, "function_call_output", input[3].(map[string]any)["type"])

	userContent := input[1].(map[string]any)["content"].([]any)
	require.Len(t, userContent, 3)
	assert.Equal(t, "input_image", userContent[1].(map[string]any)["type"])
	assert.Equal(t, "input_file", userContent[2].(map[string]any)["type"])
	assert.Equal(
		t,
		"data:application/pdf;base64,cGRm",
		userContent[2].(map[string]any)["file_data"],
	)
}

func TestProvider_RejectsStopSequences(t *testing.T) {
	t.Parallel()

	req := &llm.ChatCompletionRequest{
		Model:         "gpt-4o",
		StopSequences: []string{"STOP"},
	}
	provider := &Provider{}

	t.Run(
		"chat completion",
		func(t *testing.T) {
			t.Parallel()

			response, err := provider.ChatCompletion(context.Background(), req)

			require.Error(t, err)
			assert.Nil(t, response)
			assert.ErrorContains(t, err, "stop sequences are not supported by OpenAI Responses API")
		},
	)

	t.Run(
		"streaming chat completion",
		func(t *testing.T) {
			t.Parallel()

			stream, err := provider.ChatCompletionStream(context.Background(), req)

			require.Error(t, err)
			assert.Nil(t, stream)
			assert.ErrorContains(t, err, "stop sequences are not supported by OpenAI Responses API")
		},
	)
}

func TestProvider_StreamsResponsesEvents(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v1/responses", r.URL.Path)

				w.Header().Set("Content-Type", "text/event-stream")
				_, err := fmt.Fprint(
					w,
					"data: "+`{"type":"response.created","sequence_number":0,"response":{"model":"gpt-5.6"}}`+"\n\n"+
						"data: "+`{"type":"response.output_item.added","sequence_number":1,"output_index":1,"item":{"id":"fc_1","type":"function_call","call_id":"call_1","name":"lookup","arguments":""}}`+"\n\n"+
						"data: "+`{"type":"response.function_call_arguments.delta","sequence_number":2,"output_index":1,"item_id":"fc_1","delta":"{\"id\":"}`+"\n\n"+
						"data: "+`{"type":"response.function_call_arguments.delta","sequence_number":3,"output_index":1,"item_id":"fc_1","delta":"\"1\"}"}`+"\n\n"+
						"data: "+`{"type":"response.output_text.delta","sequence_number":4,"output_index":2,"content_index":0,"item_id":"msg_1","delta":"done"}`+"\n\n"+
						"data: "+`{"type":"response.completed","sequence_number":5,"response":{"model":"gpt-5.6","status":"completed","output":[{"type":"function_call","call_id":"call_1","name":"lookup","arguments":"{\"id\":\"1\"}"}],"usage":{"input_tokens":8,"output_tokens":3,"total_tokens":11,"input_tokens_details":{"cached_tokens":0},"output_tokens_details":{"reasoning_tokens":0}}}}`+"\n\n"+
						"data: [DONE]\n\n",
				)
				require.NoError(t, err)
			},
		),
	)
	t.Cleanup(server.Close)

	provider := NewProvider(
		"test-key",
		WithBaseURL(server.URL+"/v1"),
		WithMaxRetries(0),
	)
	stream, err := provider.ChatCompletionStream(
		context.Background(),
		&llm.ChatCompletionRequest{
			Model: "gpt-5.6",
			Messages: []llm.Message{
				{
					Role:  llm.RoleUser,
					Parts: []llm.Part{llm.TextPart{Text: "Look it up."}},
				},
			},
		},
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stream.Close()) })

	var events []llm.ChatCompletionStreamEvent
	for stream.Next() {
		events = append(events, stream.Event())
	}

	require.NoError(t, stream.Err())
	require.Len(t, events, 6)
	assert.Equal(t, "gpt-5.6", events[0].Model)
	assert.Equal(t, 0, events[1].Delta.ToolCalls[0].Index)
	assert.Equal(t, "call_1", events[1].Delta.ToolCalls[0].ID)
	assert.Equal(t, "lookup", events[1].Delta.ToolCalls[0].Name)
	assert.Equal(t, `{"id":`, events[2].Delta.ToolCalls[0].Arguments)
	assert.Equal(t, `"1"}`, events[3].Delta.ToolCalls[0].Arguments)
	assert.Equal(t, "done", events[4].Delta.Content)
	assert.Equal(t, &llm.Usage{InputTokens: 8, OutputTokens: 3}, events[5].Usage)
	assert.Equal(t, new(llm.FinishReasonToolCalls), events[5].FinishReason)
}

func TestBuildParams_RejectsUnsupportedParameters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     *llm.ChatCompletionRequest
		wantErr string
	}{
		{
			name:    "frequency penalty",
			req:     &llm.ChatCompletionRequest{Model: "gpt-4o", FrequencyPenalty: new(0.5)},
			wantErr: "frequency penalty is not supported by OpenAI Responses API",
		},
		{
			name:    "presence penalty",
			req:     &llm.ChatCompletionRequest{Model: "gpt-4o", PresencePenalty: new(0.5)},
			wantErr: "presence penalty is not supported by OpenAI Responses API",
		},
		{
			name: "invalid tool parameters",
			req: &llm.ChatCompletionRequest{
				Model: "gpt-4o",
				Tools: []llm.Tool{{Name: "lookup", Parameters: json.RawMessage(`{"type":`)}},
			},
			wantErr: `cannot decode parameters of tool "lookup"`,
		},
		{
			name: "non-object tool parameters",
			req: &llm.ChatCompletionRequest{
				Model: "gpt-4o",
				Tools: []llm.Tool{{Name: "lookup", Parameters: json.RawMessage(`null`)}},
			},
			wantErr: `cannot decode parameters of tool "lookup"`,
		},
		{
			name: "invalid response schema",
			req: &llm.ChatCompletionRequest{
				Model: "gpt-4o",
				ResponseFormat: &llm.ResponseFormat{
					Type:       llm.ResponseFormatJSONSchema,
					JSONSchema: &llm.JSONSchema{Name: "result", Schema: json.RawMessage(`[]`)},
				},
			},
			wantErr: `cannot decode JSON schema "result"`,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				_, err := buildParams(tt.req)

				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
			},
		)
	}
}

func TestBuildParams_SendsSchemasVerbatim(t *testing.T) {
	t.Parallel()

	schema := `{"type":"integer","enum":[9007199254740993]}`
	params, err := buildParams(
		&llm.ChatCompletionRequest{
			Model: "gpt-4o",
			Tools: []llm.Tool{
				{Name: "lookup", Parameters: json.RawMessage(schema)},
			},
			ResponseFormat: &llm.ResponseFormat{
				Type:       llm.ResponseFormatJSONSchema,
				JSONSchema: &llm.JSONSchema{Name: "result", Schema: json.RawMessage(schema)},
			},
		},
	)
	require.NoError(t, err)

	body, err := json.Marshal(params)
	require.NoError(t, err)

	var request struct {
		Tools []struct {
			Parameters json.RawMessage `json:"parameters"`
		} `json:"tools"`
		Text struct {
			Format struct {
				Schema json.RawMessage `json:"schema"`
			} `json:"format"`
		} `json:"text"`
	}
	require.NoError(t, json.Unmarshal(body, &request))
	require.Len(t, request.Tools, 1)
	assert.JSONEq(t, schema, string(request.Tools[0].Parameters))
	assert.Contains(t, string(request.Tools[0].Parameters), "9007199254740993")
	assert.Contains(t, string(request.Text.Format.Schema), "9007199254740993")
}

func TestProvider_ReplaysReasoningItems(t *testing.T) {
	t.Parallel()

	var requestBody map[string]any

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				err := json.NewDecoder(r.Body).Decode(&requestBody)
				assert.NoError(t, err)

				w.Header().Set("Content-Type", "application/json")
				_, err = fmt.Fprint(
					w,
					`{
						"id": "resp_2",
						"object": "response",
						"created_at": 1,
						"status": "completed",
						"model": "gpt-5.4",
						"output": [
							{"id": "rs_a", "type": "reasoning", "summary": [], "encrypted_content": "enc-a"},
							{"id": "fc_a", "type": "function_call", "status": "completed", "call_id": "call_a", "name": "lookup", "arguments": "{}"},
							{"id": "rs_b", "type": "reasoning", "summary": [], "encrypted_content": "enc-b"},
							{"id": "fc_b", "type": "function_call", "status": "completed", "call_id": "call_b", "name": "lookup", "arguments": "{}"}
						],
						"usage": {"input_tokens": 1, "output_tokens": 1, "total_tokens": 2}
					}`,
				)
				assert.NoError(t, err)
			},
		),
	)
	t.Cleanup(server.Close)

	provider := NewProvider(
		"test-key",
		WithBaseURL(server.URL+"/v1"),
		WithMaxRetries(0),
	)

	response, err := provider.ChatCompletion(
		context.Background(),
		&llm.ChatCompletionRequest{
			Model: "gpt-5.4",
			Messages: []llm.Message{
				{
					Role:  llm.RoleUser,
					Parts: []llm.Part{llm.TextPart{Text: "Look it up."}},
				},
				{
					Role: llm.RoleAssistant,
					Parts: []llm.Part{
						llm.ThinkingPart{
							Signature: `[` +
								`{"id":"rs_1","encrypted_content":"enc-1"},` +
								`{"id":"rs_2","encrypted_content":"enc-2","calls_before":1},` +
								`{"id":"rs_3","encrypted_content":"enc-3","calls_before":2}` +
								`]`,
							Provider: "openai",
						},
						llm.ThinkingPart{Text: "legacy anthropic", Signature: "sig"},
						llm.ThinkingPart{Text: "anthropic", Signature: "sig", Provider: "anthropic"},
						llm.TextPart{Text: "Checking."},
					},
					ToolCalls: []llm.ToolCall{
						{
							ID:       "call_1",
							Function: llm.FunctionCall{Name: "lookup", Arguments: `{"id":"1"}`},
						},
						{
							ID:       "call_2",
							Function: llm.FunctionCall{Name: "lookup", Arguments: `{"id":"2"}`},
						},
					},
				},
				{
					Role:       llm.RoleTool,
					Parts:      []llm.Part{llm.TextPart{Text: "found 1"}},
					ToolCallID: "call_1",
				},
				{
					Role:       llm.RoleTool,
					Parts:      []llm.Part{llm.TextPart{Text: "found 2"}},
					ToolCallID: "call_2",
				},
			},
		},
	)
	require.NoError(t, err)

	assert.NotContains(t, requestBody, "include")

	input, ok := requestBody["input"].([]any)
	require.True(t, ok)

	reasoning := func(id, encryptedContent string) map[string]any {
		return map[string]any{
			"type":              "reasoning",
			"id":                id,
			"summary":           []any{},
			"encrypted_content": encryptedContent,
		}
	}
	functionCall := func(callID, arguments string) map[string]any {
		return map[string]any{
			"type":      "function_call",
			"call_id":   callID,
			"name":      "lookup",
			"arguments": arguments,
		}
	}

	assert.Equal(
		t,
		[]any{
			map[string]any{
				"role":    "user",
				"content": []any{map[string]any{"type": "input_text", "text": "Look it up."}},
			},
			reasoning("rs_1", "enc-1"),
			map[string]any{"role": "assistant", "content": "Checking."},
			functionCall("call_1", `{"id":"1"}`),
			reasoning("rs_2", "enc-2"),
			functionCall("call_2", `{"id":"2"}`),
			reasoning("rs_3", "enc-3"),
			map[string]any{"type": "function_call_output", "call_id": "call_1", "output": "found 1"},
			map[string]any{"type": "function_call_output", "call_id": "call_2", "output": "found 2"},
		},
		input,
	)

	require.NotEmpty(t, response.Message.Parts)
	assert.Equal(
		t,
		llm.ThinkingPart{
			Signature: `[{"id":"rs_a","encrypted_content":"enc-a"},{"id":"rs_b","encrypted_content":"enc-b","calls_before":1}]`,
			Provider:  "openai",
		},
		response.Message.Parts[0],
	)
}

func TestProvider_StreamCarriesReasoningSignature(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				_, err := fmt.Fprint(
					w,
					"data: "+`{"type":"response.created","sequence_number":0,"response":{"model":"gpt-5.4"}}`+"\n\n"+
						"data: "+`{"type":"response.output_text.delta","sequence_number":1,"output_index":1,"content_index":0,"item_id":"msg_1","delta":"done"}`+"\n\n"+
						"data: "+`{"type":"response.completed","sequence_number":2,"response":{"model":"gpt-5.4","status":"completed","output":[{"id":"rs_1","type":"reasoning","summary":[],"encrypted_content":"enc-1"},{"id":"msg_1","type":"message","role":"assistant","status":"completed","content":[{"type":"output_text","text":"done","annotations":[]}]}],"usage":{"input_tokens":2,"output_tokens":1,"total_tokens":3}}}`+"\n\n",
				)
				assert.NoError(t, err)
			},
		),
	)
	t.Cleanup(server.Close)

	provider := NewProvider(
		"test-key",
		WithBaseURL(server.URL+"/v1"),
		WithMaxRetries(0),
	)
	stream, err := provider.ChatCompletionStream(
		context.Background(),
		&llm.ChatCompletionRequest{
			Model: "gpt-5.4",
			Messages: []llm.Message{
				{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "Hi."}}},
			},
		},
	)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, stream.Close()) })

	accumulator := llm.NewStreamAccumulator(stream)
	for accumulator.Next() {
	}

	require.NoError(t, accumulator.Err())

	response := accumulator.Response()
	require.Len(t, response.Message.Parts, 2)
	assert.Equal(
		t,
		llm.ThinkingPart{
			Signature: `[{"id":"rs_1","encrypted_content":"enc-1"}]`,
			Provider:  "openai",
		},
		response.Message.Parts[0],
	)
	assert.Equal(t, "done", response.Message.Text())
}

func TestProvider_StreamStopsOnError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		event   string
		wantErr string
	}{
		{
			name:    "error event",
			event:   `{"type":"error","sequence_number":1,"code":"server_error","message":"boom","param":null}`,
			wantErr: "cannot stream OpenAI response: boom",
		},
		{
			name:    "failed response",
			event:   `{"type":"response.failed","sequence_number":1,"response":{"status":"failed","error":{"code":"server_error","message":"failed"}}}`,
			wantErr: "cannot stream OpenAI response: failed",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				server := httptest.NewServer(
					http.HandlerFunc(
						func(w http.ResponseWriter, r *http.Request) {
							w.Header().Set("Content-Type", "text/event-stream")
							_, err := fmt.Fprint(
								w,
								"data: "+tt.event+"\n\n"+
									"data: "+`{"type":"response.output_text.delta","sequence_number":2,"output_index":0,"content_index":0,"item_id":"msg_1","delta":"late"}`+"\n\n"+
									"data: "+`{"type":"error","sequence_number":3,"code":"server_error","message":"later","param":null}`+"\n\n",
							)
							assert.NoError(t, err)
						},
					),
				)
				t.Cleanup(server.Close)

				provider := NewProvider(
					"test-key",
					WithBaseURL(server.URL+"/v1"),
					WithMaxRetries(0),
				)
				stream, err := provider.ChatCompletionStream(
					context.Background(),
					&llm.ChatCompletionRequest{Model: "gpt-4o"},
				)
				require.NoError(t, err)
				t.Cleanup(func() { assert.NoError(t, stream.Close()) })

				assert.False(t, stream.Next())
				assert.False(t, stream.Next())
				assert.EqualError(t, stream.Err(), tt.wantErr)
			},
		)
	}
}

func TestBuildInput_ReplaysReasoningInOrder(t *testing.T) {
	t.Parallel()

	toolCall := func(id string) llm.ToolCall {
		return llm.ToolCall{ID: id, Function: llm.FunctionCall{Name: "lookup", Arguments: "{}"}}
	}

	tests := []struct {
		name      string
		parts     []llm.Part
		toolCalls []llm.ToolCall
		want      []string
	}{
		{
			name: "no tool calls with trailing reasoning",
			parts: []llm.Part{
				llm.ThinkingPart{
					Signature: `[{"id":"rs_1","encrypted_content":"e"},{"id":"rs_2","encrypted_content":"e","calls_before":1}]`,
					Provider:  "openai",
				},
				llm.TextPart{Text: "answer"},
			},
			want: []string{"rs_1", "text", "rs_2"},
		},
		{
			name: "reasoning after the only call",
			parts: []llm.Part{
				llm.ThinkingPart{
					Signature: `[{"id":"rs_1","encrypted_content":"e"},{"id":"rs_2","encrypted_content":"e","calls_before":1}]`,
					Provider:  "openai",
				},
			},
			toolCalls: []llm.ToolCall{toolCall("call_1")},
			want:      []string{"rs_1", "call_1", "rs_2"},
		},
		{
			name: "position beyond the call count with a gap",
			parts: []llm.Part{
				llm.ThinkingPart{
					Signature: `[{"id":"rs_1","encrypted_content":"e","calls_before":3}]`,
					Provider:  "openai",
				},
			},
			toolCalls: []llm.ToolCall{toolCall("call_1")},
			want:      []string{"call_1", "rs_1"},
		},
		{
			name: "several thinking parts keep their relative order",
			parts: []llm.Part{
				llm.ThinkingPart{
					Signature: `[{"id":"rs_1","encrypted_content":"e"},{"id":"rs_3","encrypted_content":"e","calls_before":1}]`,
					Provider:  "openai",
				},
				llm.ThinkingPart{
					Signature: `[{"id":"rs_2","encrypted_content":"e"}]`,
					Provider:  "openai",
				},
			},
			toolCalls: []llm.ToolCall{toolCall("call_1"), toolCall("call_2")},
			want:      []string{"rs_1", "rs_2", "call_1", "rs_3", "call_2"},
		},
		{
			name: "undecodable and incomplete items are skipped",
			parts: []llm.Part{
				llm.ThinkingPart{Signature: "not json", Provider: "openai"},
				llm.ThinkingPart{Signature: `[{"id":"rs_1"},{"encrypted_content":"e"}]`, Provider: "openai"},
			},
			toolCalls: []llm.ToolCall{toolCall("call_1")},
			want:      []string{"call_1"},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				input := buildInput(
					[]llm.Message{
						{
							Role:      llm.RoleAssistant,
							Parts:     tt.parts,
							ToolCalls: tt.toolCalls,
						},
					},
				)

				got := make([]string, 0, len(input))
				for _, item := range input {
					switch {
					case item.OfReasoning != nil:
						got = append(got, item.OfReasoning.ID)
					case item.OfFunctionCall != nil:
						got = append(got, item.OfFunctionCall.CallID)
					case item.OfMessage != nil:
						got = append(got, "text")
					}
				}

				assert.Equal(t, tt.want, got)
			},
		)
	}
}
