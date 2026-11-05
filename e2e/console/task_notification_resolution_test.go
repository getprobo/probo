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

func TestTask_ResolveLinkedNotifications(t *testing.T) {
	t.Parallel()

	for _, state := range []struct {
		name  string
		value string
	}{
		{name: "done", value: "DONE"},
		{name: "canceled", value: "CANCELED"},
		{name: "duplicate", value: "DUPLICATE"},
	} {
		t.Run(state.name+" resolves assignment and mention notifications", func(t *testing.T) {
			t.Parallel()

			owner, assignee, mentioned, taskID, otherTaskID := seedTaskNotifications(t)

			require.Equal(t, []string{"PENDING"}, taskNotificationStatuses(t, assignee, taskID, "TASK_ASSIGNED"))
			require.Equal(t, []string{"PENDING"}, taskNotificationStatuses(t, mentioned, taskID, "TASK_MENTIONED"))
			require.Equal(t, []string{"PENDING"}, taskNotificationStatuses(t, assignee, otherTaskID, "TASK_ASSIGNED"))

			updateTask(t, owner, map[string]any{
				"taskId": taskID,
				"state":  state.value,
			})

			assert.Equal(t, []string{"RESOLVED"}, taskNotificationStatuses(t, assignee, taskID, "TASK_ASSIGNED"))
			assert.Equal(t, []string{"RESOLVED"}, taskNotificationStatuses(t, mentioned, taskID, "TASK_MENTIONED"))
			assert.Equal(t, []string{"PENDING"}, taskNotificationStatuses(t, assignee, otherTaskID, "TASK_ASSIGNED"))
		})
	}

	t.Run("deletion resolves assignment and mention notifications", func(t *testing.T) {
		t.Parallel()

		owner, assignee, mentioned, taskID, otherTaskID := seedTaskNotifications(t)

		deleteTask(t, owner, taskID)

		assert.Equal(t, []string{"RESOLVED"}, taskNotificationStatuses(t, assignee, taskID, "TASK_ASSIGNED"))
		assert.Equal(t, []string{"RESOLVED"}, taskNotificationStatuses(t, mentioned, taskID, "TASK_MENTIONED"))
		assert.Equal(t, []string{"PENDING"}, taskNotificationStatuses(t, assignee, otherTaskID, "TASK_ASSIGNED"))
	})

	t.Run("in progress leaves notifications pending", func(t *testing.T) {
		t.Parallel()

		owner, assignee, mentioned, taskID, _ := seedTaskNotifications(t)

		updateTask(t, owner, map[string]any{
			"taskId": taskID,
			"state":  "IN_PROGRESS",
		})

		assert.Equal(t, []string{"PENDING"}, taskNotificationStatuses(t, assignee, taskID, "TASK_ASSIGNED"))
		assert.Equal(t, []string{"PENDING"}, taskNotificationStatuses(t, mentioned, taskID, "TASK_MENTIONED"))
	})

	t.Run("closing and reassigning resolves the new assignment", func(t *testing.T) {
		t.Parallel()

		owner, assignee, _, taskID, _ := seedTaskNotifications(t)
		nextAssignee := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)

		updateTask(t, owner, map[string]any{
			"taskId":       taskID,
			"state":        "DONE",
			"assignedToId": nextAssignee.GetProfileID().String(),
		})

		assert.Equal(t, []string{"RESOLVED"}, taskNotificationStatuses(t, assignee, taskID, "TASK_ASSIGNED"))
		assert.Equal(t, []string{"RESOLVED"}, taskNotificationStatuses(t, nextAssignee, taskID, "TASK_ASSIGNED"))
	})

	t.Run("reassigning an already closed task leaves the new assignment resolved", func(t *testing.T) {
		t.Parallel()

		owner, _, _, taskID, _ := seedTaskNotifications(t)
		nextAssignee := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)

		updateTask(t, owner, map[string]any{
			"taskId": taskID,
			"state":  "CANCELED",
		})
		updateTask(t, owner, map[string]any{
			"taskId":       taskID,
			"assignedToId": nextAssignee.GetProfileID().String(),
		})

		assert.Equal(t, []string{"RESOLVED"}, taskNotificationStatuses(t, nextAssignee, taskID, "TASK_ASSIGNED"))
	})

	t.Run("mention on an already closed task is resolved", func(t *testing.T) {
		t.Parallel()

		owner, _, _, taskID, _ := seedTaskNotifications(t)
		late := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)

		updateTask(t, owner, map[string]any{
			"taskId": taskID,
			"state":  "DUPLICATE",
		})
		createTaskCommentWithContent(t, owner, taskID, mentionContent(t, late.GetProfileID().String(), "Late"))
		updateTaskDescription(t, owner, taskID, mentionContent(t, late.GetProfileID().String(), "Late"))

		assert.Equal(t, []string{"RESOLVED", "RESOLVED"}, taskNotificationStatuses(t, late, taskID, "TASK_MENTIONED"))
	})

	t.Run("creating a closed task resolves its notifications", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		assignee := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
		mentioned := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)

		taskID := createClosedTask(t, owner, assignee.GetProfileID().String(), mentionContent(t, mentioned.GetProfileID().String(), "Mentioned"))

		assert.Equal(t, []string{"RESOLVED"}, taskNotificationStatuses(t, assignee, taskID, "TASK_ASSIGNED"))
		assert.Equal(t, []string{"RESOLVED"}, taskNotificationStatuses(t, mentioned, taskID, "TASK_MENTIONED"))
	})
}

