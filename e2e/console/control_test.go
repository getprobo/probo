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

func TestControl_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	frameworkID := factory.CreateFramework(owner, factory.Attrs{"name": "Framework for Control Tests"})

	t.Run("with full details", func(t *testing.T) {
		query := `
			mutation CreateControl($input: CreateControlInput!) {
				createControl(input: $input) {
					controlEdge {
						node {
							id
							name
							sectionTitle
							status
						}
					}
				}
			}
		`

		var result struct {
			CreateControl struct {
				ControlEdge struct {
					Node struct {
						ID           string `json:"id"`
						Name         string `json:"name"`
						SectionTitle string `json:"sectionTitle"`
						Status       string `json:"status"`
					} `json:"node"`
				} `json:"controlEdge"`
			} `json:"createControl"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"frameworkId":  frameworkID,
				"sectionTitle": "A.5",
				"name":         "Information Security Policies",
				"description":  "Policies for information security",
				"status":       "INCLUDED",
			},
		}, &result)
		require.NoError(t, err)

		control := result.CreateControl.ControlEdge.Node
		assert.NotEmpty(t, control.ID)
		assert.Equal(t, "Information Security Policies", control.Name)
		assert.Equal(t, "A.5", control.SectionTitle)
		assert.Equal(t, "INCLUDED", control.Status)
	})

	t.Run("with excluded status and justification", func(t *testing.T) {
		query := `
			mutation CreateControl($input: CreateControlInput!) {
				createControl(input: $input) {
					controlEdge {
						node {
							id
							status
							exclusionJustification
						}
					}
				}
			}
		`

		var result struct {
			CreateControl struct {
				ControlEdge struct {
					Node struct {
						ID                     string  `json:"id"`
						Status                 string  `json:"status"`
						ExclusionJustification *string `json:"exclusionJustification"`
					} `json:"node"`
				} `json:"controlEdge"`
			} `json:"createControl"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"frameworkId":            frameworkID,
				"sectionTitle":           "A.10",
				"name":                   "Cryptography",
				"description":            "Cryptography controls",
				"status":                 "EXCLUDED",
				"exclusionJustification": "Not applicable - no cryptographic data processing",
			},
		}, &result)
		require.NoError(t, err)

		control := result.CreateControl.ControlEdge.Node
		assert.Equal(t, "EXCLUDED", control.Status)
		assert.Equal(t, "Not applicable - no cryptographic data processing", *control.ExclusionJustification)
	})
}

func TestControl_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	frameworkID := factory.CreateFramework(owner, factory.Attrs{"name": "Framework for Control Update"})
	controlID := factory.CreateControl(owner, frameworkID, factory.Attrs{
		"name":        "Control to Update",
		"description": "Original description",
	})

	t.Run("updates name and description", func(t *testing.T) {
		query := `
			mutation UpdateControl($input: UpdateControlInput!) {
				updateControl(input: $input) {
					control {
						id
						name
					}
				}
			}
		`

		var result struct {
			UpdateControl struct {
				Control struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"control"`
			} `json:"updateControl"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":          controlID,
				"name":        "Updated Control Name",
				"description": "Updated description",
			},
		}, &result)
		require.NoError(t, err)

		assert.Equal(t, controlID, result.UpdateControl.Control.ID)
		assert.Equal(t, "Updated Control Name", result.UpdateControl.Control.Name)
	})

	t.Run("changes status from INCLUDED to EXCLUDED", func(t *testing.T) {
		statusTestControlID := factory.CreateControl(owner, frameworkID, factory.Attrs{
			"name":   "Status Change Control",
			"status": "INCLUDED",
		})

		query := `
			mutation UpdateControl($input: UpdateControlInput!) {
				updateControl(input: $input) {
					control {
						id
						status
						exclusionJustification
					}
				}
			}
		`

		var result struct {
			UpdateControl struct {
				Control struct {
					ID                     string  `json:"id"`
					Status                 string  `json:"status"`
					ExclusionJustification *string `json:"exclusionJustification"`
				} `json:"control"`
			} `json:"updateControl"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":                     statusTestControlID,
				"status":                 "EXCLUDED",
				"exclusionJustification": "No physical assets in scope",
			},
		}, &result)
		require.NoError(t, err)

		assert.Equal(t, "EXCLUDED", result.UpdateControl.Control.Status)
		assert.Equal(t, "No physical assets in scope", *result.UpdateControl.Control.ExclusionJustification)
	})
}

func TestControl_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	frameworkID := factory.CreateFramework(owner, factory.Attrs{
		"name": "Framework for Delete",
	})
	controlID := factory.CreateControl(owner, frameworkID, factory.Attrs{
		"name": "Control to Delete",
	})

	query := `
		mutation DeleteControl($input: DeleteControlInput!) {
			deleteControl(input: $input) {
				deletedControlId
			}
		}
	`

	var result struct {
		DeleteControl struct {
			DeletedControlID string `json:"deletedControlId"`
		} `json:"deleteControl"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"controlId": controlID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, controlID, result.DeleteControl.DeletedControlID)
}

func TestControl_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	frameworkID := factory.CreateFramework(owner, factory.Attrs{"name": "Framework for List Test"})

	// Create multiple controls
	controlNames := []string{"Control A", "Control B", "Control C"}
	for i, name := range controlNames {
		factory.CreateControl(owner, frameworkID, factory.Attrs{
			"name":         name,
			"sectionTitle": fmt.Sprintf("A.%d", 5+i),
		})
	}

	query := `
		query GetFrameworkControls($id: ID!) {
			node(id: $id) {
				... on Framework {
					id
					name
					controls(first: 10) {
						edges {
							node {
								id
								name
								sectionTitle
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Controls struct {
				Edges []struct {
					Node struct {
						ID           string `json:"id"`
						Name         string `json:"name"`
						SectionTitle string `json:"sectionTitle"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"controls"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{"id": frameworkID}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result.Node.Controls.Edges), 3)
}

