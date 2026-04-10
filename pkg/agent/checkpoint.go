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
	CheckpointStatus string

	Checkpoint struct {
		Version       int              `json:"version"`
		Status        CheckpointStatus `json:"status"`
		AgentName     string           `json:"agent_name"`
		Messages      []llm.Message    `json:"messages"`
		Usage         llm.Usage        `json:"usage"`
		Turns         int              `json:"turns"`
		ToolUsedInRun bool             `json:"tool_used_in_run"`

		// Approval-interrupted checkpoints carry pending tool calls.
		PendingToolCalls []llm.ToolCall            `json:"pending_tool_calls,omitempty"`
		PendingApprovals []llm.ToolCall            `json:"pending_approvals,omitempty"`
		ApprovalInput    map[string]ApprovalResult `json:"approval_input,omitempty"`

		// Nested agent-as-tool suspension: one entry per suspended inner agent.
		AllToolCalls     []llm.ToolCall         `json:"all_tool_calls,omitempty"`
		InnerCheckpoints map[string]*Checkpoint `json:"inner_checkpoints,omitempty"`
		CompletedCalls   []CompletedCall        `json:"completed_calls,omitempty"`
	}

	CompletedCall struct {
		ToolCallID string     `json:"tool_call_id"`
		Result     ToolResult `json:"result"`
	}

	// CheckpointStore is supervisor-internal. Implementations may use raw
	// run IDs because public API/service methods perform tenant scoping and
	// authorization before a run reaches the supervisor.
	CheckpointStore interface {
		Save(ctx context.Context, runID string, cp *Checkpoint) error
		Load(ctx context.Context, runID string) (*Checkpoint, error)
		Delete(ctx context.Context, runID string) error
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
	CheckpointVersion  = 1
	MaxCheckpointBytes = 10 * 1024 * 1024

	CheckpointStatusSuspended        CheckpointStatus = "suspended"
	CheckpointStatusAwaitingApproval CheckpointStatus = "awaiting_approval"
)

func (e *SuspendedError) Error() string {
	return "agent run suspended"
}
