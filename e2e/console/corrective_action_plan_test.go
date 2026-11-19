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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const (
	createCorrectiveActionPlanMutation = `
		mutation($input: CreateCorrectiveActionPlanInput!) {
			createCorrectiveActionPlan(input: $input) {
				correctiveActionPlanEdge {
					node {
						id
						title
						description
						dueDate
						owner { id }
						finding { id }
						createdAt
						updatedAt
					}
				}
			}
		}
	`

	updateCorrectiveActionPlanMutation = `
		mutation($input: UpdateCorrectiveActionPlanInput!) {
			updateCorrectiveActionPlan(input: $input) {
				correctiveActionPlan {
					id
					title
					description
					dueDate
					owner { id }
					createdAt
					updatedAt
				}
			}
		}
	`

	deleteCorrectiveActionPlanMutation = `
		mutation($input: DeleteCorrectiveActionPlanInput!) {
			deleteCorrectiveActionPlan(input: $input) {
				deletedCorrectiveActionPlanId
			}
		}
	`

	createCorrectiveActionPlanInternalControlMappingMutation = `
		mutation($input: CreateCorrectiveActionPlanInternalControlMappingInput!) {
			createCorrectiveActionPlanInternalControlMapping(input: $input) {
				correctiveActionPlanEdge { node { id } }
				internalControlEdge { node { id } }
			}
		}
	`

	deleteCorrectiveActionPlanInternalControlMappingMutation = `
		mutation($input: DeleteCorrectiveActionPlanInternalControlMappingInput!) {
			deleteCorrectiveActionPlanInternalControlMapping(input: $input) {
				deletedCorrectiveActionPlanId
				deletedInternalControlId
			}
		}
	`

	correctiveActionPlanNodeQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on CorrectiveActionPlan {
					id
					title
					finding { id }
					internalControls(first: 10) {
						totalCount
						edges { node { id } }
					}
					tasks(first: 10) {
						totalCount
						edges { node { id } }
					}
				}
			}
		}
	`

	findingCorrectiveActionPlansQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on Finding {
					correctiveActionPlans(first: 10) {
						totalCount
						edges { node { id title } }
					}
					tasks(first: 10) {
						totalCount
						edges { node { id } }
					}
				}
			}
		}
	`

	taskFindingsQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on Task {
					findings(first: 10) {
						totalCount
						edges { node { id } }
					}
				}
			}
		}
	`
)

type (
	idNode struct {
		ID string `json:"id"`
	}

	idConnection struct {
		TotalCount int `json:"totalCount"`
		Edges      []struct {
			Node idNode `json:"node"`
		} `json:"edges"`
	}

	correctiveActionPlanNode struct {
		ID          string  `json:"id"`
		Title       string  `json:"title"`
		Description *string `json:"description"`
		DueDate     *string `json:"dueDate"`
		Owner       *idNode `json:"owner"`
		Finding     idNode  `json:"finding"`
		CreatedAt   string  `json:"createdAt"`
		UpdatedAt   string  `json:"updatedAt"`
	}
)

func createFindingForCorrectiveActionPlan(t *testing.T, c *testutil.Client) string {
	t.Helper()

	var result struct {
		CreateFinding struct {
			FindingEdge struct {
				Node idNode `json:"node"`
			} `json:"findingEdge"`
		} `json:"createFinding"`
	}

	err := c.Execute(`
		mutation($input: CreateFindingInput!) {
			createFinding(input: $input) {
				findingEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": c.GetOrganizationID().String(),
			"kind":           "MINOR_NONCONFORMITY",
			"title":          factory.SafeName("Finding"),
			"ownerId":        factory.CreateUser(c),
			"status":         "OPEN",
			"priority":       "HIGH",
		},
	}, &result)
	require.NoError(t, err)

	return result.CreateFinding.FindingEdge.Node.ID
}

