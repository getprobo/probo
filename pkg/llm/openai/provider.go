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

package openai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/packages/ssestream"
	"github.com/openai/openai-go/shared"
	"go.probo.inc/probo/pkg/llm"
)

type (
	Provider struct {
		client *openai.Client
	}

	Option func(*config)

	config struct {
		httpClient     *http.Client
		baseURL        string
		organization   string
		project        string
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

func WithOrganization(org string) Option {
	return func(cfg *config) { cfg.organization = org }
}

func WithProject(project string) Option {
	return func(cfg *config) { cfg.project = project }
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
	if cfg.organization != "" {
		reqOpts = append(reqOpts, option.WithOrganization(cfg.organization))
	}
	if cfg.project != "" {
		reqOpts = append(reqOpts, option.WithProject(cfg.project))
	}
	if cfg.requestTimeout > 0 {
		reqOpts = append(reqOpts, option.WithRequestTimeout(cfg.requestTimeout))
	}
	if cfg.maxRetries != nil {
		reqOpts = append(reqOpts, option.WithMaxRetries(*cfg.maxRetries))
	}

	client := openai.NewClient(reqOpts...)
	return &Provider{client: &client}
}

func (p *Provider) ChatCompletion(ctx context.Context, req *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
	params := buildParams(req)

	completion, err := p.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, mapError(err)
	}

	return mapResponse(completion), nil
}

func (p *Provider) ChatCompletionStream(ctx context.Context, req *llm.ChatCompletionRequest) (llm.ChatCompletionStream, error) {
	params := buildParams(req)
	params.StreamOptions = openai.ChatCompletionStreamOptionsParam{
		IncludeUsage: param.NewOpt(true),
	}

	stream := p.client.Chat.Completions.NewStreaming(ctx, params)
	return &openaiStream{stream: stream}, nil
}

func buildParams(req *llm.ChatCompletionRequest) openai.ChatCompletionNewParams {
	params := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(req.Model),
		Messages: buildMessages(req.Messages),
	}

	if req.MaxTokens != nil {
		params.MaxCompletionTokens = param.NewOpt(int64(*req.MaxTokens))
	}
	if req.Temperature != nil {
		params.Temperature = param.NewOpt(*req.Temperature)
	}
	if req.TopP != nil {
		params.TopP = param.NewOpt(*req.TopP)
	}
	if req.FrequencyPenalty != nil {
		params.FrequencyPenalty = param.NewOpt(*req.FrequencyPenalty)
	}
	if req.PresencePenalty != nil {
		params.PresencePenalty = param.NewOpt(*req.PresencePenalty)
	}
	if len(req.StopSequences) > 0 {
		params.Stop = openai.ChatCompletionNewParamsStopUnion{
			OfStringArray: req.StopSequences,
		}
	}
	if len(req.Tools) > 0 {
		params.Tools = buildTools(req.Tools)
	}
	if req.ToolChoice != nil {
		params.ToolChoice = buildToolChoice(req.ToolChoice)
	}
	if req.ParallelToolCalls != nil {
		params.ParallelToolCalls = param.NewOpt(*req.ParallelToolCalls)
	}
	if req.ResponseFormat != nil {
		params.ResponseFormat = buildResponseFormat(req.ResponseFormat)
	}

	return params
}

