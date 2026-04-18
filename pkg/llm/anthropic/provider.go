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

package anthropic

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
	"go.probo.inc/probo/pkg/llm"
)

type (
	Provider struct {
		client *anthropic.Client
	}

	Option func(*config)

	config struct {
		httpClient     *http.Client
		baseURL        string
		requestTimeout time.Duration
		maxRetries     *int
	}
)

func WithHTTPClient(c *http.Client) Option {
	return func(cfg *config) { cfg.httpClient = c }
}

func WithBaseURL(url string) Option {
	return func(cfg *config) { cfg.baseURL = url }
}

func WithRequestTimeout(d time.Duration) Option {
	return func(cfg *config) { cfg.requestTimeout = d }
}

func WithMaxRetries(n int) Option {
	return func(cfg *config) { cfg.maxRetries = &n }
}

func NewProvider(apiKey string, opts ...Option) *Provider {
	var cfg config
	for _, o := range opts {
		o(&cfg)
	}

	reqOpts := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}

	if cfg.httpClient != nil {
		reqOpts = append(reqOpts, option.WithHTTPClient(cfg.httpClient))
	}
	if cfg.baseURL != "" {
		reqOpts = append(reqOpts, option.WithBaseURL(cfg.baseURL))
	}
	if cfg.requestTimeout > 0 {
		reqOpts = append(reqOpts, option.WithRequestTimeout(cfg.requestTimeout))
	}
	if cfg.maxRetries != nil {
		reqOpts = append(reqOpts, option.WithMaxRetries(*cfg.maxRetries))
	}

	client := anthropic.NewClient(reqOpts...)
	return &Provider{client: &client}
}

func (p *Provider) ChatCompletion(ctx context.Context, req *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
	params, err := buildParams(req)
	if err != nil {
		return nil, err
	}

	msg, err := p.client.Messages.New(ctx, params)
	if err != nil {
		return nil, mapError(err)
	}

	return mapResponse(msg), nil
}

func (p *Provider) ChatCompletionStream(ctx context.Context, req *llm.ChatCompletionRequest) (llm.ChatCompletionStream, error) {
	params, err := buildParams(req)
	if err != nil {
		return nil, err
	}

	stream := p.client.Messages.NewStreaming(ctx, params)
	return &anthropicStream{stream: stream}, nil
}

func buildParams(req *llm.ChatCompletionRequest) (anthropic.MessageNewParams, error) {
	if req.MaxTokens == nil {
		return anthropic.MessageNewParams{}, &llm.ErrContextLength{
			Err: fmt.Errorf("MaxTokens is required for Anthropic"),
		}
	}

	system, messages := extractSystem(req.Messages)

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(req.Model),
		MaxTokens: int64(*req.MaxTokens),
		Messages:  buildMessages(messages),
	}

	if len(system) > 0 {
		blocks := make([]anthropic.TextBlockParam, len(system))
		for i, s := range system {
			blocks[i] = anthropic.TextBlockParam{Text: s}
		}
		params.System = blocks
	}

	if req.Temperature != nil {
		params.Temperature = param.NewOpt(*req.Temperature)
	}
	if req.TopP != nil {
		params.TopP = param.NewOpt(*req.TopP)
	}
	if len(req.StopSequences) > 0 {
		params.StopSequences = req.StopSequences
	}
	if len(req.Tools) > 0 {
		params.Tools = buildTools(req.Tools)
	}
	if req.ToolChoice != nil {
		params.ToolChoice = buildToolChoice(req.ToolChoice)
	}
	if req.Thinking != nil && req.Thinking.Enabled {
		params.Thinking = anthropic.ThinkingConfigParamOfEnabled(int64(req.Thinking.BudgetTokens))
	}
	if req.ResponseFormat != nil {
		switch req.ResponseFormat.Type {
		case llm.ResponseFormatJSONSchema:
			if req.ResponseFormat.JSONSchema == nil {
				return anthropic.MessageNewParams{}, fmt.Errorf("cannot apply JSON schema output format: schema is nil")
			}
			var schema map[string]any
			if err := json.Unmarshal(req.ResponseFormat.JSONSchema.Schema, &schema); err != nil {
				return anthropic.MessageNewParams{}, fmt.Errorf("cannot unmarshal JSON schema for output format: %w", err)
			}
			params.OutputConfig = anthropic.OutputConfigParam{
				Format: anthropic.JSONOutputFormatParam{Schema: schema},
			}
		case llm.ResponseFormatJSONObject:
			return anthropic.MessageNewParams{}, fmt.Errorf("anthropic does not support json_object response format without a schema; use json_schema instead")
		case llm.ResponseFormatText:
			// default behaviour, nothing to set
		}
	}

	return params, nil
}

