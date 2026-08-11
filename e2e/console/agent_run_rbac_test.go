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

func TestAgentRun_RBAC(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	runID := seedAgentRun(t, owner.GetOrganizationID(), agentRunSeed{
		agentName: "compliance-agent",
		status:    coredata.AgentRunStatusCompleted,
	})

	const getQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on AgentRun {
					id
					agentName
				}
			}
		}
	`

	roles := []testutil.TestRole{testutil.RoleAdmin, testutil.RoleViewer}
	for _, role := range roles {
		t.Run(string(role)+" can list and get agent runs", func(t *testing.T) {
			t.Parallel()
			member := testutil.NewClientInOrg(t, role, owner)

			var listResult agentRunConnectionResult

			err := member.Execute(agentRunListQuery, map[string]any{
				"orgId": member.GetOrganizationID().String(),
			}, &listResult)
			require.NoError(t, err)
			require.NotNil(t, listResult.Node, "organization node should resolve")
			assert.Equal(t, 1, listResult.Node.AgentRuns.TotalCount)

			var getResult struct {
				Node struct {
					ID        string `json:"id"`
					AgentName string `json:"agentName"`
				} `json:"node"`
			}

			err = member.Execute(getQuery, map[string]any{"id": runID.String()}, &getResult)
			require.NoError(t, err)
			assert.Equal(t, runID.String(), getResult.Node.ID)
			assert.Equal(t, "compliance-agent", getResult.Node.AgentName)
		})
	}
}
func TestAgentRun_SubmitApproval_RBAC(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	runID := seedAgentRun(t, owner.GetOrganizationID(), agentRunSeed{
		agentName:  "approval-agent",
		status:     coredata.AgentRunStatusAwaitingApproval,
		checkpoint: awaitingApprovalCheckpoint(t, "tc_1"),
	})

	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	var result submitAgentRunApprovalResult

	err := viewer.Execute(submitAgentRunApprovalMutation, map[string]any{
		"input": map[string]any{
			"agentRunId": runID.String(),
			"decisions": []map[string]any{
				{"toolCallId": "tc_1", "approved": true},
			},
		},
	}, &result)
	testutil.RequireForbiddenError(t, err, "viewer should not be able to approve agent runs")
}
