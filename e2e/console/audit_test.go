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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestAudit_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Audit").Create()

	tests := []struct {
		name        string
		input       map[string]any
		assertField string
		assertValue string
	}{
		{
			name: "with full details",
			input: map[string]any{
				"name":  "SOC 2 Type II Audit 2025",
				"state": "NOT_STARTED",
			},
			assertField: "name",
			assertValue: "SOC 2 Type II Audit 2025",
		},
		{
			name: "with NOT_STARTED state",
			input: map[string]any{
				"name":  "Audit NOT_STARTED",
				"state": "NOT_STARTED",
			},
			assertField: "state",
			assertValue: "NOT_STARTED",
		},
		{
			name: "with IN_PROGRESS state",
			input: map[string]any{
				"name":  "Audit IN_PROGRESS",
				"state": "IN_PROGRESS",
			},
			assertField: "state",
			assertValue: "IN_PROGRESS",
		},
		{
			name: "with COMPLETED state",
			input: map[string]any{
				"name":  "Audit COMPLETED",
				"state": "COMPLETED",
			},
			assertField: "state",
			assertValue: "COMPLETED",
		},
		{
			name: "with REJECTED state",
			input: map[string]any{
				"name":  "Audit REJECTED",
				"state": "REJECTED",
			},
			assertField: "state",
			assertValue: "REJECTED",
		},
		{
			name: "with OUTDATED state",
			input: map[string]any{
				"name":  "Audit OUTDATED",
				"state": "OUTDATED",
			},
			assertField: "state",
			assertValue: "OUTDATED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation CreateAudit($input: CreateAuditInput!) {
					createAudit(input: $input) {
						auditEdge {
							node {
								id
								name
								state
							}
						}
					}
				}
			`

			input := map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"frameworkId":    frameworkID,
			}
			for k, v := range tt.input {
				input[k] = v
			}

			var result struct {
				CreateAudit struct {
					AuditEdge struct {
						Node struct {
							ID    string `json:"id"`
							Name  string `json:"name"`
							State string `json:"state"`
						} `json:"node"`
					} `json:"auditEdge"`
				} `json:"createAudit"`
			}

			err := owner.Execute(query, map[string]any{"input": input}, &result)
			require.NoError(t, err)

			node := result.CreateAudit.AuditEdge.Node
			assert.NotEmpty(t, node.ID)

			switch tt.assertField {
			case "name":
				assert.Equal(t, tt.assertValue, node.Name)
			case "state":
				assert.Equal(t, tt.assertValue, node.State)
			}
		})
	}
}

func TestAudit_Create_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Audit Validation").Create()

	tests := []struct {
		name              string
		input             map[string]any
		skipOrganization  bool
		skipFramework     bool
		wantErrorContains string
	}{
		{
			name: "missing organizationId",
			input: map[string]any{
				"name": "Test Audit",
			},
			skipOrganization:  true,
			wantErrorContains: "organizationId",
		},
		{
			name: "missing frameworkId",
			input: map[string]any{
				"name": "Test Audit",
			},
			skipFramework:     true,
			wantErrorContains: "frameworkId",
		},
		{
			name: "name with HTML tags",
			input: map[string]any{
				"name": "<script>alert('xss')</script>",
			},
			wantErrorContains: "HTML",
		},
		{
			name: "name with angle brackets",
			input: map[string]any{
				"name": "Test < Audit",
			},
			wantErrorContains: "angle brackets",
		},
		{
			name: "name with newline",
			input: map[string]any{
				"name": "Test\nAudit",
			},
			wantErrorContains: "newline",
		},
		{
			name: "name with carriage return",
			input: map[string]any{
				"name": "Test\rAudit",
			},
			wantErrorContains: "carriage return",
		},
		{
			name: "name with null byte",
			input: map[string]any{
				"name": "Test\x00Audit",
			},
			wantErrorContains: "control character",
		},
		{
			name: "name with tab character",
			input: map[string]any{
				"name": "Test\tAudit",
			},
			wantErrorContains: "control character",
		},
		{
			name: "name with zero-width space",
			input: map[string]any{
				"name": "Test\u200BAudit",
			},
			wantErrorContains: "zero-width",
		},
		{
			name: "name with zero-width joiner",
			input: map[string]any{
				"name": "Test\u200DAudit",
			},
			wantErrorContains: "zero-width",
		},
		{
			name: "name with right-to-left override",
			input: map[string]any{
				"name": "Test\u202EAudit",
			},
			wantErrorContains: "bidirectional",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
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

			input := make(map[string]any)
			if !tt.skipOrganization {
				input["organizationId"] = owner.GetOrganizationID().String()
			}
			if !tt.skipFramework {
				input["frameworkId"] = frameworkID
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

func TestAudit_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Audit Update").Create()

	tests := []struct {
		name        string
		setup       func() string
		input       func(id string) map[string]any
		assertField string
		assertValue string
	}{
		{
			name: "update name",
			setup: func() string {
				return factory.NewAudit(owner, frameworkID).
					WithName("Audit to Update").
					Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{
					"id":   id,
					"name": "Updated Audit Name",
				}
			},
			assertField: "name",
			assertValue: "Updated Audit Name",
		},
		{
			name: "update to IN_PROGRESS state",
			setup: func() string {
				return factory.NewAudit(owner, frameworkID).
					WithName("State Test").
					WithState("NOT_STARTED").
					Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "state": "IN_PROGRESS"}
			},
			assertField: "state",
			assertValue: "IN_PROGRESS",
		},
		{
			name: "update to COMPLETED state",
			setup: func() string {
				return factory.NewAudit(owner, frameworkID).
					WithName("State Test").
					Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "state": "COMPLETED"}
			},
			assertField: "state",
			assertValue: "COMPLETED",
		},
		{
			name: "update to REJECTED state",
			setup: func() string {
				return factory.NewAudit(owner, frameworkID).
					WithName("State Test").
					Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "state": "REJECTED"}
			},
			assertField: "state",
			assertValue: "REJECTED",
		},
		{
			name: "update to OUTDATED state",
			setup: func() string {
				return factory.NewAudit(owner, frameworkID).
					WithName("State Test").
					Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "state": "OUTDATED"}
			},
			assertField: "state",
			assertValue: "OUTDATED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auditID := tt.setup()

			query := `
				mutation UpdateAudit($input: UpdateAuditInput!) {
					updateAudit(input: $input) {
						audit {
							id
							name
							state
						}
					}
				}
			`

			var result struct {
				UpdateAudit struct {
					Audit struct {
						ID    string `json:"id"`
						Name  string `json:"name"`
						State string `json:"state"`
					} `json:"audit"`
				} `json:"updateAudit"`
			}

			err := owner.Execute(query, map[string]any{"input": tt.input(auditID)}, &result)
			require.NoError(t, err)

			audit := result.UpdateAudit.Audit
			switch tt.assertField {
			case "name":
				assert.Equal(t, tt.assertValue, audit.Name)
			case "state":
				assert.Equal(t, tt.assertValue, audit.State)
			}
		})
	}
}

func TestAudit_Update_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Audit Update Validation").Create()
	baseAuditID := factory.NewAudit(owner, frameworkID).WithName("Validation Test Audit").Create()

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
		{
			name:  "name with HTML tags",
			setup: func() string { return baseAuditID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "<script>alert('xss')</script>"}
			},
			wantErrorContains: "HTML",
		},
		{
			name:  "name with angle brackets",
			setup: func() string { return baseAuditID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test < Audit"}
			},
			wantErrorContains: "angle brackets",
		},
		{
			name:  "name with newline",
			setup: func() string { return baseAuditID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\nAudit"}
			},
			wantErrorContains: "newline",
		},
		{
			name:  "name with carriage return",
			setup: func() string { return baseAuditID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\rAudit"}
			},
			wantErrorContains: "carriage return",
		},
		{
			name:  "name with null byte",
			setup: func() string { return baseAuditID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\x00Audit"}
			},
			wantErrorContains: "control character",
		},
		{
			name:  "name with tab character",
			setup: func() string { return baseAuditID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\tAudit"}
			},
			wantErrorContains: "control character",
		},
		{
			name:  "name with zero-width space",
			setup: func() string { return baseAuditID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\u200BAudit"}
			},
			wantErrorContains: "zero-width",
		},
		{
			name:  "name with zero-width joiner",
			setup: func() string { return baseAuditID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\u200DAudit"}
			},
			wantErrorContains: "zero-width",
		},
		{
			name:  "name with right-to-left override",
			setup: func() string { return baseAuditID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\u202EAudit"}
			},
			wantErrorContains: "bidirectional",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auditID := tt.setup()

			query := `
				mutation UpdateAudit($input: UpdateAuditInput!) {
					updateAudit(input: $input) {
						audit {
							id
						}
					}
				}
			`

			_, err := owner.Do(query, map[string]any{"input": tt.input(auditID)})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestAudit_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Audit Delete").Create()

	t.Run("delete existing audit", func(t *testing.T) {
		auditID := factory.NewAudit(owner, frameworkID).WithName("Audit to Delete").Create()

		query := `
			mutation DeleteAudit($input: DeleteAuditInput!) {
				deleteAudit(input: $input) {
					deletedAuditId
				}
			}
		`

		var result struct {
			DeleteAudit struct {
				DeletedAuditID string `json:"deletedAuditId"`
			} `json:"deleteAudit"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{"auditId": auditID},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, auditID, result.DeleteAudit.DeletedAuditID)
	})
}

