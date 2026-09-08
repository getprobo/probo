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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

type taskActivityNode struct {
	ID           string  `json:"id"`
	ActivityType string  `json:"activityType"`
	Field        *string `json:"field"`
	OldValue     *string `json:"oldValue"`
	NewValue     *string `json:"newValue"`
	Actor        *struct {
		ID string `json:"id"`
	} `json:"actor"`
}

type taskActivitiesResult struct {
	Node struct {
		Activities struct {
			TotalCount int `json:"totalCount"`
			Edges      []struct {
				Node taskActivityNode `json:"node"`
			} `json:"edges"`
		} `json:"activities"`
	} `json:"node"`
}

const taskActivitiesQuery = `
	query($id: ID!) {
		node(id: $id) {
			... on Task {
				activities(first: 20, orderBy: { field: CREATED_AT, direction: DESC }) {
					totalCount
					edges {
						node {
							id
							activityType
							field
							oldValue
							newValue
							actor {
								id
							}
						}
					}
				}
			}
		}
	}
`

func listTaskActivities(t *testing.T, client *testutil.Client, taskID string) taskActivitiesResult {
	t.Helper()

	var result taskActivitiesResult

	err := client.Execute(taskActivitiesQuery, map[string]any{"id": taskID}, &result)
	require.NoError(t, err)

	return result
}

func updateTask(t *testing.T, client *testutil.Client, input map[string]any) {
	t.Helper()

	query := `
		mutation UpdateTask($input: UpdateTaskInput!) {
			updateTask(input: $input) {
				task {
					id
				}
			}
		}
	`

	var result struct {
		UpdateTask struct {
			Task struct {
				ID string `json:"id"`
			} `json:"task"`
		} `json:"updateTask"`
	}

	err := client.Execute(query, map[string]any{"input": input}, &result)
	require.NoError(t, err)
	assert.Equal(t, input["taskId"], result.UpdateTask.Task.ID)
}

func TestTaskActivity_CreateWritesCreated(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	taskID := factory.NewTaskWithoutMeasure(owner).WithName("Created task").Create()

	result := listTaskActivities(t, owner, taskID)
	require.Equal(t, 1, result.Node.Activities.TotalCount)
	require.Len(t, result.Node.Activities.Edges, 1)

	activity := result.Node.Activities.Edges[0].Node
	assert.Equal(t, "CREATED", activity.ActivityType)
	assert.Nil(t, activity.Field)
	assert.Nil(t, activity.OldValue)
	assert.Nil(t, activity.NewValue)
	require.NotNil(t, activity.Actor)
	assert.Equal(t, owner.GetProfileID().String(), activity.Actor.ID)
}

func TestTaskActivity_UpdateStateWritesDiff(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()

	updateTask(t, owner, map[string]any{
		"taskId": taskID,
		"state":  "IN_PROGRESS",
	})

	result := listTaskActivities(t, owner, taskID)
	require.GreaterOrEqual(t, result.Node.Activities.TotalCount, 2)

	var stateActivity *taskActivityNode

	for _, edge := range result.Node.Activities.Edges {
		if edge.Node.ActivityType == "UPDATED" && edge.Node.Field != nil && *edge.Node.Field == "STATE" {
			stateActivity = &edge.Node
			break
		}
	}

	require.NotNil(t, stateActivity)
	require.NotNil(t, stateActivity.OldValue)
	require.NotNil(t, stateActivity.NewValue)
	assert.Equal(t, "TODO", *stateActivity.OldValue)
	assert.Equal(t, "IN_PROGRESS", *stateActivity.NewValue)
	require.NotNil(t, stateActivity.Actor)
	assert.Equal(t, owner.GetProfileID().String(), stateActivity.Actor.ID)
}

func TestTaskActivity_UpdateAssigneeWritesDiff(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()
	profileID := factory.CreateUser(owner)

	updateTask(t, owner, map[string]any{
		"taskId":       taskID,
		"assignedToId": profileID,
	})

	result := listTaskActivities(t, owner, taskID)

	var assigneeActivity *taskActivityNode

	for _, edge := range result.Node.Activities.Edges {
		if edge.Node.ActivityType == "UPDATED" && edge.Node.Field != nil && *edge.Node.Field == "ASSIGNED_TO" {
			assigneeActivity = &edge.Node
			break
		}
	}

	require.NotNil(t, assigneeActivity)
	assert.Nil(t, assigneeActivity.OldValue)
	require.NotNil(t, assigneeActivity.NewValue)
	assert.NotEmpty(t, *assigneeActivity.NewValue)
}

