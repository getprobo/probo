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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestTaskMention_Comment(t *testing.T) {
	t.Parallel()

	t.Run("mention notifies the mentioned member once", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		member := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
		taskID := factory.NewTaskWithoutInternalControl(owner).WithName("Mention comment task").Create()

		content := mentionContent(t, member.GetProfileID().String(), "Member")
		commentID := createTaskCommentWithContent(t, owner, taskID, content)

		require.Equal(t, 1, taskMentionNotificationCount(t, member, taskID))

		updateTaskCommentContent(t, owner, commentID, mentionContent(t, member.GetProfileID().String(), "Member"))

		assert.Equal(t, 1, taskMentionNotificationCount(t, member, taskID))
	})

	t.Run("edit adding a mention notifies the new member", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		member := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
		taskID := factory.NewTaskWithoutInternalControl(owner).WithName("Mention edit task").Create()

		commentID := createTaskCommentWithContent(t, owner, taskID, factory.ProseMirrorPlainText("No mention yet"))
		require.Equal(t, 0, taskMentionNotificationCount(t, member, taskID))

		updateTaskCommentContent(t, owner, commentID, mentionContent(t, member.GetProfileID().String(), "Member"))

		assert.Equal(t, 1, taskMentionNotificationCount(t, member, taskID))
	})

	t.Run("self mention does not notify", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		taskID := factory.NewTaskWithoutInternalControl(owner).WithName("Self mention task").Create()

		createTaskCommentWithContent(t, owner, taskID, mentionContent(t, owner.GetProfileID().String(), "Owner"))

		assert.Equal(t, 0, taskMentionNotificationCount(t, owner, taskID))
	})

	t.Run("reassigning the comment owner does not change who the editor is", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		member := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
		taskID := factory.NewTaskWithoutInternalControl(owner).WithName("Reassign owner task").Create()

		commentID := createTaskCommentWithContent(t, owner, taskID, factory.ProseMirrorPlainText("No mention yet"))

		// The owner edits their own comment, reassigning it to member while
		// also mentioning themselves in the new content. The editor is
		// still the owner (not the new owner, member), so the owner's
		// self-mention must not produce a notification.
		updateTaskCommentOwnerAndContent(
			t,
			owner,
			commentID,
			member.GetProfileID().String(),
			mentionContent(t, owner.GetProfileID().String(), "Owner"),
		)

		assert.Equal(t, 0, taskMentionNotificationCount(t, owner, taskID))
	})

	t.Run("member of another organization is ignored", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		outsider := testutil.NewClient(t, testutil.RoleOwner)
		taskID := factory.NewTaskWithoutInternalControl(owner).WithName("Cross org mention task").Create()

		createTaskCommentWithContent(t, owner, taskID, mentionContent(t, outsider.GetProfileID().String(), "Outsider"))

		assert.Equal(t, 0, taskMentionNotificationCount(t, outsider, taskID))
	})
}

func TestTaskMention_EmployeeComment(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	employee := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
	colleague := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
	taskID := factory.NewTaskWithoutInternalControl(owner).WithName("Employee mention task").Create()
	assignTask(t, owner, taskID, employee.GetProfileID().String())

	ids := mentionableProfileIDs(t, employee, employee.GetOrganizationID().String(), "Test User")
	assert.Contains(t, ids, colleague.GetProfileID().String())

	createTaskCommentWithContent(
		t,
		employee,
		taskID,
		mentionContent(t, colleague.GetProfileID().String(), "Colleague"),
	)

	assert.Equal(t, 1, taskMentionNotificationCount(t, colleague, taskID))
	assert.Equal(t, 0, taskMentionNotificationCount(t, employee, taskID))
}

func TestTaskMention_Description(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	member := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
	taskID := factory.NewTaskWithoutInternalControl(owner).WithName("Mention description task").Create()

	content := mentionContent(t, member.GetProfileID().String(), "Member")
	updateTaskDescription(t, owner, taskID, content)
	updateTaskDescription(t, owner, taskID, content)

	assert.Equal(t, 1, taskMentionNotificationCount(t, member, taskID))
}