func TestAudit_Delete_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	tests := []struct {
		name              string
		auditID           string
		wantErrorContains string
	}{
		{
			name:              "invalid ID format",
			auditID:           "invalid-id-format",
			wantErrorContains: "base64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation DeleteAudit($input: DeleteAuditInput!) {
					deleteAudit(input: $input) {
						deletedAuditId
					}
				}
			`

			_, err := owner.Do(query, map[string]any{
				"input": map[string]any{"auditId": tt.auditID},
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestAudit_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Audit List").Create()

	auditNames := []string{"Audit A", "Audit B", "Audit C"}
	for _, name := range auditNames {
		factory.NewAudit(owner, frameworkID).WithName(name).Create()
	}

	query := `
		query GetAudits($id: ID!) {
			node(id: $id) {
				... on Organization {
					audits(first: 10) {
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
			Audits struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"audits"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.Audits.TotalCount, 3)
}

func TestAudit_Query(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("query with non-existent ID returns error", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Audit {
						id
						name
					}
				}
			}
		`

		err := owner.ExecuteShouldFail(query, map[string]any{
			"id": "V0wtM0tMNmJBQ1lBQUFBQUFackhLSTJfbXJJRUFZVXo", // Valid format but doesn't exist
		})
		require.Error(t, err, "Non-existent ID should return error")
	})
}

func TestAudit_Timestamps(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Audit Timestamps").Create()

	t.Run("createdAt and updatedAt are set on create", func(t *testing.T) {
		beforeCreate := time.Now().Add(-time.Second)

		query := `
			mutation CreateAudit($input: CreateAuditInput!) {
				createAudit(input: $input) {
					auditEdge {
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
			CreateAudit struct {
				AuditEdge struct {
					Node struct {
						ID        string    `json:"id"`
						CreatedAt time.Time `json:"createdAt"`
						UpdatedAt time.Time `json:"updatedAt"`
					} `json:"node"`
				} `json:"auditEdge"`
			} `json:"createAudit"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"frameworkId":    frameworkID,
				"name":           "Timestamp Test Audit",
			},
		}, &result)
		require.NoError(t, err)

		node := result.CreateAudit.AuditEdge.Node
		testutil.AssertTimestampsOnCreate(t, node.CreatedAt, node.UpdatedAt, beforeCreate)
	})

	t.Run("updatedAt changes on update", func(t *testing.T) {
		auditID := factory.NewAudit(owner, frameworkID).WithName("Timestamp Update Test").Create()

		getQuery := `
			query($id: ID!) {
				node(id: $id) {
					... on Audit {
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

		err := owner.Execute(getQuery, map[string]any{"id": auditID}, &getResult)
		require.NoError(t, err)

		initialCreatedAt := getResult.Node.CreatedAt
		initialUpdatedAt := getResult.Node.UpdatedAt

		time.Sleep(10 * time.Millisecond)

		updateQuery := `
			mutation UpdateAudit($input: UpdateAuditInput!) {
				updateAudit(input: $input) {
					audit {
						createdAt
						updatedAt
					}
				}
			}
		`

		var updateResult struct {
			UpdateAudit struct {
				Audit struct {
					CreatedAt time.Time `json:"createdAt"`
					UpdatedAt time.Time `json:"updatedAt"`
				} `json:"audit"`
			} `json:"updateAudit"`
		}

		err = owner.Execute(updateQuery, map[string]any{
			"input": map[string]any{
				"id":   auditID,
				"name": "Updated Timestamp Test",
			},
		}, &updateResult)
		require.NoError(t, err)

		audit := updateResult.UpdateAudit.Audit
		testutil.AssertTimestampsOnUpdate(t, audit.CreatedAt, audit.UpdatedAt, initialCreatedAt, initialUpdatedAt)
	})
}

func TestAudit_SubResolvers(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Audit SubResolvers").Create()
	auditID := factory.NewAudit(owner, frameworkID).WithName("SubResolver Test Audit").Create()

	t.Run("framework sub-resolver", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Audit {
						id
						framework {
							id
							name
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID        string `json:"id"`
				Framework struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"framework"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": auditID}, &result)
		require.NoError(t, err)
		assert.Equal(t, frameworkID, result.Node.Framework.ID)
	})

	t.Run("organization sub-resolver", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Audit {
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

		err := owner.Execute(query, map[string]any{"id": auditID}, &result)
		require.NoError(t, err)
		assert.Equal(t, owner.GetOrganizationID().String(), result.Node.Organization.ID)
		assert.NotEmpty(t, result.Node.Organization.Name)
	})
}

func TestAudit_RBAC(t *testing.T) {
	t.Parallel()

	t.Run("create", func(t *testing.T) {
		t.Run("owner can create", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()

			_, err := owner.Do(`
				mutation CreateAudit($input: CreateAuditInput!) {
					createAudit(input: $input) {
						auditEdge { node { id } }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"organizationId": owner.GetOrganizationID().String(),
					"frameworkId":    frameworkID,
					"name":           "RBAC Test Audit",
				},
			})
			require.NoError(t, err, "owner should be able to create audit")
		})

		t.Run("admin can create", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()

			_, err := admin.Do(`
				mutation CreateAudit($input: CreateAuditInput!) {
					createAudit(input: $input) {
						auditEdge { node { id } }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"organizationId": admin.GetOrganizationID().String(),
					"frameworkId":    frameworkID,
					"name":           "RBAC Test Audit",
				},
			})
			require.NoError(t, err, "admin should be able to create audit")
		})

		t.Run("viewer cannot create", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()

			_, err := viewer.Do(`
				mutation CreateAudit($input: CreateAuditInput!) {
					createAudit(input: $input) {
						auditEdge { node { id } }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"organizationId": viewer.GetOrganizationID().String(),
					"frameworkId":    frameworkID,
					"name":           "RBAC Test Audit",
				},
			})
			testutil.RequireForbiddenError(t, err, "viewer should not be able to create audit")
		})
	})

	t.Run("update", func(t *testing.T) {
		t.Run("owner can update", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
			auditID := factory.NewAudit(owner, frameworkID).WithName("RBAC Update Test").Create()

			_, err := owner.Do(`
				mutation UpdateAudit($input: UpdateAuditInput!) {
					updateAudit(input: $input) {
						audit { id }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"id":   auditID,
					"name": "Updated by Owner",
				},
			})
			require.NoError(t, err, "owner should be able to update audit")
		})

		t.Run("admin can update", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
			auditID := factory.NewAudit(owner, frameworkID).WithName("RBAC Update Test").Create()

			_, err := admin.Do(`
				mutation UpdateAudit($input: UpdateAuditInput!) {
					updateAudit(input: $input) {
						audit { id }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"id":   auditID,
					"name": "Updated by Admin",
				},
			})
			require.NoError(t, err, "admin should be able to update audit")
		})

		t.Run("viewer cannot update", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
			auditID := factory.NewAudit(owner, frameworkID).WithName("RBAC Update Test").Create()

			_, err := viewer.Do(`
				mutation UpdateAudit($input: UpdateAuditInput!) {
					updateAudit(input: $input) {
						audit { id }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"id":   auditID,
					"name": "Updated by Viewer",
				},
			})
			testutil.RequireForbiddenError(t, err, "viewer should not be able to update audit")
		})
	})

	t.Run("delete", func(t *testing.T) {
		t.Run("owner can delete", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
			auditID := factory.NewAudit(owner, frameworkID).WithName("RBAC Delete Test").Create()

			_, err := owner.Do(`
				mutation DeleteAudit($input: DeleteAuditInput!) {
					deleteAudit(input: $input) {
						deletedAuditId
					}
				}
			`, map[string]any{
				"input": map[string]any{"auditId": auditID},
			})
			require.NoError(t, err, "owner should be able to delete audit")
		})

		t.Run("admin can delete", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
			auditID := factory.NewAudit(owner, frameworkID).WithName("RBAC Delete Test").Create()

			_, err := admin.Do(`
				mutation DeleteAudit($input: DeleteAuditInput!) {
					deleteAudit(input: $input) {
						deletedAuditId
					}
				}
			`, map[string]any{
				"input": map[string]any{"auditId": auditID},
			})
			require.NoError(t, err, "admin should be able to delete audit")
		})

		t.Run("viewer cannot delete", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
			auditID := factory.NewAudit(owner, frameworkID).WithName("RBAC Delete Test").Create()

			_, err := viewer.Do(`
				mutation DeleteAudit($input: DeleteAuditInput!) {
					deleteAudit(input: $input) {
						deletedAuditId
					}
				}
			`, map[string]any{
				"input": map[string]any{"auditId": auditID},
			})
			testutil.RequireForbiddenError(t, err, "viewer should not be able to delete audit")
		})
	})

	t.Run("read", func(t *testing.T) {
		t.Run("owner can read", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
			auditID := factory.NewAudit(owner, frameworkID).WithName("RBAC Read Test").Create()

			var result struct {
				Node *struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			}

			err := owner.Execute(`
				query($id: ID!) {
					node(id: $id) {
						... on Audit { id name }
					}
				}
			`, map[string]any{"id": auditID}, &result)
			require.NoError(t, err, "owner should be able to read audit")
			require.NotNil(t, result.Node, "owner should receive audit data")
		})

		t.Run("admin can read", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
			auditID := factory.NewAudit(owner, frameworkID).WithName("RBAC Read Test").Create()

			var result struct {
				Node *struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			}

			err := admin.Execute(`
				query($id: ID!) {
					node(id: $id) {
						... on Audit { id name }
					}
				}
			`, map[string]any{"id": auditID}, &result)
			require.NoError(t, err, "admin should be able to read audit")
			require.NotNil(t, result.Node, "admin should receive audit data")
		})

		t.Run("viewer can read", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
			auditID := factory.NewAudit(owner, frameworkID).WithName("RBAC Read Test").Create()

			var result struct {
				Node *struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			}

			err := viewer.Execute(`
				query($id: ID!) {
					node(id: $id) {
						... on Audit { id name }
					}
				}
			`, map[string]any{"id": auditID}, &result)
			require.NoError(t, err, "viewer should be able to read audit")
			require.NotNil(t, result.Node, "viewer should receive audit data")
		})
	})
}

