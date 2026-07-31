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

func TestResourceTag_CatalogCRUD(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	const createMutation = `
		mutation($input: CreateResourceTagInput!) {
			createResourceTag(input: $input) {
				resourceTagEdge {
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
		CreateResourceTag struct {
			ResourceTagEdge struct {
				Node struct {
					ID    string  `json:"id"`
					Key   string  `json:"key"`
					Value string  `json:"value"`
					Color *string `json:"color"`
				} `json:"node"`
			} `json:"resourceTagEdge"`
		} `json:"createResourceTag"`
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
	assert.Equal(t, "environment", createResult.CreateResourceTag.ResourceTagEdge.Node.Key)
	assert.Equal(t, "Production", createResult.CreateResourceTag.ResourceTagEdge.Node.Value)
	require.NotNil(t, createResult.CreateResourceTag.ResourceTagEdge.Node.Color)
	assert.Equal(t, "#0f0", *createResult.CreateResourceTag.ResourceTagEdge.Node.Color)

	tagID := createResult.CreateResourceTag.ResourceTagEdge.Node.ID

	const updateMutation = `
		mutation($input: UpdateResourceTagInput!) {
			updateResourceTag(input: $input) {
				resourceTag {
					id
					value
					color
				}
			}
		}
	`

	var updateResult struct {
		UpdateResourceTag struct {
			ResourceTag struct {
				ID    string  `json:"id"`
				Value string  `json:"value"`
				Color *string `json:"color"`
			} `json:"resourceTag"`
		} `json:"updateResourceTag"`
	}

	err = owner.Execute(updateMutation, map[string]any{
		"input": map[string]any{
			"id":    tagID,
			"value": "Staging",
			"color": "#00ff00",
		},
	}, &updateResult)
	require.NoError(t, err)
	assert.Equal(t, "Staging", updateResult.UpdateResourceTag.ResourceTag.Value)
	require.NotNil(t, updateResult.UpdateResourceTag.ResourceTag.Color)
	assert.Equal(t, "#00ff00", *updateResult.UpdateResourceTag.ResourceTag.Color)

	const listQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					resourceTags(first: 50) {
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
			ResourceTags struct {
				Edges []struct {
					Node struct {
						ID  string `json:"id"`
						Key string `json:"key"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"resourceTags"`
		} `json:"node"`
	}

	err = owner.Execute(listQuery, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &listResult)
	require.NoError(t, err)
	require.GreaterOrEqual(t, listResult.Node.ResourceTags.TotalCount, 1)

	found := false
	for _, edge := range listResult.Node.ResourceTags.Edges {
		if edge.Node.ID == tagID {
			found = true
			assert.Equal(t, "environment", edge.Node.Key)
		}
	}
	assert.True(t, found)

	const deleteMutation = `
		mutation($input: DeleteResourceTagInput!) {
			deleteResourceTag(input: $input) {
				deletedResourceTagId
			}
		}
	`

	var deleteResult struct {
		DeleteResourceTag struct {
			DeletedResourceTagID string `json:"deletedResourceTagId"`
		} `json:"deleteResourceTag"`
	}

	err = owner.Execute(deleteMutation, map[string]any{
		"input": map[string]any{
			"resourceTagId": tagID,
		},
	}, &deleteResult)
	require.NoError(t, err)
	assert.Equal(t, tagID, deleteResult.DeleteResourceTag.DeletedResourceTagID)
}

func TestResourceTag_AttachDetachAndConflict(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	documentID := factory.NewDocument(owner).WithTitle("Tagged Document").Create()

	const createMutation = `
		mutation($input: CreateResourceTagInput!) {
			createResourceTag(input: $input) {
				resourceTagEdge { node { id } }
			}
		}
	`

	var createResult struct {
		CreateResourceTag struct {
			ResourceTagEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"resourceTagEdge"`
		} `json:"createResourceTag"`
	}

	err := owner.Execute(createMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"key":            "department",
			"value":          "Security",
		},
	}, &createResult)
	require.NoError(t, err)

	tagID := createResult.CreateResourceTag.ResourceTagEdge.Node.ID

	const attachMutation = `
		mutation($input: AttachResourceTagInput!) {
			attachResourceTag(input: $input) {
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
			resourceTagsForResource(resourceId: $resourceId) {
				id
				key
				value
			}
		}
	`

	var forResourceResult struct {
		ResourceTagsForResource []struct {
			ID    string `json:"id"`
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"resourceTagsForResource"`
	}

	err = owner.Execute(forResourceQuery, map[string]any{
		"resourceId": documentID,
	}, &forResourceResult)
	require.NoError(t, err)
	require.Len(t, forResourceResult.ResourceTagsForResource, 1)
	assert.Equal(t, tagID, forResourceResult.ResourceTagsForResource[0].ID)
	assert.Equal(t, "department", forResourceResult.ResourceTagsForResource[0].Key)

	const nodeQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on ResourceTag {
					id
					resourceIds
				}
			}
		}
	`

	var nodeResult struct {
		Node struct {
			ID          string   `json:"id"`
			ResourceIDs []string `json:"resourceIds"`
		} `json:"node"`
	}

	err = owner.Execute(nodeQuery, map[string]any{"id": tagID}, &nodeResult)
	require.NoError(t, err)
	assert.Contains(t, nodeResult.Node.ResourceIDs, documentID)

	const detachMutation = `
		mutation($input: DetachResourceTagInput!) {
			detachResourceTag(input: $input) {
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
	assert.Empty(t, forResourceResult.ResourceTagsForResource)
}

func TestResourceTag_DuplicateKeyConflict(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	const createMutation = `
		mutation($input: CreateResourceTagInput!) {
			createResourceTag(input: $input) {
				resourceTagEdge { node { id } }
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