func extractSystem(messages []llm.Message) (system []string, rest []llm.Message) {
	for _, msg := range messages {
		if msg.Role == llm.RoleSystem {
			system = append(system, msg.Text())
		} else {
			rest = append(rest, msg)
		}
	}
	return
}

func buildMessages(messages []llm.Message) []anthropic.MessageParam {
	out := make([]anthropic.MessageParam, 0, len(messages))

	for _, msg := range messages {
		switch msg.Role {
		case llm.RoleUser:
			blocks := make([]anthropic.ContentBlockParamUnion, 0, len(msg.Parts))
			for _, p := range msg.Parts {
				switch p := p.(type) {
				case llm.TextPart:
					blocks = append(blocks, anthropic.NewTextBlock(p.Text))
				case llm.ImagePart:
					blocks = append(
						blocks,
						anthropic.NewImageBlock(
							anthropic.URLImageSourceParam{
								URL: p.URL,
							},
						),
					)
				case llm.FilePart:
					blocks = append(blocks, buildFilePart(p))
				}
			}
			out = append(out, anthropic.NewUserMessage(blocks...))
		case llm.RoleAssistant:
			var blocks []anthropic.ContentBlockParamUnion
			for _, p := range msg.Parts {
				switch part := p.(type) {
				case llm.ThinkingPart:
					blocks = append(blocks, anthropic.NewThinkingBlock(part.Signature, part.Text))
				case llm.TextPart:
					if part.Text != "" {
						blocks = append(blocks, anthropic.NewTextBlock(part.Text))
					}
				}
			}
			for _, tc := range msg.ToolCalls {
				var input any
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &input); err != nil || input == nil {
					input = map[string]any{}
				}
				blocks = append(blocks, anthropic.NewToolUseBlock(tc.ID, input, tc.Function.Name))
			}
			out = append(out, anthropic.NewAssistantMessage(blocks...))
		case llm.RoleTool:
			out = append(
				out,
				anthropic.NewUserMessage(
					anthropic.NewToolResultBlock(
						msg.ToolCallID,
						msg.Text(),
						false,
					),
				),
			)
		}
	}

	return out
}

func buildTools(tools []llm.Tool) []anthropic.ToolUnionParam {
	out := make([]anthropic.ToolUnionParam, len(tools))
	for i, t := range tools {
		tool := anthropic.ToolParam{
			Name:        t.Name,
			Description: param.NewOpt(t.Description),
		}

		if t.Parameters != nil {
			var schema map[string]any
			if err := json.Unmarshal(t.Parameters, &schema); err == nil {
				props := schema["properties"]
				required, _ := schema["required"].([]any)
				reqStrings := make([]string, 0, len(required))
				for _, r := range required {
					if s, ok := r.(string); ok {
						reqStrings = append(reqStrings, s)
					}
				}

				extra := make(map[string]any)
				for k, v := range schema {
					switch k {
					case "type", "properties", "required":
					default:
						extra[k] = v
					}
				}

				inputSchema := anthropic.ToolInputSchemaParam{
					Properties: props,
					Required:   reqStrings,
				}
				if len(extra) > 0 {
					inputSchema.ExtraFields = extra
				}
				tool.InputSchema = inputSchema
			}
		}

		out[i] = anthropic.ToolUnionParam{OfTool: &tool}
	}
	return out
}

