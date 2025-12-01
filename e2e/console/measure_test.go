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
	"go.gearno.de/x/ref"
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
				"name":        "Owner Measure",
				"description": "Created by owner",
				"category":    "POLICY",
			},
			assertField: "name",
			assertValue: "Owner Measure",
		},
		{
			name: "with POLICY category",
			input: map[string]any{
				"name":     "Policy measure",
				"category": "POLICY",
			},
			assertField: "category",
			assertValue: "POLICY",
		},
		{
			name: "with PROCEDURE category",
			input: map[string]any{
				"name":     "Procedure measure",
				"category": "PROCEDURE",
			},
			assertField: "category",
			assertValue: "PROCEDURE",
		},
		{
			name: "with TECHNICAL category",
			input: map[string]any{
				"name":     "Technical measure",
				"category": "TECHNICAL",
			},
			assertField: "category",
			assertValue: "TECHNICAL",
		},
		{
			name: "with EVIDENCE category",
			input: map[string]any{
				"name":     "Evidence measure",
				"category": "EVIDENCE",
			},
			assertField: "category",
			assertValue: "EVIDENCE",
		},
		{
			name: "with TRAINING category",
			input: map[string]any{
				"name":     "Training measure",
				"category": "TRAINING",
			},
			assertField: "category",
			assertValue: "TRAINING",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation CreateMeasure($input: CreateMeasureInput!) {
					createMeasure(input: $input) {
						measureEdge {
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
			for k, v := range tt.input {
				input[k] = v
			}

			var result struct {
				CreateMeasure struct {
					MeasureEdge struct {
						Node struct {
							ID       string `json:"id"`
							Name     string `json:"name"`
							Category string `json:"category"`
						} `json:"node"`
					} `json:"measureEdge"`
				} `json:"createMeasure"`
			}

			err := owner.Execute(query, map[string]any{"input": input}, &result)
			require.NoError(t, err)

			node := result.CreateMeasure.MeasureEdge.Node
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
				"name": "Test Measure",
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
				"name":     "Test Measure",
				"category": "",
			},
			wantErrorContains: "category",
		},
		{
			name: "missing organizationId",
			input: map[string]any{
				"name":     "Test Measure",
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
			name: "name with angle brackets",
			input: map[string]any{
				"name":     "Test < Measure",
				"category": "POLICY",
			},
			wantErrorContains: "angle brackets",
		},
		{
			name: "description with HTML tags",
			input: map[string]any{
				"name":        "Test Measure",
				"category":    "POLICY",
				"description": "<div>HTML content</div>",
			},
			wantErrorContains: "HTML",
		},
		{
			name: "category with HTML tags",
			input: map[string]any{
				"name":     "Test Measure",
				"category": "<b>POLICY</b>",
			},
			wantErrorContains: "HTML",
		},
		// Newline validation (name should not allow newlines)
		{
			name: "name with newline",
			input: map[string]any{
				"name":     "Test\nMeasure",
				"category": "POLICY",
			},
			wantErrorContains: "newline",
		},
		{
			name: "name with carriage return",
			input: map[string]any{
				"name":     "Test\rMeasure",
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
				"name":     "Test\tMeasure",
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
				mutation CreateMeasure($input: CreateMeasureInput!) {
					createMeasure(input: $input) {
						measureEdge {
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
				return factory.NewMeasure(owner).
					WithName("Measure to Update").
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
				return factory.NewMeasure(owner).WithName("State Test").Create()
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
				return factory.NewMeasure(owner).WithName("State Test").Create()
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
				return factory.NewMeasure(owner).WithName("State Test").Create()
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
			measureID := tt.setup()

			query := `
				mutation UpdateMeasure($input: UpdateMeasureInput!) {
					updateMeasure(input: $input) {
						measure {
							id
							name
							state
						}
					}
				}
			`

			var result struct {
				UpdateMeasure struct {
					Measure struct {
						ID    string `json:"id"`
						Name  string `json:"name"`
						State string `json:"state"`
					} `json:"measure"`
				} `json:"updateMeasure"`
			}

			err := owner.Execute(query, map[string]any{"input": tt.input(measureID)}, &result)
			require.NoError(t, err)

			measure := result.UpdateMeasure.Measure
			switch tt.assertField {
			case "name":
				assert.Equal(t, tt.assertValue, measure.Name)
			case "state":
				assert.Equal(t, tt.assertValue, measure.State)
			}
		})
	}
}

func TestMeasure_Update_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a measure to use for most validation tests
	baseMeasureID := factory.NewMeasure(owner).WithName("Validation Test Measure").Create()

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
			setup: func() string { return baseMeasureID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": ""}
			},
			wantErrorContains: "name",
		},
		{
			name:  "empty category",
			setup: func() string { return baseMeasureID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "category": ""}
			},
			wantErrorContains: "category",
		},
		// HTML injection validation
		{
			name:  "name with HTML tags",
			setup: func() string { return baseMeasureID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "<script>alert('xss')</script>"}
			},
			wantErrorContains: "HTML",
		},
		{
			name:  "name with angle brackets",
			setup: func() string { return baseMeasureID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test < Measure"}
			},
			wantErrorContains: "angle brackets",
		},
		{
			name:  "description with HTML tags",
			setup: func() string { return baseMeasureID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "description": "<img src=x onerror=alert(1)>"}
			},
			wantErrorContains: "HTML",
		},
		{
			name:  "category with HTML tags",
			setup: func() string { return baseMeasureID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "category": "<b>POLICY</b>"}
			},
			wantErrorContains: "HTML",
		},
		// Newline validation (name should not allow newlines)
		{
			name:  "name with newline",
			setup: func() string { return baseMeasureID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\nMeasure"}
			},
			wantErrorContains: "newline",
		},
		{
			name:  "name with carriage return",
			setup: func() string { return baseMeasureID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\rMeasure"}
			},
			wantErrorContains: "carriage return",
		},
		// Control character validation
		{
			name:  "name with null byte",
			setup: func() string { return baseMeasureID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\x00Measure"}
			},
			wantErrorContains: "control character",
		},
		{
			name:  "name with tab character",
			setup: func() string { return baseMeasureID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\tMeasure"}
			},
			wantErrorContains: "control character",
		},
		{
			name:  "description with null byte",
			setup: func() string { return baseMeasureID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "description": "Test\x00Description"}
			},
			wantErrorContains: "control character",
		},
		// Zero-width character validation
		{
			name:  "name with zero-width space",
			setup: func() string { return baseMeasureID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\u200BMeasure"}
			},
			wantErrorContains: "zero-width",
		},
		{
			name:  "description with zero-width joiner",
			setup: func() string { return baseMeasureID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "description": "Test\u200DDescription"}
			},
			wantErrorContains: "zero-width",
		},
		// Bidirectional override validation
		{
			name:  "name with right-to-left override",
			setup: func() string { return baseMeasureID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "name": "Test\u202EMeasure"}
			},
			wantErrorContains: "bidirectional",
		},
		{
			name:  "description with left-to-right override",
			setup: func() string { return baseMeasureID },
			input: func(id string) map[string]any {
				return map[string]any{"id": id, "description": "Test\u202DDescription"}
			},
			wantErrorContains: "bidirectional",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			measureID := tt.setup()

			query := `
				mutation UpdateMeasure($input: UpdateMeasureInput!) {
					updateMeasure(input: $input) {
						measure {
							id
						}
					}
				}
			`

			_, err := owner.Do(query, map[string]any{"input": tt.input(measureID)})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestMeasure_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("delete existing measure", func(t *testing.T) {
		measureID := factory.NewMeasure(owner).WithName("Measure to Delete").Create()

		query := `
			mutation DeleteMeasure($input: DeleteMeasureInput!) {
				deleteMeasure(input: $input) {
					deletedMeasureId
				}
			}
		`

		var result struct {
			DeleteMeasure struct {
				DeletedMeasureID string `json:"deletedMeasureId"`
			} `json:"deleteMeasure"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{"measureId": measureID},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, measureID, result.DeleteMeasure.DeletedMeasureID)
	})
}

func TestMeasure_Delete_Validation(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	tests := []struct {
		name              string
		measureID         string
		wantErrorContains string
	}{
		{
			name:              "invalid ID format",
			measureID:         "invalid-id-format",
			wantErrorContains: "base64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation DeleteMeasure($input: DeleteMeasureInput!) {
					deleteMeasure(input: $input) {
						deletedMeasureId
					}
				}
			`

			_, err := owner.Do(query, map[string]any{
				"input": map[string]any{"measureId": tt.measureID},
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrorContains)
		})
	}
}

func TestMeasure_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	measureNames := []string{"Measure A", "Measure B", "Measure C"}
	for _, name := range measureNames {
		factory.NewMeasure(owner).WithName(name).Create()
	}

	query := `
		query GetMeasures($id: ID!) {
			node(id: $id) {
				... on Organization {
					measures(first: 10) {
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
			Measures struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"measures"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.Measures.TotalCount, 3)
}

func TestMeasure_Query(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("query with non-existent ID returns error", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Measure {
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
			mutation CreateMeasure($input: CreateMeasureInput!) {
				createMeasure(input: $input) {
					measureEdge {
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
			CreateMeasure struct {
				MeasureEdge struct {
					Node struct {
						ID        string    `json:"id"`
						CreatedAt time.Time `json:"createdAt"`
						UpdatedAt time.Time `json:"updatedAt"`
					} `json:"node"`
				} `json:"measureEdge"`
			} `json:"createMeasure"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           "Timestamp Test Measure",
				"category":       "POLICY",
			},
		}, &result)
		require.NoError(t, err)

		node := result.CreateMeasure.MeasureEdge.Node
		testutil.AssertTimestampsOnCreate(t, node.CreatedAt, node.UpdatedAt, beforeCreate)
	})

	t.Run("updatedAt changes on update", func(t *testing.T) {
		measureID := factory.NewMeasure(owner).WithName("Timestamp Update Test").Create()

		// Get initial timestamps
		getQuery := `
			query($id: ID!) {
				node(id: $id) {
					... on Measure {
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

		err := owner.Execute(getQuery, map[string]any{"id": measureID}, &getResult)
		require.NoError(t, err)

		initialCreatedAt := getResult.Node.CreatedAt
		initialUpdatedAt := getResult.Node.UpdatedAt

		// Wait a bit to ensure timestamp difference
		time.Sleep(10 * time.Millisecond)

		updateQuery := `
			mutation UpdateMeasure($input: UpdateMeasureInput!) {
				updateMeasure(input: $input) {
					measure {
						createdAt
						updatedAt
					}
				}
			}
		`

		var updateResult struct {
			UpdateMeasure struct {
				Measure struct {
					CreatedAt time.Time `json:"createdAt"`
					UpdatedAt time.Time `json:"updatedAt"`
				} `json:"measure"`
			} `json:"updateMeasure"`
		}

		err = owner.Execute(updateQuery, map[string]any{
			"input": map[string]any{
				"id":   measureID,
				"name": "Updated Timestamp Test",
			},
		}, &updateResult)
		require.NoError(t, err)

		measure := updateResult.UpdateMeasure.Measure
		testutil.AssertTimestampsOnUpdate(t, measure.CreatedAt, measure.UpdatedAt, initialCreatedAt, initialUpdatedAt)
	})
}

func TestMeasure_OmittableDescription(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	measureID := factory.NewMeasure(owner).
		WithName("Omittable Test Measure").
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
			wantDescription: ref.Ref("Updated description"),
		},
		{
			name:            "update with null description clears it",
			input:           map[string]any{"description": nil},
			wantDescription: nil,
		},
		{
			name:            "set description again",
			input:           map[string]any{"description": "Should persist"},
			wantDescription: ref.Ref("Should persist"),
		},
		{
			name:            "update without description preserves it",
			input:           map[string]any{"name": "Updated Name"},
			wantDescription: ref.Ref("Should persist"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation($input: UpdateMeasureInput!) {
					updateMeasure(input: $input) {
						measure {
							id
							name
							description
						}
					}
				}
			`

			input := map[string]any{"id": measureID}
			for k, v := range tt.input {
				input[k] = v
			}

			var result struct {
				UpdateMeasure struct {
					Measure struct {
						ID          string  `json:"id"`
						Name        string  `json:"name"`
						Description *string `json:"description"`
					} `json:"measure"`
				} `json:"updateMeasure"`
			}

			err := owner.Execute(query, map[string]any{"input": input}, &result)
			require.NoError(t, err)

			testutil.AssertOptionalStringEqual(t, tt.wantDescription, result.UpdateMeasure.Measure.Description, "description")
		})
	}
}

