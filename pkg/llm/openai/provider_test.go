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