func buildToolChoice(tc *llm.ToolChoice) anthropic.ToolChoiceUnionParam {
	switch tc.Type {
	case llm.ToolChoiceAuto:
		return anthropic.ToolChoiceUnionParam{OfAuto: &anthropic.ToolChoiceAutoParam{}}
	case llm.ToolChoiceNone:
		return anthropic.ToolChoiceUnionParam{OfNone: &anthropic.ToolChoiceNoneParam{}}
	case llm.ToolChoiceRequired:
		return anthropic.ToolChoiceUnionParam{OfAny: &anthropic.ToolChoiceAnyParam{}}
	case llm.ToolChoiceFunction:
		return anthropic.ToolChoiceUnionParam{OfTool: &anthropic.ToolChoiceToolParam{Name: tc.Function}}
	default:
		return anthropic.ToolChoiceUnionParam{}
	}
}

func mapResponse(msg *anthropic.Message) *llm.ChatCompletionResponse {
	resp := &llm.ChatCompletionResponse{
		Model:        string(msg.Model),
		FinishReason: mapStopReason(msg.StopReason),
		Usage: llm.Usage{
			InputTokens:  int(msg.Usage.InputTokens),
			OutputTokens: int(msg.Usage.OutputTokens),
		},
		Message: llm.Message{
			Role: llm.RoleAssistant,
		},
	}

	for _, block := range msg.Content {
		switch block.Type {
		case "thinking":
			tb := block.AsThinking()
			resp.Message.Parts = append(resp.Message.Parts, llm.ThinkingPart{
				Text:      tb.Thinking,
				Signature: tb.Signature,
			})
		case "text":
			resp.Message.Parts = append(resp.Message.Parts, llm.TextPart{Text: block.Text})
		case "tool_use":
			tu := block.AsToolUse()
			resp.Message.ToolCalls = append(resp.Message.ToolCalls, llm.ToolCall{
				ID: tu.ID,
				Function: llm.FunctionCall{
					Name:      tu.Name,
					Arguments: string(tu.Input),
				},
			})
		}
	}

	return resp
}

func mapStopReason(reason anthropic.StopReason) llm.FinishReason {
	switch reason {
	case anthropic.StopReasonEndTurn, anthropic.StopReasonStopSequence:
		return llm.FinishReasonStop
	case anthropic.StopReasonMaxTokens:
		return llm.FinishReasonLength
	case anthropic.StopReasonToolUse:
		return llm.FinishReasonToolCalls
	default:
		return llm.FinishReasonStop
	}
}

func mapError(err error) error {
	// The Anthropic SDK refuses non-streaming requests client-side when
	// the expected response time exceeds 10 minutes (large max_tokens or
	// model-specific non-streaming token limits). It returns a plain
	// fmt.Errorf, not an *anthropic.Error, so we must match on the
	// message before attempting the type assertion.
	if err != nil && strings.Contains(err.Error(), "streaming is required") {
		return &llm.ErrStreamingRequired{Err: err}
	}

	var apiErr *anthropic.Error
	if !errors.As(err, &apiErr) {
		return err
	}

	switch apiErr.StatusCode {
	case http.StatusTooManyRequests:
		retryAfter := parseRetryAfter(apiErr.Response)
		return &llm.ErrRateLimit{RetryAfter: retryAfter, Err: err}
	case http.StatusUnauthorized:
		return &llm.ErrAuthentication{Err: err}
	default:
		return err
	}
}

func parseRetryAfter(resp *http.Response) time.Duration {
	if resp == nil {
		return 0
	}
	h := resp.Header.Get("Retry-After")
	if h == "" {
		return 0
	}
	if secs, err := strconv.Atoi(h); err == nil {
		return time.Duration(secs) * time.Second
	}
	return 0
}

// anthropicStream adapts an Anthropic SSE stream to our ChatCompletionStream interface.
type anthropicStream struct {
	stream  *ssestream.Stream[anthropic.MessageStreamEventUnion]
	current llm.ChatCompletionStreamEvent
	// Track tool call indices for mapping content_block_start events.
	toolCallIndex     int
	inToolUse         bool
	thinkingSignature string
}

func (s *anthropicStream) Next() bool {
	for s.stream.Next() {
		event := s.stream.Current()
		mapped, ok := s.mapStreamEvent(&event)
		if ok {
			s.current = mapped
			return true
		}
	}
	return false
}

func (s *anthropicStream) Event() llm.ChatCompletionStreamEvent {
	return s.current
}

