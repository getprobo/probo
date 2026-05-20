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
	"encoding"
	"fmt"
)

type TaskPriority string

const (
	TaskPriorityUrgent TaskPriority = "URGENT"
	TaskPriorityHigh   TaskPriority = "HIGH"
	TaskPriorityMedium TaskPriority = "MEDIUM"
	TaskPriorityLow    TaskPriority = "LOW"
)

var (
	_ fmt.Stringer             = TaskPriority("")
	_ encoding.TextMarshaler   = TaskPriority("")
	_ encoding.TextUnmarshaler = (*TaskPriority)(nil)
)

func TaskPriorities() []TaskPriority {
	return []TaskPriority{
		TaskPriorityUrgent,
		TaskPriorityHigh,
		TaskPriorityMedium,
		TaskPriorityLow,
	}
}

func (v TaskPriority) IsValid() bool {
	switch v {
	case
		TaskPriorityUrgent,
		TaskPriorityHigh,
		TaskPriorityMedium,
		TaskPriorityLow:
		return true
	}

	return false
}

func (v TaskPriority) String() string {
	return string(v)
}

func (v TaskPriority) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *TaskPriority) UnmarshalText(text []byte) error {
	val := TaskPriority(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid TaskPriority value: %q", string(text))
	}

	*v = val

	return nil
}
