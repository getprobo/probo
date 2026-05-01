// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

func TestCookieCategory_Create(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)

		const query = `
			mutation CreateCookieCategory($input: CreateCookieCategoryInput!) {
				createCookieCategory(input: $input) {
					cookieCategoryEdge {
						node {
							id
							name
							slug
							description
							kind
							rank
							gcmConsentTypes
							posthogConsent
							createdAt
							updatedAt
						}
					}
					cookieBanner {
						id
					}
				}
			}
		`

		var result struct {
			CreateCookieCategory struct {
				CookieCategoryEdge struct {
					Node struct {
						ID              string   `json:"id"`
						Name            string   `json:"name"`
						Slug            string   `json:"slug"`
						Description     string   `json:"description"`
						Kind            string   `json:"kind"`
						Rank            int      `json:"rank"`
						GcmConsentTypes []string `json:"gcmConsentTypes"`
						PosthogConsent  bool     `json:"posthogConsent"`
						CreatedAt       string   `json:"createdAt"`
						UpdatedAt       string   `json:"updatedAt"`
					} `json:"node"`
				} `json:"cookieCategoryEdge"`
				CookieBanner struct {
					ID string `json:"id"`
				} `json:"cookieBanner"`
			} `json:"createCookieCategory"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookieBannerId": bannerID,
				"name":           "Marketing",
				"slug":           "marketing",
				"description":    "Marketing cookies for tracking",
				"rank":           5,
			},
		}, &result)

		require.NoError(t, err)
		node := result.CreateCookieCategory.CookieCategoryEdge.Node
		assert.NotEmpty(t, node.ID)
		assert.Equal(t, "Marketing", node.Name)
		assert.Equal(t, "marketing", node.Slug)
		assert.Equal(t, "Marketing cookies for tracking", node.Description)
		assert.Equal(t, "NORMAL", node.Kind)
		assert.Equal(t, 5, node.Rank)
		assert.Empty(t, node.GcmConsentTypes)
		assert.False(t, node.PosthogConsent)
		assert.Equal(t, bannerID, result.CreateCookieCategory.CookieBanner.ID)
	})

	t.Run("duplicate slug conflict", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"slug": "analytics"})

		_, err := owner.Do(`
			mutation CreateCookieCategory($input: CreateCookieCategoryInput!) {
				createCookieCategory(input: $input) {
					cookieCategoryEdge { node { id } }
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"cookieBannerId": bannerID,
				"name":           "Another Analytics",
				"slug":           "analytics",
				"description":    "Duplicate slug",
				"rank":           10,
			},
		})
		require.Error(t, err)
	})
}

