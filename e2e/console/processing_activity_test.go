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
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestProcessingActivity_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	tests := []struct {
		name        string
		input       map[string]any
		assertField string
		assertValue any
	}{
		{
			name: "with full details",
			input: map[string]any{
				"name":                           "Customer Data Processing",
				"purpose":                        "Provide services",
				"lawfulBasis":                    "CONTRACTUAL_NECESSITY",
				"internationalTransfers":         false,
				"specialOrCriminalData":          "NO",
				"dataProtectionImpactAssessment": "NOT_NEEDED",
				"transferImpactAssessment":       "NOT_NEEDED",
			},
			assertField: "name",
			assertValue: "Customer Data Processing",
		},
		{
			name: "with CONSENT lawful basis",
			input: map[string]any{
				"name":                           "Consent Processing",
				"lawfulBasis":                    "CONSENT",
				"specialOrCriminalData":          "NO",
				"internationalTransfers":         false,
				"dataProtectionImpactAssessment": "NOT_NEEDED",
				"transferImpactAssessment":       "NOT_NEEDED",
			},
			assertField: "lawfulBasis",
			assertValue: "CONSENT",
		},
		{
			name: "with LEGITIMATE_INTEREST lawful basis",
			input: map[string]any{
				"name":                           "Legitimate Interest Processing",
				"lawfulBasis":                    "LEGITIMATE_INTEREST",
				"specialOrCriminalData":          "NO",
				"internationalTransfers":         false,
				"dataProtectionImpactAssessment": "NOT_NEEDED",
				"transferImpactAssessment":       "NOT_NEEDED",
			},
			assertField: "lawfulBasis",
			assertValue: "LEGITIMATE_INTEREST",
		},
		{
			name: "with international transfers enabled",
			input: map[string]any{
				"name":                           "International Processing",
				"lawfulBasis":                    "CONSENT",
				"internationalTransfers":         true,
				"specialOrCriminalData":          "NO",
				"dataProtectionImpactAssessment": "NOT_NEEDED",
				"transferImpactAssessment":       "NOT_NEEDED",
			},
			assertField: "internationalTransfers",
			assertValue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation CreateProcessingActivity($input: CreateProcessingActivityInput!) {
					createProcessingActivity(input: $input) {
						processingActivityEdge {
							node {
								id
								name
								lawfulBasis
								internationalTransfers
							}
						}
					}
				}
			`

			input := map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
			}
			for k, v := range tt.input {
				input[k] = v
			}

			var result struct {
				CreateProcessingActivity struct {
					ProcessingActivityEdge struct {
						Node struct {
							ID                     string `json:"id"`
							Name                   string `json:"name"`
							LawfulBasis            string `json:"lawfulBasis"`
							InternationalTransfers bool   `json:"internationalTransfers"`
						} `json:"node"`
					} `json:"processingActivityEdge"`
				} `json:"createProcessingActivity"`
			}

			err := owner.Execute(query, map[string]any{"input": input}, &result)
			require.NoError(t, err)

			node := result.CreateProcessingActivity.ProcessingActivityEdge.Node
			assert.NotEmpty(t, node.ID)

			switch tt.assertField {
			case "name":
				assert.Equal(t, tt.assertValue, node.Name)
			case "lawfulBasis":
				assert.Equal(t, tt.assertValue, node.LawfulBasis)
			case "internationalTransfers":
				assert.Equal(t, tt.assertValue, node.InternationalTransfers)
			}
		})
	}
}

func TestProcessingActivity_Create_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	tests := []struct {
		name              string
		input             map[string]any
		skipOrganization  bool
		wantErrorContains string
	}{
		{
			name: "missing organizationId",
			input: map[string]any{
				"name":                           "Test Processing",
				"lawfulBasis":                    "CONSENT",
				"specialOrCriminalData":          "NO",
				"internationalTransfers":         false,
				"dataProtectionImpactAssessment": "NOT_NEEDED",
				"transferImpactAssessment":       "NOT_NEEDED",
			},
			skipOrganization:  true,
			wantErrorContains: "organizationId",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation CreateProcessingActivity($input: CreateProcessingActivityInput!) {
					createProcessingActivity(input: $input) {
						processingActivityEdge {
							node {
								id
							}
						}
					}
				}
			`

			input := make(map[string]any)
			if !tt.skipOrganization {
				input["organizationId"] = owner.GetOrganizationID().String()
			}
			for k, v := range tt.input {
				input[k] = v
			}

			_, err := owner.Do(query, map[string]any{"input": input})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestProcessingActivity_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	tests := []struct {
		name        string
		setup       func() string
		input       func(id string) map[string]any
		assertField string
		assertValue any
	}{
		{
			name: "update name",
			setup: func() string {
				return factory.NewProcessingActivity(owner).
					WithName("Processing to Update").
					Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{
					"id":   id,
					"name": "Updated Processing Activity",
				}
			},
			assertField: "name",
			assertValue: "Updated Processing Activity",
		},
		{
			name: "update lawful basis",
			setup: func() string {
				return factory.NewProcessingActivity(owner).
					WithName("Lawful Basis Test").
					WithLawfulBasis("CONSENT").
					Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "lawfulBasis": "LEGITIMATE_INTEREST"}
			},
			assertField: "lawfulBasis",
			assertValue: "LEGITIMATE_INTEREST",
		},
		{
			name: "enable international transfers",
			setup: func() string {
				return factory.NewProcessingActivity(owner).
					WithName("International Test").
					WithInternationalTransfers(false).
					Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "internationalTransfers": true}
			},
			assertField: "internationalTransfers",
			assertValue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paID := tt.setup()

			query := `
				mutation UpdateProcessingActivity($input: UpdateProcessingActivityInput!) {
					updateProcessingActivity(input: $input) {
						processingActivity {
							id
							name
							lawfulBasis
							internationalTransfers
						}
					}
				}
			`

			var result struct {
				UpdateProcessingActivity struct {
					ProcessingActivity struct {
						ID                     string `json:"id"`
						Name                   string `json:"name"`
						LawfulBasis            string `json:"lawfulBasis"`
						InternationalTransfers bool   `json:"internationalTransfers"`
					} `json:"processingActivity"`
				} `json:"updateProcessingActivity"`
			}

			err := owner.Execute(query, map[string]any{"input": tt.input(paID)}, &result)
			require.NoError(t, err)

			pa := result.UpdateProcessingActivity.ProcessingActivity
			switch tt.assertField {
			case "name":
				assert.Equal(t, tt.assertValue, pa.Name)
			case "lawfulBasis":
				assert.Equal(t, tt.assertValue, pa.LawfulBasis)
			case "internationalTransfers":
				assert.Equal(t, tt.assertValue, pa.InternationalTransfers)
			}
		})
	}
}

