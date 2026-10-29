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

type DocumentSuggestionOrderField string

const (
	DocumentSuggestionOrderFieldCreatedAt DocumentSuggestionOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = DocumentSuggestionOrderField("")
	_ fmt.Stringer             = DocumentSuggestionOrderField("")
	_ encoding.TextMarshaler   = DocumentSuggestionOrderField("")
	_ encoding.TextUnmarshaler = (*DocumentSuggestionOrderField)(nil)
)

func DocumentSuggestionOrderFields() []DocumentSuggestionOrderField {
	return []DocumentSuggestionOrderField{
		DocumentSuggestionOrderFieldCreatedAt,
	}
}

func (v DocumentSuggestionOrderField) IsValid() bool {
	return isValidOrderField(v, DocumentSuggestionOrderFields())
}

func (v DocumentSuggestionOrderField) String() string {
	return string(v)
}

func (v DocumentSuggestionOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DocumentSuggestionOrderField) UnmarshalText(text []byte) error {
	return unmarshalOrderField(v, text, DocumentSuggestionOrderFields())
}

func (p DocumentSuggestionOrderField) Column() string {
	switch p {
	case DocumentSuggestionOrderFieldCreatedAt:
		return "created_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", p))
}
