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
	StreamEventSuspended  StreamEventType = "suspended"
	StreamEventError      StreamEventType = "error"
)

func (sr *StreamedRun) Wait() (*Result, error) {
	<-sr.done

	return sr.result, sr.err
}

// RunStreamed launches the agent loop and returns immediately with a
// StreamedRun whose Events channel emits incremental progress. ctx
// follows Run's graceful-suspend contract.
func (a *Agent) RunStreamed(ctx context.Context, messages []llm.Message, opts ...RunOption) *StreamedRun {
	events := make(chan StreamEvent, 64)
	sr := &StreamedRun{
		Events: events,
		done:   make(chan struct{}),
	}

	go func() {
		defer close(sr.done)
		defer close(events)

		ro := runOpts{
			callLLM: streamingCallLLM(events),
			onEvent: func(ctx context.Context, ev StreamEvent) {
				trySendEvent(ctx, events, ev)
			},
		}
		for _, opt := range opts {
			opt(&ro)
		}

		result, err := coreLoop(ctx, a, messages, ro)

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
