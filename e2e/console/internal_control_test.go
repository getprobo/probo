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

func TestMeasure_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	tests := []struct {
		name        string
		input       map[string]any
		wantError   bool
		assertField string
		assertValue string
	}{
		{
			name: "with full details",
			input: map[string]any{
				"name":        "Owner Internal control",
				"description": "Created by owner",
				"category":    "POLICY",
			},
			assertField: "name",
			assertValue: "Owner Internal control",
		},
		{
			name: "with POLICY category",
			input: map[string]any{
				"name":     "Policy internal control",
				"category": "POLICY",
			},
			assertField: "category",
			assertValue: "POLICY",
		},
		{
			name: "with PROCEDURE category",
			input: map[string]any{
				"name":     "Procedure internal control",
				"category": "PROCEDURE",
			},
			assertField: "category",
			assertValue: "PROCEDURE",
		},
		{
			name: "with TECHNICAL category",
			input: map[string]any{
				"name":     "Technical internal control",
				"category": "TECHNICAL",
			},
			assertField: "category",
			assertValue: "TECHNICAL",
		},
		{
			name: "with EVIDENCE category",
			input: map[string]any{
				"name":     "Evidence internal control",
				"category": "EVIDENCE",
			},
			assertField: "category",
			assertValue: "EVIDENCE",
		},
		{
			name: "with TRAINING category",
			input: map[string]any{
				"name":     "Training internal control",
				"category": "TRAINING",
			},
			assertField: "category",
			assertValue: "TRAINING",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation CreateInternalControl($input: CreateInternalControlInput!) {
					createInternalControl(input: $input) {
						internalControlEdge {
							node {
								id
								name
								category
							}
						}
					}
				}
			`

			input := map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
			}
			maps.Copy(input, tt.input)

			var result struct {
				CreateInternalControl struct {
					InternalControlEdge struct {
						Node struct {
							ID       string `json:"id"`
							Name     string `json:"name"`
							Category string `json:"category"`
						} `json:"node"`
					} `json:"internalControlEdge"`
				} `json:"createInternalControl"`
			}

			err := owner.Execute(query, map[string]any{"input": input}, &result)
			require.NoError(t, err)

			node := result.CreateInternalControl.InternalControlEdge.Node
			assert.NotEmpty(t, node.ID)

			switch tt.assertField {
			case "name":
				assert.Equal(t, tt.assertValue, node.Name)
			case "category":
				assert.Equal(t, tt.assertValue, node.Category)
			}
		})
	}
}

func TestMeasure_Create_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	tests := []struct {
		name              string
		input             map[string]any
		skipOrganization  bool
		wantErrorContains string
	}{
		// Required field validation
		{
			name: "missing name",
			input: map[string]any{
				"category": "POLICY",
			},
			wantErrorContains: "name",
		},
		{
			name: "missing category",
			input: map[string]any{
				"name": "Test Internal control",
			},
			wantErrorContains: "category",
		},
		{
			name: "empty name",
			input: map[string]any{
				"name":     "",
				"category": "POLICY",
			},
			wantErrorContains: "name",
		},
		{
			name: "empty category",
			input: map[string]any{
				"name":     "Test Internal control",
				"category": "",
			},
			wantErrorContains: "category",
		},
		{
			name: "missing organizationId",
			input: map[string]any{
				"name":     "Test Internal control",
				"category": "POLICY",
			},
			skipOrganization:  true,
			wantErrorContains: "organizationId",
		},
		// HTML injection validation
		{
			name: "name with HTML tags",
			input: map[string]any{
				"name":     "<script>alert('xss')</script>",
				"category": "POLICY",
			},
			wantErrorContains: "HTML",
		},
		{
			name: "description with HTML tags",
			input: map[string]any{
				"name":        "Test Internal control",
				"category":    "POLICY",
				"description": "<div>HTML content</div>",
			},
			wantErrorContains: "HTML",
		},
		{
			name: "category with HTML tags",
			input: map[string]any{
				"name":     "Test Internal control",
				"category": "<b>POLICY</b>",
			},
			wantErrorContains: "HTML",
		},
		// Newline validation (name should not allow newlines)
		{
			name: "name with newline",
			input: map[string]any{
				"name":     "Test\nInternalControl",
				"category": "POLICY",
			},
			wantErrorContains: "newline",
		},
		{
			name: "name with carriage return",
			input: map[string]any{
				"name":     "Test\rInternalControl",
				"category": "POLICY",
			},
			wantErrorContains: "carriage return",
		},
		// Control character validation
		{
			name: "name with null byte",
			input: map[string]any{
				"name":     "Test\x00Measure",
				"category": "POLICY",
			},
			wantErrorContains: "control character",
		},
		{
			name: "name with tab character",
			input: map[string]any{
				"name":     "Test\tInternalControl",
				"category": "POLICY",
			},
			wantErrorContains: "control character",
		},
		// Zero-width character validation
		{
			name: "name with zero-width space",
			input: map[string]any{
				"name":     "Test\u200BMeasure",
				"category": "POLICY",
			},
			wantErrorContains: "zero-width",
		},
		{
			name: "name with zero-width joiner",
			input: map[string]any{
				"name":     "Test\u200DMeasure",
				"category": "POLICY",
			},
			wantErrorContains: "zero-width",
		},
		// Bidirectional override validation
		{
			name: "name with right-to-left override",
			input: map[string]any{
				"name":     "Test\u202EMeasure",
				"category": "POLICY",
			},
			wantErrorContains: "bidirectional",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation CreateInternalControl($input: CreateInternalControlInput!) {
					createInternalControl(input: $input) {
						internalControlEdge {
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

			maps.Copy(input, tt.input)

			_, err := owner.Do(query, map[string]any{"input": input})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestMeasure_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	tests := []struct {
		name        string
		setup       func() string
		input       func(id string) map[string]any
		assertField string
		assertValue string
	}{
		{
			name: "update name and description",
			setup: func() string {
				return factory.NewInternalControl(owner).
					WithName("Internal control to Update").
					WithDescription("Original description").
					Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{
					"id":          id,
					"name":        "Updated by Owner",
					"description": "Owner updated this",
				}
			},
			assertField: "name",
			assertValue: "Updated by Owner",
		},
		{
			name: "update to NOT_STARTED state",
			setup: func() string {
				return factory.NewInternalControl(owner).WithName("State Test").Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "state": "NOT_STARTED"}
			},
			assertField: "state",
			assertValue: "NOT_STARTED",
		},
		{
			name: "update to IMPLEMENTED state",
			setup: func() string {
				return factory.NewInternalControl(owner).WithName("State Test").Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "state": "IMPLEMENTED"}
			},
			assertField: "state",
			assertValue: "IMPLEMENTED",
		},
		{
			name: "update to NOT_APPLICABLE state",
			setup: func() string {
				return factory.NewInternalControl(owner).WithName("State Test").Create()
			},
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "state": "NOT_APPLICABLE"}
			},
			assertField: "state",
			assertValue: "NOT_APPLICABLE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			internalControlID := tt.setup()

			query := `
				mutation UpdateInternalControl($input: UpdateInternalControlInput!) {
					updateInternalControl(input: $input) {
						internalControl {
							id
							name
							state
						}
					}
				}
			`

			var result struct {
				UpdateInternalControl struct {
					InternalControl struct {
						ID    string `json:"id"`
						Name  string `json:"name"`
						State string `json:"state"`
					} `json:"internalControl"`
				} `json:"updateInternalControl"`
			}

			err := owner.Execute(query, map[string]any{"input": tt.input(internalControlID)}, &result)
			require.NoError(t, err)

			internalControl := result.UpdateInternalControl.InternalControl

			switch tt.assertField {
			case "name":
				assert.Equal(t, tt.assertValue, internalControl.Name)
			case "state":
				assert.Equal(t, tt.assertValue, internalControl.State)
			}
		})
	}
}

func TestMeasure_Update_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create an internal control to use for most validation tests
	baseInternalControlID := factory.NewInternalControl(owner).WithName("Validation Test Internal control").Create()

	tests := []struct {
		name              string
		setup             func() string
		input             func(id string) map[string]any
		wantErrorContains string
	}{
		// ID validation
		{
			name:  "invalid ID format",
			setup: func() string { return "invalid-id-format" },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test"}
			},
			wantErrorContains: "base64",
		},
		// Empty field validation
		{
			name:  "empty name",
			setup: func() string { return baseInternalControlID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": ""}
			},
			wantErrorContains: "name",
		},
		{
			name:  "empty category",
			setup: func() string { return baseInternalControlID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "category": ""}
			},
			wantErrorContains: "category",
		},
		// HTML injection validation
		{
			name:  "name with HTML tags",
			setup: func() string { return baseInternalControlID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "<script>alert('xss')</script>"}
			},
			wantErrorContains: "HTML",
		},
		{
			name:  "description with HTML tags",
			setup: func() string { return baseInternalControlID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "description": "<img src=x onerror=alert(1)>"}
			},
			wantErrorContains: "HTML",
		},
		{
			name:  "category with HTML tags",
			setup: func() string { return baseInternalControlID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "category": "<b>POLICY</b>"}
			},
			wantErrorContains: "HTML",
		},
		// Newline validation (name should not allow newlines)
		{
			name:  "name with newline",
			setup: func() string { return baseInternalControlID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\nInternalControl"}
			},
			wantErrorContains: "newline",
		},
		{
			name:  "name with carriage return",
			setup: func() string { return baseInternalControlID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\rInternalControl"}
			},
			wantErrorContains: "carriage return",
		},
		// Control character validation
		{
			name:  "name with null byte",
			setup: func() string { return baseInternalControlID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\x00Measure"}
			},
			wantErrorContains: "control character",
		},
		{
			name:  "name with tab character",
			setup: func() string { return baseInternalControlID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\tInternalControl"}
			},
			wantErrorContains: "control character",
		},
		{
			name:  "description with null byte",
			setup: func() string { return baseInternalControlID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "description": "Test\x00Description"}
			},
			wantErrorContains: "control character",
		},
		// Zero-width character validation
		{
			name:  "name with zero-width space",
			setup: func() string { return baseInternalControlID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\u200BMeasure"}
			},
			wantErrorContains: "zero-width",
		},
		{
			name:  "description with zero-width joiner",
			setup: func() string { return baseInternalControlID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "description": "Test\u200DDescription"}
			},
			wantErrorContains: "zero-width",
		},
		// Bidirectional override validation
		{
			name:  "name with right-to-left override",
			setup: func() string { return baseInternalControlID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\u202EMeasure"}
			},
			wantErrorContains: "bidirectional",
		},
		{
			name:  "description with left-to-right override",
			setup: func() string { return baseInternalControlID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "description": "Test\u202DDescription"}
			},
			wantErrorContains: "bidirectional",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			internalControlID := tt.setup()

			query := `
				mutation UpdateInternalControl($input: UpdateInternalControlInput!) {
					updateInternalControl(input: $input) {
						internalControl {
							id
						}
					}
				}
			`

			_, err := owner.Do(query, map[string]any{"input": tt.input(internalControlID)})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestMeasure_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("delete existing internal control", func(t *testing.T) {
		internalControlID := factory.NewInternalControl(owner).WithName("Internal control to Delete").Create()

		query := `
			mutation DeleteInternalControl($input: DeleteInternalControlInput!) {
				deleteInternalControl(input: $input) {
					deletedInternalControlId
				}
			}
		`

		var result struct {
			DeleteInternalControl struct {
				DeletedInternalControlID string `json:"deletedInternalControlId"`
			} `json:"deleteInternalControl"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{"internalControlId": internalControlID},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, internalControlID, result.DeleteInternalControl.DeletedInternalControlID)
	})
}

