// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
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
							maxAgeSeconds
							description
							commonTrackerPatternId
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
						ID                     string  `json:"id"`
						Pattern                string  `json:"pattern"`
						MatchType              string  `json:"matchType"`
						TrackerType            string  `json:"trackerType"`
						MaxAgeSeconds          *int    `json:"maxAgeSeconds"`
						Description            string  `json:"description"`
						CommonTrackerPatternID *string `json:"commonTrackerPatternId"`
						CreatedAt              string  `json:"createdAt"`
						UpdatedAt              string  `json:"updatedAt"`
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
				"trackerType":      "LOCAL_STORAGE",
				"maxAgeSeconds":    maxAge,
				"description":      "Google Analytics tracking cookie",
			},
		}, &result)

		require.NoError(t, err)

		node := result.CreateTrackerPattern.TrackerPatternEdge.Node
		assert.NotEmpty(t, node.ID)
		assert.Equal(t, "_ga", node.Pattern)
		assert.Equal(t, "EXACT", node.MatchType)
		assert.Equal(t, "LOCAL_STORAGE", node.TrackerType)
		require.NotNil(t, node.MaxAgeSeconds)
		assert.Equal(t, maxAge, *node.MaxAgeSeconds)
		assert.Equal(t, "Google Analytics tracking cookie", node.Description)
		assert.Nil(t, node.CommonTrackerPatternID, "a manually created pattern is not linked to the common catalog")
		assert.Equal(t, bannerID, result.CreateTrackerPattern.CookieBanner.ID)
	})

	t.Run("with GLOB match type", func(t *testing.T) {
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
						MaxAgeSeconds *int   `json:"maxAgeSeconds"`
					} `json:"node"`
				} `json:"trackerPatternEdge"`
			} `json:"createTrackerPattern"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookieCategoryId": categoryID,
				"pattern":          "_gat_*",
				"matchType":        "GLOB",
				"description":      "Google Analytics rate limiting",
			},
		}, &result)

		require.NoError(t, err)

		node := result.CreateTrackerPattern.TrackerPatternEdge.Node
		assert.Equal(t, "_gat_*", node.Pattern)
		assert.Equal(t, "GLOB", node.MatchType)
		assert.Nil(t, node.MaxAgeSeconds)
	})

	t.Run("duplicate pattern conflict", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)

		factory.CreateTrackerPattern(owner, categoryID, factory.Attrs{
			"pattern": "duplicate_cookie",
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
				"description":      "Duplicate",
			},
		})
		require.Error(t, err)
	})
}

func TestTrackerPattern_Update(t *testing.T) {
	t.Parallel()

	t.Run("update description", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		patternID := factory.CreateTrackerPattern(owner, categoryID, factory.Attrs{
			"description": "Original description",
		})

		const query = `
			mutation UpdateTrackerPattern($input: UpdateTrackerPatternInput!) {
				updateTrackerPattern(input: $input) {
					trackerPattern {
						id
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
				"description":      "Updated description",
			},
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, patternID, result.UpdateTrackerPattern.TrackerPattern.ID)
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
							ID      string `json:"id"`
							Pattern string `json:"pattern"`
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

func TestTrackerPattern_Attribution(t *testing.T) {
	t.Parallel()

	t.Run("is null without a catalog link", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		patternID := factory.CreateTrackerPattern(owner, categoryID)

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on TrackerPattern {
						id
						attribution
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID          string  `json:"id"`
				Attribution *string `json:"attribution"`
			} `json:"node"`
		}

		require.NoError(t, owner.Execute(query, map[string]any{"id": patternID}, &result))
		assert.Nil(t, result.Node.Attribution, "a freshly created pattern has no catalog verdict")
	})

	t.Run("inherits the catalog verdict", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		patternID := factory.CreateTrackerPattern(owner, categoryID)
		commonID := seedCommonTrackerPatternWithAttribution(t, coredata.CommonTrackerPatternAttributionFirstParty)
		linkTrackerPatternToCommon(t, patternID, commonID)

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on TrackerPattern {
						id
						attribution
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID          string  `json:"id"`
				Attribution *string `json:"attribution"`
			} `json:"node"`
		}

		require.NoError(t, owner.Execute(query, map[string]any{"id": patternID}, &result))
		require.NotNil(t, result.Node.Attribution, "a catalog-linked pattern must expose the inherited verdict")
		assert.Equal(t, string(coredata.CommonTrackerPatternAttributionFirstParty), *result.Node.Attribution)
	})
}

