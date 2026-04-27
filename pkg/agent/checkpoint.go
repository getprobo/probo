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

	"go.probo.inc/probo/pkg/llm"
)

// ErrSuspendForCheckpoint is the recommended cancel cause for callers
// who want the agent loop to gracefully suspend (build a checkpoint
// and return *SuspendedError) rather than treat the cancellation as a
// silent close. The agent loop only inspects ctx.Err(); any cancel
// cause produces a graceful suspend, but using this sentinel makes
// the intent explicit and lets supervisors distinguish a user-driven
// cancel from infrastructure-level causes (lease loss, heartbeat
// failure) when they inspect context.Cause(ctx).
var ErrSuspendForCheckpoint = errors.New("agent run: graceful suspend requested")

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

	// Checkpointer is supervisor-internal. Implementations may use raw
	// run IDs because public API/service methods perform tenant scoping and
	// authorization before a run reaches the supervisor.
	Checkpointer interface {
		Save(ctx context.Context, runID string, cp *Checkpoint) error
		Load(ctx context.Context, runID string) (*Checkpoint, error)
	}

	AgentRegistry interface {
		Agent(name string) (*Agent, error)
	}

	SuspendedError struct {
		RunID      string      // Set when the outer loop has a store+runID (supervisor-managed).
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
