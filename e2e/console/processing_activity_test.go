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
				"name":                                 "Customer Data Processing",
				"purpose":                              "Provide services",
				"lawfulBasis":                          "CONTRACTUAL_NECESSITY",
				"internationalTransfers":               false,
				"specialOrCriminalData":                "NO",
				"dataProtectionImpactAssessmentNeeded": "NOT_NEEDED",
				"transferImpactAssessmentNeeded":       "NOT_NEEDED",
				"role":                                 "CONTROLLER",
			},
			assertField: "name",
			assertValue: "Customer Data Processing",
		},
		{
			name: "with CONSENT lawful basis",
			input: map[string]any{
				"name":                                 "Consent Processing",
				"lawfulBasis":                          "CONSENT",
				"specialOrCriminalData":                "NO",
				"internationalTransfers":               false,
				"dataProtectionImpactAssessmentNeeded": "NOT_NEEDED",
				"transferImpactAssessmentNeeded":       "NOT_NEEDED",
				"role":                                 "CONTROLLER",
			},
			assertField: "lawfulBasis",
			assertValue: "CONSENT",
		},
		{
			name: "with LEGITIMATE_INTEREST lawful basis",
			input: map[string]any{
				"name":                                 "Legitimate Interest Processing",
				"lawfulBasis":                          "LEGITIMATE_INTEREST",
				"specialOrCriminalData":                "NO",
				"internationalTransfers":               false,
				"dataProtectionImpactAssessmentNeeded": "NOT_NEEDED",
				"transferImpactAssessmentNeeded":       "NOT_NEEDED",
				"role":                                 "CONTROLLER",
			},
			assertField: "lawfulBasis",
			assertValue: "LEGITIMATE_INTEREST",
		},
		{
			name: "with international transfers enabled",
			input: map[string]any{
				"name":                                 "International Processing",
				"lawfulBasis":                          "CONSENT",
				"internationalTransfers":               true,
				"specialOrCriminalData":                "NO",
				"dataProtectionImpactAssessmentNeeded": "NOT_NEEDED",
				"transferImpactAssessmentNeeded":       "NOT_NEEDED",
				"role":                                 "CONTROLLER",
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
				"name":                                 "Test Processing",
				"lawfulBasis":                          "CONSENT",
				"specialOrCriminalData":                "NO",
				"internationalTransfers":               false,
				"dataProtectionImpactAssessmentNeeded": "NOT_NEEDED",
				"transferImpactAssessmentNeeded":       "NOT_NEEDED",
				"role":                                 "CONTROLLER",
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
				"organizationId":                       owner.GetOrganizationID().String(),
				"name":                                 "Timestamp Test PA",
				"lawfulBasis":                          "CONSENT",
				"specialOrCriminalData":                "NO",
				"internationalTransfers":               false,
				"dataProtectionImpactAssessmentNeeded": "NOT_NEEDED",
				"transferImpactAssessmentNeeded":       "NOT_NEEDED",
				"role":                                 "CONTROLLER",
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
					"organizationId":                       owner.GetOrganizationID().String(),
					"name":                                 "RBAC Test PA",
					"lawfulBasis":                          "CONSENT",
					"specialOrCriminalData":                "NO",
					"internationalTransfers":               false,
					"dataProtectionImpactAssessmentNeeded": "NOT_NEEDED",
					"transferImpactAssessmentNeeded":       "NOT_NEEDED",
					"role":                                 "CONTROLLER",
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
					"organizationId":                       admin.GetOrganizationID().String(),
					"name":                                 "RBAC Test PA",
					"lawfulBasis":                          "CONSENT",
					"specialOrCriminalData":                "NO",
					"internationalTransfers":               false,
					"dataProtectionImpactAssessmentNeeded": "NOT_NEEDED",
					"transferImpactAssessmentNeeded":       "NOT_NEEDED",
					"role":                                 "CONTROLLER",
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
					"organizationId":                       viewer.GetOrganizationID().String(),
					"name":                                 "RBAC Test PA",
					"lawfulBasis":                          "CONSENT",
					"specialOrCriminalData":                "NO",
					"internationalTransfers":               false,
					"dataProtectionImpactAssessmentNeeded": "NOT_NEEDED",
					"transferImpactAssessmentNeeded":       "NOT_NEEDED",
					"role":                                 "CONTROLLER",
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

func TestProcessingActivity_Role(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	roles := []string{"CONTROLLER", "PROCESSOR"}

	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			query := `
				mutation CreateProcessingActivity($input: CreateProcessingActivityInput!) {
					createProcessingActivity(input: $input) {
						processingActivityEdge {
							node {
								id
								role
							}
						}
					}
				}
			`

			var result struct {
				CreateProcessingActivity struct {
					ProcessingActivityEdge struct {
						Node struct {
							ID   string `json:"id"`
							Role string `json:"role"`
						} `json:"node"`
					} `json:"processingActivityEdge"`
				} `json:"createProcessingActivity"`
			}

			err := owner.Execute(query, map[string]any{
				"input": map[string]any{
					"organizationId":                       owner.GetOrganizationID().String(),
					"name":                                 fmt.Sprintf("PA Role %s", role),
					"lawfulBasis":                          "CONSENT",
					"specialOrCriminalData":                "NO",
					"internationalTransfers":               false,
					"dataProtectionImpactAssessmentNeeded": "NOT_NEEDED",
					"transferImpactAssessmentNeeded":       "NOT_NEEDED",
					"role":                                 role,
				},
			}, &result)
			require.NoError(t, err)
			assert.Equal(t, role, result.CreateProcessingActivity.ProcessingActivityEdge.Node.Role)
		})
	}

	t.Run("update role", func(t *testing.T) {
		paID := factory.NewProcessingActivity(owner).
			WithName("Role Update Test").
			Create()

		query := `
			mutation UpdateProcessingActivity($input: UpdateProcessingActivityInput!) {
				updateProcessingActivity(input: $input) {
					processingActivity {
						id
						role
					}
				}
			}
		`

		var result struct {
			UpdateProcessingActivity struct {
				ProcessingActivity struct {
					ID   string `json:"id"`
					Role string `json:"role"`
				} `json:"processingActivity"`
			} `json:"updateProcessingActivity"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":   paID,
				"role": "PROCESSOR",
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, "PROCESSOR", result.UpdateProcessingActivity.ProcessingActivity.Role)
	})
}

func TestProcessingActivity_ReviewDates(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("create with review dates", func(t *testing.T) {
		query := `
			mutation CreateProcessingActivity($input: CreateProcessingActivityInput!) {
				createProcessingActivity(input: $input) {
					processingActivityEdge {
						node {
							id
							lastReviewDate
							nextReviewDate
						}
					}
				}
			}
		`

		var result struct {
			CreateProcessingActivity struct {
				ProcessingActivityEdge struct {
					Node struct {
						ID             string  `json:"id"`
						LastReviewDate *string `json:"lastReviewDate"`
						NextReviewDate *string `json:"nextReviewDate"`
					} `json:"node"`
				} `json:"processingActivityEdge"`
			} `json:"createProcessingActivity"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId":                       owner.GetOrganizationID().String(),
				"name":                                 "PA Review Dates Test",
				"lawfulBasis":                          "CONSENT",
				"specialOrCriminalData":                "NO",
				"internationalTransfers":               false,
				"dataProtectionImpactAssessmentNeeded": "NOT_NEEDED",
				"transferImpactAssessmentNeeded":       "NOT_NEEDED",
				"role":                                 "CONTROLLER",
				"lastReviewDate":                       "2024-01-15T00:00:00Z",
				"nextReviewDate":                       "2025-01-15T00:00:00Z",
			},
		}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.CreateProcessingActivity.ProcessingActivityEdge.Node.LastReviewDate)
		require.NotNil(t, result.CreateProcessingActivity.ProcessingActivityEdge.Node.NextReviewDate)
	})

	t.Run("update review dates", func(t *testing.T) {
		paID := factory.NewProcessingActivity(owner).
			WithName("Review Date Update Test").
			Create()

		query := `
			mutation UpdateProcessingActivity($input: UpdateProcessingActivityInput!) {
				updateProcessingActivity(input: $input) {
					processingActivity {
						id
						lastReviewDate
						nextReviewDate
					}
				}
			}
		`

		var result struct {
			UpdateProcessingActivity struct {
				ProcessingActivity struct {
					ID             string  `json:"id"`
					LastReviewDate *string `json:"lastReviewDate"`
					NextReviewDate *string `json:"nextReviewDate"`
				} `json:"processingActivity"`
			} `json:"updateProcessingActivity"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":             paID,
				"lastReviewDate": "2024-06-01T00:00:00Z",
				"nextReviewDate": "2025-06-01T00:00:00Z",
			},
		}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.UpdateProcessingActivity.ProcessingActivity.LastReviewDate)
		require.NotNil(t, result.UpdateProcessingActivity.ProcessingActivity.NextReviewDate)
	})
}

func TestProcessingActivity_DPIA(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("create DPIA", func(t *testing.T) {
		paID := factory.NewProcessingActivity(owner).
			WithName("DPIA Create Test").
			Create()

		query := `
			mutation CreateDataProtectionImpactAssessment($input: CreateDataProtectionImpactAssessmentInput!) {
				createDataProtectionImpactAssessment(input: $input) {
					dataProtectionImpactAssessment {
						id
						description
						necessityAndProportionality
						potentialRisk
						mitigations
						residualRisk
					}
				}
			}
		`

		var result struct {
			CreateDataProtectionImpactAssessment struct {
				DataProtectionImpactAssessment struct {
					ID                          string  `json:"id"`
					Description                 *string `json:"description"`
					NecessityAndProportionality *string `json:"necessityAndProportionality"`
					PotentialRisk               *string `json:"potentialRisk"`
					Mitigations                 *string `json:"mitigations"`
					ResidualRisk                *string `json:"residualRisk"`
				} `json:"dataProtectionImpactAssessment"`
			} `json:"createDataProtectionImpactAssessment"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"processingActivityId":        paID,
				"description":                 "Test DPIA description",
				"necessityAndProportionality": "Test necessity",
				"potentialRisk":               "Test risk",
				"mitigations":                 "Test mitigations",
				"residualRisk":                "LOW",
			},
		}, &result)
		require.NoError(t, err)
		assert.NotEmpty(t, result.CreateDataProtectionImpactAssessment.DataProtectionImpactAssessment.ID)
		assert.Equal(t, "Test DPIA description", *result.CreateDataProtectionImpactAssessment.DataProtectionImpactAssessment.Description)
		assert.Equal(t, "LOW", *result.CreateDataProtectionImpactAssessment.DataProtectionImpactAssessment.ResidualRisk)
	})

	t.Run("read DPIA via processing activity", func(t *testing.T) {
		paID := factory.NewProcessingActivity(owner).
			WithName("DPIA Read Test").
			Create()

		createQuery := `
			mutation CreateDataProtectionImpactAssessment($input: CreateDataProtectionImpactAssessmentInput!) {
				createDataProtectionImpactAssessment(input: $input) {
					dataProtectionImpactAssessment { id }
				}
			}
		`
		_, err := owner.Do(createQuery, map[string]any{
			"input": map[string]any{
				"processingActivityId": paID,
				"description":          "Read test DPIA",
				"residualRisk":         "MEDIUM",
			},
		})
		require.NoError(t, err)

		readQuery := `
			query($id: ID!) {
				node(id: $id) {
					... on ProcessingActivity {
						id
						dataProtectionImpactAssessment {
							id
							description
							residualRisk
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID                             string `json:"id"`
				DataProtectionImpactAssessment *struct {
					ID           string  `json:"id"`
					Description  *string `json:"description"`
					ResidualRisk *string `json:"residualRisk"`
				} `json:"dataProtectionImpactAssessment"`
			} `json:"node"`
		}

		err = owner.Execute(readQuery, map[string]any{"id": paID}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.Node.DataProtectionImpactAssessment)
		assert.Equal(t, "Read test DPIA", *result.Node.DataProtectionImpactAssessment.Description)
		assert.Equal(t, "MEDIUM", *result.Node.DataProtectionImpactAssessment.ResidualRisk)
	})

	t.Run("update DPIA", func(t *testing.T) {
		paID := factory.NewProcessingActivity(owner).
			WithName("DPIA Update Test").
			Create()

		createQuery := `
			mutation CreateDataProtectionImpactAssessment($input: CreateDataProtectionImpactAssessmentInput!) {
				createDataProtectionImpactAssessment(input: $input) {
					dataProtectionImpactAssessment { id }
				}
			}
		`
		var createResult struct {
			CreateDataProtectionImpactAssessment struct {
				DataProtectionImpactAssessment struct {
					ID string `json:"id"`
				} `json:"dataProtectionImpactAssessment"`
			} `json:"createDataProtectionImpactAssessment"`
		}
		err := owner.Execute(createQuery, map[string]any{
			"input": map[string]any{
				"processingActivityId": paID,
				"description":          "Original description",
				"residualRisk":         "LOW",
			},
		}, &createResult)
		require.NoError(t, err)
		dpiaID := createResult.CreateDataProtectionImpactAssessment.DataProtectionImpactAssessment.ID

		updateQuery := `
			mutation UpdateDataProtectionImpactAssessment($input: UpdateDataProtectionImpactAssessmentInput!) {
				updateDataProtectionImpactAssessment(input: $input) {
					dataProtectionImpactAssessment {
						id
						description
						residualRisk
					}
				}
			}
		`

		var updateResult struct {
			UpdateDataProtectionImpactAssessment struct {
				DataProtectionImpactAssessment struct {
					ID           string  `json:"id"`
					Description  *string `json:"description"`
					ResidualRisk *string `json:"residualRisk"`
				} `json:"dataProtectionImpactAssessment"`
			} `json:"updateDataProtectionImpactAssessment"`
		}

		err = owner.Execute(updateQuery, map[string]any{
			"input": map[string]any{
				"id":           dpiaID,
				"description":  "Updated description",
				"residualRisk": "HIGH",
			},
		}, &updateResult)
		require.NoError(t, err)
		assert.Equal(t, "Updated description", *updateResult.UpdateDataProtectionImpactAssessment.DataProtectionImpactAssessment.Description)
		assert.Equal(t, "HIGH", *updateResult.UpdateDataProtectionImpactAssessment.DataProtectionImpactAssessment.ResidualRisk)
	})

	t.Run("delete DPIA", func(t *testing.T) {
		paID := factory.NewProcessingActivity(owner).
			WithName("DPIA Delete Test").
			Create()

		createQuery := `
			mutation CreateDataProtectionImpactAssessment($input: CreateDataProtectionImpactAssessmentInput!) {
				createDataProtectionImpactAssessment(input: $input) {
					dataProtectionImpactAssessment { id }
				}
			}
		`
		var createResult struct {
			CreateDataProtectionImpactAssessment struct {
				DataProtectionImpactAssessment struct {
					ID string `json:"id"`
				} `json:"dataProtectionImpactAssessment"`
			} `json:"createDataProtectionImpactAssessment"`
		}
		err := owner.Execute(createQuery, map[string]any{
			"input": map[string]any{
				"processingActivityId": paID,
				"description":          "To be deleted",
			},
		}, &createResult)
		require.NoError(t, err)
		dpiaID := createResult.CreateDataProtectionImpactAssessment.DataProtectionImpactAssessment.ID

		deleteQuery := `
			mutation DeleteDataProtectionImpactAssessment($input: DeleteDataProtectionImpactAssessmentInput!) {
				deleteDataProtectionImpactAssessment(input: $input) {
					deletedDataProtectionImpactAssessmentId
				}
			}
		`

		var deleteResult struct {
			DeleteDataProtectionImpactAssessment struct {
				DeletedDataProtectionImpactAssessmentID string `json:"deletedDataProtectionImpactAssessmentId"`
			} `json:"deleteDataProtectionImpactAssessment"`
		}

		err = owner.Execute(deleteQuery, map[string]any{
			"input": map[string]any{
				"dataProtectionImpactAssessmentId": dpiaID,
			},
		}, &deleteResult)
		require.NoError(t, err)
		assert.Equal(t, dpiaID, deleteResult.DeleteDataProtectionImpactAssessment.DeletedDataProtectionImpactAssessmentID)

		readQuery := `
			query($id: ID!) {
				node(id: $id) {
					... on ProcessingActivity {
						dataProtectionImpactAssessment { id }
					}
				}
			}
		`
		var readResult struct {
			Node struct {
				DataProtectionImpactAssessment *struct {
					ID string `json:"id"`
				} `json:"dataProtectionImpactAssessment"`
			} `json:"node"`
		}
		err = owner.Execute(readQuery, map[string]any{"id": paID}, &readResult)
		require.NoError(t, err)
		assert.Nil(t, readResult.Node.DataProtectionImpactAssessment)
	})

	t.Run("DPIA residual risk values", func(t *testing.T) {
		residualRisks := []string{"LOW", "MEDIUM", "HIGH"}

		for _, risk := range residualRisks {
			t.Run(risk, func(t *testing.T) {
				paID := factory.NewProcessingActivity(owner).
					WithName(fmt.Sprintf("DPIA Risk %s", risk)).
					Create()

				query := `
					mutation CreateDataProtectionImpactAssessment($input: CreateDataProtectionImpactAssessmentInput!) {
						createDataProtectionImpactAssessment(input: $input) {
							dataProtectionImpactAssessment {
								id
								residualRisk
							}
						}
					}
				`

				var result struct {
					CreateDataProtectionImpactAssessment struct {
						DataProtectionImpactAssessment struct {
							ID           string  `json:"id"`
							ResidualRisk *string `json:"residualRisk"`
						} `json:"dataProtectionImpactAssessment"`
					} `json:"createDataProtectionImpactAssessment"`
				}

				err := owner.Execute(query, map[string]any{
					"input": map[string]any{
						"processingActivityId": paID,
						"residualRisk":         risk,
					},
				}, &result)
				require.NoError(t, err)
				assert.Equal(t, risk, *result.CreateDataProtectionImpactAssessment.DataProtectionImpactAssessment.ResidualRisk)
			})
		}
	})
}

func TestProcessingActivity_TIA(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("create TIA", func(t *testing.T) {
		paID := factory.NewProcessingActivity(owner).
			WithName("TIA Create Test").
			Create()

		query := `
			mutation CreateTransferImpactAssessment($input: CreateTransferImpactAssessmentInput!) {
				createTransferImpactAssessment(input: $input) {
					transferImpactAssessment {
						id
						dataSubjects
						legalMechanism
						transfer
						localLawRisk
						supplementaryMeasures
					}
				}
			}
		`

		var result struct {
			CreateTransferImpactAssessment struct {
				TransferImpactAssessment struct {
					ID                    string  `json:"id"`
					DataSubjects          *string `json:"dataSubjects"`
					LegalMechanism        *string `json:"legalMechanism"`
					Transfer              *string `json:"transfer"`
					LocalLawRisk          *string `json:"localLawRisk"`
					SupplementaryMeasures *string `json:"supplementaryMeasures"`
				} `json:"transferImpactAssessment"`
			} `json:"createTransferImpactAssessment"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"processingActivityId":  paID,
				"dataSubjects":          "EU customers",
				"legalMechanism":        "Standard Contractual Clauses",
				"transfer":              "EU to US",
				"localLawRisk":          "Moderate risk due to surveillance laws",
				"supplementaryMeasures": "Encryption at rest and in transit",
			},
		}, &result)
		require.NoError(t, err)
		assert.NotEmpty(t, result.CreateTransferImpactAssessment.TransferImpactAssessment.ID)
		assert.Equal(t, "EU customers", *result.CreateTransferImpactAssessment.TransferImpactAssessment.DataSubjects)
		assert.Equal(t, "Standard Contractual Clauses", *result.CreateTransferImpactAssessment.TransferImpactAssessment.LegalMechanism)
	})

	t.Run("read TIA via processing activity", func(t *testing.T) {
		paID := factory.NewProcessingActivity(owner).
			WithName("TIA Read Test").
			Create()

		createQuery := `
			mutation CreateTransferImpactAssessment($input: CreateTransferImpactAssessmentInput!) {
				createTransferImpactAssessment(input: $input) {
					transferImpactAssessment { id }
				}
			}
		`
		_, err := owner.Do(createQuery, map[string]any{
			"input": map[string]any{
				"processingActivityId": paID,
				"dataSubjects":         "Read test subjects",
				"transfer":             "Read test transfer",
			},
		})
		require.NoError(t, err)

		readQuery := `
			query($id: ID!) {
				node(id: $id) {
					... on ProcessingActivity {
						id
						transferImpactAssessment {
							id
							dataSubjects
							transfer
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID                       string `json:"id"`
				TransferImpactAssessment *struct {
					ID           string  `json:"id"`
					DataSubjects *string `json:"dataSubjects"`
					Transfer     *string `json:"transfer"`
				} `json:"transferImpactAssessment"`
			} `json:"node"`
		}

		err = owner.Execute(readQuery, map[string]any{"id": paID}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.Node.TransferImpactAssessment)
		assert.Equal(t, "Read test subjects", *result.Node.TransferImpactAssessment.DataSubjects)
		assert.Equal(t, "Read test transfer", *result.Node.TransferImpactAssessment.Transfer)
	})

	t.Run("update TIA", func(t *testing.T) {
		paID := factory.NewProcessingActivity(owner).
			WithName("TIA Update Test").
			Create()

		createQuery := `
			mutation CreateTransferImpactAssessment($input: CreateTransferImpactAssessmentInput!) {
				createTransferImpactAssessment(input: $input) {
					transferImpactAssessment { id }
				}
			}
		`
		var createResult struct {
			CreateTransferImpactAssessment struct {
				TransferImpactAssessment struct {
					ID string `json:"id"`
				} `json:"transferImpactAssessment"`
			} `json:"createTransferImpactAssessment"`
		}
		err := owner.Execute(createQuery, map[string]any{
			"input": map[string]any{
				"processingActivityId": paID,
				"dataSubjects":         "Original subjects",
				"transfer":             "Original transfer",
			},
		}, &createResult)
		require.NoError(t, err)
		tiaID := createResult.CreateTransferImpactAssessment.TransferImpactAssessment.ID

		updateQuery := `
			mutation UpdateTransferImpactAssessment($input: UpdateTransferImpactAssessmentInput!) {
				updateTransferImpactAssessment(input: $input) {
					transferImpactAssessment {
						id
						dataSubjects
						transfer
						legalMechanism
					}
				}
			}
		`

		var updateResult struct {
			UpdateTransferImpactAssessment struct {
				TransferImpactAssessment struct {
					ID             string  `json:"id"`
					DataSubjects   *string `json:"dataSubjects"`
					Transfer       *string `json:"transfer"`
					LegalMechanism *string `json:"legalMechanism"`
				} `json:"transferImpactAssessment"`
			} `json:"updateTransferImpactAssessment"`
		}

		err = owner.Execute(updateQuery, map[string]any{
			"input": map[string]any{
				"id":             tiaID,
				"dataSubjects":   "Updated subjects",
				"transfer":       "Updated transfer",
				"legalMechanism": "Binding Corporate Rules",
			},
		}, &updateResult)
		require.NoError(t, err)
		assert.Equal(t, "Updated subjects", *updateResult.UpdateTransferImpactAssessment.TransferImpactAssessment.DataSubjects)
		assert.Equal(t, "Updated transfer", *updateResult.UpdateTransferImpactAssessment.TransferImpactAssessment.Transfer)
		assert.Equal(t, "Binding Corporate Rules", *updateResult.UpdateTransferImpactAssessment.TransferImpactAssessment.LegalMechanism)
	})

	t.Run("delete TIA", func(t *testing.T) {
		paID := factory.NewProcessingActivity(owner).
			WithName("TIA Delete Test").
			Create()

		createQuery := `
			mutation CreateTransferImpactAssessment($input: CreateTransferImpactAssessmentInput!) {
				createTransferImpactAssessment(input: $input) {
					transferImpactAssessment { id }
				}
			}
		`
		var createResult struct {
			CreateTransferImpactAssessment struct {
				TransferImpactAssessment struct {
					ID string `json:"id"`
				} `json:"transferImpactAssessment"`
			} `json:"createTransferImpactAssessment"`
		}
		err := owner.Execute(createQuery, map[string]any{
			"input": map[string]any{
				"processingActivityId": paID,
				"dataSubjects":         "To be deleted",
			},
		}, &createResult)
		require.NoError(t, err)
		tiaID := createResult.CreateTransferImpactAssessment.TransferImpactAssessment.ID

		deleteQuery := `
			mutation DeleteTransferImpactAssessment($input: DeleteTransferImpactAssessmentInput!) {
				deleteTransferImpactAssessment(input: $input) {
					deletedTransferImpactAssessmentId
				}
			}
		`

		var deleteResult struct {
			DeleteTransferImpactAssessment struct {
				DeletedTransferImpactAssessmentID string `json:"deletedTransferImpactAssessmentId"`
			} `json:"deleteTransferImpactAssessment"`
		}

		err = owner.Execute(deleteQuery, map[string]any{
			"input": map[string]any{
				"transferImpactAssessmentId": tiaID,
			},
		}, &deleteResult)
		require.NoError(t, err)
		assert.Equal(t, tiaID, deleteResult.DeleteTransferImpactAssessment.DeletedTransferImpactAssessmentID)

		readQuery := `
			query($id: ID!) {
				node(id: $id) {
					... on ProcessingActivity {
						transferImpactAssessment { id }
					}
				}
			}
		`
		var readResult struct {
			Node struct {
				TransferImpactAssessment *struct {
					ID string `json:"id"`
				} `json:"transferImpactAssessment"`
			} `json:"node"`
		}
		err = owner.Execute(readQuery, map[string]any{"id": paID}, &readResult)
		require.NoError(t, err)
		assert.Nil(t, readResult.Node.TransferImpactAssessment)
	})
}

func TestProcessingActivity_DPIA_RBAC(t *testing.T) {
	t.Parallel()

	t.Run("owner can manage DPIA", func(t *testing.T) {
		owner := testutil.NewClient(t, testutil.RoleOwner)
		paID := factory.NewProcessingActivity(owner).WithName("DPIA RBAC Owner Test").Create()

		var createResult struct {
			CreateDataProtectionImpactAssessment struct {
				DataProtectionImpactAssessment struct {
					ID string `json:"id"`
				} `json:"dataProtectionImpactAssessment"`
			} `json:"createDataProtectionImpactAssessment"`
		}
		err := owner.Execute(`
			mutation($input: CreateDataProtectionImpactAssessmentInput!) {
				createDataProtectionImpactAssessment(input: $input) {
					dataProtectionImpactAssessment { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"processingActivityId": paID,
				"description":          "Owner DPIA",
			},
		}, &createResult)
		require.NoError(t, err, "owner should be able to create DPIA")
		dpiaID := createResult.CreateDataProtectionImpactAssessment.DataProtectionImpactAssessment.ID

		_, err = owner.Do(`
			mutation($input: UpdateDataProtectionImpactAssessmentInput!) {
				updateDataProtectionImpactAssessment(input: $input) {
					dataProtectionImpactAssessment { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":          dpiaID,
				"description": "Updated by owner",
			},
		})
		require.NoError(t, err, "owner should be able to update DPIA")

		_, err = owner.Do(`
			mutation($input: DeleteDataProtectionImpactAssessmentInput!) {
				deleteDataProtectionImpactAssessment(input: $input) {
					deletedDataProtectionImpactAssessmentId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"dataProtectionImpactAssessmentId": dpiaID,
			},
		})
		require.NoError(t, err, "owner should be able to delete DPIA")
	})

	t.Run("admin can manage DPIA", func(t *testing.T) {
		owner := testutil.NewClient(t, testutil.RoleOwner)
		admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
		paID := factory.NewProcessingActivity(owner).WithName("DPIA RBAC Admin Test").Create()

		var createResult struct {
			CreateDataProtectionImpactAssessment struct {
				DataProtectionImpactAssessment struct {
					ID string `json:"id"`
				} `json:"dataProtectionImpactAssessment"`
			} `json:"createDataProtectionImpactAssessment"`
		}
		err := admin.Execute(`
			mutation($input: CreateDataProtectionImpactAssessmentInput!) {
				createDataProtectionImpactAssessment(input: $input) {
					dataProtectionImpactAssessment { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"processingActivityId": paID,
				"description":          "Admin DPIA",
			},
		}, &createResult)
		require.NoError(t, err, "admin should be able to create DPIA")
		dpiaID := createResult.CreateDataProtectionImpactAssessment.DataProtectionImpactAssessment.ID

		_, err = admin.Do(`
			mutation($input: UpdateDataProtectionImpactAssessmentInput!) {
				updateDataProtectionImpactAssessment(input: $input) {
					dataProtectionImpactAssessment { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":          dpiaID,
				"description": "Updated by admin",
			},
		})
		require.NoError(t, err, "admin should be able to update DPIA")

		_, err = admin.Do(`
			mutation($input: DeleteDataProtectionImpactAssessmentInput!) {
				deleteDataProtectionImpactAssessment(input: $input) {
					deletedDataProtectionImpactAssessmentId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"dataProtectionImpactAssessmentId": dpiaID,
			},
		})
		require.NoError(t, err, "admin should be able to delete DPIA")
	})

	t.Run("viewer cannot manage DPIA", func(t *testing.T) {
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		paID := factory.NewProcessingActivity(owner).WithName("DPIA RBAC Viewer Test").Create()

		_, err := viewer.Do(`
			mutation($input: CreateDataProtectionImpactAssessmentInput!) {
				createDataProtectionImpactAssessment(input: $input) {
					dataProtectionImpactAssessment { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"processingActivityId": paID,
				"description":          "Viewer DPIA",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to create DPIA")

		var createResult struct {
			CreateDataProtectionImpactAssessment struct {
				DataProtectionImpactAssessment struct {
					ID string `json:"id"`
				} `json:"dataProtectionImpactAssessment"`
			} `json:"createDataProtectionImpactAssessment"`
		}
		err = owner.Execute(`
			mutation($input: CreateDataProtectionImpactAssessmentInput!) {
				createDataProtectionImpactAssessment(input: $input) {
					dataProtectionImpactAssessment { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"processingActivityId": paID,
				"description":          "Owner created DPIA",
			},
		}, &createResult)
		require.NoError(t, err)
		dpiaID := createResult.CreateDataProtectionImpactAssessment.DataProtectionImpactAssessment.ID

		_, err = viewer.Do(`
			mutation($input: UpdateDataProtectionImpactAssessmentInput!) {
				updateDataProtectionImpactAssessment(input: $input) {
					dataProtectionImpactAssessment { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":          dpiaID,
				"description": "Updated by viewer",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to update DPIA")

		_, err = viewer.Do(`
			mutation($input: DeleteDataProtectionImpactAssessmentInput!) {
				deleteDataProtectionImpactAssessment(input: $input) {
					deletedDataProtectionImpactAssessmentId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"dataProtectionImpactAssessmentId": dpiaID,
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to delete DPIA")
	})
}

func TestProcessingActivity_TIA_RBAC(t *testing.T) {
	t.Parallel()

	t.Run("owner can manage TIA", func(t *testing.T) {
		owner := testutil.NewClient(t, testutil.RoleOwner)
		paID := factory.NewProcessingActivity(owner).WithName("TIA RBAC Owner Test").Create()

		var createResult struct {
			CreateTransferImpactAssessment struct {
				TransferImpactAssessment struct {
					ID string `json:"id"`
				} `json:"transferImpactAssessment"`
			} `json:"createTransferImpactAssessment"`
		}
		err := owner.Execute(`
			mutation($input: CreateTransferImpactAssessmentInput!) {
				createTransferImpactAssessment(input: $input) {
					transferImpactAssessment { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"processingActivityId": paID,
				"dataSubjects":         "Owner TIA subjects",
			},
		}, &createResult)
		require.NoError(t, err, "owner should be able to create TIA")
		tiaID := createResult.CreateTransferImpactAssessment.TransferImpactAssessment.ID

		_, err = owner.Do(`
			mutation($input: UpdateTransferImpactAssessmentInput!) {
				updateTransferImpactAssessment(input: $input) {
					transferImpactAssessment { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":           tiaID,
				"dataSubjects": "Updated by owner",
			},
		})
		require.NoError(t, err, "owner should be able to update TIA")

		_, err = owner.Do(`
			mutation($input: DeleteTransferImpactAssessmentInput!) {
				deleteTransferImpactAssessment(input: $input) {
					deletedTransferImpactAssessmentId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"transferImpactAssessmentId": tiaID,
			},
		})
		require.NoError(t, err, "owner should be able to delete TIA")
	})

	t.Run("admin can manage TIA", func(t *testing.T) {
		owner := testutil.NewClient(t, testutil.RoleOwner)
		admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
		paID := factory.NewProcessingActivity(owner).WithName("TIA RBAC Admin Test").Create()

		var createResult struct {
			CreateTransferImpactAssessment struct {
				TransferImpactAssessment struct {
					ID string `json:"id"`
				} `json:"transferImpactAssessment"`
			} `json:"createTransferImpactAssessment"`
		}
		err := admin.Execute(`
			mutation($input: CreateTransferImpactAssessmentInput!) {
				createTransferImpactAssessment(input: $input) {
					transferImpactAssessment { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"processingActivityId": paID,
				"dataSubjects":         "Admin TIA subjects",
			},
		}, &createResult)
		require.NoError(t, err, "admin should be able to create TIA")
		tiaID := createResult.CreateTransferImpactAssessment.TransferImpactAssessment.ID

		_, err = admin.Do(`
			mutation($input: UpdateTransferImpactAssessmentInput!) {
				updateTransferImpactAssessment(input: $input) {
					transferImpactAssessment { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":           tiaID,
				"dataSubjects": "Updated by admin",
			},
		})
		require.NoError(t, err, "admin should be able to update TIA")

		_, err = admin.Do(`
			mutation($input: DeleteTransferImpactAssessmentInput!) {
				deleteTransferImpactAssessment(input: $input) {
					deletedTransferImpactAssessmentId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"transferImpactAssessmentId": tiaID,
			},
		})
		require.NoError(t, err, "admin should be able to delete TIA")
	})

	t.Run("viewer cannot manage TIA", func(t *testing.T) {
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		paID := factory.NewProcessingActivity(owner).WithName("TIA RBAC Viewer Test").Create()

		_, err := viewer.Do(`
			mutation($input: CreateTransferImpactAssessmentInput!) {
				createTransferImpactAssessment(input: $input) {
					transferImpactAssessment { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"processingActivityId": paID,
				"dataSubjects":         "Viewer TIA subjects",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to create TIA")

		var createResult struct {
			CreateTransferImpactAssessment struct {
				TransferImpactAssessment struct {
					ID string `json:"id"`
				} `json:"transferImpactAssessment"`
			} `json:"createTransferImpactAssessment"`
		}
		err = owner.Execute(`
			mutation($input: CreateTransferImpactAssessmentInput!) {
				createTransferImpactAssessment(input: $input) {
					transferImpactAssessment { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"processingActivityId": paID,
				"dataSubjects":         "Owner created TIA",
			},
		}, &createResult)
		require.NoError(t, err)
		tiaID := createResult.CreateTransferImpactAssessment.TransferImpactAssessment.ID

		_, err = viewer.Do(`
			mutation($input: UpdateTransferImpactAssessmentInput!) {
				updateTransferImpactAssessment(input: $input) {
					transferImpactAssessment { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":           tiaID,
				"dataSubjects": "Updated by viewer",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to update TIA")

		_, err = viewer.Do(`
			mutation($input: DeleteTransferImpactAssessmentInput!) {
				deleteTransferImpactAssessment(input: $input) {
					deletedTransferImpactAssessmentId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"transferImpactAssessmentId": tiaID,
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to delete TIA")
	})
}

func TestProcessingActivity_Snapshot_DPIA_TIA(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("snapshot includes DPIA and TIA", func(t *testing.T) {
		paID := factory.NewProcessingActivity(owner).
			WithName("Snapshot DPIA TIA Test").
			Create()

		var dpiaResult struct {
			CreateDataProtectionImpactAssessment struct {
				DataProtectionImpactAssessment struct {
					ID string `json:"id"`
				} `json:"dataProtectionImpactAssessment"`
			} `json:"createDataProtectionImpactAssessment"`
		}
		err := owner.Execute(`
			mutation($input: CreateDataProtectionImpactAssessmentInput!) {
				createDataProtectionImpactAssessment(input: $input) {
					dataProtectionImpactAssessment { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"processingActivityId": paID,
				"description":          "DPIA for snapshot",
				"residualRisk":         "MEDIUM",
			},
		}, &dpiaResult)
		require.NoError(t, err)

		var tiaResult struct {
			CreateTransferImpactAssessment struct {
				TransferImpactAssessment struct {
					ID string `json:"id"`
				} `json:"transferImpactAssessment"`
			} `json:"createTransferImpactAssessment"`
		}
		err = owner.Execute(`
			mutation($input: CreateTransferImpactAssessmentInput!) {
				createTransferImpactAssessment(input: $input) {
					transferImpactAssessment { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"processingActivityId": paID,
				"dataSubjects":         "TIA subjects for snapshot",
				"transfer":             "EU to US",
			},
		}, &tiaResult)
		require.NoError(t, err)

		var snapshotResult struct {
			CreateSnapshot struct {
				SnapshotEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"snapshotEdge"`
			} `json:"createSnapshot"`
		}
		err = owner.Execute(`
			mutation($input: CreateSnapshotInput!) {
				createSnapshot(input: $input) {
					snapshotEdge {
						node { id }
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           fmt.Sprintf("PA Snapshot Test %d", time.Now().UnixNano()),
				"type":           "PROCESSING_ACTIVITIES",
			},
		}, &snapshotResult)
		require.NoError(t, err)
		snapshotID := snapshotResult.CreateSnapshot.SnapshotEdge.Node.ID

		var queryResult struct {
			Node struct {
				ProcessingActivities struct {
					Edges []struct {
						Node struct {
							ID                             string `json:"id"`
							Name                           string `json:"name"`
							DataProtectionImpactAssessment *struct {
								ID           string  `json:"id"`
								Description  *string `json:"description"`
								ResidualRisk *string `json:"residualRisk"`
							} `json:"dataProtectionImpactAssessment"`
							TransferImpactAssessment *struct {
								ID           string  `json:"id"`
								DataSubjects *string `json:"dataSubjects"`
								Transfer     *string `json:"transfer"`
							} `json:"transferImpactAssessment"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"processingActivities"`
			} `json:"node"`
		}
		err = owner.Execute(`
			query($orgId: ID!, $snapshotId: ID!) {
				node(id: $orgId) {
					... on Organization {
						processingActivities(first: 100, filter: { snapshotId: $snapshotId }) {
							edges {
								node {
									id
									name
									dataProtectionImpactAssessment {
										id
										description
										residualRisk
									}
									transferImpactAssessment {
										id
										dataSubjects
										transfer
									}
								}
							}
						}
					}
				}
			}
		`, map[string]any{
			"orgId":      owner.GetOrganizationID().String(),
			"snapshotId": snapshotID,
		}, &queryResult)
		require.NoError(t, err)

		var foundPA bool
		for _, edge := range queryResult.Node.ProcessingActivities.Edges {
			if edge.Node.Name == "Snapshot DPIA TIA Test" {
				foundPA = true
				require.NotNil(t, edge.Node.DataProtectionImpactAssessment, "DPIA should be included in snapshot")
				assert.Equal(t, "DPIA for snapshot", *edge.Node.DataProtectionImpactAssessment.Description)
				assert.Equal(t, "MEDIUM", *edge.Node.DataProtectionImpactAssessment.ResidualRisk)

				require.NotNil(t, edge.Node.TransferImpactAssessment, "TIA should be included in snapshot")
				assert.Equal(t, "TIA subjects for snapshot", *edge.Node.TransferImpactAssessment.DataSubjects)
				assert.Equal(t, "EU to US", *edge.Node.TransferImpactAssessment.Transfer)
				break
			}
		}
		assert.True(t, foundPA, "Processing activity should be found in snapshot")
	})

	t.Run("snapshot DPIA and TIA are independent of source", func(t *testing.T) {
		paID := factory.NewProcessingActivity(owner).
			WithName("Snapshot Independence Test").
			Create()

		err := owner.Execute(`
			mutation($input: CreateDataProtectionImpactAssessmentInput!) {
				createDataProtectionImpactAssessment(input: $input) {
					dataProtectionImpactAssessment { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"processingActivityId": paID,
				"description":          "Original DPIA",
				"residualRisk":         "LOW",
			},
		}, nil)
		require.NoError(t, err)

		var snapshotResult struct {
			CreateSnapshot struct {
				SnapshotEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"snapshotEdge"`
			} `json:"createSnapshot"`
		}
		err = owner.Execute(`
			mutation($input: CreateSnapshotInput!) {
				createSnapshot(input: $input) {
					snapshotEdge {
						node { id }
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           fmt.Sprintf("PA Independence Test %d", time.Now().UnixNano()),
				"type":           "PROCESSING_ACTIVITIES",
			},
		}, &snapshotResult)
		require.NoError(t, err)
		snapshotID := snapshotResult.CreateSnapshot.SnapshotEdge.Node.ID

		var sourceResult struct {
			Node struct {
				DataProtectionImpactAssessment *struct {
					ID           string  `json:"id"`
					Description  *string `json:"description"`
					ResidualRisk *string `json:"residualRisk"`
				} `json:"dataProtectionImpactAssessment"`
			} `json:"node"`
		}
		err = owner.Execute(`
			query($id: ID!) {
				node(id: $id) {
					... on ProcessingActivity {
						dataProtectionImpactAssessment {
							id
							description
							residualRisk
						}
					}
				}
			}
		`, map[string]any{"id": paID}, &sourceResult)
		require.NoError(t, err)
		require.NotNil(t, sourceResult.Node.DataProtectionImpactAssessment)
		sourceDpiaID := sourceResult.Node.DataProtectionImpactAssessment.ID

		_, err = owner.Do(`
			mutation($input: UpdateDataProtectionImpactAssessmentInput!) {
				updateDataProtectionImpactAssessment(input: $input) {
					dataProtectionImpactAssessment { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":           sourceDpiaID,
				"description":  "Updated after snapshot",
				"residualRisk": "HIGH",
			},
		})
		require.NoError(t, err)

		var snapshotQueryResult struct {
			Node struct {
				ProcessingActivities struct {
					Edges []struct {
						Node struct {
							Name                           string `json:"name"`
							DataProtectionImpactAssessment *struct {
								Description  *string `json:"description"`
								ResidualRisk *string `json:"residualRisk"`
							} `json:"dataProtectionImpactAssessment"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"processingActivities"`
			} `json:"node"`
		}
		err = owner.Execute(`
			query($orgId: ID!, $snapshotId: ID!) {
				node(id: $orgId) {
					... on Organization {
						processingActivities(first: 100, filter: { snapshotId: $snapshotId }) {
							edges {
								node {
									name
									dataProtectionImpactAssessment {
										description
										residualRisk
									}
								}
							}
						}
					}
				}
			}
		`, map[string]any{
			"orgId":      owner.GetOrganizationID().String(),
			"snapshotId": snapshotID,
		}, &snapshotQueryResult)
		require.NoError(t, err)

		for _, edge := range snapshotQueryResult.Node.ProcessingActivities.Edges {
			if edge.Node.Name == "Snapshot Independence Test" {
				require.NotNil(t, edge.Node.DataProtectionImpactAssessment)
				assert.Equal(t, "Original DPIA", *edge.Node.DataProtectionImpactAssessment.Description, "Snapshot DPIA should retain original value")
				assert.Equal(t, "LOW", *edge.Node.DataProtectionImpactAssessment.ResidualRisk, "Snapshot DPIA should retain original value")
				break
			}
		}
	})
}

func TestProcessingActivity_ExportPDF(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("export processing activities PDF", func(t *testing.T) {
		_ = factory.NewProcessingActivity(owner).
			WithName("PA Export Test 1").
			WithLawfulBasis("CONSENT").
			Create()
		_ = factory.NewProcessingActivity(owner).
			WithName("PA Export Test 2").
			WithLawfulBasis("LEGITIMATE_INTEREST").
			Create()

		query := `
			mutation ExportProcessingActivitiesPDF($input: ExportProcessingActivitiesPDFInput!) {
				exportProcessingActivitiesPDF(input: $input) {
					data
				}
			}
		`

		var result struct {
			ExportProcessingActivitiesPDF struct {
				Data string `json:"data"`
			} `json:"exportProcessingActivitiesPDF"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"filter":         nil,
			},
		}, &result)
		require.NoError(t, err)
		assert.NotEmpty(t, result.ExportProcessingActivitiesPDF.Data)
		assert.Contains(t, result.ExportProcessingActivitiesPDF.Data, "data:application/pdf;base64,")
	})

	t.Run("export processing activities PDF with snapshot filter", func(t *testing.T) {
		_ = factory.NewProcessingActivity(owner).
			WithName("PA Snapshot Export Test").
			Create()

		// Create snapshot
		var snapshotResult struct {
			CreateSnapshot struct {
				SnapshotEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"snapshotEdge"`
			} `json:"createSnapshot"`
		}
		err := owner.Execute(`
			mutation($input: CreateSnapshotInput!) {
				createSnapshot(input: $input) {
					snapshotEdge {
						node { id }
					}
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           fmt.Sprintf("PA Export Snapshot Test %d", time.Now().UnixNano()),
				"type":           "PROCESSING_ACTIVITIES",
			},
		}, &snapshotResult)
		require.NoError(t, err)
		snapshotID := snapshotResult.CreateSnapshot.SnapshotEdge.Node.ID

		query := `
			mutation ExportProcessingActivitiesPDF($input: ExportProcessingActivitiesPDFInput!) {
				exportProcessingActivitiesPDF(input: $input) {
					data
				}
			}
		`

		var result struct {
			ExportProcessingActivitiesPDF struct {
				Data string `json:"data"`
			} `json:"exportProcessingActivitiesPDF"`
		}

		err = owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"filter": map[string]any{
					"snapshotId": snapshotID,
				},
			},
		}, &result)
		require.NoError(t, err)
		assert.NotEmpty(t, result.ExportProcessingActivitiesPDF.Data)
		assert.Contains(t, result.ExportProcessingActivitiesPDF.Data, "data:application/pdf;base64,")
	})

	t.Run("export fails with no processing activities", func(t *testing.T) {
		newOwner := testutil.NewClient(t, testutil.RoleOwner)

		query := `
			mutation ExportProcessingActivitiesPDF($input: ExportProcessingActivitiesPDFInput!) {
				exportProcessingActivitiesPDF(input: $input) {
					data
				}
			}
		`

		_, err := newOwner.Do(query, map[string]any{
			"input": map[string]any{
				"organizationId": newOwner.GetOrganizationID().String(),
				"filter":         nil,
			},
		})
		testutil.RequireErrorCode(t, err, "NOT_FOUND")
		assert.Contains(t, err.Error(), "no processing activities found")
	})
}

func TestDataProtectionImpactAssessment_ExportPDF(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("export DPIA PDF", func(t *testing.T) {
		pa1ID := factory.NewProcessingActivity(owner).
			WithName("DPIA Export Test PA 1").
			Create()
		pa2ID := factory.NewProcessingActivity(owner).
			WithName("DPIA Export Test PA 2").
			Create()

		_, err := owner.Do(`
			mutation($input: CreateDataProtectionImpactAssessmentInput!) {
				createDataProtectionImpactAssessment(input: $input) {
					dataProtectionImpactAssessment { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"processingActivityId": pa1ID,
				"description":          "DPIA 1 description",
				"residualRisk":         "LOW",
			},
		})
		require.NoError(t, err)

		_, err = owner.Do(`
			mutation($input: CreateDataProtectionImpactAssessmentInput!) {
				createDataProtectionImpactAssessment(input: $input) {
					dataProtectionImpactAssessment { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"processingActivityId": pa2ID,
				"description":          "DPIA 2 description",
				"residualRisk":         "MEDIUM",
			},
		})
		require.NoError(t, err)

		query := `
			mutation ExportDataProtectionImpactAssessmentsPDF($input: ExportDataProtectionImpactAssessmentsPDFInput!) {
				exportDataProtectionImpactAssessmentsPDF(input: $input) {
					data
				}
			}
		`

		var result struct {
			ExportDataProtectionImpactAssessmentsPDF struct {
				Data string `json:"data"`
			} `json:"exportDataProtectionImpactAssessmentsPDF"`
		}

		err = owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"filter":         nil,
			},
		}, &result)
		require.NoError(t, err)
		assert.NotEmpty(t, result.ExportDataProtectionImpactAssessmentsPDF.Data)
		assert.Contains(t, result.ExportDataProtectionImpactAssessmentsPDF.Data, "data:application/pdf;base64,")
	})

	t.Run("export fails with no DPIAs", func(t *testing.T) {
		newOwner := testutil.NewClient(t, testutil.RoleOwner)

		query := `
			mutation ExportDataProtectionImpactAssessmentsPDF($input: ExportDataProtectionImpactAssessmentsPDFInput!) {
				exportDataProtectionImpactAssessmentsPDF(input: $input) {
					data
				}
			}
		`

		_, err := newOwner.Do(query, map[string]any{
			"input": map[string]any{
				"organizationId": newOwner.GetOrganizationID().String(),
				"filter":         nil,
			},
		})
		testutil.RequireErrorCode(t, err, "NOT_FOUND")
		assert.Contains(t, err.Error(), "no data protection impact assessments found")
	})
}

func TestTransferImpactAssessment_ExportPDF(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("export TIA PDF", func(t *testing.T) {
		pa1ID := factory.NewProcessingActivity(owner).
			WithName("TIA Export Test PA 1").
			Create()
		pa2ID := factory.NewProcessingActivity(owner).
			WithName("TIA Export Test PA 2").
			Create()

		_, err := owner.Do(`
			mutation($input: CreateTransferImpactAssessmentInput!) {
				createTransferImpactAssessment(input: $input) {
					transferImpactAssessment { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"processingActivityId": pa1ID,
				"dataSubjects":         "TIA 1 subjects",
				"transfer":             "EU to US",
			},
		})
		require.NoError(t, err)

		_, err = owner.Do(`
			mutation($input: CreateTransferImpactAssessmentInput!) {
				createTransferImpactAssessment(input: $input) {
					transferImpactAssessment { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"processingActivityId": pa2ID,
				"dataSubjects":         "TIA 2 subjects",
				"transfer":             "EU to UK",
			},
		})
		require.NoError(t, err)

		query := `
			mutation ExportTransferImpactAssessmentsPDF($input: ExportTransferImpactAssessmentsPDFInput!) {
				exportTransferImpactAssessmentsPDF(input: $input) {
					data
				}
			}
		`

		var result struct {
			ExportTransferImpactAssessmentsPDF struct {
				Data string `json:"data"`
			} `json:"exportTransferImpactAssessmentsPDF"`
		}

		err = owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"filter":         nil,
			},
		}, &result)
		require.NoError(t, err)
		assert.NotEmpty(t, result.ExportTransferImpactAssessmentsPDF.Data)
		assert.Contains(t, result.ExportTransferImpactAssessmentsPDF.Data, "data:application/pdf;base64,")
	})

	t.Run("export fails with no TIAs", func(t *testing.T) {
		newOwner := testutil.NewClient(t, testutil.RoleOwner)

		query := `
			mutation ExportTransferImpactAssessmentsPDF($input: ExportTransferImpactAssessmentsPDFInput!) {
				exportTransferImpactAssessmentsPDF(input: $input) {
					data
				}
			}
		`

		_, err := newOwner.Do(query, map[string]any{
			"input": map[string]any{
				"organizationId": newOwner.GetOrganizationID().String(),
				"filter":         nil,
			},
		})
		testutil.RequireErrorCode(t, err, "NOT_FOUND")
		assert.Contains(t, err.Error(), "no transfer impact assessments found")
	})
}

func TestProcessingActivity_ExportPDF_RBAC(t *testing.T) {
	t.Parallel()

	t.Run("owner can export PDF", func(t *testing.T) {
		owner := testutil.NewClient(t, testutil.RoleOwner)
		_ = factory.NewProcessingActivity(owner).WithName("RBAC Export Test").Create()

		query := `
			mutation ExportProcessingActivitiesPDF($input: ExportProcessingActivitiesPDFInput!) {
				exportProcessingActivitiesPDF(input: $input) {
					data
				}
			}
		`

		var result struct {
			ExportProcessingActivitiesPDF struct {
				Data string `json:"data"`
			} `json:"exportProcessingActivitiesPDF"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"filter":         nil,
			},
		}, &result)
		require.NoError(t, err, "owner should be able to export PDF")
		assert.NotEmpty(t, result.ExportProcessingActivitiesPDF.Data)
	})

	t.Run("admin can export PDF", func(t *testing.T) {
		owner := testutil.NewClient(t, testutil.RoleOwner)
		admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
		_ = factory.NewProcessingActivity(owner).WithName("RBAC Export Test").Create()

		query := `
			mutation ExportProcessingActivitiesPDF($input: ExportProcessingActivitiesPDFInput!) {
				exportProcessingActivitiesPDF(input: $input) {
					data
				}
			}
		`

		var result struct {
			ExportProcessingActivitiesPDF struct {
				Data string `json:"data"`
			} `json:"exportProcessingActivitiesPDF"`
		}

		err := admin.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId": admin.GetOrganizationID().String(),
				"filter":         nil,
			},
		}, &result)
		require.NoError(t, err, "admin should be able to export PDF")
		assert.NotEmpty(t, result.ExportProcessingActivitiesPDF.Data)
	})

	t.Run("viewer can export PDF", func(t *testing.T) {
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		_ = factory.NewProcessingActivity(owner).WithName("RBAC Export Test").Create()

		query := `
			mutation ExportProcessingActivitiesPDF($input: ExportProcessingActivitiesPDFInput!) {
				exportProcessingActivitiesPDF(input: $input) {
					data
				}
			}
		`

		var result struct {
			ExportProcessingActivitiesPDF struct {
				Data string `json:"data"`
			} `json:"exportProcessingActivitiesPDF"`
		}

		err := viewer.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId": viewer.GetOrganizationID().String(),
				"filter":         nil,
			},
		}, &result)
		require.NoError(t, err, "viewer should be able to export PDF")
		assert.NotEmpty(t, result.ExportProcessingActivitiesPDF.Data)
	})
}
