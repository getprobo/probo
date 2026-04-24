// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestThirdParty_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("with full details", func(t *testing.T) {
		const query = `
			mutation($input: CreateThirdPartyInput!) {
				createThirdParty(input: $input) {
					thirdPartyEdge {
						node {
							id
							name
							description
						}
					}
				}
			}
		`

		var result struct {
			CreateThirdParty struct {
				ThirdPartyEdge struct {
					Node struct {
						ID          string  `json:"id"`
						Name        string  `json:"name"`
						Description *string `json:"description"`
					} `json:"node"`
				} `json:"thirdPartyEdge"`
			} `json:"createThirdParty"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           "AWS",
				"description":    "Amazon Web Services - Cloud Provider",
				"websiteUrl":     "https://aws.amazon.com",
			},
		}, &result)
		require.NoError(t, err)

		assert.NotEmpty(t, result.CreateThirdParty.ThirdPartyEdge.Node.ID)
		assert.Equal(t, "AWS", result.CreateThirdParty.ThirdPartyEdge.Node.Name)
		assert.Equal(t, "Amazon Web Services - Cloud Provider", *result.CreateThirdParty.ThirdPartyEdge.Node.Description)
	})

	t.Run("with all optional fields", func(t *testing.T) {
		const query = `
			mutation($input: CreateThirdPartyInput!) {
				createThirdParty(input: $input) {
					thirdPartyEdge {
						node {
							id
							name
							legalName
							headquarterAddress
							privacyPolicyUrl
							termsOfServiceUrl
							certifications
						}
					}
				}
			}
		`

		var result struct {
			CreateThirdParty struct {
				ThirdPartyEdge struct {
					Node struct {
						ID                 string   `json:"id"`
						Name               string   `json:"name"`
						LegalName          *string  `json:"legalName"`
						HeadquarterAddress *string  `json:"headquarterAddress"`
						PrivacyPolicyUrl   *string  `json:"privacyPolicyUrl"`
						TermsOfServiceUrl  *string  `json:"termsOfServiceUrl"`
						Certifications     []string `json:"certifications"`
					} `json:"node"`
				} `json:"thirdPartyEdge"`
			} `json:"createThirdParty"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"organizationId":     owner.GetOrganizationID().String(),
				"name":               "Stripe",
				"legalName":          "Stripe, Inc.",
				"headquarterAddress": "354 Oyster Point Blvd, South San Francisco, CA",
				"privacyPolicyUrl":   "https://stripe.com/privacy",
				"termsOfServiceUrl":  "https://stripe.com/legal",
				"certifications":     []string{"SOC 2", "PCI DSS"},
			},
		}, &result)
		require.NoError(t, err)

		assert.Equal(t, "Stripe", result.CreateThirdParty.ThirdPartyEdge.Node.Name)
		assert.Equal(t, "Stripe, Inc.", *result.CreateThirdParty.ThirdPartyEdge.Node.LegalName)
		assert.Contains(t, result.CreateThirdParty.ThirdPartyEdge.Node.Certifications, "SOC 2")
	})
}

func TestThirdParty_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	thirdPartyID := factory.CreateThirdParty(owner, factory.Attrs{
		"name":        "ThirdParty to Update",
		"description": "Original description",
	})

	const query = `
		mutation($input: UpdateThirdPartyInput!) {
			updateThirdParty(input: $input) {
				thirdParty {
					id
					name
				}
			}
		}
	`

	var result struct {
		UpdateThirdParty struct {
			ThirdParty struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"thirdParty"`
		} `json:"updateThirdParty"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id":          thirdPartyID,
			"name":        "Updated ThirdParty Name",
			"description": "Updated description",
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, thirdPartyID, result.UpdateThirdParty.ThirdParty.ID)
	assert.Equal(t, "Updated ThirdParty Name", result.UpdateThirdParty.ThirdParty.Name)
}

