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
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestOrganization_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	t.Run("update name and description", func(t *testing.T) {
		newName := fmt.Sprintf("Updated Org %d", time.Now().UnixNano())

		query := `
			mutation UpdateOrganization($input: UpdateOrganizationInput!) {
				updateOrganization(input: $input) {
					organization {
						id
						name
						description
					}
				}
			}
		`

		var result struct {
			UpdateOrganization struct {
				Organization struct {
					ID          string `json:"id"`
					Name        string `json:"name"`
					Description string `json:"description"`
				} `json:"organization"`
			} `json:"updateOrganization"`
		}

		err := owner.ExecuteConnect(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"name":           newName,
				"description":    "Updated organization description",
			},
		}, &result)
		require.NoError(t, err)

		assert.Equal(t, owner.GetOrganizationID().String(), result.UpdateOrganization.Organization.ID)
		assert.Equal(t, newName, result.UpdateOrganization.Organization.Name)
		assert.Equal(t, "Updated organization description", result.UpdateOrganization.Organization.Description)
	})

	t.Run("update website and email", func(t *testing.T) {
		query := `
			mutation UpdateOrganization($input: UpdateOrganizationInput!) {
				updateOrganization(input: $input) {
					organization {
						id
						websiteUrl
						email
					}
				}
			}
		`

		var result struct {
			UpdateOrganization struct {
				Organization struct {
					ID         string `json:"id"`
					WebsiteUrl string `json:"websiteUrl"`
					Email      string `json:"email"`
				} `json:"organization"`
			} `json:"updateOrganization"`
		}

		err := owner.ExecuteConnect(query, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"websiteUrl":     "https://example.com",
				"email":          "contact@example.com",
			},
		}, &result)
		require.NoError(t, err)

		assert.Equal(t, "https://example.com", result.UpdateOrganization.Organization.WebsiteUrl)
		assert.Equal(t, "contact@example.com", result.UpdateOrganization.Organization.Email)
	})

	t.Run("update headquarter address", func(t *testing.T) {
		query := `
			mutation UpdateOrganization($input: UpdateOrganizationInput!) {
				updateOrganization(input: $input) {
					organization {
						id
						headquarterAddress
					}
				}
			}
		`

		var result struct {
			UpdateOrganization struct {
				Organization struct {
					ID                 string `json:"id"`
					HeadquarterAddress string `json:"headquarterAddress"`
				} `json:"organization"`
			} `json:"updateOrganization"`
		}

		err := owner.ExecuteConnect(query, map[string]any{
			"input": map[string]any{
				"organizationId":     owner.GetOrganizationID().String(),
				"headquarterAddress": "123 Main St, Suite 100, San Francisco, CA 94102",
			},
		}, &result)
		require.NoError(t, err)

		assert.Equal(t, "123 Main St, Suite 100, San Francisco, CA 94102", result.UpdateOrganization.Organization.HeadquarterAddress)
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
					summary
				}
			}
		}
	`

	var result struct {
		UpdateOrganizationContext struct {
			Context struct {
				OrganizationID string  `json:"organizationId"`
				Summary        *string `json:"summary"`
			} `json:"context"`
		} `json:"updateOrganizationContext"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"summary":        "Our organization provides compliance solutions.",
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, owner.GetOrganizationID().String(), result.UpdateOrganizationContext.Context.OrganizationID)
	require.NotNil(t, result.UpdateOrganizationContext.Context.Summary)
	assert.Equal(t, "Our organization provides compliance solutions.", *result.UpdateOrganizationContext.Context.Summary)
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
					description
					websiteUrl
					email
					headquarterAddress
				}
			}
		}
	`

	var result struct {
		Node struct {
			ID                 string `json:"id"`
			Name               string `json:"name"`
			Description        string `json:"description"`
			WebsiteUrl         string `json:"websiteUrl"`
			Email              string `json:"email"`
			HeadquarterAddress string `json:"headquarterAddress"`
		} `json:"node"`
	}

	err := owner.ExecuteConnect(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, owner.GetOrganizationID().String(), result.Node.ID)
	assert.NotEmpty(t, result.Node.Name)
}
