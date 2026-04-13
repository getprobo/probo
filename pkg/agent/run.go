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

package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.gearno.de/kit/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/llm"
)

const (
	tracerName = "go.probo.inc/probo/pkg/agent"

	// synthesisNudge is the static user message appended after tool
	// exploration completes, asking the model to produce the final
	// structured output on the next (synthesis) turn.
	synthesisNudge = "Based on everything you have gathered, produce the final structured output now."
)

type (
	CallLLMFunc func(ctx context.Context, agent *Agent, req *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error)

	RunOption func(*runOpts)

	runOpts struct {
		callLLM             CallLLMFunc
		onEvent             func(ctx context.Context, ev StreamEvent)
		skipInputGuardrails bool
		skipSessionLoad     bool
		initialUsage        llm.Usage
		initialTurns        int
		checkpointer        Checkpointer
		runID               string
		toolUsedInRun       bool
	}

	loopState struct {
		agent         *Agent
		toolMap       map[string]ToolDescriptor
		toolDefs      []llm.Tool
		messages      []llm.Message
		inputMessages []llm.Message
		systemPrompt  string
		totalUsage    llm.Usage
		turns         int
		toolUsedInRun bool

		tracer  trace.Tracer
		runSpan trace.Span
		opts    runOpts
		logger  *log.Logger
	}

	parallelToolEntry struct {
		result ToolResult
		err    error
	}
)

func WithCheckpointer(cp Checkpointer, runID string) RunOption {
	return func(o *runOpts) {
		o.checkpointer = cp
		o.runID = runID
	}
}

func noopEvent(_ context.Context, _ StreamEvent) {}

func blockingCallLLM(ctx context.Context, agent *Agent, req *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
	resp, err := agent.client.ChatCompletion(ctx, req)
	if err == nil {
		return resp, nil
	}

	// Some providers (e.g. Anthropic) require streaming for large
	// max_tokens or when thinking is enabled. Fall back to streaming
	// transparently when the blocking call returns ErrStreamingRequired.
	var streamRequired *llm.ErrStreamingRequired
	if !errors.As(err, &streamRequired) {
		return nil, err
	}

	stream, sErr := agent.client.ChatCompletionStream(ctx, req)
	if sErr != nil {
		return nil, err // return the original error
	}
	defer func() { _ = stream.Close() }()

	acc := llm.NewStreamAccumulator(stream)
	for acc.Next() {
	}
	if sErr := acc.Err(); sErr != nil {
		return nil, sErr
	}
	return acc.Response(), nil
}

func (a *Agent) Run(ctx context.Context, messages []llm.Message) (*Result, error) {
	return a.RunWithOpts(ctx, messages)
}

func (a *Agent) RunWithOpts(ctx context.Context, messages []llm.Message, opts ...RunOption) (*Result, error) {
	ro := runOpts{
		callLLM: blockingCallLLM,
		onEvent: noopEvent,
	}
	for _, opt := range opts {
		opt(&ro)
	}

	return coreLoop(ctx, a, messages, ro)
}

func (s *loopState) resolveAgentTools(ctx context.Context) error {
	tools, toolMap, err := s.agent.resolveTools(ctx)
	if err != nil {
		return err
	}

	toolDefs := make([]llm.Tool, len(tools))
	for i, t := range tools {
		toolDefs[i] = t.Definition()
	}

	s.toolMap = toolMap
	s.toolDefs = toolDefs
	return nil
}

func (s *loopState) finishRun(ctx context.Context, result *Result, err error) (*Result, error) {
	defer s.runSpan.End()
	defer func() {
		emitHook(s.agent, func(h RunHooks) { h.OnRunEnd(ctx, s.agent, result, err) })
	}()

	if err != nil {
		if _, ok := errors.AsType[*SuspendedError](err); ok {
			s.runSpan.SetAttributes(attribute.Bool("agent.suspended", true))
			s.opts.onEvent(ctx, StreamEvent{Type: StreamEventSuspended, Agent: s.agent})

			s.logger.InfoCtx(
				ctx,
				"agent run suspended",
				log.String("agent", s.agent.name),
				log.Int("turns", s.turns),
			)

			return result, err
		}

		s.runSpan.RecordError(err)
		s.runSpan.SetStatus(codes.Error, err.Error())
		s.opts.onEvent(ctx, StreamEvent{Type: StreamEventError, Agent: s.agent, Err: err})

		s.logger.ErrorCtx(
			ctx,
			"agent run failed",
			log.String("agent", s.agent.name),
			log.Int("turns", s.turns),
			log.Error(err),
		)

		return result, err
	}

	if s.agent.session != nil {
		if saveErr := s.agent.session.Save(ctx, s.agent.sessionID, result.Messages); saveErr != nil {
			err = fmt.Errorf("cannot save session: %w", saveErr)
			result = nil

			s.runSpan.RecordError(err)
			s.runSpan.SetStatus(codes.Error, err.Error())
			s.opts.onEvent(ctx, StreamEvent{Type: StreamEventError, Agent: s.agent, Err: err})

			s.logger.ErrorCtx(
				ctx,
				"cannot save session",
				log.String("agent", s.agent.name),
				log.Error(err),
			)

			return result, err
		}
	}

	s.runSpan.SetAttributes(
		attribute.Int("agent.turns", result.Turns),
		attribute.Int("agent.usage.input_tokens", result.Usage.InputTokens),
		attribute.Int("agent.usage.output_tokens", result.Usage.OutputTokens),
	)

	s.logger.InfoCtx(
		ctx,
		"agent run completed",
		log.String("agent", s.agent.name),
		log.Int("turns", result.Turns),
		log.Int("input_tokens", result.Usage.InputTokens),
		log.Int("output_tokens", result.Usage.OutputTokens),
	)

	s.opts.onEvent(ctx, StreamEvent{Type: StreamEventComplete, Agent: s.agent, Result: result})
	return result, err
}