func TestAudit_MaxLength_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Max Length").Create()
	longName := strings.Repeat("a", 1001)

	t.Run("create", func(t *testing.T) {
		query := `
			mutation CreateAudit($input: CreateAuditInput!) {
				createAudit(input: $input) {
					auditEdge {
						node { id }
					}
				}
			}
		`

		_, err := owner.Do(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"frameworkId":    frameworkID,
				"name":           longName,
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("update", func(t *testing.T) {
		auditID := factory.NewAudit(owner, frameworkID).WithName("Max Length Test").Create()

		query := `
			mutation UpdateAudit($input: UpdateAuditInput!) {
				updateAudit(input: $input) {
					audit { id }
				}
			}
		`

		_, err := owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":   auditID,
				"name": longName,
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})
}

func TestAudit_Pagination(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Pagination").Create()

	for i := 0; i < 5; i++ {
		factory.NewAudit(owner, frameworkID).
			WithName(fmt.Sprintf("Pagination Audit %d", i)).
			Create()
	}

	t.Run("first/after pagination", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						audits(first: 2) {
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
				Audits struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
						Cursor string `json:"cursor"`
					} `json:"edges"`
					PageInfo   testutil.PageInfo `json:"pageInfo"`
					TotalCount int               `json:"totalCount"`
				} `json:"audits"`
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

		testutil.AssertFirstPage(t, len(result.Node.Audits.Edges), result.Node.Audits.PageInfo, 2, true)
		assert.GreaterOrEqual(t, result.Node.Audits.TotalCount, 5)

		testutil.AssertHasMorePages(t, result.Node.Audits.PageInfo)
		queryAfter := `
			query($id: ID!, $after: CursorKey) {
				node(id: $id) {
					... on Organization {
						audits(first: 2, after: $after) {
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
				Audits struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo testutil.PageInfo `json:"pageInfo"`
				} `json:"audits"`
			} `json:"node"`
		}

		err = owner.Execute(queryAfter, map[string]any{
			"id":    owner.GetOrganizationID().String(),
			"after": *result.Node.Audits.PageInfo.EndCursor,
		}, &resultAfter)
		require.NoError(t, err)

		testutil.AssertMiddlePage(t, len(resultAfter.Node.Audits.Edges), resultAfter.Node.Audits.PageInfo, 2)
	})

	t.Run("last/before pagination", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						audits(last: 2) {
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
				Audits struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo testutil.PageInfo `json:"pageInfo"`
				} `json:"audits"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id": owner.GetOrganizationID().String(),
		}, &result)
		require.NoError(t, err)

		testutil.AssertLastPage(t, len(result.Node.Audits.Edges), result.Node.Audits.PageInfo, 2, true)
	})
}

