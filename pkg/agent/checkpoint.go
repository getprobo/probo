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
	AgentStatus string

	// AgentConfig captures the subset of agent options that must remain
	// stable across a suspend/restore cycle to keep the run coherent.
	// Currently that is only MaxTurns, because Checkpoint.Turns is a
	// counter compared against it — if the live agent's bound were
	// lowered below the saved counter we would either short-circuit the
	// restored run or fail the warning at restoreSuspended. Other loop
	// bounds (maxEmptyOutputRetries, maxToolDepth) reset per turn and
	// are safe to change mid-suspension. Live references (tools,
	// handoffs, hooks, LLM client, approval callbacks, guardrails) are
	// intentionally not snapshotted so deploys can update behavior
	// while runs are paused.
	AgentConfig struct {
		MaxTurns int
	}

	Checkpoint struct {
		Status        AgentStatus
		AgentName     string
		Config        AgentConfig
		Messages      []llm.Message
		Usage         llm.Usage
		Turns         int
		ToolUsedInRun bool

		// Approval-interrupted checkpoints carry pending tool calls.
		PendingToolCalls []llm.ToolCall
		PendingApprovals []llm.ToolCall
		ApprovalInput    map[string]ApprovalResult // keyed by tool call ID

		// Nested agent-as-tool suspension: one entry per suspended inner agent.
		AllToolCalls     []llm.ToolCall
		InnerCheckpoints map[string]*Checkpoint
		CompletedCalls   []CompletedCall
	}

	CompletedCall struct {
		ToolCallID string
		Result     ToolResult
	}

	// Checkpointer is worker-internal. Implementations may use raw
	// run IDs because public API/service methods perform tenant scoping and
	// authorization before a run reaches the worker.
	Checkpointer interface {
		Save(ctx context.Context, runID string, cp *Checkpoint) error
		Load(ctx context.Context, runID string) (*Checkpoint, error)
	}

	AgentRegistry interface {
		Agent(name string) (*Agent, error)
	}

	SuspendedError struct {
		RunID      string      // Set when the outer loop has a store+runID (worker-managed).
		Checkpoint *Checkpoint // Set when returning from an inner agent-as-tool (no store).
	}
)

const (
	AgentStatusSuspended        AgentStatus = "suspended"
	AgentStatusAwaitingApproval AgentStatus = "awaiting_approval"
)

func (e *SuspendedError) Error() string {
	return "agent run suspended"
}
