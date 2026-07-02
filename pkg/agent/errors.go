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
	"errors"
	"fmt"

	"go.probo.inc/probo/pkg/llm"
)

// ErrSuspendForCheckpoint is the cancel cause to use when the caller
// wants the agent loop to gracefully suspend.
var ErrSuspendForCheckpoint = errors.New("agent run graceful suspend requested")

type (
	MaxTurnsExceededError struct {
		MaxTurns int
	}

	MaxToolDepthExceededError struct {
		MaxDepth int
	}

	InputGuardrailTrippedError struct {
		Guardrail string
		Message   string
	}

	OutputGuardrailTrippedError struct {
		Guardrail string
		Message   string
	}

	InterruptedError struct {
		ToolCalls        []llm.ToolCall
		PendingApprovals []llm.ToolCall
		Agent            *Agent
		Messages         []llm.Message
		Usage            llm.Usage
		Turns            int

		outerState *outerLoopState
	}

	needsApprovalError struct {
		allToolCalls     []llm.ToolCall
		pendingApprovals []llm.ToolCall
	}

	nestedInterruptionError struct {
		inner          *InterruptedError
		toolCallID     string
		allToolCalls   []llm.ToolCall
		completedCalls []CompletedCall
	}

	outerLoopState struct {
		agent          *Agent
		messages       []llm.Message
		usage          llm.Usage
		turns          int
		allToolCalls   []llm.ToolCall
		toolCallID     string
		completedCalls []CompletedCall
		innerInterrupt *InterruptedError
	}
)

func (e *MaxTurnsExceededError) Error() string {
	return fmt.Sprintf("agent exceeded maximum number of turns (%d)", e.MaxTurns)
}

func (e *MaxToolDepthExceededError) Error() string {
	return fmt.Sprintf("agent-tool delegation exceeded maximum depth (%d)", e.MaxDepth)
}

func (e *InputGuardrailTrippedError) Error() string {
	return fmt.Sprintf("input guardrail %q tripped: %s", e.Guardrail, e.Message)
}

func (e *OutputGuardrailTrippedError) Error() string {
	return fmt.Sprintf("output guardrail %q tripped: %s", e.Guardrail, e.Message)
}

func (e *InterruptedError) Error() string {
	return fmt.Sprintf("run interrupted: %d tool call(s) require approval", len(e.PendingApprovals))
}

func (e *needsApprovalError) Error() string {
	return fmt.Sprintf("%d tool call(s) require approval", len(e.pendingApprovals))
}

func (e *nestedInterruptionError) Error() string {
	return fmt.Sprintf("nested agent interrupted: %s", e.inner.Error())
}