func (s *loopState) buildCheckpoint(status AgentStatus) *Checkpoint {
	msgsCopy := make([]llm.Message, len(s.messages))
	copy(msgsCopy, s.messages)

	return &Checkpoint{
		Status:        status,
		AgentName:     s.agent.name,
		Messages:      msgsCopy,
		Usage:         s.totalUsage,
		Turns:         s.turns,
		ToolUsedInRun: s.toolUsedInRun,
	}
}

func (s *loopState) applyHandoff(ctx context.Context, handoffTarget *Handoff) error {
	emitHook(s.agent, func(h RunHooks) { h.OnHandoff(ctx, s.agent, handoffTarget.Agent) })
	emitAgentHook(handoffTarget.Agent, func(h AgentHooks) { h.OnHandoff(ctx, handoffTarget.Agent, s.agent) })

	s.opts.onEvent(ctx, StreamEvent{Type: StreamEventHandoff, Agent: handoffTarget.Agent})

	s.logger.InfoCtx(
		ctx,
		"agent handoff",
		log.String("from", s.agent.name),
		log.String("to", handoffTarget.Agent.name),
	)

	if handoffTarget.InputFilter != nil {
		s.messages = handoffTarget.InputFilter(
			HandoffInputData{
				InputHistory: s.inputMessages,
				NewItems:     s.messages[len(s.inputMessages):],
			},
		)
	}

	s.agent = handoffTarget.Agent
	s.logger = s.agent.logger

	if err := s.resolveAgentTools(ctx); err != nil {
		return err
	}

	emitAgentHook(s.agent, func(h AgentHooks) { h.OnStart(ctx, s.agent) })

	s.opts.onEvent(ctx, StreamEvent{Type: StreamEventAgentStart, Agent: s.agent})

	s.systemPrompt = s.agent.buildSystemPrompt(ctx)
	s.toolUsedInRun = false

	return nil
}

