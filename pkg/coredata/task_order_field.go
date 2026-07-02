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
