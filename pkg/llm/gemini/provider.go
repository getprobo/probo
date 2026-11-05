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

// Package gemini talks to Google's Generative Language API natively --
// :generateContent and :streamGenerateContent -- rather than through Gemini's
// OpenAI compatibility shim.
//
// The shim is reachable via the openaicompat provider and is the quicker route,
// but it is a lossy translation on Google's side. Going direct is what gives
// access to system instructions as a first-class field, inline file data, and
// usage metadata, and it is the surface Google documents and versions.
//
// Written against net/http and encoding/json rather than a Google SDK so this
// adds no dependency to the module.
package gemini

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/llm"
)

const (
	defaultBaseURL        = "https://generativelanguage.googleapis.com/v1beta"
	defaultRequestTimeout = 10 * time.Minute

	roleUser  = "user"
	roleModel = "model"
)

type (
	Provider struct {
		httpClient     *http.Client
		baseURL        string
		apiKey         string
		requestTimeout time.Duration
	}

	Option func(*Provider)

	generateRequest struct {
		Contents          []content         `json:"contents"`
		SystemInstruction *content          `json:"systemInstruction,omitempty"`
		Tools             []toolSpec        `json:"tools,omitempty"`
		ToolConfig        *toolConfig       `json:"toolConfig,omitempty"`
		GenerationConfig  *generationConfig `json:"generationConfig,omitempty"`
	}

	content struct {
		Role  string `json:"role,omitempty"`
		Parts []part `json:"parts"`
	}

	// part is a union: exactly one field is set. Gemini rejects a part with
	// more than one populated member, so every field is omitempty and the
	// builders below set a single one.
	part struct {
		Text             string            `json:"text,omitempty"`
		InlineData       *inlineData       `json:"inlineData,omitempty"`
		FunctionCall     *functionCall     `json:"functionCall,omitempty"`
		FunctionResponse *functionResponse `json:"functionResponse,omitempty"`
	}

	inlineData struct {
		MimeType string `json:"mimeType"`
		Data     string `json:"data"`
	}

	functionCall struct {
		Name string          `json:"name"`
		Args json.RawMessage `json:"args,omitempty"`
	}

	functionResponse struct {
		Name     string          `json:"name"`
		Response json.RawMessage `json:"response"`
	}

	toolSpec struct {
		FunctionDeclarations []functionDeclaration `json:"functionDeclarations,omitempty"`
	}

	functionDeclaration struct {
		Name        string          `json:"name"`
		Description string          `json:"description,omitempty"`
		Parameters  json.RawMessage `json:"parameters,omitempty"`
	}

	toolConfig struct {
		FunctionCallingConfig functionCallingConfig `json:"functionCallingConfig"`
	}

	functionCallingConfig struct {
		Mode                 string   `json:"mode"`
		AllowedFunctionNames []string `json:"allowedFunctionNames,omitempty"`
	}

	generationConfig struct {
		MaxOutputTokens  *int            `json:"maxOutputTokens,omitempty"`
		Temperature      *float64        `json:"temperature,omitempty"`
		TopP             *float64        `json:"topP,omitempty"`
		StopSequences    []string        `json:"stopSequences,omitempty"`
		ResponseMimeType string          `json:"responseMimeType,omitempty"`
		ResponseSchema   json.RawMessage `json:"responseSchema,omitempty"`
	}

	generateResponse struct {
		Candidates    []candidate    `json:"candidates"`
		UsageMetadata *usageMetadata `json:"usageMetadata"`
		ModelVersion  string         `json:"modelVersion"`
		Error         *apiError      `json:"error"`
	}

	candidate struct {
		Content      content `json:"content"`
		FinishReason string  `json:"finishReason"`
	}

	usageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
	}

	apiError struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	}
)

func WithHTTPClient(c *http.Client) Option {
	return func(p *Provider) { p.httpClient = c }
}

// WithBaseURL overrides the API root. Point it at a Vertex AI endpoint or a
// regional mirror; the default is the public Generative Language API.
func WithBaseURL(baseURL string) Option {
	return func(p *Provider) {
		if baseURL != "" {
			p.baseURL = strings.TrimSuffix(baseURL, "/")
		}
	}
}

func WithRequestTimeout(d time.Duration) Option {
	return func(p *Provider) { p.requestTimeout = d }
}