func TestMeasure_SubResolvers(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	measureID := factory.NewMeasure(owner).
		WithName("SubResolver Test Measure").
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
						... on Measure {
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

			resp, err := owner.Do(query, map[string]any{"id": measureID})
			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}

func TestMeasure_RBAC(t *testing.T) {
	t.Parallel()

	t.Run("create", func(t *testing.T) {
		t.Run("owner can create", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)

			_, err := owner.Do(`
				mutation CreateMeasure($input: CreateMeasureInput!) {
					createMeasure(input: $input) {
						measureEdge { node { id } }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"organizationId": owner.GetOrganizationID().String(),
					"name":           "RBAC Test Measure",
					"category":       "POLICY",
				},
			})
			require.NoError(t, err, "owner should be able to create measure")
		})

		t.Run("admin can create", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)

			_, err := admin.Do(`
				mutation CreateMeasure($input: CreateMeasureInput!) {
					createMeasure(input: $input) {
						measureEdge { node { id } }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"organizationId": admin.GetOrganizationID().String(),
					"name":           "RBAC Test Measure",
					"category":       "POLICY",
				},
			})
			require.NoError(t, err, "admin should be able to create measure")
		})

		t.Run("viewer cannot create", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

			_, err := viewer.Do(`
				mutation CreateMeasure($input: CreateMeasureInput!) {
					createMeasure(input: $input) {
						measureEdge { node { id } }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"organizationId": viewer.GetOrganizationID().String(),
					"name":           "RBAC Test Measure",
					"category":       "POLICY",
				},
			})
			testutil.RequireForbiddenError(t, err, "viewer should not be able to create measure")
		})
	})

	t.Run("update", func(t *testing.T) {
		t.Run("owner can update", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			measureID := factory.NewMeasure(owner).WithName("RBAC Update Test").Create()

			_, err := owner.Do(`
				mutation UpdateMeasure($input: UpdateMeasureInput!) {
					updateMeasure(input: $input) {
						measure { id }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"id":   measureID,
					"name": "Updated by Owner",
				},
			})
			require.NoError(t, err, "owner should be able to update measure")
		})

		t.Run("admin can update", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
			measureID := factory.NewMeasure(owner).WithName("RBAC Update Test").Create()

			_, err := admin.Do(`
				mutation UpdateMeasure($input: UpdateMeasureInput!) {
					updateMeasure(input: $input) {
						measure { id }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"id":   measureID,
					"name": "Updated by Admin",
				},
			})
			require.NoError(t, err, "admin should be able to update measure")
		})

		t.Run("viewer cannot update", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			measureID := factory.NewMeasure(owner).WithName("RBAC Update Test").Create()

			_, err := viewer.Do(`
				mutation UpdateMeasure($input: UpdateMeasureInput!) {
					updateMeasure(input: $input) {
						measure { id }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"id":   measureID,
					"name": "Updated by Viewer",
				},
			})
			testutil.RequireForbiddenError(t, err, "viewer should not be able to update measure")
		})
	})

	t.Run("delete", func(t *testing.T) {
		t.Run("owner can delete", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			measureID := factory.NewMeasure(owner).WithName("RBAC Delete Test").Create()

			_, err := owner.Do(`
				mutation DeleteMeasure($input: DeleteMeasureInput!) {
					deleteMeasure(input: $input) {
						deletedMeasureId
					}
				}
			`, map[string]any{
				"input": map[string]any{"measureId": measureID},
			})
			require.NoError(t, err, "owner should be able to delete measure")
		})

		t.Run("admin can delete", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
			measureID := factory.NewMeasure(owner).WithName("RBAC Delete Test").Create()

			_, err := admin.Do(`
				mutation DeleteMeasure($input: DeleteMeasureInput!) {
					deleteMeasure(input: $input) {
						deletedMeasureId
					}
				}
			`, map[string]any{
				"input": map[string]any{"measureId": measureID},
			})
			require.NoError(t, err, "admin should be able to delete measure")
		})

		t.Run("viewer cannot delete", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			measureID := factory.NewMeasure(owner).WithName("RBAC Delete Test").Create()

			_, err := viewer.Do(`
				mutation DeleteMeasure($input: DeleteMeasureInput!) {
					deleteMeasure(input: $input) {
						deletedMeasureId
					}
				}
			`, map[string]any{
				"input": map[string]any{"measureId": measureID},
			})
			testutil.RequireForbiddenError(t, err, "viewer should not be able to delete measure")
		})
	})

	t.Run("read", func(t *testing.T) {
		t.Run("owner can read", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			measureID := factory.NewMeasure(owner).WithName("RBAC Read Test").Create()

			var result struct {
				Node *struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			}

			err := owner.Execute(`
				query($id: ID!) {
					node(id: $id) {
						... on Measure { id name }
					}
				}
			`, map[string]any{"id": measureID}, &result)
			require.NoError(t, err, "owner should be able to read measure")
			require.NotNil(t, result.Node, "owner should receive measure data")
		})

		t.Run("admin can read", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
			measureID := factory.NewMeasure(owner).WithName("RBAC Read Test").Create()

			var result struct {
				Node *struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			}

			err := admin.Execute(`
				query($id: ID!) {
					node(id: $id) {
						... on Measure { id name }
					}
				}
			`, map[string]any{"id": measureID}, &result)
			require.NoError(t, err, "admin should be able to read measure")
			require.NotNil(t, result.Node, "admin should receive measure data")
		})

		t.Run("viewer can read", func(t *testing.T) {
			owner := testutil.NewClient(t, testutil.RoleOwner)
			viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
			measureID := factory.NewMeasure(owner).WithName("RBAC Read Test").Create()

			var result struct {
				Node *struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			}

			err := viewer.Execute(`
				query($id: ID!) {
					node(id: $id) {
						... on Measure { id name }
					}
				}
			`, map[string]any{"id": measureID}, &result)
			require.NoError(t, err, "viewer should be able to read measure")
			require.NotNil(t, result.Node, "viewer should receive measure data")
		})
	})
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
					"name":     "Test Measure",
					"category": longCategory,
				},
				wantErrorContains: "category",
			},
			{
				name: "description exceeds max length",
				input: map[string]any{
					"name":        "Test Measure",
					"category":    "POLICY",
					"description": longDescription,
				},
				wantErrorContains: "description",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				query := `
					mutation CreateMeasure($input: CreateMeasureInput!) {
						createMeasure(input: $input) {
							measureEdge {
								node { id }
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

				_, err := owner.Do(query, map[string]any{"input": input})
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrorContains)
			})
		}
	})

	t.Run("update", func(t *testing.T) {
		measureID := factory.NewMeasure(owner).WithName("Max Length Test").Create()

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
					mutation UpdateMeasure($input: UpdateMeasureInput!) {
						updateMeasure(input: $input) {
							measure { id }
						}
					}
				`

				input := map[string]any{"id": measureID}
				for k, v := range tt.input {
					input[k] = v
				}

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
		measureID := factory.NewMeasure(owner).WithName("Measure with Tasks").Create()

		// Create tasks linked to the measure
		task1ID := factory.NewTask(owner, measureID).WithName("Task 1").Create()
		task2ID := factory.NewTask(owner, measureID).WithName("Task 2").Create()

		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Measure {
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

		err := owner.Execute(query, map[string]any{"id": measureID}, &result)
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
		measureID := factory.NewMeasure(owner).WithName("Measure with Controls").Create()

		// Create framework and control
		frameworkID := factory.NewFramework(owner).WithName("Test Framework").Create()
		controlID := factory.NewControl(owner, frameworkID).WithName("Test Control").Create()

		// Link control to measure
		linkQuery := `
			mutation($input: CreateControlMeasureMappingInput!) {
				createControlMeasureMapping(input: $input) {
					controlEdge { node { id } }
				}
			}
		`
		_, err := owner.Do(linkQuery, map[string]any{
			"input": map[string]any{
				"controlId": controlID,
				"measureId": measureID,
			},
		})
		require.NoError(t, err)

		// Query the measure's controls
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Measure {
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

		err = owner.Execute(query, map[string]any{"id": measureID}, &result)
		require.NoError(t, err)
		assert.Equal(t, 1, result.Node.Controls.TotalCount)
		assert.Equal(t, controlID, result.Node.Controls.Edges[0].Node.ID)
	})

	t.Run("risks sub-resolver with linked risks", func(t *testing.T) {
		measureID := factory.NewMeasure(owner).WithName("Measure with Risks").Create()

		// Create risk
		riskID := factory.NewRisk(owner).WithName("Test Risk").Create()

		// Link risk to measure
		linkQuery := `
			mutation($input: CreateRiskMeasureMappingInput!) {
				createRiskMeasureMapping(input: $input) {
					riskEdge { node { id } }
				}
			}
		`
		_, err := owner.Do(linkQuery, map[string]any{
			"input": map[string]any{
				"riskId":    riskID,
				"measureId": measureID,
			},
		})
		require.NoError(t, err)

		// Query the measure's risks
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Measure {
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

		err = owner.Execute(query, map[string]any{"id": measureID}, &result)
		require.NoError(t, err)
		assert.Equal(t, 1, result.Node.Risks.TotalCount)
		assert.Equal(t, riskID, result.Node.Risks.Edges[0].Node.ID)
	})
}

func TestMeasure_Pagination(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create multiple measures for pagination testing
	measureIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		measureIDs[i] = factory.NewMeasure(owner).
			WithName(fmt.Sprintf("Pagination Measure %d", i)).
			Create()
	}

	t.Run("first/after pagination", func(t *testing.T) {
		// Get first 2 measures
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						measures(first: 2) {
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
				Measures struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
						Cursor string `json:"cursor"`
					} `json:"edges"`
					PageInfo   testutil.PageInfo `json:"pageInfo"`
					TotalCount int               `json:"totalCount"`
				} `json:"measures"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id": owner.GetOrganizationID().String(),
		}, &result)
		require.NoError(t, err)

		testutil.AssertFirstPage(t, len(result.Node.Measures.Edges), result.Node.Measures.PageInfo, 2, true)
		assert.GreaterOrEqual(t, result.Node.Measures.TotalCount, 5)

		// Get next page using cursor
		testutil.AssertHasMorePages(t, result.Node.Measures.PageInfo)
		queryAfter := `
			query($id: ID!, $after: CursorKey) {
				node(id: $id) {
					... on Organization {
						measures(first: 2, after: $after) {
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
				Measures struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo testutil.PageInfo `json:"pageInfo"`
				} `json:"measures"`
			} `json:"node"`
		}

		err = owner.Execute(queryAfter, map[string]any{
			"id":    owner.GetOrganizationID().String(),
			"after": *result.Node.Measures.PageInfo.EndCursor,
		}, &resultAfter)
		require.NoError(t, err)

		testutil.AssertMiddlePage(t, len(resultAfter.Node.Measures.Edges), resultAfter.Node.Measures.PageInfo, 2)
	})

	t.Run("last/before pagination", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						measures(last: 2) {
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
				Measures struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo testutil.PageInfo `json:"pageInfo"`
				} `json:"measures"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id": owner.GetOrganizationID().String(),
		}, &result)
		require.NoError(t, err)

		testutil.AssertLastPage(t, len(result.Node.Measures.Edges), result.Node.Measures.PageInfo, 2, true)
	})
}

func TestMeasure_Filtering(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create measures with different states
	measure1ID := factory.NewMeasure(owner).WithName("Filter Test Implemented").Create()
	measure2ID := factory.NewMeasure(owner).WithName("Filter Test Not Started").Create()

	// Update measure1 to IMPLEMENTED state
	updateQuery := `
		mutation($input: UpdateMeasureInput!) {
			updateMeasure(input: $input) {
				measure { id state }
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
			query($id: ID!, $filter: MeasureFilter) {
				node(id: $id) {
					... on Organization {
						measures(first: 100, filter: $filter) {
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
				Measures struct {
					Edges []struct {
						Node struct {
							ID    string `json:"id"`
							Name  string `json:"name"`
							State string `json:"state"`
						} `json:"node"`
					} `json:"edges"`
					TotalCount int `json:"totalCount"`
				} `json:"measures"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id":     owner.GetOrganizationID().String(),
			"filter": map[string]any{"state": "IMPLEMENTED"},
		}, &result)
		require.NoError(t, err)

		// All returned measures should be IMPLEMENTED
		for _, edge := range result.Node.Measures.Edges {
			assert.Equal(t, "IMPLEMENTED", edge.Node.State)
		}

		// Should contain our implemented measure
		found := false
		for _, edge := range result.Node.Measures.Edges {
			if edge.Node.ID == measure1ID {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected to find implemented measure in filtered results")
	})

	t.Run("filter by query string", func(t *testing.T) {
		query := `
			query($id: ID!, $filter: MeasureFilter) {
				node(id: $id) {
					... on Organization {
						measures(first: 100, filter: $filter) {
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
				Measures struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"measures"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{
			"id":     owner.GetOrganizationID().String(),
			"filter": map[string]any{"query": "Filter Test"},
		}, &result)
		require.NoError(t, err)

		// Should find measures matching the query
		foundIDs := make([]string, len(result.Node.Measures.Edges))
		for i, edge := range result.Node.Measures.Edges {
			foundIDs[i] = edge.Node.ID
		}
		assert.Contains(t, foundIDs, measure1ID)
		assert.Contains(t, foundIDs, measure2ID)
	})
}

func TestMeasure_TenantIsolation(t *testing.T) {
	t.Parallel()

	// Create two separate organizations with their own owners
	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a measure in org1
	measureID := factory.NewMeasure(org1Owner).WithName("Org1 Measure").Create()

	t.Run("cannot read measure from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Measure {
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

		err := org2Owner.Execute(query, map[string]any{"id": measureID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "measure")
	})

	t.Run("cannot update measure from another organization", func(t *testing.T) {
		query := `
			mutation UpdateMeasure($input: UpdateMeasureInput!) {
				updateMeasure(input: $input) {
					measure { id }
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":   measureID,
				"name": "Hijacked Measure",
			},
		})
		require.Error(t, err, "Should not be able to update measure from another org")
	})

	t.Run("cannot delete measure from another organization", func(t *testing.T) {
		query := `
			mutation DeleteMeasure($input: DeleteMeasureInput!) {
				deleteMeasure(input: $input) {
					deletedMeasureId
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"measureId": measureID,
			},
		})
		require.Error(t, err, "Should not be able to delete measure from another org")
	})

	t.Run("cannot list measures from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Organization {
						measures(first: 100) {
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
				Measures struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"measures"`
			} `json:"node"`
		}

		// Query org1's measures as org2
		err := org2Owner.Execute(query, map[string]any{
			"id": org1Owner.GetOrganizationID().String(),
		}, &result)

		// Should either error or return empty list (can't access other org's data)
		if err == nil {
			// If no error, the measure from org1 should not be in the list
			for _, edge := range result.Node.Measures.Edges {
				assert.NotEqual(t, measureID, edge.Node.ID, "Should not see measure from another org")
			}
		}
	})
}

func TestMeasure_Ordering(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create measures with distinct names for ordering
	factory.NewMeasure(owner).WithName("AAA Order Test").Create()
	factory.NewMeasure(owner).WithName("ZZZ Order Test").Create()

	t.Run("order by name ascending", func(t *testing.T) {
		query := `
			query($id: ID!, $orderBy: MeasureOrder) {
				node(id: $id) {
					... on Organization {
						measures(first: 100, orderBy: $orderBy) {
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
				Measures struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"measures"`
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
		names := make([]string, len(result.Node.Measures.Edges))
		for i, edge := range result.Node.Measures.Edges {
			names[i] = edge.Node.Name
		}
		testutil.AssertOrderedAscending(t, names, "name")
	})

	t.Run("order by name descending", func(t *testing.T) {
		query := `
			query($id: ID!, $orderBy: MeasureOrder) {
				node(id: $id) {
					... on Organization {
						measures(first: 100, orderBy: $orderBy) {
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
				Measures struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"measures"`
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
		names := make([]string, len(result.Node.Measures.Edges))
		for i, edge := range result.Node.Measures.Edges {
			names[i] = edge.Node.Name
		}
		testutil.AssertOrderedDescending(t, names, "name")
	})

	t.Run("order by created_at", func(t *testing.T) {
		query := `
			query($id: ID!, $orderBy: MeasureOrder) {
				node(id: $id) {
					... on Organization {
						measures(first: 100, orderBy: $orderBy) {
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
				Measures struct {
					Edges []struct {
						Node struct {
							ID        string    `json:"id"`
							CreatedAt time.Time `json:"createdAt"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"measures"`
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
		times := make([]time.Time, len(result.Node.Measures.Edges))
		for i, edge := range result.Node.Measures.Edges {
			times[i] = edge.Node.CreatedAt
		}
		testutil.AssertTimesOrderedDescending(t, times, "createdAt")
	})
}

