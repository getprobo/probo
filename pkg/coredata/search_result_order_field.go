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
	SearchResultOrderField string
)

const (
	SearchResultOrderFieldID   SearchResultOrderField = "ID"
	SearchResultOrderFieldRank SearchResultOrderField = "RANK"
)

var (
	_ page.OrderField          = SearchResultOrderField("")
	_ fmt.Stringer             = SearchResultOrderField("")
	_ encoding.TextMarshaler   = SearchResultOrderField("")
	_ encoding.TextUnmarshaler = (*SearchResultOrderField)(nil)
)

func SearchResultOrderFields() []SearchResultOrderField {
	return []SearchResultOrderField{
		SearchResultOrderFieldID,
		SearchResultOrderFieldRank,
	}
}

func (v SearchResultOrderField) IsValid() bool {
	switch v {
	case SearchResultOrderFieldID, SearchResultOrderFieldRank:
		return true
	}

	return false
}

func (v SearchResultOrderField) String() string {
	return string(v)
}

func (v SearchResultOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *SearchResultOrderField) UnmarshalText(text []byte) error {
	val := SearchResultOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid SearchResultOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p SearchResultOrderField) Column() string {
	switch p {
	case SearchResultOrderFieldID:
		return "id"
	case SearchResultOrderFieldRank:
		return "rank"
	}

	return string(p)
}