func (s *anthropicStream) Err() error {
	err := s.stream.Err()
	if err != nil {
		return mapError(err)
	}
	return nil
}

func (s *anthropicStream) Close() error {
	return s.stream.Close()
}

func (s *anthropicStream) mapStreamEvent(event *anthropic.MessageStreamEventUnion) (llm.ChatCompletionStreamEvent, bool) {
	switch event.Type {
	case "content_block_start":
		cb := event.ContentBlock
		switch cb.Type {
		case "tool_use":
			s.inToolUse = true
			tu := cb.AsToolUse()
			return llm.ChatCompletionStreamEvent{
				Delta: llm.MessageDelta{
					ToolCalls: []llm.ToolCallDelta{{
						Index: s.toolCallIndex,
						ID:    tu.ID,
						Name:  tu.Name,
					}},
				},
			}, true
		case "thinking":
			return llm.ChatCompletionStreamEvent{}, false
		}
		return llm.ChatCompletionStreamEvent{}, false

	case "content_block_delta":
		delta := event.Delta
		switch delta.Type {
		case "text_delta":
			return llm.ChatCompletionStreamEvent{
				Delta: llm.MessageDelta{Content: delta.Text},
			}, true
		case "thinking_delta":
			return llm.ChatCompletionStreamEvent{
				Delta: llm.MessageDelta{Thinking: delta.Thinking},
			}, true
		case "signature_delta":
			s.thinkingSignature = delta.Signature
			return llm.ChatCompletionStreamEvent{
				Delta: llm.MessageDelta{ThinkingSignature: delta.Signature},
			}, true
		case "input_json_delta":
			return llm.ChatCompletionStreamEvent{
				Delta: llm.MessageDelta{
					ToolCalls: []llm.ToolCallDelta{{
						Index:     s.toolCallIndex,
						Arguments: delta.PartialJSON,
					}},
				},
			}, true
		}
		return llm.ChatCompletionStreamEvent{}, false

	case "content_block_stop":
		if s.inToolUse {
			s.toolCallIndex++
			s.inToolUse = false
		}
		return llm.ChatCompletionStreamEvent{}, false

	case "message_delta":
		fr := mapStopReason(anthropic.StopReason(event.Delta.StopReason))
		evt := llm.ChatCompletionStreamEvent{
			FinishReason: &fr,
		}
		if event.Usage.OutputTokens > 0 || event.Usage.InputTokens > 0 {
			evt.Usage = &llm.Usage{
				InputTokens:  int(event.Usage.InputTokens),
				OutputTokens: int(event.Usage.OutputTokens),
			}
		}
		return evt, true

	case "message_start":
		evt := llm.ChatCompletionStreamEvent{
			Model: string(event.Message.Model),
		}
		if event.Message.Usage.InputTokens > 0 || event.Message.Usage.OutputTokens > 0 {
			evt.Usage = &llm.Usage{
				InputTokens:  int(event.Message.Usage.InputTokens),
				OutputTokens: int(event.Message.Usage.OutputTokens),
			}
		}
		return evt, true

	default:
		return llm.ChatCompletionStreamEvent{}, false
	}
}

func buildFilePart(p llm.FilePart) anthropic.ContentBlockParamUnion {
	switch {
	case strings.HasPrefix(p.MimeType, "image/"):
		return anthropic.NewImageBlockBase64(p.MimeType, p.Data)
	case p.MimeType == "application/pdf":
		return anthropic.NewDocumentBlock(anthropic.Base64PDFSourceParam{
			Data: p.Data,
		})
	case strings.HasPrefix(p.MimeType, "text/"):
		decoded, err := base64.StdEncoding.DecodeString(p.Data)
		if err != nil {
			return anthropic.NewTextBlock(fmt.Sprintf("[file: %s, type: %s, error decoding content]", p.Filename, p.MimeType))
		}
		return anthropic.NewDocumentBlock(anthropic.PlainTextSourceParam{
			Data: string(decoded),
		})
	default:
		return anthropic.NewTextBlock(fmt.Sprintf("[file: %s, type: %s, unsupported format]", p.Filename, p.MimeType))
	}
}
