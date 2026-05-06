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

func TestTrackerPattern_Create(t *testing.T) {
	t.Parallel()

	t.Run("with EXACT match type", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)

		const query = `
			mutation CreateTrackerPattern($input: CreateTrackerPatternInput!) {
				createTrackerPattern(input: $input) {
					trackerPatternEdge {
						node {
							id
							pattern
							matchType
							trackerType
							displayName
							maxAgeSeconds
							description
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
			CreateTrackerPattern struct {
				TrackerPatternEdge struct {
					Node struct {
						ID            string `json:"id"`
						Pattern       string `json:"pattern"`
						MatchType     string `json:"matchType"`
						TrackerType   string `json:"trackerType"`
						DisplayName   string `json:"displayName"`
						MaxAgeSeconds *int   `json:"maxAgeSeconds"`
						Description   string `json:"description"`
						CreatedAt     string `json:"createdAt"`
						UpdatedAt     string `json:"updatedAt"`
					} `json:"node"`
				} `json:"trackerPatternEdge"`
				CookieBanner struct {
					ID string `json:"id"`
				} `json:"cookieBanner"`
			} `json:"createTrackerPattern"`
		}

		maxAge := 86400
		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookieCategoryId": categoryID,
				"pattern":          "_ga",
				"matchType":        "EXACT",
				"trackerType":      "SCRIPT",
				"displayName":      "Google Analytics",
				"maxAgeSeconds":    maxAge,
				"description":      "Google Analytics tracking cookie",
			},
		}, &result)

		require.NoError(t, err)
		node := result.CreateTrackerPattern.TrackerPatternEdge.Node
		assert.NotEmpty(t, node.ID)
		assert.Equal(t, "_ga", node.Pattern)
		assert.Equal(t, "EXACT", node.MatchType)
		assert.Equal(t, "SCRIPT", node.TrackerType)
		assert.Equal(t, "Google Analytics", node.DisplayName)
		require.NotNil(t, node.MaxAgeSeconds)
		assert.Equal(t, maxAge, *node.MaxAgeSeconds)
		assert.Equal(t, "Google Analytics tracking cookie", node.Description)
		assert.Equal(t, bannerID, result.CreateTrackerPattern.CookieBanner.ID)
	})

	t.Run("with PREFIX match type", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)

		const query = `
			mutation CreateTrackerPattern($input: CreateTrackerPatternInput!) {
				createTrackerPattern(input: $input) {
					trackerPatternEdge {
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
			CreateTrackerPattern struct {
				TrackerPatternEdge struct {
					Node struct {
						ID            string `json:"id"`
						Pattern       string `json:"pattern"`
						MatchType     string `json:"matchType"`
						DisplayName   string `json:"displayName"`
						MaxAgeSeconds *int   `json:"maxAgeSeconds"`
					} `json:"node"`
				} `json:"trackerPatternEdge"`
			} `json:"createTrackerPattern"`
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
		node := result.CreateTrackerPattern.TrackerPatternEdge.Node
		assert.Equal(t, "_gat_", node.Pattern)
		assert.Equal(t, "PREFIX", node.MatchType)
		assert.Nil(t, node.MaxAgeSeconds)
	})

	t.Run("duplicate pattern conflict", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)

		factory.CreateTrackerPattern(owner, categoryID, factory.Attrs{
			"pattern":     "duplicate_cookie",
			"displayName": "First",
		})

		_, err := owner.Do(`
			mutation CreateTrackerPattern($input: CreateTrackerPatternInput!) {
				createTrackerPattern(input: $input) {
					trackerPatternEdge { node { id } }
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

func TestTrackerPattern_Update(t *testing.T) {
	t.Parallel()

	t.Run("update displayName and description", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		patternID := factory.CreateTrackerPattern(owner, categoryID, factory.Attrs{
			"displayName": "Original Name",
			"description": "Original description",
		})

		const query = `
			mutation UpdateTrackerPattern($input: UpdateTrackerPatternInput!) {
				updateTrackerPattern(input: $input) {
					trackerPattern {
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
			UpdateTrackerPattern struct {
				TrackerPattern struct {
					ID          string `json:"id"`
					DisplayName string `json:"displayName"`
					Description string `json:"description"`
				} `json:"trackerPattern"`
				CookieBanner struct {
					ID string `json:"id"`
				} `json:"cookieBanner"`
			} `json:"updateTrackerPattern"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"trackerPatternId": patternID,
				"displayName":      "Updated Name",
				"description":      "Updated description",
			},
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, patternID, result.UpdateTrackerPattern.TrackerPattern.ID)
		assert.Equal(t, "Updated Name", result.UpdateTrackerPattern.TrackerPattern.DisplayName)
		assert.Equal(t, "Updated description", result.UpdateTrackerPattern.TrackerPattern.Description)
		assert.Equal(t, bannerID, result.UpdateTrackerPattern.CookieBanner.ID)
	})

	t.Run("update maxAgeSeconds", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		patternID := factory.CreateTrackerPattern(owner, categoryID)

		const query = `
			mutation UpdateTrackerPattern($input: UpdateTrackerPatternInput!) {
				updateTrackerPattern(input: $input) {
					trackerPattern {
						id
						maxAgeSeconds
					}
				}
			}
		`

		var result struct {
			UpdateTrackerPattern struct {
				TrackerPattern struct {
					ID            string `json:"id"`
					MaxAgeSeconds *int   `json:"maxAgeSeconds"`
				} `json:"trackerPattern"`
			} `json:"updateTrackerPattern"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"trackerPatternId": patternID,
				"maxAgeSeconds":    7200,
			},
		}, &result)

		require.NoError(t, err)
		require.NotNil(t, result.UpdateTrackerPattern.TrackerPattern.MaxAgeSeconds)
		assert.Equal(t, 7200, *result.UpdateTrackerPattern.TrackerPattern.MaxAgeSeconds)
	})
}

