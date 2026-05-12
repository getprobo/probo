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

func TestTrackerResource_Create(t *testing.T) {
	t.Parallel()

	t.Run("with all fields", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)

		const query = `
			mutation CreateTrackerResource($input: CreateTrackerResourceInput!) {
				createTrackerResource(input: $input) {
					trackerResourceEdge {
						node {
							id
							type
							origin
							path
							displayName
							description
							excluded
						}
					}
					cookieBanner {
						id
					}
				}
			}
		`

		var result struct {
			CreateTrackerResource struct {
				TrackerResourceEdge struct {
					Node struct {
						ID          string `json:"id"`
						Type        string `json:"type"`
						Origin      string `json:"origin"`
						Path        string `json:"path"`
						DisplayName string `json:"displayName"`
						Description string `json:"description"`
						Excluded    bool   `json:"excluded"`
					} `json:"node"`
				} `json:"trackerResourceEdge"`
				CookieBanner struct {
					ID string `json:"id"`
				} `json:"cookieBanner"`
			} `json:"createTrackerResource"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"cookieCategoryId": categoryID,
				"type":             "SCRIPT",
				"origin":           "https://example.com",
				"path":             "/script.js",
				"displayName":      "Example Script",
				"description":      "Test",
			},
		}, &result)

		require.NoError(t, err)
		node := result.CreateTrackerResource.TrackerResourceEdge.Node
		assert.NotEmpty(t, node.ID)
		assert.Equal(t, "SCRIPT", node.Type)
		assert.Equal(t, "https://example.com", node.Origin)
		assert.Equal(t, "/script.js", node.Path)
		assert.Equal(t, "Example Script", node.DisplayName)
		assert.Equal(t, "Test", node.Description)
		assert.False(t, node.Excluded)
		assert.Equal(t, bannerID, result.CreateTrackerResource.CookieBanner.ID)
	})
}

func TestTrackerResource_List(t *testing.T) {
	t.Parallel()

	t.Run("via uncategorisedTrackerResources connection", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		factory.ReportDetectedResources(owner, bannerID, 2)

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on CookieBanner {
						uncategorisedTrackerResources(first: 10) {
							totalCount
							edges {
								node {
									id
									type
									origin
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
				UncategorisedTrackerResources struct {
					TotalCount int `json:"totalCount"`
					Edges      []struct {
						Node struct {
							ID          string `json:"id"`
							Type        string `json:"type"`
							Origin      string `json:"origin"`
							DisplayName string `json:"displayName"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						HasNextPage     bool `json:"hasNextPage"`
						HasPreviousPage bool `json:"hasPreviousPage"`
					} `json:"pageInfo"`
				} `json:"uncategorisedTrackerResources"`
			} `json:"node"`
		}

		err := owner.Execute(query, map[string]any{"id": bannerID}, &result)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Node.UncategorisedTrackerResources.TotalCount, 2)
		assert.GreaterOrEqual(t, len(result.Node.UncategorisedTrackerResources.Edges), 2)
	})
}