func TestTrackerPattern_CommonTrackerPatternID(t *testing.T) {
	t.Parallel()

	t.Run("reflects the common catalog link", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		patternID := factory.CreateTrackerPattern(owner, categoryID)

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on TrackerPattern {
						id
						commonTrackerPatternId
					}
				}
			}
		`

		var result struct {
			Node struct {
				ID                     string  `json:"id"`
				CommonTrackerPatternID *string `json:"commonTrackerPatternId"`
			} `json:"node"`
		}

		require.NoError(t, owner.Execute(query, map[string]any{"id": patternID}, &result))
		assert.Nil(t, result.Node.CommonTrackerPatternID, "a freshly created pattern has no catalog link")

		commonID := seedCommonTrackerPattern(t)
		linkTrackerPatternToCommon(t, patternID, commonID)

		require.NoError(t, owner.Execute(query, map[string]any{"id": patternID}, &result))
		require.NotNil(t, result.Node.CommonTrackerPatternID, "the catalog link must surface once set")
		assert.Equal(t, commonID.String(), *result.Node.CommonTrackerPatternID)
	})
}

func TestTrackerPattern_SampleHits(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	bannerID := factory.CreateCookieBanner(owner)
	categoryID := factory.CreateCookieCategory(owner, bannerID)
	patternID := factory.CreateTrackerPattern(owner, categoryID)

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on TrackerPattern {
					sampleHits {
						family
						count
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			SampleHits []struct {
				Family string `json:"family"`
				Count  int    `json:"count"`
			} `json:"sampleHits"`
		} `json:"node"`
	}

	require.NoError(t, owner.Execute(query, map[string]any{"id": patternID}, &result))
	assert.Empty(t, result.Node.SampleHits)
}

func TestTrackerPattern_ObservedBy(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	bannerID := factory.CreateCookieBanner(owner)
	categoryID := factory.CreateCookieCategory(owner, bannerID)

	bothID := factory.CreateTrackerPattern(owner, categoryID, factory.Attrs{"pattern": "both"})
	scanID := factory.CreateTrackerPattern(owner, categoryID, factory.Attrs{"pattern": "scan"})
	sdkID := factory.CreateTrackerPattern(owner, categoryID, factory.Attrs{"pattern": "sdk"})
	factory.CreateTrackerPattern(owner, categoryID, factory.Attrs{"pattern": "unseen"})

	scanCookie := &detectedTrackerCookie{
		valueSize:   24,
		httpOnly:    true,
		secure:      true,
		sameSite:    "LAX",
		partitioned: true,
		setByURL:    "https://tracker.example/set",
	}

	seedDetectedTracker(t, bannerID, bothID, "both", coredata.DetectedTrackerObserverSDK, nil)
	seedDetectedTracker(t, bannerID, bothID, "both", coredata.DetectedTrackerObserverScanner, scanCookie)
	seedDetectedTracker(t, bannerID, scanID, "scan", coredata.DetectedTrackerObserverScanner, scanCookie)
	seedDetectedTracker(t, bannerID, sdkID, "sdk", coredata.DetectedTrackerObserverSDK, nil)

	const listQuery = `
		query($id: ID!, $filter: TrackerPatternFilter) {
			node(id: $id) {
				... on CookieBanner {
					trackerPatterns(
						first: 10
						filter: $filter
						orderBy: { field: NAME, direction: ASC }
					) {
						edges {
							node {
								pattern
								observedBy
							}
						}
					}
				}
			}
		}
	`

	type listedPattern struct {
		Pattern    string   `json:"pattern"`
		ObservedBy []string `json:"observedBy"`
	}

	var listed struct {
		Node struct {
			TrackerPatterns struct {
				Edges []struct {
					Node listedPattern `json:"node"`
				} `json:"edges"`
			} `json:"trackerPatterns"`
		} `json:"node"`
	}

	load := func(t *testing.T, observedBy []string) []listedPattern {
		t.Helper()

		require.NoError(t, owner.Execute(
			listQuery,
			map[string]any{
				"id": bannerID,
				"filter": map[string]any{
					"observedBy": observedBy,
				},
			},
			&listed,
		))

		patterns := make([]listedPattern, 0, len(listed.Node.TrackerPatterns.Edges))
		for _, edge := range listed.Node.TrackerPatterns.Edges {
			patterns = append(patterns, edge.Node)
		}

		return patterns
	}

	assert.Equal(t, []listedPattern{
		{Pattern: "both", ObservedBy: []string{"SCANNER", "SDK"}},
		{Pattern: "sdk", ObservedBy: []string{"SDK"}},
	}, load(t, []string{"SDK"}))
	assert.Equal(t, []listedPattern{
		{Pattern: "both", ObservedBy: []string{"SCANNER", "SDK"}},
		{Pattern: "scan", ObservedBy: []string{"SCANNER"}},
	}, load(t, []string{"SCANNER"}))
	assert.Equal(t, []listedPattern{
		{Pattern: "both", ObservedBy: []string{"SCANNER", "SDK"}},
		{Pattern: "scan", ObservedBy: []string{"SCANNER"}},
		{Pattern: "sdk", ObservedBy: []string{"SDK"}},
	}, load(t, []string{"SDK", "SCANNER"}))

	const detectionQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on TrackerPattern {
					detectedTrackers(first: 10) {
						edges {
							node {
								identifier
								observedBy
								valueSize
								httpOnly
								secure
								sameSite
								partitioned
								setByUrl
							}
						}
					}
				}
			}
		}
	`

	var detection struct {
		Node struct {
			DetectedTrackers struct {
				Edges []struct {
					Node struct {
						Identifier  string  `json:"identifier"`
						ObservedBy  string  `json:"observedBy"`
						ValueSize   *int    `json:"valueSize"`
						HTTPOnly    *bool   `json:"httpOnly"`
						Secure      *bool   `json:"secure"`
						SameSite    *string `json:"sameSite"`
						Partitioned *bool   `json:"partitioned"`
						SetByURL    *string `json:"setByUrl"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"detectedTrackers"`
		} `json:"node"`
	}

	require.NoError(t, owner.Execute(detectionQuery, map[string]any{"id": scanID}, &detection))
	require.Len(t, detection.Node.DetectedTrackers.Edges, 1)

	got := detection.Node.DetectedTrackers.Edges[0].Node
	assert.Equal(t, "scan", got.Identifier)
	assert.Equal(t, "SCANNER", got.ObservedBy)
	require.NotNil(t, got.ValueSize)
	assert.Equal(t, 24, *got.ValueSize)
	require.NotNil(t, got.HTTPOnly)
	assert.True(t, *got.HTTPOnly)
	require.NotNil(t, got.Secure)
	assert.True(t, *got.Secure)
	require.NotNil(t, got.SameSite)
	assert.Equal(t, "LAX", *got.SameSite)
	require.NotNil(t, got.Partitioned)
	assert.True(t, *got.Partitioned)
	require.NotNil(t, got.SetByURL)
	assert.Equal(t, "https://tracker.example/set", *got.SetByURL)
}

type detectedTrackerCookie struct {
	valueSize   int
	httpOnly    bool
	secure      bool
	sameSite    string
	partitioned bool
	setByURL    string
}

func seedDetectedTracker(
	t *testing.T,
	bannerID string,
	patternID string,
	identifier string,
	observedBy coredata.DetectedTrackerObserver,
	cookie *detectedTrackerCookie,
) {
	t.Helper()

	patternGID, err := gid.ParseGID(patternID)
	require.NoError(t, err)

	ctx := context.Background()
	conn := dialTestPg(t, ctx)
	t.Cleanup(func() { _ = conn.Close(ctx) })

	id := gid.New(patternGID.TenantID(), coredata.DetectedTrackerEntityType)
	now := time.Now().UTC()

	var (
		valueSize   *int
		httpOnly    *bool
		secure      *bool
		sameSite    *string
		partitioned *bool
		setByURL    *string
		source      = string(coredata.CookieSourceScript)
	)

	if cookie != nil {
		source = string(coredata.CookieSourceHTTP)
		valueSize = &cookie.valueSize
		httpOnly = &cookie.httpOnly
		secure = &cookie.secure
		sameSite = &cookie.sameSite
		partitioned = &cookie.partitioned
		setByURL = &cookie.setByURL
	}

	_, err = conn.Exec(
		ctx,
		`
		INSERT INTO detected_trackers (
			id, tenant_id, cookie_banner_id, tracker_pattern_id, tracker_type,
			identifier, source, observed_by, value_size, http_only, secure,
			same_site, partitioned, set_by_url, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, 'COOKIE',
			$5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15
		)
	`,
		id,
		patternGID.TenantID(),
		bannerID,
		patternID,
		identifier,
		source,
		observedBy,
		valueSize,
		httpOnly,
		secure,
		sameSite,
		partitioned,
		setByURL,
		now,
		now,
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cleanupConn := dialTestPg(t, cleanupCtx)

		defer func() { _ = cleanupConn.Close(cleanupCtx) }()

		_, err := cleanupConn.Exec(cleanupCtx, `DELETE FROM detected_trackers WHERE id = $1`, id)
		assert.NoError(t, err, "cleanup: cannot delete seeded detected tracker %s", id)
	})
}

func seedCommonTrackerPattern(t *testing.T) gid.GID {
	t.Helper()

	return seedCommonTrackerPatternWithAttribution(t, coredata.CommonTrackerPatternAttributionUndetermined)
}

func seedCommonTrackerPatternWithAttribution(
	t *testing.T,
	attribution coredata.CommonTrackerPatternAttribution,
) gid.GID {
	t.Helper()

	ctx := context.Background()
	conn := dialTestPg(t, ctx)
	t.Cleanup(func() { _ = conn.Close(ctx) })

	id := gid.New(gid.NilTenant, coredata.CommonTrackerPatternEntityType)
	now := time.Now().UTC()

	_, err := conn.Exec(ctx, `
		INSERT INTO common_tracker_patterns (
			id, tracker_type, pattern, match_type, description, confidence, attribution, enrichment_attempts, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
	`, id, "COOKIE", "e2e_common_"+id.String(), "EXACT", "Seeded catalog description", 1.0, attribution, 0, now, now)
	require.NoError(t, err)

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cleanupConn := dialTestPg(t, cleanupCtx)

		defer func() { _ = cleanupConn.Close(cleanupCtx) }()

		_, err := cleanupConn.Exec(cleanupCtx, `DELETE FROM common_tracker_patterns WHERE id = $1`, id)
		assert.NoError(t, err, "cleanup: cannot delete seeded common tracker pattern %s", id)
	})

	return id
}

func linkTrackerPatternToCommon(t *testing.T, patternID string, commonID gid.GID) {
	t.Helper()

	ctx := context.Background()
	conn := dialTestPg(t, ctx)
	t.Cleanup(func() { _ = conn.Close(ctx) })

	_, err := conn.Exec(ctx, `
		UPDATE tracker_patterns SET common_tracker_pattern_id = $1 WHERE id = $2
	`, commonID, patternID)
	require.NoError(t, err)
}
