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

package coredata_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agentruntest"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/llm"
)

func TestPGCheckpointer(t *testing.T) {
	t.Parallel()

	client := agentruntest.PGClient(t)
	store := coredata.NewPGCheckpointer(client)

	t.Run(
		"load returns nil when no checkpoint exists",
		func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			run := agentruntest.InsertPendingRun(
				t,
				client,
				"test-agent",
				[]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hello"}}}},
			)

			cp, err := store.Load(ctx, run.ID.String())
			require.NoError(t, err)
			assert.Nil(t, cp)
		},
	)

	t.Run(
		"save and load round-trip",
		func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			run := agentruntest.InsertPendingRun(
				t,
				client,
				"test-agent",
				[]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hello"}}}},
			)
			runID := run.ID.String()

			original := &agent.Checkpoint{
				Status:    agent.AgentStatusSuspended,
				AgentName: "test-agent",
				Messages: []llm.Message{
					{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hello"}}},
					{Role: llm.RoleAssistant, Parts: []llm.Part{llm.TextPart{Text: "working..."}}},
				},
				Usage:         llm.Usage{InputTokens: 20, OutputTokens: 10},
				Turns:         1,
				ToolUsedInRun: true,
			}

			err := store.Save(ctx, runID, original)
			require.NoError(t, err)

			loaded, err := store.Load(ctx, runID)
			require.NoError(t, err)
			require.NotNil(t, loaded)

			assert.Equal(t, original.Status, loaded.Status)
			assert.Equal(t, original.AgentName, loaded.AgentName)
			assert.Equal(t, original.Usage, loaded.Usage)
			assert.Equal(t, original.Turns, loaded.Turns)
			assert.Equal(t, original.ToolUsedInRun, loaded.ToolUsedInRun)
			assert.Len(t, loaded.Messages, 2)
		},
	)

	t.Run(
		"save overwrites previous checkpoint",
		func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			run := agentruntest.InsertPendingRun(
				t,
				client,
				"test-agent",
				[]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hello"}}}},
			)
			runID := run.ID.String()

			first := &agent.Checkpoint{
				Status:    agent.AgentStatusSuspended,
				AgentName: "test-agent",
				Turns:     1,
			}
			require.NoError(t, store.Save(ctx, runID, first))

			updated := &agent.Checkpoint{
				Status:    agent.AgentStatusSuspended,
				AgentName: "test-agent",
				Messages: []llm.Message{
					{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hello"}}},
					{Role: llm.RoleAssistant, Parts: []llm.Part{llm.TextPart{Text: "working..."}}},
					{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "continue"}}},
				},
				Usage: llm.Usage{InputTokens: 30, OutputTokens: 15},
				Turns: 2,
			}

			require.NoError(t, store.Save(ctx, runID, updated))

			loaded, err := store.Load(ctx, runID)
			require.NoError(t, err)
			require.NotNil(t, loaded)
			assert.Equal(t, 2, loaded.Turns)
			assert.Len(t, loaded.Messages, 3)
		},
	)

	t.Run(
		"save and load preserves approval state",
		func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			run := agentruntest.InsertPendingRun(
				t,
				client,
				"test-agent",
				[]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hello"}}}},
			)
			runID := run.ID.String()

			original := &agent.Checkpoint{
				Status:    agent.AgentStatusAwaitingApproval,
				AgentName: "test-agent",
				Turns:     3,
				PendingToolCalls: []llm.ToolCall{
					{ID: "call-1", Function: llm.FunctionCall{Name: "send_email", Arguments: `{"to":"a@b.c"}`}},
				},
				PendingApprovals: []llm.ToolCall{
					{ID: "call-1", Function: llm.FunctionCall{Name: "send_email", Arguments: `{"to":"a@b.c"}`}},
				},
				ApprovalInput: map[string]agent.ApprovalResult{
					"call-2": {Approved: true},
				},
				AllToolCalls: []llm.ToolCall{
					{ID: "call-1", Function: llm.FunctionCall{Name: "send_email"}},
					{ID: "call-2", Function: llm.FunctionCall{Name: "log_event"}},
				},
				InnerCheckpoints: map[string]*agent.Checkpoint{
					"call-3": {
						Status:    agent.AgentStatusSuspended,
						AgentName: "inner-agent",
						Turns:     1,
					},
				},
				CompletedCalls: []agent.CompletedCall{
					{ToolCallID: "call-2", Result: agent.ToolResult{Content: "ok"}},
				},
			}

			require.NoError(t, store.Save(ctx, runID, original))

			loaded, err := store.Load(ctx, runID)
			require.NoError(t, err)
			require.NotNil(t, loaded)
			assert.Len(t, loaded.PendingToolCalls, 1)
			assert.Len(t, loaded.PendingApprovals, 1)
			assert.True(t, loaded.ApprovalInput["call-2"].Approved)
			assert.Len(t, loaded.AllToolCalls, 2)
			require.Contains(t, loaded.InnerCheckpoints, "call-3")
			assert.Equal(t, "inner-agent", loaded.InnerCheckpoints["call-3"].AgentName)
			require.Len(t, loaded.CompletedCalls, 1)
			assert.Equal(t, "ok", loaded.CompletedCalls[0].Result.Content)
		},
	)

	t.Run(
		"save to nonexistent run returns error",
		func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			run := agentruntest.InsertPendingRun(
				t,
				client,
				"test-agent",
				[]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hello"}}}},
			)
			// Build a syntactically valid GID for the same tenant but a
			// different (unknown) entity so the tenant-scope check does
			// not short-circuit with a parse error.
			otherID := gid.New(run.ID.TenantID(), coredata.AgentRunEntityType)

			cp := &agent.Checkpoint{
				Status:    agent.AgentStatusSuspended,
				AgentName: "test-agent",
			}
			err := store.Save(ctx, otherID.String(), cp)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not found")
		},
	)
}