func TestProcessingActivity_Update_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	tests := []struct {
		name              string
		setup             func() string
		input             func(id string) map[string]any
		wantErrorContains string
	}{
		{
			name:  "invalid ID format",
			setup: func() string { return "invalid-id-format" },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test"}
			},
			wantErrorContains: "base64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paID := tt.setup()

			query := `
				mutation UpdateProcessingActivity($input: UpdateProcessingActivityInput!) {
					updateProcessingActivity(input: $input) {
						processingActivity {
							id
						}
					}
				}
			`

			_, err := owner.Do(query, map[string]any{"input": tt.input(paID)})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestProcessingActivity_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("delete existing processing activity", func(t *testing.T) {
		paID := factory.NewProcessingActivity(owner).WithName("PA to Delete").Create()

		query := `
			mutation DeleteProcessingActivity($input: DeleteProcessingActivityInput!) {
				deleteProcessingActivity(input: $input) {
					deletedProcessingActivityId
				}
			}
		`

		var result struct {
			DeleteProcessingActivity struct {
				DeletedProcessingActivityID string `json:"deletedProcessingActivityId"`
			} `json:"deleteProcessingActivity"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{"processingActivityId": paID},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, paID, result.DeleteProcessingActivity.DeletedProcessingActivityID)
	})
}

func TestProcessingActivity_Delete_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	tests := []struct {
		name              string
		paID              string
		wantErrorContains string
	}{
		{
			name:              "invalid ID format",
			paID:              "invalid-id-format",
			wantErrorContains: "base64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation DeleteProcessingActivity($input: DeleteProcessingActivityInput!) {
					deleteProcessingActivity(input: $input) {
						deletedProcessingActivityId
					}
				}
			`

			_, err := owner.Do(query, map[string]any{
				"input": map[string]any{"processingActivityId": tt.paID},
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestProcessingActivity_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	paNames := []string{"Processing A", "Processing B", "Processing C"}
	for _, name := range paNames {
		factory.NewProcessingActivity(owner).WithName(name).Create()
	}

	query := `
		query GetProcessingActivities($id: ID!) {
			node(id: $id) {
				... on Organization {
					processingActivities(first: 10) {
						edges {
							node {
								id
								name
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
			ProcessingActivities struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"processingActivities"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.ProcessingActivities.TotalCount, 3)
}

func TestProcessingActivity_Query(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("query with non-existent ID returns error", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on ProcessingActivity {
						id
						name
					}
				}
			}
		`

		err := owner.ExecuteShouldFail(query, map[string]any{
			"id": "V0wtM0tMNmJBQ1lBQUFBQUFackhLSTJfbXJJRUFZVXo",
		})
		require.Error(t, err, "Non-existent ID should return error")
	})
}

func TestProcessingActivity_Timestamps(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("createdAt and updatedAt are set on create", func(t *testing.T) {
		beforeCreate := time.Now().Add(-time.Second)

		query := `
			mutation CreateProcessingActivity($input: CreateProcessingActivityInput!) {
				createProcessingActivity(input: $input) {
					processingActivityEdge {
						node {
							id
							createdAt
							updatedAt
						}
					}
				}
			}
		`

		var result struct {
			CreateProcessingActivity struct {
				ProcessingActivityEdge struct {
					Node struct {
						ID        string    `json:"id"`
						CreatedAt time.Time `json:"createdAt"`
						UpdatedAt time.Time `json:"updatedAt"`
					} `json:"node"`
				} `json:"processingActivityEdge"`
			} `json:"createProcessingActivity"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId":                 owner.GetOrganizationID().String(),
				"name":                           "Timestamp Test PA",
				"lawfulBasis":                    "CONSENT",
				"specialOrCriminalData":          "NO",
				"internationalTransfers":         false,
				"dataProtectionImpactAssessment": "NOT_NEEDED",
				"transferImpactAssessment":       "NOT_NEEDED",
			},
		}, &result)
		require.NoError(t, err)

		node := result.CreateProcessingActivity.ProcessingActivityEdge.Node
		testutil.AssertTimestampsOnCreate(t, node.CreatedAt, node.UpdatedAt, beforeCreate)
	})

	t.Run("updatedAt changes on update", func(t *testing.T) {
		paID := factory.NewProcessingActivity(owner).WithName("Timestamp Update Test").Create()

		getQuery := `
			query($id: ID!) {
				node(id: $id) {
					... on ProcessingActivity {
						createdAt
						updatedAt
					}
				}
			}
		`

		var getResult struct {
			Node struct {
				CreatedAt time.Time `json:"createdAt"`
				UpdatedAt time.Time `json:"updatedAt"`
			} `json:"node"`
		}

		err := owner.Execute(getQuery, map[string]any{"id": paID}, &getResult)
		require.NoError(t, err)

		initialCreatedAt := getResult.Node.CreatedAt
		initialUpdatedAt := getResult.Node.UpdatedAt

		time.Sleep(10 * time.Millisecond)

		updateQuery := `
			mutation UpdateProcessingActivity($input: UpdateProcessingActivityInput!) {
				updateProcessingActivity(input: $input) {
					processingActivity {
						createdAt
						updatedAt
					}
				}
			}
		`

		var updateResult struct {
			UpdateProcessingActivity struct {
				ProcessingActivity struct {
					CreatedAt time.Time `json:"createdAt"`
					UpdatedAt time.Time `json:"updatedAt"`
				} `json:"processingActivity"`
			} `json:"updateProcessingActivity"`
		}

		err = owner.Execute(updateQuery, map[string]any{
			"input": map[string]any{
				"id":   paID,
				"name": "Updated Timestamp Test",
			},
		}, &updateResult)
		require.NoError(t, err)

		pa := updateResult.UpdateProcessingActivity.ProcessingActivity
		testutil.AssertTimestampsOnUpdate(t, pa.CreatedAt, pa.UpdatedAt, initialCreatedAt, initialUpdatedAt)
	})
}

func TestProcessingActivity_SubResolvers(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	paID := factory.NewProcessingActivity(owner).WithName("SubResolver Test PA").Create()

	t.Run("organization sub-resolver", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on ProcessingActivity {
						id
						organization {
							id
							name
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID           string `json:"id"`
				Organization struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"organization"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": paID}, &result)
		require.NoError(t, err)
		assert.Equal(t, owner.GetOrganizationID().String(), result.Node.Organization.ID)
		assert.NotEmpty(t, result.Node.Organization.Name)
	})
}

func TestProcessingActivity_RBAC(t *testing.T) {
	t.Parallel()

	t.Run("create", func(t *testing.T) {
		t.Run("owner can create", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)

			_, err := owner.Do(`
				mutation CreateProcessingActivity($input: CreateProcessingActivityInput!) {
					createProcessingActivity(input: $input) {
						processingActivityEdge { node { id } }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"organizationId":                 owner.GetOrganizationID().String(),
					"name":                           "RBAC Test PA",
					"lawfulBasis":                    "CONSENT",
					"specialOrCriminalData":          "NO",
					"internationalTransfers":         false,
					"dataProtectionImpactAssessment": "NOT_NEEDED",
					"transferImpactAssessment":       "NOT_NEEDED",
				},
			})
			require.NoError(t, err, "owner should be able to create processing activity")
		})

		t.Run("admin can create", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)

			_, err := admin.Do(`
				mutation CreateProcessingActivity($input: CreateProcessingActivityInput!) {
					createProcessingActivity(input: $input) {
						processingActivityEdge { node { id } }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"organizationId":                 admin.GetOrganizationID().String(),
					"name":                           "RBAC Test PA",
					"lawfulBasis":                    "CONSENT",
					"specialOrCriminalData":          "NO",
					"internationalTransfers":         false,
					"dataProtectionImpactAssessment": "NOT_NEEDED",
					"transferImpactAssessment":       "NOT_NEEDED",
				},
			})
			require.NoError(t, err, "admin should be able to create processing activity")
		})

		t.Run("viewer cannot create", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

			_, err := viewer.Do(`
				mutation CreateProcessingActivity($input: CreateProcessingActivityInput!) {
					createProcessingActivity(input: $input) {
						processingActivityEdge { node { id } }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"organizationId":                 viewer.GetOrganizationID().String(),
					"name":                           "RBAC Test PA",
					"lawfulBasis":                    "CONSENT",
					"specialOrCriminalData":          "NO",
					"internationalTransfers":         false,
					"dataProtectionImpactAssessment": "NOT_NEEDED",
					"transferImpactAssessment":       "NOT_NEEDED",
				},
			})
			testutil.RequireForbiddenError(t, err, "viewer should not be able to create processing activity")
		})
	})

	t.Run("update", func(t *testing.T) {
		t.Run("owner can update", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			paID := factory.NewProcessingActivity(owner).WithName("RBAC Update Test").Create()

			_, err := owner.Do(`
				mutation UpdateProcessingActivity($input: UpdateProcessingActivityInput!) {
					updateProcessingActivity(input: $input) {
						processingActivity { id }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"id":   paID,
					"name": "Updated by Owner",
				},
			})
			require.NoError(t, err, "owner should be able to update processing activity")
		})

		t.Run("admin can update", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
			paID := factory.NewProcessingActivity(owner).WithName("RBAC Update Test").Create()

			_, err := admin.Do(`
				mutation UpdateProcessingActivity($input: UpdateProcessingActivityInput!) {
					updateProcessingActivity(input: $input) {
						processingActivity { id }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"id":   paID,
					"name": "Updated by Admin",
				},
			})
			require.NoError(t, err, "admin should be able to update processing activity")
		})

		t.Run("viewer cannot update", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			paID := factory.NewProcessingActivity(owner).WithName("RBAC Update Test").Create()

			_, err := viewer.Do(`
				mutation UpdateProcessingActivity($input: UpdateProcessingActivityInput!) {
					updateProcessingActivity(input: $input) {
						processingActivity { id }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"id":   paID,
					"name": "Updated by Viewer",
				},
			})
			testutil.RequireForbiddenError(t, err, "viewer should not be able to update processing activity")
		})
	})

	t.Run("delete", func(t *testing.T) {
		t.Run("owner can delete", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			paID := factory.NewProcessingActivity(owner).WithName("RBAC Delete Test").Create()

			_, err := owner.Do(`
				mutation DeleteProcessingActivity($input: DeleteProcessingActivityInput!) {
					deleteProcessingActivity(input: $input) {
						deletedProcessingActivityId
					}
				}
			`, map[string]any{
				"input": map[string]any{"processingActivityId": paID},
			})
			require.NoError(t, err, "owner should be able to delete processing activity")
		})

		t.Run("admin can delete", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
			paID := factory.NewProcessingActivity(owner).WithName("RBAC Delete Test").Create()

			_, err := admin.Do(`
				mutation DeleteProcessingActivity($input: DeleteProcessingActivityInput!) {
					deleteProcessingActivity(input: $input) {
						deletedProcessingActivityId
					}
				}
			`, map[string]any{
				"input": map[string]any{"processingActivityId": paID},
			})
			require.NoError(t, err, "admin should be able to delete processing activity")
		})

		t.Run("viewer cannot delete", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			paID := factory.NewProcessingActivity(owner).WithName("RBAC Delete Test").Create()

			_, err := viewer.Do(`
				mutation DeleteProcessingActivity($input: DeleteProcessingActivityInput!) {
					deleteProcessingActivity(input: $input) {
						deletedProcessingActivityId
					}
				}
			`, map[string]any{
				"input": map[string]any{"processingActivityId": paID},
			})
			testutil.RequireForbiddenError(t, err, "viewer should not be able to delete processing activity")
		})
	})

	t.Run("read", func(t *testing.T) {
		t.Run("owner can read", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			paID := factory.NewProcessingActivity(owner).WithName("RBAC Read Test").Create()

			var result struct {
				Node *struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			}

			err := owner.Execute(`
				query($id: ID!) {
					node(id: $id) {
						... on ProcessingActivity { id name }
					}
				}
			`, map[string]any{"id": paID}, &result)
			require.NoError(t, err, "owner should be able to read processing activity")
			require.NotNil(t, result.Node, "owner should receive processing activity data")
		})

		t.Run("admin can read", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
			paID := factory.NewProcessingActivity(owner).WithName("RBAC Read Test").Create()

			var result struct {
				Node *struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			}

			err := admin.Execute(`
				query($id: ID!) {
					node(id: $id) {
						... on ProcessingActivity { id name }
					}
				}
			`, map[string]any{"id": paID}, &result)
			require.NoError(t, err, "admin should be able to read processing activity")
			require.NotNil(t, result.Node, "admin should receive processing activity data")
		})

		t.Run("viewer can read", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			paID := factory.NewProcessingActivity(owner).WithName("RBAC Read Test").Create()

			var result struct {
				Node *struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			}

			err := viewer.Execute(`
				query($id: ID!) {
					node(id: $id) {
						... on ProcessingActivity { id name }
					}
				}
			`, map[string]any{"id": paID}, &result)
			require.NoError(t, err, "viewer should be able to read processing activity")
			require.NotNil(t, result.Node, "viewer should receive processing activity data")
		})
	})
}


func TestProcessingActivity_Pagination(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	for i := 0; i < 5; i++ {
		factory.NewProcessingActivity(owner).
			WithName(fmt.Sprintf("Pagination PA %d", i)).
			Create()
	}

	t.Run("first/after pagination", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						processingActivities(first: 2) {
							edges {
								node { id name }
								cursor
							}
							pageInfo {
								hasNextPage
								hasPreviousPage
								startCursor
								endCursor
							}
							totalCount
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ProcessingActivities struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
						Cursor string `json:"cursor"`
					} `json:"edges"`
					PageInfo   testutil.PageInfo `json:"pageInfo"`
					TotalCount int               `json:"totalCount"`
				} `json:"processingActivities"`
			} `json:"node"`
		}

		err := owner.Execute(
			query,
			map[string]any{
				"id": owner.GetOrganizationID().String(),
			},
			&result,
		)
		require.NoError(t, err)

		testutil.AssertFirstPage(t, len(result.Node.ProcessingActivities.Edges), result.Node.ProcessingActivities.PageInfo, 2, true)
		assert.GreaterOrEqual(t, result.Node.ProcessingActivities.TotalCount, 5)

		testutil.AssertHasMorePages(t, result.Node.ProcessingActivities.PageInfo)
		queryAfter := `
			query($id: ID!, $after: CursorKey) {
				node(id: $id) {
					... on Organization {
						processingActivities(first: 2, after: $after) {
							edges {
								node { id name }
							}
							pageInfo {
								hasNextPage
								hasPreviousPage
							}
						}
					}
				}
			}
		`

		var resultAfter struct {
			Node struct {
				ProcessingActivities struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo testutil.PageInfo `json:"pageInfo"`
				} `json:"processingActivities"`
			} `json:"node"`
		}

		err = owner.Execute(queryAfter, map[string]any{
			"id":    owner.GetOrganizationID().String(),
			"after": *result.Node.ProcessingActivities.PageInfo.EndCursor,
		}, &resultAfter)
		require.NoError(t, err)

		testutil.AssertMiddlePage(t, len(resultAfter.Node.ProcessingActivities.Edges), resultAfter.Node.ProcessingActivities.PageInfo, 2)
	})

	t.Run("last/before pagination", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						processingActivities(last: 2) {
							edges {
								node { id name }
							}
							pageInfo {
								hasNextPage
								hasPreviousPage
							}
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ProcessingActivities struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo testutil.PageInfo `json:"pageInfo"`
				} `json:"processingActivities"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id": owner.GetOrganizationID().String(),
		}, &result)
		require.NoError(t, err)

		testutil.AssertLastPage(t, len(result.Node.ProcessingActivities.Edges), result.Node.ProcessingActivities.PageInfo, 2, true)
	})
}

