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
	AccessReviewEntryOrderField string
)

const (
	AccessReviewEntryOrderFieldCreatedAt AccessReviewEntryOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = AccessReviewEntryOrderField("")
	_ fmt.Stringer             = AccessReviewEntryOrderField("")
	_ encoding.TextMarshaler   = AccessReviewEntryOrderField("")
	_ encoding.TextUnmarshaler = (*AccessReviewEntryOrderField)(nil)
)

func AccessReviewEntryOrderFields() []AccessReviewEntryOrderField {
	return []AccessReviewEntryOrderField{
		AccessReviewEntryOrderFieldCreatedAt,
	}
}

func (v AccessReviewEntryOrderField) IsValid() bool {
	switch v {
	case
		AccessReviewEntryOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v AccessReviewEntryOrderField) String() string {
	return string(v)
}

func (v AccessReviewEntryOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AccessReviewEntryOrderField) UnmarshalText(text []byte) error {
	val := AccessReviewEntryOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid AccessReviewEntryOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p AccessReviewEntryOrderField) Column() string {
	switch p {
	case AccessReviewEntryOrderFieldCreatedAt:
		return "created_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", p))
}
