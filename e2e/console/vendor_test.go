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

func TestVendor_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("with full details", func(t *testing.T) {
		const query = `
			mutation($input: CreateVendorInput!) {
				createVendor(input: $input) {
					vendorEdge {
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
			CreateVendor struct {
				VendorEdge struct {
					Node struct {
						ID          string  `json:"id"`
						Name        string  `json:"name"`
						Description *string `json:"description"`
					} `json:"node"`
				} `json:"vendorEdge"`
			} `json:"createVendor"`
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

		assert.NotEmpty(t, result.CreateVendor.VendorEdge.Node.ID)
		assert.Equal(t, "AWS", result.CreateVendor.VendorEdge.Node.Name)
		assert.Equal(t, "Amazon Web Services - Cloud Provider", *result.CreateVendor.VendorEdge.Node.Description)
	})

	t.Run("with all optional fields", func(t *testing.T) {
		const query = `
			mutation($input: CreateVendorInput!) {
				createVendor(input: $input) {
					vendorEdge {
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
			CreateVendor struct {
				VendorEdge struct {
					Node struct {
						ID                 string   `json:"id"`
						Name               string   `json:"name"`
						LegalName          *string  `json:"legalName"`
						HeadquarterAddress *string  `json:"headquarterAddress"`
						PrivacyPolicyUrl   *string  `json:"privacyPolicyUrl"`
						TermsOfServiceUrl  *string  `json:"termsOfServiceUrl"`
						Certifications     []string `json:"certifications"`
					} `json:"node"`
				} `json:"vendorEdge"`
			} `json:"createVendor"`
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

		assert.Equal(t, "Stripe", result.CreateVendor.VendorEdge.Node.Name)
		assert.Equal(t, "Stripe, Inc.", *result.CreateVendor.VendorEdge.Node.LegalName)
		assert.Contains(t, result.CreateVendor.VendorEdge.Node.Certifications, "SOC 2")
	})
}

func TestVendor_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	vendorID := factory.CreateVendor(owner, factory.Attrs{
		"name":        "Vendor to Update",
		"description": "Original description",
	})

	const query = `
		mutation($input: UpdateVendorInput!) {
			updateVendor(input: $input) {
				vendor {
					id
					name
				}
			}
		}
	`

	var result struct {
		UpdateVendor struct {
			Vendor struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"vendor"`
		} `json:"updateVendor"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id":          vendorID,
			"name":        "Updated Vendor Name",
			"description": "Updated description",
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, vendorID, result.UpdateVendor.Vendor.ID)
	assert.Equal(t, "Updated Vendor Name", result.UpdateVendor.Vendor.Name)
}

func TestVendor_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	vendorID := factory.CreateVendor(owner, factory.Attrs{
		"name": "Vendor to Delete",
	})

	const query = `
		mutation($input: DeleteVendorInput!) {
			deleteVendor(input: $input) {
				deletedVendorId
			}
		}
	`

	var result struct {
		DeleteVendor struct {
			DeletedVendorID string `json:"deletedVendorId"`
		} `json:"deleteVendor"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"vendorId": vendorID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, vendorID, result.DeleteVendor.DeletedVendorID)
}

func TestVendor_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create multiple vendors
	vendorNames := []string{"GitHub", "Slack", "Datadog"}
	for _, name := range vendorNames {
		factory.CreateVendor(owner, factory.Attrs{"name": name})
	}

	const query = `
		query($orgId: ID!) {
			node(id: $orgId) {
				... on Organization {
					vendors(first: 10) {
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
			Vendors struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"vendors"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"orgId": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Node.Vendors.TotalCount, 3)
}

func TestVendor_CreateContact(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	vendorID := factory.CreateVendor(owner, factory.Attrs{"name": "Vendor With Contact"})

	const query = `
		mutation($input: CreateVendorContactInput!) {
			createVendorContact(input: $input) {
				vendorContactEdge {
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
		CreateVendorContact struct {
			VendorContactEdge struct {
				Node struct {
					ID       string  `json:"id"`
					FullName string  `json:"fullName"`
					Email    string  `json:"email"`
					Role     *string `json:"role"`
				} `json:"node"`
			} `json:"vendorContactEdge"`
		} `json:"createVendorContact"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"vendorId": vendorID,
			"fullName": "John Contact",
			"email":    "john@vendor.com",
			"role":     "Account Manager",
		},
	}, &result)
	require.NoError(t, err)

	assert.NotEmpty(t, result.CreateVendorContact.VendorContactEdge.Node.ID)
	assert.Equal(t, "John Contact", result.CreateVendorContact.VendorContactEdge.Node.FullName)
}

func TestVendor_RequiredFields(t *testing.T) {
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
				"name": "Test Vendor",
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
				mutation CreateVendor($input: CreateVendorInput!) {
					createVendor(input: $input) {
						vendorEdge {
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

func TestVendor_CategoryEnum(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	categories := []string{
		"ANALYTICS",
		"CLOUD_PROVIDER",
		"COLLABORATION",
	}

	for _, category := range categories {
		t.Run("create with category "+category, func(t *testing.T) {
			vendorID := factory.NewVendor(owner).
				WithName("Category Test " + category).
				WithCategory(category).
				Create()

			query := `
				query($id: ID!) {
					node(id: $id) {
						... on Vendor {
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

			err := owner.Execute(query, map[string]any{"id": vendorID}, &result)
			require.NoError(t, err)
			require.NotNil(t, result.Node.Category)
			assert.Equal(t, category, *result.Node.Category)
		})
	}
}

func TestVendor_SubResolvers(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	vendorID := factory.NewVendor(owner).
		WithName("SubResolver Test Vendor").
		Create()

	t.Run("vendor node query", func(t *testing.T) {
		query := `
			query GetVendor($id: ID!) {
				node(id: $id) {
					... on Vendor {
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

		err := owner.Execute(query, map[string]any{"id": vendorID}, &result)
		require.NoError(t, err)
		assert.Equal(t, vendorID, result.Node.ID)
		assert.Equal(t, "SubResolver Test Vendor", result.Node.Name)
	})

	t.Run("organization sub-resolver", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Vendor {
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

		err := owner.Execute(query, map[string]any{"id": vendorID}, &result)
		require.NoError(t, err)
		assert.Equal(t, owner.GetOrganizationID().String(), result.Node.Organization.ID)
		assert.NotEmpty(t, result.Node.Organization.Name)
	})

	t.Run("services sub-resolver (empty)", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Vendor {
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

		err := owner.Execute(query, map[string]any{"id": vendorID}, &result)
		require.NoError(t, err)
		assert.NotNil(t, result.Node.Services.Edges)
	})

	t.Run("businessOwner sub-resolver (null)", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Vendor {
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

		err := owner.Execute(query, map[string]any{"id": vendorID}, &result)
		require.NoError(t, err)
		assert.Nil(t, result.Node.BusinessOwner)
	})

	t.Run("securityOwner sub-resolver (null)", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Vendor {
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

		err := owner.Execute(query, map[string]any{"id": vendorID}, &result)
		require.NoError(t, err)
		assert.Nil(t, result.Node.SecurityOwner)
	})
}

func TestVendor_InvalidID(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("update with invalid ID", func(t *testing.T) {
		query := `
			mutation UpdateVendor($input: UpdateVendorInput!) {
				updateVendor(input: $input) {
					vendor {
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
			mutation DeleteVendor($input: DeleteVendorInput!) {
				deleteVendor(input: $input) {
					deletedVendorId
				}
			}
		`

		_, err := owner.Do(query, map[string]any{
			"input": map[string]any{
				"vendorId": "invalid-id-format",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "base64")
	})

	t.Run("query with non-existent ID", func(t *testing.T) {
		query := `
			query GetVendor($id: ID!) {
				node(id: $id) {
					... on Vendor {
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

func TestVendor_OmittableDescription(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	vendorID := factory.NewVendor(owner).
		WithName("Description Test Vendor").
		WithDescription("Initial description").
		Create()

	t.Run("set description", func(t *testing.T) {
		query := `
			mutation UpdateVendor($input: UpdateVendorInput!) {
				updateVendor(input: $input) {
					vendor {
						id
						description
					}
				}
			}
		`

		var result struct {
			UpdateVendor struct {
				Vendor struct {
					ID          string  `json:"id"`
					Description *string `json:"description"`
				} `json:"vendor"`
			} `json:"updateVendor"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":          vendorID,
				"description": "Updated description",
			},
		}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.UpdateVendor.Vendor.Description)
		assert.Equal(t, "Updated description", *result.UpdateVendor.Vendor.Description)
	})

	t.Run("clear description with null", func(t *testing.T) {
		query := `
			mutation UpdateVendor($input: UpdateVendorInput!) {
				updateVendor(input: $input) {
					vendor {
						id
						description
					}
				}
			}
		`

		var result struct {
			UpdateVendor struct {
				Vendor struct {
					ID          string  `json:"id"`
					Description *string `json:"description"`
				} `json:"vendor"`
			} `json:"updateVendor"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":          vendorID,
				"description": nil,
			},
		}, &result)
		require.NoError(t, err)
		assert.Nil(t, result.UpdateVendor.Vendor.Description)
	})

	t.Run("update without description preserves value", func(t *testing.T) {
		// First set a description
		setQuery := `
			mutation UpdateVendor($input: UpdateVendorInput!) {
				updateVendor(input: $input) {
					vendor {
						id
					}
				}
			}
		`

		err := owner.Execute(setQuery, map[string]any{
			"input": map[string]any{
				"id":          vendorID,
				"description": "Should persist",
			},
		}, nil)
		require.NoError(t, err)

		// Update only name
		query := `
			mutation UpdateVendor($input: UpdateVendorInput!) {
				updateVendor(input: $input) {
					vendor {
						id
						name
						description
					}
				}
			}
		`

		var result struct {
			UpdateVendor struct {
				Vendor struct {
					ID          string  `json:"id"`
					Name        string  `json:"name"`
					Description *string `json:"description"`
				} `json:"vendor"`
			} `json:"updateVendor"`
		}

		err = owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":   vendorID,
				"name": "Updated Name",
			},
		}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.UpdateVendor.Vendor.Description)
		assert.Equal(t, "Should persist", *result.UpdateVendor.Vendor.Description)
	})
}

func TestVendor_OmittableBusinessOwner(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a profile for owner assignment
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
	vendorID := factory.NewVendor(owner).
		WithName("BusinessOwner Test Vendor").
		Create()

	t.Run("set business owner", func(t *testing.T) {
		query := `
			mutation UpdateVendor($input: UpdateVendorInput!) {
				updateVendor(input: $input) {
					vendor {
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
			UpdateVendor struct {
				Vendor struct {
					ID            string `json:"id"`
					BusinessOwner struct {
						ID       string `json:"id"`
						FullName string `json:"fullName"`
					} `json:"businessOwner"`
				} `json:"vendor"`
			} `json:"updateVendor"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":              vendorID,
				"businessOwnerId": profileID.String(),
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, profileID.String(), result.UpdateVendor.Vendor.BusinessOwner.ID)
	})

	t.Run("clear business owner with null", func(t *testing.T) {
		query := `
			mutation UpdateVendor($input: UpdateVendorInput!) {
				updateVendor(input: $input) {
					vendor {
						id
						businessOwner {
							id
						}
					}
				}
			}
		`

		var result struct {
			UpdateVendor struct {
				Vendor struct {
					ID            string `json:"id"`
					BusinessOwner *struct {
						ID string `json:"id"`
					} `json:"businessOwner"`
				} `json:"vendor"`
			} `json:"updateVendor"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":              vendorID,
				"businessOwnerId": nil,
			},
		}, &result)
		require.NoError(t, err)
		assert.Nil(t, result.UpdateVendor.Vendor.BusinessOwner)
	})
}

func TestVendor_OmittableSecurityOwner(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a profile for owner assignment
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()
	vendorID := factory.NewVendor(owner).WithName("SecurityOwner Test Vendor").Create()

	t.Run("set security owner", func(t *testing.T) {
		query := `
			mutation UpdateVendor($input: UpdateVendorInput!) {
				updateVendor(input: $input) {
					vendor {
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
			UpdateVendor struct {
				Vendor struct {
					ID            string `json:"id"`
					SecurityOwner struct {
						ID       string `json:"id"`
						FullName string `json:"fullName"`
					} `json:"securityOwner"`
				} `json:"vendor"`
			} `json:"updateVendor"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":              vendorID,
				"securityOwnerId": profileID.String(),
			},
		}, &result)
		require.NoError(t, err)
		assert.Equal(t, profileID.String(), result.UpdateVendor.Vendor.SecurityOwner.ID)
	})

	t.Run("clear security owner with null", func(t *testing.T) {
		query := `
			mutation UpdateVendor($input: UpdateVendorInput!) {
				updateVendor(input: $input) {
					vendor {
						id
						securityOwner {
							id
						}
					}
				}
			}
		`

		var result struct {
			UpdateVendor struct {
				Vendor struct {
					ID            string `json:"id"`
					SecurityOwner *struct {
						ID string `json:"id"`
					} `json:"securityOwner"`
				} `json:"vendor"`
			} `json:"updateVendor"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":              vendorID,
				"securityOwnerId": nil,
			},
		}, &result)
		require.NoError(t, err)
		assert.Nil(t, result.UpdateVendor.Vendor.SecurityOwner)
	})
}

func TestVendor_OmittableWebsiteUrl(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	vendorID := factory.NewVendor(owner).
		WithName("WebsiteUrl Test Vendor").
		WithWebsiteUrl("https://example.com").
		Create()

	t.Run("set websiteUrl", func(t *testing.T) {
		query := `
			mutation UpdateVendor($input: UpdateVendorInput!) {
				updateVendor(input: $input) {
					vendor {
						id
						websiteUrl
					}
				}
			}
		`

		var result struct {
			UpdateVendor struct {
				Vendor struct {
					ID         string  `json:"id"`
					WebsiteUrl *string `json:"websiteUrl"`
				} `json:"vendor"`
			} `json:"updateVendor"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":         vendorID,
				"websiteUrl": "https://updated.example.com",
			},
		}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.UpdateVendor.Vendor.WebsiteUrl)
		assert.Equal(t, "https://updated.example.com", *result.UpdateVendor.Vendor.WebsiteUrl)
	})

	t.Run("clear websiteUrl with null", func(t *testing.T) {
		query := `
			mutation UpdateVendor($input: UpdateVendorInput!) {
				updateVendor(input: $input) {
					vendor {
						id
						websiteUrl
					}
				}
			}
		`

		var result struct {
			UpdateVendor struct {
				Vendor struct {
					ID         string  `json:"id"`
					WebsiteUrl *string `json:"websiteUrl"`
				} `json:"vendor"`
			} `json:"updateVendor"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"id":         vendorID,
				"websiteUrl": nil,
			},
		}, &result)
		require.NoError(t, err)
		assert.Nil(t, result.UpdateVendor.Vendor.WebsiteUrl)
	})
}

func TestVendor_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	vendorID := factory.NewVendor(org1Owner).WithName("Org1 Vendor").Create()

	t.Run("cannot read vendor from another organization", func(t *testing.T) {
		query := `
			query($id: ID!) {
				node(id: $id) {
					... on Vendor {
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

		err := org2Owner.Execute(query, map[string]any{"id": vendorID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "vendor")
	})

	t.Run("cannot update vendor from another organization", func(t *testing.T) {
		query := `
			mutation UpdateVendor($input: UpdateVendorInput!) {
				updateVendor(input: $input) {
					vendor { id }
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"id":   vendorID,
				"name": "Hijacked Vendor",
			},
		})
		require.Error(t, err, "Should not be able to update vendor from another org")
	})

	t.Run("cannot delete vendor from another organization", func(t *testing.T) {
		query := `
			mutation DeleteVendor($input: DeleteVendorInput!) {
				deleteVendor(input: $input) {
					deletedVendorId
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"vendorId": vendorID,
			},
		})
		require.Error(t, err, "Should not be able to delete vendor from another org")
	})
}