func TestControl_RequiredFields(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

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

	var frameworkResult struct {
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
			"name":           fmt.Sprintf("Control Required Fields Test %d", time.Now().UnixNano()),
		},
	}, &frameworkResult)
	require.NoError(t, err)
	frameworkID := frameworkResult.CreateFramework.FrameworkEdge.Node.ID

	createControlQuery := `
		mutation CreateControl($input: CreateControlInput!) {
			createControl(input: $input) {
				controlEdge {
					node {
						id
					}
				}
			}
		}
	`

	tests := []struct {
		name      string
		variables map[string]any
		wantError bool
	}{
		{
			name: "Missing frameworkId should fail",
			variables: map[string]any{
				"input": map[string]any{
					"name":         "Test Control",
					"description":  "Test",
					"sectionTitle": "Section 1",
					"status":       "INCLUDED",
				},
			},
			wantError: true,
		},
		{
			name: "Missing name should fail",
			variables: map[string]any{
				"input": map[string]any{
					"frameworkId":  frameworkID,
					"description":  "Test",
					"sectionTitle": "Section 1",
					"status":       "INCLUDED",
				},
			},
			wantError: true,
		},
		{
			name: "Missing sectionTitle should fail",
			variables: map[string]any{
				"input": map[string]any{
					"frameworkId": frameworkID,
					"name":        "Test Control",
					"description": "Test",
					"status":      "INCLUDED",
				},
			},
			wantError: true,
		},
		{
			name: "Missing status should fail",
			variables: map[string]any{
				"input": map[string]any{
					"frameworkId":  frameworkID,
					"name":         "Test Control",
					"description":  "Test",
					"sectionTitle": "Section 1",
				},
			},
			wantError: true,
		},
		{
			name: "Missing description should fail (required field)",
			variables: map[string]any{
				"input": map[string]any{
					"frameworkId":  frameworkID,
					"name":         "Test Control",
					"sectionTitle": "Section 1",
					"status":       "INCLUDED",
				},
			},
			wantError: true,
		},
		{
			name: "Invalid status enum should fail",
			variables: map[string]any{
				"input": map[string]any{
					"frameworkId":  frameworkID,
					"name":         "Test Control",
					"description":  "Test",
					"sectionTitle": "Section 1",
					"status":       "INVALID_STATUS",
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := owner.Do(createControlQuery, tt.variables)

			if tt.wantError {
				require.Error(t, err, "Expected validation error")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestControl_OmittableDescription(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create framework
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

	var frameworkResult struct {
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
			"name":           fmt.Sprintf("Control Omittable Test %d", time.Now().UnixNano()),
		},
	}, &frameworkResult)
	require.NoError(t, err)
	frameworkID := frameworkResult.CreateFramework.FrameworkEdge.Node.ID

	// Create control with description
	createControlQuery := `
		mutation CreateControl($input: CreateControlInput!) {
			createControl(input: $input) {
				controlEdge {
					node {
						id
						description
					}
				}
			}
		}
	`

	var createResult struct {
		CreateControl struct {
			ControlEdge struct {
				Node struct {
					ID          string `json:"id"`
					Description string `json:"description"`
				} `json:"node"`
			} `json:"controlEdge"`
		} `json:"createControl"`
	}

	err = owner.Execute(createControlQuery, map[string]any{
		"input": map[string]any{
			"frameworkId":  frameworkID,
			"name":         "Omittable Test Control",
			"description":  "Initial description",
			"sectionTitle": "Section 1",
			"status":       "INCLUDED",
		},
	}, &createResult)
	require.NoError(t, err)
	controlID := createResult.CreateControl.ControlEdge.Node.ID

	t.Run("Update with null description should clear it", func(t *testing.T) {
		updateControlQuery := `
			mutation UpdateControl($input: UpdateControlInput!) {
				updateControl(input: $input) {
					control {
						id
						description
					}
				}
			}
		`

		var updateResult struct {
			UpdateControl struct {
				Control struct {
					ID          string  `json:"id"`
					Description *string `json:"description"`
				} `json:"control"`
			} `json:"updateControl"`
		}

		err := owner.Execute(updateControlQuery, map[string]any{
			"input": map[string]any{
				"id":          controlID,
				"description": nil,
			},
		}, &updateResult)
		require.NoError(t, err)
		assert.Nil(t, updateResult.UpdateControl.Control.Description)
	})

	t.Run("Update without description should not change it", func(t *testing.T) {
		// Set description first
		setDescQuery := `
			mutation UpdateControl($input: UpdateControlInput!) {
				updateControl(input: $input) {
					control {
						id
					}
				}
			}
		`

		var setDescResult struct {
			UpdateControl struct {
				Control struct {
					ID string `json:"id"`
				} `json:"control"`
			} `json:"updateControl"`
		}

		err := owner.Execute(setDescQuery, map[string]any{
			"input": map[string]any{
				"id":          controlID,
				"description": "Should persist",
			},
		}, &setDescResult)
		require.NoError(t, err)

		// Update only name
		updateNameQuery := `
			mutation UpdateControl($input: UpdateControlInput!) {
				updateControl(input: $input) {
					control {
						id
						name
						description
					}
				}
			}
		`

		var updateResult struct {
			UpdateControl struct {
				Control struct {
					ID          string `json:"id"`
					Name        string `json:"name"`
					Description string `json:"description"`
				} `json:"control"`
			} `json:"updateControl"`
		}

		err = owner.Execute(updateNameQuery, map[string]any{
			"input": map[string]any{
				"id":   controlID,
				"name": "Updated Name",
			},
		}, &updateResult)
		require.NoError(t, err)
		assert.Equal(t, "Should persist", updateResult.UpdateControl.Control.Description)
	})
}

func TestControl_SubResolvers(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create framework
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

	var frameworkResult struct {
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
			"name":           fmt.Sprintf("Control SubResolver Test %d", time.Now().UnixNano()),
		},
	}, &frameworkResult)
	require.NoError(t, err)
	frameworkID := frameworkResult.CreateFramework.FrameworkEdge.Node.ID

	// Create control
	createControlQuery := `
		mutation CreateControl($input: CreateControlInput!) {
			createControl(input: $input) {
				controlEdge {
					node {
						id
					}
				}
			}
		}
	`

	var controlResult struct {
		CreateControl struct {
			ControlEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"controlEdge"`
		} `json:"createControl"`
	}

	err = owner.Execute(createControlQuery, map[string]any{
		"input": map[string]any{
			"frameworkId":  frameworkID,
			"name":         "SubResolver Test Control",
			"description":  "Test description",
			"sectionTitle": "Section 1",
			"status":       "INCLUDED",
		},
	}, &controlResult)
	require.NoError(t, err)
	controlID := controlResult.CreateControl.ControlEdge.Node.ID

	// Create a measure and link it
	createMeasureQuery := `
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

	var measureResult struct {
		CreateMeasure struct {
			MeasureEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"measureEdge"`
		} `json:"createMeasure"`
	}

	err = owner.Execute(createMeasureQuery, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "Test Measure for Control",
			"category":       "POLICY",
		},
	}, &measureResult)
	require.NoError(t, err)
	measureID := measureResult.CreateMeasure.MeasureEdge.Node.ID

	// Create mapping
	createMappingQuery := `
		mutation CreateControlMeasureMapping($input: CreateControlMeasureMappingInput!) {
			createControlMeasureMapping(input: $input) {
				controlEdge {
					node {
						id
					}
				}
			}
		}
	`

	var mappingResult struct {
		CreateControlMeasureMapping struct {
			ControlEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"controlEdge"`
		} `json:"createControlMeasureMapping"`
	}

	err = owner.Execute(createMappingQuery, map[string]any{
		"input": map[string]any{
			"controlId": controlID,
			"measureId": measureID,
		},
	}, &mappingResult)
	require.NoError(t, err)

	t.Run("Control framework sub-resolver", func(t *testing.T) {
		query := `
			query GetControlFramework($id: ID!) {
				node(id: $id) {
					... on Control {
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

		err := owner.Execute(query, map[string]any{
			"id": controlID,
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, frameworkID, result.Node.Framework.ID)
	})

	t.Run("Control measures sub-resolver", func(t *testing.T) {
		query := `
			query GetControlMeasures($id: ID!) {
				node(id: $id) {
					... on Control {
						id
						measures(first: 10) {
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
				ID       string `json:"id"`
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
			"id": controlID,
		}, &result)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result.Node.Measures.Edges), 1)
	})
}

func TestControl_ExclusionJustification(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create framework
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

	var frameworkResult struct {
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
			"name":           fmt.Sprintf("Exclusion Test %d", time.Now().UnixNano()),
		},
	}, &frameworkResult)
	require.NoError(t, err)
	frameworkID := frameworkResult.CreateFramework.FrameworkEdge.Node.ID

	t.Run("EXCLUDED status should have exclusionJustification", func(t *testing.T) {
		createControlQuery := `
			mutation CreateControl($input: CreateControlInput!) {
				createControl(input: $input) {
					controlEdge {
						node {
							id
							status
							exclusionJustification
						}
					}
				}
			}
		`

		var result struct {
			CreateControl struct {
				ControlEdge struct {
					Node struct {
						ID                     string `json:"id"`
						Status                 string `json:"status"`
						ExclusionJustification string `json:"exclusionJustification"`
					} `json:"node"`
				} `json:"controlEdge"`
			} `json:"createControl"`
		}

		err := owner.Execute(createControlQuery, map[string]any{
			"input": map[string]any{
				"frameworkId":            frameworkID,
				"name":                   "Excluded Control",
				"description":            "Test",
				"sectionTitle":           "Section 1",
				"status":                 "EXCLUDED",
				"exclusionJustification": "Not applicable to our business",
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, "EXCLUDED", result.CreateControl.ControlEdge.Node.Status)
		assert.Equal(t, "Not applicable to our business", result.CreateControl.ControlEdge.Node.ExclusionJustification)
	})

	t.Run("Change status from INCLUDED to EXCLUDED", func(t *testing.T) {
		// Create included control first
		createControlQuery := `
			mutation CreateControl($input: CreateControlInput!) {
				createControl(input: $input) {
					controlEdge {
						node {
							id
						}
					}
				}
			}
		`

		var createResult struct {
			CreateControl struct {
				ControlEdge struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"controlEdge"`
			} `json:"createControl"`
		}

		err := owner.Execute(createControlQuery, map[string]any{
			"input": map[string]any{
				"frameworkId":  frameworkID,
				"name":         "To Be Excluded Control",
				"description":  "Test",
				"sectionTitle": "Section 2",
				"status":       "INCLUDED",
			},
		}, &createResult)
		require.NoError(t, err)

		// Update to excluded
		updateControlQuery := `
			mutation UpdateControl($input: UpdateControlInput!) {
				updateControl(input: $input) {
					control {
						id
						status
						exclusionJustification
					}
				}
			}
		`

		var updateResult struct {
			UpdateControl struct {
				Control struct {
					ID                     string `json:"id"`
					Status                 string `json:"status"`
					ExclusionJustification string `json:"exclusionJustification"`
				} `json:"control"`
			} `json:"updateControl"`
		}

		err = owner.Execute(updateControlQuery, map[string]any{
			"input": map[string]any{
				"id":                     createResult.CreateControl.ControlEdge.Node.ID,
				"status":                 "EXCLUDED",
				"exclusionJustification": "Decided to exclude",
			},
		}, &updateResult)
		require.NoError(t, err)
	})
}
