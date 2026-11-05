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

// Package openaicompat talks to any server implementing the OpenAI
// /v1/chat/completions wire format: llama.cpp's llama-server, vLLM, Ollama,
// LM Studio, text-generation-webui, LiteLLM, OpenRouter, Together, Groq, and
// Gemini's compatibility endpoint among them.
//
// It is deliberately NOT the openai package. That one targets OpenAI's
// Responses API (/v1/responses), which is OpenAI-specific and which none of
// the self-hosted servers above implement -- pointing it at a base URL does
// not reach them. Chat completions is the format they actually speak, and it
// is the de facto portable one, so it gets its own provider rather than a flag
// on that one.
//
// The request is hand-rolled rather than built with an SDK on purpose. The
// value of this provider is tolerance of servers that implement a SUBSET of
// the spec, and an SDK tuned to one vendor's API tends to emit fields such a
// server rejects outright. Every optional field here is omitted when unset.
package openaicompat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/llm"
)

const (
	defaultRequestTimeout = 10 * time.Minute
	chatCompletionsPath   = "/chat/completions"
)

type (
	Provider struct {
		httpClient     *http.Client
		baseURL        string
		apiKey         string
		requestTimeout time.Duration
		extraHeaders   map[string]string
	}

	Option func(*Provider)

	// chatRequest mirrors the subset of the chat completions body that is
	// portable across implementations. Everything optional is omitempty so a
	// minimal server never sees a field it does not understand.
	chatRequest struct {
		Model       string        `json:"model"`
		Messages    []chatMessage `json:"messages"`
		MaxTokens   *int          `json:"max_tokens,omitempty"`
		Temperature *float64      `json:"temperature,omitempty"`
		TopP        *float64      `json:"top_p,omitempty"`
		Stop        []string      `json:"stop,omitempty"`
		Tools       []chatTool    `json:"tools,omitempty"`
		ToolChoice  any           `json:"tool_choice,omitempty"`
		Stream      bool          `json:"stream,omitempty"`
		// StreamOptions asks for a final usage chunk. Servers that do not
		// know the field ignore it; servers that do give us token counts we
		// would otherwise have to leave at zero.
		StreamOptions  *streamOptions  `json:"stream_options,omitempty"`
		ResponseFormat *responseFormat `json:"response_format,omitempty"`
	}

	streamOptions struct {
		IncludeUsage bool `json:"include_usage"`
	}

	responseFormat struct {
		Type       string          `json:"type"`
		JSONSchema *jsonSchemaSpec `json:"json_schema,omitempty"`
	}

	jsonSchemaSpec struct {
		Name        string          `json:"name"`
		Description string          `json:"description,omitempty"`
		Schema      json.RawMessage `json:"schema,omitempty"`
		Strict      bool            `json:"strict,omitempty"`
	}

	chatMessage struct {
		Role       string         `json:"role"`
		Content    string         `json:"content,omitempty"`
		ToolCalls  []chatToolCall `json:"tool_calls,omitempty"`
		ToolCallID string         `json:"tool_call_id,omitempty"`
	}

	chatTool struct {
		Type     string           `json:"type"`
		Function chatToolFunction `json:"function"`
	}

	chatToolFunction struct {
		Name        string          `json:"name"`
		Description string          `json:"description,omitempty"`
		Parameters  json.RawMessage `json:"parameters,omitempty"`
	}

	chatToolCall struct {
		Index    int              `json:"index"`
		ID       string           `json:"id"`
		Type     string           `json:"type"`
		Function chatToolCallFunc `json:"function"`
	}

	chatToolCallFunc struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	}

	chatResponse struct {
		Model   string       `json:"model"`
		Choices []chatChoice `json:"choices"`
		Usage   *chatUsage   `json:"usage"`
		Error   *chatError   `json:"error"`
	}

	chatChoice struct {
		Message      chatMessage `json:"message"`
		Delta        *chatDelta  `json:"delta"`
		FinishReason string      `json:"finish_reason"`
	}

	chatDelta struct {
		Role      string         `json:"role"`
		Content   string         `json:"content"`
		ToolCalls []chatToolCall `json:"tool_calls"`
	}

	chatUsage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	}

	chatError struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	}
)

func WithHTTPClient(c *http.Client) Option {
	return func(p *Provider) { p.httpClient = c }
}

func WithRequestTimeout(d time.Duration) Option {
	return func(p *Provider) { p.requestTimeout = d }
}

// WithHeader adds a header to every request. Some gateways in front of a
// self-hosted model authenticate on something other than Authorization.
func WithHeader(name, value string) Option {
	return func(p *Provider) {
		if p.extraHeaders == nil {
			p.extraHeaders = map[string]string{}
		}
		p.extraHeaders[name] = value
	}
}