func TestMeasure_Delete_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	tests := []struct {
		name              string
		internalControlID string
		wantErrorContains string
	}{
		{
			name:              "invalid ID format",
			internalControlID: "invalid-id-format",
			wantErrorContains: "base64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation DeleteInternalControl($input: DeleteInternalControlInput!) {
					deleteInternalControl(input: $input) {
						deletedInternalControlId
					}
				}
			`

			_, err := owner.Do(query, map[string]any{
				"input": map[string]any{"internalControlId": tt.internalControlID},
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestMeasure_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	internalControlNames := []string{"Internal control A", "Internal control B", "Internal control C"}
	for _, name := range internalControlNames {
		factory.NewInternalControl(owner).WithName(name).Create()
	}

	query := `
		query GetInternalControls($id: ID!) {
			node(id: $id) {
				... on Organization {
					internalControls(first: 10) {
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
			InternalControls struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"internalControls"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.InternalControls.TotalCount, 3)
}

func TestMeasure_Query(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("query with non-existent ID returns error", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on InternalControl {
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

func TestMeasure_Timestamps(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("createdAt and updatedAt are set on create", func(t *testing.T) {
		beforeCreate := time.Now().Add(-time.Second)

		query := `
			mutation CreateInternalControl($input: CreateInternalControlInput!) {
				createInternalControl(input: $input) {
					internalControlEdge {
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
			CreateInternalControl struct {
				InternalControlEdge struct {
					Node struct {
						ID        string    `json:"id"`
						CreatedAt time.Time `json:"createdAt"`
						UpdatedAt time.Time `json:"updatedAt"`
					} `json:"node"`
				} `json:"internalControlEdge"`
			} `json:"createInternalControl"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           "Timestamp Test Internal control",
				"category":       "POLICY",
			},
		}, &result)
		require.NoError(t, err)

		node := result.CreateInternalControl.InternalControlEdge.Node
		testutil.AssertTimestampsOnCreate(t, node.CreatedAt, node.UpdatedAt, beforeCreate)
	})

	t.Run("updatedAt changes on update", func(t *testing.T) {
		internalControlID := factory.NewInternalControl(owner).WithName("Timestamp Update Test").Create()

		// Get initial timestamps
		getQuery := `
			query($id: ID!) {
				node(id: $id) {
					... on InternalControl {
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

		err := owner.Execute(getQuery, map[string]any{"id": internalControlID}, &getResult)
		require.NoError(t, err)

		initialCreatedAt := getResult.Node.CreatedAt
		initialUpdatedAt := getResult.Node.UpdatedAt

		updateQuery := `
			mutation UpdateInternalControl($input: UpdateInternalControlInput!) {
				updateInternalControl(input: $input) {
					internalControl {
						createdAt
						updatedAt
					}
				}
			}
		`

		var updateResult struct {
			UpdateInternalControl struct {
				InternalControl struct {
					CreatedAt time.Time `json:"createdAt"`
					UpdatedAt time.Time `json:"updatedAt"`
				} `json:"internalControl"`
			} `json:"updateInternalControl"`
		}

		err = owner.Execute(updateQuery, map[string]any{
			"input": map[string]any{
				"id":   internalControlID,
				"name": "Updated Timestamp Test",
			},
		}, &updateResult)
		require.NoError(t, err)

		internalControl := updateResult.UpdateInternalControl.InternalControl
		testutil.AssertTimestampsOnUpdate(t, internalControl.CreatedAt, internalControl.UpdatedAt, initialCreatedAt, initialUpdatedAt)
	})
}

func TestMeasure_OmittableDescription(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	internalControlID := factory.NewInternalControl(owner).
		WithName("Omittable Test Internal control").
		WithDescription("Initial description").
		Create()

	tests := []struct {
		name            string
		input           map[string]any
		wantDescription *string
	}{
		{
			name:            "update with new description",
			input:           map[string]any{"description": "Updated description"},
			wantDescription: new("Updated description"),
		},
		{
			name:            "update with null description clears it",
			input:           map[string]any{"description": nil},
			wantDescription: nil,
		},
		{
			name:            "set description again",
			input:           map[string]any{"description": "Should persist"},
			wantDescription: new("Should persist"),
		},
		{
			name:            "update without description preserves it",
			input:           map[string]any{"name": "Updated Name"},
			wantDescription: new("Should persist"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation($input: UpdateInternalControlInput!) {
					updateInternalControl(input: $input) {
						internalControl {
							id
							name
							description
						}
					}
				}
			`

			input := map[string]any{"id": internalControlID}
			maps.Copy(input, tt.input)

			var result struct {
				UpdateInternalControl struct {
					InternalControl struct {
						ID          string  `json:"id"`
						Name        string  `json:"name"`
						Description *string `json:"description"`
					} `json:"internalControl"`
				} `json:"updateInternalControl"`
			}

			err := owner.Execute(query, map[string]any{"input": input}, &result)
			require.NoError(t, err)

			testutil.AssertOptionalStringEqual(t, tt.wantDescription, result.UpdateInternalControl.InternalControl.Description, "description")
		})
	}
}

func TestMeasure_SubResolvers(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	internalControlID := factory.NewInternalControl(owner).
		WithName("SubResolver Test Internal control").
		Create()

	subResolvers := []struct {
		name  string
		field string
	}{
		{"controls", "controls"},
		{"risks", "risks"},
		{"tasks", "tasks"},
		{"evidences", "evidences"},
	}

	for _, sr := range subResolvers {
		t.Run(sr.name+" sub-resolver returns empty list", func(t *testing.T) {
			query := fmt.Sprintf(`
				query($id: ID!) {
					node(id: $id) {
						... on InternalControl {
							id
							%s(first: 10) {
								edges {
									node {
										id
									}
								}
							}
						}
					}
				}
			`, sr.field)

			resp, err := owner.Do(query, map[string]any{"id": internalControlID})
			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}

func TestMeasure_MaxLength_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// TitleMaxLength = 1000, ContentMaxLength = 5000
	longName := strings.Repeat("a", 1001)
	longCategory := strings.Repeat("b", 1001)
	longDescription := strings.Repeat("c", 5001)

	t.Run("create", func(t *testing.T) {
		tests := []struct {
			name              string
			input             map[string]any
			wantErrorContains string
		}{
			{
				name: "name exceeds max length",
				input: map[string]any{
					"name":     longName,
					"category": "POLICY",
				},
				wantErrorContains: "name",
			},
			{
				name: "category exceeds max length",
				input: map[string]any{
					"name":     "Test Internal control",
					"category": longCategory,
				},
				wantErrorContains: "category",
			},
			{
				name: "description exceeds max length",
				input: map[string]any{
					"name":        "Test Internal control",
					"category":    "POLICY",
					"description": longDescription,
				},
				wantErrorContains: "description",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				query := `
					mutation CreateInternalControl($input: CreateInternalControlInput!) {
						createInternalControl(input: $input) {
							internalControlEdge {
								node { id }
							}
						}
					}
				`

				input := map[string]any{
					"organizationId": owner.GetOrganizationID().String(),
				}
				maps.Copy(input, tt.input)

				_, err := owner.Do(query, map[string]any{"input": input})
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrorContains)
			})
		}
	})

	t.Run("update", func(t *testing.T) {
		internalControlID := factory.NewInternalControl(owner).WithName("Max Length Test").Create()

		tests := []struct {
			name              string
			input             map[string]any
			wantErrorContains string
		}{
			{
				name:              "name exceeds max length",
				input:             map[string]any{"name": longName},
				wantErrorContains: "name",
			},
			{
				name:              "category exceeds max length",
				input:             map[string]any{"category": longCategory},
				wantErrorContains: "category",
			},
			{
				name:              "description exceeds max length",
				input:             map[string]any{"description": longDescription},
				wantErrorContains: "description",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				query := `
					mutation UpdateInternalControl($input: UpdateInternalControlInput!) {
						updateInternalControl(input: $input) {
							internalControl { id }
						}
					}
				`

				input := map[string]any{"id": internalControlID}
				maps.Copy(input, tt.input)

				_, err := owner.Do(query, map[string]any{"input": input})
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrorContains)
			})
		}
	})
}

