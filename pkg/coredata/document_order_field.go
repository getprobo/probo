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
	DocumentOrderField string
)

const (
	DocumentOrderFieldCreatedAt    DocumentOrderField = "CREATED_AT"
	DocumentOrderFieldUpdatedAt    DocumentOrderField = "UPDATED_AT"
	DocumentOrderFieldTitle        DocumentOrderField = "TITLE"
	DocumentOrderFieldDocumentType DocumentOrderField = "DOCUMENT_TYPE"
)

var (
	_ page.OrderField          = DocumentOrderField("")
	_ fmt.Stringer             = DocumentOrderField("")
	_ encoding.TextMarshaler   = DocumentOrderField("")
	_ encoding.TextUnmarshaler = (*DocumentOrderField)(nil)
)

func DocumentOrderFields() []DocumentOrderField {
	return []DocumentOrderField{
		DocumentOrderFieldCreatedAt,
		DocumentOrderFieldUpdatedAt,
		DocumentOrderFieldTitle,
		DocumentOrderFieldDocumentType,
	}
}

func (v DocumentOrderField) IsValid() bool {
	switch v {
	case
		DocumentOrderFieldCreatedAt,
		DocumentOrderFieldUpdatedAt,
		DocumentOrderFieldTitle,
		DocumentOrderFieldDocumentType:
		return true
	}

	return false
}

func (v DocumentOrderField) String() string {
	return string(v)
}

func (v DocumentOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DocumentOrderField) UnmarshalText(text []byte) error {
	val := DocumentOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DocumentOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p DocumentOrderField) Column() string {
	switch p {
	case DocumentOrderFieldCreatedAt:
		return "created_at"
	case DocumentOrderFieldUpdatedAt:
		return "updated_at"
	case DocumentOrderFieldTitle:
		return "title"
	case DocumentOrderFieldDocumentType:
		return "document_type"
	}

	panic(fmt.Sprintf("unsupported order by: %s", p))
}
