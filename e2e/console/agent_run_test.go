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

package console_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// agentRunSeed describes the agent run row inserted directly into the test
// database. Agent runs have no creation mutation on the console API (they are
// produced by the agent run worker), so e2e coverage seeds them straight into
// Postgres, mirroring the common-third-party catalog seeding helper.
type agentRunSeed struct {
	agentName    string
	status       coredata.AgentRunStatus
	errorMessage *string
	startedAt    *time.Time
	createdAt    time.Time
}

func seedAgentRun(t *testing.T, organizationID gid.GID, seed agentRunSeed) gid.GID {
	t.Helper()

	ctx := context.Background()
	conn := dialTestPg(t, ctx)
	t.Cleanup(func() { _ = conn.Close(ctx) })

	if seed.agentName == "" {
		seed.agentName = "test-agent"
	}

	if seed.status == "" {
		seed.status = coredata.AgentRunStatusPending
	}

	if seed.createdAt.IsZero() {
		seed.createdAt = time.Now().UTC()
	}

	id := gid.New(organizationID.TenantID(), coredata.AgentRunEntityType)

	_, err := conn.Exec(ctx, `
		INSERT INTO agent_runs (
			id, tenant_id, organization_id, start_agent_name, status,
			input_messages, error_message, started_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $9
		)
	`,
		id,
		organizationID.TenantID(),
		organizationID,
		seed.agentName,
		seed.status,
		[]byte(`[]`),
		seed.errorMessage,
		seed.startedAt,
		seed.createdAt,
	)
	require.NoError(t, err, "cannot seed agent run")

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cleanupConn := dialTestPg(t, cleanupCtx)

		defer func() { _ = cleanupConn.Close(cleanupCtx) }()

		_, err := cleanupConn.Exec(cleanupCtx, `DELETE FROM agent_runs WHERE id = $1`, id)
		assert.NoError(t, err, "cleanup: cannot delete seeded agent run %s", id)
	})

	return id
}

const agentRunListQuery = `
	query($orgId: ID!, $orderBy: AgentRunOrder) {
		node(id: $orgId) {
			... on Organization {
				agentRuns(first: 50, orderBy: $orderBy) {
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

type agentRunNode struct {
	ID           string  `json:"id"`
	AgentName    string  `json:"agentName"`
	Status       string  `json:"status"`
	ErrorMessage *string `json:"errorMessage"`
	StartedAt    *string `json:"startedAt"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

type agentRunConnectionResult struct {
	Node struct {
		AgentRuns struct {
			TotalCount int `json:"totalCount"`
			Edges      []struct {
				Cursor string       `json:"cursor"`
				Node   agentRunNode `json:"node"`
			} `json:"edges"`
			PageInfo testutil.PageInfo `json:"pageInfo"`
		} `json:"agentRuns"`
	} `json:"node"`
}