func TestMeasure_SubResolvers_WithData(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("tasks sub-resolver with linked tasks", func(t *testing.T) {
		internalControlID := factory.NewInternalControl(owner).WithName("Internal control with Tasks").Create()

		// Create tasks linked to the internal control
		task1ID := factory.NewTask(owner, internalControlID).WithName("Task 1").Create()
		task2ID := factory.NewTask(owner, internalControlID).WithName("Task 2").Create()

		query := `
			query($id: ID!) {
				node(id: $id) {
					... on InternalControl {
						id
						tasks(first: 10) {
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
				ID    string `json:"id"`
				Tasks struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
					TotalCount int `json:"totalCount"`
				} `json:"tasks"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": internalControlID}, &result)
		require.NoError(t, err)
		assert.Equal(t, 2, result.Node.Tasks.TotalCount)

		taskIDs := make([]string, len(result.Node.Tasks.Edges))
		for i, edge := range result.Node.Tasks.Edges {
			taskIDs[i] = edge.Node.ID
		}

		assert.Contains(t, taskIDs, task1ID)
		assert.Contains(t, taskIDs, task2ID)
	})

	t.Run("controls sub-resolver with linked controls", func(t *testing.T) {
		internalControlID := factory.NewInternalControl(owner).WithName("Internal control with Controls").Create()

		// Create framework and control
		frameworkID := factory.NewFramework(owner).WithName("Test Framework").Create()
		controlID := factory.NewControl(owner, frameworkID).WithName("Test Control").Create()

		// Link control to internal control
		linkQuery := `
			mutation($input: CreateControlInternalControlMappingInput!) {
				createControlInternalControlMapping(input: $input) {
					controlEdge { node { id } }
				}
			}
		`
		_, err := owner.Do(linkQuery, map[string]any{
			"input": map[string]any{
				"controlId":         controlID,
				"internalControlId": internalControlID,
			},
		})
		require.NoError(t, err)

		// Query the internalControl's controls
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on InternalControl {
						id
						controls(first: 10) {
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
				ID       string `json:"id"`
				Controls struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
					TotalCount int `json:"totalCount"`
				} `json:"controls"`
			} `json:"node"`
		}

		err = owner.Execute(query, map[string]any{"id": internalControlID}, &result)
		require.NoError(t, err)
		assert.Equal(t, 1, result.Node.Controls.TotalCount)
		assert.Equal(t, controlID, result.Node.Controls.Edges[0].Node.ID)
	})

	t.Run("risks sub-resolver with linked risks", func(t *testing.T) {
		internalControlID := factory.NewInternalControl(owner).WithName("Internal control with Risks").Create()

		// Create risk
		riskID := factory.NewRisk(owner).WithName("Test Risk").Create()

		// Link risk to internal control
		linkQuery := `
			mutation($input: CreateRiskInternalControlMappingInput!) {
				createRiskInternalControlMapping(input: $input) {
					riskEdge { node { id } }
				}
			}
		`
		_, err := owner.Do(linkQuery, map[string]any{
			"input": map[string]any{
				"riskId":            riskID,
				"internalControlId": internalControlID,
			},
		})
		require.NoError(t, err)

		// Query the internalControl's risks
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on InternalControl {
						id
						risks(first: 10) {
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
				ID    string `json:"id"`
				Risks struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
					TotalCount int `json:"totalCount"`
				} `json:"risks"`
			} `json:"node"`
		}

		err = owner.Execute(query, map[string]any{"id": internalControlID}, &result)
		require.NoError(t, err)
		assert.Equal(t, 1, result.Node.Risks.TotalCount)
		assert.Equal(t, riskID, result.Node.Risks.Edges[0].Node.ID)
	})
}