// NewProvider returns a provider talking to baseURL, which must point at the
// API root (the part before /chat/completions), e.g.
// "http://127.0.0.1:8080/v1".
//
// apiKey may be empty. A locally hosted llama-server or Ollama usually has no
// authentication at all, and sending an empty bearer token upsets some of
// them, so the Authorization header is omitted entirely when the key is empty.
func NewProvider(baseURL, apiKey string, opts ...Option) *Provider {
	p := &Provider{
		baseURL:        strings.TrimSuffix(baseURL, "/"),
		apiKey:         apiKey,
		requestTimeout: defaultRequestTimeout,
		httpClient:     http.DefaultClient,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *Provider) ChatCompletion(
	ctx context.Context,
	req *llm.ChatCompletionRequest,
) (*llm.ChatCompletionResponse, error) {
	body, err := buildRequest(req, false)
	if err != nil {
		return nil, err
	}

	httpResp, err := p.do(ctx, body)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	var resp chatResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode chat completion response: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("chat completion failed: %s", resp.Error.Message)
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("chat completion returned no choices")
	}

	choice := resp.Choices[0]

	message := llm.Message{Role: llm.RoleAssistant}
	if choice.Message.Content != "" {
		message.Parts = []llm.Part{llm.TextPart{Text: choice.Message.Content}}
	}

	for _, tc := range choice.Message.ToolCalls {
		message.ToolCalls = append(message.ToolCalls, llm.ToolCall{
			ID: tc.ID,
			Function: llm.FunctionCall{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			},
		})
	}

	out := &llm.ChatCompletionResponse{
		Model:        resp.Model,
		Message:      message,
		FinishReason: mapFinishReason(choice.FinishReason),
	}

	if resp.Usage != nil {
		out.Usage = llm.Usage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
		}
	}

	return out, nil
}

func (p *Provider) ChatCompletionStream(
	ctx context.Context,
	req *llm.ChatCompletionRequest,
) (llm.ChatCompletionStream, error) {
	body, err := buildRequest(req, true)
	if err != nil {
		return nil, err
	}

	httpResp, err := p.do(ctx, body)
	if err != nil {
		return nil, err
	}

	return &stream{
		body:    httpResp.Body,
		scanner: bufio.NewScanner(httpResp.Body),
	}, nil
}

func (p *Provider) do(ctx context.Context, body []byte) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, p.requestTimeout)

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		p.baseURL+chatCompletionsPath,
		bytes.NewReader(body),
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("cannot build chat completion request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	}
	for name, value := range p.extraHeaders {
		httpReq.Header.Set(name, value)
	}

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("cannot call chat completion endpoint: %w", err)
	}

	// The caller closes the body, which releases the context. Streaming reads
	// outlive this function, so cancel cannot be deferred here.
	resp.Body = &cancelOnClose{ReadCloser: resp.Body, cancel: cancel}

	if resp.StatusCode >= http.StatusBadRequest {
		defer resp.Body.Close()

		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))

		return nil, classifyError(resp, string(payload))
	}

	return resp, nil
}

type cancelOnClose struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (c *cancelOnClose) Close() error {
	err := c.ReadCloser.Close()
	c.cancel()

	return err
}

func classifyError(resp *http.Response, payload string) error {
	var envelope chatResponse
	message := strings.TrimSpace(payload)
	if err := json.Unmarshal([]byte(payload), &envelope); err == nil && envelope.Error != nil {
		message = envelope.Error.Message
	}

	err := fmt.Errorf("chat completion failed with status %d: %s", resp.StatusCode, message)

	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		var retryAfter time.Duration
		if v := resp.Header.Get("Retry-After"); v != "" {
			if seconds, convErr := strconv.Atoi(v); convErr == nil {
				retryAfter = time.Duration(seconds) * time.Second
			}
		}

		return &llm.ErrRateLimit{RetryAfter: retryAfter, Err: err}
	case http.StatusUnauthorized, http.StatusForbidden:
		return &llm.ErrAuthentication{Err: err}
	case http.StatusRequestEntityTooLarge:
		return &llm.ErrContextLength{Err: err}
	}

	// Context-length failures are a 400 on most servers and are only
	// identifiable from the message, which is why this is a substring match
	// rather than a status check.
	if resp.StatusCode == http.StatusBadRequest {
		lowered := strings.ToLower(message)
		if strings.Contains(lowered, "context length") ||
			strings.Contains(lowered, "context window") ||
			strings.Contains(lowered, "too many tokens") ||
			strings.Contains(lowered, "maximum context") {
			return &llm.ErrContextLength{Err: err}
		}
	}

	return err
}

func buildRequest(req *llm.ChatCompletionRequest, streaming bool) ([]byte, error) {
	if req.Thinking != nil && req.Thinking.Enabled {
		return nil, errors.New("thinking is not supported by the OpenAI-compatible provider")
	}

	out := chatRequest{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        req.StopSequences,
		Stream:      streaming,
	}

	if streaming {
		out.StreamOptions = &streamOptions{IncludeUsage: true}
	}

	messages, err := buildMessages(req.Messages)
	if err != nil {
		return nil, err
	}
	out.Messages = messages

	for _, tool := range req.Tools {
		out.Tools = append(out.Tools, chatTool{
			Type: "function",
			Function: chatToolFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Parameters,
			},
		})
	}

	if choice := buildToolChoice(req.ToolChoice); choice != nil {
		out.ToolChoice = choice
	}

	if format, err := buildResponseFormat(req.ResponseFormat); err != nil {
		return nil, err
	} else if format != nil {
		out.ResponseFormat = format
	}

	body, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("cannot encode chat completion request: %w", err)
	}

	return body, nil
}