func coreLoop(ctx context.Context, startAgent *Agent, inputMessages []llm.Message, opts runOpts) (*Result, error) {
	s := &loopState{
		agent:         startAgent,
		inputMessages: inputMessages,
		systemPrompt:  startAgent.buildSystemPrompt(ctx),
		totalUsage:    opts.initialUsage,
		turns:         opts.initialTurns,
		toolUsedInRun: opts.toolUsedInRun,
		tracer:        otel.GetTracerProvider().Tracer(tracerName),
		opts:          opts,
		logger:        startAgent.logger,
	}

	s.messages = make([]llm.Message, 0, len(inputMessages))

	if !opts.skipSessionLoad && s.agent.session != nil {
		prev, err := s.agent.session.Load(ctx, s.agent.sessionID)
		if err != nil {
			return nil, fmt.Errorf("cannot load session: %w", err)
		}

		s.logger.InfoCtx(
			ctx,
			"session loaded",
			log.String("session_id", s.agent.sessionID),
			log.Int("message_count", len(prev)),
		)

		s.messages = append(s.messages, prev...)
	}

	s.messages = append(s.messages, inputMessages...)

	if !opts.skipInputGuardrails {
		if err := runInputGuardrails(ctx, s.agent, s.messages); err != nil {
			opts.onEvent(ctx, StreamEvent{Type: StreamEventError, Agent: s.agent, Err: err})

			s.logger.ErrorCtx(
				ctx,
				"input guardrail tripped",
				log.Error(err),
			)

			return nil, err
		}
	}

	emitHook(s.agent, func(h RunHooks) { h.OnRunStart(ctx, s.agent, s.messages) })
	emitAgentHook(s.agent, func(h AgentHooks) { h.OnStart(ctx, s.agent) })

	opts.onEvent(ctx, StreamEvent{Type: StreamEventAgentStart, Agent: s.agent})

	ctx, s.runSpan = s.tracer.Start(
		ctx,
		fmt.Sprintf("agent.run %s", s.agent.name),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("agent.name", s.agent.name),
			attribute.String("agent.model", s.agent.model),
		),
	)

	if err := s.resolveAgentTools(ctx); err != nil {
		s.runSpan.End()
		return nil, err
	}

	s.logger.InfoCtx(
		ctx,
		"agent run started",
		log.String("model", s.agent.model),
		log.Int("max_turns", s.agent.maxTurns),
		log.Int("tool_count", len(s.toolDefs)),
	)

	emptyOutputRetries := 0

	structuredFormat := resolveStructuredFormat(s.agent)

	// When the agent has both tools and a structured output request,
	// we delay structured output enforcement until a dedicated
	// synthesis turn. Enforcing the schema during tool exploration
	// causes models with extended thinking to stuff planning prose
	// into the first text field of the schema as a scratchpad,
	// burning the entire max_tokens budget on thinking-inside-JSON
	// before ever producing a valid object. Instead, we let the
	// model freely call tools without a schema, then force one final
	// synthesis turn with ToolChoice=none + schema enforced once the
	// model signals it has enough information (finish_reason=stop).
	// Agents without tools or without a structured output request
	// do not need this dance and enforce the schema immediately.
	exploring := structuredFormat != nil && len(s.toolDefs) > 0

	for {
		if err := ctx.Err(); err != nil {
			return s.finishRun(ctx, nil, fmt.Errorf("cannot complete: %w", err))
		}

		if s.turns >= s.agent.maxTurns {
			return s.finishRun(ctx, nil, &MaxTurnsExceededError{MaxTurns: s.agent.maxTurns})
		}

		if ch := stopSignalFrom(ctx); ch != nil {
			select {
			case <-ch:
				cp := s.buildCheckpoint(AgentStatusSuspended)
				se := &SuspendedError{RunID: s.opts.runID}

				if s.opts.checkpointer != nil && s.opts.runID != "" {
					if saveErr := s.opts.checkpointer.Save(ctx, s.opts.runID, cp); saveErr != nil {
						s.logger.ErrorCtx(ctx, "cannot save suspension checkpoint", log.Error(saveErr))
						se.Checkpoint = cp
					}
				} else {
					se.Checkpoint = cp
				}

				return s.finishRun(ctx, nil, se)
			default:
			}
		}

		fullMessages := buildFullMessages(s.systemPrompt, s.messages)

		var responseFormat *llm.ResponseFormat
		if !exploring {
			responseFormat = structuredFormat
		}

		toolChoice := s.agent.modelSettings.ToolChoice
		if s.toolUsedInRun && s.agent.resetToolChoice && toolChoice != nil {
			toolChoice = nil
		}
		if !exploring && structuredFormat != nil && len(s.toolDefs) > 0 {
			// On the synthesis turn, forbid further tool calls so the
			// model is forced to convert what it has into JSON.
			none := llm.ToolChoice{Type: llm.ToolChoiceNone}
			toolChoice = &none
		}

		req := &llm.ChatCompletionRequest{
			Model:             s.agent.model,
			Messages:          fullMessages,
			Tools:             s.toolDefs,
			Temperature:       s.agent.modelSettings.Temperature,
			TopP:              s.agent.modelSettings.TopP,
			FrequencyPenalty:  s.agent.modelSettings.FrequencyPenalty,
			PresencePenalty:   s.agent.modelSettings.PresencePenalty,
			MaxTokens:         s.agent.modelSettings.MaxTokens,
			ToolChoice:        toolChoice,
			ParallelToolCalls: s.agent.modelSettings.ParallelToolCalls,
			ResponseFormat:    responseFormat,
			Thinking:          s.agent.modelSettings.Thinking,
		}

		s.logger.InfoCtx(
			ctx,
			"calling LLM",
			log.Int("turn", s.turns+1),
			log.Int("message_count", len(fullMessages)),
		)

		resp, err := callLLMWithHooks(ctx, s.agent, req, opts)
		if err != nil {
			return s.finishRun(ctx, nil, fmt.Errorf("cannot call LLM: %w", err))
		}

		s.totalUsage = s.totalUsage.Add(resp.Usage)
		s.turns++

		s.logger.InfoCtx(
			ctx,
			"LLM response received",
			log.Int("turn", s.turns),
			log.String("finish_reason", string(resp.FinishReason)),
			log.Int("input_tokens", resp.Usage.InputTokens),
			log.Int("output_tokens", resp.Usage.OutputTokens),
		)

		s.messages = append(s.messages, resp.Message)

		switch resp.FinishReason {
		case llm.FinishReasonStop, llm.FinishReasonLength:
			// Model signalled it has nothing more to do with tools.
			// If we have a structured output request but haven't
			// enforced the schema yet, promote this turn to the
			// synthesis turn: the next iteration runs with
			// ToolChoice=none and the schema enforced, so the model
			// converts what it has gathered into JSON in one shot.
			//
			// Anthropic requires the last message in the conversation
			// to be a user message, so we cannot simply continue after
			// an assistant stop turn. Drop empty (thinking-only) turns
			// from history and append a user nudge that asks for the
			// final structured output. Non-empty assistant turns stay
			// in history so the model can reference its own
			// conclusions during synthesis.
			if exploring && s.turns < s.agent.maxTurns {
				exploring = false
				if resp.Message.Text() == "" {
					s.messages = s.messages[:len(s.messages)-1]
				}
				s.messages = append(
					s.messages,
					llm.Message{
						Role:  llm.RoleUser,
						Parts: []llm.Part{llm.TextPart{Text: synthesisNudge}},
					},
				)
				s.logger.WarnCtx(
					ctx,
					"entering synthesis turn: forcing structured output with tool_choice=none",
					log.Int("turn", s.turns),
					log.Int("output_tokens", resp.Usage.OutputTokens),
				)
				continue
			}

			// Anthropic extended-thinking models can return a synthesis turn
			// that contains only thinking blocks and no text part, leaving us
			// with no structured output to validate. Retry the same turn a
			// bounded number of times so the model gets another chance to
			// emit the required JSON output. The empty assistant turn must be
			// dropped from history because Anthropic rejects requests where
			// the last message is a thinking-only assistant turn.
			if structuredFormat != nil && resp.Message.Text() == "" && emptyOutputRetries < s.agent.maxEmptyOutputRetries && s.turns < s.agent.maxTurns {
				emptyOutputRetries++
				s.messages = s.messages[:len(s.messages)-1]
				s.logger.WarnCtx(
					ctx,
					"retrying turn: structured output expected but got empty text",
					log.Int("turn", s.turns),
					log.Int("retry", emptyOutputRetries),
					log.Int("output_tokens", resp.Usage.OutputTokens),
				)
				continue
			}
			if err := runOutputGuardrails(ctx, s.agent, resp.Message); err != nil {
				return s.finishRun(ctx, nil, err)
			}

			result := &Result{
				Messages:  s.messages,
				Usage:     s.totalUsage,
				Turns:     s.turns,
				LastAgent: s.agent,
			}

			emitAgentHook(s.agent, func(h AgentHooks) { h.OnEnd(ctx, s.agent, resp.Message.Text()) })

			opts.onEvent(ctx, StreamEvent{Type: StreamEventAgentEnd, Agent: s.agent})
			return s.finishRun(ctx, result, nil)

		case llm.FinishReasonToolCalls:
			s.toolUsedInRun = true
			emptyOutputRetries = 0

			s.logger.InfoCtx(
				ctx,
				"dispatching tool calls",
				log.Int("count", len(resp.Message.ToolCalls)),
			)

			handoffTarget, toolResults, toolMsgs, err := dispatchToolCalls(
				ctx,
				s.tracer,
				s.agent,
				resp.Message.ToolCalls,
				s.toolMap,
				opts.onEvent,
				s.logger,
			)
			s.messages = append(s.messages, toolMsgs...)

			if err != nil {
				if se, ok := errors.AsType[*SuspendedError](err); ok {
					outerCP := s.buildCheckpoint(AgentStatusSuspended)
					if se.Checkpoint != nil {
						outerCP.AllToolCalls = se.Checkpoint.AllToolCalls
						outerCP.InnerCheckpoints = se.Checkpoint.InnerCheckpoints
						outerCP.CompletedCalls = se.Checkpoint.CompletedCalls
					}
					if s.opts.checkpointer != nil && s.opts.runID != "" {
						if saveErr := s.opts.checkpointer.Save(ctx, s.opts.runID, outerCP); saveErr != nil {
							s.logger.ErrorCtx(ctx, "cannot save checkpoint", log.Error(saveErr))
						}
					}
					return s.finishRun(ctx, nil, &SuspendedError{RunID: s.opts.runID, Checkpoint: outerCP})
				}

				if nae, ok := errors.AsType[*needsApprovalError](err); ok {
					s.logger.InfoCtx(
						ctx,
						"run interrupted, approval required",
						log.Int("pending_count", len(nae.pendingApprovals)),
					)

					msgsCopy := make([]llm.Message, len(s.messages))
					copy(msgsCopy, s.messages)

					if s.opts.checkpointer != nil && s.opts.runID != "" {
						cp := s.buildCheckpoint(AgentStatusAwaitingApproval)
						cp.PendingToolCalls = nae.allToolCalls
						cp.PendingApprovals = nae.pendingApprovals
						if saveErr := s.opts.checkpointer.Save(ctx, s.opts.runID, cp); saveErr != nil {
							s.logger.ErrorCtx(ctx, "cannot save approval checkpoint", log.Error(saveErr))
						}
					}

					return s.finishRun(
						ctx,
						nil,
						&InterruptedError{
							ToolCalls:        nae.allToolCalls,
							PendingApprovals: nae.pendingApprovals,
							Agent:            s.agent,
							Messages:         msgsCopy,
							Usage:            s.totalUsage,
							Turns:            s.turns,
						},
					)
				}

				if nie, ok := errors.AsType[*nestedInterruptionError](err); ok {
					s.logger.InfoCtx(
						ctx,
						"nested agent interrupted, approval required",
						log.Int("pending_count", len(nie.inner.PendingApprovals)),
						log.String("nested_agent", nie.inner.Agent.name),
					)

					msgsCopy := make([]llm.Message, len(s.messages))
					copy(msgsCopy, s.messages)

					if s.opts.checkpointer != nil && s.opts.runID != "" {
						cp := s.buildCheckpoint(AgentStatusAwaitingApproval)
						cp.PendingToolCalls = nie.inner.ToolCalls
						cp.PendingApprovals = nie.inner.PendingApprovals
						cp.AllToolCalls = nie.allToolCalls
						cp.CompletedCalls = nie.completedCalls
						cp.InnerCheckpoints = map[string]*Checkpoint{
							nie.toolCallID: {
								Status:           AgentStatusAwaitingApproval,
								AgentName:        nie.inner.Agent.name,
								Messages:         nie.inner.Messages,
								Usage:            nie.inner.Usage,
								Turns:            nie.inner.Turns,
								PendingToolCalls: nie.inner.ToolCalls,
								PendingApprovals: nie.inner.PendingApprovals,
							},
						}
						if saveErr := s.opts.checkpointer.Save(ctx, s.opts.runID, cp); saveErr != nil {
							s.logger.ErrorCtx(ctx, "cannot save nested approval checkpoint", log.Error(saveErr))
						}
					}

					return s.finishRun(
						ctx,
						nil,
						&InterruptedError{
							ToolCalls:        nie.inner.ToolCalls,
							PendingApprovals: nie.inner.PendingApprovals,
							Agent:            nie.inner.Agent,
							Messages:         nie.inner.Messages,
							Usage:            nie.inner.Usage,
							Turns:            nie.inner.Turns,
							outerState: &outerLoopState{
								agent:          s.agent,
								messages:       msgsCopy,
								usage:          s.totalUsage,
								turns:          s.turns,
								allToolCalls:   nie.allToolCalls,
								toolCallID:     nie.toolCallID,
								completedCalls: nie.completedCalls,
								innerInterrupt: nie.inner,
							},
						},
					)
				}

				return s.finishRun(ctx, nil, err)
			}

			finalOutput, isFinal, behaviorErr := s.agent.toolUseBehavior(ctx, toolResults)
			if behaviorErr != nil {
				return s.finishRun(ctx, nil, fmt.Errorf("cannot evaluate tool use behavior: %w", behaviorErr))
			}

			if isFinal {
				s.messages = append(
					s.messages,
					llm.Message{
						Role:  llm.RoleAssistant,
						Parts: []llm.Part{llm.TextPart{Text: finalOutput}},
					},
				)

				result := &Result{
					Messages:  s.messages,
					Usage:     s.totalUsage,
					Turns:     s.turns,
					LastAgent: s.agent,
				}

				emitAgentHook(s.agent, func(h AgentHooks) { h.OnEnd(ctx, s.agent, finalOutput) })

				opts.onEvent(ctx, StreamEvent{Type: StreamEventAgentEnd, Agent: s.agent})
				return s.finishRun(ctx, result, nil)
			}

			if handoffTarget != nil {
				if handoffErr := s.applyHandoff(ctx, handoffTarget); handoffErr != nil {
					return s.finishRun(ctx, nil, handoffErr)
				}
			}

			// Save incremental checkpoint after completed tool-call turn.
			if s.opts.checkpointer != nil && s.opts.runID != "" {
				cp := s.buildCheckpoint(AgentStatusSuspended)
				if saveErr := s.opts.checkpointer.Save(ctx, s.opts.runID, cp); saveErr != nil {
					s.logger.ErrorCtx(ctx, "cannot save checkpoint", log.Error(saveErr))
				}
			}

		case llm.FinishReasonContentFilter:
			return s.finishRun(ctx, nil, fmt.Errorf("cannot complete: content was filtered by the provider"))

		default:
			return s.finishRun(ctx, nil, fmt.Errorf("cannot complete: unexpected finish reason %q", resp.FinishReason))
		}
	}
}