func TestTrackerPattern_Excluded(t *testing.T) {
	t.Parallel()

	t.Run("defaults to false on create", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)

		const query = `
			mutation CreateTrackerPattern($input: CreateTrackerPatternInput!) {
				createTrackerPattern(input: $input) {
					trackerPatternEdge {
						node {
							id
							excluded
						}
					}
				}
			}
		`

		var result struct {
			CreateTrackerPattern struct {
				TrackerPatternEdge struct {
					Node struct {
						ID       string `json:"id"`
						Excluded bool   `json:"excluded"`
					} `json:"node"`
				} `json:"trackerPatternEdge"`
			} `json:"createTrackerPattern"`
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
		assert.False(t, result.CreateTrackerPattern.TrackerPatternEdge.Node.Excluded)
	})

	t.Run("can be set to true via update", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		patternID := factory.CreateTrackerPattern(owner, categoryID)

		const query = `
			mutation UpdateTrackerPattern($input: UpdateTrackerPatternInput!) {
				updateTrackerPattern(input: $input) {
					trackerPattern {
						id
						excluded
					}
				}
			}
		`

		var result struct {
			UpdateTrackerPattern struct {
				TrackerPattern struct {
					ID       string `json:"id"`
					Excluded bool   `json:"excluded"`
				} `json:"trackerPattern"`
			} `json:"updateTrackerPattern"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"trackerPatternId": patternID,
				"excluded":         true,
			},
		}, &result)

		require.NoError(t, err)
		assert.True(t, result.UpdateTrackerPattern.TrackerPattern.Excluded)
	})

	t.Run("can be toggled back to false", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		patternID := factory.CreateTrackerPattern(owner, categoryID)

		const updateQuery = `
			mutation UpdateTrackerPattern($input: UpdateTrackerPatternInput!) {
				updateTrackerPattern(input: $input) {
					trackerPattern {
						id
						excluded
					}
				}
			}
		`

		var result struct {
			UpdateTrackerPattern struct {
				TrackerPattern struct {
					ID       string `json:"id"`
					Excluded bool   `json:"excluded"`
				} `json:"trackerPattern"`
			} `json:"updateTrackerPattern"`
		}

		err := owner.Execute(updateQuery, map[string]any{
			"input": map[string]any{
				"trackerPatternId": patternID,
				"excluded":         true,
			},
		}, &result)
		require.NoError(t, err)
		assert.True(t, result.UpdateTrackerPattern.TrackerPattern.Excluded)

		err = owner.Execute(updateQuery, map[string]any{
			"input": map[string]any{
				"trackerPatternId": patternID,
				"excluded":         false,
			},
		}, &result)
		require.NoError(t, err)
		assert.False(t, result.UpdateTrackerPattern.TrackerPattern.Excluded)
	})
}

func TestTrackerPattern_Delete(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		patternID := factory.CreateTrackerPattern(owner, categoryID)

		const query = `
			mutation DeleteTrackerPattern($input: DeleteTrackerPatternInput!) {
				deleteTrackerPattern(input: $input) {
					deletedTrackerPatternId
					cookieBanner {
						id
					}
				}
			}
		`

		var result struct {
			DeleteTrackerPattern struct {
				DeletedTrackerPatternID string `json:"deletedTrackerPatternId"`
				CookieBanner            struct {
					ID string `json:"id"`
				} `json:"cookieBanner"`
			} `json:"deleteTrackerPattern"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{"trackerPatternId": patternID},
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, patternID, result.DeleteTrackerPattern.DeletedTrackerPatternID)
		assert.Equal(t, bannerID, result.DeleteTrackerPattern.CookieBanner.ID)
	})
}

