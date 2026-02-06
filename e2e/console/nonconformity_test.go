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

// createAuditForNC creates a framework and audit for nonconformity testing
func createAuditForNC(t *testing.T, owner *testutil.Client, name string) string {
	t.Helper()

	// Create a framework first
	createFrameworkQuery := `
		mutation CreateFramework($input: CreateFrameworkInput!) {
			createFramework(input: $input) {
				frameworkEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createFrameworkResult struct {
		CreateFramework struct {
			FrameworkEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"frameworkEdge"`
		} `json:"createFramework"`
	}

	err := owner.Execute(createFrameworkQuery, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "Framework for " + name,
		},
	}, &createFrameworkResult)
	require.NoError(t, err)
	frameworkID := createFrameworkResult.CreateFramework.FrameworkEdge.Node.ID

	// Create an audit
	createAuditQuery := `
		mutation CreateAudit($input: CreateAuditInput!) {
			createAudit(input: $input) {
				auditEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createAuditResult struct {
		CreateAudit struct {
			AuditEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"auditEdge"`
		} `json:"createAudit"`
	}

	err = owner.Execute(createAuditQuery, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"frameworkId":    frameworkID,
			"name":           name,
			"state":          "NOT_STARTED",
		},
	}, &createAuditResult)
	require.NoError(t, err)

	return createAuditResult.CreateAudit.AuditEdge.Node.ID
}

func TestNonconformity_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
	auditID := createAuditForNC(t, owner, "NC Test Audit")

	query := `
		mutation CreateNonconformity($input: CreateNonconformityInput!) {
			createNonconformity(input: $input) {
				nonconformityEdge {
					node {
						id
						referenceId
						description
						rootCause
						correctiveAction
						status
					}
				}
			}
		}
	`

	var result struct {
		CreateNonconformity struct {
			NonconformityEdge struct {
				Node struct {
					ID               string `json:"id"`
					ReferenceID      string `json:"referenceId"`
					Description      string `json:"description"`
					RootCause        string `json:"rootCause"`
					CorrectiveAction string `json:"correctiveAction"`
					Status           string `json:"status"`
				} `json:"node"`
			} `json:"nonconformityEdge"`
		} `json:"createNonconformity"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId":   owner.GetOrganizationID().String(),
			"referenceId":      fmt.Sprintf("NC-%d", time.Now().UnixNano()),
			"description":      "Unauthorized access detected",
			"auditId":          auditID,
			"rootCause":        "Insufficient access controls",
			"correctiveAction": "Implement MFA",
			"ownerId":          profileID.String(),
			"status":           "OPEN",
		},
	}, &result)
	require.NoError(t, err)

	nc := result.CreateNonconformity.NonconformityEdge.Node
	assert.NotEmpty(t, nc.ID)
	assert.Equal(t, "Unauthorized access detected", nc.Description)
	assert.Equal(t, "Insufficient access controls", nc.RootCause)
	assert.Equal(t, "OPEN", nc.Status)
}

func TestNonconformity_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
	auditID := createAuditForNC(t, owner, "NC Update Test Audit")

	// Create a nonconformity to update
	createQuery := `
		mutation CreateNonconformity($input: CreateNonconformityInput!) {
			createNonconformity(input: $input) {
				nonconformityEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createResult struct {
		CreateNonconformity struct {
			NonconformityEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"nonconformityEdge"`
		} `json:"createNonconformity"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"referenceId":    fmt.Sprintf("NC-UPDATE-%d", time.Now().UnixNano()),
			"auditId":        auditID,
			"rootCause":      "Original root cause",
			"ownerId":        profileID.String(),
			"status":         "OPEN",
		},
	}, &createResult)
	require.NoError(t, err)
	ncID := createResult.CreateNonconformity.NonconformityEdge.Node.ID

	query := `
		mutation UpdateNonconformity($input: UpdateNonconformityInput!) {
			updateNonconformity(input: $input) {
				nonconformity {
					id
					rootCause
					correctiveAction
					status
				}
			}
		}
	`

	var result struct {
		UpdateNonconformity struct {
			Nonconformity struct {
				ID               string `json:"id"`
				RootCause        string `json:"rootCause"`
				CorrectiveAction string `json:"correctiveAction"`
				Status           string `json:"status"`
			} `json:"nonconformity"`
		} `json:"updateNonconformity"`
	}

	err = owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id":               ncID,
			"rootCause":        "Updated root cause",
			"correctiveAction": "New corrective action",
			"status":           "IN_PROGRESS",
		},
	}, &result)
	require.NoError(t, err)

	nc := result.UpdateNonconformity.Nonconformity
	assert.Equal(t, ncID, nc.ID)
	assert.Equal(t, "Updated root cause", nc.RootCause)
	assert.Equal(t, "New corrective action", nc.CorrectiveAction)
	assert.Equal(t, "IN_PROGRESS", nc.Status)
}

