// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"fmt"
	"maps"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestDatum_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)

	tests := []struct {
		name        string
		input       map[string]any
		assertField string
		assertValue string
	}{
		{
			name: "with full details",
			input: map[string]any{
				"name":               "Customer PII",
				"dataClassification": "CONFIDENTIAL",
			},
			assertField: "name",
			assertValue: "Customer PII",
		},
		{
			name: "with PUBLIC classification",
			input: map[string]any{
				"name":               "Public Data",
				"dataClassification": "PUBLIC",
			},
			assertField: "dataClassification",
			assertValue: "PUBLIC",
		},
		{
			name: "with INTERNAL classification",
			input: map[string]any{
				"name":               "Internal Data",
				"dataClassification": "INTERNAL",
			},
			assertField: "dataClassification",
			assertValue: "INTERNAL",
		},
		{
			name: "with CONFIDENTIAL classification",
			input: map[string]any{
				"name":               "Confidential Data",
				"dataClassification": "CONFIDENTIAL",
			},
			assertField: "dataClassification",
			assertValue: "CONFIDENTIAL",
		},
		{
			name: "with SECRET classification",
			input: map[string]any{
				"name":               "Secret Data",
				"dataClassification": "SECRET",
			},
			assertField: "dataClassification",
			assertValue: "SECRET",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation CreateDatum($input: CreateDatumInput!) {
					createDatum(input: $input) {
						datumEdge {
							node {
								id
								name
								dataClassification
							}
						}
					}
				}
			`

			input := map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"ownerId":        profileID,
			}
			maps.Copy(input, tt.input)

			var result struct {
				CreateDatum struct {
					DatumEdge struct {
						Node struct {
							ID                 string `json:"id"`
							Name               string `json:"name"`
							DataClassification string `json:"dataClassification"`
						} `json:"node"`
					} `json:"datumEdge"`
				} `json:"createDatum"`
			}

			err := owner.Execute(query, map[string]any{"input": input}, &result)
			require.NoError(t, err)

			node := result.CreateDatum.DatumEdge.Node
			assert.NotEmpty(t, node.ID)

			switch tt.assertField {
			case "name":
				assert.Equal(t, tt.assertValue, node.Name)
			case "dataClassification":
				assert.Equal(t, tt.assertValue, node.DataClassification)
			}
		})
	}
}

func TestDatum_Create_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)

	tests := []struct {
		name              string
		input             map[string]any
		skipOrganization  bool
		skipOwner         bool
		wantErrorContains string
	}{
		{
			name: "missing organizationId",
			input: map[string]any{
				"name":               "Test Datum",
				"dataClassification": "INTERNAL",
			},
			skipOrganization:  true,
			wantErrorContains: "organizationId",
		},
		{
			name: "missing ownerId",
			input: map[string]any{
				"name":               "Test Datum",
				"dataClassification": "INTERNAL",
			},
			skipOwner:         true,
			wantErrorContains: "ownerId",
		},
		{
			name: "name with HTML tags",
			input: map[string]any{
				"name":               "<script>alert('xss')</script>",
				"dataClassification": "INTERNAL",
			},
			wantErrorContains: "HTML",
		},
		{
			name: "name with newline",
			input: map[string]any{
				"name":               "Test\nDatum",
				"dataClassification": "INTERNAL",
			},
			wantErrorContains: "newline",
		},
		{
			name: "name with carriage return",
			input: map[string]any{
				"name":               "Test\rDatum",
				"dataClassification": "INTERNAL",
			},
			wantErrorContains: "carriage return",
		},
		{
			name: "name with null byte",
			input: map[string]any{
				"name":               "Test\x00Datum",
				"dataClassification": "INTERNAL",
			},
			wantErrorContains: "control character",
		},
		{
			name: "name with tab character",
			input: map[string]any{
				"name":               "Test\tDatum",
				"dataClassification": "INTERNAL",
			},
			wantErrorContains: "control character",
		},
		{
			name: "name with zero-width space",
			input: map[string]any{
				"name":               "Test\u200BDatum",
				"dataClassification": "INTERNAL",
			},
			wantErrorContains: "zero-width",
		},
		{
			name: "name with zero-width joiner",
			input: map[string]any{
				"name":               "Test\u200DDatum",
				"dataClassification": "INTERNAL",
			},
			wantErrorContains: "zero-width",
		},
		{
			name: "name with right-to-left override",
			input: map[string]any{
				"name":               "Test\u202EDatum",
				"dataClassification": "INTERNAL",
			},
			wantErrorContains: "bidirectional",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation CreateDatum($input: CreateDatumInput!) {
					createDatum(input: $input) {
						datumEdge {
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

			if !tt.skipOwner {
				input["ownerId"] = profileID
			}

			maps.Copy(input, tt.input)

			_, err := owner.Do(query, map[string]any{"input": input})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestDatum_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)

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
				return factory.NewDatum(owner, profileID).
					WithName("Datum to Update").
					Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{
					"id":   id,
					"name": "Updated Datum Name",
				}
			},
			assertField: "name",
			assertValue: "Updated Datum Name",
		},
		{
			name: "update to PUBLIC classification",
			setup: func() string {
				return factory.NewDatum(owner, profileID).
					WithName("Classification Test").
					WithDataClassification("INTERNAL").
					Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "dataClassification": "PUBLIC"}
			},
			assertField: "dataClassification",
			assertValue: "PUBLIC",
		},
		{
			name: "update to SECRET classification",
			setup: func() string {
				return factory.NewDatum(owner, profileID).
					WithName("Classification Test").
					Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "dataClassification": "SECRET"}
			},
			assertField: "dataClassification",
			assertValue: "SECRET",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			datumID := tt.setup()

			query := `
				mutation UpdateDatum($input: UpdateDatumInput!) {
					updateDatum(input: $input) {
						datum {
							id
							name
							dataClassification
						}
					}
				}
			`

			var result struct {
				UpdateDatum struct {
					Datum struct {
						ID                 string `json:"id"`
						Name               string `json:"name"`
						DataClassification string `json:"dataClassification"`
					} `json:"datum"`
				} `json:"updateDatum"`
			}

			err := owner.Execute(query, map[string]any{"input": tt.input(datumID)}, &result)
			require.NoError(t, err)

			datum := result.UpdateDatum.Datum

			switch tt.assertField {
			case "name":
				assert.Equal(t, tt.assertValue, datum.Name)
			case "dataClassification":
				assert.Equal(t, tt.assertValue, datum.DataClassification)
			}
		})
	}
}

func TestDatum_Update_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)
	baseDatumID := factory.NewDatum(owner, profileID).WithName("Validation Test Datum").Create()

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
			setup: func() string { return baseDatumID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "<script>alert('xss')</script>"}
			},
			wantErrorContains: "HTML",
		},
		{
			name:  "name with newline",
			setup: func() string { return baseDatumID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\nDatum"}
			},
			wantErrorContains: "newline",
		},
		{
			name:  "name with carriage return",
			setup: func() string { return baseDatumID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\rDatum"}
			},
			wantErrorContains: "carriage return",
		},
		{
			name:  "name with null byte",
			setup: func() string { return baseDatumID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\x00Datum"}
			},
			wantErrorContains: "control character",
		},
		{
			name:  "name with tab character",
			setup: func() string { return baseDatumID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\tDatum"}
			},
			wantErrorContains: "control character",
		},
		{
			name:  "name with zero-width space",
			setup: func() string { return baseDatumID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\u200BDatum"}
			},
			wantErrorContains: "zero-width",
		},
		{
			name:  "name with zero-width joiner",
			setup: func() string { return baseDatumID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\u200DDatum"}
			},
			wantErrorContains: "zero-width",
		},
		{
			name:  "name with right-to-left override",
			setup: func() string { return baseDatumID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\u202EDatum"}
			},
			wantErrorContains: "bidirectional",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			datumID := tt.setup()

			query := `
				mutation UpdateDatum($input: UpdateDatumInput!) {
					updateDatum(input: $input) {
						datum {
							id
						}
					}
				}
			`

			_, err := owner.Do(query, map[string]any{"input": tt.input(datumID)})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestDatum_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)

	t.Run("delete existing datum", func(t *testing.T) {
		datumID := factory.NewDatum(owner, profileID).WithName("Datum to Delete").Create()

		query := `
			mutation DeleteDatum($input: DeleteDatumInput!) {
				deleteDatum(input: $input) {
					deletedDatumId
				}
			}
		`

		var result struct {
			DeleteDatum struct {
				DeletedDatumID string `json:"deletedDatumId"`
			} `json:"deleteDatum"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{"datumId": datumID},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, datumID, result.DeleteDatum.DeletedDatumID)
	})
}