func buildMessages(messages []llm.Message) ([]chatMessage, error) {
	out := make([]chatMessage, 0, len(messages))

	for _, msg := range messages {
		converted := chatMessage{
			Role:       string(msg.Role),
			ToolCallID: msg.ToolCallID,
		}

		var text strings.Builder
		for _, part := range msg.Parts {
			switch p := part.(type) {
			case llm.TextPart:
				text.WriteString(p.Text)
			case llm.ThinkingPart:
				// Thinking is not representable in chat completions and is
				// dropped rather than inlined, which would feed the model its
				// own reasoning back as if the user had written it.
			default:
				return nil, fmt.Errorf(
					"content part %T is not supported by the OpenAI-compatible provider",
					part,
				)
			}
		}
		converted.Content = text.String()

		for i, tc := range msg.ToolCalls {
			converted.ToolCalls = append(converted.ToolCalls, chatToolCall{
				Index: i,
				ID:    tc.ID,
				Type:  "function",
				Function: chatToolCallFunc{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			})
		}

		out = append(out, converted)
	}

	return out, nil
}

func buildToolChoice(choice *llm.ToolChoice) any {
	if choice == nil {
		return nil
	}

	switch choice.Type {
	case llm.ToolChoiceAuto:
		return "auto"
	case llm.ToolChoiceNone:
		return "none"
	case llm.ToolChoiceRequired:
		return "required"
	case llm.ToolChoiceFunction:
		return map[string]any{
			"type":     "function",
			"function": map[string]string{"name": choice.Function},
		}
	}

	return nil
}

func buildResponseFormat(format *llm.ResponseFormat) (*responseFormat, error) {
	if format == nil {
		return nil, nil
	}

	switch format.Type {
	case llm.ResponseFormatText:
		return &responseFormat{Type: "text"}, nil
	case llm.ResponseFormatJSONObject:
		return &responseFormat{Type: "json_object"}, nil
	case llm.ResponseFormatJSONSchema:
		if format.JSONSchema == nil {
			return nil, errors.New("response format json_schema requires a schema")
		}

		return &responseFormat{
			Type: "json_schema",
			JSONSchema: &jsonSchemaSpec{
				Name:        format.JSONSchema.Name,
				Description: format.JSONSchema.Description,
				Schema:      format.JSONSchema.Schema,
				Strict:      format.JSONSchema.Strict,
			},
		}, nil
	}

	return nil, fmt.Errorf("unsupported response format %q", format.Type)
}

func mapFinishReason(reason string) llm.FinishReason {
	switch reason {
	case "tool_calls", "function_call":
		return llm.FinishReasonToolCalls
	case "length":
		return llm.FinishReasonLength
	case "content_filter":
		return llm.FinishReasonContentFilter
	default:
		return llm.FinishReasonStop
	}
}

type stream struct {
	body    io.ReadCloser
	scanner *bufio.Scanner
	current llm.ChatCompletionStreamEvent
	err     error
	done    bool
}

func (s *stream) Next() bool {
	if s.done || s.err != nil {
		return false
	}

	for s.scanner.Scan() {
		line := strings.TrimSpace(s.scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}

		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			s.done = true

			return false
		}

		var chunk chatResponse
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			s.err = fmt.Errorf("cannot decode stream chunk: %w", err)

			return false
		}

		if chunk.Error != nil {
			s.err = fmt.Errorf("stream failed: %s", chunk.Error.Message)

			return false
		}

		event := llm.ChatCompletionStreamEvent{Model: chunk.Model}

		if chunk.Usage != nil {
			event.Usage = &llm.Usage{
				InputTokens:  chunk.Usage.PromptTokens,
				OutputTokens: chunk.Usage.CompletionTokens,
			}
		}

		if len(chunk.Choices) > 0 {
			choice := chunk.Choices[0]

			if choice.Delta != nil {
				event.Delta.Content = choice.Delta.Content

				for _, tc := range choice.Delta.ToolCalls {
					event.Delta.ToolCalls = append(event.Delta.ToolCalls, llm.ToolCallDelta{
						Index:     tc.Index,
						ID:        tc.ID,
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					})
				}
			}

			if choice.FinishReason != "" {
				reason := mapFinishReason(choice.FinishReason)
				event.FinishReason = &reason
			}
		}

		// A usage-only final chunk carries no choices; emitting it is how the
		// accumulator learns the token counts.
		s.current = event

		return true
	}

	if err := s.scanner.Err(); err != nil {
		s.err = fmt.Errorf("cannot read stream: %w", err)

		return false
	}

	s.done = true

	return false
}

func (s *stream) Event() llm.ChatCompletionStreamEvent { return s.current }

func (s *stream) Err() error { return s.err }

func (s *stream) Close() error { return s.body.Close() }
