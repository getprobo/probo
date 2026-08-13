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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/pkg/coredata"
)

func TestAgentExecution_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	runID := seedAgentExecution(t, org1Owner.GetOrganizationID(), agentExecutionSeed{
		agentName: "compliance-agent",
		status:    coredata.AgentExecutionStatusCompleted,
	})

	t.Run("other org cannot fetch the run by id", func(t *testing.T) {
		t.Parallel()

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on AgentExecution { id }
				}
			}
		`

		var result struct {
			Node *struct {
				ID string `json:"id"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{"id": runID.String()}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "AgentExecution")
	})

	t.Run("other org list does not include the run", func(t *testing.T) {
		t.Parallel()

		var result agentExecutionConnectionResult

		err := org2Owner.Execute(agentExecutionListQuery, map[string]any{
			"orgId": org2Owner.GetOrganizationID().String(),
		}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.Node, "organization node should resolve")

		assert.Equal(t, 0, result.Node.AgentExecutions.TotalCount)
		assert.Empty(t, result.Node.AgentExecutions.Edges)
	})
}
func TestAgentExecution_SubmitApproval_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	runID := seedAgentExecution(t, org1Owner.GetOrganizationID(), agentExecutionSeed{
		agentName:  "approval-agent",
		status:     coredata.AgentExecutionStatusAwaitingApproval,
		checkpoint: awaitingApprovalCheckpoint(t, "tc_1"),
	})

	var result submitAgentExecutionApprovalResult

	err := org2Owner.Execute(submitAgentExecutionApprovalMutation, map[string]any{
		"input": map[string]any{
			"agentExecutionId": runID.String(),
			"decisions": []map[string]any{
				{"toolCallId": "tc_1", "approved": true},
			},
		},
	}, &result)
	testutil.RequireForbiddenError(t, err, "other org should not be able to approve the run")
}
