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

type MailingListUpdateOrderField string

const (
	MailingListUpdateOrderFieldCreatedAt MailingListUpdateOrderField = "CREATED_AT"
	MailingListUpdateOrderFieldUpdatedAt MailingListUpdateOrderField = "UPDATED_AT"
)

var (
	_ page.OrderField          = MailingListUpdateOrderField("")
	_ fmt.Stringer             = MailingListUpdateOrderField("")
	_ encoding.TextMarshaler   = MailingListUpdateOrderField("")
	_ encoding.TextUnmarshaler = (*MailingListUpdateOrderField)(nil)
)

func MailingListUpdateOrderFields() []MailingListUpdateOrderField {
	return []MailingListUpdateOrderField{
		MailingListUpdateOrderFieldCreatedAt,
		MailingListUpdateOrderFieldUpdatedAt,
	}
}

func (v MailingListUpdateOrderField) IsValid() bool {
	switch v {
	case
		MailingListUpdateOrderFieldCreatedAt,
		MailingListUpdateOrderFieldUpdatedAt:
		return true
	}

	return false
}

func (v MailingListUpdateOrderField) String() string {
	return string(v)
}

func (v MailingListUpdateOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *MailingListUpdateOrderField) UnmarshalText(text []byte) error {
	val := MailingListUpdateOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid MailingListUpdateOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (f MailingListUpdateOrderField) Column() string {
	switch f {
	case MailingListUpdateOrderFieldCreatedAt:
		return "created_at"
	case MailingListUpdateOrderFieldUpdatedAt:
		return "updated_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", f))
}