func mentionableProfileIDs(t *testing.T, client *testutil.Client, organizationID string, query string) []string {
	t.Helper()

	const gql = `
		query($organizationId: ID!, $first: Int!, $filter: ProfileFilter) {
			organization: node(id: $organizationId) {
				... on Organization {
					profiles(
						first: $first
						orderBy: { direction: ASC, field: FULL_NAME }
						filter: $filter
					) {
						edges {
							node {
								id
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Organization struct {
			Profiles struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"profiles"`
		} `json:"organization"`
	}

	err := client.Execute(
		gql,
		map[string]any{
			"organizationId": organizationID,
			"first":          8,
			"filter": map[string]any{
				"states":        []string{"ACTIVE", "PENDING"},
				"contractEnded": false,
				"query":         query,
			},
		},
		&result,
	)
	require.NoError(t, err)

	ids := make([]string, 0, len(result.Organization.Profiles.Edges))
	for _, edge := range result.Organization.Profiles.Edges {
		ids = append(ids, edge.Node.ID)
	}

	return ids
}

func mentionContent(t *testing.T, profileID string, label string) string {
	t.Helper()

	doc := map[string]any{
		"type": "doc",
		"content": []any{
			map[string]any{
				"type": "paragraph",
				"content": []any{
					map[string]any{"type": "text", "text": "Please review "},
					map[string]any{
						"type":  "mention",
						"attrs": map[string]any{"id": profileID, "label": label},
					},
				},
			},
		},
	}

	raw, err := json.Marshal(doc)
	require.NoError(t, err)

	return string(raw)
}

func createTaskCommentWithContent(t *testing.T, client *testutil.Client, taskID string, content string) string {
	t.Helper()

	const query = `
		mutation($input: CreateTaskCommentInput!) {
			createTaskComment(input: $input) {
				taskCommentEdge {
					node {
						id
					}
				}
			}
		}
	`

	var result struct {
		CreateTaskComment struct {
			TaskCommentEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"taskCommentEdge"`
		} `json:"createTaskComment"`
	}

	err := client.Execute(query, map[string]any{
		"input": map[string]any{
			"taskId":  taskID,
			"content": content,
		},
	}, &result)
	require.NoError(t, err)

	return result.CreateTaskComment.TaskCommentEdge.Node.ID
}

func updateTaskCommentContent(t *testing.T, client *testutil.Client, commentID string, content string) {
	t.Helper()

	const query = `
		mutation($input: UpdateTaskCommentInput!) {
			updateTaskComment(input: $input) {
				taskComment {
					id
				}
			}
		}
	`

	err := client.Execute(query, map[string]any{
		"input": map[string]any{
			"taskCommentId": commentID,
			"content":       content,
		},
	}, nil)
	require.NoError(t, err)
}

func updateTaskCommentOwnerAndContent(
	t *testing.T,
	client *testutil.Client,
	commentID string,
	ownerID string,
	content string,
) {
	t.Helper()

	const query = `
		mutation($input: UpdateTaskCommentInput!) {
			updateTaskComment(input: $input) {
				taskComment {
					id
				}
			}
		}
	`

	err := client.Execute(query, map[string]any{
		"input": map[string]any{
			"taskCommentId": commentID,
			"ownerId":       ownerID,
			"content":       content,
		},
	}, nil)
	require.NoError(t, err)
}

func updateTaskDescription(t *testing.T, client *testutil.Client, taskID string, content string) {
	t.Helper()

	const query = `
		mutation($input: UpdateTaskInput!) {
			updateTask(input: $input) {
				task {
					id
				}
			}
		}
	`

	err := client.Execute(query, map[string]any{
		"input": map[string]any{
			"taskId":  taskID,
			"content": content,
		},
	}, nil)
	require.NoError(t, err)
}

func taskMentionNotificationCount(t *testing.T, client *testutil.Client, taskID string) int {
	t.Helper()

	const query = `
		query($organizationId: ID!) {
			viewer {
				notifications(
					organizationId: $organizationId
					filter: { types: [TASK_MENTIONED] }
					first: 20
				) {
					edges {
						node {
							subjectId
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
						SubjectID string `json:"subjectId"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"notifications"`
		} `json:"viewer"`
	}

	err := client.Execute(query, map[string]any{
		"organizationId": client.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)

	count := 0

	for _, edge := range result.Viewer.Notifications.Edges {
		if edge.Node.SubjectID == taskID {
			count++
		}
	}

	return count
}
