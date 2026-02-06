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

func TestContinualImprovement_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	// TODO: right now we need to invite and accept invite to get new profile.
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	query := `
		mutation CreateContinualImprovement($input: CreateContinualImprovementInput!) {
			createContinualImprovement(input: $input) {
				continualImprovementEdge {
					node {
						id
						referenceId
						description
						source
						status
						priority
					}
				}
			}
		}
	`

	var result struct {
		CreateContinualImprovement struct {
			ContinualImprovementEdge struct {
				Node struct {
					ID          string `json:"id"`
					ReferenceID string `json:"referenceId"`
					Description string `json:"description"`
					Source      string `json:"source"`
					Status      string `json:"status"`
					Priority    string `json:"priority"`
				} `json:"node"`
			} `json:"continualImprovementEdge"`
		} `json:"createContinualImprovement"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"referenceId":    fmt.Sprintf("CI-%d", time.Now().UnixNano()),
			"description":    "Improve security training program",
			"source":         "Internal Audit",
			"ownerId":        profileID.String(),
			"status":         "OPEN",
			"priority":       "HIGH",
		},
	}, &result)
	require.NoError(t, err)

	ci := result.CreateContinualImprovement.ContinualImprovementEdge.Node
	assert.NotEmpty(t, ci.ID)
	assert.Equal(t, "Improve security training program", ci.Description)
	assert.Equal(t, "Internal Audit", ci.Source)
	assert.Equal(t, "OPEN", ci.Status)
	assert.Equal(t, "HIGH", ci.Priority)
}

func TestContinualImprovement_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	// TODO: right now we need to invite and accept invite to get new profile.
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	createQuery := `
		mutation CreateContinualImprovement($input: CreateContinualImprovementInput!) {
			createContinualImprovement(input: $input) {
				continualImprovementEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createResult struct {
		CreateContinualImprovement struct {
			ContinualImprovementEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"continualImprovementEdge"`
		} `json:"createContinualImprovement"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"referenceId":    fmt.Sprintf("CI-UPDATE-%d", time.Now().UnixNano()),
			"description":    "Original description",
			"ownerId":        profileID.String(),
			"status":         "OPEN",
			"priority":       "LOW",
		},
	}, &createResult)
	require.NoError(t, err)
	ciID := createResult.CreateContinualImprovement.ContinualImprovementEdge.Node.ID

	query := `
		mutation UpdateContinualImprovement($input: UpdateContinualImprovementInput!) {
			updateContinualImprovement(input: $input) {
				continualImprovement {
					id
					description
					status
					priority
				}
			}
		}
	`

	var result struct {
		UpdateContinualImprovement struct {
			ContinualImprovement struct {
				ID          string `json:"id"`
				Description string `json:"description"`
				Status      string `json:"status"`
				Priority    string `json:"priority"`
			} `json:"continualImprovement"`
		} `json:"updateContinualImprovement"`
	}

	err = owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id":          ciID,
			"description": "Updated description",
			"status":      "IN_PROGRESS",
			"priority":    "HIGH",
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, ciID, result.UpdateContinualImprovement.ContinualImprovement.ID)
	assert.Equal(t, "Updated description", result.UpdateContinualImprovement.ContinualImprovement.Description)
	assert.Equal(t, "IN_PROGRESS", result.UpdateContinualImprovement.ContinualImprovement.Status)
	assert.Equal(t, "HIGH", result.UpdateContinualImprovement.ContinualImprovement.Priority)
}

