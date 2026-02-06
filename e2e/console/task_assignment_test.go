// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

func TestTask_Assign(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create measure and task
	measureID := factory.NewMeasure(owner).Create()
	taskID := factory.NewTask(owner, measureID).Create()
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	query := `
		mutation UpdateTask($input: UpdateTaskInput!) {
			updateTask(input: $input) {
				task {
					id
					assignedTo {
						id
						fullName
					}
				}
			}
		}
	`

	var result struct {
		UpdateTask struct {
			Task struct {
				ID         string `json:"id"`
				AssignedTo struct {
					ID       string `json:"id"`
					FullName string `json:"fullName"`
				} `json:"assignedTo"`
			} `json:"task"`
		} `json:"updateTask"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"taskId":       taskID,
			"assignedToId": profileID.String(),
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, taskID, result.UpdateTask.Task.ID)
	assert.Equal(t, profileID.String(), result.UpdateTask.Task.AssignedTo.ID)
}

func TestTask_Unassign(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create measure, task, people and assign
	measureID := factory.NewMeasure(owner).Create()
	taskID := factory.NewTask(owner, measureID).Create()
	profileID := testutil.NewClientInOrg(t, testutil.RoleViewer, owner).GetProfileID()

	// First assign the task
	assignQuery := `
		mutation UpdateTask($input: UpdateTaskInput!) {
			updateTask(input: $input) {
				task {
					id
				}
			}
		}
	`

	_, err := owner.Do(assignQuery, map[string]any{
		"input": map[string]any{
			"taskId":       taskID,
			"assignedToId": profileID.String(),
		},
	})
	require.NoError(t, err)

	query := `
		mutation UpdateTask($input: UpdateTaskInput!) {
			updateTask(input: $input) {
				task {
					id
					assignedTo {
						id
					}
				}
			}
		}
	`

	var result struct {
		UpdateTask struct {
			Task struct {
				ID         string `json:"id"`
				AssignedTo *struct {
					ID string `json:"id"`
				} `json:"assignedTo"`
			} `json:"task"`
		} `json:"updateTask"`
	}

	err = owner.Execute(query, map[string]any{
		"input": map[string]any{
			"taskId":       taskID,
			"assignedToId": nil,
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, taskID, result.UpdateTask.Task.ID)
	assert.Nil(t, result.UpdateTask.Task.AssignedTo)
}
