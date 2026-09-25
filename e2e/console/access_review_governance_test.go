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
)

func TestAccessReviewGovernance_FlagEntry(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	fixture := startAccessReviewGovernanceCampaign(
		t,
		owner,
		"governance flag",
		accessReviewGovernanceCSVMultiRow,
	)
	require.NotEmpty(t, fixture.Entries)

	entryID := fixture.Entries[0].ID
	flags := []string{"PRIVILEGED_ACCESS", "ROLE_MISMATCH", "EXCESSIVE"}
	reasons := []string{
		"Admin role without ticket",
		"Title does not match role",
		"Too many SaaS seats",
	}

	flagged := flagAccessReviewEntryLeaf(t, owner, entryID, flags, reasons)
	assert.ElementsMatch(t, flags, flagged.Flags)
	assert.Equal(t, reasons, flagged.FlagReasons)
	assert.True(t, flagged.CanDecide)
	assert.True(t, flagged.CanFlag)
	assert.Equal(t, fixture.CampaignSourceID, flagged.CampaignSource.ID)

	roundtrip := queryAccessReviewEntryGovernance(t, owner, entryID)
	assert.ElementsMatch(t, flags, roundtrip.Flags)
	assert.Equal(t, reasons, roundtrip.FlagReasons)
	assert.Equal(t, "PENDING", roundtrip.Decision)

	stats := queryAccessReviewCampaignStatistics(t, owner, fixture.CampaignID)
	require.NotEmpty(t, stats.FlagCounts)

	found := false

	for _, fc := range stats.FlagCounts {
		if fc.Flag == "PRIVILEGED_ACCESS" && fc.Count >= 1 {
			found = true

			break
		}
	}

	assert.True(t, found, "flagCounts should include PRIVILEGED_ACCESS after flagging")
}

func TestAccessReviewGovernance_BulkDecisionsAndClose(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	fixture := startAccessReviewGovernanceCampaign(
		t,
		owner,
		"governance bulk decide",
		accessReviewGovernanceCSVMultiRow,
	)
	require.Len(t, fixture.Entries, 3)

	entryByEmail := map[string]accessReviewGovernanceEntry{}
	for _, entry := range fixture.Entries {
		entryByEmail[entry.Email] = entry
	}

	decisions := []map[string]any{
		{
			"accessReviewEntryId": entryByEmail["jane@example.com"].ID,
			"decision":            "APPROVED",
			"decisionNote":        "Still needs CTO access",
		},
		{
			"accessReviewEntryId": entryByEmail["bob@example.com"].ID,
			"decision":            "REVOKE",
			"decisionNote":        "Offboarded from project",
		},
		{
			"accessReviewEntryId": entryByEmail["alice@example.com"].ID,
			"decision":            "DEFER",
			"decisionNote":        "Waiting on manager sign-off",
		},
	}

	updated := recordAccessReviewEntryDecisionsLeaf(t, owner, decisions)
	require.Len(t, updated, 3)

	stats := queryAccessReviewCampaignStatistics(t, owner, fixture.CampaignID)
	assert.Equal(t, 3, stats.TotalCount)
	assert.Equal(t, 1, decisionCount(stats, "APPROVED"))
	assert.Equal(t, 1, decisionCount(stats, "REVOKE"))
	assert.Equal(t, 1, decisionCount(stats, "DEFER"))
	assert.Equal(t, 0, decisionCount(stats, "PENDING"))

	closeAccessReviewCampaignLeaf(t, owner, fixture.CampaignID)
}

func TestAccessReviewGovernance_CampaignSourceGraph(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	fixture := startAccessReviewGovernanceCampaign(
		t,
		owner,
		"governance graph",
		accessReviewGovernanceCSVMultiRow,
	)

	const sourceQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on AccessReviewCampaignSource {
					id
					name
					campaign {
						id
						canClose: permission(action: "access-review:campaign:close")
						statistics {
							totalCount
						}
					}
					statistics {
						totalCount
						decisionCounts { decision count }
					}
					fetchAttempts(first: 5) {
						totalCount
						edges {
							node {
								status
								fetchedAccountsCount
								error
							}
						}
					}
					entries(first: 10) {
						totalCount
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Campaign struct {
				ID         string `json:"id"`
				CanClose   bool   `json:"canClose"`
				Statistics struct {
					TotalCount int `json:"totalCount"`
				} `json:"statistics"`
			} `json:"campaign"`
			Statistics struct {
				TotalCount     int `json:"totalCount"`
				DecisionCounts []struct {
					Decision string `json:"decision"`
					Count    int    `json:"count"`
				} `json:"decisionCounts"`
			} `json:"statistics"`
			FetchAttempts struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						Status               string  `json:"status"`
						FetchedAccountsCount int     `json:"fetchedAccountsCount"`
						Error                *string `json:"error"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"fetchAttempts"`
			Entries struct {
				TotalCount int `json:"totalCount"`
			} `json:"entries"`
		} `json:"node"`
	}

	err := owner.Execute(sourceQuery, map[string]any{"id": fixture.CampaignSourceID}, &result)
	require.NoError(t, err)

	node := result.Node
	assert.Equal(t, fixture.CampaignSourceID, node.ID)
	assert.Equal(t, fixture.CampaignID, node.Campaign.ID)
	assert.Contains(t, node.Name, "governance graph")
	assert.True(t, node.Campaign.CanClose)
	assert.Equal(t, len(fixture.Entries), node.Campaign.Statistics.TotalCount)
	assert.Equal(t, len(fixture.Entries), node.Statistics.TotalCount)
	assert.Equal(t, len(fixture.Entries), node.Entries.TotalCount)
	assert.GreaterOrEqual(t, node.FetchAttempts.TotalCount, 1)
	require.NotEmpty(t, node.FetchAttempts.Edges)
	assert.Equal(t, "SUCCESS", node.FetchAttempts.Edges[0].Node.Status)
	assert.Greater(t, node.FetchAttempts.Edges[0].Node.FetchedAccountsCount, 0)
	assert.Nil(t, node.FetchAttempts.Edges[0].Node.Error)

	entry := queryAccessReviewEntryGovernance(t, owner, fixture.Entries[0].ID)
	assert.Equal(t, fixture.CampaignSourceID, entry.CampaignSource.ID)
	assert.True(t, entry.CanDecide)
	assert.True(t, entry.CanFlag)
}