func NewProvider(apiKey string, opts ...Option) *Provider {
	p := &Provider{
		apiKey:         apiKey,
		baseURL:        defaultBaseURL,
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
	body, err := buildRequest(req)
	if err != nil {
		return nil, err
	}

	httpResp, err := p.do(ctx, req.Model, "generateContent", body, false)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	var resp generateResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot decode generateContent response: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("generateContent failed: %s", resp.Error.Message)
	}

	if len(resp.Candidates) == 0 {
		// Gemini returns zero candidates when every one was filtered, which
		// is a content-filter outcome rather than an empty answer.
		return nil, &llm.ErrContentFilter{
			Err: errors.New("generateContent returned no candidates"),
		}
	}

	cand := resp.Candidates[0]

	message := llm.Message{Role: llm.RoleAssistant}

	var text strings.Builder
	for _, prt := range cand.Content.Parts {
		switch {
		case prt.FunctionCall != nil:
			args := string(prt.FunctionCall.Args)
			if args == "" {
				args = "{}"
			}

			message.ToolCalls = append(message.ToolCalls, llm.ToolCall{
				// Gemini matches function responses by name rather than by
				// id, and returns no id of its own. The name is therefore the
				// only stable correlator available to the caller.
				ID: prt.FunctionCall.Name,
				Function: llm.FunctionCall{
					Name:      prt.FunctionCall.Name,
					Arguments: args,
				},
			})
		case prt.Text != "":
			text.WriteString(prt.Text)
		}
	}

	if text.Len() > 0 {
		message.Parts = []llm.Part{llm.TextPart{Text: text.String()}}
	}

	model := resp.ModelVersion
	if model == "" {
		model = req.Model
	}

	out := &llm.ChatCompletionResponse{
		Model:        model,
		Message:      message,
		FinishReason: mapFinishReason(cand.FinishReason, len(message.ToolCalls) > 0),
	}

	if resp.UsageMetadata != nil {
		out.Usage = llm.Usage{
			InputTokens:  resp.UsageMetadata.PromptTokenCount,
			OutputTokens: resp.UsageMetadata.CandidatesTokenCount,
		}
	}

	return out, nil
}

func (p *Provider) ChatCompletionStream(
	ctx context.Context,
	req *llm.ChatCompletionRequest,
) (llm.ChatCompletionStream, error) {
	body, err := buildRequest(req)
	if err != nil {
		return nil, err
	}

	httpResp, err := p.do(ctx, req.Model, "streamGenerateContent", body, true)
	if err != nil {
		return nil, err
	}

	return &stream{
		body:     httpResp.Body,
		scanner:  bufio.NewScanner(httpResp.Body),
		fallback: req.Model,
	}, nil
}

func (p *Provider) do(
	ctx context.Context,
	model, method string,
	body []byte,
	streaming bool,
) (*http.Response, error) {
	if model == "" {
		return nil, errors.New("model is required")
	}

	endpoint := fmt.Sprintf("%s/models/%s:%s", p.baseURL, url.PathEscape(model), method)
	if streaming {
		// Without alt=sse the streaming endpoint returns a JSON array that
		// only completes at the end, which is not a stream in any useful sense.
		endpoint += "?alt=sse"
	}

	ctx, cancel := context.WithTimeout(ctx, p.requestTimeout)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("cannot build %s request: %w", method, err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	// Header rather than ?key=, so the credential stays out of request URLs
	// and therefore out of access logs.
	httpReq.Header.Set("x-goog-api-key", p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("cannot call %s: %w", method, err)
	}

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
	message := strings.TrimSpace(payload)

	var envelope generateResponse
	if err := json.Unmarshal([]byte(payload), &envelope); err == nil && envelope.Error != nil {
		message = envelope.Error.Message
	}

	err := fmt.Errorf("gemini request failed with status %d: %s", resp.StatusCode, message)

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
	}

	if resp.StatusCode == http.StatusBadRequest {
		lowered := strings.ToLower(message)
		if strings.Contains(lowered, "token count") ||
			strings.Contains(lowered, "context length") ||
			strings.Contains(lowered, "exceeds the maximum") {
			return &llm.ErrContextLength{Err: err}
		}
	}

	return err
}

