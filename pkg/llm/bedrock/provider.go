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

package bedrock

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"go.probo.inc/probo/pkg/llm"
)

type (
	Provider struct {
		client *bedrockruntime.Client
	}

	Option func(*bedrockruntime.Options)
)

// WithBaseEndpoint overrides the Bedrock service endpoint URL.
func WithBaseEndpoint(url string) Option {
	return func(o *bedrockruntime.Options) { o.BaseEndpoint = &url }
}

func NewProvider(cfg aws.Config, opts ...Option) *Provider {
	fns := make([]func(*bedrockruntime.Options), len(opts))
	for i, o := range opts {
		fns[i] = func(bo *bedrockruntime.Options) { o(bo) }
	}

	client := bedrockruntime.NewFromConfig(cfg, fns...)
	return &Provider{client: client}
}

func (p *Provider) ChatCompletion(ctx context.Context, req *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
	input := buildInput(req)

	output, err := p.client.Converse(ctx, input)
	if err != nil {
		return nil, mapError(err)
	}

	return mapResponse(output, req.Model), nil
}

func (p *Provider) ChatCompletionStream(ctx context.Context, req *llm.ChatCompletionRequest) (llm.ChatCompletionStream, error) {
	input := &bedrockruntime.ConverseStreamInput{
		ModelId:         aws.String(req.Model),
		Messages:        buildMessages(req.Messages),
		InferenceConfig: buildInferenceConfig(req),
	}

	system := buildSystem(req.Messages)
	if len(system) > 0 {
		input.System = system
	}

	if len(req.Tools) > 0 && (req.ToolChoice == nil || req.ToolChoice.Type != llm.ToolChoiceNone) {
		input.ToolConfig = buildToolConfig(req)
	}

	output, err := p.client.ConverseStream(ctx, input)
	if err != nil {
		return nil, mapError(err)
	}

	return newBedrockStream(output.GetStream()), nil
}

func buildInput(req *llm.ChatCompletionRequest) *bedrockruntime.ConverseInput {
	input := &bedrockruntime.ConverseInput{
		ModelId:         aws.String(req.Model),
		Messages:        buildMessages(req.Messages),
		InferenceConfig: buildInferenceConfig(req),
	}

	system := buildSystem(req.Messages)
	if len(system) > 0 {
		input.System = system
	}

	if len(req.Tools) > 0 && (req.ToolChoice == nil || req.ToolChoice.Type != llm.ToolChoiceNone) {
		input.ToolConfig = buildToolConfig(req)
	}

	return input
}

func buildInferenceConfig(req *llm.ChatCompletionRequest) *types.InferenceConfiguration {
	cfg := &types.InferenceConfiguration{}

	if req.MaxTokens != nil {
		v := int32(*req.MaxTokens)
		cfg.MaxTokens = &v
	}
	if req.Temperature != nil {
		v := float32(*req.Temperature)
		cfg.Temperature = &v
	}
	if req.TopP != nil {
		v := float32(*req.TopP)
		cfg.TopP = &v
	}
	if len(req.StopSequences) > 0 {
		cfg.StopSequences = req.StopSequences
	}

	return cfg
}

func buildSystem(messages []llm.Message) []types.SystemContentBlock {
	var system []types.SystemContentBlock
	for _, msg := range messages {
		if msg.Role == llm.RoleSystem {
			system = append(
				system,
				&types.SystemContentBlockMemberText{
					Value: msg.Text(),
				},
			)
		}
	}
	return system
}

func buildMessages(messages []llm.Message) []types.Message {
	var out []types.Message

	for _, msg := range messages {
		switch msg.Role {
		case llm.RoleSystem:
			continue
		case llm.RoleUser:
			var content []types.ContentBlock
			for _, p := range msg.Parts {
				if tp, ok := p.(llm.TextPart); ok {
					content = append(content, &types.ContentBlockMemberText{Value: tp.Text})
				}
			}
			out = append(
				out, types.Message{
					Role:    types.ConversationRoleUser,
					Content: content,
				},
			)

		case llm.RoleAssistant:
			var content []types.ContentBlock
			if text := msg.Text(); text != "" {
				content = append(content, &types.ContentBlockMemberText{Value: text})
			}
			for _, tc := range msg.ToolCalls {
				var input any
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &input)
				content = append(
					content,
					&types.ContentBlockMemberToolUse{
						Value: types.ToolUseBlock{
							ToolUseId: aws.String(tc.ID),
							Name:      aws.String(tc.Function.Name),
							Input:     document.NewLazyDocument(input),
						},
					},
				)
			}
			out = append(
				out, types.Message{
					Role:    types.ConversationRoleAssistant,
					Content: content,
				},
			)

		case llm.RoleTool:
			out = append(out, types.Message{
				Role: types.ConversationRoleUser,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberToolResult{
						Value: types.ToolResultBlock{
							ToolUseId: aws.String(msg.ToolCallID),
							Content: []types.ToolResultContentBlock{
								&types.ToolResultContentBlockMemberText{Value: msg.Text()},
							},
						},
					},
				},
			})
		}
	}

	return out
}