func TestNonconformity_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
	auditID := createAuditForNC(t, owner, "NC Delete Test Audit")

	// Create a nonconformity to delete
	createQuery := `
		mutation CreateNonconformity($input: CreateNonconformityInput!) {
			createNonconformity(input: $input) {
				nonconformityEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createResult struct {
		CreateNonconformity struct {
			NonconformityEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"nonconformityEdge"`
		} `json:"createNonconformity"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"referenceId":    fmt.Sprintf("NC-DELETE-%d", time.Now().UnixNano()),
			"auditId":        auditID,
			"rootCause":      "Test root cause",
			"ownerId":        profileID.String(),
			"status":         "OPEN",
		},
	}, &createResult)
	require.NoError(t, err)
	ncID := createResult.CreateNonconformity.NonconformityEdge.Node.ID

	deleteQuery := `
		mutation DeleteNonconformity($input: DeleteNonconformityInput!) {
			deleteNonconformity(input: $input) {
				deletedNonconformityId
			}
		}
	`

	var deleteResult struct {
		DeleteNonconformity struct {
			DeletedNonconformityID string `json:"deletedNonconformityId"`
		} `json:"deleteNonconformity"`
	}

	err = owner.Execute(deleteQuery, map[string]any{
		"input": map[string]any{
			"nonconformityId": ncID,
		},
	}, &deleteResult)
	require.NoError(t, err)
	assert.Equal(t, ncID, deleteResult.DeleteNonconformity.DeletedNonconformityID)
}

func TestNonconformity_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
	auditID := createAuditForNC(t, owner, "NC List Test Audit")

	// Create multiple nonconformities
	createQuery := `
		mutation CreateNonconformity($input: CreateNonconformityInput!) {
			createNonconformity(input: $input) {
				nonconformityEdge {
					node {
						id
					}
				}
			}
		}
	`

	for i := 0; i < 3; i++ {
		var createResult struct {
			CreateNonconformity struct {
				NonconformityEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"nonconformityEdge"`
			} `json:"createNonconformity"`
		}

		err := owner.Execute(createQuery, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"referenceId":    fmt.Sprintf("NC-LIST-%d-%d", i, time.Now().UnixNano()),
				"auditId":        auditID,
				"rootCause":      fmt.Sprintf("Root cause %d", i),
				"ownerId":        profileID.String(),
				"status":         "OPEN",
			},
		}, &createResult)
		require.NoError(t, err)
	}

	query := `
		query GetNonconformities($id: ID!) {
			node(id: $id) {
				... on Organization {
					nonconformities(first: 10) {
						edges {
							node {
								id
								referenceId
								status
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
			Nonconformities struct {
				Edges []struct {
					Node struct {
						ID          string `json:"id"`
						ReferenceID string `json:"referenceId"`
						Status      string `json:"status"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"nonconformities"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.Nonconformities.TotalCount, 3)
}

func TestNonconformity_StatusValues(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
	auditID := createAuditForNC(t, owner, "NC Status Test Audit")

	statuses := []string{"OPEN", "IN_PROGRESS", "CLOSED"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			query := `
				mutation CreateNonconformity($input: CreateNonconformityInput!) {
					createNonconformity(input: $input) {
						nonconformityEdge {
							node {
								id
								status
							}
						}
					}
				}
			`

			var result struct {
				CreateNonconformity struct {
					NonconformityEdge struct {
						Node struct {
							ID     string `json:"id"`
							Status string `json:"status"`
						} `json:"node"`
					} `json:"nonconformityEdge"`
				} `json:"createNonconformity"`
			}

			err := owner.Execute(query, map[string]any{
				"input": map[string]any{
					"organizationId": owner.GetOrganizationID().String(),
					"referenceId":    fmt.Sprintf("NC-STATUS-%s-%d", status, time.Now().UnixNano()),
					"auditId":        auditID,
					"rootCause":      "Test root cause",
					"ownerId":        profileID.String(),
					"status":         status,
				},
			}, &result)
			require.NoError(t, err)
			assert.Equal(t, status, result.CreateNonconformity.NonconformityEdge.Node.Status)
		})
	}
}