func seedTaskNotifications(t *testing.T) (
	owner *testutil.Client,
	assignee *testutil.Client,
	mentioned *testutil.Client,
	taskID string,
	otherTaskID string,
) {
	t.Helper()

	owner = testutil.NewClient(t, testutil.RoleOwner)
	assignee = testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
	mentioned = testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)

	taskID = factory.NewTaskWithoutInternalControl(owner).WithName("Notification task").Create()
	otherTaskID = factory.NewTaskWithoutInternalControl(owner).WithName("Other notification task").Create()

	assignTask(t, owner, taskID, assignee.GetProfileID().String())
	assignTask(t, owner, otherTaskID, assignee.GetProfileID().String())
	createTaskCommentWithContent(t, owner, taskID, mentionContent(t, mentioned.GetProfileID().String(), "Mentioned"))

	return owner, assignee, mentioned, taskID, otherTaskID
}

func createClosedTask(t *testing.T, client *testutil.Client, assigneeID string, content string) string {
	t.Helper()

	const query = `
		mutation($input: CreateTaskInput!) {
			createTask(input: $input) {
				taskEdge {
					node { id }
				}
			}
		}
	`

	var result struct {
		CreateTask struct {
			TaskEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"taskEdge"`
		} `json:"createTask"`
	}

	err := client.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": client.GetOrganizationID().String(),
			"name":           "Created closed",
			"priority":       "MEDIUM",
			"state":          "DONE",
			"assignedToId":   assigneeID,
			"content":        content,
		},
	}, &result)
	require.NoError(t, err)

	return result.CreateTask.TaskEdge.Node.ID
}

func deleteTask(t *testing.T, client *testutil.Client, taskID string) {
	t.Helper()

	const query = `
		mutation($input: DeleteTaskInput!) {
			deleteTask(input: $input) {
				deletedTaskId
			}
		}
	`

	err := client.Execute(query, map[string]any{
		"input": map[string]any{
			"taskId": taskID,
		},
	}, nil)
	require.NoError(t, err)
}

func taskNotificationStatuses(
	t *testing.T,
	client *testutil.Client,
	taskID string,
	notificationType string,
) []string {
	t.Helper()

	const query = `
		query($organizationId: ID!) {
			viewer {
				notifications(
					organizationId: $organizationId
					filter: { types: [TASK_ASSIGNED, TASK_MENTIONED] }
					first: 20
				) {
					edges {
						node {
							type
							subjectId
							status
						}
					}
				}
			}
		}
	`

	var result struct {
		Viewer struct {
			Notifications struct {
				Edges []struct {
					Node struct {
						Type      string `json:"type"`
						SubjectID string `json:"subjectId"`
						Status    string `json:"status"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"notifications"`
		} `json:"viewer"`
	}

	err := client.Execute(query, map[string]any{
		"organizationId": client.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)

	var statuses []string

	for _, edge := range result.Viewer.Notifications.Edges {
		if edge.Node.SubjectID == taskID && edge.Node.Type == notificationType {
			statuses = append(statuses, edge.Node.Status)
		}
	}

	return statuses
}