func TestContinualImprovement_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	// TODO: right now we need to invite and accept invite to get new profile.
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	createQuery := `
		mutation CreateContinualImprovement($input: CreateContinualImprovementInput!) {
			createContinualImprovement(input: $input) {
				continualImprovementEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createResult struct {
		CreateContinualImprovement struct {
			ContinualImprovementEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"continualImprovementEdge"`
		} `json:"createContinualImprovement"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"referenceId":    fmt.Sprintf("CI-DELETE-%d", time.Now().UnixNano()),
			"ownerId":        profileID.String(),
			"status":         "OPEN",
			"priority":       "LOW",
		},
	}, &createResult)
	require.NoError(t, err)
	ciID := createResult.CreateContinualImprovement.ContinualImprovementEdge.Node.ID

	query := `
		mutation DeleteContinualImprovement($input: DeleteContinualImprovementInput!) {
			deleteContinualImprovement(input: $input) {
				deletedContinualImprovementId
			}
		}
	`

	var result struct {
		DeleteContinualImprovement struct {
			DeletedContinualImprovementID string `json:"deletedContinualImprovementId"`
		} `json:"deleteContinualImprovement"`
	}

	err = owner.Execute(query, map[string]any{
		"input": map[string]any{
			"continualImprovementId": ciID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, ciID, result.DeleteContinualImprovement.DeletedContinualImprovementID)
}

func TestContinualImprovement_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	// TODO: right now we need to invite and accept invite to get new profile.
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	createQuery := `
		mutation CreateContinualImprovement($input: CreateContinualImprovementInput!) {
			createContinualImprovement(input: $input) {
				continualImprovementEdge {
					node {
						id
					}
				}
			}
		}
	`

	for i := 0; i < 3; i++ {
		_, err := owner.Do(createQuery, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"referenceId":    fmt.Sprintf("CI-LIST-%d-%d", i, time.Now().UnixNano()),
				"description":    fmt.Sprintf("Improvement %d", i),
				"ownerId":        profileID.String(),
				"status":         "OPEN",
				"priority":       "MEDIUM",
			},
		})
		require.NoError(t, err)
	}

	query := `
		query GetContinualImprovements($id: ID!) {
			node(id: $id) {
				... on Organization {
					continualImprovements(first: 10) {
						edges {
							node {
								id
								referenceId
								status
								priority
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
			ContinualImprovements struct {
				Edges []struct {
					Node struct {
						ID          string `json:"id"`
						ReferenceID string `json:"referenceId"`
						Status      string `json:"status"`
						Priority    string `json:"priority"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"continualImprovements"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.ContinualImprovements.TotalCount, 3)
}

func TestContinualImprovement_StatusAndPriorityValues(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	// TODO: right now we need to invite and accept invite to get new profile.
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	t.Run("status values", func(t *testing.T) {
		statuses := []string{"OPEN", "IN_PROGRESS", "CLOSED"}

		for _, status := range statuses {
			t.Run(status, func(t *testing.T) {
				query := `
					mutation CreateContinualImprovement($input: CreateContinualImprovementInput!) {
						createContinualImprovement(input: $input) {
							continualImprovementEdge {
								node {
									id
									status
								}
							}
						}
					}
				`

				var result struct {
					CreateContinualImprovement struct {
						ContinualImprovementEdge struct {
							Node struct {
								ID     string `json:"id"`
								Status string `json:"status"`
							} `json:"node"`
						} `json:"continualImprovementEdge"`
					} `json:"createContinualImprovement"`
				}

				err := owner.Execute(query, map[string]any{
					"input": map[string]any{
						"organizationId": owner.GetOrganizationID().String(),
						"referenceId":    fmt.Sprintf("CI-STATUS-%s-%d", status, time.Now().UnixNano()),
						"ownerId":        profileID.String(),
						"status":         status,
						"priority":       "LOW",
					},
				}, &result)
				require.NoError(t, err)
				assert.Equal(t, status, result.CreateContinualImprovement.ContinualImprovementEdge.Node.Status)
			})
		}
	})

	t.Run("priority values", func(t *testing.T) {
		priorities := []string{"LOW", "MEDIUM", "HIGH"}

		for _, priority := range priorities {
			t.Run(priority, func(t *testing.T) {
				query := `
					mutation CreateContinualImprovement($input: CreateContinualImprovementInput!) {
						createContinualImprovement(input: $input) {
							continualImprovementEdge {
								node {
									id
									priority
								}
							}
						}
					}
				`

				var result struct {
					CreateContinualImprovement struct {
						ContinualImprovementEdge struct {
							Node struct {
								ID       string `json:"id"`
								Priority string `json:"priority"`
							} `json:"node"`
						} `json:"continualImprovementEdge"`
					} `json:"createContinualImprovement"`
				}

				err := owner.Execute(query, map[string]any{
					"input": map[string]any{
						"organizationId": owner.GetOrganizationID().String(),
						"referenceId":    fmt.Sprintf("CI-PRIORITY-%s-%d", priority, time.Now().UnixNano()),
						"ownerId":        profileID.String(),
						"status":         "OPEN",
						"priority":       priority,
					},
				}, &result)
				require.NoError(t, err)
				assert.Equal(t, priority, result.CreateContinualImprovement.ContinualImprovementEdge.Node.Priority)
			})
		}
	})
}
