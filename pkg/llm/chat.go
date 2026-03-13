// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package llm

import (
	"encoding/json"
	"strings"
)

type (
	ChatCompletionRequest struct {
		Model              string
		Messages           []Message
		MaxTokens          *int
		Temperature        *float64
		TopP               *float64
		FrequencyPenalty   *float64
		PresencePenalty    *float64
		StopSequences      []string
		Tools              []Tool
		ToolChoice         *ToolChoice
		ParallelToolCalls  *bool
		ResponseFormat     *ResponseFormat
	}

	ToolChoiceType string

	ToolChoice struct {
		Type     ToolChoiceType
		Function string // required when Type is ToolChoiceFunction
	}

	ResponseFormatType string

	ResponseFormat struct {
		Type       ResponseFormatType
		JSONSchema *JSONSchema // required when Type is ResponseFormatJSONSchema
	}

	JSONSchema struct {
		Name        string
		Description string
		Schema      json.RawMessage
		Strict      bool
	}

	FinishReason string

	Usage struct {
		InputTokens  int
		OutputTokens int
	}

	ChatCompletionResponse struct {
		Model        string
		Message      Message
		Usage        Usage
		FinishReason FinishReason
	}

	// ChatCompletionStream is an iterator over streaming chat completion events.
	// Callers must call Close when done, even if Next returns false.
	// The typical usage pattern is:
	//
	//	stream, err := client.ChatCompletionStream(ctx, req)
	//	if err != nil { ... }
	//	defer stream.Close()
	//	for stream.Next() {
	//	    event := stream.Event()
	//	    // process event
	//	}
	//	if err := stream.Err(); err != nil { ... }
	ChatCompletionStream interface {
		Next() bool
		Event() ChatCompletionStreamEvent
		Err() error
		Close() error
	}

	ChatCompletionStreamEvent struct {
		Delta        MessageDelta
		Usage        *Usage        // present on final event if provider supports it
		FinishReason *FinishReason // present on final event
	}

	MessageDelta struct {
		Content   string
		ToolCalls []ToolCallDelta
	}

	ToolCallDelta struct {
		Index     int
		ID        string // set on first chunk for this tool call
		Name      string // set on first chunk for this tool call
		Arguments string // incremental JSON fragment
	}
)

const (
	ToolChoiceAuto     ToolChoiceType = "auto"
	ToolChoiceNone     ToolChoiceType = "none"
	ToolChoiceRequired ToolChoiceType = "required"
	ToolChoiceFunction ToolChoiceType = "function"
)

const (
	ResponseFormatText       ResponseFormatType = "text"
	ResponseFormatJSONObject ResponseFormatType = "json_object"
	ResponseFormatJSONSchema ResponseFormatType = "json_schema"
)

const (
	FinishReasonStop          FinishReason = "stop"
	FinishReasonToolCalls     FinishReason = "tool_calls"
	FinishReasonLength        FinishReason = "length"
	FinishReasonContentFilter FinishReason = "content_filter"
)

func (u Usage) Add(other Usage) Usage {
	return Usage{
		InputTokens:  u.InputTokens + other.InputTokens,
		OutputTokens: u.OutputTokens + other.OutputTokens,
	}
}

// StreamAccumulator wraps a ChatCompletionStream and reassembles the
// streamed deltas into a full ChatCompletionResponse. It proxies
// Next/Event/Err/Close transparently so callers can still observe
// individual deltas while accumulating the final result.
//
// After the stream is exhausted (Next returns false), call Response
// to get the fully assembled ChatCompletionResponse.
type StreamAccumulator struct {
	stream       ChatCompletionStream
	current      ChatCompletionStreamEvent
	content      strings.Builder
	toolCalls    map[int]*ToolCall
	usage        Usage
	finishReason FinishReason
	model        string
}

func NewStreamAccumulator(stream ChatCompletionStream) *StreamAccumulator {
	return &StreamAccumulator{
		stream:    stream,
		toolCalls: make(map[int]*ToolCall),
	}
}

func (a *StreamAccumulator) Next() bool {
	if !a.stream.Next() {
		return false
	}

	a.current = a.stream.Event()
	a.accumulate(a.current)

	return true
}

func (a *StreamAccumulator) Event() ChatCompletionStreamEvent {
	return a.current
}

func (a *StreamAccumulator) Err() error {
	return a.stream.Err()
}

func (a *StreamAccumulator) Close() error {
	return a.stream.Close()
}

// Response returns the fully assembled ChatCompletionResponse after
// the stream has been exhausted. Must only be called after Next
// returns false and Err returns nil.
func (a *StreamAccumulator) Response() *ChatCompletionResponse {
	toolCalls := make([]ToolCall, 0, len(a.toolCalls))
	for i := 0; i < len(a.toolCalls); i++ {
		if tc, ok := a.toolCalls[i]; ok {
			toolCalls = append(toolCalls, *tc)
		}
	}

	return &ChatCompletionResponse{
		Model: a.model,
		Message: Message{
			Role:      RoleAssistant,
			Parts:     []Part{TextPart{Text: a.content.String()}},
			ToolCalls: toolCalls,
		},
		Usage:        a.usage,
		FinishReason: a.finishReason,
	}
}

func (a *StreamAccumulator) accumulate(event ChatCompletionStreamEvent) {
	a.content.WriteString(event.Delta.Content)

	for _, tcd := range event.Delta.ToolCalls {
		tc, ok := a.toolCalls[tcd.Index]
		if !ok {
			tc = &ToolCall{}
			a.toolCalls[tcd.Index] = tc
		}

		if tcd.ID != "" {
			tc.ID = tcd.ID
		}
		if tcd.Name != "" {
			tc.Function.Name = tcd.Name
		}
		tc.Function.Arguments += tcd.Arguments
	}

	if event.Usage != nil {
		a.usage = *event.Usage
	}
	if event.FinishReason != nil {
		a.finishReason = *event.FinishReason
	}
}