func TestMeasure_Pagination(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create multiple internal controls for pagination testing
	internalControlIDs := make([]string, 5)
	for i := range 5 {
		internalControlIDs[i] = factory.NewInternalControl(owner).
			WithName(fmt.Sprintf("Pagination Internal control %d", i)).
			Create()
	}

	t.Run("first/after pagination", func(t *testing.T) {
		// Get first 2 internal controls
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						internalControls(first: 2) {
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
				InternalControls struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
						Cursor string `json:"cursor"`
					} `json:"edges"`
					PageInfo   testutil.PageInfo `json:"pageInfo"`
					TotalCount int               `json:"totalCount"`
				} `json:"internalControls"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id": owner.GetOrganizationID().String(),
		}, &result)
		require.NoError(t, err)

		testutil.AssertFirstPage(t, len(result.Node.InternalControls.Edges), result.Node.InternalControls.PageInfo, 2, true)
		assert.GreaterOrEqual(t, result.Node.InternalControls.TotalCount, 5)

		// Get next page using cursor
		testutil.AssertHasMorePages(t, result.Node.InternalControls.PageInfo)

		queryAfter := `
			query($id: ID!, $after: CursorKey) {
				node(id: $id) {
					... on Organization {
						internalControls(first: 2, after: $after) {
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
				InternalControls struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo testutil.PageInfo `json:"pageInfo"`
				} `json:"internalControls"`
			} `json:"node"`
		}

		err = owner.Execute(queryAfter, map[string]any{
			"id":    owner.GetOrganizationID().String(),
			"after": *result.Node.InternalControls.PageInfo.EndCursor,
		}, &resultAfter)
		require.NoError(t, err)

		testutil.AssertMiddlePage(t, len(resultAfter.Node.InternalControls.Edges), resultAfter.Node.InternalControls.PageInfo, 2)
	})

	t.Run("last/before pagination", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						internalControls(last: 2) {
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
				InternalControls struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo testutil.PageInfo `json:"pageInfo"`
				} `json:"internalControls"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id": owner.GetOrganizationID().String(),
		}, &result)
		require.NoError(t, err)

		testutil.AssertLastPage(t, len(result.Node.InternalControls.Edges), result.Node.InternalControls.PageInfo, 2, true)
	})
}

func TestMeasure_Filtering(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create internal controls with different states
	measure1ID := factory.NewInternalControl(owner).WithName("Filter Test Implemented").Create()
	measure2ID := factory.NewInternalControl(owner).WithName("Filter Test Not Started").Create()

	// Update measure1 to IMPLEMENTED state
	updateQuery := `
		mutation($input: UpdateInternalControlInput!) {
			updateInternalControl(input: $input) {
				internalControl { id state }
			}
		}
	`
	_, err := owner.Do(updateQuery, map[string]any{
		"input": map[string]any{
			"id":    measure1ID,
			"state": "IMPLEMENTED",
		},
	})
	require.NoError(t, err)

	t.Run("filter by state", func(t *testing.T) {
		query := `
			query($id: ID!, $filter: InternalControlFilter) {
				node(id: $id) {
					... on Organization {
						internalControls(first: 100, filter: $filter) {
							edges {
								node {
									id
									name
									state
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
				InternalControls struct {
					Edges []struct {
						Node struct {
							ID    string `json:"id"`
							Name  string `json:"name"`
							State string `json:"state"`
						} `json:"node"`
					} `json:"edges"`
					TotalCount int `json:"totalCount"`
				} `json:"internalControls"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id":     owner.GetOrganizationID().String(),
			"filter": map[string]any{"state": "IMPLEMENTED"},
		}, &result)
		require.NoError(t, err)

		// All returned internal controls should be IMPLEMENTED
		for _, edge := range result.Node.InternalControls.Edges {
			assert.Equal(t, "IMPLEMENTED", edge.Node.State)
		}

		// Should contain our implemented internal control
		found := false

		for _, edge := range result.Node.InternalControls.Edges {
			if edge.Node.ID == measure1ID {
				found = true
				break
			}
		}

		assert.True(t, found, "Expected to find implemented internal control in filtered results")
	})

	t.Run("filter by query string", func(t *testing.T) {
		query := `
			query($id: ID!, $filter: InternalControlFilter) {
				node(id: $id) {
					... on Organization {
						internalControls(first: 100, filter: $filter) {
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
				InternalControls struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"internalControls"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id":     owner.GetOrganizationID().String(),
			"filter": map[string]any{"query": "Filter Test"},
		}, &result)
		require.NoError(t, err)

		// Should find internal controls matching the query
		foundIDs := make([]string, len(result.Node.InternalControls.Edges))
		for i, edge := range result.Node.InternalControls.Edges {
			foundIDs[i] = edge.Node.ID
		}

		assert.Contains(t, foundIDs, measure1ID)
		assert.Contains(t, foundIDs, measure2ID)
	})
}

func TestMeasure_FilterByCategory(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create internal controls with different categories
	policyID := factory.NewInternalControl(owner).WithName("Category Policy Internal control").WithCategory("POLICY").Create()
	factory.NewInternalControl(owner).WithName("Category Technical Internal control").WithCategory("TECHNICAL").Create()

	t.Run("filter by category on organization", func(t *testing.T) {
		t.Parallel()

		const query = `
			query($id: ID!, $filter: InternalControlFilter) {
				node(id: $id) {
					... on Organization {
						internalControls(first: 100, filter: $filter) {
							edges {
								node {
									id
									category
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
				InternalControls struct {
					Edges []struct {
						Node struct {
							ID       string `json:"id"`
							Category string `json:"category"`
						} `json:"node"`
					} `json:"edges"`
					TotalCount int `json:"totalCount"`
				} `json:"internalControls"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id":     owner.GetOrganizationID().String(),
			"filter": map[string]any{"category": "POLICY"},
		}, &result)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, result.Node.InternalControls.TotalCount, 1)

		for _, edge := range result.Node.InternalControls.Edges {
			assert.Equal(t, "POLICY", edge.Node.Category)
		}

		found := false

		for _, edge := range result.Node.InternalControls.Edges {
			if edge.Node.ID == policyID {
				found = true
				break
			}
		}

		assert.True(t, found, "Expected to find POLICY internal control in filtered results")
	})

	t.Run("filter by category on risk", func(t *testing.T) {
		t.Parallel()

		riskID := factory.NewRisk(owner).WithName("Category Filter Risk").Create()

		policyInternalControlID := factory.NewInternalControl(owner).WithName("Risk Policy Internal control").WithCategory("POLICY").Create()
		techInternalControlID := factory.NewInternalControl(owner).WithName("Risk Technical Internal control").WithCategory("TECHNICAL").Create()

		// Link both internal controls to the risk
		const linkQuery = `
			mutation($input: CreateRiskInternalControlMappingInput!) {
				createRiskInternalControlMapping(input: $input) {
					riskEdge { node { id } }
				}
			}
		`
		for _, mID := range []string{policyInternalControlID, techInternalControlID} {
			_, err := owner.Do(linkQuery, map[string]any{
				"input": map[string]any{
					"riskId":            riskID,
					"internalControlId": mID,
				},
			})
			require.NoError(t, err)
		}

		var err error

		const query = `
			query($id: ID!, $filter: InternalControlFilter) {
				node(id: $id) {
					... on Risk {
						internalControls(first: 100, filter: $filter) {
							edges {
								node {
									id
									category
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
				InternalControls struct {
					Edges []struct {
						Node struct {
							ID       string `json:"id"`
							Category string `json:"category"`
						} `json:"node"`
					} `json:"edges"`
					TotalCount int `json:"totalCount"`
				} `json:"internalControls"`
			} `json:"node"`
		}

		err = owner.Execute(query, map[string]any{
			"id":     riskID,
			"filter": map[string]any{"category": "POLICY"},
		}, &result)
		require.NoError(t, err)

		assert.Equal(t, 1, result.Node.InternalControls.TotalCount)
		assert.Equal(t, policyInternalControlID, result.Node.InternalControls.Edges[0].Node.ID)
		assert.Equal(t, "POLICY", result.Node.InternalControls.Edges[0].Node.Category)
	})

	t.Run("filter by category on control", func(t *testing.T) {
		t.Parallel()

		frameworkID := factory.NewFramework(owner).WithName("Category Filter Framework").Create()
		controlID := factory.NewControl(owner, frameworkID).WithName("Category Filter Control").Create()

		policyInternalControlID := factory.NewInternalControl(owner).WithName("Control Policy Internal control").WithCategory("POLICY").Create()
		techInternalControlID := factory.NewInternalControl(owner).WithName("Control Technical Internal control").WithCategory("TECHNICAL").Create()

		// Link both internal controls to the control
		const linkQuery = `
			mutation($input: CreateControlInternalControlMappingInput!) {
				createControlInternalControlMapping(input: $input) {
					controlEdge { node { id } }
				}
			}
		`
		for _, mID := range []string{policyInternalControlID, techInternalControlID} {
			_, err := owner.Do(linkQuery, map[string]any{
				"input": map[string]any{
					"controlId":         controlID,
					"internalControlId": mID,
				},
			})
			require.NoError(t, err)
		}

		const query = `
			query($id: ID!, $filter: InternalControlFilter) {
				node(id: $id) {
					... on Control {
						internalControls(first: 100, filter: $filter) {
							edges {
								node {
									id
									category
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
				InternalControls struct {
					Edges []struct {
						Node struct {
							ID       string `json:"id"`
							Category string `json:"category"`
						} `json:"node"`
					} `json:"edges"`
					TotalCount int `json:"totalCount"`
				} `json:"internalControls"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id":     controlID,
			"filter": map[string]any{"category": "POLICY"},
		}, &result)
		require.NoError(t, err)

		assert.Equal(t, 1, result.Node.InternalControls.TotalCount)
		assert.Equal(t, policyInternalControlID, result.Node.InternalControls.Edges[0].Node.ID)
		assert.Equal(t, "POLICY", result.Node.InternalControls.Edges[0].Node.Category)
	})
}

func TestMeasure_Ordering(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create internal controls with distinct names for ordering
	factory.NewInternalControl(owner).WithName("AAA Order Test").Create()
	factory.NewInternalControl(owner).WithName("ZZZ Order Test").Create()

	t.Run("order by name ascending", func(t *testing.T) {
		query := `
			query($id: ID!, $orderBy: InternalControlOrder) {
				node(id: $id) {
					... on Organization {
						internalControls(first: 100, orderBy: $orderBy) {
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
				InternalControls struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"internalControls"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id": owner.GetOrganizationID().String(),
			"orderBy": map[string]any{
				"field":     "NAME",
				"direction": "ASC",
			},
		}, &result)
		require.NoError(t, err)

		// Verify ordering - names should be in ascending order
		names := make([]string, len(result.Node.InternalControls.Edges))
		for i, edge := range result.Node.InternalControls.Edges {
			names[i] = edge.Node.Name
		}

		testutil.AssertOrderedAscending(t, names, "name")
	})

	t.Run("order by name descending", func(t *testing.T) {
		query := `
			query($id: ID!, $orderBy: InternalControlOrder) {
				node(id: $id) {
					... on Organization {
						internalControls(first: 100, orderBy: $orderBy) {
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
				InternalControls struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"internalControls"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id": owner.GetOrganizationID().String(),
			"orderBy": map[string]any{
				"field":     "NAME",
				"direction": "DESC",
			},
		}, &result)
		require.NoError(t, err)

		// Verify ordering - names should be in descending order
		names := make([]string, len(result.Node.InternalControls.Edges))
		for i, edge := range result.Node.InternalControls.Edges {
			names[i] = edge.Node.Name
		}

		testutil.AssertOrderedDescending(t, names, "name")
	})

	t.Run("order by created_at", func(t *testing.T) {
		query := `
			query($id: ID!, $orderBy: InternalControlOrder) {
				node(id: $id) {
					... on Organization {
						internalControls(first: 100, orderBy: $orderBy) {
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
				InternalControls struct {
					Edges []struct {
						Node struct {
							ID        string    `json:"id"`
							CreatedAt time.Time `json:"createdAt"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"internalControls"`
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

		// Verify ordering - createdAt should be in descending order
		times := make([]time.Time, len(result.Node.InternalControls.Edges))
		for i, edge := range result.Node.InternalControls.Edges {
			times[i] = edge.Node.CreatedAt
		}

		testutil.AssertTimesOrderedDescending(t, times, "createdAt")
	})
}

func TestMeasure_ThirdPartyMapping(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	const createMutation = `
		mutation($input: CreateInternalControlThirdPartyMappingInput!) {
			createInternalControlThirdPartyMapping(input: $input) {
				internalControlEdge { node { id } }
				thirdPartyEdge { node { id } }
			}
		}
	`

	const deleteMutation = `
		mutation($input: DeleteInternalControlThirdPartyMappingInput!) {
			deleteInternalControlThirdPartyMapping(input: $input) {
				deletedInternalControlId
				deletedThirdPartyId
			}
		}
	`

	t.Run("create mapping links internal control to third party on both sides", func(t *testing.T) {
		t.Parallel()

		internalControlID := factory.NewInternalControl(owner).WithName("Mapping Internal control").Create()
		thirdPartyID := factory.NewThirdParty(owner).WithName("Mapping Third Party").Create()

		_, err := owner.Do(createMutation, map[string]any{
			"input": map[string]any{
				"internalControlId": internalControlID,
				"thirdPartyId":      thirdPartyID,
			},
		})
		require.NoError(t, err)

		const internalControlQuery = `
			query($id: ID!) {
				node(id: $id) {
					... on InternalControl {
						thirdParties(first: 10) {
							edges { node { id } }
							totalCount
						}
					}
				}
			}
		`

		var internalControlResult struct {
			Node struct {
				ThirdParties struct {
					Edges []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
					TotalCount int `json:"totalCount"`
				} `json:"thirdParties"`
			} `json:"node"`
		}

		err = owner.Execute(internalControlQuery, map[string]any{"id": internalControlID}, &internalControlResult)
		require.NoError(t, err)
		assert.Equal(t, 1, internalControlResult.Node.ThirdParties.TotalCount)
		assert.Equal(t, thirdPartyID, internalControlResult.Node.ThirdParties.Edges[0].Node.ID)

		const thirdPartyQuery = `
			query($id: ID!) {
				node(id: $id) {
					... on ThirdParty {
						internalControls(first: 10) {
							edges { node { id } }
							totalCount
						}
					}
				}
			}
		`

		var tpResult struct {
			Node struct {
				InternalControls struct {
					Edges []struct {
						Node struct {
							ID string `json:"id"`
						} `json:"node"`
					} `json:"edges"`
					TotalCount int `json:"totalCount"`
				} `json:"internalControls"`
			} `json:"node"`
		}

		err = owner.Execute(thirdPartyQuery, map[string]any{"id": thirdPartyID}, &tpResult)
		require.NoError(t, err)
		assert.Equal(t, 1, tpResult.Node.InternalControls.TotalCount)
		assert.Equal(t, internalControlID, tpResult.Node.InternalControls.Edges[0].Node.ID)
	})

	t.Run("create mapping is idempotent", func(t *testing.T) {
		t.Parallel()

		internalControlID := factory.NewInternalControl(owner).WithName("Idempotent Internal control").Create()
		thirdPartyID := factory.NewThirdParty(owner).WithName("Idempotent Third Party").Create()

		input := map[string]any{
			"internalControlId": internalControlID,
			"thirdPartyId":      thirdPartyID,
		}

		_, err := owner.Do(createMutation, map[string]any{"input": input})
		require.NoError(t, err)

		_, err = owner.Do(createMutation, map[string]any{"input": input})
		require.NoError(t, err, "second mapping creation should be idempotent")

		const countQuery = `
			query($id: ID!) {
				node(id: $id) {
					... on InternalControl {
						thirdParties(first: 10) { totalCount }
					}
				}
			}
		`

		var result struct {
			Node struct {
				ThirdParties struct {
					TotalCount int `json:"totalCount"`
				} `json:"thirdParties"`
			} `json:"node"`
		}

		err = owner.Execute(countQuery, map[string]any{"id": internalControlID}, &result)
		require.NoError(t, err)
		assert.Equal(t, 1, result.Node.ThirdParties.TotalCount, "duplicate create must not produce a second row")
	})

	t.Run("delete mapping removes link from both sides", func(t *testing.T) {
		t.Parallel()

		internalControlID := factory.NewInternalControl(owner).WithName("Unlink Internal control").Create()
		thirdPartyID := factory.NewThirdParty(owner).WithName("Unlink Third Party").Create()

		input := map[string]any{
			"internalControlId": internalControlID,
			"thirdPartyId":      thirdPartyID,
		}

		_, err := owner.Do(createMutation, map[string]any{"input": input})
		require.NoError(t, err)

		_, err = owner.Do(deleteMutation, map[string]any{"input": input})
		require.NoError(t, err)

		const internalControlCountQuery = `
			query($id: ID!) {
				node(id: $id) {
					... on InternalControl {
						thirdParties(first: 10) { totalCount }
					}
				}
			}
		`

		var internalControlResult struct {
			Node struct {
				ThirdParties struct {
					TotalCount int `json:"totalCount"`
				} `json:"thirdParties"`
			} `json:"node"`
		}

		err = owner.Execute(internalControlCountQuery, map[string]any{"id": internalControlID}, &internalControlResult)
		require.NoError(t, err)
		assert.Equal(t, 0, internalControlResult.Node.ThirdParties.TotalCount, "internal control side should have no linked third parties")

		const thirdPartyCountQuery = `
			query($id: ID!) {
				node(id: $id) {
					... on ThirdParty {
						internalControls(first: 10) { totalCount }
					}
				}
			}
		`

		var thirdPartyResult struct {
			Node struct {
				InternalControls struct {
					TotalCount int `json:"totalCount"`
				} `json:"internalControls"`
			} `json:"node"`
		}

		err = owner.Execute(thirdPartyCountQuery, map[string]any{"id": thirdPartyID}, &thirdPartyResult)
		require.NoError(t, err)
		assert.Equal(t, 0, thirdPartyResult.Node.InternalControls.TotalCount, "third-party side should have no linked internal controls")
	})

	t.Run("viewer cannot create mapping", func(t *testing.T) {
		t.Parallel()

		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		internalControlID := factory.NewInternalControl(owner).WithName("RBAC Internal control").Create()
		thirdPartyID := factory.NewThirdParty(owner).WithName("RBAC Third Party").Create()

		_, err := viewer.Do(createMutation, map[string]any{
			"input": map[string]any{
				"internalControlId": internalControlID,
				"thirdPartyId":      thirdPartyID,
			},
		})
		require.Error(t, err)
	})

	t.Run("tenant isolation on mapping creation", func(t *testing.T) {
		t.Parallel()

		otherOwner := testutil.NewClient(t, testutil.RoleOwner)

		internalControlID := factory.NewInternalControl(owner).WithName("Tenant Internal control").Create()
		otherThirdPartyID := factory.NewThirdParty(otherOwner).WithName("Other Tenant Third Party").Create()

		_, err := owner.Do(createMutation, map[string]any{
			"input": map[string]any{
				"internalControlId": internalControlID,
				"thirdPartyId":      otherThirdPartyID,
			},
		})
		require.Error(t, err, "should not link a third party from another tenant")
	})
}
