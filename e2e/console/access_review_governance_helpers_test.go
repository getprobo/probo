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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const (
	accessReviewGovernanceCSVHeader = "email,full_name,role,job_title,is_admin,active,mfa_status,auth_method,last_login,account_created_at,external_id"

	accessReviewGovernanceCSVMultiRow = accessReviewGovernanceCSVHeader + `
jane@example.com,Jane Smith,admin,CTO,true,true,ENABLED,SSO,2026-01-15T00:00:00Z,2024-06-01T00:00:00Z,ext-jane
bob@example.com,Bob Jones,viewer,Engineer,false,true,ENABLED,SSO,2026-01-14T00:00:00Z,2024-05-01T00:00:00Z,ext-bob
alice@example.com,Alice Lee,admin,Manager,true,true,DISABLED,PASSWORD,2026-01-10T00:00:00Z,2024-01-01T00:00:00Z,ext-alice`

	accessReviewCampaignWorkerPollTimeout = 60 * time.Second
	accessReviewCampaignWorkerPollTick    = time.Second
	accessReviewGovernanceExpectedEntries = 3
)

const accessReviewEntryGovernanceSelection = `
	id
	email
	flags
	flagReasons
	decision
	decisionNote
	campaignSource {
		id
		name
	}
	canDecide: permission(action: "access-review:entry:decide")
	canFlag: permission(action: "access-review:entry:flag")
`