func TestThirdParty_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	thirdPartyID := factory.CreateThirdParty(owner, factory.Attrs{
		"name": "ThirdParty to Delete",
	})

	const query = `
		mutation($input: DeleteThirdPartyInput!) {
			deleteThirdParty(input: $input) {
				deletedThirdPartyId
			}
		}
	`

	var result struct {
		DeleteThirdParty struct {
			DeletedThirdPartyID string `json:"deletedThirdPartyId"`
		} `json:"deleteThirdParty"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"thirdPartyId": thirdPartyID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, thirdPartyID, result.DeleteThirdParty.DeletedThirdPartyID)
}

func TestThirdParty_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create multiple thirdParties
	thirdPartyNames := []string{"GitHub", "Slack", "Datadog"}
	for _, name := range thirdPartyNames {
		factory.CreateThirdParty(owner, factory.Attrs{"name": name})
	}

	const query = `
		query($orgId: ID!) {
			node(id: $orgId) {
				... on Organization {
					thirdParties(first: 10) {
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
			ThirdParties struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"thirdParties"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"orgId": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.ThirdParties.TotalCount, 3)
}

func TestThirdParty_CreateContact(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	thirdPartyID := factory.CreateThirdParty(owner, factory.Attrs{"name": "ThirdParty With Contact"})

	const query = `
		mutation($input: CreateThirdPartyContactInput!) {
			createThirdPartyContact(input: $input) {
				thirdPartyContactEdge {
					node {
						id
						fullName
						email
						role
					}
				}
			}
		}
	`

	var result struct {
		CreateThirdPartyContact struct {
			ThirdPartyContactEdge struct {
				Node struct {
					ID       string  `json:"id"`
					FullName string  `json:"fullName"`
					Email    string  `json:"email"`
					Role     *string `json:"role"`
				} `json:"node"`
			} `json:"thirdPartyContactEdge"`
		} `json:"createThirdPartyContact"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"thirdPartyId": thirdPartyID,
			"fullName":     "John Contact",
			"email":        "john@thirdParty.com",
			"role":         "Account Manager",
		},
	}, &result)
	require.NoError(t, err)

	assert.NotEmpty(t, result.CreateThirdPartyContact.ThirdPartyContactEdge.Node.ID)
	assert.Equal(t, "John Contact", result.CreateThirdPartyContact.ThirdPartyContactEdge.Node.FullName)
}

func TestThirdParty_RequiredFields(t *testing.T) {
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
				"name": "Test ThirdParty",
			},
			skipOrganization:  true,
			wantErrorContains: "organizationId",
		},
		{
			name:              "missing name",
			input:             map[string]any{},
			wantErrorContains: "name",
		},
		{
			name: "empty name",
			input: map[string]any{
				"name": "",
			},
			wantErrorContains: "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := `
				mutation CreateThirdParty($input: CreateThirdPartyInput!) {
					createThirdParty(input: $input) {
						thirdPartyEdge {
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

func TestThirdParty_CategoryEnum(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	categories := []string{
		"ANALYTICS",
		"CLOUD_PROVIDER",
		"COLLABORATION",
	}

	for _, category := range categories {
		t.Run("create with category "+category, func(t *testing.T) {
			thirdPartyID := factory.NewThirdParty(owner).
				WithName("Category Test " + category).
				WithCategory(category).
				Create()

			query := `
				query($id: ID!) {
					node(id: $id) {
						... on ThirdParty {
							id
							category
						}
					}
				}
			`

			var result struct {
				Node struct {
					ID       string  `json:"id"`
					Category *string `json:"category"`
				} `json:"node"`
			}

			err := owner.Execute(query, map[string]any{"id": thirdPartyID}, &result)
			require.NoError(t, err)
			require.NotNil(t, result.Node.Category)
			assert.Equal(t, category, *result.Node.Category)
		})
	}
}

func TestThirdParty_SubResolvers(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	thirdPartyID := factory.NewThirdParty(owner).
		WithName("SubResolver Test ThirdParty").
		Create()

	t.Run("thirdParty node query", func(t *testing.T) {
		query := `
			query GetThirdParty($id: ID!) {
				node(id: $id) {
					... on ThirdParty {
						id
						name
						description
						websiteUrl
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID          string  `json:"id"`
				Name        string  `json:"name"`
				Description *string `json:"description"`
				WebsiteUrl  *string `json:"websiteUrl"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": thirdPartyID}, &result)
		require.NoError(t, err)
		assert.Equal(t, thirdPartyID, result.Node.ID)
		assert.Equal(t, "SubResolver Test ThirdParty", result.Node.Name)
	})

	t.Run("organization sub-resolver", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on ThirdParty {
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

		err := owner.Execute(query, map[string]any{"id": thirdPartyID}, &result)
		require.NoError(t, err)
		assert.Equal(t, owner.GetOrganizationID().String(), result.Node.Organization.ID)
		assert.NotEmpty(t, result.Node.Organization.Name)
	})

	t.Run("services sub-resolver (empty)", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on ThirdParty {
						id
						services(first: 10) {
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
				Services struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"services"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": thirdPartyID}, &result)
		require.NoError(t, err)
		assert.NotNil(t, result.Node.Services.Edges)
	})

	t.Run("businessOwner sub-resolver (null)", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on ThirdParty {
						id
						businessOwner {
							id
							fullName
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID            string `json:"id"`
				BusinessOwner *struct {
					ID       string `json:"id"`
					FullName string `json:"fullName"`
				} `json:"businessOwner"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": thirdPartyID}, &result)
		require.NoError(t, err)
		assert.Nil(t, result.Node.BusinessOwner)
	})

	t.Run("securityOwner sub-resolver (null)", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on ThirdParty {
						id
						securityOwner {
							id
							fullName
						}
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID            string `json:"id"`
				SecurityOwner *struct {
					ID       string `json:"id"`
					FullName string `json:"fullName"`
				} `json:"securityOwner"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": thirdPartyID}, &result)
		require.NoError(t, err)
		assert.Nil(t, result.Node.SecurityOwner)
	})
}

func TestThirdParty_InvalidID(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("update with invalid ID", func(t *testing.T) {
		query := `
			mutation UpdateThirdParty($input: UpdateThirdPartyInput!) {
				updateThirdParty(input: $input) {
					thirdParty {
						id
					}
				}
			}
		`

		_, err := owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":   "invalid-id-format",
				"name": "Test",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "base64")
	})

	t.Run("delete with invalid ID", func(t *testing.T) {
		query := `
			mutation DeleteThirdParty($input: DeleteThirdPartyInput!) {
				deleteThirdParty(input: $input) {
					deletedThirdPartyId
				}
			}
		`

		_, err := owner.Do(query, map[string]any{
			"input": map[string]any{
				"thirdPartyId": "invalid-id-format",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "base64")
	})

	t.Run("query with non-existent ID", func(t *testing.T) {
		query := `
			query GetThirdParty($id: ID!) {
				node(id: $id) {
					... on ThirdParty {
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

func TestThirdParty_OmittableDescription(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	thirdPartyID := factory.NewThirdParty(owner).
		WithName("Description Test ThirdParty").
		WithDescription("Initial description").
		Create()

	t.Run("set description", func(t *testing.T) {
		query := `
			mutation UpdateThirdParty($input: UpdateThirdPartyInput!) {
				updateThirdParty(input: $input) {
					thirdParty {
						id
						description
					}
				}
			}
		`

		var result struct {
			UpdateThirdParty struct {
				ThirdParty struct {
					ID          string  `json:"id"`
					Description *string `json:"description"`
				} `json:"thirdParty"`
			} `json:"updateThirdParty"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":          thirdPartyID,
				"description": "Updated description",
			},
		}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.UpdateThirdParty.ThirdParty.Description)
		assert.Equal(t, "Updated description", *result.UpdateThirdParty.ThirdParty.Description)
	})

	t.Run("clear description with null", func(t *testing.T) {
		query := `
			mutation UpdateThirdParty($input: UpdateThirdPartyInput!) {
				updateThirdParty(input: $input) {
					thirdParty {
						id
						description
					}
				}
			}
		`

		var result struct {
			UpdateThirdParty struct {
				ThirdParty struct {
					ID          string  `json:"id"`
					Description *string `json:"description"`
				} `json:"thirdParty"`
			} `json:"updateThirdParty"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":          thirdPartyID,
				"description": nil,
			},
		}, &result)
		require.NoError(t, err)
		assert.Nil(t, result.UpdateThirdParty.ThirdParty.Description)
	})

	t.Run("update without description preserves value", func(t *testing.T) {
		// First set a description
		setQuery := `
			mutation UpdateThirdParty($input: UpdateThirdPartyInput!) {
				updateThirdParty(input: $input) {
					thirdParty {
						id
					}
				}
			}
		`

		err := owner.Execute(setQuery, map[string]any{
			"input": map[string]any{
				"id":          thirdPartyID,
				"description": "Should persist",
			},
		}, nil)
		require.NoError(t, err)

		// Update only name
		query := `
			mutation UpdateThirdParty($input: UpdateThirdPartyInput!) {
				updateThirdParty(input: $input) {
					thirdParty {
						id
						name
						description
					}
				}
			}
		`

		var result struct {
			UpdateThirdParty struct {
				ThirdParty struct {
					ID          string  `json:"id"`
					Name        string  `json:"name"`
					Description *string `json:"description"`
				} `json:"thirdParty"`
			} `json:"updateThirdParty"`
		}

		err = owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":   thirdPartyID,
				"name": "Updated Name",
			},
		}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.UpdateThirdParty.ThirdParty.Description)
		assert.Equal(t, "Should persist", *result.UpdateThirdParty.ThirdParty.Description)
	})
}

func TestThirdParty_OmittableBusinessOwner(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a profile for owner assignment
	profileID := factory.CreateUser(owner)
	thirdPartyID := factory.NewThirdParty(owner).
		WithName("BusinessOwner Test ThirdParty").
		Create()

	t.Run("set business owner", func(t *testing.T) {
		query := `
			mutation UpdateThirdParty($input: UpdateThirdPartyInput!) {
				updateThirdParty(input: $input) {
					thirdParty {
						id
						businessOwner {
							id
							fullName
						}
					}
				}
			}
		`

		var result struct {
			UpdateThirdParty struct {
				ThirdParty struct {
					ID            string `json:"id"`
					BusinessOwner struct {
						ID       string `json:"id"`
						FullName string `json:"fullName"`
					} `json:"businessOwner"`
				} `json:"thirdParty"`
			} `json:"updateThirdParty"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":              thirdPartyID,
				"businessOwnerId": profileID,
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, profileID, result.UpdateThirdParty.ThirdParty.BusinessOwner.ID)
	})

	t.Run("clear business owner with null", func(t *testing.T) {
		query := `
			mutation UpdateThirdParty($input: UpdateThirdPartyInput!) {
				updateThirdParty(input: $input) {
					thirdParty {
						id
						businessOwner {
							id
						}
					}
				}
			}
		`

		var result struct {
			UpdateThirdParty struct {
				ThirdParty struct {
					ID            string `json:"id"`
					BusinessOwner *struct {
						ID string `json:"id"`
					} `json:"businessOwner"`
				} `json:"thirdParty"`
			} `json:"updateThirdParty"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":              thirdPartyID,
				"businessOwnerId": nil,
			},
		}, &result)
		require.NoError(t, err)
		assert.Nil(t, result.UpdateThirdParty.ThirdParty.BusinessOwner)
	})
}

func TestThirdParty_OmittableSecurityOwner(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a profile for owner assignment
	profileID := factory.CreateUser(owner)
	thirdPartyID := factory.NewThirdParty(owner).WithName("SecurityOwner Test ThirdParty").Create()

	t.Run("set security owner", func(t *testing.T) {
		query := `
			mutation UpdateThirdParty($input: UpdateThirdPartyInput!) {
				updateThirdParty(input: $input) {
					thirdParty {
						id
						securityOwner {
							id
							fullName
						}
					}
				}
			}
		`

		var result struct {
			UpdateThirdParty struct {
				ThirdParty struct {
					ID            string `json:"id"`
					SecurityOwner struct {
						ID       string `json:"id"`
						FullName string `json:"fullName"`
					} `json:"securityOwner"`
				} `json:"thirdParty"`
			} `json:"updateThirdParty"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":              thirdPartyID,
				"securityOwnerId": profileID,
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, profileID, result.UpdateThirdParty.ThirdParty.SecurityOwner.ID)
	})

	t.Run("clear security owner with null", func(t *testing.T) {
		query := `
			mutation UpdateThirdParty($input: UpdateThirdPartyInput!) {
				updateThirdParty(input: $input) {
					thirdParty {
						id
						securityOwner {
							id
						}
					}
				}
			}
		`

		var result struct {
			UpdateThirdParty struct {
				ThirdParty struct {
					ID            string `json:"id"`
					SecurityOwner *struct {
						ID string `json:"id"`
					} `json:"securityOwner"`
				} `json:"thirdParty"`
			} `json:"updateThirdParty"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":              thirdPartyID,
				"securityOwnerId": nil,
			},
		}, &result)
		require.NoError(t, err)
		assert.Nil(t, result.UpdateThirdParty.ThirdParty.SecurityOwner)
	})
}

func TestThirdParty_OmittableWebsiteUrl(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	thirdPartyID := factory.NewThirdParty(owner).
		WithName("WebsiteUrl Test ThirdParty").
		WithWebsiteUrl("https://example.com").
		Create()

	t.Run("set websiteUrl", func(t *testing.T) {
		query := `
			mutation UpdateThirdParty($input: UpdateThirdPartyInput!) {
				updateThirdParty(input: $input) {
					thirdParty {
						id
						websiteUrl
					}
				}
			}
		`

		var result struct {
			UpdateThirdParty struct {
				ThirdParty struct {
					ID         string  `json:"id"`
					WebsiteUrl *string `json:"websiteUrl"`
				} `json:"thirdParty"`
			} `json:"updateThirdParty"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":         thirdPartyID,
				"websiteUrl": "https://updated.example.com",
			},
		}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.UpdateThirdParty.ThirdParty.WebsiteUrl)
		assert.Equal(t, "https://updated.example.com", *result.UpdateThirdParty.ThirdParty.WebsiteUrl)
	})

	t.Run("clear websiteUrl with null", func(t *testing.T) {
		query := `
			mutation UpdateThirdParty($input: UpdateThirdPartyInput!) {
				updateThirdParty(input: $input) {
					thirdParty {
						id
						websiteUrl
					}
				}
			}
		`

		var result struct {
			UpdateThirdParty struct {
				ThirdParty struct {
					ID         string  `json:"id"`
					WebsiteUrl *string `json:"websiteUrl"`
				} `json:"thirdParty"`
			} `json:"updateThirdParty"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":         thirdPartyID,
				"websiteUrl": nil,
			},
		}, &result)
		require.NoError(t, err)
		assert.Nil(t, result.UpdateThirdParty.ThirdParty.WebsiteUrl)
	})
}

// TestThirdParty_Assess exercises the assessThirdParty mutation through authorization
// and tenant-isolation paths without running the real LLM/browser pipeline.
// The e2e config deliberately omits `llm.vendor-assessor.provider`, so an
// authorized call reaches DisabledVendorAssessor and surfaces a stable
// UNAVAILABLE error. Happy-path payload shape is covered by unit tests in
// pkg/probo.
func TestThirdParty_Assess(t *testing.T) {
	t.Parallel()

	const query = `
		mutation AssessThirdParty($input: AssessThirdPartyInput!) {
			assessThirdParty(input: $input) {
				thirdParty {
					id
				}
			}
		}
	`

	type resultShape struct {
		AssessThirdParty struct {
			ThirdParty struct {
				ID string `json:"id"`
			} `json:"thirdParty"`
		} `json:"assessThirdParty"`
	}

	t.Run("owner call surfaces the disabled error", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		thirdPartyID := factory.NewThirdParty(owner).WithName("Unconfigured assess").Create()

		var result resultShape
		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":         thirdPartyID,
				"websiteUrl": "https://vendor.example.com",
			},
		}, &result)
		testutil.RequireErrorCode(t, err, "UNAVAILABLE")
	})

	t.Run("admin call surfaces the disabled error", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
		thirdPartyID := factory.NewThirdParty(owner).WithName("Admin-assessed third party").Create()

		var result resultShape
		err := admin.Execute(query, map[string]any{
			"input": map[string]any{
				"id":         thirdPartyID,
				"websiteUrl": "https://admin.example.com",
			},
		}, &result)
		testutil.RequireErrorCode(t, err, "UNAVAILABLE")
	})

	t.Run("viewer cannot assess a third party", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		thirdPartyID := factory.NewThirdParty(owner).WithName("Viewer attempt").Create()

		var result resultShape
		err := viewer.Execute(query, map[string]any{
			"input": map[string]any{
				"id":         thirdPartyID,
				"websiteUrl": "https://viewer.example.com",
			},
		}, &result)
		testutil.RequireForbiddenError(t, err)
	})

	t.Run("cannot assess third party from another organization", func(t *testing.T) {
		t.Parallel()

		org1Owner := testutil.NewClient(t, testutil.RoleOwner)
		org2Owner := testutil.NewClient(t, testutil.RoleOwner)
		thirdPartyID := factory.NewThirdParty(org1Owner).WithName("Org1 third party").Create()

		var result resultShape
		err := org2Owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":         thirdPartyID,
				"websiteUrl": "https://cross-tenant.example.com",
			},
		}, &result)
		require.Error(t, err, "third party assess must not cross tenant boundaries")
	})

	t.Run("procedure is accepted on the input", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		thirdPartyID := factory.NewThirdParty(owner).WithName("Procedure test").Create()

		var result resultShape
		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":         thirdPartyID,
				"websiteUrl": "https://procedure.example.com",
				"procedure":  "Focus on SOC 2 controls and data residency",
			},
		}, &result)
		testutil.RequireErrorCode(t, err, "UNAVAILABLE")
	})
}

func TestThirdParty_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	thirdPartyID := factory.NewThirdParty(org1Owner).WithName("Org1 ThirdParty").Create()

	t.Run("cannot read thirdParty from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on ThirdParty {
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

		err := org2Owner.Execute(query, map[string]any{"id": thirdPartyID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "thirdParty")
	})

	t.Run("cannot update thirdParty from another organization", func(t *testing.T) {
		query := `
			mutation UpdateThirdParty($input: UpdateThirdPartyInput!) {
				updateThirdParty(input: $input) {
					thirdParty { id }
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":   thirdPartyID,
				"name": "Hijacked ThirdParty",
			},
		})
		require.Error(t, err, "Should not be able to update thirdParty from another org")
	})

	t.Run("cannot delete thirdParty from another organization", func(t *testing.T) {
		query := `
			mutation DeleteThirdParty($input: DeleteThirdPartyInput!) {
				deleteThirdParty(input: $input) {
					deletedThirdPartyId
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"thirdPartyId": thirdPartyID,
			},
		})
		require.Error(t, err, "Should not be able to delete thirdParty from another org")
	})
}
