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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/packages/ssestream"
	"github.com/openai/openai-go/responses"
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
	params, err := buildParams(req)
	if err != nil {
		return nil, fmt.Errorf("cannot build OpenAI response parameters: %w", err)
	}

	response, err := p.client.Responses.New(ctx, params)
	if err != nil {
		return nil, mapError(err)
	}

	if response.Status == responses.ResponseStatusFailed {
		return nil, fmt.Errorf("cannot complete OpenAI response: %s", response.Error.Message)
	}

	return mapResponse(response), nil
}

func (p *Provider) ChatCompletionStream(ctx context.Context, req *llm.ChatCompletionRequest) (llm.ChatCompletionStream, error) {
	params, err := buildParams(req)
	if err != nil {
		return nil, fmt.Errorf("cannot build OpenAI response parameters: %w", err)
	}

	stream := p.client.Responses.NewStreaming(ctx, params)

	return &openaiStream{
		stream:      stream,
		toolIndexes: make(map[int64]int),
	}, nil
}

func buildParams(req *llm.ChatCompletionRequest) (responses.ResponseNewParams, error) {
	if len(req.StopSequences) > 0 {
		return responses.ResponseNewParams{}, errors.New("stop sequences are not supported by OpenAI Responses API")
	}

	params := responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: buildInput(req.Messages),
		},
		Model: shared.ResponsesModel(req.Model),
		Store: param.NewOpt(false),
	}

	if req.MaxTokens != nil {
		params.MaxOutputTokens = param.NewOpt(int64(*req.MaxTokens))
	}

	if req.Temperature != nil {
		params.Temperature = param.NewOpt(*req.Temperature)
	}

	if req.TopP != nil {
		params.TopP = param.NewOpt(*req.TopP)
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
		params.Text.Format = buildResponseFormat(req.ResponseFormat)
	}

	if req.Thinking != nil && req.Thinking.Enabled && isReasoningModel(req.Model) {
		switch {
		case req.Thinking.BudgetTokens <= 1024:
			params.Reasoning.Effort = shared.ReasoningEffortLow
		case req.Thinking.BudgetTokens <= 8192:
			params.Reasoning.Effort = shared.ReasoningEffortMedium
		default:
			params.Reasoning.Effort = shared.ReasoningEffortHigh
		}
	}

	return params, nil
}

func buildInput(messages []llm.Message) responses.ResponseInputParam {
	input := make(responses.ResponseInputParam, 0, len(messages))

	for _, msg := range messages {
		switch msg.Role {
		case llm.RoleSystem:
			input = append(
				input,
				responses.ResponseInputItemParamOfMessage(
					msg.Text(),
					responses.EasyInputMessageRoleSystem,
				),
			)
		case llm.RoleUser:
			input = append(
				input,
				responses.ResponseInputItemParamOfMessage(
					buildUserContent(msg.Parts),
					responses.EasyInputMessageRoleUser,
				),
			)
		case llm.RoleAssistant:
			if text := msg.Text(); text != "" {
				input = append(
					input,
					responses.ResponseInputItemParamOfMessage(
						text,
						responses.EasyInputMessageRoleAssistant,
					),
				)
			}

			for _, tc := range msg.ToolCalls {
				input = append(
					input,
					responses.ResponseInputItemParamOfFunctionCall(
						tc.Function.Arguments,
						tc.ID,
						tc.Function.Name,
					),
				)
			}
		case llm.RoleTool:
			input = append(
				input,
				responses.ResponseInputItemParamOfFunctionCallOutput(
					msg.ToolCallID,
					msg.Text(),
				),
			)
		}
	}

	return input
}

func buildUserContent(parts []llm.Part) responses.ResponseInputMessageContentListParam {
	content := make(responses.ResponseInputMessageContentListParam, 0, len(parts))

	for _, part := range parts {
		switch part := part.(type) {
		case llm.TextPart:
			content = append(content, responses.ResponseInputContentParamOfInputText(part.Text))
		case llm.ImagePart:
			image := responses.ResponseInputImageParam{
				Detail:   responses.ResponseInputImageDetailAuto,
				ImageURL: param.NewOpt(part.URL),
			}
			content = append(content, responses.ResponseInputContentUnionParam{OfInputImage: &image})
		case llm.FilePart:
			content = append(content, buildFilePart(part))
		}
	}

	return content
}