func TestAudit_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(org1Owner).WithName("Org1 Framework").Create()
	auditID := factory.NewAudit(org1Owner, frameworkID).WithName("Org1 Audit").Create()

	t.Run("cannot read audit from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Audit {
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

		err := org2Owner.Execute(query, map[string]any{"id": auditID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "audit")
	})

	t.Run("cannot update audit from another organization", func(t *testing.T) {
		query := `
			mutation UpdateAudit($input: UpdateAuditInput!) {
				updateAudit(input: $input) {
					audit { id }
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":   auditID,
				"name": "Hijacked Audit",
			},
		})
		require.Error(t, err, "Should not be able to update audit from another org")
	})

	t.Run("cannot delete audit from another organization", func(t *testing.T) {
		query := `
			mutation DeleteAudit($input: DeleteAuditInput!) {
				deleteAudit(input: $input) {
					deletedAuditId
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"auditId": auditID,
			},
		})
		require.Error(t, err, "Should not be able to delete audit from another org")
	})

	t.Run("cannot list audits from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						audits(first: 100) {
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
				Audits struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"audits"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{
			"id": org1Owner.GetOrganizationID().String(),
		}, &result)

		if err == nil {
			for _, edge := range result.Node.Audits.Edges {
				assert.NotEqual(t, auditID, edge.Node.ID, "Should not see audit from another org")
			}
		}
	})
}