func buildMessages(messages []llm.Message) []openai.ChatCompletionMessageParamUnion {
	out := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages))

	for _, msg := range messages {
		switch msg.Role {
		case llm.RoleSystem:
			out = append(out, openai.SystemMessage(msg.Text()))
		case llm.RoleUser:
			parts := make([]openai.ChatCompletionContentPartUnionParam, 0, len(msg.Parts))
			for _, p := range msg.Parts {
				switch p := p.(type) {
				case llm.TextPart:
					parts = append(parts, openai.TextContentPart(p.Text))
				case llm.ImagePart:
					parts = append(parts, openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
						URL: p.URL,
					}))
				}
			}
			out = append(out, openai.UserMessage(parts))
		case llm.RoleAssistant:
			m := openai.ChatCompletionAssistantMessageParam{
				Content: openai.ChatCompletionAssistantMessageParamContentUnion{
					OfString: param.NewOpt(msg.Text()),
				},
			}
			if len(msg.ToolCalls) > 0 {
				m.ToolCalls = make([]openai.ChatCompletionMessageToolCallParam, len(msg.ToolCalls))
				for i, tc := range msg.ToolCalls {
					m.ToolCalls[i] = openai.ChatCompletionMessageToolCallParam{
						ID: tc.ID,
						Function: openai.ChatCompletionMessageToolCallFunctionParam{
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
					}
				}
			}
			out = append(out, openai.ChatCompletionMessageParamUnion{OfAssistant: &m})
		case llm.RoleTool:
			out = append(out, openai.ToolMessage(msg.Text(), msg.ToolCallID))
		}
	}

	return out
}

func buildTools(tools []llm.Tool) []openai.ChatCompletionToolParam {
	out := make([]openai.ChatCompletionToolParam, len(tools))
	for i, t := range tools {
		fn := shared.FunctionDefinitionParam{
			Name:        t.Name,
			Description: param.NewOpt(t.Description),
			Strict:      param.NewOpt(true),
		}
		if t.Parameters != nil {
			var params shared.FunctionParameters
			if err := json.Unmarshal(t.Parameters, &params); err == nil {
				fn.Parameters = params
			}
		}
		out[i] = openai.ChatCompletionToolParam{Function: fn}
	}
	return out
}

func buildToolChoice(tc *llm.ToolChoice) openai.ChatCompletionToolChoiceOptionUnionParam {
	switch tc.Type {
	case llm.ToolChoiceAuto:
		return openai.ChatCompletionToolChoiceOptionUnionParam{
			OfAuto: param.NewOpt(string(openai.ChatCompletionToolChoiceOptionAutoAuto)),
		}
	case llm.ToolChoiceNone:
		return openai.ChatCompletionToolChoiceOptionUnionParam{
			OfAuto: param.NewOpt(string(openai.ChatCompletionToolChoiceOptionAutoNone)),
		}
	case llm.ToolChoiceRequired:
		return openai.ChatCompletionToolChoiceOptionUnionParam{
			OfAuto: param.NewOpt(string(openai.ChatCompletionToolChoiceOptionAutoRequired)),
		}
	case llm.ToolChoiceFunction:
		return openai.ChatCompletionToolChoiceOptionParamOfChatCompletionNamedToolChoice(
			openai.ChatCompletionNamedToolChoiceFunctionParam{Name: tc.Function},
		)
	default:
		return openai.ChatCompletionToolChoiceOptionUnionParam{}
	}
}

func buildResponseFormat(rf *llm.ResponseFormat) openai.ChatCompletionNewParamsResponseFormatUnion {
	switch rf.Type {
	case llm.ResponseFormatText:
		return openai.ChatCompletionNewParamsResponseFormatUnion{
			OfText: &shared.ResponseFormatTextParam{},
		}
	case llm.ResponseFormatJSONObject:
		return openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &shared.ResponseFormatJSONObjectParam{},
		}
	case llm.ResponseFormatJSONSchema:
		if rf.JSONSchema != nil {
			schema := shared.ResponseFormatJSONSchemaJSONSchemaParam{
				Name:   rf.JSONSchema.Name,
				Strict: param.NewOpt(rf.JSONSchema.Strict),
			}
			if rf.JSONSchema.Description != "" {
				schema.Description = param.NewOpt(rf.JSONSchema.Description)
			}
			if rf.JSONSchema.Schema != nil {
				schema.Schema = rf.JSONSchema.Schema
			}
			return openai.ChatCompletionNewParamsResponseFormatUnion{
				OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{JSONSchema: schema},
			}
		}
		return openai.ChatCompletionNewParamsResponseFormatUnion{}
	default:
		return openai.ChatCompletionNewParamsResponseFormatUnion{}
	}
}

