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

type DocumentCommentThreadOrderField string

const (
	DocumentCommentThreadOrderFieldCreatedAt DocumentCommentThreadOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = DocumentCommentThreadOrderField("")
	_ fmt.Stringer             = DocumentCommentThreadOrderField("")
	_ encoding.TextMarshaler   = DocumentCommentThreadOrderField("")
	_ encoding.TextUnmarshaler = (*DocumentCommentThreadOrderField)(nil)
)

func DocumentCommentThreadOrderFields() []DocumentCommentThreadOrderField {
	return []DocumentCommentThreadOrderField{
		DocumentCommentThreadOrderFieldCreatedAt,
	}
}

func (v DocumentCommentThreadOrderField) IsValid() bool {
	return isValidOrderField(v, DocumentCommentThreadOrderFields())
}

func (v DocumentCommentThreadOrderField) String() string {
	return string(v)
}

func (v DocumentCommentThreadOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DocumentCommentThreadOrderField) UnmarshalText(text []byte) error {
	return unmarshalOrderField(v, text, DocumentCommentThreadOrderFields())
}

func (p DocumentCommentThreadOrderField) Column() string {
	switch p {
	case DocumentCommentThreadOrderFieldCreatedAt:
		return "created_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", p))
}