func buildRequest(req *llm.ChatCompletionRequest) ([]byte, error) {
	if req.Thinking != nil && req.Thinking.Enabled {
		return nil, errors.New("thinking is not supported by the gemini provider")
	}

	if req.FrequencyPenalty != nil {
		return nil, errors.New("frequency penalty is not supported by the gemini provider")
	}

	if req.PresencePenalty != nil {
		return nil, errors.New("presence penalty is not supported by the gemini provider")
	}

	out := generateRequest{}

	contents, system, err := buildContents(req.Messages)
	if err != nil {
		return nil, err
	}

	out.Contents = contents
	out.SystemInstruction = system

	if len(req.Tools) > 0 {
		decls := make([]functionDeclaration, 0, len(req.Tools))
		for _, tool := range req.Tools {
			decls = append(decls, functionDeclaration{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Parameters,
			})
		}

		out.Tools = []toolSpec{{FunctionDeclarations: decls}}
	}

	if cfg := buildToolConfig(req.ToolChoice); cfg != nil {
		out.ToolConfig = cfg
	}

	genCfg := &generationConfig{
		MaxOutputTokens: req.MaxTokens,
		Temperature:     req.Temperature,
		TopP:            req.TopP,
		StopSequences:   req.StopSequences,
	}

	if err := applyResponseFormat(genCfg, req.ResponseFormat); err != nil {
		return nil, err
	}

	out.GenerationConfig = genCfg

	body, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("cannot encode gemini request: %w", err)
	}

	return body, nil
}

// buildContents splits the conversation into Gemini's shape: system turns are
// hoisted into a dedicated systemInstruction field rather than staying inline,
// and the remaining roles collapse to user/model.
func buildContents(messages []llm.Message) ([]content, *content, error) {
	var (
		contents []content
		system   *content
	)

	for _, msg := range messages {
		if msg.Role == llm.RoleSystem {
			text, err := joinText(msg.Parts)
			if err != nil {
				return nil, nil, err
			}

			if text == "" {
				continue
			}

			if system == nil {
				system = &content{Parts: []part{{Text: text}}}
			} else {
				system.Parts = append(system.Parts, part{Text: text})
			}

			continue
		}

		if msg.Role == llm.RoleTool {
			// A tool result is a user-role functionResponse part. The name is
			// carried in ToolCallID because that is where the matching
			// ToolCall.ID was put when the call came back.
			contents = append(contents, content{
				Role: roleUser,
				Parts: []part{{
					FunctionResponse: &functionResponse{
						Name:     msg.ToolCallID,
						Response: toolResultPayload(msg.Parts),
					},
				}},
			})

			continue
		}

		converted := content{Role: roleUser}
		if msg.Role == llm.RoleAssistant {
			converted.Role = roleModel
		}

		for _, prt := range msg.Parts {
			switch p := prt.(type) {
			case llm.TextPart:
				if p.Text != "" {
					converted.Parts = append(converted.Parts, part{Text: p.Text})
				}
			case llm.FilePart:
				converted.Parts = append(converted.Parts, part{
					InlineData: &inlineData{MimeType: p.MimeType, Data: p.Data},
				})
			case llm.ThinkingPart:
				// Dropped: replaying it as text would feed the model its own
				// reasoning back as conversation.
			default:
				return nil, nil, fmt.Errorf(
					"content part %T is not supported by the gemini provider",
					prt,
				)
			}
		}

		for _, tc := range msg.ToolCalls {
			args := json.RawMessage(tc.Function.Arguments)
			if len(args) == 0 {
				args = json.RawMessage("{}")
			}

			converted.Parts = append(converted.Parts, part{
				FunctionCall: &functionCall{Name: tc.Function.Name, Args: args},
			})
		}

		if len(converted.Parts) == 0 {
			continue
		}

		contents = append(contents, converted)
	}

	return contents, system, nil
}

// toolResultPayload wraps a tool result so it is always a JSON object, which
// is what functionResponse.response must be. A plain string result is wrapped
// under "output" rather than sent bare.
func toolResultPayload(parts []llm.Part) json.RawMessage {
	text, err := joinText(parts)
	if err != nil || text == "" {
		return json.RawMessage(`{}`)
	}

	trimmed := strings.TrimSpace(text)
	if strings.HasPrefix(trimmed, "{") && json.Valid([]byte(trimmed)) {
		return json.RawMessage(trimmed)
	}

	wrapped, err := json.Marshal(map[string]string{"output": text})
	if err != nil {
		return json.RawMessage(`{}`)
	}

	return wrapped
}

func joinText(parts []llm.Part) (string, error) {
	var out strings.Builder

	for _, prt := range parts {
		switch p := prt.(type) {
		case llm.TextPart:
			out.WriteString(p.Text)
		case llm.ThinkingPart:
		default:
			return "", fmt.Errorf("content part %T is not supported by the gemini provider", prt)
		}
	}

	return out.String(), nil
}

