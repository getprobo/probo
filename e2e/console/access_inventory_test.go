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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestAccessInventory_ListGrantsAfterCSVSync(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	sourceID := factory.NewAccessReviewSource(owner, orgID).
		WithName("Inventory CSV").
		WithCsvData(testCsvData).
		Create()

	waitForAccessInventoryGrants(t, owner, orgID, 1)

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					accessInventoryGrants(first: 10) {
						totalCount
						edges {
							node {
								id
								principalKey
								entitlementKey
								presence
								principal {
									email
									fullName
								}
								resource {
									displayName
								}
								source {
									id
								}
							}
						}
					}
					accessInventoryPrincipals(first: 10) {
						totalCount
					}
					accessInventoryResources(first: 10) {
						totalCount
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			AccessInventoryGrants struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID             string `json:"id"`
						PrincipalKey   string `json:"principalKey"`
						EntitlementKey string `json:"entitlementKey"`
						Presence       string `json:"presence"`
						Principal      struct {
							Email    string `json:"email"`
							FullName string `json:"fullName"`
						} `json:"principal"`
						Resource struct {
							DisplayName string `json:"displayName"`
						} `json:"resource"`
						Source struct {
							ID string `json:"id"`
						} `json:"source"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"accessInventoryGrants"`
			AccessInventoryPrincipals struct {
				TotalCount int `json:"totalCount"`
			} `json:"accessInventoryPrincipals"`
			AccessInventoryResources struct {
				TotalCount int `json:"totalCount"`
			} `json:"accessInventoryResources"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{"id": orgID}, &result)
	require.NoError(t, err)
	require.GreaterOrEqual(t, result.Node.AccessInventoryGrants.TotalCount, 1)
	require.NotEmpty(t, result.Node.AccessInventoryGrants.Edges)

	grant := result.Node.AccessInventoryGrants.Edges[0].Node
	assert.Equal(t, "PRESENT", grant.Presence)
	assert.Equal(t, "admin", grant.EntitlementKey)
	assert.Equal(t, sourceID, grant.Source.ID)
	assert.Equal(t, "jane@example.com", grant.Principal.Email)
	assert.NotEmpty(t, grant.Resource.DisplayName)
	assert.GreaterOrEqual(t, result.Node.AccessInventoryPrincipals.TotalCount, 1)
	assert.GreaterOrEqual(t, result.Node.AccessInventoryResources.TotalCount, 1)

	const nodeQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on AccessInventoryGrant {
					id
					entitlementKey
					presence
				}
			}
		}
	`

	var nodeResult struct {
		Node struct {
			ID             string `json:"id"`
			EntitlementKey string `json:"entitlementKey"`
			Presence       string `json:"presence"`
		} `json:"node"`
	}

	err = owner.Execute(nodeQuery, map[string]any{"id": grant.ID}, &nodeResult)
	require.NoError(t, err)
	assert.Equal(t, grant.ID, nodeResult.Node.ID)
	assert.Equal(t, "admin", nodeResult.Node.EntitlementKey)
	assert.Equal(t, "PRESENT", nodeResult.Node.Presence)
}

func TestAccessInventory_ChangeEventsChangelogAndFeed(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	sourceID := factory.NewAccessReviewSource(owner, orgID).
		WithName("Inventory Change Events").
		WithCsvData(testCsvData).
		Create()

	waitForAccessInventoryGrants(t, owner, orgID, 1)

	const listQuery = `
		query($id: ID!, $sourceId: ID!) {
			node(id: $id) {
				... on Organization {
					accessInventoryGrants(first: 10, filter: { accessReviewSourceId: $sourceId }) {
						edges {
							node {
								id
								changeEvents(first: 10) {
									totalCount
									edges {
										node {
											id
											changeType
											effectiveAt
											after {
												presence
												entitlementDisplayName
											}
										}
									}
								}
							}
						}
					}
					accessInventoryChangeEvents(
						first: 10
						filter: {
							accessReviewSourceId: $sourceId
							changeTypes: [APPEARED]
						}
					) {
						totalCount
						edges {
							node {
								id
								changeType
								principalKey
								source {
									id
								}
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			AccessInventoryGrants struct {
				Edges []struct {
					Node struct {
						ID           string `json:"id"`
						ChangeEvents struct {
							TotalCount int `json:"totalCount"`
							Edges      []struct {
								Node struct {
									ID          string `json:"id"`
									ChangeType  string `json:"changeType"`
									EffectiveAt string `json:"effectiveAt"`
									After       *struct {
										Presence               string `json:"presence"`
										EntitlementDisplayName string `json:"entitlementDisplayName"`
									} `json:"after"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"changeEvents"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"accessInventoryGrants"`
			AccessInventoryChangeEvents struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID           string `json:"id"`
						ChangeType   string `json:"changeType"`
						PrincipalKey string `json:"principalKey"`
						Source       struct {
							ID string `json:"id"`
						} `json:"source"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"accessInventoryChangeEvents"`
		} `json:"node"`
	}

	err := owner.Execute(
		listQuery,
		map[string]any{
			"id":       orgID,
			"sourceId": sourceID,
		},
		&result,
	)
	require.NoError(t, err)
	require.NotEmpty(t, result.Node.AccessInventoryGrants.Edges)

	grant := result.Node.AccessInventoryGrants.Edges[0].Node
	require.GreaterOrEqual(t, grant.ChangeEvents.TotalCount, 1)
	require.NotEmpty(t, grant.ChangeEvents.Edges)
	assert.Equal(t, "APPEARED", grant.ChangeEvents.Edges[0].Node.ChangeType)
	require.NotNil(t, grant.ChangeEvents.Edges[0].Node.After)
	assert.Equal(t, "PRESENT", grant.ChangeEvents.Edges[0].Node.After.Presence)
	assert.Equal(t, "admin", grant.ChangeEvents.Edges[0].Node.After.EntitlementDisplayName)
	assert.NotEmpty(t, grant.ChangeEvents.Edges[0].Node.EffectiveAt)

	require.GreaterOrEqual(t, result.Node.AccessInventoryChangeEvents.TotalCount, 1)
	require.NotEmpty(t, result.Node.AccessInventoryChangeEvents.Edges)
	event := result.Node.AccessInventoryChangeEvents.Edges[0].Node
	assert.Equal(t, "APPEARED", event.ChangeType)
	assert.Equal(t, sourceID, event.Source.ID)
	assert.NotEmpty(t, event.PrincipalKey)

	const nodeQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on AccessInventoryChangeEvent {
					id
					changeType
					effectiveAt
				}
			}
		}
	`

	var nodeResult struct {
		Node struct {
			ID          string `json:"id"`
			ChangeType  string `json:"changeType"`
			EffectiveAt string `json:"effectiveAt"`
		} `json:"node"`
	}

	err = owner.Execute(nodeQuery, map[string]any{"id": event.ID}, &nodeResult)
	require.NoError(t, err)
	assert.Equal(t, event.ID, nodeResult.Node.ID)
	assert.Equal(t, "APPEARED", nodeResult.Node.ChangeType)
	assert.NotEmpty(t, nodeResult.Node.EffectiveAt)
}

func TestAccessInventory_SyncNowAndStatus(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	sourceID := factory.NewAccessReviewSource(owner, orgID).
		WithName("Inventory Sync Status").
		WithCsvData(testCsvData).
		Create()

	waitForAccessInventoryGrants(t, owner, orgID, 1)

	const syncMutation = `
		mutation($input: SyncAccessInventoryInput!) {
			syncAccessInventory(input: $input) {
				accessReviewSource {
					id
					syncConsecutiveFailures
					latestSyncRun {
						id
						status
					}
				}
				accessSourceSyncRun {
					id
					status
					source {
						id
					}
				}
			}
		}
	`

	var syncResult struct {
		SyncAccessInventory struct {
			AccessReviewSource struct {
				ID                      string `json:"id"`
				SyncConsecutiveFailures int    `json:"syncConsecutiveFailures"`
				LatestSyncRun           *struct {
					ID     string `json:"id"`
					Status string `json:"status"`
				} `json:"latestSyncRun"`
			} `json:"accessReviewSource"`
			AccessSourceSyncRun struct {
				ID     string `json:"id"`
				Status string `json:"status"`
				Source struct {
					ID string `json:"id"`
				} `json:"source"`
			} `json:"accessSourceSyncRun"`
		} `json:"syncAccessInventory"`
	}

	err := owner.Execute(
		syncMutation,
		map[string]any{
			"input": map[string]any{
				"accessReviewSourceId": sourceID,
			},
		},
		&syncResult,
	)
	require.NoError(t, err)

	payload := syncResult.SyncAccessInventory
	assert.Equal(t, sourceID, payload.AccessReviewSource.ID)
	assert.Equal(t, sourceID, payload.AccessSourceSyncRun.Source.ID)
	require.NotEmpty(t, payload.AccessSourceSyncRun.ID)
	require.NotNil(t, payload.AccessReviewSource.LatestSyncRun)
	assert.Equal(t, payload.AccessSourceSyncRun.ID, payload.AccessReviewSource.LatestSyncRun.ID)

	runID := payload.AccessSourceSyncRun.ID
	waitForAccessSourceSyncRunSuccess(t, owner, runID)

	const statusQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on AccessReviewSource {
					id
					lastSuccessfulSyncAt
					nextSyncAt
					syncConsecutiveFailures
					latestSyncRun {
						id
						status
					}
				}
			}
		}
	`

	var statusResult struct {
		Node struct {
			ID                      string  `json:"id"`
			LastSuccessfulSyncAt    *string `json:"lastSuccessfulSyncAt"`
			NextSyncAt              *string `json:"nextSyncAt"`
			SyncConsecutiveFailures int     `json:"syncConsecutiveFailures"`
			LatestSyncRun           *struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"latestSyncRun"`
		} `json:"node"`
	}

	err = owner.Execute(statusQuery, map[string]any{"id": sourceID}, &statusResult)
	require.NoError(t, err)
	assert.Equal(t, sourceID, statusResult.Node.ID)
	require.NotNil(t, statusResult.Node.LatestSyncRun)
	assert.Equal(t, runID, statusResult.Node.LatestSyncRun.ID)
	assert.Equal(t, "SUCCESS", statusResult.Node.LatestSyncRun.Status)
	require.NotNil(t, statusResult.Node.LastSuccessfulSyncAt)

	const nodeQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on AccessSourceSyncRun {
					id
					status
				}
			}
		}
	`

	var nodeResult struct {
		Node struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"node"`
	}

	err = owner.Execute(nodeQuery, map[string]any{"id": runID}, &nodeResult)
	require.NoError(t, err)
	assert.Equal(t, runID, nodeResult.Node.ID)
	assert.Equal(t, "SUCCESS", nodeResult.Node.Status)
}