func TestAgentRun_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	errMsg := "boom"
	startedAt := time.Now().UTC().Add(-time.Minute)

	completedID := seedAgentRun(t, owner.GetOrganizationID(), agentRunSeed{
		agentName: "compliance-agent",
		status:    coredata.AgentRunStatusCompleted,
		startedAt: &startedAt,
	})
	failedID := seedAgentRun(t, owner.GetOrganizationID(), agentRunSeed{
		agentName:    "vetting-agent",
		status:       coredata.AgentRunStatusFailed,
		errorMessage: &errMsg,
		startedAt:    &startedAt,
	})

	var result agentRunConnectionResult

	err := owner.Execute(agentRunListQuery, map[string]any{
		"orgId": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, 2, result.Node.AgentRuns.TotalCount)
	require.Len(t, result.Node.AgentRuns.Edges, 2)

	byID := make(map[string]agentRunNode, 2)

	for _, edge := range result.Node.AgentRuns.Edges {
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

func TestAgentRun_ListEmpty(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	var result agentRunConnectionResult

	err := owner.Execute(agentRunListQuery, map[string]any{
		"orgId": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, 0, result.Node.AgentRuns.TotalCount)
	assert.Empty(t, result.Node.AgentRuns.Edges)
	assert.False(t, result.Node.AgentRuns.PageInfo.HasNextPage)
	assert.False(t, result.Node.AgentRuns.PageInfo.HasPreviousPage)
}

func TestAgentRun_Ordering(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	base := time.Now().UTC().Add(-time.Hour)
	oldestID := seedAgentRun(t, owner.GetOrganizationID(), agentRunSeed{createdAt: base})
	middleID := seedAgentRun(t, owner.GetOrganizationID(), agentRunSeed{createdAt: base.Add(time.Minute)})
	newestID := seedAgentRun(t, owner.GetOrganizationID(), agentRunSeed{createdAt: base.Add(2 * time.Minute)})

	t.Run("ascending by createdAt", func(t *testing.T) {
		t.Parallel()

		var result agentRunConnectionResult

		err := owner.Execute(agentRunListQuery, map[string]any{
			"orgId":   owner.GetOrganizationID().String(),
			"orderBy": map[string]any{"direction": "ASC", "field": "CREATED_AT"},
		}, &result)
		require.NoError(t, err)
		require.Len(t, result.Node.AgentRuns.Edges, 3)

		assert.Equal(t, oldestID.String(), result.Node.AgentRuns.Edges[0].Node.ID)
		assert.Equal(t, middleID.String(), result.Node.AgentRuns.Edges[1].Node.ID)
		assert.Equal(t, newestID.String(), result.Node.AgentRuns.Edges[2].Node.ID)
	})

	t.Run("descending by createdAt", func(t *testing.T) {
		t.Parallel()

		var result agentRunConnectionResult

		err := owner.Execute(agentRunListQuery, map[string]any{
			"orgId":   owner.GetOrganizationID().String(),
			"orderBy": map[string]any{"direction": "DESC", "field": "CREATED_AT"},
		}, &result)
		require.NoError(t, err)
		require.Len(t, result.Node.AgentRuns.Edges, 3)

		assert.Equal(t, newestID.String(), result.Node.AgentRuns.Edges[0].Node.ID)
		assert.Equal(t, middleID.String(), result.Node.AgentRuns.Edges[1].Node.ID)
		assert.Equal(t, oldestID.String(), result.Node.AgentRuns.Edges[2].Node.ID)
	})
}

func TestAgentRun_Pagination(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	base := time.Now().UTC().Add(-time.Hour)
	for i := range 3 {
		seedAgentRun(t, owner.GetOrganizationID(), agentRunSeed{
			createdAt: base.Add(time.Duration(i) * time.Minute),
		})
	}

	const query = `
		query($orgId: ID!, $first: Int, $after: CursorKey) {
			node(id: $orgId) {
				... on Organization {
					agentRuns(first: $first, after: $after, orderBy: {direction: ASC, field: CREATED_AT}) {
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

	var firstPage agentRunConnectionResult

	err := owner.Execute(query, map[string]any{
		"orgId": owner.GetOrganizationID().String(),
		"first": 2,
	}, &firstPage)
	require.NoError(t, err)

	assert.Equal(t, 3, firstPage.Node.AgentRuns.TotalCount)
	testutil.AssertFirstPage(t, len(firstPage.Node.AgentRuns.Edges), firstPage.Node.AgentRuns.PageInfo, 2, true)
	require.NotNil(t, firstPage.Node.AgentRuns.PageInfo.EndCursor)

	var secondPage agentRunConnectionResult

	err = owner.Execute(query, map[string]any{
		"orgId": owner.GetOrganizationID().String(),
		"first": 2,
		"after": *firstPage.Node.AgentRuns.PageInfo.EndCursor,
	}, &secondPage)
	require.NoError(t, err)

	testutil.AssertLastPage(t, len(secondPage.Node.AgentRuns.Edges), secondPage.Node.AgentRuns.PageInfo, 1, true)

	// The page boundary must not overlap.
	firstIDs := map[string]struct{}{}
	for _, edge := range firstPage.Node.AgentRuns.Edges {
		firstIDs[edge.Node.ID] = struct{}{}
	}

	for _, edge := range secondPage.Node.AgentRuns.Edges {
		_, overlap := firstIDs[edge.Node.ID]
		assert.False(t, overlap, "second page must not repeat a first-page run")
	}
}

func TestAgentRun_Get(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	runID := seedAgentRun(t, owner.GetOrganizationID(), agentRunSeed{
		agentName: "compliance-agent",
		status:    coredata.AgentRunStatusRunning,
	})

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on AgentRun {
					id
					agentName
					status
					errorMessage
					startedAt
					createdAt
					updatedAt
					organization { id }
					permission(action: "core:agent-run:get")
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
	assert.True(t, result.Node.Permission, "owner should have agent-run:get permission")
}

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

func TestAgentRun_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	runID := seedAgentRun(t, org1Owner.GetOrganizationID(), agentRunSeed{
		agentName: "compliance-agent",
		status:    coredata.AgentRunStatusCompleted,
	})

	t.Run("other org cannot fetch the run by id", func(t *testing.T) {
		t.Parallel()

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on AgentRun { id }
				}
			}
		`

		var result struct {
			Node *struct {
				ID string `json:"id"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{"id": runID.String()}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "AgentRun")
	})

	t.Run("other org list does not include the run", func(t *testing.T) {
		t.Parallel()

		var result agentRunConnectionResult

		err := org2Owner.Execute(agentRunListQuery, map[string]any{
			"orgId": org2Owner.GetOrganizationID().String(),
		}, &result)
		require.NoError(t, err)

		assert.Equal(t, 0, result.Node.AgentRuns.TotalCount)
		assert.Empty(t, result.Node.AgentRuns.Edges)
	})
}
