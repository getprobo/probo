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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestFinding_CreateNonconformity(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)

	query := `
		mutation CreateFinding($input: CreateFindingInput!) {
			createFinding(input: $input) {
				findingEdge {
					node {
						id
						kind
						referenceId
						description
						rootCause
						correctiveAction
						status
						priority
					}
				}
			}
		}
	`

	var result struct {
		CreateFinding struct {
			FindingEdge struct {
				Node struct {
					ID               string `json:"id"`
					Kind             string `json:"kind"`
					ReferenceID      string `json:"referenceId"`
					Description      string `json:"description"`
					RootCause        string `json:"rootCause"`
					CorrectiveAction string `json:"correctiveAction"`
					Status           string `json:"status"`
					Priority         string `json:"priority"`
				} `json:"node"`
			} `json:"findingEdge"`
		} `json:"createFinding"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId":   owner.GetOrganizationID().String(),
			"kind":             "MINOR_NONCONFORMITY",
			"description":      "Unauthorized access detected",
			"rootCause":        "Insufficient access controls",
			"correctiveAction": "Implement MFA",
			"ownerId":          profileID,
			"status":           "OPEN",
			"priority":         "HIGH",
		},
	}, &result)
	require.NoError(t, err)

	node := result.CreateFinding.FindingEdge.Node
	assert.NotEmpty(t, node.ID)
	assert.Equal(t, "MINOR_NONCONFORMITY", node.Kind)
	assert.Equal(t, "Unauthorized access detected", node.Description)
	assert.Equal(t, "Insufficient access controls", node.RootCause)
	assert.Equal(t, "Implement MFA", node.CorrectiveAction)
	assert.Equal(t, "OPEN", node.Status)
	assert.Equal(t, "HIGH", node.Priority)
}

func TestFinding_CreateObservation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)

	query := `
		mutation CreateFinding($input: CreateFindingInput!) {
			createFinding(input: $input) {
				findingEdge {
					node {
						id
						kind
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
		CreateFinding struct {
			FindingEdge struct {
				Node struct {
					ID          string `json:"id"`
					Kind        string `json:"kind"`
					Description string `json:"description"`
					Source      string `json:"source"`
					Status      string `json:"status"`
					Priority    string `json:"priority"`
				} `json:"node"`
			} `json:"findingEdge"`
		} `json:"createFinding"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"kind":           "OBSERVATION",
			"description":    "Improve security training program",
			"source":         "Internal Audit",
			"ownerId":        profileID,
			"status":         "OPEN",
			"priority":       "MEDIUM",
		},
	}, &result)
	require.NoError(t, err)

	node := result.CreateFinding.FindingEdge.Node
	assert.NotEmpty(t, node.ID)
	assert.Equal(t, "OBSERVATION", node.Kind)
	assert.Equal(t, "Improve security training program", node.Description)
	assert.Equal(t, "Internal Audit", node.Source)
	assert.Equal(t, "OPEN", node.Status)
	assert.Equal(t, "MEDIUM", node.Priority)
}

func TestFinding_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)

	createQuery := `
		mutation CreateFinding($input: CreateFindingInput!) {
			createFinding(input: $input) {
				findingEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createResult struct {
		CreateFinding struct {
			FindingEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"findingEdge"`
		} `json:"createFinding"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"kind":           "MINOR_NONCONFORMITY",
			"description":    "Original description",
			"ownerId":        profileID,
			"status":         "OPEN",
			"priority":       "LOW",
		},
	}, &createResult)
	require.NoError(t, err)
	findingID := createResult.CreateFinding.FindingEdge.Node.ID

	query := `
		mutation UpdateFinding($input: UpdateFindingInput!) {
			updateFinding(input: $input) {
				finding {
					id
					description
					rootCause
					correctiveAction
					status
					priority
				}
			}
		}
	`

	var result struct {
		UpdateFinding struct {
			Finding struct {
				ID               string `json:"id"`
				Description      string `json:"description"`
				RootCause        string `json:"rootCause"`
				CorrectiveAction string `json:"correctiveAction"`
				Status           string `json:"status"`
				Priority         string `json:"priority"`
			} `json:"finding"`
		} `json:"updateFinding"`
	}

	err = owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id":               findingID,
			"description":      "Updated description",
			"rootCause":        "Updated root cause",
			"correctiveAction": "New corrective action",
			"status":           "IN_PROGRESS",
			"priority":         "HIGH",
		},
	}, &result)
	require.NoError(t, err)

	f := result.UpdateFinding.Finding
	assert.Equal(t, findingID, f.ID)
	assert.Equal(t, "Updated description", f.Description)
	assert.Equal(t, "Updated root cause", f.RootCause)
	assert.Equal(t, "New corrective action", f.CorrectiveAction)
	assert.Equal(t, "IN_PROGRESS", f.Status)
	assert.Equal(t, "HIGH", f.Priority)
}

func TestFinding_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)

	createQuery := `
		mutation CreateFinding($input: CreateFindingInput!) {
			createFinding(input: $input) {
				findingEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createResult struct {
		CreateFinding struct {
			FindingEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"findingEdge"`
		} `json:"createFinding"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"kind":           "OBSERVATION",
			"ownerId":        profileID,
			"status":         "OPEN",
			"priority":       "LOW",
		},
	}, &createResult)
	require.NoError(t, err)
	findingID := createResult.CreateFinding.FindingEdge.Node.ID

	query := `
		mutation DeleteFinding($input: DeleteFindingInput!) {
			deleteFinding(input: $input) {
				deletedFindingId
			}
		}
	`

	var result struct {
		DeleteFinding struct {
			DeletedFindingID string `json:"deletedFindingId"`
		} `json:"deleteFinding"`
	}

	err = owner.Execute(query, map[string]any{
		"input": map[string]any{
			"findingId": findingID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, findingID, result.DeleteFinding.DeletedFindingID)
}

func TestFinding_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)

	createQuery := `
		mutation CreateFinding($input: CreateFindingInput!) {
			createFinding(input: $input) {
				findingEdge {
					node {
						id
					}
				}
			}
		}
	`

	for i := range 3 {
		var createResult struct {
			CreateFinding struct {
				FindingEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"findingEdge"`
			} `json:"createFinding"`
		}

		err := owner.Execute(createQuery, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"kind":           "MINOR_NONCONFORMITY",
				"description":    fmt.Sprintf("Finding %d", i),
				"ownerId":        profileID,
				"status":         "OPEN",
				"priority":       "MEDIUM",
			},
		}, &createResult)
		require.NoError(t, err)
	}

	query := `
		query GetFindings($id: ID!) {
			node(id: $id) {
				... on Organization {
					findings(first: 10) {
						edges {
							node {
								id
								kind
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
			Findings struct {
				Edges []struct {
					Node struct {
						ID          string `json:"id"`
						Kind        string `json:"kind"`
						ReferenceID string `json:"referenceId"`
						Status      string `json:"status"`
						Priority    string `json:"priority"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"findings"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.Findings.TotalCount, 3)
}

