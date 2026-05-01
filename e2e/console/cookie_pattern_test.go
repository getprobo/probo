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

func TestCookiePattern_Create(t *testing.T) {
	t.Parallel()

	t.Run("with EXACT match type", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)

		const query = `
			mutation CreateCookiePattern($input: CreateCookiePatternInput!) {
				createCookiePattern(input: $input) {
					cookiePatternEdge {
						node {
							id
							pattern
							matchType
							displayName
							maxAgeSeconds
							description
							source
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
			CreateCookiePattern struct {
				CookiePatternEdge struct {
					Node struct {
						ID            string `json:"id"`
						Pattern       string `json:"pattern"`
						MatchType     string `json:"matchType"`
						DisplayName   string `json:"displayName"`
						MaxAgeSeconds *int   `json:"maxAgeSeconds"`
						Description   string `json:"description"`
						Source        string `json:"source"`
						CreatedAt     string `json:"createdAt"`
						UpdatedAt     string `json:"updatedAt"`
					} `json:"node"`
				} `json:"cookiePatternEdge"`
				CookieBanner struct {
					ID string `json:"id"`
				} `json:"cookieBanner"`
			} `json:"createCookiePattern"`
		}

		maxAge := 86400
		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookieCategoryId": categoryID,
				"pattern":          "_ga",
				"matchType":        "EXACT",
				"displayName":      "Google Analytics",
				"maxAgeSeconds":    maxAge,
				"description":      "Google Analytics tracking cookie",
			},
		}, &result)

		require.NoError(t, err)
		node := result.CreateCookiePattern.CookiePatternEdge.Node
		assert.NotEmpty(t, node.ID)
		assert.Equal(t, "_ga", node.Pattern)
		assert.Equal(t, "EXACT", node.MatchType)
		assert.Equal(t, "Google Analytics", node.DisplayName)
		require.NotNil(t, node.MaxAgeSeconds)
		assert.Equal(t, maxAge, *node.MaxAgeSeconds)
		assert.Equal(t, "Google Analytics tracking cookie", node.Description)
		assert.Equal(t, "SCRIPT", node.Source)
		assert.Equal(t, bannerID, result.CreateCookiePattern.CookieBanner.ID)
	})

	t.Run("with PREFIX match type", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)

		const query = `
			mutation CreateCookiePattern($input: CreateCookiePatternInput!) {
				createCookiePattern(input: $input) {
					cookiePatternEdge {
						node {
							id
							pattern
							matchType
							displayName
							maxAgeSeconds
						}
					}
				}
			}
		`

		var result struct {
			CreateCookiePattern struct {
				CookiePatternEdge struct {
					Node struct {
						ID            string `json:"id"`
						Pattern       string `json:"pattern"`
						MatchType     string `json:"matchType"`
						DisplayName   string `json:"displayName"`
						MaxAgeSeconds *int   `json:"maxAgeSeconds"`
					} `json:"node"`
				} `json:"cookiePatternEdge"`
			} `json:"createCookiePattern"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookieCategoryId": categoryID,
				"pattern":          "_gat_",
				"matchType":        "PREFIX",
				"displayName":      "GA Throttle",
				"description":      "Google Analytics rate limiting",
			},
		}, &result)

		require.NoError(t, err)
		node := result.CreateCookiePattern.CookiePatternEdge.Node
		assert.Equal(t, "_gat_", node.Pattern)
		assert.Equal(t, "PREFIX", node.MatchType)
		assert.Nil(t, node.MaxAgeSeconds)
	})

	t.Run("duplicate pattern conflict", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)

		factory.CreateCookiePattern(owner, categoryID, factory.Attrs{
			"pattern":     "duplicate_cookie",
			"displayName": "First",
		})

		_, err := owner.Do(`
			mutation CreateCookiePattern($input: CreateCookiePatternInput!) {
				createCookiePattern(input: $input) {
					cookiePatternEdge { node { id } }
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"cookieCategoryId": categoryID,
				"pattern":          "duplicate_cookie",
				"matchType":        "EXACT",
				"displayName":      "Second",
				"description":      "Duplicate",
			},
		})
		require.Error(t, err)
	})
}

func TestCookiePattern_Update(t *testing.T) {
	t.Parallel()

	t.Run("update displayName and description", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		patternID := factory.CreateCookiePattern(owner, categoryID, factory.Attrs{
			"displayName": "Original Name",
			"description": "Original description",
		})

		const query = `
			mutation UpdateCookiePattern($input: UpdateCookiePatternInput!) {
				updateCookiePattern(input: $input) {
					cookiePattern {
						id
						displayName
						description
					}
					cookieBanner {
						id
					}
				}
			}
		`

		var result struct {
			UpdateCookiePattern struct {
				CookiePattern struct {
					ID          string `json:"id"`
					DisplayName string `json:"displayName"`
					Description string `json:"description"`
				} `json:"cookiePattern"`
				CookieBanner struct {
					ID string `json:"id"`
				} `json:"cookieBanner"`
			} `json:"updateCookiePattern"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookiePatternId": patternID,
				"displayName":     "Updated Name",
				"description":     "Updated description",
			},
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, patternID, result.UpdateCookiePattern.CookiePattern.ID)
		assert.Equal(t, "Updated Name", result.UpdateCookiePattern.CookiePattern.DisplayName)
		assert.Equal(t, "Updated description", result.UpdateCookiePattern.CookiePattern.Description)
		assert.Equal(t, bannerID, result.UpdateCookiePattern.CookieBanner.ID)
	})

	t.Run("update maxAgeSeconds", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		patternID := factory.CreateCookiePattern(owner, categoryID)

		const query = `
			mutation UpdateCookiePattern($input: UpdateCookiePatternInput!) {
				updateCookiePattern(input: $input) {
					cookiePattern {
						id
						maxAgeSeconds
					}
				}
			}
		`

		var result struct {
			UpdateCookiePattern struct {
				CookiePattern struct {
					ID            string `json:"id"`
					MaxAgeSeconds *int   `json:"maxAgeSeconds"`
				} `json:"cookiePattern"`
			} `json:"updateCookiePattern"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookiePatternId": patternID,
				"maxAgeSeconds":   7200,
			},
		}, &result)

		require.NoError(t, err)
		require.NotNil(t, result.UpdateCookiePattern.CookiePattern.MaxAgeSeconds)
		assert.Equal(t, 7200, *result.UpdateCookiePattern.CookiePattern.MaxAgeSeconds)
	})
}

func TestCookiePattern_Excluded(t *testing.T) {
	t.Parallel()

	t.Run("defaults to false on create", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)

		const query = `
			mutation CreateCookiePattern($input: CreateCookiePatternInput!) {
				createCookiePattern(input: $input) {
					cookiePatternEdge {
						node {
							id
							excluded
						}
					}
				}
			}
		`

		var result struct {
			CreateCookiePattern struct {
				CookiePatternEdge struct {
					Node struct {
						ID       string `json:"id"`
						Excluded bool   `json:"excluded"`
					} `json:"node"`
				} `json:"cookiePatternEdge"`
			} `json:"createCookiePattern"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookieCategoryId": categoryID,
				"pattern":          "test_excluded_default",
				"matchType":        "EXACT",
				"displayName":      "Test Excluded Default",
				"description":      "Should default to not excluded",
			},
		}, &result)

		require.NoError(t, err)
		assert.False(t, result.CreateCookiePattern.CookiePatternEdge.Node.Excluded)
	})

	t.Run("can be set to true via update", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		patternID := factory.CreateCookiePattern(owner, categoryID)

		const query = `
			mutation UpdateCookiePattern($input: UpdateCookiePatternInput!) {
				updateCookiePattern(input: $input) {
					cookiePattern {
						id
						excluded
					}
				}
			}
		`

		var result struct {
			UpdateCookiePattern struct {
				CookiePattern struct {
					ID       string `json:"id"`
					Excluded bool   `json:"excluded"`
				} `json:"cookiePattern"`
			} `json:"updateCookiePattern"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookiePatternId": patternID,
				"excluded":        true,
			},
		}, &result)

		require.NoError(t, err)
		assert.True(t, result.UpdateCookiePattern.CookiePattern.Excluded)
	})

	t.Run("can be toggled back to false", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		patternID := factory.CreateCookiePattern(owner, categoryID)

		const updateQuery = `
			mutation UpdateCookiePattern($input: UpdateCookiePatternInput!) {
				updateCookiePattern(input: $input) {
					cookiePattern {
						id
						excluded
					}
				}
			}
		`

		var result struct {
			UpdateCookiePattern struct {
				CookiePattern struct {
					ID       string `json:"id"`
					Excluded bool   `json:"excluded"`
				} `json:"cookiePattern"`
			} `json:"updateCookiePattern"`
		}

		err := owner.Execute(updateQuery, map[string]any{
			"input": map[string]any{
				"cookiePatternId": patternID,
				"excluded":        true,
			},
		}, &result)
		require.NoError(t, err)
		assert.True(t, result.UpdateCookiePattern.CookiePattern.Excluded)

		err = owner.Execute(updateQuery, map[string]any{
			"input": map[string]any{
				"cookiePatternId": patternID,
				"excluded":        false,
			},
		}, &result)
		require.NoError(t, err)
		assert.False(t, result.UpdateCookiePattern.CookiePattern.Excluded)
	})
}

func TestCookiePattern_Delete(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		patternID := factory.CreateCookiePattern(owner, categoryID)

		const query = `
			mutation DeleteCookiePattern($input: DeleteCookiePatternInput!) {
				deleteCookiePattern(input: $input) {
					deletedCookiePatternId
					cookieBanner {
						id
					}
				}
			}
		`

		var result struct {
			DeleteCookiePattern struct {
				DeletedCookiePatternID string `json:"deletedCookiePatternId"`
				CookieBanner           struct {
					ID string `json:"id"`
				} `json:"cookieBanner"`
			} `json:"deleteCookiePattern"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{"cookiePatternId": patternID},
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, patternID, result.DeleteCookiePattern.DeletedCookiePatternID)
		assert.Equal(t, bannerID, result.DeleteCookiePattern.CookieBanner.ID)
	})
}