func createCorrectiveActionPlan(t *testing.T, c *testutil.Client, findingID string, input map[string]any) correctiveActionPlanNode {
	t.Helper()

	if input == nil {
		input = map[string]any{}
	}
	input["findingId"] = findingID
	if _, ok := input["title"]; !ok {
		input["title"] = factory.SafeName("Plan")
	}

	var result struct {
		CreateCorrectiveActionPlan struct {
			CorrectiveActionPlanEdge struct {
				Node correctiveActionPlanNode `json:"node"`
			} `json:"correctiveActionPlanEdge"`
		} `json:"createCorrectiveActionPlan"`
	}

	err := c.Execute(createCorrectiveActionPlanMutation, map[string]any{"input": input}, &result)
	require.NoError(t, err)

	return result.CreateCorrectiveActionPlan.CorrectiveActionPlanEdge.Node
}

func linkCorrectiveActionPlanInternalControl(t *testing.T, c *testutil.Client, planID string, internalControlID string) {
	t.Helper()

	err := c.Execute(createCorrectiveActionPlanInternalControlMappingMutation, map[string]any{
		"input": map[string]any{
			"correctiveActionPlanId": planID,
			"internalControlId":      internalControlID,
		},
	}, nil)
	require.NoError(t, err)
}

func TestCorrectiveActionPlan_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	findingID := createFindingForCorrectiveActionPlan(t, owner)

	t.Run("with required fields", func(t *testing.T) {
		t.Parallel()

		plan := createCorrectiveActionPlan(t, owner, findingID, map[string]any{"title": "Restore access reviews"})

		assert.NotEmpty(t, plan.ID)
		assert.Equal(t, "Restore access reviews", plan.Title)
		assert.Nil(t, plan.Description)
		assert.Nil(t, plan.DueDate)
		assert.Nil(t, plan.Owner)
		assert.Equal(t, findingID, plan.Finding.ID)
		assert.Equal(t, plan.CreatedAt, plan.UpdatedAt)
	})

	t.Run("with all fields", func(t *testing.T) {
		t.Parallel()

		profileID := factory.CreateUser(owner)
		plan := createCorrectiveActionPlan(t, owner, findingID, map[string]any{
			"title":       "Rotate credentials",
			"description": "Rotate all shared credentials",
			"ownerId":     profileID,
			"dueDate":     "2027-01-15T00:00:00Z",
		})

		require.NotNil(t, plan.Description)
		assert.Equal(t, "Rotate all shared credentials", *plan.Description)
		require.NotNil(t, plan.Owner)
		assert.Equal(t, profileID, plan.Owner.ID)
		require.NotNil(t, plan.DueDate)
		assert.Contains(t, *plan.DueDate, "2027-01-15")
	})

	t.Run("rejects HTML in title", func(t *testing.T) {
		t.Parallel()

		_, err := owner.Do(createCorrectiveActionPlanMutation, map[string]any{
			"input": map[string]any{
				"findingId": findingID,
				"title":     "<script>alert(1)</script>",
			},
		})
		require.Error(t, err)
	})

	t.Run("rejects empty title", func(t *testing.T) {
		t.Parallel()

		_, err := owner.Do(createCorrectiveActionPlanMutation, map[string]any{
			"input": map[string]any{
				"findingId": findingID,
				"title":     "",
			},
		})
		require.Error(t, err)
	})
}

