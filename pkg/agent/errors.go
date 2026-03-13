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
	"fmt"

	"go.probo.inc/probo/pkg/llm"
)

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
		completedCalls []completedCall
	}

	outerLoopState struct {
		agent          *Agent
		messages       []llm.Message
		usage          llm.Usage
		turns          int
		allToolCalls   []llm.ToolCall
		toolCallID     string
		completedCalls []completedCall
		innerInterrupt *InterruptedError
	}

	completedCall struct {
		toolCallID string
		result     ToolResult
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
