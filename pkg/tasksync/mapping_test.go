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

package tasksync

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/tasksync/linear"
)

func TestTaskStateToLinearType(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "backlog", TaskStateToLinearType(coredata.TaskStateBacklog))
	assert.Equal(t, "unstarted", TaskStateToLinearType(coredata.TaskStateTodo))
	assert.Equal(t, "started", TaskStateToLinearType(coredata.TaskStateInProgress))
	assert.Equal(t, "completed", TaskStateToLinearType(coredata.TaskStateDone))
	assert.Equal(t, "canceled", TaskStateToLinearType(coredata.TaskStateCanceled))
	assert.Equal(t, "canceled", TaskStateToLinearType(coredata.TaskStateDuplicate))
}

func TestLinearTypeToTaskState(t *testing.T) {
	t.Parallel()

	assert.Equal(t, coredata.TaskStateBacklog, LinearTypeToTaskState("backlog"))
	assert.Equal(t, coredata.TaskStateBacklog, LinearTypeToTaskState("triage"))
	assert.Equal(t, coredata.TaskStateTodo, LinearTypeToTaskState("unstarted"))
	assert.Equal(t, coredata.TaskStateInProgress, LinearTypeToTaskState("started"))
	assert.Equal(t, coredata.TaskStateDone, LinearTypeToTaskState("completed"))
	assert.Equal(t, coredata.TaskStateCanceled, LinearTypeToTaskState("canceled"))
}

func TestTaskPriorityMapping(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 1, TaskPriorityToLinear(coredata.TaskPriorityUrgent))
	assert.Equal(t, 2, TaskPriorityToLinear(coredata.TaskPriorityHigh))
	assert.Equal(t, 3, TaskPriorityToLinear(coredata.TaskPriorityMedium))
	assert.Equal(t, 4, TaskPriorityToLinear(coredata.TaskPriorityLow))

	assert.Equal(t, coredata.TaskPriorityUrgent, LinearPriorityToTask(1))
	assert.Equal(t, coredata.TaskPriorityHigh, LinearPriorityToTask(2))
	assert.Equal(t, coredata.TaskPriorityMedium, LinearPriorityToTask(3))
	assert.Equal(t, coredata.TaskPriorityLow, LinearPriorityToTask(4))
	assert.Equal(t, coredata.TaskPriorityMedium, LinearPriorityToTask(0))
}

func TestPickWorkflowStateID(t *testing.T) {
	t.Parallel()

	states := []linear.WorkflowState{
		{ID: "s1", Type: "unstarted", Position: 1},
		{ID: "s3", Type: "started", Position: 20},
		{ID: "s2", Type: "started", Position: 5},
	}

	id, err := PickWorkflowStateID(states, coredata.TaskStateInProgress)
	require.NoError(t, err)
	assert.Equal(t, "s2", id)

	_, err = PickWorkflowStateID(states, coredata.TaskStateDone)
	require.Error(t, err)
}

func TestDeadlineMapping(t *testing.T) {
	t.Parallel()

	deadline := time.Date(2026, 9, 14, 15, 4, 5, 0, time.UTC)
	got := DeadlineToLinearDate(&deadline)
	require.NotNil(t, got)
	assert.Equal(t, "2026-09-14", *got)

	parsed := LinearDateToDeadline("2026-09-14")
	require.NotNil(t, parsed)
	assert.Equal(t, 2026, parsed.Year())
	assert.Equal(t, time.September, parsed.Month())
	assert.Equal(t, 14, parsed.Day())
}

func TestMissingTaskSyncScopes(t *testing.T) {
	t.Parallel()

	assert.Empty(t, missingTaskSyncScopes([]string{"read", "write", "issues:create"}))
	assert.Equal(t, []string{"write", "issues:create"}, missingTaskSyncScopes([]string{"read"}))
	assert.Equal(t, []string{"read", "issues:create"}, missingTaskSyncScopes([]string{"write"}))
	assert.Equal(t, []string{"read"}, missingTaskSyncScopes([]string{"write", "issues:create"}))
}

func TestMarkdownRoundTrip(t *testing.T) {
	t.Parallel()

	content, err := MarkdownToContent("Hello **world**")
	require.NoError(t, err)
	require.NotEmpty(t, content)

	markdown, err := ContentToMarkdown(content)
	require.NoError(t, err)
	assert.Contains(t, markdown, "Hello")
	assert.Contains(t, markdown, "world")
}

func TestContentHashStable(t *testing.T) {
	t.Parallel()

	first := ContentHash("title", "body", coredata.TaskStateTodo, coredata.TaskPriorityHigh, nil)
	second := ContentHash("title", "body", coredata.TaskStateTodo, coredata.TaskPriorityHigh, nil)
	assert.Equal(t, first, second)
	assert.NotEqual(t, first, ContentHash("other", "body", coredata.TaskStateTodo, coredata.TaskPriorityHigh, nil))
}
