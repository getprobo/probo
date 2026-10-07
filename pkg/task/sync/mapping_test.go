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
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/task/sync/linear"
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

	priority, ok := LinearPriorityToTask(1)
	require.True(t, ok)
	assert.Equal(t, coredata.TaskPriorityUrgent, priority)

	priority, ok = LinearPriorityToTask(2)
	require.True(t, ok)
	assert.Equal(t, coredata.TaskPriorityHigh, priority)

	priority, ok = LinearPriorityToTask(3)
	require.True(t, ok)
	assert.Equal(t, coredata.TaskPriorityMedium, priority)

	priority, ok = LinearPriorityToTask(4)
	require.True(t, ok)
	assert.Equal(t, coredata.TaskPriorityLow, priority)

	priority, ok = LinearPriorityToTask(0)
	assert.False(t, ok)
	assert.Empty(t, priority)
	assert.True(t, LinearPriorityIsNone(0))
	assert.False(t, LinearPriorityIsNone(3))

	priority, ok = LinearPriorityToTask(99)
	require.True(t, ok)
	assert.Equal(t, coredata.TaskPriorityMedium, priority)
}

func TestOutboundLinearPriority_OmitsWhenPreservingNone(t *testing.T) {
	t.Parallel()

	destination := []byte(
		`{"team_id":"team-1","preserve_linear_none":true,"linear_none_snapshot":"HIGH"}`,
	)

	priority, err := outboundLinearPriority(coredata.TaskPriorityHigh, destination)
	require.NoError(t, err)
	assert.Nil(t, priority)

	priority, err = outboundLinearPriority(coredata.TaskPriorityUrgent, destination)
	require.NoError(t, err)
	require.NotNil(t, priority)
	assert.Equal(t, 1, *priority)

	priority, err = outboundLinearPriority(coredata.TaskPriorityHigh, []byte(`{"team_id":"team-1"}`))
	require.NoError(t, err)
	require.NotNil(t, priority)
	assert.Equal(t, 2, *priority)
}

