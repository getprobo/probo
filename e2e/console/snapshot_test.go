// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestSnapshot_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	query := `
		mutation CreateSnapshot($input: CreateSnapshotInput!) {
			createSnapshot(input: $input) {
				snapshotEdge {
					node {
						id
						name
						description
						type
					}
				}
			}
		}
	`

	var result struct {
		CreateSnapshot struct {
			SnapshotEdge struct {
				Node struct {
					ID          string `json:"id"`
					Name        string `json:"name"`
					Description string `json:"description"`
					Type        string `json:"type"`
				} `json:"node"`
			} `json:"snapshotEdge"`
		} `json:"createSnapshot"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "Q4 2024 Risk Snapshot",
			"description":    "Quarterly risk assessment snapshot",
			"type":           "RISKS",
		},
	}, &result)
	require.NoError(t, err)

	snapshot := result.CreateSnapshot.SnapshotEdge.Node
	assert.NotEmpty(t, snapshot.ID)
	assert.Equal(t, "Q4 2024 Risk Snapshot", snapshot.Name)
	assert.Equal(t, "Quarterly risk assessment snapshot", snapshot.Description)
	assert.Equal(t, "RISKS", snapshot.Type)
}

func TestSnapshot_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a snapshot to delete
	createQuery := `
		mutation CreateSnapshot($input: CreateSnapshotInput!) {
			createSnapshot(input: $input) {
				snapshotEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createResult struct {
		CreateSnapshot struct {
			SnapshotEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"snapshotEdge"`
		} `json:"createSnapshot"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           fmt.Sprintf("Snapshot to Delete %d", time.Now().UnixNano()),
			"type":           "ASSETS",
		},
	}, &createResult)
	require.NoError(t, err)

	snapshotID := createResult.CreateSnapshot.SnapshotEdge.Node.ID

	deleteQuery := `
		mutation DeleteSnapshot($input: DeleteSnapshotInput!) {
			deleteSnapshot(input: $input) {
				deletedSnapshotId
			}
		}
	`

	var deleteResult struct {
		DeleteSnapshot struct {
			DeletedSnapshotID string `json:"deletedSnapshotId"`
		} `json:"deleteSnapshot"`
	}

	err = owner.Execute(deleteQuery, map[string]any{
		"input": map[string]any{
			"snapshotId": snapshotID,
		},
	}, &deleteResult)

	require.NoError(t, err)
	assert.Equal(t, snapshotID, deleteResult.DeleteSnapshot.DeletedSnapshotID)
}

func TestSnapshot_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create multiple snapshots
	snapshotTypes := []string{"RISKS", "VENDORS", "ASSETS", "DATA"}
	for i, snapshotType := range snapshotTypes {
		query := `
			mutation CreateSnapshot($input: CreateSnapshotInput!) {
				createSnapshot(input: $input) {
					snapshotEdge {
						node {
							id
						}
					}
				}
			}
		`

		var result struct {
			CreateSnapshot struct {
				SnapshotEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"snapshotEdge"`
			} `json:"createSnapshot"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           fmt.Sprintf("Snapshot %d %d", i, time.Now().UnixNano()),
				"type":           snapshotType,
			},
		}, &result)
		require.NoError(t, err)
	}

	query := `
		query GetSnapshots($id: ID!) {
			node(id: $id) {
				... on Organization {
					snapshots(first: 10) {
						edges {
							node {
								id
								name
								type
							}
						}
						totalCount
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			Snapshots struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
						Type string `json:"type"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"snapshots"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.Snapshots.TotalCount, 4)
}

func TestSnapshot_Types(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	snapshotTypes := []string{"RISKS", "VENDORS", "ASSETS", "DATA"}

	for _, snapshotType := range snapshotTypes {
		t.Run(snapshotType, func(t *testing.T) {
			query := `
				mutation CreateSnapshot($input: CreateSnapshotInput!) {
					createSnapshot(input: $input) {
						snapshotEdge {
							node {
								id
								type
							}
						}
					}
				}
			`

			var result struct {
				CreateSnapshot struct {
					SnapshotEdge struct {
						Node struct {
							ID   string `json:"id"`
							Type string `json:"type"`
						} `json:"node"`
					} `json:"snapshotEdge"`
				} `json:"createSnapshot"`
			}

			err := owner.Execute(query, map[string]any{
				"input": map[string]any{
					"organizationId": owner.GetOrganizationID().String(),
					"name":           fmt.Sprintf("Snapshot Type %s %d", snapshotType, time.Now().UnixNano()),
					"type":           snapshotType,
				},
			}, &result)
			require.NoError(t, err)
			assert.Equal(t, snapshotType, result.CreateSnapshot.SnapshotEdge.Node.Type)
		})
	}
}