func TestTrackerPattern_MoveToCategory(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryA := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"slug": "cat-a-move"})
		categoryB := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"slug": "cat-b-move"})
		patternID := factory.CreateTrackerPattern(owner, categoryA)

		const query = `
			mutation MoveTrackerPatternToCategory($input: MoveTrackerPatternToCategoryInput!) {
				moveTrackerPatternToCategory(input: $input) {
					trackerPattern {
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
			MoveTrackerPatternToCategory struct {
				TrackerPattern struct {
					ID             string `json:"id"`
					CookieCategory struct {
						ID string `json:"id"`
					} `json:"cookieCategory"`
				} `json:"trackerPattern"`
				CookieBanner struct {
					ID string `json:"id"`
				} `json:"cookieBanner"`
			} `json:"moveTrackerPatternToCategory"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"trackerPatternId":       patternID,
				"targetCookieCategoryId": categoryB,
			},
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, patternID, result.MoveTrackerPatternToCategory.TrackerPattern.ID)
		assert.Equal(t, categoryB, result.MoveTrackerPatternToCategory.TrackerPattern.CookieCategory.ID)
		assert.Equal(t, bannerID, result.MoveTrackerPatternToCategory.CookieBanner.ID)
	})

	t.Run("cross-banner mismatch error", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		banner1 := factory.CreateCookieBanner(owner)
		banner2 := factory.CreateCookieBanner(owner)
		category1 := factory.CreateCookieCategory(owner, banner1, factory.Attrs{"slug": "cat-x-mismatch"})
		category2 := factory.CreateCookieCategory(owner, banner2, factory.Attrs{"slug": "cat-y-mismatch"})
		patternID := factory.CreateTrackerPattern(owner, category1)

		_, err := owner.Do(`
			mutation MoveTrackerPatternToCategory($input: MoveTrackerPatternToCategoryInput!) {
				moveTrackerPatternToCategory(input: $input) {
					trackerPattern { id }
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"trackerPatternId":       patternID,
				"targetCookieCategoryId": category2,
			},
		})
		require.Error(t, err)
	})
}

func TestTrackerPattern_List(t *testing.T) {
	t.Parallel()

	t.Run("via category trackerPatterns connection", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		factory.CreateTrackerPattern(owner, categoryID)
		factory.CreateTrackerPattern(owner, categoryID)

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on CookieCategory {
						trackerPatterns(first: 10) {
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
				TrackerPatterns struct {
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
				} `json:"trackerPatterns"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": categoryID}, &result)
		require.NoError(t, err)
		assert.Equal(t, 2, result.Node.TrackerPatterns.TotalCount)
		assert.Len(t, result.Node.TrackerPatterns.Edges, 2)
	})
}

func TestTrackerPattern_RBAC(t *testing.T) {
	t.Parallel()

	t.Run("viewer cannot create pattern", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)

		_, err := viewer.Do(`
			mutation CreateTrackerPattern($input: CreateTrackerPatternInput!) {
				createTrackerPattern(input: $input) {
					trackerPatternEdge { node { id } }
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
		testutil.RequireForbiddenError(t, err, "viewer should not be able to create tracker pattern")
	})

	t.Run("viewer cannot update pattern", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		patternID := factory.CreateTrackerPattern(owner, categoryID)

		_, err := viewer.Do(`
			mutation UpdateTrackerPattern($input: UpdateTrackerPatternInput!) {
				updateTrackerPattern(input: $input) {
					trackerPattern { id }
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"trackerPatternId": patternID,
				"displayName":      "Updated by Viewer",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to update tracker pattern")
	})

	t.Run("viewer cannot delete pattern", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		patternID := factory.CreateTrackerPattern(owner, categoryID)

		_, err := viewer.Do(`
			mutation DeleteTrackerPattern($input: DeleteTrackerPatternInput!) {
				deleteTrackerPattern(input: $input) {
					deletedTrackerPatternId
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{"trackerPatternId": patternID},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to delete tracker pattern")
	})

	t.Run("viewer cannot move pattern", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryA := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"slug": "rbac-move-a"})
		categoryB := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"slug": "rbac-move-b"})
		patternID := factory.CreateTrackerPattern(owner, categoryA)

		_, err := viewer.Do(`
			mutation MoveTrackerPatternToCategory($input: MoveTrackerPatternToCategoryInput!) {
				moveTrackerPatternToCategory(input: $input) {
					trackerPattern { id }
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"trackerPatternId":       patternID,
				"targetCookieCategoryId": categoryB,
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to move tracker pattern")
	})
}
