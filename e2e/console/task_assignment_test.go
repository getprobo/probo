// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

func TestTask_Assign(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create measure and task
	measureID := factory.NewMeasure(owner).Create()
	taskID := factory.NewTask(owner, measureID).Create()
	profileID := factory.CreateUser(owner)

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
			"assignedToId": profileID,
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, taskID, result.UpdateTask.Task.ID)
	assert.Equal(t, profileID, result.UpdateTask.Task.AssignedTo.ID)
}

func TestTask_Unassign(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create measure, task, people and assign
	measureID := factory.NewMeasure(owner).Create()
	taskID := factory.NewTask(owner, measureID).Create()
	profileID := factory.CreateUser(owner)

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
			"assignedToId": profileID,
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
