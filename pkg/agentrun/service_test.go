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

package agentrun_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agentrun"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func TestService_Get(t *testing.T) {
	client := test.PGClient(t)
	svc := agentrun.NewService(client)

	run := insertPendingRun(
		t,
		client,
		"service-get-agent",
		nil,
	)

	got, err := svc.Get(context.Background(), coredata.NewNoScope(), run.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, run.ID, got.ID)

	missingID := gid.New(run.ID.TenantID(), coredata.AgentRunEntityType)
	_, err = svc.Get(context.Background(), coredata.NewNoScope(), missingID)
	require.Error(t, err)
	assert.ErrorIs(t, err, coredata.ErrResourceNotFound)
}

func TestService_ListForOrganizationID(t *testing.T) {
	client := test.PGClient(t)
	svc := agentrun.NewService(client)

	orgID := insertTestOrganization(t, client)

	runA := insertPendingRunInOrg(t, client, orgID, "service-list-agent-a", nil)
	runB := insertPendingRunInOrg(t, client, orgID, "service-list-agent-b", nil)

	cursor := page.NewCursor(
		10,
		nil,
		page.Head,
		page.OrderBy[coredata.AgentRunOrderField]{
			Field:     coredata.AgentRunOrderFieldCreatedAt,
			Direction: page.OrderDirectionDesc,
		},
	)

	got, err := svc.ListForOrganizationID(context.Background(), coredata.NewNoScope(), orgID, cursor)
	require.NoError(t, err)
	require.NotNil(t, got)

	ids := make(map[gid.GID]bool)
	for _, run := range got.Data {
		ids[run.ID] = true
	}

	assert.True(t, ids[runA.ID])
	assert.True(t, ids[runB.ID])
}

func TestService_SubmitApproval_NotAwaitingApproval(t *testing.T) {
	client := test.PGClient(t)
	svc := agentrun.NewService(client)

	// A freshly inserted run is PENDING, not AWAITING_APPROVAL.
	run := insertPendingRun(t, client, "service-approval-agent", nil)

	_, err := svc.SubmitApproval(
		context.Background(),
		coredata.NewNoScope(),
		run.ID,
		map[string]agent.ApprovalResult{"tc_x": {Approved: true}},
	)
	require.Error(t, err)
	assert.ErrorIs(t, err, agentrun.ErrNotAwaitingApproval)
}

func TestService_CountForOrganizationID(t *testing.T) {
	client := test.PGClient(t)
	svc := agentrun.NewService(client)

	orgID := insertTestOrganization(t, client)
	_ = insertPendingRunInOrg(t, client, orgID, "service-count-agent-a", nil)
	_ = insertPendingRunInOrg(t, client, orgID, "service-count-agent-b", nil)
	_ = insertPendingRunInOrg(t, client, orgID, "service-count-agent-c", nil)

	count, err := svc.CountForOrganizationID(context.Background(), coredata.NewNoScope(), orgID)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}
