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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/task/sync/linear"
)

const (
	linearStateTypeBacklog   = "backlog"
	linearStateTypeTriage    = "triage"
	linearStateTypeUnstarted = "unstarted"
	linearStateTypeStarted   = "started"
	linearStateTypeCompleted = "completed"
	linearStateTypeCanceled  = "canceled"
	linearPriorityNone       = 0
)

func TaskStateToLinearType(state coredata.TaskState) string {
	switch state {
	case coredata.TaskStateBacklog:
		return linearStateTypeBacklog
	case coredata.TaskStateTodo:
		return linearStateTypeUnstarted
	case coredata.TaskStateInProgress:
		return linearStateTypeStarted
	case coredata.TaskStateDone:
		return linearStateTypeCompleted
	case coredata.TaskStateCanceled, coredata.TaskStateDuplicate:
		return linearStateTypeCanceled
	default:
		return linearStateTypeUnstarted
	}
}

func LinearTypeToTaskState(stateType string) coredata.TaskState {
	switch stateType {
	case linearStateTypeBacklog, linearStateTypeTriage:
		return coredata.TaskStateBacklog
	case linearStateTypeUnstarted:
		return coredata.TaskStateTodo
	case linearStateTypeStarted:
		return coredata.TaskStateInProgress
	case linearStateTypeCompleted:
		return coredata.TaskStateDone
	case linearStateTypeCanceled:
		return coredata.TaskStateCanceled
	default:
		return coredata.TaskStateTodo
	}
}

func TaskPriorityToLinear(priority coredata.TaskPriority) int {
	switch priority {
	case coredata.TaskPriorityUrgent:
		return 1
	case coredata.TaskPriorityHigh:
		return 2
	case coredata.TaskPriorityMedium:
		return 3
	case coredata.TaskPriorityLow:
		return 4
	default:
		return 3
	}
}

func LinearPriorityIsNone(priority int) bool {
	return priority == linearPriorityNone
}

func LinearPriorityToTask(priority int) (coredata.TaskPriority, bool) {
	switch priority {
	case linearPriorityNone:
		return "", false
	case 1:
		return coredata.TaskPriorityUrgent, true
	case 2:
		return coredata.TaskPriorityHigh, true
	case 3:
		return coredata.TaskPriorityMedium, true
	case 4:
		return coredata.TaskPriorityLow, true
	default:
		return coredata.TaskPriorityMedium, true
	}
}

func outboundLinearPriority(
	taskPriority coredata.TaskPriority,
	destination json.RawMessage,
) (*int, error) {
	dest, err := parseTaskExternalLinkDestination(destination)
	if err != nil {
		return nil, err
	}

	if dest.PreserveLinearNone && dest.LinearNoneSnapshot == taskPriority.String() {
		return nil, nil
	}

	priority := TaskPriorityToLinear(taskPriority)

	return &priority, nil
}

func applyLinearPriorityToDestination(
	destination json.RawMessage,
	linearPriority int,
	proboPriority coredata.TaskPriority,
) (json.RawMessage, error) {
	dest, err := parseTaskExternalLinkDestination(destination)
	if err != nil {
		return nil, err
	}

	if LinearPriorityIsNone(linearPriority) {
		dest.PreserveLinearNone = true
		dest.LinearNoneSnapshot = proboPriority.String()
	} else {
		dest.PreserveLinearNone = false
		dest.LinearNoneSnapshot = ""
	}

	encoded, err := json.Marshal(dest)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal destination: %w", err)
	}

	return encoded, nil
}

func clearLinearNoneFromDestination(destination json.RawMessage) (json.RawMessage, error) {
	dest, err := parseTaskExternalLinkDestination(destination)
	if err != nil {
		return nil, err
	}

	if !dest.PreserveLinearNone && dest.LinearNoneSnapshot == "" {
		return destination, nil
	}

	dest.PreserveLinearNone = false
	dest.LinearNoneSnapshot = ""

	encoded, err := json.Marshal(dest)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal destination: %w", err)
	}

	return encoded, nil
}