func buildFullMessages(systemPrompt string, messages []llm.Message) []llm.Message {
	fullMessages := make([]llm.Message, 0, len(messages)+1)
	if systemPrompt != "" {
		fullMessages = append(
			fullMessages,
			llm.Message{
				Role:  llm.RoleSystem,
				Parts: []llm.Part{llm.TextPart{Text: systemPrompt}},
			},
		)
	}

	return append(fullMessages, messages...)
}

func callLLMWithHooks(
	ctx context.Context,
	agent *Agent,
	req *llm.ChatCompletionRequest,
	opts runOpts,
) (*llm.ChatCompletionResponse, error) {
	emitHook(agent, func(h RunHooks) { h.OnLLMStart(ctx, agent, req.Messages) })
	emitAgentHook(agent, func(h AgentHooks) { h.OnLLMStart(ctx, agent, req.Messages) })

	resp, err := opts.callLLM(ctx, agent, req)
	if err != nil {
		emitHook(agent, func(h RunHooks) { h.OnLLMEnd(ctx, agent, nil, err) })
		emitAgentHook(agent, func(h AgentHooks) { h.OnLLMEnd(ctx, agent, nil, err) })
		return nil, err
	}

	emitHook(agent, func(h RunHooks) { h.OnLLMEnd(ctx, agent, resp, nil) })
	emitAgentHook(agent, func(h AgentHooks) { h.OnLLMEnd(ctx, agent, resp, nil) })

	return resp, nil
}

