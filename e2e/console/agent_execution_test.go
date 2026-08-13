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

package console_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/llm"
)

// agentExecutionSeed describes the agent execution row inserted directly into the test
// database. Agent executions have no creation mutation on the console API (they are
// produced by the agent execution worker), so e2e coverage seeds them straight into
// Postgres, mirroring the common-third-party catalog seeding helper.
type agentExecutionSeed struct {
	agentName    string
	status       coredata.AgentExecutionStatus
	errorMessage *string
	startedAt    *time.Time
	createdAt    time.Time
	checkpoint   []byte
}

func seedAgentExecution(t *testing.T, organizationID gid.GID, seed agentExecutionSeed) gid.GID {
	t.Helper()

	ctx := context.Background()
	conn := dialTestPg(t, ctx)
	t.Cleanup(func() { _ = conn.Close(ctx) })

	if seed.agentName == "" {
		seed.agentName = "test-agent"
	}

	if seed.status == "" {
		seed.status = coredata.AgentExecutionStatusPending
	}

	if seed.createdAt.IsZero() {
		seed.createdAt = time.Now().UTC()
	}

	id := gid.New(organizationID.TenantID(), coredata.AgentExecutionEntityType)

	var checkpoint any
	if len(seed.checkpoint) > 0 {
		checkpoint = string(seed.checkpoint)
	}

	_, err := conn.Exec(ctx, `
		INSERT INTO agent_executions (
			id, tenant_id, organization_id, start_agent_name, status,
			input_messages, checkpoint, error_message, started_at,
			session_messages, processing_input_ids,
			attempt_count, max_attempts,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6::jsonb, $7::jsonb, $8, $9,
			$6::jsonb, $10, $11, $12, $13, $13
		)
	`,
		id,
		organizationID.TenantID(),
		organizationID,
		seed.agentName,
		seed.status,
		"[]",
		checkpoint,
		seed.errorMessage,
		seed.startedAt,
		[]string{},
		0,
		1,
		seed.createdAt,
	)
	require.NoError(t, err, "cannot seed agent execution")

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cleanupConn := dialTestPg(t, cleanupCtx)

		defer func() { _ = cleanupConn.Close(cleanupCtx) }()

		_, err := cleanupConn.Exec(cleanupCtx, `DELETE FROM agent_executions WHERE id = $1`, id)
		assert.NoError(t, err, "cleanup: cannot delete seeded agent execution %s", id)
	})

	return id
}

const agentExecutionListQuery = `
	query($orgId: ID!, $orderBy: AgentExecutionOrder) {
		node(id: $orgId) {
			... on Organization {
				agentExecutions(first: 50, orderBy: $orderBy) {
					totalCount
					edges {
						cursor
						node {
							id
							agentName
							status
							errorMessage
							startedAt
							createdAt
							updatedAt
						}
					}
					pageInfo {
						hasNextPage
						hasPreviousPage
						startCursor
						endCursor
					}
				}
			}
		}
	}
`