func TestAudit_Ordering(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Ordering").Create()

	factory.NewAudit(owner, frameworkID).WithName("AAA Order Test").Create()
	factory.NewAudit(owner, frameworkID).WithName("ZZZ Order Test").Create()

	t.Run("order by created_at descending", func(t *testing.T) {
		query := `
			query($id: ID!, $orderBy: AuditOrder) {
				node(id: $id) {
					... on Organization {
						audits(first: 100, orderBy: $orderBy) {
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
				Audits struct {
					Edges []struct {
						Node struct {
							ID        string    `json:"id"`
							CreatedAt time.Time `json:"createdAt"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"audits"`
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

		times := make([]time.Time, len(result.Node.Audits.Edges))
		for i, edge := range result.Node.Audits.Edges {
			times[i] = edge.Node.CreatedAt
		}
		testutil.AssertTimesOrderedDescending(t, times, "createdAt")
	})
}

func TestAudit_UploadReport(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Upload").Create()

	t.Run("upload valid PDF report", func(t *testing.T) {
		auditID := factory.NewAudit(owner, frameworkID).WithName("Upload Test").Create()

		query := `
			mutation UploadAuditReport($input: UploadAuditReportInput!) {
				uploadAuditReport(input: $input) {
					audit {
						id
						report {
							id
							filename
							size
						}
					}
				}
			}
		`

		// Create a minimal valid PDF content
		pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

		var result struct {
			UploadAuditReport struct {
				Audit struct {
					ID     string `json:"id"`
					Report *struct {
						ID       string `json:"id"`
						Filename string `json:"filename"`
						Size     int    `json:"size"`
					} `json:"report"`
				} `json:"audit"`
			} `json:"uploadAuditReport"`
		}

		err := owner.ExecuteWithFile(query, map[string]any{
			"input": map[string]any{
				"auditId": auditID,
				"file":    nil, // Will be replaced by the file
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "audit-report.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent,
		}, &result)
		require.NoError(t, err)

		assert.Equal(t, auditID, result.UploadAuditReport.Audit.ID)
		require.NotNil(t, result.UploadAuditReport.Audit.Report)
		assert.Equal(t, "audit-report.pdf", result.UploadAuditReport.Audit.Report.Filename)
		assert.Equal(t, len(pdfContent), result.UploadAuditReport.Audit.Report.Size)
	})

	t.Run("upload replaces existing report", func(t *testing.T) {
		auditID := factory.NewAudit(owner, frameworkID).WithName("Replace Report Test").Create()

		query := `
			mutation UploadAuditReport($input: UploadAuditReportInput!) {
				uploadAuditReport(input: $input) {
					audit {
						id
						report {
							id
							filename
						}
					}
				}
			}
		`

		// Upload first report
		pdfContent1 := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

		var result1 struct {
			UploadAuditReport struct {
				Audit struct {
					ID     string `json:"id"`
					Report *struct {
						ID       string `json:"id"`
						Filename string `json:"filename"`
					} `json:"report"`
				} `json:"audit"`
			} `json:"uploadAuditReport"`
		}

		err := owner.ExecuteWithFile(query, map[string]any{
			"input": map[string]any{
				"auditId": auditID,
				"file":    nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "first-report.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent1,
		}, &result1)
		require.NoError(t, err)
		firstReportID := result1.UploadAuditReport.Audit.Report.ID

		// Upload second report (should replace)
		pdfContent2 := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Version /1.4 >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

		var result2 struct {
			UploadAuditReport struct {
				Audit struct {
					ID     string `json:"id"`
					Report *struct {
						ID       string `json:"id"`
						Filename string `json:"filename"`
					} `json:"report"`
				} `json:"audit"`
			} `json:"uploadAuditReport"`
		}

		err = owner.ExecuteWithFile(query, map[string]any{
			"input": map[string]any{
				"auditId": auditID,
				"file":    nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "second-report.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent2,
		}, &result2)
		require.NoError(t, err)

		assert.Equal(t, "second-report.pdf", result2.UploadAuditReport.Audit.Report.Filename)
		assert.NotEqual(t, firstReportID, result2.UploadAuditReport.Audit.Report.ID, "Report ID should change when replaced")
	})
}

func TestAudit_UploadReport_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Upload Validation").Create()

	t.Run("reject non document file", func(t *testing.T) {
		auditID := factory.NewAudit(owner, frameworkID).WithName("Invalid File Test").Create()

		query := `
			mutation UploadAuditReport($input: UploadAuditReportInput!) {
				uploadAuditReport(input: $input) {
					audit {
						id
					}
				}
			}
		`

		// Try to upload a text file
		textContent := []byte("This is not a document file")

		err := owner.ExecuteWithFile(query, map[string]any{
			"input": map[string]any{
				"auditId": auditID,
				"file":    nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "not-a-document.txt",
			ContentType: "text/plain",
			Content:     textContent,
		}, nil)
		require.Error(t, err, "Should reject non-document file")
	})

	t.Run("reject file with wrong extension but document content-type", func(t *testing.T) {
		auditID := factory.NewAudit(owner, frameworkID).WithName("Wrong Extension Test").Create()

		query := `
			mutation UploadAuditReport($input: UploadAuditReportInput!) {
				uploadAuditReport(input: $input) {
					audit {
						id
					}
				}
			}
		`

		// Try to upload with wrong extension
		textContent := []byte("Not a real document")

		err := owner.ExecuteWithFile(query, map[string]any{
			"input": map[string]any{
				"auditId": auditID,
				"file":    nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "fake.exe",
			ContentType: "application/pdf",
			Content:     textContent,
		}, nil)
		require.Error(t, err, "Should reject file with wrong extension")
	})

	t.Run("reject empty file", func(t *testing.T) {
		auditID := factory.NewAudit(owner, frameworkID).WithName("Empty File Test").Create()

		query := `
			mutation UploadAuditReport($input: UploadAuditReportInput!) {
				uploadAuditReport(input: $input) {
					audit {
						id
					}
				}
			}
		`

		err := owner.ExecuteWithFile(query, map[string]any{
			"input": map[string]any{
				"auditId": auditID,
				"file":    nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "empty.pdf",
			ContentType: "application/pdf",
			Content:     []byte{},
		}, nil)
		require.Error(t, err, "Should reject empty file")
	})

	t.Run("reject invalid audit ID", func(t *testing.T) {
		query := `
			mutation UploadAuditReport($input: UploadAuditReportInput!) {
				uploadAuditReport(input: $input) {
					audit {
						id
					}
				}
			}
		`

		pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

		err := owner.ExecuteWithFile(query, map[string]any{
			"input": map[string]any{
				"auditId": "invalid-id",
				"file":    nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "report.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent,
		}, nil)
		require.Error(t, err, "Should reject invalid audit ID")
	})
}

func TestAudit_UploadReport_RBAC(t *testing.T) {
	t.Parallel()

	query := `
		mutation UploadAuditReport($input: UploadAuditReportInput!) {
			uploadAuditReport(input: $input) {
				audit {
					id
				}
			}
		}
	`

	pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

	t.Run("owner can upload", func(t *testing.T) {
		owner := testutil.NewClient(t, testutil.RoleOwner)
		frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
		auditID := factory.NewAudit(owner, frameworkID).WithName("RBAC Upload Test").Create()

		err := owner.ExecuteWithFile(query, map[string]any{
			"input": map[string]any{
				"auditId": auditID,
				"file":    nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "report.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent,
		}, nil)
		require.NoError(t, err, "owner should be able to upload report")
	})

	t.Run("admin can upload", func(t *testing.T) {
		owner := testutil.NewClient(t, testutil.RoleOwner)
		admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
		frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
		auditID := factory.NewAudit(owner, frameworkID).WithName("RBAC Upload Test").Create()

		err := admin.ExecuteWithFile(query, map[string]any{
			"input": map[string]any{
				"auditId": auditID,
				"file":    nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "report.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent,
		}, nil)
		require.NoError(t, err, "admin should be able to upload report")
	})

	t.Run("viewer cannot upload", func(t *testing.T) {
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
		auditID := factory.NewAudit(owner, frameworkID).WithName("RBAC Upload Test").Create()

		err := viewer.ExecuteWithFile(query, map[string]any{
			"input": map[string]any{
				"auditId": auditID,
				"file":    nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "report.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent,
		}, nil)
		testutil.RequireForbiddenError(t, err, "viewer should not be able to upload report")
	})
}

func TestAudit_DeleteReport(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	frameworkID := factory.NewFramework(owner).WithName("Framework for Delete Report").Create()

	t.Run("delete existing report", func(t *testing.T) {
		auditID := factory.NewAudit(owner, frameworkID).WithName("Delete Report Test").Create()

		// First upload a report
		uploadQuery := `
			mutation UploadAuditReport($input: UploadAuditReportInput!) {
				uploadAuditReport(input: $input) {
					audit {
						id
						report {
							id
						}
					}
				}
			}
		`

		pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

		err := owner.ExecuteWithFile(uploadQuery, map[string]any{
			"input": map[string]any{
				"auditId": auditID,
				"file":    nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "report.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent,
		}, nil)
		require.NoError(t, err)

		// Now delete the report
		deleteQuery := `
			mutation DeleteAuditReport($input: DeleteAuditReportInput!) {
				deleteAuditReport(input: $input) {
					audit {
						id
						report {
							id
						}
					}
				}
			}
		`

		var deleteResult struct {
			DeleteAuditReport struct {
				Audit struct {
					ID     string `json:"id"`
					Report *struct {
						ID string `json:"id"`
					} `json:"report"`
				} `json:"audit"`
			} `json:"deleteAuditReport"`
		}

		err = owner.Execute(deleteQuery, map[string]any{
			"input": map[string]any{
				"auditId": auditID,
			},
		}, &deleteResult)
		require.NoError(t, err)
		assert.Equal(t, auditID, deleteResult.DeleteAuditReport.Audit.ID)
		assert.Nil(t, deleteResult.DeleteAuditReport.Audit.Report, "Report should be nil after deletion")
	})
}

func TestAudit_DeleteReport_RBAC(t *testing.T) {
	t.Parallel()

	uploadQuery := `
		mutation UploadAuditReport($input: UploadAuditReportInput!) {
			uploadAuditReport(input: $input) {
				audit { id }
			}
		}
	`

	deleteQuery := `
		mutation DeleteAuditReport($input: DeleteAuditReportInput!) {
			deleteAuditReport(input: $input) {
				audit { id }
			}
		}
	`

	pdfContent := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF")

	t.Run("viewer cannot delete report", func(t *testing.T) {
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		frameworkID := factory.NewFramework(owner).WithName("RBAC Framework").Create()
		auditID := factory.NewAudit(owner, frameworkID).WithName("RBAC Delete Report Test").Create()

		// Owner uploads the report
		err := owner.ExecuteWithFile(uploadQuery, map[string]any{
			"input": map[string]any{
				"auditId": auditID,
				"file":    nil,
			},
		}, "input.file", testutil.UploadFile{
			Filename:    "report.pdf",
			ContentType: "application/pdf",
			Content:     pdfContent,
		}, nil)
		require.NoError(t, err)

		// Viewer tries to delete
		_, err = viewer.Do(deleteQuery, map[string]any{
			"input": map[string]any{
				"auditId": auditID,
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to delete report")
	})
}