func buildToolConfig(choice *llm.ToolChoice) *toolConfig {
	if choice == nil {
		return nil
	}

	switch choice.Type {
	case llm.ToolChoiceAuto:
		return &toolConfig{FunctionCallingConfig: functionCallingConfig{Mode: "AUTO"}}
	case llm.ToolChoiceNone:
		return &toolConfig{FunctionCallingConfig: functionCallingConfig{Mode: "NONE"}}
	case llm.ToolChoiceRequired:
		return &toolConfig{FunctionCallingConfig: functionCallingConfig{Mode: "ANY"}}
	case llm.ToolChoiceFunction:
		return &toolConfig{FunctionCallingConfig: functionCallingConfig{
			Mode:                 "ANY",
			AllowedFunctionNames: []string{choice.Function},
		}}
	}

	return nil
}

func applyResponseFormat(cfg *generationConfig, format *llm.ResponseFormat) error {
	if format == nil {
		return nil
	}

	switch format.Type {
	case llm.ResponseFormatText:
		return nil
	case llm.ResponseFormatJSONObject:
		cfg.ResponseMimeType = "application/json"

		return nil
	case llm.ResponseFormatJSONSchema:
		if format.JSONSchema == nil {
			return errors.New("response format json_schema requires a schema")
		}

		cfg.ResponseMimeType = "application/json"
		cfg.ResponseSchema = format.JSONSchema.Schema

		return nil
	}

	return fmt.Errorf("unsupported response format %q", format.Type)
}

func mapFinishReason(reason string, hasToolCalls bool) llm.FinishReason {
	// Gemini reports STOP even when the turn ends in a function call, so the
	// presence of a call is what distinguishes the two.
	if hasToolCalls {
		return llm.FinishReasonToolCalls
	}

	switch reason {
	case "MAX_TOKENS":
		return llm.FinishReasonLength
	case "SAFETY", "RECITATION", "BLOCKLIST", "PROHIBITED_CONTENT", "SPII":
		return llm.FinishReasonContentFilter
	default:
		return llm.FinishReasonStop
	}
}

type stream struct {
	body     io.ReadCloser
	scanner  *bufio.Scanner
	fallback string
	current  llm.ChatCompletionStreamEvent
	err      error
	done     bool
	// toolIndex assigns stable indices across chunks. Gemini emits whole
	// function calls rather than argument fragments, so each one is a new
	// index rather than a continuation.
	toolIndex int
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
		if payload == "" || payload == "[DONE]" {
			continue
		}

		var chunk generateResponse
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			s.err = fmt.Errorf("cannot decode gemini stream chunk: %w", err)

			return false
		}

		if chunk.Error != nil {
			s.err = fmt.Errorf("gemini stream failed: %s", chunk.Error.Message)

			return false
		}

		model := chunk.ModelVersion
		if model == "" {
			model = s.fallback
		}

		event := llm.ChatCompletionStreamEvent{Model: model}

		if chunk.UsageMetadata != nil {
			event.Usage = &llm.Usage{
				InputTokens:  chunk.UsageMetadata.PromptTokenCount,
				OutputTokens: chunk.UsageMetadata.CandidatesTokenCount,
			}
		}

		var sawToolCall bool

		if len(chunk.Candidates) > 0 {
			cand := chunk.Candidates[0]

			var text strings.Builder
			for _, prt := range cand.Content.Parts {
				switch {
				case prt.FunctionCall != nil:
					args := string(prt.FunctionCall.Args)
					if args == "" {
						args = "{}"
					}

					event.Delta.ToolCalls = append(event.Delta.ToolCalls, llm.ToolCallDelta{
						Index:     s.toolIndex,
						ID:        prt.FunctionCall.Name,
						Name:      prt.FunctionCall.Name,
						Arguments: args,
					})
					s.toolIndex++
					sawToolCall = true
				case prt.Text != "":
					text.WriteString(prt.Text)
				}
			}

			event.Delta.Content = text.String()

			if cand.FinishReason != "" {
				reason := mapFinishReason(cand.FinishReason, sawToolCall || s.toolIndex > 0)
				event.FinishReason = &reason
			}
		}

		s.current = event

		return true
	}

	if err := s.scanner.Err(); err != nil {
		s.err = fmt.Errorf("cannot read gemini stream: %w", err)

		return false
	}

	s.done = true

	return false
}

func (s *stream) Event() llm.ChatCompletionStreamEvent { return s.current }

func (s *stream) Err() error { return s.err }

func (s *stream) Close() error { return s.body.Close() }
