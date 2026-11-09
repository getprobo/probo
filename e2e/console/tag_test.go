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
						key
						value
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
				replacedTagAssignmentId
			}
		}
	`

	detachTagMutation = `
		mutation($input: DetachTagInput!) {
			detachTag(input: $input) {
				tag {
					id
				}
				deletedTagAssignmentId
			}
		}
	`

	updateTagMutation = `
		mutation($input: UpdateTagInput!) {
			updateTag(input: $input) {
				tag {
					id
					value
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

	tagsForResourceQuery = `
		query($resourceId: ID!) {
			tagsForResource(resourceId: $resourceId) {
				id
				key
				value
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
					Key   string  `json:"key"`
					Value string  `json:"value"`
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
			ReplacedTagAssignmentID *string `json:"replacedTagAssignmentId"`
		} `json:"attachTag"`
	}

	detachTagResult struct {
		DetachTag struct {
			Tag struct {
				ID string `json:"id"`
			} `json:"tag"`
			DeletedTagAssignmentID *string `json:"deletedTagAssignmentId"`
		} `json:"detachTag"`
	}

	tagsForResourceResult struct {
		TagsForResource []struct {
			ID    string `json:"id"`
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"tagsForResource"`
	}
)

func createTag(t *testing.T, c *testutil.Client, key string, value string) string {
	t.Helper()

	var result createTagResult

	err := c.Execute(createTagMutation, map[string]any{
		"input": map[string]any{
			"organizationId": c.GetOrganizationID().String(),
			"key":            key,
			"value":          value,
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

func tagsForResource(t *testing.T, c *testutil.Client, resourceID string) tagsForResourceResult {
	t.Helper()

	var result tagsForResourceResult

	err := c.Execute(tagsForResourceQuery, map[string]any{
		"resourceId": resourceID,
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
			"key":            "environment",
			"value":          "Production",
			"color":          "#0f0",
		},
	}, &createResult)
	require.NoError(t, err)
	assert.Equal(t, "environment", createResult.CreateTag.TagEdge.Node.Key)
	assert.Equal(t, "Production", createResult.CreateTag.TagEdge.Node.Value)
	require.NotNil(t, createResult.CreateTag.TagEdge.Node.Color)
	assert.Equal(t, "#0f0", *createResult.CreateTag.TagEdge.Node.Color)

	tagID := createResult.CreateTag.TagEdge.Node.ID

	var updateResult struct {
		UpdateTag struct {
			Tag struct {
				ID    string  `json:"id"`
				Value string  `json:"value"`
				Color *string `json:"color"`
			} `json:"tag"`
		} `json:"updateTag"`
	}

	err = owner.Execute(updateTagMutation, map[string]any{
		"input": map[string]any{
			"id":    tagID,
			"value": "Staging",
			"color": "#00ff00",
		},
	}, &updateResult)
	require.NoError(t, err)
	assert.Equal(t, "Staging", updateResult.UpdateTag.Tag.Value)
	require.NotNil(t, updateResult.UpdateTag.Tag.Color)
	assert.Equal(t, "#00ff00", *updateResult.UpdateTag.Tag.Color)

	err = owner.Execute(updateTagMutation, map[string]any{
		"input": map[string]any{
			"id":    tagID,
			"value": "Staging",
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
								key
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
						ID  string `json:"id"`
						Key string `json:"key"`
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

			assert.Equal(t, "environment", edge.Node.Key)
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

func TestTag_KeyHoldsSeveralValues(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	securityID := createTag(t, owner, "department", "Security")
	engineeringID := createTag(t, owner, "department", "Engineering")
	assert.NotEqual(t, securityID, engineeringID)

	_, err := owner.Do(createTagMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"key":            "department",
			"value":          "Security",
		},
	})
	testutil.RequireErrorCode(t, err, "CONFLICT")
}

func TestTag_AttachAndDetach(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	documentID := factory.NewDocument(owner).WithTitle("Tagged Document").Create()
	tagID := createTag(t, owner, "classification", "Internal")

	first := attachTag(t, owner, documentID, tagID)
	assert.Equal(t, tagID, first.AttachTag.Tag.ID)
	assert.NotEmpty(t, first.AttachTag.TagAssignmentEdge.Node.ID)
	assert.Nil(t, first.AttachTag.ReplacedTagAssignmentID)

	second := attachTag(t, owner, documentID, tagID)
	assert.Equal(
		t,
		first.AttachTag.TagAssignmentEdge.Node.ID,
		second.AttachTag.TagAssignmentEdge.Node.ID,
		"attaching the same tag again returns the existing assignment",
	)
	assert.Nil(t, second.AttachTag.ReplacedTagAssignmentID)

	forResource := tagsForResource(t, owner, documentID)
	require.Len(t, forResource.TagsForResource, 1)
	assert.Equal(t, tagID, forResource.TagsForResource[0].ID)

	const nodeQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on Tag {
					id
					assignments(first: 50) {
						edges {
							node {
								id
								resource {
									id
									__typename
								}
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
					Node struct {
						ID       string `json:"id"`
						Resource *struct {
							ID       string `json:"id"`
							Typename string `json:"__typename"`
						} `json:"resource"`
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
	require.NotNil(t, nodeResult.Node.Assignments.Edges[0].Node.Resource)
	assert.Equal(t, documentID, nodeResult.Node.Assignments.Edges[0].Node.Resource.ID)
	assert.Equal(t, "Document", nodeResult.Node.Assignments.Edges[0].Node.Resource.Typename)

	detached := detachTag(t, owner, documentID, tagID)
	require.NotNil(t, detached.DetachTag.DeletedTagAssignmentID)
	assert.Equal(t, first.AttachTag.TagAssignmentEdge.Node.ID, *detached.DetachTag.DeletedTagAssignmentID)

	again := detachTag(t, owner, documentID, tagID)
	assert.Nil(t, again.DetachTag.DeletedTagAssignmentID, "detaching a tag that is not attached is a no-op")

	assert.Empty(t, tagsForResource(t, owner, documentID).TagsForResource)
}

func TestTag_AttachReplacesValueOfSameKey(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	documentID := factory.NewDocument(owner).WithTitle("Environment Document").Create()
	productionID := createTag(t, owner, "env", "Production")
	stagingID := createTag(t, owner, "env", "Staging")
	ownerTagID := createTag(t, owner, "owner", "Platform")

	production := attachTag(t, owner, documentID, productionID)
	attachTag(t, owner, documentID, ownerTagID)

	staging := attachTag(t, owner, documentID, stagingID)
	require.NotNil(t, staging.AttachTag.ReplacedTagAssignmentID)
	assert.Equal(t, production.AttachTag.TagAssignmentEdge.Node.ID, *staging.AttachTag.ReplacedTagAssignmentID)

	tags := tagsForResource(t, owner, documentID).TagsForResource
	require.Len(t, tags, 2, "another key on the same resource is kept")

	values := map[string]string{}
	for _, tag := range tags {
		values[tag.Key] = tag.Value
	}

	assert.Equal(t, "Staging", values["env"])
	assert.Equal(t, "Platform", values["owner"])
}

func TestTag_AttachRejectsUntaggableResource(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	tagID := createTag(t, owner, "scope", "Global")
	otherTagID := createTag(t, owner, "scope", "Local")

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
	tagID := createTag(t, owner, "lifecycle", "Deprecated")

	attachTag(t, owner, documentID, tagID)
	require.Len(t, tagsForResource(t, owner, documentID).TagsForResource, 1)

	_, err := owner.Do(deleteTagMutation, map[string]any{
		"input": map[string]any{"tagId": tagID},
	})
	require.NoError(t, err)

	assert.Empty(t, tagsForResource(t, owner, documentID).TagsForResource)

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
		tagID := createTag(t, owner, "team", "Security")
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

		tags := tagsForResource(t, viewer, documentID).TagsForResource
		require.Len(t, tags, 1)
		assert.Equal(t, tagID, tags[0].ID)
	})

	t.Run("viewer cannot change tags", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
		documentID := factory.NewDocument(owner).WithTitle("Viewer Document").Create()
		tagID := createTag(t, owner, "team", "Platform")
		attachedTagID := createTag(t, owner, "region", "EU")
		attachTag(t, owner, documentID, attachedTagID)

		_, err := viewer.Do(createTagMutation, map[string]any{
			"input": map[string]any{
				"organizationId": viewer.GetOrganizationID().String(),
				"key":            "team",
				"value":          "Viewer",
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer cannot create")

		_, err = viewer.Do(updateTagMutation, map[string]any{
			"input": map[string]any{"id": tagID, "value": "Changed"},
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

		tagID := createTag(t, admin, "team", "Compliance")
		attached := attachTag(t, admin, documentID, tagID)
		assert.Equal(t, tagID, attached.AttachTag.Tag.ID)

		detached := detachTag(t, admin, documentID, tagID)
		require.NotNil(t, detached.DetachTag.DeletedTagAssignmentID)

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
	tag1ID := createTag(t, owner1, "env", "Production")
	tag2ID := createTag(t, owner2, "env", "Production")
	attachTag(t, owner1, document1ID, tag1ID)

	t.Run("other org cannot read the tag", func(t *testing.T) {
		var nodeResult struct {
			Node *struct {
				ID string `json:"id"`
			} `json:"node"`
		}

		err := owner2.Execute(tagNodeQuery, map[string]any{"id": tag1ID}, &nodeResult)
		testutil.AssertNodeNotAccessible(t, err, nodeResult.Node == nil, "Tag")
	})

	t.Run("other org cannot list tags of a resource", func(t *testing.T) {
		_, err := owner2.Do(tagsForResourceQuery, map[string]any{"resourceId": document1ID})
		require.Error(t, err)
	})

	t.Run("other org cannot change the tag", func(t *testing.T) {
		_, err := owner2.Do(updateTagMutation, map[string]any{
			"input": map[string]any{"id": tag1ID, "value": "Hijacked"},
		})
		require.Error(t, err)

		_, err = owner2.Do(deleteTagMutation, map[string]any{
			"input": map[string]any{"tagId": tag1ID},
		})
		require.Error(t, err)
	})

	t.Run("cannot attach a tag from another org", func(t *testing.T) {
		_, err := owner2.Do(attachTagMutation, map[string]any{
			"input": map[string]any{"resourceId": document2ID, "tagId": tag1ID},
		})
		require.Error(t, err)

		assert.Empty(t, tagsForResource(t, owner2, document2ID).TagsForResource)
	})

	t.Run("cannot tag a resource of another org", func(t *testing.T) {
		_, err := owner2.Do(attachTagMutation, map[string]any{
			"input": map[string]any{"resourceId": document1ID, "tagId": tag2ID},
		})
		require.Error(t, err)

		_, err = owner2.Do(detachTagMutation, map[string]any{
			"input": map[string]any{"resourceId": document1ID, "tagId": tag1ID},
		})
		require.Error(t, err)
	})

	tags := tagsForResource(t, owner1, document1ID).TagsForResource
	require.Len(t, tags, 1)
	assert.Equal(t, tag1ID, tags[0].ID)
}