type (
	accessReviewGovernanceEntry struct {
		ID             string   `json:"id"`
		Email          string   `json:"email"`
		Flags          []string `json:"flags"`
		FlagReasons    []string `json:"flagReasons"`
		Decision       string   `json:"decision"`
		DecisionNote   *string  `json:"decisionNote"`
		CampaignSource struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"campaignSource"`
		CanDecide bool `json:"canDecide"`
		CanFlag   bool `json:"canFlag"`
	}

	accessReviewGovernanceCampaign struct {
		CampaignID       string
		CampaignSourceID string
		Entries          []accessReviewGovernanceEntry
	}

	accessReviewGovernanceStatistics struct {
		TotalCount     int `json:"totalCount"`
		DecisionCounts []struct {
			Decision string `json:"decision"`
			Count    int    `json:"count"`
		} `json:"decisionCounts"`
		FlagCounts []struct {
			Flag  string `json:"flag"`
			Count int    `json:"count"`
		} `json:"flagCounts"`
	}
)

func replaceAccessReviewEntryGovernanceSelection(query string) string {
	return strings.ReplaceAll(query, "ENTRY", accessReviewEntryGovernanceSelection)
}

func startAccessReviewGovernanceCampaign(
	t *testing.T,
	client *testutil.Client,
	campaignName string,
	csvData string,
) accessReviewGovernanceCampaign {
	t.Helper()

	orgID := client.GetOrganizationID().String()
	sourceID := factory.NewAccessReviewSource(client, orgID).
		WithName(factory.SafeName(campaignName + " source")).
		WithCsvData(csvData).
		Create()

	campaignID := factory.NewAccessReviewCampaign(client, orgID).
		WithName(factory.SafeName(campaignName)).
		WithAccessReviewSourceIDs([]string{sourceID}).
		Create()

	const startMutation = `
		mutation($input: StartAccessReviewCampaignInput!) {
			startAccessReviewCampaign(input: $input) {
				accessReviewCampaign { id status }
			}
		}
	`

	err := client.Execute(
		startMutation,
		map[string]any{
			"input": map[string]any{
				"accessReviewCampaignId": campaignID,
			},
		},
		nil,
	)
	require.NoError(t, err)

	return requireAccessReviewCampaignPendingActions(t, client, campaignID)
}

func requireAccessReviewCampaignPendingActions(
	t *testing.T,
	client *testutil.Client,
	campaignID string,
) accessReviewGovernanceCampaign {
	t.Helper()

	const pollQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on AccessReviewCampaign {
					status
					sources { id }
					entries(first: 20) {
						totalCount
						edges {
							node {
								ENTRY
							}
						}
					}
				}
			}
		}
	`

	query := replaceAccessReviewEntryGovernanceSelection(pollQuery)

	var (
		lastErr    error
		lastStatus string
		lastCount  int
		result     struct {
			Node struct {
				Status  string `json:"status"`
				Sources []struct {
					ID string `json:"id"`
				} `json:"sources"`
				Entries struct {
					TotalCount int `json:"totalCount"`
					Edges      []struct {
						Node accessReviewGovernanceEntry `json:"node"`
					} `json:"edges"`
				} `json:"entries"`
			} `json:"node"`
		}
	)

	ok := assert.Eventually(
		t,
		func() bool {
			lastErr = client.Execute(query, map[string]any{"id": campaignID}, &result)
			if lastErr != nil {
				return false
			}

			lastStatus = result.Node.Status
			lastCount = result.Node.Entries.TotalCount

			return lastStatus == "PENDING_ACTIONS" &&
				lastCount == accessReviewGovernanceExpectedEntries &&
				len(result.Node.Entries.Edges) == accessReviewGovernanceExpectedEntries
		},
		accessReviewCampaignWorkerPollTimeout,
		accessReviewCampaignWorkerPollTick,
		"campaign worker should import CSV entries and reach PENDING_ACTIONS",
	)
	if !ok {
		if lastErr != nil {
			require.NoError(
				t,
				lastErr,
				"last campaign poll failed (status=%q totalCount=%d)",
				lastStatus,
				lastCount,
			)
		}

		require.FailNow(
			t,
			"timed out waiting for PENDING_ACTIONS with entries (last status=%q totalCount=%d)",
			lastStatus,
			lastCount,
		)
	}

	require.Len(t, result.Node.Sources, 1)

	entries := make([]accessReviewGovernanceEntry, len(result.Node.Entries.Edges))
	for i, edge := range result.Node.Entries.Edges {
		entries[i] = edge.Node
	}

	return accessReviewGovernanceCampaign{
		CampaignID:       campaignID,
		CampaignSourceID: result.Node.Sources[0].ID,
		Entries:          entries,
	}
}

func flagAccessReviewEntryLeaf(
	t *testing.T,
	client *testutil.Client,
	entryID string,
	flags []string,
	flagReasons []string,
) accessReviewGovernanceEntry {
	t.Helper()

	const mutation = `
		mutation($input: FlagAccessReviewEntryInput!) {
			flagAccessReviewEntry(input: $input) {
				accessReviewEntry {
					ENTRY
				}
			}
		}
	`

	var result struct {
		FlagAccessReviewEntry struct {
			AccessReviewEntry accessReviewGovernanceEntry `json:"accessReviewEntry"`
		} `json:"flagAccessReviewEntry"`
	}

	query := replaceAccessReviewEntryGovernanceSelection(mutation)

	err := client.Execute(
		query,
		map[string]any{
			"input": map[string]any{
				"accessReviewEntryId": entryID,
				"flags":               flags,
				"flagReasons":         flagReasons,
			},
		},
		&result,
	)
	require.NoError(t, err)

	return result.FlagAccessReviewEntry.AccessReviewEntry
}

func flagAccessReviewEntryExpectError(
	t *testing.T,
	client *testutil.Client,
	entryID string,
	flags []string,
) error {
	t.Helper()

	const mutation = `
		mutation($input: FlagAccessReviewEntryInput!) {
			flagAccessReviewEntry(input: $input) {
				accessReviewEntry { id }
			}
		}
	`

	return client.Execute(
		mutation,
		map[string]any{
			"input": map[string]any{
				"accessReviewEntryId": entryID,
				"flags":               flags,
			},
		},
		nil,
	)
}

func recordAccessReviewEntryDecisionsLeaf(
	t *testing.T,
	client *testutil.Client,
	decisions []map[string]any,
) []accessReviewGovernanceEntry {
	t.Helper()

	const mutation = `
		mutation($input: RecordAccessReviewEntryDecisionsInput!) {
			recordAccessReviewEntryDecisions(input: $input) {
				accessReviewEntries {
					ENTRY
				}
			}
		}
	`

	var result struct {
		RecordAccessReviewEntryDecisions struct {
			AccessReviewEntries []accessReviewGovernanceEntry `json:"accessReviewEntries"`
		} `json:"recordAccessReviewEntryDecisions"`
	}

	query := replaceAccessReviewEntryGovernanceSelection(mutation)

	err := client.Execute(
		query,
		map[string]any{
			"input": map[string]any{
				"decisions": decisions,
			},
		},
		&result,
	)
	require.NoError(t, err)

	return result.RecordAccessReviewEntryDecisions.AccessReviewEntries
}

func recordAccessReviewEntryDecisionsExpectError(
	t *testing.T,
	client *testutil.Client,
	decisions []map[string]any,
) error {
	t.Helper()

	const mutation = `
		mutation($input: RecordAccessReviewEntryDecisionsInput!) {
			recordAccessReviewEntryDecisions(input: $input) {
				accessReviewEntries { id }
			}
		}
	`

	return client.Execute(
		mutation,
		map[string]any{
			"input": map[string]any{
				"decisions": decisions,
			},
		},
		nil,
	)
}

func queryAccessReviewEntryGovernance(
	t *testing.T,
	client *testutil.Client,
	entryID string,
) accessReviewGovernanceEntry {
	t.Helper()

	const nodeQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on AccessReviewEntry {
					ENTRY
				}
			}
		}
	`

	var result struct {
		Node accessReviewGovernanceEntry `json:"node"`
	}

	query := replaceAccessReviewEntryGovernanceSelection(nodeQuery)

	err := client.Execute(query, map[string]any{"id": entryID}, &result)
	require.NoError(t, err)

	return result.Node
}

func queryAccessReviewCampaignStatistics(
	t *testing.T,
	client *testutil.Client,
	campaignID string,
) accessReviewGovernanceStatistics {
	t.Helper()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on AccessReviewCampaign {
					statistics {
						totalCount
						decisionCounts { decision count }
						flagCounts { flag count }
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			Statistics accessReviewGovernanceStatistics `json:"statistics"`
		} `json:"node"`
	}

	err := client.Execute(query, map[string]any{"id": campaignID}, &result)
	require.NoError(t, err)

	return result.Node.Statistics
}

func closeAccessReviewCampaignLeaf(t *testing.T, client *testutil.Client, campaignID string) string {
	t.Helper()

	const mutation = `
		mutation($input: CloseAccessReviewCampaignInput!) {
			closeAccessReviewCampaign(input: $input) {
				accessReviewCampaign { id status }
			}
		}
	`

	var result struct {
		CloseAccessReviewCampaign struct {
			AccessReviewCampaign struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"accessReviewCampaign"`
		} `json:"closeAccessReviewCampaign"`
	}

	err := client.Execute(
		mutation,
		map[string]any{
			"input": map[string]any{
				"accessReviewCampaignId": campaignID,
			},
		},
		&result,
	)
	require.NoError(t, err)
	require.Equal(t, "COMPLETED", result.CloseAccessReviewCampaign.AccessReviewCampaign.Status)

	return result.CloseAccessReviewCampaign.AccessReviewCampaign.ID
}

func decisionCount(
	stats accessReviewGovernanceStatistics,
	decision string,
) int {
	for _, dc := range stats.DecisionCounts {
		if dc.Decision == decision {
			return dc.Count
		}
	}

	return 0
}
