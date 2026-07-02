// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

	"go.probo.inc/probo/pkg/page"
)

type (
	TaskOrderField string
)

const (
	TaskOrderFieldPriorityRank TaskOrderField = "PRIORITY_RANK" // ordering only
	TaskOrderFieldCreatedAt    TaskOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = TaskOrderField("")
	_ fmt.Stringer             = TaskOrderField("")
	_ encoding.TextMarshaler   = TaskOrderField("")
	_ encoding.TextUnmarshaler = (*TaskOrderField)(nil)
)

func TaskOrderFields() []TaskOrderField {
	return []TaskOrderField{
		TaskOrderFieldPriorityRank,
		TaskOrderFieldCreatedAt,
	}
}

func (v TaskOrderField) IsValid() bool {
	switch v {
	case
		TaskOrderFieldPriorityRank,
		TaskOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v TaskOrderField) String() string {
	return string(v)
}

func (v TaskOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *TaskOrderField) UnmarshalText(text []byte) error {
	val := TaskOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid TaskOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p TaskOrderField) Column() string {
	switch p {
	case TaskOrderFieldPriorityRank:
		return "priority_rank"
	case TaskOrderFieldCreatedAt:
		return "created_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", p))
}
