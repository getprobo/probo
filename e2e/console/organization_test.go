// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestOrganization_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("update name", func(t *testing.T) {
		newName := fmt.Sprintf("Updated Org %d", time.Now().UnixNano())

		query := `
			mutation UpdateOrganization($input: UpdateOrganizationInput!) {
				updateOrganization(input: $input) {
					organization {
						id
						name
					}
				}
			}
		`

		var result struct {
			UpdateOrganization struct {
				Organization struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"organization"`
			} `json:"updateOrganization"`
		}

		err := owner.ExecuteConnect(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           newName,
			},
		}, &result)
		require.NoError(t, err)

		assert.Equal(t, owner.GetOrganizationID().String(), result.UpdateOrganization.Organization.ID)
		assert.Equal(t, newName, result.UpdateOrganization.Organization.Name)
	})
}

func TestTrustCenter_UpdateContactInfo(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	organizationID := owner.GetOrganizationID().String()

	const trustCenterQuery = `
		query($organizationId: ID!) {
			node(id: $organizationId) {
				... on Organization {
					trustCenter {
						id
					}
				}
			}
		}
	`

	var trustCenterLookup struct {
		Node struct {
			TrustCenter struct {
				ID string `json:"id"`
			} `json:"trustCenter"`
		} `json:"node"`
	}

	err := owner.Execute(trustCenterQuery, map[string]any{
		"organizationId": organizationID,
	}, &trustCenterLookup)
	require.NoError(t, err)
	require.NotEmpty(t, trustCenterLookup.Node.TrustCenter.ID)

	trustCenterID := trustCenterLookup.Node.TrustCenter.ID

	t.Run("update description", func(t *testing.T) {
		query := `
			mutation UpdateTrustCenterBrand($input: UpdateTrustCenterBrandInput!) {
				updateTrustCenterBrand(input: $input) {
					trustCenter {
						id
						description
					}
				}
			}
		`

		var result struct {
			UpdateTrustCenterBrand struct {
				TrustCenter struct {
					ID          string `json:"id"`
					Description string `json:"description"`
				} `json:"trustCenter"`
			} `json:"updateTrustCenterBrand"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"trustCenterId": trustCenterID,
				"description":   "Updated compliance page description",
			},
		}, &result)
		require.NoError(t, err)

		assert.Equal(t, trustCenterID, result.UpdateTrustCenterBrand.TrustCenter.ID)
		assert.Equal(t, "Updated compliance page description", result.UpdateTrustCenterBrand.TrustCenter.Description)
	})

	t.Run("update website and email", func(t *testing.T) {
		query := `
			mutation UpdateTrustCenterBrand($input: UpdateTrustCenterBrandInput!) {
				updateTrustCenterBrand(input: $input) {
					trustCenter {
						id
						websiteUrl
						email
					}
				}
			}
		`

		var result struct {
			UpdateTrustCenterBrand struct {
				TrustCenter struct {
					ID         string `json:"id"`
					WebsiteUrl string `json:"websiteUrl"`
					Email      string `json:"email"`
				} `json:"trustCenter"`
			} `json:"updateTrustCenterBrand"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"trustCenterId": trustCenterID,
				"websiteUrl":    "https://example.com",
				"email":         "contact@example.com",
			},
		}, &result)
		require.NoError(t, err)

		assert.Equal(t, "https://example.com", result.UpdateTrustCenterBrand.TrustCenter.WebsiteUrl)
		assert.Equal(t, "contact@example.com", result.UpdateTrustCenterBrand.TrustCenter.Email)
	})

	t.Run("update headquarter address", func(t *testing.T) {
		query := `
			mutation UpdateTrustCenterBrand($input: UpdateTrustCenterBrandInput!) {
				updateTrustCenterBrand(input: $input) {
					trustCenter {
						id
						headquarterAddress
					}
				}
			}
		`

		var result struct {
			UpdateTrustCenterBrand struct {
				TrustCenter struct {
					ID                 string `json:"id"`
					HeadquarterAddress string `json:"headquarterAddress"`
				} `json:"trustCenter"`
			} `json:"updateTrustCenterBrand"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"trustCenterId":      trustCenterID,
				"headquarterAddress": "123 Main St, Suite 100, San Francisco, CA 94102",
			},
		}, &result)
		require.NoError(t, err)

		assert.Equal(t, "123 Main St, Suite 100, San Francisco, CA 94102", result.UpdateTrustCenterBrand.TrustCenter.HeadquarterAddress)
	})
}

func TestOrganization_UpdateContext(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	query := `
		mutation UpdateOrganizationContext($input: UpdateOrganizationContextInput!) {
			updateOrganizationContext(input: $input) {
				context {
					organizationId
					product
					architecture
					team
				}
			}
		}
	`

	var result struct {
		UpdateOrganizationContext struct {
			Context struct {
				OrganizationID string  `json:"organizationId"`
				Product        *string `json:"product"`
				Architecture   *string `json:"architecture"`
				Team           *string `json:"team"`
			} `json:"context"`
		} `json:"updateOrganizationContext"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"product":        "Our product provides compliance solutions.",
			"architecture":   "Microservices architecture on AWS.",
			"team":           "Small engineering team.",
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, owner.GetOrganizationID().String(), result.UpdateOrganizationContext.Context.OrganizationID)
	require.NotNil(t, result.UpdateOrganizationContext.Context.Product)
	assert.Equal(t, "Our product provides compliance solutions.", *result.UpdateOrganizationContext.Context.Product)
	require.NotNil(t, result.UpdateOrganizationContext.Context.Architecture)
	assert.Equal(t, "Microservices architecture on AWS.", *result.UpdateOrganizationContext.Context.Architecture)
}

func TestOrganization_Get(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	query := `
		query GetOrganization($id: ID!) {
			node(id: $id) {
				... on Organization {
					id
					name
				}
			}
		}
	`

	var result struct {
		Node struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"node"`
	}

	err := owner.ExecuteConnect(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, owner.GetOrganizationID().String(), result.Node.ID)
	assert.NotEmpty(t, result.Node.Name)
}