func buildToolConfig(req *llm.ChatCompletionRequest) *types.ToolConfiguration {
	config := &types.ToolConfiguration{}

	tools := make([]types.Tool, len(req.Tools))
	for i, t := range req.Tools {
		spec := types.ToolSpecification{
			Name:        aws.String(t.Name),
			Description: aws.String(t.Description),
		}
		if t.Parameters != nil {
			var schema any
			_ = json.Unmarshal(t.Parameters, &schema)
			spec.InputSchema = &types.ToolInputSchemaMemberJson{
				Value: document.NewLazyDocument(schema),
			}
		}
		tools[i] = &types.ToolMemberToolSpec{Value: spec}
	}
	config.Tools = tools

	if req.ToolChoice != nil {
		config.ToolChoice = buildToolChoice(req.ToolChoice)
	}

	return config
}

func buildToolChoice(tc *llm.ToolChoice) types.ToolChoice {
	switch tc.Type {
	case llm.ToolChoiceAuto:
		return &types.ToolChoiceMemberAuto{Value: types.AutoToolChoice{}}
	case llm.ToolChoiceRequired:
		return &types.ToolChoiceMemberAny{Value: types.AnyToolChoice{}}
	case llm.ToolChoiceFunction:
		return &types.ToolChoiceMemberTool{
			Value: types.SpecificToolChoice{
				Name: aws.String(tc.Function),
			},
		}
	case llm.ToolChoiceNone:
		// Bedrock doesn't have a "none" tool choice; omit tools instead.
		return nil
	default:
		return nil
	}
}

func mapResponse(output *bedrockruntime.ConverseOutput, model string) *llm.ChatCompletionResponse {
	resp := &llm.ChatCompletionResponse{
		Model:        model,
		FinishReason: mapStopReason(output.StopReason),
		Message: llm.Message{
			Role: llm.RoleAssistant,
		},
	}

	if output.Usage != nil {
		resp.Usage = llm.Usage{
			InputTokens:  int(aws.ToInt32(output.Usage.InputTokens)),
			OutputTokens: int(aws.ToInt32(output.Usage.OutputTokens)),
		}
	}

	// Extract message content from the response output union.
	if msgOutput, ok := output.Output.(*types.ConverseOutputMemberMessage); ok {
		for _, block := range msgOutput.Value.Content {
			switch b := block.(type) {
			case *types.ContentBlockMemberText:
				resp.Message.Parts = append(resp.Message.Parts, llm.TextPart{Text: b.Value})
			case *types.ContentBlockMemberToolUse:
				var args any
				if b.Value.Input != nil {
					_ = b.Value.Input.UnmarshalSmithyDocument(&args)
				}
				argsJSON, _ := json.Marshal(args)
				resp.Message.ToolCalls = append(resp.Message.ToolCalls, llm.ToolCall{
					ID: aws.ToString(b.Value.ToolUseId),
					Function: llm.FunctionCall{
						Name:      aws.ToString(b.Value.Name),
						Arguments: string(argsJSON),
					},
				})
			}
		}
	}

	return resp
}

func mapStopReason(reason types.StopReason) llm.FinishReason {
	switch reason {
	case types.StopReasonEndTurn, types.StopReasonStopSequence:
		return llm.FinishReasonStop
	case types.StopReasonMaxTokens:
		return llm.FinishReasonLength
	case types.StopReasonToolUse:
		return llm.FinishReasonToolCalls
	case types.StopReasonContentFiltered, types.StopReasonGuardrailIntervened:
		return llm.FinishReasonContentFilter
	default:
		return llm.FinishReasonStop
	}
}