func TestCookiePattern_MoveToCategory(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryA := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"slug": "cat-a-move"})
		categoryB := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"slug": "cat-b-move"})
		patternID := factory.CreateCookiePattern(owner, categoryA)

		const query = `
			mutation MoveCookiePatternToCategory($input: MoveCookiePatternToCategoryInput!) {
				moveCookiePatternToCategory(input: $input) {
					cookiePattern {
						id
						cookieCategory {
							id
						}
					}
					cookieBanner {
						id
					}
				}
			}
		`

		var result struct {
			MoveCookiePatternToCategory struct {
				CookiePattern struct {
					ID             string `json:"id"`
					CookieCategory struct {
						ID string `json:"id"`
					} `json:"cookieCategory"`
				} `json:"cookiePattern"`
				CookieBanner struct {
					ID string `json:"id"`
				} `json:"cookieBanner"`
			} `json:"moveCookiePatternToCategory"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookiePatternId":        patternID,
				"targetCookieCategoryId": categoryB,
			},
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, patternID, result.MoveCookiePatternToCategory.CookiePattern.ID)
		assert.Equal(t, categoryB, result.MoveCookiePatternToCategory.CookiePattern.CookieCategory.ID)
		assert.Equal(t, bannerID, result.MoveCookiePatternToCategory.CookieBanner.ID)
	})

	t.Run("cross-banner mismatch error", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		banner1 := factory.CreateCookieBanner(owner)
		banner2 := factory.CreateCookieBanner(owner)
		category1 := factory.CreateCookieCategory(owner, banner1, factory.Attrs{"slug": "cat-x-mismatch"})
		category2 := factory.CreateCookieCategory(owner, banner2, factory.Attrs{"slug": "cat-y-mismatch"})
		patternID := factory.CreateCookiePattern(owner, category1)

		_, err := owner.Do(`
			mutation MoveCookiePatternToCategory($input: MoveCookiePatternToCategoryInput!) {
				moveCookiePatternToCategory(input: $input) {
					cookiePattern { id }
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"cookiePatternId":        patternID,
				"targetCookieCategoryId": category2,
			},
		})
		require.Error(t, err)
	})
}