func TestCookieCategory_Update(t *testing.T) {
	t.Parallel()

	t.Run("partial update", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{
			"name":        "Analytics",
			"slug":        "analytics-update",
			"description": "Original description",
		})

		const query = `
			mutation UpdateCookieCategory($input: UpdateCookieCategoryInput!) {
				updateCookieCategory(input: $input) {
					cookieCategory {
						id
						name
						description
						slug
					}
					cookieBanner {
						id
					}
				}
			}
		`

		var result struct {
			UpdateCookieCategory struct {
				CookieCategory struct {
					ID          string `json:"id"`
					Name        string `json:"name"`
					Description string `json:"description"`
					Slug        string `json:"slug"`
				} `json:"cookieCategory"`
				CookieBanner struct {
					ID string `json:"id"`
				} `json:"cookieBanner"`
			} `json:"updateCookieCategory"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookieCategoryId": categoryID,
				"name":             "Updated Analytics",
			},
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, categoryID, result.UpdateCookieCategory.CookieCategory.ID)
		assert.Equal(t, "Updated Analytics", result.UpdateCookieCategory.CookieCategory.Name)
		assert.Equal(t, "Original description", result.UpdateCookieCategory.CookieCategory.Description)
		assert.Equal(t, bannerID, result.UpdateCookieCategory.CookieBanner.ID)
	})

	t.Run("update gcmConsentTypes", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)

		const query = `
			mutation UpdateCookieCategory($input: UpdateCookieCategoryInput!) {
				updateCookieCategory(input: $input) {
					cookieCategory {
						id
						gcmConsentTypes
					}
				}
			}
		`

		var result struct {
			UpdateCookieCategory struct {
				CookieCategory struct {
					ID              string   `json:"id"`
					GcmConsentTypes []string `json:"gcmConsentTypes"`
				} `json:"cookieCategory"`
			} `json:"updateCookieCategory"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookieCategoryId": categoryID,
				"gcmConsentTypes":  []string{"ad_storage", "analytics_storage"},
			},
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, []string{"ad_storage", "analytics_storage"}, result.UpdateCookieCategory.CookieCategory.GcmConsentTypes)
	})

	t.Run("update posthogConsent", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)

		const query = `
			mutation UpdateCookieCategory($input: UpdateCookieCategoryInput!) {
				updateCookieCategory(input: $input) {
					cookieCategory {
						id
						posthogConsent
					}
				}
			}
		`

		var result struct {
			UpdateCookieCategory struct {
				CookieCategory struct {
					ID             string `json:"id"`
					PosthogConsent bool   `json:"posthogConsent"`
				} `json:"cookieCategory"`
			} `json:"updateCookieCategory"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookieCategoryId": categoryID,
				"posthogConsent":   true,
			},
		}, &result)

		require.NoError(t, err)
		assert.True(t, result.UpdateCookieCategory.CookieCategory.PosthogConsent)
	})
}

func TestCookieCategory_Delete(t *testing.T) {
	t.Parallel()

	t.Run("success for normal category", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)

		const query = `
			mutation DeleteCookieCategory($input: DeleteCookieCategoryInput!) {
				deleteCookieCategory(input: $input) {
					deletedCookieCategoryId
					cookieBanner {
						id
					}
				}
			}
		`

		var result struct {
			DeleteCookieCategory struct {
				DeletedCookieCategoryID string `json:"deletedCookieCategoryId"`
				CookieBanner            struct {
					ID string `json:"id"`
				} `json:"cookieBanner"`
			} `json:"deleteCookieCategory"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{"cookieCategoryId": categoryID},
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, categoryID, result.DeleteCookieCategory.DeletedCookieCategoryID)
		assert.Equal(t, bannerID, result.DeleteCookieCategory.CookieBanner.ID)
	})

	t.Run("cannot delete system category", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)

		// Fetch the NECESSARY category
		const listQuery = `
			query($id: ID!) {
				node(id: $id) {
					... on CookieBanner {
						categories(first: 20) {
							edges {
								node {
									id
									kind
								}
							}
						}
					}
				}
			}
		`

		var listResult struct {
			Node struct {
				Categories struct {
					Edges []struct {
						Node struct {
							ID   string `json:"id"`
							Kind string `json:"kind"`
						} `json:"node"`
					} `json:"edges"`
				} `json:"categories"`
			} `json:"node"`
		}

		err := owner.Execute(listQuery, map[string]any{"id": bannerID}, &listResult)
		require.NoError(t, err)

		var necessaryCategoryID string
		for _, e := range listResult.Node.Categories.Edges {
			if e.Node.Kind == "NECESSARY" {
				necessaryCategoryID = e.Node.ID
				break
			}
		}
		require.NotEmpty(t, necessaryCategoryID, "should find a NECESSARY category")

		_, err = owner.Do(`
			mutation DeleteCookieCategory($input: DeleteCookieCategoryInput!) {
				deleteCookieCategory(input: $input) {
					deletedCookieCategoryId
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{"cookieCategoryId": necessaryCategoryID},
		})
		require.Error(t, err)
	})
}

func TestCookieCategory_Reorder(t *testing.T) {
	t.Parallel()

	t.Run("change rank", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"rank": 5})

		const query = `
			mutation ReorderCookieCategory($input: ReorderCookieCategoryInput!) {
				reorderCookieCategory(input: $input) {
					cookieBanner {
						id
					}
				}
			}
		`

		var result struct {
			ReorderCookieCategory struct {
				CookieBanner struct {
					ID string `json:"id"`
				} `json:"cookieBanner"`
			} `json:"reorderCookieCategory"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookieCategoryId": categoryID,
				"rank":             1,
			},
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, bannerID, result.ReorderCookieCategory.CookieBanner.ID)
	})
}

func TestCookieCategory_List(t *testing.T) {
	t.Parallel()

	t.Run("via banner categories connection", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"rank": 20})
		factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"rank": 30})

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on CookieBanner {
						categories(first: 20, orderBy: {field: RANK, direction: ASC}) {
							totalCount
							edges {
								node {
									id
									rank
								}
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
				Categories struct {
					TotalCount int `json:"totalCount"`
					Edges      []struct {
						Node struct {
							ID   string `json:"id"`
							Rank int    `json:"rank"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						HasNextPage     bool `json:"hasNextPage"`
						HasPreviousPage bool `json:"hasPreviousPage"`
					} `json:"pageInfo"`
				} `json:"categories"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": bannerID}, &result)
		require.NoError(t, err)

		// Default categories + 2 custom ones
		assert.GreaterOrEqual(t, result.Node.Categories.TotalCount, 4)

		// Verify ordering (ranks should be ascending)
		for i := 1; i < len(result.Node.Categories.Edges); i++ {
			assert.GreaterOrEqual(t,
				result.Node.Categories.Edges[i].Node.Rank,
				result.Node.Categories.Edges[i-1].Node.Rank,
			)
		}
	})
}

func TestCookieCategory_RBAC(t *testing.T) {
	t.Parallel()

	t.Run("viewer cannot create category", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)

		_, err := viewer.Do(`
			mutation CreateCookieCategory($input: CreateCookieCategoryInput!) {
				createCookieCategory(input: $input) {
					cookieCategoryEdge { node { id } }
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"cookieBannerId": bannerID,
				"name":           "Test",
				"slug":           "test-rbac",
				"description":    "Test category",
				"rank":           10,
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to create cookie category")
	})

	t.Run("viewer cannot update category", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)

		_, err := viewer.Do(`
			mutation UpdateCookieCategory($input: UpdateCookieCategoryInput!) {
				updateCookieCategory(input: $input) {
					cookieCategory { id }
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"cookieCategoryId": categoryID,
				"name":             "Updated",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to update cookie category")
	})

	t.Run("viewer cannot delete category", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)

		_, err := viewer.Do(`
			mutation DeleteCookieCategory($input: DeleteCookieCategoryInput!) {
				deleteCookieCategory(input: $input) {
					deletedCookieCategoryId
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{"cookieCategoryId": categoryID},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to delete cookie category")
	})
}
