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

	"go.probo.inc/probo/pkg/llm"
)

type (
	StreamEventType string

	StreamEvent struct {
		Type       StreamEventType
		Agent      *Agent
		Delta      string
		Tool       Tool
		ToolResult *ToolResult
		Result     *Result
		Err        error
	}

	StreamedRun struct {
		Events <-chan StreamEvent
		done   chan struct{}
		result *Result
		err    error
	}
)

const (
	StreamEventAgentStart StreamEventType = "agent_start"
	StreamEventAgentEnd   StreamEventType = "agent_end"
	StreamEventLLMDelta   StreamEventType = "llm_delta"
	StreamEventToolStart  StreamEventType = "tool_start"
	StreamEventToolEnd    StreamEventType = "tool_end"
	StreamEventHandoff    StreamEventType = "handoff"
	StreamEventComplete   StreamEventType = "complete"
	StreamEventError      StreamEventType = "error"
)

func (sr *StreamedRun) Wait() (*Result, error) {
	<-sr.done

	return sr.result, sr.err
}

func (a *Agent) RunStreamed(ctx context.Context, messages []llm.Message) *StreamedRun {
	events := make(chan StreamEvent, 64)
	sr := &StreamedRun{
		Events: events,
		done:   make(chan struct{}),
	}

	go func() {
		defer close(sr.done)
		defer close(events)

		result, err := coreLoop(
			ctx,
			a,
			messages,
			runOpts{
				callLLM: streamingCallLLM(events),
				onEvent: func(ctx context.Context, ev StreamEvent) {
					trySendEvent(ctx, events, ev)
				},
			},
		)

		sr.result = result
		sr.err = err
	}()

	return sr
}

func streamingCallLLM(events chan<- StreamEvent) CallLLMFunc {
	return func(ctx context.Context, agent *Agent, req *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
		stream, err := agent.client.ChatCompletionStream(ctx, req)
		if err != nil {
			return nil, err
		}

		acc := llm.NewStreamAccumulator(stream)
		for acc.Next() {
			ev := acc.Event()
			if ev.Delta.Content != "" {
				trySendEvent(
					ctx,
					events,
					StreamEvent{
						Type:  StreamEventLLMDelta,
						Agent: agent,
						Delta: ev.Delta.Content,
					},
				)
			}
		}

		if err := acc.Err(); err != nil {
			_ = stream.Close()
			return nil, err
		}

		if err := stream.Close(); err != nil {
			return nil, err
		}

		return acc.Response(), nil
	}
}

func trySendEvent(ctx context.Context, events chan<- StreamEvent, ev StreamEvent) {
	select {
	case events <- ev:
	case <-ctx.Done():
	}
}