func dispatchToolCalls(
	ctx context.Context,
	tracer trace.Tracer,
	agent *Agent,
	toolCalls []llm.ToolCall,
	toolMap map[string]ToolDescriptor,
	onEvent func(context.Context, StreamEvent),
	logger *log.Logger,
) (*Handoff, []ToolCallResult, []llm.Message, error) {
	if err := checkApproval(ctx, agent, toolCalls); err != nil {
		return nil, nil, nil, err
	}

	return executeToolCalls(ctx, tracer, agent, toolCalls, toolMap, onEvent, logger)
}

func executeToolCalls(
	ctx context.Context,
	tracer trace.Tracer,
	agent *Agent,
	toolCalls []llm.ToolCall,
	toolMap map[string]ToolDescriptor,
	onEvent func(context.Context, StreamEvent),
	logger *log.Logger,
) (*Handoff, []ToolCallResult, []llm.Message, error) {
	descriptors := make([]ToolDescriptor, len(toolCalls))
	handoffIdx := -1

	for i, tc := range toolCalls {
		desc, ok := toolMap[tc.Function.Name]
		if !ok {
			return nil, nil, nil, fmt.Errorf("cannot dispatch tool call: unknown tool %q", tc.Function.Name)
		}
		descriptors[i] = desc
		if _, isHandoff := desc.(*handoffToolAdapter); isHandoff && handoffIdx == -1 {
			handoffIdx = i
		}
	}

	if handoffIdx >= 0 {
		return executeWithHandoff(ctx, tracer, agent, toolCalls, descriptors, handoffIdx, onEvent, logger)
	}

	tools := make([]Tool, len(descriptors))
	for i, d := range descriptors {
		tools[i] = d.(Tool)
	}

	results, msgs, err := executeParallel(ctx, tracer, agent, toolCalls, tools, onEvent, logger)
	return nil, results, msgs, err
}

func executeWithHandoff(
	ctx context.Context,
	tracer trace.Tracer,
	agent *Agent,
	toolCalls []llm.ToolCall,
	descriptors []ToolDescriptor,
	handoffIdx int,
	onEvent func(context.Context, StreamEvent),
	logger *log.Logger,
) (*Handoff, []ToolCallResult, []llm.Message, error) {
	var (
		results []ToolCallResult
		msgs    []llm.Message
	)

	for i := range handoffIdx {
		logger.InfoCtx(
			ctx,
			"executing tool before handoff",
			log.String("tool", toolCalls[i].Function.Name),
		)

		tr, err := executeSingleTool(ctx, tracer, agent, toolCalls[i], descriptors[i].(Tool), onEvent, logger)
		if err != nil {
			if ie, ok := errors.AsType[*InterruptedError](err); ok {
				var completed []CompletedCall
				for j := range results {
					completed = append(
						completed,
						CompletedCall{
							ToolCallID: toolCalls[j].ID,
							Result:     results[j].Result,
						},
					)
				}
				return nil, nil, msgs, &nestedInterruptionError{
					inner:          ie,
					toolCallID:     toolCalls[i].ID,
					allToolCalls:   toolCalls,
					completedCalls: completed,
				}
			}
			return nil, nil, msgs, err
		}

		msgs = append(
			msgs,
			llm.Message{
				Role:       llm.RoleTool,
				ToolCallID: toolCalls[i].ID,
				Parts:      []llm.Part{llm.TextPart{Text: tr.Content}},
			},
		)

		results = append(
			results,
			ToolCallResult{
				ToolName:  toolCalls[i].Function.Name,
				Arguments: toolCalls[i].Function.Arguments,
				Result:    tr,
			},
		)
	}

	ht := descriptors[handoffIdx].(*handoffToolAdapter)
	if ht.handoff.OnHandoff != nil {
		if err := ht.handoff.OnHandoff(ctx); err != nil {
			return nil, nil, msgs, fmt.Errorf("cannot execute handoff callback for %q: %w", ht.handoff.Agent.name, err)
		}
	}

	msgs = append(
		msgs,
		llm.Message{
			Role:       llm.RoleTool,
			ToolCallID: toolCalls[handoffIdx].ID,
			Parts:      []llm.Part{llm.TextPart{Text: fmt.Sprintf("Transferred to %s", ht.handoff.Agent.name)}},
		},
	)

	for i := handoffIdx + 1; i < len(toolCalls); i++ {
		msgs = append(
			msgs,
			llm.Message{
				Role:       llm.RoleTool,
				ToolCallID: toolCalls[i].ID,
				Parts: []llm.Part{
					llm.TextPart{Text: "Tool call was not executed because a handoff occurred."},
				},
			},
		)
	}

	return ht.handoff, results, msgs, nil
}

