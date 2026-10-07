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
	ConnectorAccountOrderField string
)

const (
	ConnectorAccountOrderFieldCreatedAt         ConnectorAccountOrderField = "CREATED_AT"
	ConnectorAccountOrderFieldExternalAccountID ConnectorAccountOrderField = "EXTERNAL_ACCOUNT_ID"
)

var (
	_ page.OrderField          = ConnectorAccountOrderField("")
	_ fmt.Stringer             = ConnectorAccountOrderField("")
	_ encoding.TextMarshaler   = ConnectorAccountOrderField("")
	_ encoding.TextUnmarshaler = (*ConnectorAccountOrderField)(nil)
)

func ConnectorAccountOrderFields() []ConnectorAccountOrderField {
	return []ConnectorAccountOrderField{
		ConnectorAccountOrderFieldCreatedAt,
		ConnectorAccountOrderFieldExternalAccountID,
	}
}

func (v ConnectorAccountOrderField) IsValid() bool {
	return isValidOrderField(v, ConnectorAccountOrderFields())
}

func (v ConnectorAccountOrderField) String() string {
	return string(v)
}

func (v ConnectorAccountOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ConnectorAccountOrderField) UnmarshalText(text []byte) error {
	return unmarshalOrderField(v, text, ConnectorAccountOrderFields())
}

func (p ConnectorAccountOrderField) Column() string {
	switch p {
	case ConnectorAccountOrderFieldCreatedAt:
		return "created_at"
	case ConnectorAccountOrderFieldExternalAccountID:
		return "external_account_id"
	}

	panic(fmt.Sprintf("unsupported order by: %s", p))
}
