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

package agentrun_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/agentrun"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func TestService_Get(t *testing.T) {
	client := pgClient(t)
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
	client := pgClient(t)
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

func TestService_CountForOrganizationID(t *testing.T) {
	client := pgClient(t)
	svc := agentrun.NewService(client)

	orgID := insertTestOrganization(t, client)
	_ = insertPendingRunInOrg(t, client, orgID, "service-count-agent-a", nil)
	_ = insertPendingRunInOrg(t, client, orgID, "service-count-agent-b", nil)
	_ = insertPendingRunInOrg(t, client, orgID, "service-count-agent-c", nil)

	count, err := svc.CountForOrganizationID(context.Background(), coredata.NewNoScope(), orgID)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}