func buildTools(tools []llm.Tool) []responses.ToolUnionParam {
	out := make([]responses.ToolUnionParam, len(tools))
	for i, t := range tools {
		var parameters map[string]any
		if t.Parameters != nil {
			_ = json.Unmarshal(t.Parameters, &parameters)
		}

		tool := responses.ToolParamOfFunction(t.Name, parameters, true)
		if t.Description != "" {
			tool.OfFunction.Description = param.NewOpt(t.Description)
		}

		out[i] = tool
	}

	return out
}

func buildToolChoice(tc *llm.ToolChoice) responses.ResponseNewParamsToolChoiceUnion {
	switch tc.Type {
	case llm.ToolChoiceAuto:
		return responses.ResponseNewParamsToolChoiceUnion{
			OfToolChoiceMode: param.NewOpt(responses.ToolChoiceOptionsAuto),
		}
	case llm.ToolChoiceNone:
		return responses.ResponseNewParamsToolChoiceUnion{
			OfToolChoiceMode: param.NewOpt(responses.ToolChoiceOptionsNone),
		}
	case llm.ToolChoiceRequired:
		return responses.ResponseNewParamsToolChoiceUnion{
			OfToolChoiceMode: param.NewOpt(responses.ToolChoiceOptionsRequired),
		}
	case llm.ToolChoiceFunction:
		return responses.ResponseNewParamsToolChoiceUnion{
			OfFunctionTool: &responses.ToolChoiceFunctionParam{Name: tc.Function},
		}
	default:
		return responses.ResponseNewParamsToolChoiceUnion{}
	}
}

func buildResponseFormat(rf *llm.ResponseFormat) responses.ResponseFormatTextConfigUnionParam {
	switch rf.Type {
	case llm.ResponseFormatText:
		return responses.ResponseFormatTextConfigUnionParam{
			OfText: &shared.ResponseFormatTextParam{},
		}
	case llm.ResponseFormatJSONObject:
		return responses.ResponseFormatTextConfigUnionParam{
			OfJSONObject: &shared.ResponseFormatJSONObjectParam{},
		}
	case llm.ResponseFormatJSONSchema:
		if rf.JSONSchema != nil {
			var schema map[string]any

			_ = json.Unmarshal(rf.JSONSchema.Schema, &schema)

			format := responses.ResponseFormatTextJSONSchemaConfigParam{
				Name:   rf.JSONSchema.Name,
				Strict: param.NewOpt(rf.JSONSchema.Strict),
				Schema: schema,
			}
			if rf.JSONSchema.Description != "" {
				format.Description = param.NewOpt(rf.JSONSchema.Description)
			}

			return responses.ResponseFormatTextConfigUnionParam{
				OfJSONSchema: &format,
			}
		}

		return responses.ResponseFormatTextConfigUnionParam{}
	default:
		return responses.ResponseFormatTextConfigUnionParam{}
	}
}

func mapResponse(response *responses.Response) *llm.ChatCompletionResponse {
	resp := &llm.ChatCompletionResponse{
		Model: string(response.Model),
		Message: llm.Message{
			Role:  llm.RoleAssistant,
			Parts: []llm.Part{llm.TextPart{Text: response.OutputText()}},
		},
		Usage: llm.Usage{
			InputTokens:  int(response.Usage.InputTokens),
			OutputTokens: int(response.Usage.OutputTokens),
		},
		FinishReason: mapFinishReason(response),
	}

	for _, output := range response.Output {
		if output.Type == "function_call" {
			resp.Message.ToolCalls = append(
				resp.Message.ToolCalls,
				llm.ToolCall{
					ID: output.CallID,
					Function: llm.FunctionCall{
						Name:      output.Name,
						Arguments: output.Arguments,
					},
				},
			)
		}
	}

	return resp
}

func mapFinishReason(response *responses.Response) llm.FinishReason {
	if response.IncompleteDetails.Reason == "max_output_tokens" {
		return llm.FinishReasonLength
	}

	if response.IncompleteDetails.Reason == "content_filter" {
		return llm.FinishReasonContentFilter
	}

	for _, output := range response.Output {
		if output.Type == "function_call" {
			return llm.FinishReasonToolCalls
		}
	}

	return llm.FinishReasonStop
}