func TestProcessingActivity_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	paID := factory.NewProcessingActivity(org1Owner).WithName("Org1 PA").Create()

	t.Run("cannot read processing activity from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on ProcessingActivity {
						id
						name
					}
				}
			}
		`

		var result struct {
			Node *struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{"id": paID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "processing activity")
	})

	t.Run("cannot update processing activity from another organization", func(t *testing.T) {
		query := `
			mutation UpdateProcessingActivity($input: UpdateProcessingActivityInput!) {
				updateProcessingActivity(input: $input) {
					processingActivity { id }
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":   paID,
				"name": "Hijacked PA",
			},
		})
		require.Error(t, err, "Should not be able to update processing activity from another org")
	})

	t.Run("cannot delete processing activity from another organization", func(t *testing.T) {
		query := `
			mutation DeleteProcessingActivity($input: DeleteProcessingActivityInput!) {
				deleteProcessingActivity(input: $input) {
					deletedProcessingActivityId
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"processingActivityId": paID,
			},
		})
		require.Error(t, err, "Should not be able to delete processing activity from another org")
	})

	t.Run("cannot list processing activities from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						processingActivities(first: 100) {
							edges {
								node {
									id
									name
								}
							}
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ProcessingActivities struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"processingActivities"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{
			"id": org1Owner.GetOrganizationID().String(),
		}, &result)

		if err == nil {
			for _, edge := range result.Node.ProcessingActivities.Edges {
				assert.NotEqual(t, paID, edge.Node.ID, "Should not see processing activity from another org")
			}
		}
	})
}

func TestProcessingActivity_Ordering(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	factory.NewProcessingActivity(owner).WithName("AAA Order Test").Create()
	factory.NewProcessingActivity(owner).WithName("ZZZ Order Test").Create()

	t.Run("order by created_at descending", func(t *testing.T) {
		query := `
			query($id: ID!, $orderBy: ProcessingActivityOrder) {
				node(id: $id) {
					... on Organization {
						processingActivities(first: 100, orderBy: $orderBy) {
							edges {
								node {
									id
									createdAt
								}
							}
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ProcessingActivities struct {
					Edges []struct {
						Node struct {
							ID        string    `json:"id"`
							CreatedAt time.Time `json:"createdAt"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"processingActivities"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id": owner.GetOrganizationID().String(),
			"orderBy": map[string]any{
				"field":     "CREATED_AT",
				"direction": "DESC",
			},
		}, &result)
		require.NoError(t, err)

		times := make([]time.Time, len(result.Node.ProcessingActivities.Edges))
		for i, edge := range result.Node.ProcessingActivities.Edges {
			times[i] = edge.Node.CreatedAt
		}
		testutil.AssertTimesOrderedDescending(t, times, "createdAt")
	})
}

func TestProcessingActivity_LawfulBasis(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	lawfulBases := []string{"CONSENT", "CONTRACTUAL_NECESSITY", "LEGITIMATE_INTEREST"}

	for _, basis := range lawfulBases {
		t.Run(basis, func(t *testing.T) {
			paID := factory.NewProcessingActivity(owner).
				WithName(fmt.Sprintf("PA Lawful Basis %s", basis)).
				WithLawfulBasis(basis).
				Create()

			query := `
				query($id: ID!) {
					node(id: $id) {
						... on ProcessingActivity {
							id
							lawfulBasis
						}
					}
				}
			`

			var result struct {
				Node struct {
					ID          string `json:"id"`
					LawfulBasis string `json:"lawfulBasis"`
				} `json:"node"`
			}

			err := owner.Execute(query, map[string]any{"id": paID}, &result)
			require.NoError(t, err)
			assert.Equal(t, basis, result.Node.LawfulBasis)
		})
	}
}