func TestApplyLinearPriorityToDestination(t *testing.T) {
	t.Parallel()

	destination, err := applyLinearPriorityToDestination(
		[]byte(`{"team_id":"team-1","linear_organization_id":"org-a"}`),
		0,
		coredata.TaskPriorityHigh,
	)
	require.NoError(t, err)

	var dest coredata.TaskExternalLinkDestination

	require.NoError(t, json.Unmarshal(destination, &dest))
	assert.Equal(t, "team-1", dest.TeamID)
	assert.Equal(t, "org-a", dest.LinearOrganizationID)
	assert.True(t, dest.PreserveLinearNone)
	assert.Equal(t, "HIGH", dest.LinearNoneSnapshot)

	destination, err = applyLinearPriorityToDestination(destination, 2, coredata.TaskPriorityHigh)
	require.NoError(t, err)

	dest = coredata.TaskExternalLinkDestination{}

	require.NoError(t, json.Unmarshal(destination, &dest))
	assert.False(t, dest.PreserveLinearNone)
	assert.Empty(t, dest.LinearNoneSnapshot)

	cleared, err := clearLinearNoneFromDestination(
		[]byte(`{"team_id":"team-1","preserve_linear_none":true,"linear_none_snapshot":"HIGH"}`),
	)
	require.NoError(t, err)

	dest = coredata.TaskExternalLinkDestination{}

	require.NoError(t, json.Unmarshal(cleared, &dest))
	assert.Equal(t, "team-1", dest.TeamID)
	assert.False(t, dest.PreserveLinearNone)
	assert.Empty(t, dest.LinearNoneSnapshot)
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

func TestPickWorkflowStateID_BacklogFallsBackToTriage(t *testing.T) {
	t.Parallel()

	triageOnly := []linear.WorkflowState{
		{ID: "t2", Type: "triage", Position: 10},
		{ID: "t1", Type: "triage", Position: 2},
		{ID: "u1", Type: "unstarted", Position: 1},
	}

	id, err := PickWorkflowStateID(triageOnly, coredata.TaskStateBacklog)
	require.NoError(t, err)
	assert.Equal(t, "t1", id)

	withBacklog := append(
		[]linear.WorkflowState{{ID: "b1", Type: "backlog", Position: 50}},
		triageOnly...,
	)

	id, err = PickWorkflowStateID(withBacklog, coredata.TaskStateBacklog)
	require.NoError(t, err)
	assert.Equal(t, "b1", id)
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

func TestMarkdownToContent_Empty(t *testing.T) {
	t.Parallel()

	for _, markdown := range []string{"", "   ", "\n"} {
		content, err := MarkdownToContent(markdown)
		require.NoError(t, err)
		assert.NotEmpty(t, content)

		roundTrip, err := ContentToMarkdown(content)
		require.NoError(t, err)
		assert.Empty(t, strings.TrimSpace(roundTrip))
	}
}

func TestMarkdownToContent_DropsUnsupported(t *testing.T) {
	t.Parallel()

	content, err := MarkdownToContent("Hello ![alt](https://example.com/a.png) world")
	require.NoError(t, err)
	assert.NotContains(t, content, "image")
	assert.Contains(t, content, "Hello")
	assert.Contains(t, content, "world")

	onlyImage, err := MarkdownToContent("![alt](https://example.com/a.png)")
	require.NoError(t, err)
	assert.NotContains(t, onlyImage, "image")
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

	first := ContentHash("title", "body", coredata.TaskStateTodo, coredata.TaskPriorityHigh, nil, nil)
	second := ContentHash("title", "body", coredata.TaskStateTodo, coredata.TaskPriorityHigh, nil, nil)
	assert.Equal(t, first, second)
	assert.NotEqual(t, first, ContentHash("other", "body", coredata.TaskStateTodo, coredata.TaskPriorityHigh, nil, nil))
}

func TestContentHash_SameLinearDateIsEqual(t *testing.T) {
	t.Parallel()

	afternoon := time.Date(2026, 9, 14, 15, 4, 5, 0, time.UTC)
	midnight := LinearDateToDeadline(*DeadlineToLinearDate(&afternoon))
	require.NotNil(t, midnight)

	afternoonHash := ContentHash("title", "body", coredata.TaskStateTodo, coredata.TaskPriorityHigh, &afternoon, nil)
	midnightHash := ContentHash("title", "body", coredata.TaskStateTodo, coredata.TaskPriorityHigh, midnight, nil)
	assert.Equal(t, afternoonHash, midnightHash)

	nextDay := time.Date(2026, 9, 15, 15, 4, 5, 0, time.UTC)
	assert.NotEqual(
		t,
		afternoonHash,
		ContentHash("title", "body", coredata.TaskStateTodo, coredata.TaskPriorityHigh, &nextDay, nil),
	)
}

func TestTaskNeedsOutboundReconcile(t *testing.T) {
	t.Parallel()

	task := &coredata.Task{
		Name:     "title",
		State:    coredata.TaskStateTodo,
		Priority: coredata.TaskPriorityHigh,
	}

	published, err := taskContentHash(task)
	require.NoError(t, err)

	needsSync, err := taskNeedsOutboundReconcile(task, published)
	require.NoError(t, err)
	assert.False(t, needsSync)

	task.Name = "renamed"
	needsSync, err = taskNeedsOutboundReconcile(task, published)
	require.NoError(t, err)
	assert.True(t, needsSync)

	task.Name = "title"
	assignee := gid.New(gid.NewTenantID(), coredata.MembershipProfileEntityType)
	task.AssignedToID = &assignee
	needsSync, err = taskNeedsOutboundReconcile(task, published)
	require.NoError(t, err)
	assert.True(t, needsSync)
}

func TestContentHash_AssigneeChanges(t *testing.T) {
	t.Parallel()

	unassigned := ContentHash("title", "body", coredata.TaskStateTodo, coredata.TaskPriorityHigh, nil, nil)
	assignee := gid.New(gid.NewTenantID(), coredata.MembershipProfileEntityType)
	assigned := ContentHash("title", "body", coredata.TaskStateTodo, coredata.TaskPriorityHigh, nil, &assignee)
	other := gid.New(gid.NewTenantID(), coredata.MembershipProfileEntityType)

	assert.NotEqual(t, unassigned, assigned)
	assert.NotEqual(
		t,
		assigned,
		ContentHash("title", "body", coredata.TaskStateTodo, coredata.TaskPriorityHigh, nil, &other),
	)
	assert.Equal(
		t,
		assigned,
		ContentHash("title", "body", coredata.TaskStateTodo, coredata.TaskPriorityHigh, nil, &assignee),
	)
}
