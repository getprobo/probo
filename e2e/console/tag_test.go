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

func TestTag_CatalogCRUD(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	const createMutation = `
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

	var createResult struct {
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

	err := owner.Execute(createMutation, map[string]any{
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

	const updateMutation = `
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

	var updateResult struct {
		UpdateTag struct {
			Tag struct {
				ID    string  `json:"id"`
				Value string  `json:"value"`
				Color *string `json:"color"`
			} `json:"tag"`
		} `json:"updateTag"`
	}

	err = owner.Execute(updateMutation, map[string]any{
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

	const deleteMutation = `
		mutation($input: DeleteTagInput!) {
			deleteTag(input: $input) {
				deletedTagId
			}
		}
	`

	var deleteResult struct {
		DeleteTag struct {
			DeletedTagID string `json:"deletedTagId"`
		} `json:"deleteTag"`
	}

	err = owner.Execute(deleteMutation, map[string]any{
		"input": map[string]any{
			"tagId": tagID,
		},
	}, &deleteResult)
	require.NoError(t, err)
	assert.Equal(t, tagID, deleteResult.DeleteTag.DeletedTagID)
}

func TestTag_AttachDetachAndConflict(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	documentID := factory.NewDocument(owner).WithTitle("Tagged Document").Create()

	const createMutation = `
		mutation($input: CreateTagInput!) {
			createTag(input: $input) {
				tagEdge { node { id } }
			}
		}
	`

	var createResult struct {
		CreateTag struct {
			TagEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"tagEdge"`
		} `json:"createTag"`
	}

	err := owner.Execute(createMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"key":            "department",
			"value":          "Security",
		},
	}, &createResult)
	require.NoError(t, err)

	tagID := createResult.CreateTag.TagEdge.Node.ID

	const attachMutation = `
		mutation($input: AttachTagInput!) {
			attachTag(input: $input) {
				resourceId
				tagId
			}
		}
	`

	err = owner.Execute(attachMutation, map[string]any{
		"input": map[string]any{
			"resourceId": documentID,
			"tagId":      tagID,
		},
	}, nil)
	require.NoError(t, err)

	_, err = owner.Do(attachMutation, map[string]any{
		"input": map[string]any{
			"resourceId": documentID,
			"tagId":      tagID,
		},
	})
	require.Error(t, err)

	const forResourceQuery = `
		query($resourceId: ID!) {
			tagsForResource(resourceId: $resourceId) {
				id
				key
				value
			}
		}
	`

	var forResourceResult struct {
		TagsForResource []struct {
			ID    string `json:"id"`
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"tagsForResource"`
	}

	err = owner.Execute(forResourceQuery, map[string]any{
		"resourceId": documentID,
	}, &forResourceResult)
	require.NoError(t, err)
	require.Len(t, forResourceResult.TagsForResource, 1)
	assert.Equal(t, tagID, forResourceResult.TagsForResource[0].ID)
	assert.Equal(t, "department", forResourceResult.TagsForResource[0].Key)

	const nodeQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on Tag {
					id
					assignments(first: 50) {
						edges {
							node {
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
						Resource struct {
							ID       string `json:"id"`
							Typename string `json:"__typename"`
						} `json:"resource"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"assignments"`
		} `json:"node"`
	}

	err = owner.Execute(nodeQuery, map[string]any{"id": tagID}, &nodeResult)
	require.NoError(t, err)
	assert.Equal(t, 1, nodeResult.Node.Assignments.TotalCount)
	require.Len(t, nodeResult.Node.Assignments.Edges, 1)
	assert.Equal(t, documentID, nodeResult.Node.Assignments.Edges[0].Node.Resource.ID)
	assert.Equal(t, "Document", nodeResult.Node.Assignments.Edges[0].Node.Resource.Typename)

	const detachMutation = `
		mutation($input: DetachTagInput!) {
			detachTag(input: $input) {
				resourceId
				tagId
			}
		}
	`

	err = owner.Execute(detachMutation, map[string]any{
		"input": map[string]any{
			"resourceId": documentID,
			"tagId":      tagID,
		},
	}, nil)
	require.NoError(t, err)

	err = owner.Execute(forResourceQuery, map[string]any{
		"resourceId": documentID,
	}, &forResourceResult)
	require.NoError(t, err)
	assert.Empty(t, forResourceResult.TagsForResource)
}

func TestTag_DuplicateKeyConflict(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	const createMutation = `
		mutation($input: CreateTagInput!) {
			createTag(input: $input) {
				tagEdge { node { id } }
			}
		}
	`

	err := owner.Execute(createMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"key":            "duplicate-key",
			"value":          "One",
		},
	}, nil)
	require.NoError(t, err)

	_, err = owner.Do(createMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"key":            "duplicate-key",
			"value":          "Two",
		},
	})
	require.Error(t, err)
}