func TestTaskActivity_UpdateDescriptionWritesValues(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	taskID := factory.NewTaskWithoutMeasure(owner).
		WithContent("Original description").
		Create()

	updateTask(t, owner, map[string]any{
		"taskId":  taskID,
		"content": factory.ProseMirrorPlainText("A new description"),
	})

	result := listTaskActivities(t, owner, taskID)

	var descriptionActivity *taskActivityNode

	for _, edge := range result.Node.Activities.Edges {
		if edge.Node.ActivityType == "UPDATED" && edge.Node.Field != nil && *edge.Node.Field == "DESCRIPTION" {
			descriptionActivity = &edge.Node
			break
		}
	}

	require.NotNil(t, descriptionActivity)
	require.NotNil(t, descriptionActivity.OldValue)
	require.NotNil(t, descriptionActivity.NewValue)
	factory.AssertProseMirrorPlainText(t, "Original description", *descriptionActivity.OldValue)
	factory.AssertProseMirrorPlainText(t, "A new description", *descriptionActivity.NewValue)
}

func TestTaskActivity_RankOnlyUpdateWritesNothing(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()

	before := listTaskActivities(t, owner, taskID)
	require.Equal(t, 1, before.Node.Activities.TotalCount)

	updateTask(t, owner, map[string]any{
		"taskId": taskID,
		"rank":   1,
	})

	after := listTaskActivities(t, owner, taskID)
	assert.Equal(t, 1, after.Node.Activities.TotalCount)
	assert.Equal(t, "CREATED", after.Node.Activities.Edges[0].Node.ActivityType)
}

func TestTaskActivity_ListNewestFirst(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()

	updateTask(t, owner, map[string]any{
		"taskId": taskID,
		"state":  "IN_PROGRESS",
	})
	updateTask(t, owner, map[string]any{
		"taskId":   taskID,
		"priority": "HIGH",
	})

	result := listTaskActivities(t, owner, taskID)
	require.GreaterOrEqual(t, len(result.Node.Activities.Edges), 3)

	assert.Equal(t, "UPDATED", result.Node.Activities.Edges[0].Node.ActivityType)
	require.NotNil(t, result.Node.Activities.Edges[0].Node.Field)
	assert.Equal(t, "PRIORITY", *result.Node.Activities.Edges[0].Node.Field)

	assert.Equal(t, "UPDATED", result.Node.Activities.Edges[1].Node.ActivityType)
	require.NotNil(t, result.Node.Activities.Edges[1].Node.Field)
	assert.Equal(t, "STATE", *result.Node.Activities.Edges[1].Node.Field)

	assert.Equal(t, "CREATED", result.Node.Activities.Edges[2].Node.ActivityType)
}

func TestTaskActivity_ImportReuseWritesFieldDiffs(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	measureRef := fmt.Sprintf("measure-%s", owner.GetOrganizationID())
	taskRef := fmt.Sprintf("task-%s", owner.GetOrganizationID())

	first := importMeasureTasks(t, owner, measureRef, taskRef, "Imported task", "First description")
	require.Len(t, first, 1)

	created := listTaskActivities(t, owner, first[0].ID)
	require.Equal(t, 1, created.Node.Activities.TotalCount)
	assert.Equal(t, "CREATED", created.Node.Activities.Edges[0].Node.ActivityType)
	require.NotNil(t, created.Node.Activities.Edges[0].Node.Actor)
	assert.Equal(t, owner.GetProfileID().String(), created.Node.Activities.Edges[0].Node.Actor.ID)

	second := importMeasureTasks(t, owner, measureRef, taskRef, "Renamed import task", "Second description")
	require.Len(t, second, 1)
	assert.Equal(t, first[0].ID, second[0].ID)
	assert.Equal(t, "Renamed import task", second[0].Name)

	result := listTaskActivities(t, owner, second[0].ID)
	require.GreaterOrEqual(t, result.Node.Activities.TotalCount, 3)

	var nameActivity *taskActivityNode

	var descriptionActivity *taskActivityNode

	for _, edge := range result.Node.Activities.Edges {
		if edge.Node.ActivityType != "UPDATED" || edge.Node.Field == nil {
			continue
		}

		switch *edge.Node.Field {
		case "NAME":
			nameActivity = &edge.Node
		case "DESCRIPTION":
			descriptionActivity = &edge.Node
		}
	}

	require.NotNil(t, nameActivity)
	require.NotNil(t, nameActivity.OldValue)
	require.NotNil(t, nameActivity.NewValue)
	assert.Equal(t, "Imported task", *nameActivity.OldValue)
	assert.Equal(t, "Renamed import task", *nameActivity.NewValue)
	require.NotNil(t, nameActivity.Actor)
	assert.Equal(t, owner.GetProfileID().String(), nameActivity.Actor.ID)

	require.NotNil(t, descriptionActivity)
	require.NotNil(t, descriptionActivity.OldValue)
	require.NotNil(t, descriptionActivity.NewValue)
	factory.AssertProseMirrorPlainText(t, "First description", *descriptionActivity.OldValue)
	factory.AssertProseMirrorPlainText(t, "Second description", *descriptionActivity.NewValue)
}

