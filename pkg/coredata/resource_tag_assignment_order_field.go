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

package coredata

import (
	"encoding"
	"fmt"

	"go.probo.inc/probo/pkg/page"
)

type (
	ResourceTagAssignmentOrderField string
)

const (
	ResourceTagAssignmentOrderFieldCreatedAt ResourceTagAssignmentOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = ResourceTagAssignmentOrderField("")
	_ fmt.Stringer             = ResourceTagAssignmentOrderField("")
	_ encoding.TextMarshaler   = ResourceTagAssignmentOrderField("")
	_ encoding.TextUnmarshaler = (*ResourceTagAssignmentOrderField)(nil)
)

func ResourceTagAssignmentOrderFields() []ResourceTagAssignmentOrderField {
	return []ResourceTagAssignmentOrderField{
		ResourceTagAssignmentOrderFieldCreatedAt,
	}
}

func (v ResourceTagAssignmentOrderField) IsValid() bool {
	switch v {
	case ResourceTagAssignmentOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v ResourceTagAssignmentOrderField) String() string {
	return string(v)
}

func (v ResourceTagAssignmentOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ResourceTagAssignmentOrderField) UnmarshalText(text []byte) error {
	val := ResourceTagAssignmentOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ResourceTagAssignmentOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p ResourceTagAssignmentOrderField) Column() string {
	switch p {
	case ResourceTagAssignmentOrderFieldCreatedAt:
		return "created_at"
	}

	return string(p)
}