func mapResponse(c *openai.ChatCompletion) *llm.ChatCompletionResponse {
	resp := &llm.ChatCompletionResponse{
		Model: c.Model,
		Usage: llm.Usage{
			InputTokens:  int(c.Usage.PromptTokens),
			OutputTokens: int(c.Usage.CompletionTokens),
		},
	}

	if len(c.Choices) > 0 {
		choice := c.Choices[0]
		resp.FinishReason = mapFinishReason(choice.FinishReason)
		resp.Message = llm.Message{
			Role:  llm.RoleAssistant,
			Parts: []llm.Part{llm.TextPart{Text: choice.Message.Content}},
		}
		if len(choice.Message.ToolCalls) > 0 {
			resp.Message.ToolCalls = make([]llm.ToolCall, len(choice.Message.ToolCalls))
			for i, tc := range choice.Message.ToolCalls {
				resp.Message.ToolCalls[i] = llm.ToolCall{
					ID: tc.ID,
					Function: llm.FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
		}
	}

	return resp
}

func mapFinishReason(reason string) llm.FinishReason {
	switch reason {
	case "stop":
		return llm.FinishReasonStop
	case "tool_calls":
		return llm.FinishReasonToolCalls
	case "length":
		return llm.FinishReasonLength
	case "content_filter":
		return llm.FinishReasonContentFilter
	default:
		return llm.FinishReasonStop
	}
}

func mapError(err error) error {
	var apiErr *openai.Error
	if !errors.As(err, &apiErr) {
		return err
	}

	switch apiErr.StatusCode {
	case http.StatusTooManyRequests:
		retryAfter := parseRetryAfter(apiErr.Response)
		return &llm.ErrRateLimit{RetryAfter: retryAfter, Err: err}
	case http.StatusUnauthorized:
		return &llm.ErrAuthentication{Err: err}
	case http.StatusBadRequest:
		if apiErr.Code == "context_length_exceeded" {
			return &llm.ErrContextLength{Err: err}
		}
		if apiErr.Code == "content_filter" {
			return &llm.ErrContentFilter{Err: err}
		}
		return err
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

// openaiStream adapts an OpenAI SSE stream to our ChatCompletionStream interface.
type openaiStream struct {
	stream  *ssestream.Stream[openai.ChatCompletionChunk]
	current llm.ChatCompletionStreamEvent
}

func (s *openaiStream) Next() bool {
	if !s.stream.Next() {
		return false
	}

	chunk := s.stream.Current()
	s.current = mapChunkToEvent(&chunk)
	return true
}

func (s *openaiStream) Event() llm.ChatCompletionStreamEvent {
	return s.current
}

func (s *openaiStream) Err() error {
	err := s.stream.Err()
	if err != nil {
		return mapError(err)
	}
	return nil
}

func (s *openaiStream) Close() error {
	return s.stream.Close()
}

func mapChunkToEvent(chunk *openai.ChatCompletionChunk) llm.ChatCompletionStreamEvent {
	event := llm.ChatCompletionStreamEvent{}

	if chunk.Usage.PromptTokens > 0 || chunk.Usage.CompletionTokens > 0 {
		usage := llm.Usage{
			InputTokens:  int(chunk.Usage.PromptTokens),
			OutputTokens: int(chunk.Usage.CompletionTokens),
		}
		event.Usage = &usage
	}

	if len(chunk.Choices) > 0 {
		choice := chunk.Choices[0]
		delta := choice.Delta

		event.Delta.Content = delta.Content

		if len(delta.ToolCalls) > 0 {
			event.Delta.ToolCalls = make([]llm.ToolCallDelta, len(delta.ToolCalls))
			for i, tc := range delta.ToolCalls {
				event.Delta.ToolCalls[i] = llm.ToolCallDelta{
					Index:     int(tc.Index),
					ID:        tc.ID,
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				}
			}
		}

		if choice.FinishReason != "" {
			fr := mapFinishReason(choice.FinishReason)
			event.FinishReason = &fr
		}
	}

	return event
}
