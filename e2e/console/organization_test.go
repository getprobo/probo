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
					product
					architecture
					team
					processes
					customers
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
				Processes      *string `json:"processes"`
				Customers      *string `json:"customers"`
			} `json:"context"`
		} `json:"updateOrganizationContext"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"product":        "Our product provides compliance solutions.",
			"architecture":   "Microservices architecture on AWS.",
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