func executeParallel(
	ctx context.Context,
	tracer trace.Tracer,
	agent *Agent,
	toolCalls []llm.ToolCall,
	tools []Tool,
	onEvent func(context.Context, StreamEvent),
	logger *log.Logger,
) ([]ToolCallResult, []llm.Message, error) {
	entries := make([]parallelToolEntry, len(toolCalls))
	var wg sync.WaitGroup

	wg.Add(len(toolCalls))
	for i := range toolCalls {
		go func(idx int, tc llm.ToolCall, tool Tool) {
			defer wg.Done()

			tr, err := executeSingleTool(ctx, tracer, agent, tc, tool, onEvent, logger)
			if err != nil {
				entries[idx] = parallelToolEntry{err: err}
				return
			}
			entries[idx] = parallelToolEntry{result: tr}
		}(i, toolCalls[i], tools[i])
	}

	wg.Wait()

	for i, entry := range entries {
		ie, ok := errors.AsType[*InterruptedError](entry.err)
		if !ok {
			continue
		}

		var completed []CompletedCall
		for j, other := range entries {
			if j == i {
				continue
			}
			if other.err != nil {
				completed = append(
					completed,
					CompletedCall{
						ToolCallID: toolCalls[j].ID,
						Result: ToolResult{
							Content: fmt.Sprintf("Error: %s", other.err.Error()),
							IsError: true,
						},
					},
				)
				continue
			}
			completed = append(
				completed,
				CompletedCall{
					ToolCallID: toolCalls[j].ID,
					Result:     other.result,
				},
			)
		}
		return nil, nil, &nestedInterruptionError{
			inner:          ie,
			toolCallID:     toolCalls[i].ID,
			allToolCalls:   toolCalls,
			completedCalls: completed,
		}
	}

	// Check for suspended inner agents (stop signal propagated).
	for i, entry := range entries {
		if entry.err == nil {
			continue
		}
		se, ok := errors.AsType[*SuspendedError](entry.err)
		if ok {
			innerCheckpoints := make(map[string]*Checkpoint)
			var completed []CompletedCall

			for j, other := range entries {
				if j == i {
					continue
				}
				if other.err == nil {
					completed = append(
						completed,
						CompletedCall{
							ToolCallID: toolCalls[j].ID,
							Result:     other.result,
						},
					)
					continue
				}
				otherSE, ok := errors.AsType[*SuspendedError](other.err)
				if ok && otherSE.Checkpoint != nil {
					innerCheckpoints[toolCalls[j].ID] = otherSE.Checkpoint
				} else {
					completed = append(
						completed,
						CompletedCall{
							ToolCallID: toolCalls[j].ID,
							Result: ToolResult{
								Content: fmt.Sprintf("Error: %s", other.err.Error()),
								IsError: true,
							},
						},
					)
				}
			}

			innerCheckpoints[toolCalls[i].ID] = se.Checkpoint

			outerSE := &SuspendedError{
				Checkpoint: &Checkpoint{
					Status:           AgentStatusSuspended,
					AllToolCalls:     toolCalls,
					InnerCheckpoints: innerCheckpoints,
					CompletedCalls:   completed,
				},
			}
			return nil, nil, outerSE
		}
	}

	var (
		results []ToolCallResult
		msgs    []llm.Message
	)

	for i, tc := range toolCalls {
		entry := entries[i]

		if entry.err != nil {
			msgs = append(
				msgs,
				llm.Message{
					Role:       llm.RoleTool,
					ToolCallID: tc.ID,
					Parts: []llm.Part{
						llm.TextPart{
							Text: fmt.Sprintf("Error: %s", entry.err.Error()),
						},
					},
				},
			)
			results = append(
				results,
				ToolCallResult{
					ToolName:  tc.Function.Name,
					Arguments: tc.Function.Arguments,
					Result: ToolResult{
						Content: fmt.Sprintf("Error: %s", entry.err.Error()),
						IsError: true,
					},
				},
			)
			continue
		}

		msgs = append(
			msgs,
			llm.Message{
				Role:       llm.RoleTool,
				ToolCallID: tc.ID,
				Parts:      []llm.Part{llm.TextPart{Text: entry.result.Content}},
			},
		)

		results = append(
			results,
			ToolCallResult{
				ToolName:  tc.Function.Name,
				Arguments: tc.Function.Arguments,
				Result:    entry.result,
			},
		)
	}

	return results, msgs, nil
}

func executeSingleTool(
	ctx context.Context,
	tracer trace.Tracer,
	agent *Agent,
	tc llm.ToolCall,
	tool Tool,
	onEvent func(context.Context, StreamEvent),
	logger *log.Logger,
) (ToolResult, error) {
	onEvent(ctx, StreamEvent{Type: StreamEventToolStart, Agent: agent, Tool: tool})

	emitHook(agent, func(h RunHooks) { h.OnToolStart(ctx, agent, tool, tc.Function.Arguments) })
	emitAgentHook(agent, func(h AgentHooks) { h.OnToolStart(ctx, agent, tool) })

	toolCtx, toolSpan := tracer.Start(
		ctx,
		fmt.Sprintf("agent.tool %s", tool.Name()),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("tool.name", tool.Name()),
		),
	)

	logger.InfoCtx(
		ctx,
		"executing tool",
		log.String("tool", tool.Name()),
	)

	result, err := tool.Execute(toolCtx, tc.Function.Arguments)
	if err != nil {
		if _, ok := errors.AsType[*InterruptedError](err); ok {
			toolSpan.SetAttributes(attribute.Bool("tool.interrupted", true))
			toolSpan.End()

			onEvent(ctx, StreamEvent{Type: StreamEventToolEnd, Agent: agent, Tool: tool, Err: err})
			emitHook(agent, func(h RunHooks) { h.OnToolEnd(ctx, agent, tool, ToolResult{}, err) })
			emitAgentHook(agent, func(h AgentHooks) { h.OnToolEnd(ctx, agent, tool, ToolResult{}) })

			return ToolResult{}, err
		}

		if _, ok := errors.AsType[*SuspendedError](err); ok {
			toolSpan.SetAttributes(attribute.Bool("tool.suspended", true))
			toolSpan.End()

			onEvent(ctx, StreamEvent{Type: StreamEventToolEnd, Agent: agent, Tool: tool, Err: err})
			emitHook(agent, func(h RunHooks) { h.OnToolEnd(ctx, agent, tool, ToolResult{}, err) })
			emitAgentHook(agent, func(h AgentHooks) { h.OnToolEnd(ctx, agent, tool, ToolResult{}) })

			return ToolResult{}, err
		}

		toolSpan.RecordError(err)
		toolSpan.SetStatus(codes.Error, err.Error())
		toolSpan.End()

		onEvent(ctx, StreamEvent{Type: StreamEventToolEnd, Agent: agent, Tool: tool, Err: err})
		emitHook(agent, func(h RunHooks) { h.OnToolEnd(ctx, agent, tool, result, err) })
		emitAgentHook(agent, func(h AgentHooks) { h.OnToolEnd(ctx, agent, tool, result) })

		logger.ErrorCtx(
			ctx,
			"tool execution failed",
			log.String("tool", tool.Name()),
			log.Error(err),
		)

		return ToolResult{}, fmt.Errorf("cannot execute tool %q: %w", tool.Name(), err)
	}

	toolSpan.SetAttributes(attribute.Bool("tool.is_error", result.IsError))
	toolSpan.End()

	onEvent(ctx, StreamEvent{Type: StreamEventToolEnd, Agent: agent, Tool: tool, ToolResult: &result})

	emitHook(agent, func(h RunHooks) { h.OnToolEnd(ctx, agent, tool, result, nil) })
	emitAgentHook(agent, func(h AgentHooks) { h.OnToolEnd(ctx, agent, tool, result) })

	if result.IsError {
		content := result.Content
		if len(content) > 200 {
			content = content[:200] + "... (truncated)"
		}
		logger.WarnCtx(
			ctx,
			"tool returned error",
			log.String("tool", tool.Name()),
			log.String("content", content),
		)
	} else {
		logger.InfoCtx(
			ctx,
			"tool execution completed",
			log.String("tool", tool.Name()),
		)
	}

	return result, nil
}

