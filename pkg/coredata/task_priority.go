// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

type TaskPriority string

const (
	TaskPriorityUrgent TaskPriority = "URGENT"
	TaskPriorityHigh   TaskPriority = "HIGH"
	TaskPriorityMedium TaskPriority = "MEDIUM"
	TaskPriorityLow    TaskPriority = "LOW"
)

func TaskPriorities() []TaskPriority {
	return []TaskPriority{
		TaskPriorityUrgent,
		TaskPriorityHigh,
		TaskPriorityMedium,
		TaskPriorityLow,
	}
}

func (tp TaskPriority) String() string {
	return string(tp)
}

func (tp *TaskPriority) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("unsupported type for TaskPriority: %T", value)
	}

	switch s {
	case "URGENT":
		*tp = TaskPriorityUrgent
	case "HIGH":
		*tp = TaskPriorityHigh
	case "MEDIUM":
		*tp = TaskPriorityMedium
	case "LOW":
		*tp = TaskPriorityLow
	default:
		return fmt.Errorf("invalid TaskPriority value: %q", s)
	}
	return nil
}

func (tp TaskPriority) Value() (driver.Value, error) {
	return tp.String(), nil
}