func waitForAccessSourceSyncRunSuccess(
	t *testing.T,
	client *testutil.Client,
	runID string,
) {
	t.Helper()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on AccessSourceSyncRun {
					id
					status
				}
			}
		}
	`

	require.Eventually(t, func() bool {
		var result struct {
			Node struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"node"`
		}

		err := client.Execute(query, map[string]any{"id": runID}, &result)
		if err != nil {
			t.Logf("access source sync run poll: %v", err)
			return false
		}

		switch result.Node.Status {
		case "SUCCESS":
			return true
		case "FAILED":
			t.Fatalf("access source sync run %s failed", runID)
			return false
		default:
			return false
		}
	}, 60*time.Second, 500*time.Millisecond)
}

func waitForAccessInventoryGrants(
	t *testing.T,
	client *testutil.Client,
	organizationID string,
	minCount int,
) {
	t.Helper()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					accessInventoryGrants(first: 1) {
						totalCount
					}
				}
			}
		}
	`

	require.Eventually(t, func() bool {
		var result struct {
			Node struct {
				AccessInventoryGrants struct {
					TotalCount int `json:"totalCount"`
				} `json:"accessInventoryGrants"`
			} `json:"node"`
		}

		err := client.Execute(query, map[string]any{"id": organizationID}, &result)
		if err != nil {
			t.Logf("access inventory poll: %v", err)
			return false
		}

		return result.Node.AccessInventoryGrants.TotalCount >= minCount
	}, 60*time.Second, 500*time.Millisecond)
}