func checkApproval(ctx context.Context, a *Agent, toolCalls []llm.ToolCall) error {
	if a.approval == nil {
		return nil
	}

	var pending []llm.ToolCall
	for _, tc := range toolCalls {
		if a.approval.requiresApproval(ctx, tc) {
			pending = append(pending, tc)
		}
	}

	if len(pending) > 0 {
		return &needsApprovalError{
			allToolCalls:     toolCalls,
			pendingApprovals: pending,
		}
	}

	return nil
}

func runInputGuardrails(ctx context.Context, agent *Agent, messages []llm.Message) error {
	for _, g := range agent.inputGuardrails {
		result, err := g.Check(ctx, messages)
		if err != nil {
			return fmt.Errorf("cannot run input guardrail %q: %w", g.Name(), err)
		}

		if result != nil && result.Tripwire {
			emitHook(agent, func(h RunHooks) { h.OnGuardrailTripped(ctx, agent, g.Name(), result) })

			return &InputGuardrailTrippedError{
				Guardrail: g.Name(),
				Message:   result.Message,
			}
		}
	}

	return nil
}

func runOutputGuardrails(ctx context.Context, agent *Agent, message llm.Message) error {
	for _, g := range agent.outputGuardrails {
		result, err := g.Check(ctx, message)
		if err != nil {
			return fmt.Errorf("cannot run output guardrail %q: %w", g.Name(), err)
		}

		if result != nil && result.Tripwire {
			emitHook(agent, func(h RunHooks) { h.OnGuardrailTripped(ctx, agent, g.Name(), result) })

			return &OutputGuardrailTrippedError{
				Guardrail: g.Name(),
				Message:   result.Message,
			}
		}
	}

	return nil
}

// Resume continues an interrupted run after human approval decisions have
// been collected. It executes or denies each pending tool call according to
// the provided ResumeInput, then re-enters the agent loop. Input guardrails
// are not re-evaluated because the messages were already validated in the
// original Run call.
func Resume(ctx context.Context, interrupted *InterruptedError, input ResumeInput) (*Result, error) {
	return ResumeWithOpts(ctx, interrupted, input)
}

func ResumeWithOpts(ctx context.Context, interrupted *InterruptedError, input ResumeInput, opts ...RunOption) (*Result, error) {
	ro := runOpts{
		callLLM: blockingCallLLM,
		onEvent: noopEvent,
	}
	for _, opt := range opts {
		opt(&ro)
	}

	return resumeWithOpts(ctx, interrupted, input, ro)
}