func buildFilePart(part llm.FilePart) responses.ResponseInputContentUnionParam {
	dataURL := (&url.URL{
		Scheme: "data",
		Opaque: part.MimeType + ";base64," + part.Data,
	}).String()

	if strings.HasPrefix(part.MimeType, "image/") {
		image := responses.ResponseInputImageParam{
			Detail:   responses.ResponseInputImageDetailAuto,
			ImageURL: param.NewOpt(dataURL),
		}

		return responses.ResponseInputContentUnionParam{OfInputImage: &image}
	}

	file := responses.ResponseInputFileParam{
		FileData: param.NewOpt(dataURL),
		Filename: param.NewOpt(part.Filename),
	}

	return responses.ResponseInputContentUnionParam{OfInputFile: &file}
}

func isReasoningModel(model string) bool {
	for _, prefix := range []string{"o1", "o3", "o4", "gpt-5"} {
		if model == prefix || strings.HasPrefix(model, prefix+"-") || strings.HasPrefix(model, prefix+".") {
			return true
		}
	}

	return false
}

func mapError(err error) error {
	apiErr, ok := errors.AsType[*openai.Error](err)
	if !ok {
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

type openaiStream struct {
	stream      *ssestream.Stream[responses.ResponseStreamEventUnion]
	current     llm.ChatCompletionStreamEvent
	toolIndexes map[int64]int
	err         error
}

func (s *openaiStream) Next() bool {
	for s.stream.Next() {
		event, ok := s.mapEvent(s.stream.Current())
		if !ok {
			continue
		}

		s.current = event

		return true
	}

	return false
}

func (s *openaiStream) Event() llm.ChatCompletionStreamEvent {
	return s.current
}

func (s *openaiStream) Err() error {
	if s.err != nil {
		return s.err
	}

	err := s.stream.Err()
	if err != nil {
		return mapError(err)
	}

	return nil
}

func (s *openaiStream) Close() error {
	return s.stream.Close()
}

func (s *openaiStream) mapEvent(raw responses.ResponseStreamEventUnion) (llm.ChatCompletionStreamEvent, bool) {
	switch event := raw.AsAny().(type) {
	case responses.ResponseCreatedEvent:
		return llm.ChatCompletionStreamEvent{
			Model: string(event.Response.Model),
		}, true
	case responses.ResponseTextDeltaEvent:
		return llm.ChatCompletionStreamEvent{
			Delta: llm.MessageDelta{Content: event.Delta},
		}, true
	case responses.ResponseOutputItemAddedEvent:
		if event.Item.Type != "function_call" {
			return llm.ChatCompletionStreamEvent{}, false
		}

		toolIndex := len(s.toolIndexes)
		s.toolIndexes[event.OutputIndex] = toolIndex

		return llm.ChatCompletionStreamEvent{
			Delta: llm.MessageDelta{
				ToolCalls: []llm.ToolCallDelta{
					{
						Index: toolIndex,
						ID:    event.Item.CallID,
						Name:  event.Item.Name,
					},
				},
			},
		}, true
	case responses.ResponseFunctionCallArgumentsDeltaEvent:
		toolIndex, ok := s.toolIndexes[event.OutputIndex]
		if !ok {
			return llm.ChatCompletionStreamEvent{}, false
		}

		return llm.ChatCompletionStreamEvent{
			Delta: llm.MessageDelta{
				ToolCalls: []llm.ToolCallDelta{
					{
						Index:     toolIndex,
						Arguments: event.Delta,
					},
				},
			},
		}, true
	case responses.ResponseCompletedEvent:
		return finalStreamEvent(&event.Response), true
	case responses.ResponseIncompleteEvent:
		return finalStreamEvent(&event.Response), true
	case responses.ResponseErrorEvent:
		s.err = fmt.Errorf("cannot stream OpenAI response: %s", event.Message)
		return llm.ChatCompletionStreamEvent{}, false
	case responses.ResponseFailedEvent:
		s.err = fmt.Errorf("cannot stream OpenAI response: %s", event.Response.Error.Message)
		return llm.ChatCompletionStreamEvent{}, false
	default:
		return llm.ChatCompletionStreamEvent{}, false
	}
}

func finalStreamEvent(response *responses.Response) llm.ChatCompletionStreamEvent {
	finishReason := mapFinishReason(response)

	return llm.ChatCompletionStreamEvent{
		Model: string(response.Model),
		Usage: &llm.Usage{
			InputTokens:  int(response.Usage.InputTokens),
			OutputTokens: int(response.Usage.OutputTokens),
		},
		FinishReason: &finishReason,
	}
}