func TestCorrectiveActionPlan_UpdateAndDelete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	findingID := createFindingForCorrectiveActionPlan(t, owner)
	profileID := factory.CreateUser(owner)

	plan := createCorrectiveActionPlan(t, owner, findingID, map[string]any{
		"description": "Initial",
		"ownerId":     profileID,
		"dueDate":     "2027-01-15T00:00:00Z",
	})

	var updated struct {
		UpdateCorrectiveActionPlan struct {
			CorrectiveActionPlan correctiveActionPlanNode `json:"correctiveActionPlan"`
		} `json:"updateCorrectiveActionPlan"`
	}

	err := owner.Execute(updateCorrectiveActionPlanMutation, map[string]any{
		"input": map[string]any{
			"id":          plan.ID,
			"title":       "Updated plan",
			"description": nil,
			"ownerId":     nil,
			"dueDate":     nil,
		},
	}, &updated)
	require.NoError(t, err)

	got := updated.UpdateCorrectiveActionPlan.CorrectiveActionPlan
	assert.Equal(t, "Updated plan", got.Title)
	assert.Nil(t, got.Description)
	assert.Nil(t, got.Owner)
	assert.Nil(t, got.DueDate)
	origCreatedAt, err := time.Parse(time.RFC3339Nano, plan.CreatedAt)
	require.NoError(t, err)
	gotCreatedAt, err := time.Parse(time.RFC3339Nano, got.CreatedAt)
	require.NoError(t, err)
	gotUpdatedAt, err := time.Parse(time.RFC3339Nano, got.UpdatedAt)
	require.NoError(t, err)
	assert.WithinDuration(t, origCreatedAt, gotCreatedAt, time.Microsecond)
	assert.True(t, gotUpdatedAt.After(gotCreatedAt), "updatedAt advances on update")

	var deleted struct {
		DeleteCorrectiveActionPlan struct {
			DeletedCorrectiveActionPlanID string `json:"deletedCorrectiveActionPlanId"`
		} `json:"deleteCorrectiveActionPlan"`
	}

	err = owner.Execute(deleteCorrectiveActionPlanMutation, map[string]any{
		"input": map[string]any{"correctiveActionPlanId": plan.ID},
	}, &deleted)
	require.NoError(t, err)
	assert.Equal(t, plan.ID, deleted.DeleteCorrectiveActionPlan.DeletedCorrectiveActionPlanID)

	var listed struct {
		Node struct {
			CorrectiveActionPlans idConnection `json:"correctiveActionPlans"`
		} `json:"node"`
	}

	err = owner.Execute(findingCorrectiveActionPlansQuery, map[string]any{"id": findingID}, &listed)
	require.NoError(t, err)
	assert.Equal(t, 0, listed.Node.CorrectiveActionPlans.TotalCount)
}

