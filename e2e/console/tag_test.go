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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const (
	createTagMutation = `
		mutation($input: CreateTagInput!) {
			createTag(input: $input) {
				tagEdge {
					node {
						id
						name
						color
					}
				}
			}
		}
	`

	attachTagMutation = `
		mutation($input: AttachTagInput!) {
			attachTag(input: $input) {
				tag {
					id
				}
				tagAssignmentEdge {
					node {
						id
					}
				}
			}
		}
	`

	detachTagMutation = `
		mutation($input: DetachTagInput!) {
			detachTag(input: $input) {
				tag {
					id
				}
				detached
			}
		}
	`

	updateTagMutation = `
		mutation($input: UpdateTagInput!) {
			updateTag(input: $input) {
				tag {
					id
					name
					color
				}
			}
		}
	`

	deleteTagMutation = `
		mutation($input: DeleteTagInput!) {
			deleteTag(input: $input) {
				deletedTagId
			}
		}
	`

	tagNodeQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on Tag {
					id
				}
			}
		}
	`

	resourceTagsQuery = `
		query($resourceId: ID!) {
			node(id: $resourceId) {
				... on Document {
					tags(first: 50) {
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

	searchQuery = `
		query(
			$organizationId: ID!
			$query: String!
		) {
			search(
				organizationId: $organizationId
				query: $query
				first: 50
			) {
				edges {
					title
					node {
						id
						__typename
					}
				}
				totalCount
			}
		}
	`

	deleteDocumentMutation = `
		mutation($input: DeleteDocumentInput!) {
			deleteDocument(input: $input) {
				deletedDocumentId
			}
		}
	`
)

type (
	createTagResult struct {
		CreateTag struct {
			TagEdge struct {
				Node struct {
					ID    string  `json:"id"`
					Name  string  `json:"name"`
					Color *string `json:"color"`
				} `json:"node"`
			} `json:"tagEdge"`
		} `json:"createTag"`
	}

	attachTagResult struct {
		AttachTag struct {
			Tag struct {
				ID string `json:"id"`
			} `json:"tag"`
			TagAssignmentEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"tagAssignmentEdge"`
		} `json:"attachTag"`
	}

	detachTagResult struct {
		DetachTag struct {
			Tag struct {
				ID string `json:"id"`
			} `json:"tag"`
			Detached bool `json:"detached"`
		} `json:"detachTag"`
	}

	resourceTag struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	resourceTagsResult struct {
		Node *struct {
			Tags struct {
				Edges []struct {
					Node resourceTag `json:"node"`
				} `json:"edges"`
			} `json:"tags"`
		} `json:"node"`
	}

	searchResult struct {
		Search struct {
			Edges []struct {
				Title string `json:"title"`
				Node  struct {
					ID       string `json:"id"`
					Typename string `json:"__typename"`
				} `json:"node"`
			} `json:"edges"`
			TotalCount int `json:"totalCount"`
		} `json:"search"`
	}
)

func createTag(t *testing.T, c *testutil.Client, name string) string {
	t.Helper()

	var result createTagResult

	err := c.Execute(createTagMutation, map[string]any{
		"input": map[string]any{
			"organizationId": c.GetOrganizationID().String(),
			"name":           name,
		},
	}, &result)
	require.NoError(t, err)

	return result.CreateTag.TagEdge.Node.ID
}

func attachTag(t *testing.T, c *testutil.Client, resourceID string, tagID string) attachTagResult {
	t.Helper()

	var result attachTagResult

	err := c.Execute(attachTagMutation, map[string]any{
		"input": map[string]any{
			"resourceId": resourceID,
			"tagId":      tagID,
		},
	}, &result)
	require.NoError(t, err)

	return result
}

func detachTag(t *testing.T, c *testutil.Client, resourceID string, tagID string) detachTagResult {
	t.Helper()

	var result detachTagResult

	err := c.Execute(detachTagMutation, map[string]any{
		"input": map[string]any{
			"resourceId": resourceID,
			"tagId":      tagID,
		},
	}, &result)
	require.NoError(t, err)

	return result
}

func resourceTags(t *testing.T, c *testutil.Client, resourceID string) []resourceTag {
	t.Helper()

	var result resourceTagsResult

	err := c.Execute(resourceTagsQuery, map[string]any{
		"resourceId": resourceID,
	}, &result)
	require.NoError(t, err)
	require.NotNil(t, result.Node)

	tags := make([]resourceTag, len(result.Node.Tags.Edges))
	for i, edge := range result.Node.Tags.Edges {
		tags[i] = edge.Node
	}

	return tags
}

func searchResources(
	t *testing.T,
	c *testutil.Client,
	organizationID string,
	query string,
) searchResult {
	t.Helper()

	var result searchResult

	err := c.Execute(searchQuery, map[string]any{
		"organizationId": organizationID,
		"query":          query,
	}, &result)
	require.NoError(t, err)

	return result
}

func TestTag_CatalogCRUD(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	var createResult createTagResult

	err := owner.Execute(createTagMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "Production",
			"color":          "#0f0",
		},
	}, &createResult)
	require.NoError(t, err)
	assert.Equal(t, "Production", createResult.CreateTag.TagEdge.Node.Name)
	require.NotNil(t, createResult.CreateTag.TagEdge.Node.Color)
	assert.Equal(t, "#0f0", *createResult.CreateTag.TagEdge.Node.Color)

	tagID := createResult.CreateTag.TagEdge.Node.ID

	var updateResult struct {
		UpdateTag struct {
			Tag struct {
				ID    string  `json:"id"`
				Name  string  `json:"name"`
				Color *string `json:"color"`
			} `json:"tag"`
		} `json:"updateTag"`
	}

	err = owner.Execute(updateTagMutation, map[string]any{
		"input": map[string]any{
			"id":    tagID,
			"name":  "Staging",
			"color": "#00ff00",
		},
	}, &updateResult)
	require.NoError(t, err)
	assert.Equal(t, "Staging", updateResult.UpdateTag.Tag.Name)
	require.NotNil(t, updateResult.UpdateTag.Tag.Color)
	assert.Equal(t, "#00ff00", *updateResult.UpdateTag.Tag.Color)

	err = owner.Execute(updateTagMutation, map[string]any{
		"input": map[string]any{
			"id":   tagID,
			"name": "Staging",
		},
	}, &updateResult)
	require.NoError(t, err)
	require.NotNil(t, updateResult.UpdateTag.Tag.Color, "omitting color keeps it")
	assert.Equal(t, "#00ff00", *updateResult.UpdateTag.Tag.Color)

	err = owner.Execute(updateTagMutation, map[string]any{
		"input": map[string]any{
			"id":    tagID,
			"color": nil,
		},
	}, &updateResult)
	require.NoError(t, err)
	assert.Nil(t, updateResult.UpdateTag.Tag.Color, "null color clears it")

	const listQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					tags(first: 50) {
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

	var listResult struct {
		Node struct {
			Tags struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"tags"`
		} `json:"node"`
	}

	err = owner.Execute(listQuery, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &listResult)
	require.NoError(t, err)
	require.GreaterOrEqual(t, listResult.Node.Tags.TotalCount, 1)

	found := false

	for _, edge := range listResult.Node.Tags.Edges {
		if edge.Node.ID == tagID {
			found = true

			assert.Equal(t, "Staging", edge.Node.Name)
		}
	}

	assert.True(t, found)

	var deleteResult struct {
		DeleteTag struct {
			DeletedTagID string `json:"deletedTagId"`
		} `json:"deleteTag"`
	}

	err = owner.Execute(deleteTagMutation, map[string]any{
		"input": map[string]any{
			"tagId": tagID,
		},
	}, &deleteResult)
	require.NoError(t, err)
	assert.Equal(t, tagID, deleteResult.DeleteTag.DeletedTagID)
}

func TestTag_NameIsUnique(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	securityID := createTag(t, owner, "Security")
	engineeringID := createTag(t, owner, "Engineering")
	assert.NotEqual(t, securityID, engineeringID)

	_, err := owner.Do(createTagMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"name":           "Security",
		},
	})
	testutil.RequireErrorCode(t, err, "CONFLICT")
}

func TestTag_AttachAndDetach(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	documentID := factory.NewDocument(owner).WithTitle("Tagged Document").Create()
	tagID := createTag(t, owner, "Internal")

	first := attachTag(t, owner, documentID, tagID)
	assert.Equal(t, tagID, first.AttachTag.Tag.ID)
	assert.Equal(t, documentID, first.AttachTag.TagAssignmentEdge.Node.ID)

	second := attachTag(t, owner, documentID, tagID)
	assert.Equal(
		t,
		documentID,
		second.AttachTag.TagAssignmentEdge.Node.ID,
		"attaching the same tag again returns the resource",
	)

	forResource := resourceTags(t, owner, documentID)
	require.Len(t, forResource, 1)
	assert.Equal(t, tagID, forResource[0].ID)

	const nodeQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on Tag {
					id
					assignments(first: 50) {
						edges {
							createdAt
							node {
								id
								__typename
							}
						}
						totalCount
					}
				}
			}
		}
	`

	var nodeResult struct {
		Node struct {
			ID          string `json:"id"`
			Assignments struct {
				Edges []struct {
					CreatedAt string `json:"createdAt"`
					Node      *struct {
						ID       string `json:"id"`
						Typename string `json:"__typename"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"assignments"`
		} `json:"node"`
	}

	err := owner.Execute(nodeQuery, map[string]any{"id": tagID}, &nodeResult)
	require.NoError(t, err)
	assert.Equal(t, 1, nodeResult.Node.Assignments.TotalCount)
	require.Len(t, nodeResult.Node.Assignments.Edges, 1)
	require.NotNil(t, nodeResult.Node.Assignments.Edges[0].Node)
	assert.NotEmpty(t, nodeResult.Node.Assignments.Edges[0].CreatedAt)
	assert.Equal(t, documentID, nodeResult.Node.Assignments.Edges[0].Node.ID)
	assert.Equal(t, "Document", nodeResult.Node.Assignments.Edges[0].Node.Typename)

	detached := detachTag(t, owner, documentID, tagID)
	assert.True(t, detached.DetachTag.Detached)

	again := detachTag(t, owner, documentID, tagID)
	assert.False(t, again.DetachTag.Detached, "detaching a tag that is not attached is a no-op")

	assert.Empty(t, resourceTags(t, owner, documentID))
}

func TestTag_AssignmentsPaginate(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	tagID := createTag(t, owner, "Paginated")

	documentIDs := map[string]struct{}{}

	for _, title := range []string{"Page Doc A", "Page Doc B", "Page Doc C"} {
		documentID := factory.NewDocument(owner).WithTitle(title).Create()
		attachTag(t, owner, documentID, tagID)

		documentIDs[documentID] = struct{}{}
	}

	const query = `
		query($id: ID!, $first: Int!, $after: CursorKey) {
			node(id: $id) {
				... on Tag {
					assignments(first: $first, after: $after) {
						edges {
							node {
								id
							}
						}
						pageInfo {
							hasNextPage
							endCursor
						}
						totalCount
					}
				}
			}
		}
	`

	type assignmentsPage struct {
		Node struct {
			Assignments struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
				PageInfo struct {
					HasNextPage bool    `json:"hasNextPage"`
					EndCursor   *string `json:"endCursor"`
				} `json:"pageInfo"`
				TotalCount int `json:"totalCount"`
			} `json:"assignments"`
		} `json:"node"`
	}

	var firstPage assignmentsPage

	err := owner.Execute(query, map[string]any{"id": tagID, "first": 2}, &firstPage)
	require.NoError(t, err)
	assert.Equal(t, 3, firstPage.Node.Assignments.TotalCount)
	require.Len(t, firstPage.Node.Assignments.Edges, 2)
	require.True(t, firstPage.Node.Assignments.PageInfo.HasNextPage)
	require.NotNil(t, firstPage.Node.Assignments.PageInfo.EndCursor)

	var secondPage assignmentsPage

	err = owner.Execute(query, map[string]any{
		"id":    tagID,
		"first": 2,
		"after": *firstPage.Node.Assignments.PageInfo.EndCursor,
	}, &secondPage)
	require.NoError(t, err)
	require.Len(t, secondPage.Node.Assignments.Edges, 1)
	assert.False(t, secondPage.Node.Assignments.PageInfo.HasNextPage)

	seen := map[string]struct{}{}

	for _, page := range []assignmentsPage{firstPage, secondPage} {
		for _, edge := range page.Node.Assignments.Edges {
			seen[edge.Node.ID] = struct{}{}
		}
	}

	assert.Equal(t, documentIDs, seen, "both pages together return every assignment exactly once")
}

func TestTag_AttachKeepsManyTags(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	documentID := factory.NewDocument(owner).WithTitle("Multi Tag Document").Create()
	isoID := createTag(t, owner, "ISO27001")
	socID := createTag(t, owner, "SOC2")
	platformID := createTag(t, owner, "Platform")

	attachTag(t, owner, documentID, isoID)
	attachTag(t, owner, documentID, platformID)
	attachTag(t, owner, documentID, socID)

	tags := resourceTags(t, owner, documentID)
	require.Len(t, tags, 3)

	names := map[string]struct{}{}
	for _, tag := range tags {
		names[tag.Name] = struct{}{}
	}

	assert.Contains(t, names, "ISO27001")
	assert.Contains(t, names, "SOC2")
	assert.Contains(t, names, "Platform")
}

func TestTag_AttachRejectsUntaggableResource(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	tagID := createTag(t, owner, "Global")
	otherTagID := createTag(t, owner, "Local")

	_, err := owner.Do(attachTagMutation, map[string]any{
		"input": map[string]any{
			"resourceId": otherTagID,
			"tagId":      tagID,
		},
	})
	testutil.RequireErrorCode(t, err, "INVALID")
}

func TestTag_DeleteRemovesAssignments(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	documentID := factory.NewDocument(owner).WithTitle("Cascade Document").Create()
	tagID := createTag(t, owner, "Deprecated")

	attachTag(t, owner, documentID, tagID)
	require.Len(t, resourceTags(t, owner, documentID), 1)

	_, err := owner.Do(deleteTagMutation, map[string]any{
		"input": map[string]any{"tagId": tagID},
	})
	require.NoError(t, err)

	assert.Empty(t, resourceTags(t, owner, documentID))

	var nodeResult struct {
		Node *struct {
			ID string `json:"id"`
		} `json:"node"`
	}

	err = owner.Execute(tagNodeQuery, map[string]any{"id": tagID}, &nodeResult)
	testutil.AssertNodeNotAccessible(t, err, nodeResult.Node == nil, "Tag")
}

func TestTag_RBAC(t *testing.T) {
	t.Parallel()

	t.Run("viewer can read tags", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		documentID := factory.NewDocument(owner).WithTitle("Viewer Document").Create()
		tagID := createTag(t, owner, "Security")
		attachTag(t, owner, documentID, tagID)

		var nodeResult struct {
			Node *struct {
				ID string `json:"id"`
			} `json:"node"`
		}

		err := viewer.Execute(tagNodeQuery, map[string]any{"id": tagID}, &nodeResult)
		require.NoError(t, err)
		require.NotNil(t, nodeResult.Node)
		assert.Equal(t, tagID, nodeResult.Node.ID)

		tags := resourceTags(t, viewer, documentID)
		require.Len(t, tags, 1)
		assert.Equal(t, tagID, tags[0].ID)
	})

	t.Run("viewer cannot change tags", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		documentID := factory.NewDocument(owner).WithTitle("Viewer Document").Create()
		tagID := createTag(t, owner, "Platform")
		attachedTagID := createTag(t, owner, "EU")
		attachTag(t, owner, documentID, attachedTagID)

		_, err := viewer.Do(createTagMutation, map[string]any{
			"input": map[string]any{
				"organizationId": viewer.GetOrganizationID().String(),
				"name":           "Viewer",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer cannot create")

		_, err = viewer.Do(updateTagMutation, map[string]any{
			"input": map[string]any{"id": tagID, "name": "Changed"},
		})
		testutil.RequireForbiddenError(t, err, "viewer cannot update")

		_, err = viewer.Do(deleteTagMutation, map[string]any{
			"input": map[string]any{"tagId": tagID},
		})
		testutil.RequireForbiddenError(t, err, "viewer cannot delete")

		_, err = viewer.Do(attachTagMutation, map[string]any{
			"input": map[string]any{"resourceId": documentID, "tagId": tagID},
		})
		testutil.RequireForbiddenError(t, err, "viewer cannot attach")

		_, err = viewer.Do(detachTagMutation, map[string]any{
			"input": map[string]any{"resourceId": documentID, "tagId": attachedTagID},
		})
		testutil.RequireForbiddenError(t, err, "viewer cannot detach")
	})

	t.Run("admin can manage tags", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
		documentID := factory.NewDocument(owner).WithTitle("Admin Document").Create()

		tagID := createTag(t, admin, "Compliance")
		attached := attachTag(t, admin, documentID, tagID)
		assert.Equal(t, tagID, attached.AttachTag.Tag.ID)

		detached := detachTag(t, admin, documentID, tagID)
		assert.True(t, detached.DetachTag.Detached)

		_, err := admin.Do(deleteTagMutation, map[string]any{
			"input": map[string]any{"tagId": tagID},
		})
		require.NoError(t, err)
	})
}

func TestTag_TenantIsolation(t *testing.T) {
	t.Parallel()

	owner1 := testutil.NewClient(t, testutil.RoleOwner)
	owner2 := testutil.NewClient(t, testutil.RoleOwner)

	document1ID := factory.NewDocument(owner1).WithTitle("Org 1 Document").Create()
	document2ID := factory.NewDocument(owner2).WithTitle("Org 2 Document").Create()
	tag1ID := createTag(t, owner1, "Production")
	tag2ID := createTag(t, owner2, "Production")
	attachTag(t, owner1, document1ID, tag1ID)

	// The group returns only once its parallel subtests finish, so the
	// final assertions observe every cross-org attempt.
	t.Run("cross-org access", func(t *testing.T) {
		t.Run("other org cannot read the tag", func(t *testing.T) {
			t.Parallel()

			var nodeResult struct {
				Node *struct {
					ID string `json:"id"`
				} `json:"node"`
			}

			err := owner2.Execute(tagNodeQuery, map[string]any{"id": tag1ID}, &nodeResult)
			testutil.AssertNodeNotAccessible(t, err, nodeResult.Node == nil, "Tag")
		})

		t.Run("other org cannot list tags of a resource", func(t *testing.T) {
			t.Parallel()

			var result resourceTagsResult

			err := owner2.Execute(resourceTagsQuery, map[string]any{"resourceId": document1ID}, &result)
			testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "Document")
		})

		t.Run("other org cannot change the tag", func(t *testing.T) {
			t.Parallel()

			_, err := owner2.Do(updateTagMutation, map[string]any{
				"input": map[string]any{"id": tag1ID, "name": "Hijacked"},
			})
			require.Error(t, err)

			_, err = owner2.Do(deleteTagMutation, map[string]any{
				"input": map[string]any{"tagId": tag1ID},
			})
			require.Error(t, err)
		})

		t.Run("cannot attach a tag from another org", func(t *testing.T) {
			t.Parallel()

			_, err := owner2.Do(attachTagMutation, map[string]any{
				"input": map[string]any{"resourceId": document2ID, "tagId": tag1ID},
			})
			require.Error(t, err)

			assert.Empty(t, resourceTags(t, owner2, document2ID))
		})

		t.Run("cannot tag a resource of another org", func(t *testing.T) {
			t.Parallel()

			_, err := owner2.Do(attachTagMutation, map[string]any{
				"input": map[string]any{"resourceId": document1ID, "tagId": tag2ID},
			})
			require.Error(t, err)

			_, err = owner2.Do(detachTagMutation, map[string]any{
				"input": map[string]any{"resourceId": document1ID, "tagId": tag1ID},
			})
			require.Error(t, err)
		})
	})

	tags := resourceTags(t, owner1, document1ID)
	require.Len(t, tags, 1)
	assert.Equal(t, tag1ID, tags[0].ID)
}

func TestTag_Search(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	other := testutil.NewClient(t, testutil.RoleOwner)
	documentID := factory.NewDocument(owner).WithTitle("Search Document UniqueTitle").Create()
	orgID := owner.GetOrganizationID().String()

	productionID := createTag(t, owner, "SearchProduction")
	complianceID := createTag(t, owner, "SearchCompliance")
	createTag(t, other, "SearchProduction")

	attachTag(t, owner, documentID, productionID)

	byTitle := searchResources(t, owner, orgID, "UniqueTitle")
	require.Len(t, byTitle.Search.Edges, 1)
	assert.Equal(t, documentID, byTitle.Search.Edges[0].Node.ID)
	assert.Equal(t, "Search Document UniqueTitle", byTitle.Search.Edges[0].Title)
	assert.Equal(t, "Document", byTitle.Search.Edges[0].Node.Typename)

	anyMatch := searchResources(t, owner, orgID, "tag:SearchProduction")
	require.Len(t, anyMatch.Search.Edges, 1)
	assert.Equal(t, documentID, anyMatch.Search.Edges[0].Node.ID)
	assert.Equal(t, 1, anyMatch.Search.TotalCount)

	allMatch := searchResources(
		t,
		owner,
		orgID,
		"tag:SearchProduction tag:SearchCompliance",
	)
	assert.Empty(t, allMatch.Search.Edges)
	assert.Equal(t, 0, allMatch.Search.TotalCount)

	attachTag(t, owner, documentID, complianceID)

	allMatch = searchResources(
		t,
		owner,
		orgID,
		"tag:SearchProduction tag:SearchCompliance",
	)
	require.Len(t, allMatch.Search.Edges, 1)
	assert.Equal(t, documentID, allMatch.Search.Edges[0].Node.ID)

	taggedAndTitled := searchResources(
		t,
		owner,
		orgID,
		"UniqueTitle tag:SearchProduction",
	)
	require.Len(t, taggedAndTitled.Search.Edges, 1)
	assert.Equal(t, documentID, taggedAndTitled.Search.Edges[0].Node.ID)

	documentsOnly := searchResources(t, owner, orgID, "tag:SearchProduction type:document")
	require.Len(t, documentsOnly.Search.Edges, 1)
	assert.Equal(t, documentID, documentsOnly.Search.Edges[0].Node.ID)

	risksOnly := searchResources(t, owner, orgID, "tag:SearchProduction type:risk")
	assert.Empty(t, risksOnly.Search.Edges)

	err := owner.ExecuteShouldFail(searchQuery, map[string]any{
		"organizationId": orgID,
		"query":          "type:document",
	})
	assert.ErrorContains(t, err, "query")

	err = owner.ExecuteShouldFail(searchQuery, map[string]any{
		"organizationId": orgID,
		"query":          "   ",
	})
	assert.ErrorContains(t, err, "query")

	otherOrg := searchResources(
		t,
		other,
		other.GetOrganizationID().String(),
		"tag:SearchProduction",
	)
	assert.Empty(t, otherOrg.Search.Edges)

	err = owner.Execute(deleteDocumentMutation, map[string]any{
		"input": map[string]any{
			"documentId": documentID,
		},
	}, &struct {
		DeleteDocument struct {
			DeletedDocumentID string `json:"deletedDocumentId"`
		} `json:"deleteDocument"`
	}{})
	require.NoError(t, err)

	afterDelete := searchResources(t, owner, orgID, "tag:SearchProduction")
	assert.Empty(t, afterDelete.Search.Edges)
	assert.Equal(t, 0, afterDelete.Search.TotalCount)

	afterDeleteTitle := searchResources(t, owner, orgID, "UniqueTitle")
	assert.Empty(t, afterDeleteTitle.Search.Edges)

	analysisID := factory.CreateRiskAnalysis(owner, factory.Attrs{"name": "Cascade Analysis UniqueParent"})
	factory.CreateRiskAnalysisDiagram(owner, analysisID, factory.Attrs{
		"name": "Cascade Diagram UniqueChild",
	})

	byChild := searchResources(t, owner, orgID, "UniqueChild")
	assert.Empty(t, byChild.Search.Edges)

	byParent := searchResources(t, owner, orgID, "UniqueParent")
	require.Len(t, byParent.Search.Edges, 1)
	assert.Equal(t, analysisID, byParent.Search.Edges[0].Node.ID)

	_, err = owner.Do(`
		mutation($input: DeleteRiskAnalysisInput!) {
			deleteRiskAnalysis(input: $input) { deletedRiskAnalysisId }
		}
	`, map[string]any{"input": map[string]any{"riskAnalysisId": analysisID}})
	require.NoError(t, err)

	afterDeleteParent := searchResources(t, owner, orgID, "UniqueParent")
	assert.Empty(t, afterDeleteParent.Search.Edges)
}