func importMeasureTasks(
	t *testing.T,
	client *testutil.Client,
	measureRef string,
	taskRef string,
	taskName string,
	taskDescription string,
) []struct {
	ID   string
	Name string
} {
	t.Helper()

	payload := fmt.Sprintf(
		`[{
			"name": "Imported measure",
			"category": "Security",
			"reference-id": %q,
			"tasks": [{
				"name": %q,
				"description": %q,
				"reference-id": %q
			}]
		}]`,
		measureRef,
		taskName,
		taskDescription,
		taskRef,
	)

	query := `
		mutation ImportMeasure($input: ImportMeasureInput!) {
			importMeasure(input: $input) {
				measureEdges {
					node {
						id
						tasks(first: 10) {
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
		}
	`

	var result struct {
		ImportMeasure struct {
			MeasureEdges []struct {
				Node struct {
					ID    string `json:"id"`
					Tasks struct {
						Edges []struct {
							Node struct {
								ID   string `json:"id"`
								Name string `json:"name"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"tasks"`
				} `json:"node"`
			} `json:"measureEdges"`
		} `json:"importMeasure"`
	}

	err := client.ExecuteWithFile(
		query,
		map[string]any{
			"input": map[string]any{
				"organizationId": client.GetOrganizationID(),
				"file":           nil,
			},
		},
		"input.file",
		testutil.UploadFile{
			Filename:    "measures.json",
			ContentType: "application/json",
			Content:     []byte(payload),
		},
		&result,
	)
	require.NoError(t, err)
	require.Len(t, result.ImportMeasure.MeasureEdges, 1)

	tasks := result.ImportMeasure.MeasureEdges[0].Node.Tasks.Edges
	imported := make([]struct {
		ID   string
		Name string
	}, 0, len(tasks))

	for _, edge := range tasks {
		imported = append(imported, struct {
			ID   string
			Name string
		}{ID: edge.Node.ID, Name: edge.Node.Name})
	}

	return imported
}

func TestTaskActivity_ViewerCanList(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()

	result := listTaskActivities(t, viewer, taskID)
	require.Equal(t, 1, result.Node.Activities.TotalCount)
	assert.Equal(t, "CREATED", result.Node.Activities.Edges[0].Node.ActivityType)
}

func TestTaskActivity_AuditorCannotAccess(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	auditor := testutil.NewClientInOrg(t, testutil.RoleAuditor, owner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()

	activities := listTaskActivities(t, owner, taskID)
	require.Len(t, activities.Node.Activities.Edges, 1)
	activityID := activities.Node.Activities.Edges[0].Node.ID

	t.Run("cannot list", func(t *testing.T) {
		t.Parallel()

		_, err := auditor.Do(taskActivitiesQuery, map[string]any{"id": taskID})
		require.Error(t, err)
	})

	t.Run("cannot get", func(t *testing.T) {
		t.Parallel()

		query := `
			query($id: ID!) {
				node(id: $id) {
					... on TaskActivity {
						id
						activityType
					}
				}
			}
		`

		var result struct {
			Node *struct {
				ID string `json:"id"`
			} `json:"node"`
		}

		err := auditor.Execute(query, map[string]any{"id": activityID}, &result)
		testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "task activity")
	})
}