func TestCookiePattern_List(t *testing.T) {
	t.Parallel()

	t.Run("via category cookiePatterns connection", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		factory.CreateCookiePattern(owner, categoryID)
		factory.CreateCookiePattern(owner, categoryID)

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on CookieCategory {
						cookiePatterns(first: 10) {
							totalCount
							edges {
								node {
									id
									pattern
									displayName
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
				CookiePatterns struct {
					TotalCount int `json:"totalCount"`
					Edges      []struct {
						Node struct {
							ID          string `json:"id"`
							Pattern     string `json:"pattern"`
							DisplayName string `json:"displayName"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						HasNextPage     bool `json:"hasNextPage"`
						HasPreviousPage bool `json:"hasPreviousPage"`
					} `json:"pageInfo"`
				} `json:"cookiePatterns"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": categoryID}, &result)
		require.NoError(t, err)
		assert.Equal(t, 2, result.Node.CookiePatterns.TotalCount)
		assert.Len(t, result.Node.CookiePatterns.Edges, 2)
	})
}

func TestCookiePattern_RBAC(t *testing.T) {
	t.Parallel()

	t.Run("viewer cannot create pattern", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)

		_, err := viewer.Do(`
			mutation CreateCookiePattern($input: CreateCookiePatternInput!) {
				createCookiePattern(input: $input) {
					cookiePatternEdge { node { id } }
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"cookieCategoryId": categoryID,
				"pattern":          "test_viewer",
				"matchType":        "EXACT",
				"displayName":      "Test Viewer Pattern",
				"description":      "Should fail",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to create cookie pattern")
	})

	t.Run("viewer cannot update pattern", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		patternID := factory.CreateCookiePattern(owner, categoryID)

		_, err := viewer.Do(`
			mutation UpdateCookiePattern($input: UpdateCookiePatternInput!) {
				updateCookiePattern(input: $input) {
					cookiePattern { id }
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"cookiePatternId": patternID,
				"displayName":     "Updated by Viewer",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to update cookie pattern")
	})

	t.Run("viewer cannot delete pattern", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		patternID := factory.CreateCookiePattern(owner, categoryID)

		_, err := viewer.Do(`
			mutation DeleteCookiePattern($input: DeleteCookiePatternInput!) {
				deleteCookiePattern(input: $input) {
					deletedCookiePatternId
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{"cookiePatternId": patternID},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to delete cookie pattern")
	})

	t.Run("viewer cannot move pattern", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryA := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"slug": "rbac-move-a"})
		categoryB := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"slug": "rbac-move-b"})
		patternID := factory.CreateCookiePattern(owner, categoryA)

		_, err := viewer.Do(`
			mutation MoveCookiePatternToCategory($input: MoveCookiePatternToCategoryInput!) {
				moveCookiePatternToCategory(input: $input) {
					cookiePattern { id }
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"cookiePatternId":        patternID,
				"targetCookieCategoryId": categoryB,
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to move cookie pattern")
	})
}