func parseTaskExternalLinkDestination(
	raw json.RawMessage,
) (coredata.TaskExternalLinkDestination, error) {
	var dest coredata.TaskExternalLinkDestination
	if len(raw) == 0 {
		return dest, nil
	}

	if err := json.Unmarshal(raw, &dest); err != nil {
		return dest, fmt.Errorf("cannot unmarshal destination: %w", err)
	}

	return dest, nil
}

func PickWorkflowStateID(states []linear.WorkflowState, taskState coredata.TaskState) (string, error) {
	want := TaskStateToLinearType(taskState)
	if id, ok := pickWorkflowStateIDByType(states, want); ok {
		return id, nil
	}

	if taskState == coredata.TaskStateBacklog {
		if id, ok := pickWorkflowStateIDByType(states, linearStateTypeTriage); ok {
			return id, nil
		}
	}

	return "", fmt.Errorf("cannot find Linear workflow state for type %q", want)
}

func pickWorkflowStateIDByType(states []linear.WorkflowState, want string) (string, bool) {
	var best *linear.WorkflowState

	for i := range states {
		state := &states[i]
		if state.Type != want {
			continue
		}

		if best == nil || state.Position < best.Position {
			best = state
		}
	}

	if best == nil {
		return "", false
	}

	return best.ID, true
}

func ContentToMarkdown(content string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", nil
	}

	node, err := prosemirror.Parse(content)
	if err != nil {
		return "", fmt.Errorf("cannot parse prosemirror json: %w", err)
	}

	md, err := prosemirror.RenderMarkdown(node)
	if err != nil {
		return "", fmt.Errorf("cannot render markdown: %w", err)
	}

	return md, nil
}

func MarkdownToContent(markdown string) (string, error) {
	node, err := prosemirror.ParseMarkdown(markdown)
	if err != nil {
		return "", fmt.Errorf("cannot parse markdown: %w", err)
	}

	encoded, err := json.Marshal(node)
	if err != nil {
		return "", fmt.Errorf("cannot marshal prosemirror json: %w", err)
	}

	encodedJSON := string(encoded)

	content, err := prosemirror.DefaultDocumentJSON(&encodedJSON)
	if err != nil {
		return "", fmt.Errorf("cannot sanitize task content: %w", err)
	}

	return content, nil
}

func DeadlineToLinearDate(deadline *time.Time) *string {
	if deadline == nil {
		return nil
	}

	value := deadline.UTC().Format("2006-01-02")

	return &value
}

func LinearDateToDeadline(value string) *time.Time {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return &t
	}

	if t, err := time.Parse("2006-01-02", value); err == nil {
		return &t
	}

	return nil
}

func ContentHash(
	name string,
	markdown string,
	state coredata.TaskState,
	priority coredata.TaskPriority,
	deadline *time.Time,
	assignedToID *gid.GID,
) string {
	var deadlineValue string
	if date := DeadlineToLinearDate(deadline); date != nil {
		deadlineValue = *date
	}

	sum := sha256.Sum256(
		[]byte(strings.Join(
			[]string{
				name,
				markdown,
				state.String(),
				priority.String(),
				deadlineValue,
				assignedToHash(assignedToID),
			},
			"\x1f",
		)),
	)

	return hex.EncodeToString(sum[:])
}

func assignedToHash(assignedToID *gid.GID) string {
	if assignedToID == nil {
		return ""
	}

	return assignedToID.String()
}

func taskContentHash(task *coredata.Task) (string, error) {
	markdown, err := ContentToMarkdown(task.Content)
	if err != nil {
		return "", err
	}

	return ContentHash(
		task.Name,
		markdown,
		task.State,
		task.Priority,
		task.Deadline,
		task.AssignedToID,
	), nil
}

func taskNeedsOutboundReconcile(task *coredata.Task, publishedHash string) (bool, error) {
	current, err := taskContentHash(task)
	if err != nil {
		return false, err
	}

	return current != publishedHash, nil
}