func resumeWithOpts(ctx context.Context, interrupted *InterruptedError, input ResumeInput, ro runOpts) (*Result, error) {
	if interrupted.outerState != nil {
		return resumeNested(ctx, interrupted, input, ro)
	}

	tracer := otel.GetTracerProvider().Tracer(tracerName)
	logger := interrupted.Agent.logger

	agent := interrupted.Agent
	messages := interrupted.Messages

	logger.InfoCtx(
		ctx,
		"resuming interrupted run",
		log.String("agent", agent.name),
		log.Int("pending_approvals", len(interrupted.PendingApprovals)),
		log.Int("total_tool_calls", len(interrupted.ToolCalls)),
	)

	_, toolMap, err := agent.resolveTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve tools for resume: %w", err)
	}

	pendingSet := make(map[string]struct{}, len(interrupted.PendingApprovals))
	for _, tc := range interrupted.PendingApprovals {
		pendingSet[tc.ID] = struct{}{}
	}

	var handoffTarget *Handoff

	for _, tc := range interrupted.ToolCalls {
		if _, needsApproval := pendingSet[tc.ID]; needsApproval {
			approval, ok := input.Approvals[tc.ID]
			if !ok || !approval.Approved {
				reason := "Tool call was denied by human review."
				if ok && approval.Message != "" {
					reason = approval.Message
				}

				logger.InfoCtx(
					ctx,
					"tool call denied",
					log.String("tool", tc.Function.Name),
					log.String("tool_call_id", tc.ID),
					log.String("reason", reason),
				)

				messages = append(
					messages,
					llm.Message{
						Role:       llm.RoleTool,
						ToolCallID: tc.ID,
						Parts: []llm.Part{
							llm.TextPart{Text: reason},
						},
					},
				)

				continue
			}

			logger.InfoCtx(
				ctx,
				"tool call approved",
				log.String("tool", tc.Function.Name),
				log.String("tool_call_id", tc.ID),
			)
		}

		desc, toolOK := toolMap[tc.Function.Name]
		if !toolOK {
			return nil, fmt.Errorf("unknown tool %q", tc.Function.Name)
		}

		if ht, ok := desc.(*handoffToolAdapter); ok {
			if ht.handoff.OnHandoff != nil {
				if err := ht.handoff.OnHandoff(ctx); err != nil {
					return nil, fmt.Errorf("cannot execute handoff callback for %q: %w", ht.handoff.Agent.name, err)
				}
			}

			messages = append(
				messages,
				llm.Message{
					Role:       llm.RoleTool,
					ToolCallID: tc.ID,
					Parts: []llm.Part{
						llm.TextPart{
							Text: fmt.Sprintf("Transferred to %s", ht.handoff.Agent.name),
						},
					},
				},
			)

			handoffTarget = ht.handoff
			break
		}

		tr, execErr := executeSingleTool(ctx, tracer, agent, tc, desc.(Tool), ro.onEvent, logger)
		if execErr != nil {
			return nil, execErr
		}

		messages = append(
			messages,
			llm.Message{
				Role:       llm.RoleTool,
				ToolCallID: tc.ID,
				Parts: []llm.Part{
					llm.TextPart{
						Text: tr.Content,
					},
				},
			},
		)
	}

	resumeAgent := agent
	if handoffTarget != nil {
		logger.InfoCtx(
			ctx,
			"resume handoff",
			log.String("from", agent.name),
			log.String("to", handoffTarget.Agent.name),
		)

		emitHook(agent, func(h RunHooks) { h.OnHandoff(ctx, agent, handoffTarget.Agent) })
		emitAgentHook(handoffTarget.Agent, func(h AgentHooks) { h.OnHandoff(ctx, handoffTarget.Agent, agent) })

		if handoffTarget.InputFilter != nil {
			filtered := handoffTarget.InputFilter(
				HandoffInputData{
					InputHistory: interrupted.Messages,
					NewItems:     messages[len(interrupted.Messages):],
				},
			)
			messages = filtered
		}

		resumeAgent = handoffTarget.Agent
	}

	return coreLoop(
		ctx,
		resumeAgent,
		messages,
		runOpts{
			callLLM:             ro.callLLM,
			onEvent:             ro.onEvent,
			skipInputGuardrails: true,
			skipSessionLoad:     true,
			initialUsage:        interrupted.Usage,
			initialTurns:        interrupted.Turns,
			checkpointer:        ro.checkpointer,
			runID:               ro.runID,
			toolUsedInRun:       ro.toolUsedInRun,
		},
	)
}

func resumeNested(ctx context.Context, interrupted *InterruptedError, input ResumeInput, ro runOpts) (*Result, error) {
	outer := interrupted.outerState
	logger := outer.agent.logger

	logger.InfoCtx(
		ctx,
		"resuming nested agent interruption",
		log.String("outer_agent", outer.agent.name),
		log.String("inner_agent", interrupted.Agent.name),
	)

	innerResult, err := resumeWithOpts(ctx, outer.innerInterrupt, input, ro)
	if err != nil {
		innerIE, ok := errors.AsType[*InterruptedError](err)
		if ok {
			return nil, &InterruptedError{
				ToolCalls:        innerIE.ToolCalls,
				PendingApprovals: innerIE.PendingApprovals,
				Agent:            innerIE.Agent,
				Messages:         innerIE.Messages,
				Usage:            innerIE.Usage,
				Turns:            innerIE.Turns,
				outerState: &outerLoopState{
					agent:          outer.agent,
					messages:       outer.messages,
					usage:          outer.usage,
					turns:          outer.turns,
					allToolCalls:   outer.allToolCalls,
					toolCallID:     outer.toolCallID,
					completedCalls: outer.completedCalls,
					innerInterrupt: innerIE,
				},
			}
		}
		return nil, fmt.Errorf("cannot resume nested agent: %w", err)
	}

	completedMap := make(map[string]ToolResult, len(outer.completedCalls))
	for _, cc := range outer.completedCalls {
		completedMap[cc.ToolCallID] = cc.Result
	}

	messages := make([]llm.Message, len(outer.messages))
	copy(messages, outer.messages)

	for _, tc := range outer.allToolCalls {
		var content string
		if tc.ID == outer.toolCallID {
			content = innerResult.FinalMessage().Text()
		} else if cr, ok := completedMap[tc.ID]; ok {
			content = cr.Content
		} else {
			content = "Error: tool execution was interrupted"
		}

		messages = append(
			messages,
			llm.Message{
				Role:       llm.RoleTool,
				ToolCallID: tc.ID,
				Parts:      []llm.Part{llm.TextPart{Text: content}},
			},
		)
	}

	return coreLoop(
		ctx,
		outer.agent,
		messages,
		runOpts{
			callLLM:             ro.callLLM,
			onEvent:             ro.onEvent,
			skipInputGuardrails: true,
			skipSessionLoad:     true,
			initialUsage:        outer.usage,
			initialTurns:        outer.turns,
			checkpointer:        ro.checkpointer,
			runID:               ro.runID,
			toolUsedInRun:       ro.toolUsedInRun,
		},
	)
}

func emitHook(agent *Agent, fn func(RunHooks)) {
	for _, h := range agent.hooks {
		fn(h)
	}
}

func emitAgentHook(agent *Agent, fn func(AgentHooks)) {
	if agent.agentHooks != nil {
		fn(agent.agentHooks)
	}
}

// resolveStructuredFormat returns the structured output request the
// agent wants enforced on its final turn, or nil if none. An agent can
// declare structured output through either WithOutputType (typed
// sub-agents) or a directly-set responseFormat (the RunTyped
// convenience wrapper).
func resolveStructuredFormat(a *Agent) *llm.ResponseFormat {
	if a.responseFormat != nil {
		return a.responseFormat
	}
	if a.outputType != nil {
		return a.outputType.responseFormat()
	}
	return nil
}