func TestCorrectiveActionPlan_InternalControlsAndDerivedTasks(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	findingID := createFindingForCorrectiveActionPlan(t, owner)
	internalControlID := factory.CreateInternalControl(owner)
	taskID := factory.CreateTask(owner, &internalControlID)
	unrelatedTaskID := factory.CreateTask(owner, nil)

	plan := createCorrectiveActionPlan(t, owner, findingID, nil)
	linkCorrectiveActionPlanInternalControl(t, owner, plan.ID, internalControlID)

	_, err := owner.Do(createCorrectiveActionPlanInternalControlMappingMutation, map[string]any{
		"input": map[string]any{
			"correctiveActionPlanId": plan.ID,
			"internalControlId":      internalControlID,
		},
	})
	testutil.RequireErrorCode(t, err, "CONFLICT", "linking the same internal control twice")

	var planNode struct {
		Node struct {
			Finding          idNode       `json:"finding"`
			InternalControls idConnection `json:"internalControls"`
			Tasks            idConnection `json:"tasks"`
		} `json:"node"`
	}

	err = owner.Execute(correctiveActionPlanNodeQuery, map[string]any{"id": plan.ID}, &planNode)
	require.NoError(t, err)
	assert.Equal(t, findingID, planNode.Node.Finding.ID)
	assert.Equal(t, 1, planNode.Node.InternalControls.TotalCount)
	require.Len(t, planNode.Node.InternalControls.Edges, 1)
	assert.Equal(t, internalControlID, planNode.Node.InternalControls.Edges[0].Node.ID)
	assert.Equal(t, 1, planNode.Node.Tasks.TotalCount)
	require.Len(t, planNode.Node.Tasks.Edges, 1)
	assert.Equal(t, taskID, planNode.Node.Tasks.Edges[0].Node.ID)

	var findingNode struct {
		Node struct {
			CorrectiveActionPlans idConnection `json:"correctiveActionPlans"`
			Tasks                 idConnection `json:"tasks"`
		} `json:"node"`
	}

	err = owner.Execute(findingCorrectiveActionPlansQuery, map[string]any{"id": findingID}, &findingNode)
	require.NoError(t, err)
	assert.Equal(t, 1, findingNode.Node.CorrectiveActionPlans.TotalCount)
	assert.Equal(t, 1, findingNode.Node.Tasks.TotalCount)
	require.Len(t, findingNode.Node.Tasks.Edges, 1)
	assert.Equal(t, taskID, findingNode.Node.Tasks.Edges[0].Node.ID)

	var taskNode struct {
		Node struct {
			Findings idConnection `json:"findings"`
		} `json:"node"`
	}

	err = owner.Execute(taskFindingsQuery, map[string]any{"id": taskID}, &taskNode)
	require.NoError(t, err)
	assert.Equal(t, 1, taskNode.Node.Findings.TotalCount)
	require.Len(t, taskNode.Node.Findings.Edges, 1)
	assert.Equal(t, findingID, taskNode.Node.Findings.Edges[0].Node.ID)

	err = owner.Execute(taskFindingsQuery, map[string]any{"id": unrelatedTaskID}, &taskNode)
	require.NoError(t, err)
	assert.Equal(t, 0, taskNode.Node.Findings.TotalCount)

	var internalControlNode struct {
		Node struct {
			CorrectiveActionPlans idConnection `json:"correctiveActionPlans"`
		} `json:"node"`
	}

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on InternalControl {
					correctiveActionPlans(first: 10) {
						totalCount
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{"id": internalControlID}, &internalControlNode)
	require.NoError(t, err)
	assert.Equal(t, 1, internalControlNode.Node.CorrectiveActionPlans.TotalCount)
	require.Len(t, internalControlNode.Node.CorrectiveActionPlans.Edges, 1)
	assert.Equal(t, plan.ID, internalControlNode.Node.CorrectiveActionPlans.Edges[0].Node.ID)

	var unlinked struct {
		DeleteCorrectiveActionPlanInternalControlMapping struct {
			DeletedCorrectiveActionPlanID string `json:"deletedCorrectiveActionPlanId"`
			DeletedInternalControlID      string `json:"deletedInternalControlId"`
		} `json:"deleteCorrectiveActionPlanInternalControlMapping"`
	}

	err = owner.Execute(deleteCorrectiveActionPlanInternalControlMappingMutation, map[string]any{
		"input": map[string]any{
			"correctiveActionPlanId": plan.ID,
			"internalControlId":      internalControlID,
		},
	}, &unlinked)
	require.NoError(t, err)
	assert.Equal(t, plan.ID, unlinked.DeleteCorrectiveActionPlanInternalControlMapping.DeletedCorrectiveActionPlanID)
	assert.Equal(t, internalControlID, unlinked.DeleteCorrectiveActionPlanInternalControlMapping.DeletedInternalControlID)

	err = owner.Execute(findingCorrectiveActionPlansQuery, map[string]any{"id": findingID}, &findingNode)
	require.NoError(t, err)
	assert.Equal(t, 0, findingNode.Node.Tasks.TotalCount)

	err = owner.Execute(taskFindingsQuery, map[string]any{"id": taskID}, &taskNode)
	require.NoError(t, err)
	assert.Equal(t, 0, taskNode.Node.Findings.TotalCount)
}

func TestCorrectiveActionPlan_RBAC(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	findingID := createFindingForCorrectiveActionPlan(t, owner)
	plan := createCorrectiveActionPlan(t, owner, findingID, nil)
	internalControlID := factory.CreateInternalControl(owner)

	t.Run("admin can create", func(t *testing.T) {
		t.Parallel()

		created := createCorrectiveActionPlan(t, admin, findingID, nil)
		assert.NotEmpty(t, created.ID)
	})

	t.Run("viewer can read", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Node *struct {
				ID string `json:"id"`
			} `json:"node"`
		}

		err := viewer.Execute(correctiveActionPlanNodeQuery, map[string]any{"id": plan.ID}, &result)
		require.NoError(t, err)
		require.NotNil(t, result.Node)
		assert.Equal(t, plan.ID, result.Node.ID)
	})

	t.Run("viewer cannot create", func(t *testing.T) {
		t.Parallel()

		_, err := viewer.Do(createCorrectiveActionPlanMutation, map[string]any{
			"input": map[string]any{"findingId": findingID, "title": "Nope"},
		})
		testutil.RequireForbiddenError(t, err, "viewer cannot create")
	})

	t.Run("viewer cannot update", func(t *testing.T) {
		t.Parallel()

		_, err := viewer.Do(updateCorrectiveActionPlanMutation, map[string]any{
			"input": map[string]any{"id": plan.ID, "title": "Nope"},
		})
		testutil.RequireForbiddenError(t, err, "viewer cannot update")
	})

	t.Run("viewer cannot delete", func(t *testing.T) {
		t.Parallel()

		_, err := viewer.Do(deleteCorrectiveActionPlanMutation, map[string]any{
			"input": map[string]any{"correctiveActionPlanId": plan.ID},
		})
		testutil.RequireForbiddenError(t, err, "viewer cannot delete")
	})

	t.Run("viewer cannot link internal control", func(t *testing.T) {
		t.Parallel()

		_, err := viewer.Do(createCorrectiveActionPlanInternalControlMappingMutation, map[string]any{
			"input": map[string]any{
				"correctiveActionPlanId": plan.ID,
				"internalControlId":      internalControlID,
			},
		})
		testutil.RequireForbiddenError(t, err, "viewer cannot link internal control")
	})
}

