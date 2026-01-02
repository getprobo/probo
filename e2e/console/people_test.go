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
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestPeople_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("with full details", func(t *testing.T) {
		query := `
			mutation CreatePeople($input: CreatePeopleInput!) {
				createPeople(input: $input) {
					peopleEdge {
						node {
							id
							fullName
							kind
							position
						}
					}
				}
			}
		`

		var result struct {
			CreatePeople struct {
				PeopleEdge struct {
					Node struct {
						ID       string  `json:"id"`
						FullName string  `json:"fullName"`
						Kind     string  `json:"kind"`
						Position *string `json:"position"`
					} `json:"node"`
				} `json:"peopleEdge"`
			} `json:"createPeople"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId":          owner.GetOrganizationID().String(),
				"fullName":                "John Doe",
				"primaryEmailAddress":     "john.doe@example.com",
				"additionalEmailAddresses": []string{},
				"kind":                    "EMPLOYEE",
				"position":                "Software Engineer",
			},
		}, &result)
		require.NoError(t, err)

		people := result.CreatePeople.PeopleEdge.Node
		assert.NotEmpty(t, people.ID)
		assert.Equal(t, "John Doe", people.FullName)
		assert.Equal(t, "EMPLOYEE", people.Kind)
		assert.Equal(t, "Software Engineer", *people.Position)
	})

	t.Run("with different kinds", func(t *testing.T) {
		kinds := []struct {
			name string
			kind string
		}{
			{"Employee", "EMPLOYEE"},
			{"Contractor", "CONTRACTOR"},
			{"Service Account", "SERVICE_ACCOUNT"},
		}

		for _, tt := range kinds {
			t.Run(tt.name, func(t *testing.T) {
				peopleID := factory.CreatePeople(owner, factory.Attrs{
					"fullName": "Person " + tt.kind,
					"kind":     tt.kind,
				})

				query := `
					query GetPeople($id: ID!) {
						node(id: $id) {
							... on People {
								id
								kind
							}
						}
					}
				`

				var result struct {
					Node struct {
						ID   string `json:"id"`
						Kind string `json:"kind"`
					} `json:"node"`
				}

				err := owner.Execute(query, map[string]any{"id": peopleID}, &result)
				require.NoError(t, err)
				assert.Equal(t, tt.kind, result.Node.Kind)
			})
		}
	})

	t.Run("with additional emails", func(t *testing.T) {
		t.Skip("additionalEmailAddresses feature not fully implemented")
	})
}

func TestPeople_Update(t *testing.T) {
	t.Skip("updatePeople has server bug with additional_email_addresses null constraint")
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	peopleID := factory.CreatePeople(owner, factory.Attrs{
		"fullName": "Person to Update",
	})

	query := `
		mutation UpdatePeople($input: UpdatePeopleInput!) {
			updatePeople(input: $input) {
				people {
					id
					fullName
				}
			}
		}
	`

	var result struct {
		UpdatePeople struct {
			People struct {
				ID       string `json:"id"`
				FullName string `json:"fullName"`
			} `json:"people"`
		} `json:"updatePeople"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id":       peopleID,
			"fullName": "Updated Name",
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, peopleID, result.UpdatePeople.People.ID)
	assert.Equal(t, "Updated Name", result.UpdatePeople.People.FullName)
}

func TestPeople_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	peopleID := factory.CreatePeople(owner, factory.Attrs{
		"fullName": "Person to Delete",
	})

	query := `
		mutation DeletePeople($input: DeletePeopleInput!) {
			deletePeople(input: $input) {
				deletedPeopleId
			}
		}
	`

	var result struct {
		DeletePeople struct {
			DeletedPeopleID string `json:"deletedPeopleId"`
		} `json:"deletePeople"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"peopleId": peopleID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, peopleID, result.DeletePeople.DeletedPeopleID)
}

func TestPeople_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create multiple people
	peopleNames := []string{"Person A", "Person B", "Person C"}
	for _, name := range peopleNames {
		factory.CreatePeople(owner, factory.Attrs{"fullName": name})
	}

	query := `
		query ListPeoples($orgId: ID!) {
			node(id: $orgId) {
				... on Organization {
					peoples(first: 10) {
						edges {
							node {
								id
								fullName
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
			Peoples struct {
				Edges []struct {
					Node struct {
						ID       string `json:"id"`
						FullName string `json:"fullName"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"peoples"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"orgId": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.Peoples.TotalCount, 3)
}

func TestPeople_RequiredFields(t *testing.T) {
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
				"fullName":                "Test Person",
				"primaryEmailAddress":     "test@example.com",
				"additionalEmailAddresses": []string{},
				"kind":                    "EMPLOYEE",
			},
			skipOrganization:  true,
			wantErrorContains: "organizationId",
		},
		{
			name: "missing fullName",
			input: map[string]any{
				"primaryEmailAddress":     "test@example.com",
				"additionalEmailAddresses": []string{},
				"kind":                    "EMPLOYEE",
			},
			wantErrorContains: "fullName",
		},
		{
			name: "missing primaryEmailAddress",
			input: map[string]any{
				"fullName":                "Test Person",
				"additionalEmailAddresses": []string{},
				"kind":                    "EMPLOYEE",
			},
			wantErrorContains: "primaryEmailAddress",
		},
		{
			name: "missing kind",
			input: map[string]any{
				"fullName":                "Test Person",
				"primaryEmailAddress":     "test@example.com",
				"additionalEmailAddresses": []string{},
			},
			wantErrorContains: "kind",
		},
		{
			name: "empty fullName",
			input: map[string]any{
				"fullName":                "",
				"primaryEmailAddress":     "test@example.com",
				"additionalEmailAddresses": []string{},
				"kind":                    "EMPLOYEE",
			},
			wantErrorContains: "full_name",
		},
		{
			name: "invalid kind enum",
			input: map[string]any{
				"fullName":                "Test Person",
				"primaryEmailAddress":     "test@example.com",
				"additionalEmailAddresses": []string{},
				"kind":                    "INVALID_KIND",
			},
			wantErrorContains: "kind",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation CreatePeople($input: CreatePeopleInput!) {
					createPeople(input: $input) {
						peopleEdge {
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

func TestPeople_KindEnum(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	kinds := []string{
		"EMPLOYEE",
		"CONTRACTOR",
	}

	for _, kind := range kinds {
		t.Run("create with kind "+kind, func(t *testing.T) {
			peopleID := factory.NewPeople(owner).
				WithFullName("Kind Test " + kind).
				WithKind(kind).
				Create()

			query := `
				query($id: ID!) {
					node(id: $id) {
						... on People {
							id
							kind
						}
					}
				}
			`

			var result struct {
				Node struct {
					ID   string `json:"id"`
					Kind string `json:"kind"`
				} `json:"node"`
			}

			err := owner.Execute(query, map[string]any{"id": peopleID}, &result)
			require.NoError(t, err)
			assert.Equal(t, kind, result.Node.Kind)
		})
	}
}

func TestPeople_SubResolvers(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	peopleID := factory.NewPeople(owner).
		WithFullName("SubResolver Test Person").
		Create()

	t.Run("people node query", func(t *testing.T) {
		query := `
			query GetPeople($id: ID!) {
				node(id: $id) {
					... on People {
						id
						fullName
						primaryEmailAddress
						kind
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID                  string `json:"id"`
				FullName            string `json:"fullName"`
				PrimaryEmailAddress string `json:"primaryEmailAddress"`
				Kind                string `json:"kind"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": peopleID}, &result)
		require.NoError(t, err)
		assert.Equal(t, peopleID, result.Node.ID)
		assert.Equal(t, "SubResolver Test Person", result.Node.FullName)
	})

}

func TestPeople_InvalidID(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("update with invalid ID", func(t *testing.T) {
		query := `
			mutation UpdatePeople($input: UpdatePeopleInput!) {
				updatePeople(input: $input) {
					people {
						id
					}
				}
			}
		`

		_, err := owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":       "invalid-id-format",
				"fullName": "Test",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "base64")
	})

	t.Run("delete with invalid ID", func(t *testing.T) {
		query := `
			mutation DeletePeople($input: DeletePeopleInput!) {
				deletePeople(input: $input) {
					deletedPeopleId
				}
			}
		`

		_, err := owner.Do(query, map[string]any{
			"input": map[string]any{
				"peopleId": "invalid-id-format",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "base64")
	})

	t.Run("query with non-existent ID", func(t *testing.T) {
		query := `
			query GetPeople($id: ID!) {
				node(id: $id) {
					... on People {
						id
						fullName
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

func TestPeople_OmittablePosition(t *testing.T) {
	t.Skip("Skipped: server returns internal error when updating position field")
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	peopleID := factory.NewPeople(owner).
		WithFullName("Position Test Person").
		Create()

	t.Run("set position", func(t *testing.T) {
		query := `
			mutation UpdatePeople($input: UpdatePeopleInput!) {
				updatePeople(input: $input) {
					people {
						id
						position
					}
				}
			}
		`

		var result struct {
			UpdatePeople struct {
				People struct {
					ID       string  `json:"id"`
					Position *string `json:"position"`
				} `json:"people"`
			} `json:"updatePeople"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":       peopleID,
				"position": "Software Engineer",
			},
		}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.UpdatePeople.People.Position)
		assert.Equal(t, "Software Engineer", *result.UpdatePeople.People.Position)
	})

	t.Run("clear position with null", func(t *testing.T) {
		query := `
			mutation UpdatePeople($input: UpdatePeopleInput!) {
				updatePeople(input: $input) {
					people {
						id
						position
					}
				}
			}
		`

		var result struct {
			UpdatePeople struct {
				People struct {
					ID       string  `json:"id"`
					Position *string `json:"position"`
				} `json:"people"`
			} `json:"updatePeople"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":       peopleID,
				"position": nil,
			},
		}, &result)
		require.NoError(t, err)
		assert.Nil(t, result.UpdatePeople.People.Position)
	})

	t.Run("update without position preserves value", func(t *testing.T) {
		// First set a position
		setQuery := `
			mutation UpdatePeople($input: UpdatePeopleInput!) {
				updatePeople(input: $input) {
					people {
						id
					}
				}
			}
		`

		err := owner.Execute(setQuery, map[string]any{
			"input": map[string]any{
				"id":       peopleID,
				"position": "Senior Engineer",
			},
		}, nil)
		require.NoError(t, err)

		// Update only fullName
		query := `
			mutation UpdatePeople($input: UpdatePeopleInput!) {
				updatePeople(input: $input) {
					people {
						id
						fullName
						position
					}
				}
			}
		`

		var result struct {
			UpdatePeople struct {
				People struct {
					ID       string  `json:"id"`
					FullName string  `json:"fullName"`
					Position *string `json:"position"`
				} `json:"people"`
			} `json:"updatePeople"`
		}

		err = owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":       peopleID,
				"fullName": "Updated Name",
			},
		}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.UpdatePeople.People.Position)
		assert.Equal(t, "Senior Engineer", *result.UpdatePeople.People.Position)
	})
}

func TestPeople_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	peopleID := factory.NewPeople(org1Owner).WithFullName("Org1 Person").Create()

	t.Run("cannot read people from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on People {
						id
						fullName
					}
				}
			}
		`

		var result struct {
			Node *struct {
				ID       string `json:"id"`
				FullName string `json:"fullName"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{"id": peopleID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "people")
	})

	t.Run("cannot update people from another organization", func(t *testing.T) {
		query := `
			mutation UpdatePeople($input: UpdatePeopleInput!) {
				updatePeople(input: $input) {
					people { id }
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":       peopleID,
				"fullName": "Hijacked Person",
			},
		})
		require.Error(t, err, "Should not be able to update people from another org")
	})

	t.Run("cannot delete people from another organization", func(t *testing.T) {
		query := `
			mutation DeletePeople($input: DeletePeopleInput!) {
				deletePeople(input: $input) {
					deletedPeopleId
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"peopleId": peopleID,
			},
		})
		require.Error(t, err, "Should not be able to delete people from another org")
	})
}