func TestTrackerResource_Update(t *testing.T) {
	t.Parallel()

	t.Run("update displayName, description, and excluded", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		resourceID := factory.CreateTrackerResource(owner, categoryID, factory.Attrs{
			"displayName": "Original Name",
			"description": "Original description",
		})

		const query = `
			mutation UpdateTrackerResource($input: UpdateTrackerResourceInput!) {
				updateTrackerResource(input: $input) {
					trackerResource {
						id
						displayName
						description
						excluded
					}
					cookieBanner {
						id
					}
				}
			}
		`

		var result struct {
			UpdateTrackerResource struct {
				TrackerResource struct {
					ID          string `json:"id"`
					DisplayName string `json:"displayName"`
					Description string `json:"description"`
					Excluded    bool   `json:"excluded"`
				} `json:"trackerResource"`
				CookieBanner struct {
					ID string `json:"id"`
				} `json:"cookieBanner"`
			} `json:"updateTrackerResource"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"trackerResourceId": resourceID,
				"displayName":       "Updated Name",
				"description":       "Updated description",
				"excluded":          true,
			},
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, resourceID, result.UpdateTrackerResource.TrackerResource.ID)
		assert.Equal(t, "Updated Name", result.UpdateTrackerResource.TrackerResource.DisplayName)
		assert.Equal(t, "Updated description", result.UpdateTrackerResource.TrackerResource.Description)
		assert.True(t, result.UpdateTrackerResource.TrackerResource.Excluded)
		assert.Equal(t, bannerID, result.UpdateTrackerResource.CookieBanner.ID)
	})
}

func TestTrackerResource_MoveToCategory(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryA := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"slug": "res-cat-a-move"})
		categoryB := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"slug": "res-cat-b-move"})
		resourceID := factory.CreateTrackerResource(owner, categoryA)

		const query = `
			mutation MoveTrackerResourceToCategory($input: MoveTrackerResourceToCategoryInput!) {
				moveTrackerResourceToCategory(input: $input) {
					trackerResource {
						id
					}
					cookieBanner {
						id
					}
				}
			}
		`

		var result struct {
			MoveTrackerResourceToCategory struct {
				TrackerResource struct {
					ID string `json:"id"`
				} `json:"trackerResource"`
				CookieBanner struct {
					ID string `json:"id"`
				} `json:"cookieBanner"`
			} `json:"moveTrackerResourceToCategory"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{
				"trackerResourceId":      resourceID,
				"targetCookieCategoryId": categoryB,
			},
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, resourceID, result.MoveTrackerResourceToCategory.TrackerResource.ID)
		assert.Equal(t, bannerID, result.MoveTrackerResourceToCategory.CookieBanner.ID)
	})

	t.Run("cross-banner mismatch error", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		banner1 := factory.CreateCookieBanner(owner)
		banner2 := factory.CreateCookieBanner(owner)
		category1 := factory.CreateCookieCategory(owner, banner1, factory.Attrs{"slug": "res-cat-x-mismatch"})
		category2 := factory.CreateCookieCategory(owner, banner2, factory.Attrs{"slug": "res-cat-y-mismatch"})
		resourceID := factory.CreateTrackerResource(owner, category1)

		_, err := owner.Do(`
			mutation MoveTrackerResourceToCategory($input: MoveTrackerResourceToCategoryInput!) {
				moveTrackerResourceToCategory(input: $input) {
					trackerResource { id }
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"trackerResourceId":      resourceID,
				"targetCookieCategoryId": category2,
			},
		})
		require.Error(t, err)
	})
}

func TestTrackerResource_Delete(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		resourceID := factory.CreateTrackerResource(owner, categoryID)

		const query = `
			mutation DeleteTrackerResource($input: DeleteTrackerResourceInput!) {
				deleteTrackerResource(input: $input) {
					deletedTrackerResourceId
					cookieBanner {
						id
					}
				}
			}
		`

		var result struct {
			DeleteTrackerResource struct {
				DeletedTrackerResourceID string `json:"deletedTrackerResourceId"`
				CookieBanner             struct {
					ID string `json:"id"`
				} `json:"cookieBanner"`
			} `json:"deleteTrackerResource"`
		}

		err := owner.Execute(query, map[string]any{
			"input": map[string]any{"trackerResourceId": resourceID},
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, resourceID, result.DeleteTrackerResource.DeletedTrackerResourceID)
		assert.Equal(t, bannerID, result.DeleteTrackerResource.CookieBanner.ID)
	})
}

func TestTrackerResource_RBAC(t *testing.T) {
	t.Parallel()

	t.Run("viewer cannot create resource", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)

		_, err := viewer.Do(`
			mutation CreateTrackerResource($input: CreateTrackerResourceInput!) {
				createTrackerResource(input: $input) {
					trackerResourceEdge { node { id } }
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"cookieCategoryId": categoryID,
				"type":             "SCRIPT",
				"origin":           "https://example.com",
				"path":             "/viewer.js",
				"displayName":      "Test Viewer Resource",
				"description":      "Should fail",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to create tracker resource")
	})

	t.Run("viewer cannot update resource", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		resourceID := factory.CreateTrackerResource(owner, categoryID)

		_, err := viewer.Do(`
			mutation UpdateTrackerResource($input: UpdateTrackerResourceInput!) {
				updateTrackerResource(input: $input) {
					trackerResource { id }
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"trackerResourceId": resourceID,
				"description":       "Updated by Viewer",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to update tracker resource")
	})

	t.Run("viewer cannot delete resource", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryID := factory.CreateCookieCategory(owner, bannerID)
		resourceID := factory.CreateTrackerResource(owner, categoryID)

		_, err := viewer.Do(`
			mutation DeleteTrackerResource($input: DeleteTrackerResourceInput!) {
				deleteTrackerResource(input: $input) {
					deletedTrackerResourceId
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{"trackerResourceId": resourceID},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to delete tracker resource")
	})

	t.Run("viewer cannot move resource", func(t *testing.T) {
		t.Parallel()
		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		bannerID := factory.CreateCookieBanner(owner)
		categoryA := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"slug": "rbac-res-move-a"})
		categoryB := factory.CreateCookieCategory(owner, bannerID, factory.Attrs{"slug": "rbac-res-move-b"})
		resourceID := factory.CreateTrackerResource(owner, categoryA)

		_, err := viewer.Do(`
			mutation MoveTrackerResourceToCategory($input: MoveTrackerResourceToCategoryInput!) {
				moveTrackerResourceToCategory(input: $input) {
					trackerResource { id }
					cookieBanner { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"trackerResourceId":      resourceID,
				"targetCookieCategoryId": categoryB,
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer should not be able to move tracker resource")
	})
}