func TestFinding_ListWithKindFilter(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)

	createQuery := `
		mutation CreateFinding($input: CreateFindingInput!) {
			createFinding(input: $input) {
				findingEdge {
					node {
						id
					}
				}
			}
		}
	`

	kinds := []string{"MINOR_NONCONFORMITY", "MAJOR_NONCONFORMITY", "OBSERVATION", "EXCEPTION"}
	for _, kind := range kinds {
		var createResult struct {
			CreateFinding struct {
				FindingEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"findingEdge"`
			} `json:"createFinding"`
		}

		err := owner.Execute(createQuery, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"kind":           kind,
				"ownerId":        profileID,
				"status":         "OPEN",
				"priority":       "LOW",
			},
		}, &createResult)
		require.NoError(t, err)
	}

	query := `
		query GetFindings($id: ID!, $filter: FindingFilter) {
			node(id: $id) {
				... on Organization {
					findings(first: 10, filter: $filter) {
						edges {
							node {
								id
								kind
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
			Findings struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Kind string `json:"kind"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"findings"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id":     owner.GetOrganizationID().String(),
		"filter": map[string]any{"kind": "OBSERVATION"},
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.Findings.TotalCount, 1)

	for _, edge := range result.Node.Findings.Edges {
		assert.Equal(t, "OBSERVATION", edge.Node.Kind)
	}
}

func TestFinding_CreateAuditMapping(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)
	frameworkID := factory.CreateFramework(owner)
	auditID := factory.CreateAudit(owner, frameworkID)

	// Create a finding first
	createQuery := `
		mutation CreateFinding($input: CreateFindingInput!) {
			createFinding(input: $input) {
				findingEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createResult struct {
		CreateFinding struct {
			FindingEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"findingEdge"`
		} `json:"createFinding"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"kind":           "MINOR_NONCONFORMITY",
			"ownerId":        profileID,
			"status":         "OPEN",
			"priority":       "HIGH",
		},
	}, &createResult)
	require.NoError(t, err)
	findingID := createResult.CreateFinding.FindingEdge.Node.ID

	// Link finding to audit
	linkQuery := `
		mutation CreateFindingAuditMapping($input: CreateFindingAuditMappingInput!) {
			createFindingAuditMapping(input: $input) {
				findingEdge {
					node {
						id
					}
				}
				auditEdge {
					node {
						id
					}
				}
			}
		}
	`

	var linkResult struct {
		CreateFindingAuditMapping struct {
			FindingEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"findingEdge"`
			AuditEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"auditEdge"`
		} `json:"createFindingAuditMapping"`
	}

	err = owner.Execute(linkQuery, map[string]any{
		"input": map[string]any{
			"findingId":   findingID,
			"auditId":     auditID,
			"referenceId": "MinNC/001",
		},
	}, &linkResult)
	require.NoError(t, err)
	assert.Equal(t, findingID, linkResult.CreateFindingAuditMapping.FindingEdge.Node.ID)
	assert.Equal(t, auditID, linkResult.CreateFindingAuditMapping.AuditEdge.Node.ID)

	// Verify audits appear on finding
	auditsQuery := `
		query GetFinding($id: ID!) {
			node(id: $id) {
				... on Finding {
					audits(first: 10) {
						edges {
							node {
								id
							}
						}
						totalCount
					}
				}
			}
		}
	`

	var auditsResult struct {
		Node struct {
			Audits struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"audits"`
		} `json:"node"`
	}

	err = owner.Execute(auditsQuery, map[string]any{
		"id": findingID,
	}, &auditsResult)
	require.NoError(t, err)
	assert.Equal(t, 1, auditsResult.Node.Audits.TotalCount)
	assert.Equal(t, auditID, auditsResult.Node.Audits.Edges[0].Node.ID)
}

func TestFinding_DeleteAuditMapping(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)
	frameworkID := factory.CreateFramework(owner)
	auditID := factory.CreateAudit(owner, frameworkID)

	// Create a finding
	createQuery := `
		mutation CreateFinding($input: CreateFindingInput!) {
			createFinding(input: $input) {
				findingEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createResult struct {
		CreateFinding struct {
			FindingEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"findingEdge"`
		} `json:"createFinding"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"kind":           "OBSERVATION",
			"ownerId":        profileID,
			"status":         "OPEN",
			"priority":       "LOW",
		},
	}, &createResult)
	require.NoError(t, err)
	findingID := createResult.CreateFinding.FindingEdge.Node.ID

	// Link finding to audit
	linkQuery := `
		mutation CreateFindingAuditMapping($input: CreateFindingAuditMappingInput!) {
			createFindingAuditMapping(input: $input) {
				findingEdge {
					node { id }
				}
				auditEdge {
					node { id }
				}
			}
		}
	`

	var linkResult struct {
		CreateFindingAuditMapping struct {
			FindingEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"findingEdge"`
			AuditEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"auditEdge"`
		} `json:"createFindingAuditMapping"`
	}

	err = owner.Execute(linkQuery, map[string]any{
		"input": map[string]any{
			"findingId":   findingID,
			"auditId":     auditID,
			"referenceId": "MajNC/001",
		},
	}, &linkResult)
	require.NoError(t, err)

	// Delete the mapping
	unlinkQuery := `
		mutation DeleteFindingAuditMapping($input: DeleteFindingAuditMappingInput!) {
			deleteFindingAuditMapping(input: $input) {
				deletedFindingId
				deletedAuditId
			}
		}
	`

	var unlinkResult struct {
		DeleteFindingAuditMapping struct {
			DeletedFindingID string `json:"deletedFindingId"`
			DeletedAuditID   string `json:"deletedAuditId"`
		} `json:"deleteFindingAuditMapping"`
	}

	err = owner.Execute(unlinkQuery, map[string]any{
		"input": map[string]any{
			"findingId": findingID,
			"auditId":   auditID,
		},
	}, &unlinkResult)
	require.NoError(t, err)
	assert.Equal(t, findingID, unlinkResult.DeleteFindingAuditMapping.DeletedFindingID)
	assert.Equal(t, auditID, unlinkResult.DeleteFindingAuditMapping.DeletedAuditID)

	// Verify no audits on finding
	auditsQuery := `
		query GetFinding($id: ID!) {
			node(id: $id) {
				... on Finding {
					audits(first: 10) {
						totalCount
					}
				}
			}
		}
	`

	var auditsResult struct {
		Node struct {
			Audits struct {
				TotalCount int `json:"totalCount"`
			} `json:"audits"`
		} `json:"node"`
	}

	err = owner.Execute(auditsQuery, map[string]any{
		"id": findingID,
	}, &auditsResult)
	require.NoError(t, err)
	assert.Equal(t, 0, auditsResult.Node.Audits.TotalCount)
}

func TestFinding_StatusAndPriorityValues(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)

	createQuery := `
		mutation CreateFinding($input: CreateFindingInput!) {
			createFinding(input: $input) {
				findingEdge {
					node {
						id
						status
						priority
					}
				}
			}
		}
	`

	t.Run(
		"status values",
		func(t *testing.T) {
			t.Parallel()
			riskID := factory.CreateRisk(owner)
			statuses := []string{"OPEN", "IN_PROGRESS", "CLOSED", "RISK_ACCEPTED", "MITIGATED", "FALSE_POSITIVE"}

			for _, status := range statuses {
				t.Run(
					status,
					func(t *testing.T) {
						t.Parallel()
						var result struct {
							CreateFinding struct {
								FindingEdge struct {
									Node struct {
										ID     string `json:"id"`
										Status string `json:"status"`
									} `json:"node"`
								} `json:"findingEdge"`
							} `json:"createFinding"`
						}

						input := map[string]any{
							"organizationId": owner.GetOrganizationID().String(),
							"kind":           "MINOR_NONCONFORMITY",
							"ownerId":        profileID,
							"status":         status,
							"priority":       "LOW",
						}

						if status == "RISK_ACCEPTED" {
							input["riskId"] = riskID
						}

						err := owner.Execute(createQuery, map[string]any{
							"input": input,
						}, &result)
						require.NoError(t, err)
						assert.Equal(t, status, result.CreateFinding.FindingEdge.Node.Status)
					},
				)
			}
		},
	)

	t.Run(
		"priority values",
		func(t *testing.T) {
			t.Parallel()
			priorities := []string{"LOW", "MEDIUM", "HIGH"}

			for _, priority := range priorities {
				t.Run(
					priority,
					func(t *testing.T) {
						t.Parallel()
						var result struct {
							CreateFinding struct {
								FindingEdge struct {
									Node struct {
										ID       string `json:"id"`
										Priority string `json:"priority"`
									} `json:"node"`
								} `json:"findingEdge"`
							} `json:"createFinding"`
						}

						err := owner.Execute(createQuery, map[string]any{
							"input": map[string]any{
								"organizationId": owner.GetOrganizationID().String(),
								"kind":           "OBSERVATION",
								"ownerId":        profileID,
								"status":         "OPEN",
								"priority":       priority,
							},
						}, &result)
						require.NoError(t, err)
						assert.Equal(t, priority, result.CreateFinding.FindingEdge.Node.Priority)
					},
				)
			}
		},
	)
}
