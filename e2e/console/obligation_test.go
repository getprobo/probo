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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestObligation_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	query := `
		mutation CreateObligation($input: CreateObligationInput!) {
			createObligation(input: $input) {
				obligationEdge {
					node {
						id
						area
						source
						requirement
						regulator
						status
					}
				}
			}
		}
	`

	var result struct {
		CreateObligation struct {
			ObligationEdge struct {
				Node struct {
					ID          string `json:"id"`
					Area        string `json:"area"`
					Source      string `json:"source"`
					Requirement string `json:"requirement"`
					Regulator   string `json:"regulator"`
					Status      string `json:"status"`
				} `json:"node"`
			} `json:"obligationEdge"`
		} `json:"createObligation"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"area":           "Data Protection",
			"source":         "GDPR Article 5",
			"requirement":    "Data must be processed lawfully",
			"regulator":      "ICO",
			"ownerId":        profileID.String(),
			"status":         "NON_COMPLIANT",
			"type":           "LEGAL",
		},
	}, &result)
	require.NoError(t, err)

	assert.NotEmpty(t, result.CreateObligation.ObligationEdge.Node.ID)
	assert.Equal(t, "Data Protection", result.CreateObligation.ObligationEdge.Node.Area)
	assert.Equal(t, "GDPR Article 5", result.CreateObligation.ObligationEdge.Node.Source)
	assert.Equal(t, "Data must be processed lawfully", result.CreateObligation.ObligationEdge.Node.Requirement)
	assert.Equal(t, "ICO", result.CreateObligation.ObligationEdge.Node.Regulator)
	assert.Equal(t, "NON_COMPLIANT", result.CreateObligation.ObligationEdge.Node.Status)
}

func TestObligation_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	// Create an obligation to update
	createQuery := `
		mutation CreateObligation($input: CreateObligationInput!) {
			createObligation(input: $input) {
				obligationEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createResult struct {
		CreateObligation struct {
			ObligationEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"obligationEdge"`
		} `json:"createObligation"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"area":           "Original Area",
			"ownerId":        profileID.String(),
			"status":         "NON_COMPLIANT",
			"type":           "LEGAL",
		},
	}, &createResult)
	require.NoError(t, err)

	obligationID := createResult.CreateObligation.ObligationEdge.Node.ID

	query := `
		mutation UpdateObligation($input: UpdateObligationInput!) {
			updateObligation(input: $input) {
				obligation {
					id
					area
					status
				}
			}
		}
	`

	var result struct {
		UpdateObligation struct {
			Obligation struct {
				ID     string `json:"id"`
				Area   string `json:"area"`
				Status string `json:"status"`
			} `json:"obligation"`
		} `json:"updateObligation"`
	}

	err = owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id":     obligationID,
			"area":   "Updated Area",
			"status": "COMPLIANT",
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, obligationID, result.UpdateObligation.Obligation.ID)
	assert.Equal(t, "Updated Area", result.UpdateObligation.Obligation.Area)
	assert.Equal(t, "COMPLIANT", result.UpdateObligation.Obligation.Status)
}

func TestObligation_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	// Create an obligation to delete
	createQuery := `
		mutation CreateObligation($input: CreateObligationInput!) {
			createObligation(input: $input) {
				obligationEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createResult struct {
		CreateObligation struct {
			ObligationEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"obligationEdge"`
		} `json:"createObligation"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"area":           "Obligation to Delete",
			"ownerId":        profileID.String(),
			"status":         "NON_COMPLIANT",
			"type":           "LEGAL",
		},
	}, &createResult)
	require.NoError(t, err)

	obligationID := createResult.CreateObligation.ObligationEdge.Node.ID

	deleteQuery := `
		mutation DeleteObligation($input: DeleteObligationInput!) {
			deleteObligation(input: $input) {
				deletedObligationId
			}
		}
	`

	var deleteResult struct {
		DeleteObligation struct {
			DeletedObligationID string `json:"deletedObligationId"`
		} `json:"deleteObligation"`
	}

	err = owner.Execute(deleteQuery, map[string]any{
		"input": map[string]any{
			"obligationId": obligationID,
		},
	}, &deleteResult)
	require.NoError(t, err)
	assert.Equal(t, obligationID, deleteResult.DeleteObligation.DeletedObligationID)
}

func TestObligation_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	// Create multiple obligations
	areas := []string{"Area A", "Area B", "Area C"}
	for _, area := range areas {
		query := `
			mutation CreateObligation($input: CreateObligationInput!) {
				createObligation(input: $input) {
					obligationEdge {
						node {
							id
						}
					}
				}
			}
		`

		var result struct {
			CreateObligation struct {
				ObligationEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"obligationEdge"`
			} `json:"createObligation"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"area":           area,
				"ownerId":        profileID.String(),
				"status":         "NON_COMPLIANT",
				"type":           "LEGAL",
			},
		}, &result)
		require.NoError(t, err)
	}

	query := `
		query GetObligations($id: ID!) {
			node(id: $id) {
				... on Organization {
					obligations(first: 10) {
						edges {
							node {
								id
								area
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
			Obligations struct {
				Edges []struct {
					Node struct {
						ID     string `json:"id"`
						Area   string `json:"area"`
						Status string `json:"status"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"obligations"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, result.Node.Obligations.TotalCount, 3)
}

func TestObligation_StatusValues(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	statuses := []string{"NON_COMPLIANT", "PARTIALLY_COMPLIANT", "COMPLIANT"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			query := `
				mutation CreateObligation($input: CreateObligationInput!) {
					createObligation(input: $input) {
						obligationEdge {
							node {
								id
								status
							}
						}
					}
				}
			`

			var result struct {
				CreateObligation struct {
					ObligationEdge struct {
						Node struct {
							ID     string `json:"id"`
							Status string `json:"status"`
						} `json:"node"`
					} `json:"obligationEdge"`
				} `json:"createObligation"`
			}

			err := owner.Execute(query, map[string]any{
				"input": map[string]any{
					"organizationId": owner.GetOrganizationID().String(),
					"area":           "Status Test " + status,
					"ownerId":        profileID.String(),
					"status":         status,
					"type":           "LEGAL",
				},
			}, &result)
			require.NoError(t, err)

			assert.Equal(t, status, result.CreateObligation.ObligationEdge.Node.Status)
		})
	}
}
