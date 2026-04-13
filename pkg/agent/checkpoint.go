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
	AgentStatus string

	Checkpoint struct {
		Status        AgentStatus
		AgentName     string
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
	MaxCheckpointBytes = 10 * 1024 * 1024

	AgentStatusSuspended        AgentStatus = "suspended"
	AgentStatusAwaitingApproval AgentStatus = "awaiting_approval"
)

func (e *SuspendedError) Error() string {
	return "agent run suspended"
}
