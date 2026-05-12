// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package coredata

import (
	"database/sql/driver"
	"fmt"
)

type (
	TaskState string
)

const (
	TaskStateTodo       TaskState = "TODO"
	TaskStateInProgress TaskState = "IN_PROGRESS"
	TaskStateDone       TaskState = "DONE"
)

func TaskStates() []TaskState {
	return []TaskState{
		TaskStateTodo,
		TaskStateInProgress,
		TaskStateDone,
	}
}

func (ts TaskState) MarshalText() ([]byte, error) {
	return []byte(ts.String()), nil
}

func (ts *TaskState) UnmarshalText(data []byte) error {
	val := TaskState(data)

	switch val {
	case TaskStateTodo, TaskStateInProgress, TaskStateDone:
		*ts = val
	default:
		return fmt.Errorf("invalid TaskState value: %q", val)
	}

	return nil
}

func (ts TaskState) String() string {
	return string(ts)
}

func (ts *TaskState) Scan(value any) error {
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid scan source for TaskState, expected string got %T", value)
	}

	return ts.UnmarshalText([]byte(val))
}

func (ts TaskState) Value() (driver.Value, error) {
	return ts.String(), nil
}