func mapError(err error) error {
	var respErr *smithyhttp.ResponseError
	if !errors.As(err, &respErr) {
		// Check for common error types by message content.
		msg := err.Error()
		if strings.Contains(msg, "throttling") || strings.Contains(msg, "ThrottlingException") {
			return &llm.ErrRateLimit{Err: err}
		}
		return err
	}

	switch respErr.HTTPStatusCode() {
	case 429:
		return &llm.ErrRateLimit{Err: err}
	case 401, 403:
		return &llm.ErrAuthentication{Err: err}
	case 400:
		msg := err.Error()
		if strings.Contains(msg, "context") || strings.Contains(msg, "token") {
			return &llm.ErrContextLength{Err: err}
		}
		return err
	default:
		return err
	}
}

// bedrockStream adapts a Bedrock ConverseStream to our ChatCompletionStream interface.
type bedrockStream struct {
	eventStream *bedrockruntime.ConverseStreamEventStream
	events      <-chan types.ConverseStreamOutput
	current     llm.ChatCompletionStreamEvent
	err         error
	toolIndex   int
	inToolUse   bool
}

func newBedrockStream(eventStream *bedrockruntime.ConverseStreamEventStream) *bedrockStream {
	return &bedrockStream{
		eventStream: eventStream,
		events:      eventStream.Events(),
	}
}

func (s *bedrockStream) Next() bool {
	for event := range s.events {
		mapped, ok := s.mapEvent(event)
		if ok {
			s.current = mapped
			return true
		}
	}

	if err := s.eventStream.Err(); err != nil {
		s.err = mapError(err)
	}
	return false
}

func (s *bedrockStream) Event() llm.ChatCompletionStreamEvent {
	return s.current
}

func (s *bedrockStream) Err() error {
	return s.err
}

func (s *bedrockStream) Close() error {
	return s.eventStream.Close()
}

func (s *bedrockStream) mapEvent(event types.ConverseStreamOutput) (llm.ChatCompletionStreamEvent, bool) {
	switch e := event.(type) {
	case *types.ConverseStreamOutputMemberContentBlockStart:
		if start, ok := e.Value.Start.(*types.ContentBlockStartMemberToolUse); ok {
			s.inToolUse = true
			return llm.ChatCompletionStreamEvent{
				Delta: llm.MessageDelta{
					ToolCalls: []llm.ToolCallDelta{{
						Index: s.toolIndex,
						ID:    aws.ToString(start.Value.ToolUseId),
						Name:  aws.ToString(start.Value.Name),
					}},
				},
			}, true
		}
		s.inToolUse = false
		return llm.ChatCompletionStreamEvent{}, false

	case *types.ConverseStreamOutputMemberContentBlockDelta:
		switch d := e.Value.Delta.(type) {
		case *types.ContentBlockDeltaMemberText:
			return llm.ChatCompletionStreamEvent{
				Delta: llm.MessageDelta{Content: d.Value},
			}, true
		case *types.ContentBlockDeltaMemberToolUse:
			return llm.ChatCompletionStreamEvent{
				Delta: llm.MessageDelta{
					ToolCalls: []llm.ToolCallDelta{{
						Index:     s.toolIndex,
						Arguments: aws.ToString(d.Value.Input),
					}},
				},
			}, true
		}
		return llm.ChatCompletionStreamEvent{}, false

	case *types.ConverseStreamOutputMemberContentBlockStop:
		if s.inToolUse {
			s.toolIndex++
			s.inToolUse = false
		}
		return llm.ChatCompletionStreamEvent{}, false

	case *types.ConverseStreamOutputMemberMessageStop:
		fr := mapStopReason(e.Value.StopReason)
		return llm.ChatCompletionStreamEvent{
			FinishReason: &fr,
		}, true

	case *types.ConverseStreamOutputMemberMetadata:
		if e.Value.Usage != nil {
			return llm.ChatCompletionStreamEvent{
				Usage: &llm.Usage{
					InputTokens:  int(aws.ToInt32(e.Value.Usage.InputTokens)),
					OutputTokens: int(aws.ToInt32(e.Value.Usage.OutputTokens)),
				},
			}, true
		}
		return llm.ChatCompletionStreamEvent{}, false

	default:
		return llm.ChatCompletionStreamEvent{}, false
	}
}