func TestCorrectiveActionPlan_TenantIsolation(t *testing.T) {
	t.Parallel()
	owner1 := testutil.NewClient(t, testutil.RoleOwner)
	owner2 := testutil.NewClient(t, testutil.RoleOwner)
	findingID := createFindingForCorrectiveActionPlan(t, owner1)
	plan := createCorrectiveActionPlan(t, owner1, findingID, nil)
	otherFindingID := createFindingForCorrectiveActionPlan(t, owner2)
	otherPlan := createCorrectiveActionPlan(t, owner2, otherFindingID, nil)
	otherInternalControlID := factory.CreateInternalControl(owner2)
	otherProfileID := factory.CreateUser(owner2)

	t.Run("other org cannot read plan", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Node *struct {
				ID string `json:"id"`
			} `json:"node"`
		}

		err := owner2.Execute(correctiveActionPlanNodeQuery, map[string]any{"id": plan.ID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "CorrectiveActionPlan")
	})

	t.Run("other org cannot create plan on finding", func(t *testing.T) {
		t.Parallel()

		_, err := owner2.Do(createCorrectiveActionPlanMutation, map[string]any{
			"input": map[string]any{"findingId": findingID, "title": "Nope"},
		})
		require.Error(t, err)
	})

	t.Run("other org cannot delete plan", func(t *testing.T) {
		t.Parallel()

		_, err := owner2.Do(deleteCorrectiveActionPlanMutation, map[string]any{
			"input": map[string]any{"correctiveActionPlanId": plan.ID},
		})
		require.Error(t, err)
	})

	t.Run("cannot link internal control of another org", func(t *testing.T) {
		t.Parallel()

		_, err := owner1.Do(createCorrectiveActionPlanInternalControlMappingMutation, map[string]any{
			"input": map[string]any{
				"correctiveActionPlanId": plan.ID,
				"internalControlId":      otherInternalControlID,
			},
		})
		require.Error(t, err)
	})

	t.Run("cannot link own internal control to another org plan", func(t *testing.T) {
		t.Parallel()

		ownInternalControlID := factory.CreateInternalControl(owner1)

		_, err := owner1.Do(createCorrectiveActionPlanInternalControlMappingMutation, map[string]any{
			"input": map[string]any{
				"correctiveActionPlanId": otherPlan.ID,
				"internalControlId":      ownInternalControlID,
			},
		})
		require.Error(t, err)
	})

	t.Run("cannot assign owner from another org", func(t *testing.T) {
		t.Parallel()

		_, err := owner1.Do(createCorrectiveActionPlanMutation, map[string]any{
			"input": map[string]any{
				"findingId": findingID,
				"title":     "Cross org owner",
				"ownerId":   otherProfileID,
			},
		})
		require.Error(t, err)
	})
}