func TestDatum_Delete_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	tests := []struct {
		name              string
		datumID           string
		wantErrorContains string
	}{
		{
			name:              "invalid ID format",
			datumID:           "invalid-id-format",
			wantErrorContains: "base64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation DeleteDatum($input: DeleteDatumInput!) {
					deleteDatum(input: $input) {
						deletedDatumId
					}
				}
			`

			_, err := owner.Do(query, map[string]any{
				"input": map[string]any{"datumId": tt.datumID},
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestDatum_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)

	datumNames := []string{"Datum A", "Datum B", "Datum C"}
	for _, name := range datumNames {
		factory.NewDatum(owner, profileID).WithName(name).Create()
	}

	query := `
		query GetData($id: ID!) {
			node(id: $id) {
				... on Organization {
					data(first: 10) {
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
			Data struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"data"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.Data.TotalCount, 3)
}

func TestDatum_Query(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("query with non-existent ID returns error", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Datum {
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

func TestDatum_Timestamps(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)

	t.Run("createdAt and updatedAt are set on create", func(t *testing.T) {
		beforeCreate := time.Now().Add(-time.Second)

		query := `
			mutation CreateDatum($input: CreateDatumInput!) {
				createDatum(input: $input) {
					datumEdge {
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
			CreateDatum struct {
				DatumEdge struct {
					Node struct {
						ID        string    `json:"id"`
						CreatedAt time.Time `json:"createdAt"`
						UpdatedAt time.Time `json:"updatedAt"`
					} `json:"node"`
				} `json:"datumEdge"`
			} `json:"createDatum"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId":     owner.GetOrganizationID().String(),
				"ownerId":            profileID,
				"name":               "Timestamp Test Datum",
				"dataClassification": "INTERNAL",
			},
		}, &result)
		require.NoError(t, err)

		node := result.CreateDatum.DatumEdge.Node
		testutil.AssertTimestampsOnCreate(t, node.CreatedAt, node.UpdatedAt, beforeCreate)
	})

	t.Run("updatedAt changes on update", func(t *testing.T) {
		datumID := factory.NewDatum(owner, profileID).WithName("Timestamp Update Test").Create()

		getQuery := `
			query($id: ID!) {
				node(id: $id) {
					... on Datum {
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

		err := owner.Execute(getQuery, map[string]any{"id": datumID}, &getResult)
		require.NoError(t, err)

		initialCreatedAt := getResult.Node.CreatedAt
		initialUpdatedAt := getResult.Node.UpdatedAt

		// Wait long enough for timestamp to change (database may have second precision)
		time.Sleep(1100 * time.Millisecond)

		updateQuery := `
			mutation UpdateDatum($input: UpdateDatumInput!) {
				updateDatum(input: $input) {
					datum {
						createdAt
						updatedAt
					}
				}
			}
		`

		var updateResult struct {
			UpdateDatum struct {
				Datum struct {
					CreatedAt time.Time `json:"createdAt"`
					UpdatedAt time.Time `json:"updatedAt"`
				} `json:"datum"`
			} `json:"updateDatum"`
		}

		err = owner.Execute(updateQuery, map[string]any{
			"input": map[string]any{
				"id":   datumID,
				"name": "Updated Timestamp Test",
			},
		}, &updateResult)
		require.NoError(t, err)

		datum := updateResult.UpdateDatum.Datum
		testutil.AssertTimestampsOnUpdate(t, datum.CreatedAt, datum.UpdatedAt, initialCreatedAt, initialUpdatedAt)
	})
}

func TestDatum_SubResolvers(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)
	datumID := factory.NewDatum(owner, profileID).WithName("SubResolver Test Datum").Create()

	t.Run("owner sub-resolver", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Datum {
						id
						owner {
							id
							fullName
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID    string `json:"id"`
				Owner struct {
					ID       string `json:"id"`
					FullName string `json:"fullName"`
				} `json:"owner"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": datumID}, &result)
		require.NoError(t, err)
		assert.Equal(t, profileID, result.Node.Owner.ID)
	})

	t.Run("organization sub-resolver", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Datum {
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

		err := owner.Execute(query, map[string]any{"id": datumID}, &result)
		require.NoError(t, err)
		assert.Equal(t, owner.GetOrganizationID().String(), result.Node.Organization.ID)
		assert.NotEmpty(t, result.Node.Organization.Name)
	})
}

func TestDatum_RBAC(t *testing.T) {
	t.Parallel()

	org := testutil.NewOrganizationRoles(t)

	const createDatumMutation = `
		mutation CreateDatum($input: CreateDatumInput!) {
			createDatum(input: $input) {
				datumEdge { node { id } }
			}
		}
	`

	const updateDatumMutation = `
		mutation UpdateDatum($input: UpdateDatumInput!) {
			updateDatum(input: $input) {
				datum { id }
			}
		}
	`

	const deleteDatumMutation = `
		mutation DeleteDatum($input: DeleteDatumInput!) {
			deleteDatum(input: $input) {
				deletedDatumId
			}
		}
	`

	const readDatumQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on Datum { id name }
			}
		}
	`

	t.Run(
		"create",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name         string
				role         testutil.TestRole
				forbidden    bool
				allowMsg     string
				forbiddenMsg string
			}{
				{
					name:     "owner can create",
					role:     testutil.RoleOwner,
					allowMsg: "owner should be able to create datum",
				},
				{
					name:     "admin can create",
					role:     testutil.RoleAdmin,
					allowMsg: "admin should be able to create datum",
				},
				{
					name:         "viewer cannot create",
					role:         testutil.RoleViewer,
					forbidden:    true,
					forbiddenMsg: "viewer should not be able to create datum",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						client := org.Client(t, tt.role)
						ownerClient := org.Client(t, testutil.RoleOwner)
						profileID := factory.CreateUser(ownerClient)

						_, err := client.Do(
							createDatumMutation,
							map[string]any{
								"input": map[string]any{
									"organizationId":     client.GetOrganizationID().String(),
									"ownerId":            profileID,
									"name":               factory.SafeName("RBAC create " + string(tt.role)),
									"dataClassification": "INTERNAL",
								},
							},
						)
						if tt.forbidden {
							testutil.RequireForbiddenError(t, err, tt.forbiddenMsg)
						} else {
							require.NoError(t, err, tt.allowMsg)
						}
					},
				)
			}
		},
	)

	t.Run(
		"update",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name         string
				role         testutil.TestRole
				forbidden    bool
				updateName   string
				allowMsg     string
				forbiddenMsg string
			}{
				{
					name:       "owner can update",
					role:       testutil.RoleOwner,
					updateName: "Updated by Owner",
					allowMsg:   "owner should be able to update datum",
				},
				{
					name:       "admin can update",
					role:       testutil.RoleAdmin,
					updateName: "Updated by Admin",
					allowMsg:   "admin should be able to update datum",
				},
				{
					name:         "viewer cannot update",
					role:         testutil.RoleViewer,
					forbidden:    true,
					updateName:   "Updated by Viewer",
					forbiddenMsg: "viewer should not be able to update datum",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ownerClient := org.Client(t, testutil.RoleOwner)
						profileID := factory.CreateUser(ownerClient)
						datumID := factory.NewDatum(ownerClient, profileID).
							WithName(factory.SafeName("RBAC update " + string(tt.role))).
							Create()

						client := org.Client(t, tt.role)

						_, err := client.Do(
							updateDatumMutation,
							map[string]any{
								"input": map[string]any{
									"id":   datumID,
									"name": tt.updateName,
								},
							},
						)
						if tt.forbidden {
							testutil.RequireForbiddenError(t, err, tt.forbiddenMsg)
						} else {
							require.NoError(t, err, tt.allowMsg)
						}
					},
				)
			}
		},
	)

	t.Run(
		"delete",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name         string
				role         testutil.TestRole
				forbidden    bool
				allowMsg     string
				forbiddenMsg string
			}{
				{
					name:     "owner can delete",
					role:     testutil.RoleOwner,
					allowMsg: "owner should be able to delete datum",
				},
				{
					name:     "admin can delete",
					role:     testutil.RoleAdmin,
					allowMsg: "admin should be able to delete datum",
				},
				{
					name:         "viewer cannot delete",
					role:         testutil.RoleViewer,
					forbidden:    true,
					forbiddenMsg: "viewer should not be able to delete datum",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ownerClient := org.Client(t, testutil.RoleOwner)
						profileID := factory.CreateUser(ownerClient)
						datumID := factory.NewDatum(ownerClient, profileID).
							WithName(factory.SafeName("RBAC delete " + string(tt.role))).
							Create()

						client := org.Client(t, tt.role)

						_, err := client.Do(
							deleteDatumMutation,
							map[string]any{
								"input": map[string]any{
									"datumId": datumID,
								},
							},
						)
						if tt.forbidden {
							testutil.RequireForbiddenError(t, err, tt.forbiddenMsg)
						} else {
							require.NoError(t, err, tt.allowMsg)
						}
					},
				)
			}
		},
	)

	t.Run(
		"read",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name        string
				role        testutil.TestRole
				allowMsg    string
				nodePresent string
			}{
				{
					name:        "owner can read",
					role:        testutil.RoleOwner,
					allowMsg:    "owner should be able to read datum",
					nodePresent: "owner should receive datum data",
				},
				{
					name:        "admin can read",
					role:        testutil.RoleAdmin,
					allowMsg:    "admin should be able to read datum",
					nodePresent: "admin should receive datum data",
				},
				{
					name:        "viewer can read",
					role:        testutil.RoleViewer,
					allowMsg:    "viewer should be able to read datum",
					nodePresent: "viewer should receive datum data",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ownerClient := org.Client(t, testutil.RoleOwner)
						profileID := factory.CreateUser(ownerClient)
						datumID := factory.NewDatum(ownerClient, profileID).
							WithName(factory.SafeName("RBAC read " + string(tt.role))).
							Create()

						client := org.Client(t, tt.role)

						var result struct {
							Node *struct {
								ID   string `json:"id"`
								Name string `json:"name"`
							} `json:"node"`
						}

						err := client.Execute(
							readDatumQuery,
							map[string]any{
								"id": datumID,
							},
							&result,
						)
						require.NoError(t, err, tt.allowMsg)
						require.NotNil(t, result.Node, tt.nodePresent)
					},
				)
			}
		},
	)
}

func TestDatum_MaxLength_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)

	longName := strings.Repeat("a", 1001)

	t.Run("create", func(t *testing.T) {
		query := `
			mutation CreateDatum($input: CreateDatumInput!) {
				createDatum(input: $input) {
					datumEdge {
						node { id }
					}
				}
			}
		`

		_, err := owner.Do(query, map[string]any{
			"input": map[string]any{
				"organizationId":     owner.GetOrganizationID().String(),
				"ownerId":            profileID,
				"name":               longName,
				"dataClassification": "INTERNAL",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("update", func(t *testing.T) {
		datumID := factory.NewDatum(owner, profileID).WithName("Max Length Test").Create()

		query := `
			mutation UpdateDatum($input: UpdateDatumInput!) {
				updateDatum(input: $input) {
					datum { id }
				}
			}
		`

		_, err := owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":   datumID,
				"name": longName,
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})
}

func TestDatum_Pagination(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)

	for i := range 5 {
		factory.NewDatum(owner, profileID).
			WithName(fmt.Sprintf("Pagination Datum %d", i)).
			Create()
	}

	t.Run("first/after pagination", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						data(first: 2) {
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
				Data struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
						Cursor string `json:"cursor"`
					} `json:"edges"`
					PageInfo   testutil.PageInfo `json:"pageInfo"`
					TotalCount int               `json:"totalCount"`
				} `json:"data"`
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

		testutil.AssertFirstPage(t, len(result.Node.Data.Edges), result.Node.Data.PageInfo, 2, true)
		assert.GreaterOrEqual(t, result.Node.Data.TotalCount, 5)

		testutil.AssertHasMorePages(t, result.Node.Data.PageInfo)

		queryAfter := `
			query($id: ID!, $after: CursorKey) {
				node(id: $id) {
					... on Organization {
						data(first: 2, after: $after) {
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
				Data struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo testutil.PageInfo `json:"pageInfo"`
				} `json:"data"`
			} `json:"node"`
		}

		err = owner.Execute(queryAfter, map[string]any{
			"id":    owner.GetOrganizationID().String(),
			"after": *result.Node.Data.PageInfo.EndCursor,
		}, &resultAfter)
		require.NoError(t, err)

		testutil.AssertMiddlePage(t, len(resultAfter.Node.Data.Edges), resultAfter.Node.Data.PageInfo, 2)
	})

	t.Run("last/before pagination", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						data(last: 2) {
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
				Data struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo testutil.PageInfo `json:"pageInfo"`
				} `json:"data"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id": owner.GetOrganizationID().String(),
		}, &result)
		require.NoError(t, err)

		testutil.AssertLastPage(t, len(result.Node.Data.Edges), result.Node.Data.PageInfo, 2, true)
	})
}

func TestDatum_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	profileID := factory.CreateUser(org1Owner)
	datumID := factory.NewDatum(org1Owner, profileID).WithName("Org1 Datum").Create()

	t.Run("cannot read datum from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Datum {
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

		err := org2Owner.Execute(query, map[string]any{"id": datumID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "datum")
	})

	t.Run("cannot update datum from another organization", func(t *testing.T) {
		query := `
			mutation UpdateDatum($input: UpdateDatumInput!) {
				updateDatum(input: $input) {
					datum { id }
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":   datumID,
				"name": "Hijacked Datum",
			},
		})
		require.Error(t, err, "Should not be able to update datum from another org")
	})

	t.Run("cannot delete datum from another organization", func(t *testing.T) {
		query := `
			mutation DeleteDatum($input: DeleteDatumInput!) {
				deleteDatum(input: $input) {
					deletedDatumId
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"datumId": datumID,
			},
		})
		require.Error(t, err, "Should not be able to delete datum from another org")
	})

	t.Run("cannot list data from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						data(first: 100) {
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
				Data struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"data"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{
			"id": org1Owner.GetOrganizationID().String(),
		}, &result)
		if err == nil {
			for _, edge := range result.Node.Data.Edges {
				assert.NotEqual(t, datumID, edge.Node.ID, "Should not see datum from another org")
			}
		}
	})

	t.Run("cannot create datum referencing a thirdParty from another organization", func(t *testing.T) {
		org2ThirdPartyID := factory.NewThirdParty(org2Owner).WithName("Org2 ThirdParty").Create()

		_, err := org1Owner.Do(`
			mutation($input: CreateDatumInput!) {
				createDatum(input: $input) {
					datumEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"organizationId":     org1Owner.GetOrganizationID().String(),
				"name":               factory.SafeName("Datum"),
				"dataClassification": "CONFIDENTIAL",
				"ownerId":            profileID,
				"thirdPartyIds":      []string{org2ThirdPartyID},
			},
		})
		require.Error(t, err, "must not accept a thirdPartyId belonging to another organization")
	})

	t.Run("cannot update datum to reference a thirdParty from another organization", func(t *testing.T) {
		org2ThirdPartyID := factory.NewThirdParty(org2Owner).WithName("Org2 ThirdParty for Update").Create()
		otherDatumID := factory.NewDatum(org1Owner, profileID).WithName("Org1 Datum for ThirdPartyIDs").Create()

		_, err := org1Owner.Do(`
			mutation($input: UpdateDatumInput!) {
				updateDatum(input: $input) {
					datum { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":            otherDatumID,
				"thirdPartyIds": []string{org2ThirdPartyID},
			},
		})
		require.Error(t, err, "must not accept a thirdPartyId belonging to another organization")
	})
}

func TestDatum_Ordering(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	profileID := factory.CreateUser(owner)

	factory.NewDatum(owner, profileID).WithName("AAA Order Test").Create()
	factory.NewDatum(owner, profileID).WithName("ZZZ Order Test").Create()

	t.Run("order by created_at descending", func(t *testing.T) {
		query := `
			query($id: ID!, $orderBy: DatumOrder) {
				node(id: $id) {
					... on Organization {
						data(first: 100, orderBy: $orderBy) {
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
				Data struct {
					Edges []struct {
						Node struct {
							ID        string    `json:"id"`
							CreatedAt time.Time `json:"createdAt"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"data"`
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

		times := make([]time.Time, len(result.Node.Data.Edges))
		for i, edge := range result.Node.Data.Edges {
			times[i] = edge.Node.CreatedAt
		}

		testutil.AssertTimesOrderedDescending(t, times, "createdAt")
	})
}