type agentExecutionNode struct {
	ID           string  `json:"id"`
	AgentName    string  `json:"agentName"`
	Status       string  `json:"status"`
	ErrorMessage *string `json:"errorMessage"`
	StartedAt    *string `json:"startedAt"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

type agentExecutionConnectionResult struct {
	Node *struct {
		AgentExecutions struct {
			TotalCount int `json:"totalCount"`
			Edges      []struct {
				Cursor string             `json:"cursor"`
				Node   agentExecutionNode `json:"node"`
			} `json:"edges"`
			PageInfo testutil.PageInfo `json:"pageInfo"`
		} `json:"agentExecutions"`
	} `json:"node"`
}

func TestAgentExecution_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	errMsg := "boom"
	startedAt := time.Now().UTC().Add(-time.Minute)

	completedID := seedAgentExecution(t, owner.GetOrganizationID(), agentExecutionSeed{
		agentName: "compliance-agent",
		status:    coredata.AgentExecutionStatusCompleted,
		startedAt: &startedAt,
	})
	failedID := seedAgentExecution(t, owner.GetOrganizationID(), agentExecutionSeed{
		agentName:    "vetting-agent",
		status:       coredata.AgentExecutionStatusFailed,
		errorMessage: &errMsg,
		startedAt:    &startedAt,
	})

	var result agentExecutionConnectionResult

	err := owner.Execute(agentExecutionListQuery, map[string]any{
		"orgId": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	require.NotNil(t, result.Node, "organization node should resolve")

	assert.Equal(t, 2, result.Node.AgentExecutions.TotalCount)
	require.Len(t, result.Node.AgentExecutions.Edges, 2)

	byID := make(map[string]agentExecutionNode, 2)

	for _, edge := range result.Node.AgentExecutions.Edges {
		assert.NotEmpty(t, edge.Cursor, "edge cursor should be set")
		byID[edge.Node.ID] = edge.Node
	}

	completed, ok := byID[completedID.String()]
	require.True(t, ok, "completed run not returned in list")
	assert.Equal(t, "compliance-agent", completed.AgentName)
	assert.Equal(t, "COMPLETED", completed.Status)
	assert.Nil(t, completed.ErrorMessage)
	assert.NotNil(t, completed.StartedAt)
	assert.NotEmpty(t, completed.CreatedAt)
	assert.NotEmpty(t, completed.UpdatedAt)

	failed, ok := byID[failedID.String()]
	require.True(t, ok, "failed run not returned in list")
	assert.Equal(t, "vetting-agent", failed.AgentName)
	assert.Equal(t, "FAILED", failed.Status)
	require.NotNil(t, failed.ErrorMessage)
	assert.Equal(t, "boom", *failed.ErrorMessage)
}

func TestAgentExecution_ListEmpty(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	var result agentExecutionConnectionResult

	err := owner.Execute(agentExecutionListQuery, map[string]any{
		"orgId": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	require.NotNil(t, result.Node, "organization node should resolve")

	assert.Equal(t, 0, result.Node.AgentExecutions.TotalCount)
	assert.Empty(t, result.Node.AgentExecutions.Edges)
	assert.False(t, result.Node.AgentExecutions.PageInfo.HasNextPage)
	assert.False(t, result.Node.AgentExecutions.PageInfo.HasPreviousPage)
}

func TestAgentExecution_Ordering(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	base := time.Now().UTC().Add(-time.Hour)
	oldestID := seedAgentExecution(t, owner.GetOrganizationID(), agentExecutionSeed{createdAt: base})
	middleID := seedAgentExecution(t, owner.GetOrganizationID(), agentExecutionSeed{createdAt: base.Add(time.Minute)})
	newestID := seedAgentExecution(t, owner.GetOrganizationID(), agentExecutionSeed{createdAt: base.Add(2 * time.Minute)})

	t.Run("ascending by createdAt", func(t *testing.T) {
		t.Parallel()

		var result agentExecutionConnectionResult

		err := owner.Execute(agentExecutionListQuery, map[string]any{
			"orgId":   owner.GetOrganizationID().String(),
			"orderBy": map[string]any{"direction": "ASC", "field": "CREATED_AT"},
		}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.Node, "organization node should resolve")
		require.Len(t, result.Node.AgentExecutions.Edges, 3)

		assert.Equal(t, oldestID.String(), result.Node.AgentExecutions.Edges[0].Node.ID)
		assert.Equal(t, middleID.String(), result.Node.AgentExecutions.Edges[1].Node.ID)
		assert.Equal(t, newestID.String(), result.Node.AgentExecutions.Edges[2].Node.ID)
	})

	t.Run("descending by createdAt", func(t *testing.T) {
		t.Parallel()

		var result agentExecutionConnectionResult

		err := owner.Execute(agentExecutionListQuery, map[string]any{
			"orgId":   owner.GetOrganizationID().String(),
			"orderBy": map[string]any{"direction": "DESC", "field": "CREATED_AT"},
		}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.Node, "organization node should resolve")
		require.Len(t, result.Node.AgentExecutions.Edges, 3)

		assert.Equal(t, newestID.String(), result.Node.AgentExecutions.Edges[0].Node.ID)
		assert.Equal(t, middleID.String(), result.Node.AgentExecutions.Edges[1].Node.ID)
		assert.Equal(t, oldestID.String(), result.Node.AgentExecutions.Edges[2].Node.ID)
	})
}

func TestAgentExecution_Pagination(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	base := time.Now().UTC().Add(-time.Hour)
	for i := range 3 {
		seedAgentExecution(t, owner.GetOrganizationID(), agentExecutionSeed{
			createdAt: base.Add(time.Duration(i) * time.Minute),
		})
	}

	const query = `
		query($orgId: ID!, $first: Int, $after: CursorKey) {
			node(id: $orgId) {
				... on Organization {
					agentExecutions(first: $first, after: $after, orderBy: {direction: ASC, field: CREATED_AT}) {
						totalCount
						edges {
							cursor
							node { id }
						}
						pageInfo {
							hasNextPage
							hasPreviousPage
							startCursor
							endCursor
						}
					}
				}
			}
		}
	`

	var firstPage agentExecutionConnectionResult

	err := owner.Execute(query, map[string]any{
		"orgId": owner.GetOrganizationID().String(),
		"first": 2,
	}, &firstPage)
	require.NoError(t, err)
	require.NotNil(t, firstPage.Node, "organization node should resolve")

	assert.Equal(t, 3, firstPage.Node.AgentExecutions.TotalCount)
	testutil.AssertFirstPage(t, len(firstPage.Node.AgentExecutions.Edges), firstPage.Node.AgentExecutions.PageInfo, 2, true)
	require.NotNil(t, firstPage.Node.AgentExecutions.PageInfo.EndCursor)

	var secondPage agentExecutionConnectionResult

	err = owner.Execute(query, map[string]any{
		"orgId": owner.GetOrganizationID().String(),
		"first": 2,
		"after": *firstPage.Node.AgentExecutions.PageInfo.EndCursor,
	}, &secondPage)
	require.NoError(t, err)
	require.NotNil(t, secondPage.Node, "organization node should resolve")

	testutil.AssertLastPage(t, len(secondPage.Node.AgentExecutions.Edges), secondPage.Node.AgentExecutions.PageInfo, 1, true)

	// The page boundary must not overlap.
	firstIDs := map[string]struct{}{}
	for _, edge := range firstPage.Node.AgentExecutions.Edges {
		firstIDs[edge.Node.ID] = struct{}{}
	}

	for _, edge := range secondPage.Node.AgentExecutions.Edges {
		_, overlap := firstIDs[edge.Node.ID]
		assert.False(t, overlap, "second page must not repeat a first-page run")
	}
}

func TestAgentExecution_Get(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	runID := seedAgentExecution(t, owner.GetOrganizationID(), agentExecutionSeed{
		agentName: "compliance-agent",
		status:    coredata.AgentExecutionStatusRunning,
	})

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on AgentExecution {
					id
					agentName
					status
					errorMessage
					startedAt
					createdAt
					updatedAt
					organization { id }
					permission(action: "agent:execution:get")
				}
			}
		}
	`

	var result struct {
		Node struct {
			ID           string  `json:"id"`
			AgentName    string  `json:"agentName"`
			Status       string  `json:"status"`
			ErrorMessage *string `json:"errorMessage"`
			StartedAt    *string `json:"startedAt"`
			CreatedAt    string  `json:"createdAt"`
			UpdatedAt    string  `json:"updatedAt"`
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
			Permission bool `json:"permission"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{"id": runID.String()}, &result)
	require.NoError(t, err)

	assert.Equal(t, runID.String(), result.Node.ID)
	assert.Equal(t, "compliance-agent", result.Node.AgentName)
	assert.Equal(t, "RUNNING", result.Node.Status)
	assert.Nil(t, result.Node.ErrorMessage)
	assert.Equal(t, owner.GetOrganizationID().String(), result.Node.Organization.ID)
	assert.True(t, result.Node.Permission, "owner should have agent-execution:get permission")
}

// awaitingApprovalCheckpoint builds the JSON checkpoint a worker persists
// when a run pauses for approval, carrying the pending tool-call IDs the
// approval mutation must reconcile against.
func awaitingApprovalCheckpoint(t *testing.T, toolCallIDs ...string) []byte {
	t.Helper()

	approvals := make([]llm.ToolCall, len(toolCallIDs))
	for i, id := range toolCallIDs {
		approvals[i] = llm.ToolCall{
			ID:       id,
			Function: llm.FunctionCall{Name: "danger", Arguments: "{}"},
		}
	}

	cp := agent.Checkpoint{
		Status:    agent.AgentStatusAwaitingApproval,
		AgentName: "approval-agent",
		Messages: []llm.Message{
			{Role: llm.RoleAssistant, ToolCalls: approvals},
		},
		PendingToolCalls: approvals,
		PendingApprovals: approvals,
	}

	data, err := json.Marshal(&cp)
	require.NoError(t, err, "cannot marshal approval checkpoint")

	return data
}

const submitAgentExecutionApprovalMutation = `
	mutation($input: SubmitAgentExecutionApprovalInput!) {
		submitAgentExecutionApproval(input: $input) {
			agentExecution {
				id
				status
			}
		}
	}
`

type submitAgentExecutionApprovalResult struct {
	SubmitAgentExecutionApproval struct {
		AgentExecution struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"agentExecution"`
	} `json:"submitAgentExecutionApproval"`
}

func TestAgentExecution_SubmitApproval(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	runID := seedAgentExecution(t, owner.GetOrganizationID(), agentExecutionSeed{
		agentName:  "approval-agent",
		status:     coredata.AgentExecutionStatusAwaitingApproval,
		checkpoint: awaitingApprovalCheckpoint(t, "tc_1"),
	})

	var result submitAgentExecutionApprovalResult

	err := owner.Execute(submitAgentExecutionApprovalMutation, map[string]any{
		"input": map[string]any{
			"agentExecutionId": runID.String(),
			"decisions": []map[string]any{
				{"toolCallId": "tc_1", "approved": true},
			},
		},
	}, &result)
	require.NoError(t, err)

	// A submitted decision requeues the run so a worker resumes it.
	assert.Equal(t, runID.String(), result.SubmitAgentExecutionApproval.AgentExecution.ID)
	assert.Equal(t, "PENDING", result.SubmitAgentExecutionApproval.AgentExecution.Status)
}

func TestAgentExecution_SubmitApproval_NotAwaiting(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	runID := seedAgentExecution(t, owner.GetOrganizationID(), agentExecutionSeed{
		agentName: "approval-agent",
		status:    coredata.AgentExecutionStatusCompleted,
	})

	var result submitAgentExecutionApprovalResult

	err := owner.Execute(submitAgentExecutionApprovalMutation, map[string]any{
		"input": map[string]any{
			"agentExecutionId": runID.String(),
			"decisions": []map[string]any{
				{"toolCallId": "tc_1", "approved": true},
			},
		},
	}, &result)
	testutil.RequireErrorCode(t, err, "CONFLICT")
}

func TestAgentExecution_SubmitApproval_IncompleteDecisions(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Two pending approvals, but only one decision is supplied.
	runID := seedAgentExecution(t, owner.GetOrganizationID(), agentExecutionSeed{
		agentName:  "approval-agent",
		status:     coredata.AgentExecutionStatusAwaitingApproval,
		checkpoint: awaitingApprovalCheckpoint(t, "tc_1", "tc_2"),
	})

	var result submitAgentExecutionApprovalResult

	err := owner.Execute(submitAgentExecutionApprovalMutation, map[string]any{
		"input": map[string]any{
			"agentExecutionId": runID.String(),
			"decisions": []map[string]any{
				{"toolCallId": "tc_1", "approved": true},
			},
		},
	}, &result)
	testutil.RequireErrorCode(t, err, "INVALID")
}
