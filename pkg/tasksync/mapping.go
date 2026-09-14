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
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/tasksync/linear"
)

const (
	linearStateTypeBacklog   = "backlog"
	linearStateTypeUnstarted = "unstarted"
	linearStateTypeStarted   = "started"
	linearStateTypeCompleted = "completed"
	linearStateTypeCanceled  = "canceled"
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
	case linearStateTypeBacklog, "triage":
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

func LinearPriorityToTask(priority int) coredata.TaskPriority {
	switch priority {
	case 1:
		return coredata.TaskPriorityUrgent
	case 2:
		return coredata.TaskPriorityHigh
	case 4:
		return coredata.TaskPriorityLow
	default:
		return coredata.TaskPriorityMedium
	}
}

func PickWorkflowStateID(states []linear.WorkflowState, taskState coredata.TaskState) (string, error) {
	want := TaskStateToLinearType(taskState)

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
		return "", fmt.Errorf("cannot find Linear workflow state for type %q", want)
	}

	return best.ID, nil
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
) string {
	var deadlineValue string
	if deadline != nil {
		deadlineValue = deadline.UTC().Format(time.RFC3339)
	}

	sum := sha256.Sum256(
		[]byte(strings.Join(
			[]string{
				name,
				markdown,
				state.String(),
				priority.String(),
				deadlineValue,
			},
			"\x1f",
		)),
	)

	return hex.EncodeToString(sum[:])
}
