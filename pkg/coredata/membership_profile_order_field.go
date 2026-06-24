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
	MembershipProfileOrderField string
)

const (
	MembershipProfileOrderFieldCreatedAt        MembershipProfileOrderField = "CREATED_AT"
	MembershipProfileOrderFieldFullName         MembershipProfileOrderField = "FULL_NAME"
	MembershipProfileOrderFieldEmailAddress     MembershipProfileOrderField = "EMAIL_ADDRESS"
	MembershipProfileOrderFieldKind             MembershipProfileOrderField = "KIND"
	MembershipProfileOrderFieldOrganizationName MembershipProfileOrderField = "ORGANIZATION_NAME"
	MembershipProfileOrderFieldState            MembershipProfileOrderField = "STATE"
)

var (
	_ page.OrderField          = MembershipProfileOrderField("")
	_ fmt.Stringer             = MembershipProfileOrderField("")
	_ encoding.TextMarshaler   = MembershipProfileOrderField("")
	_ encoding.TextUnmarshaler = (*MembershipProfileOrderField)(nil)
)

func MembershipProfileOrderFields() []MembershipProfileOrderField {
	return []MembershipProfileOrderField{
		MembershipProfileOrderFieldCreatedAt,
		MembershipProfileOrderFieldFullName,
		MembershipProfileOrderFieldEmailAddress,
		MembershipProfileOrderFieldKind,
		MembershipProfileOrderFieldOrganizationName,
		MembershipProfileOrderFieldState,
	}
}

func (v MembershipProfileOrderField) IsValid() bool {
	switch v {
	case
		MembershipProfileOrderFieldCreatedAt,
		MembershipProfileOrderFieldFullName,
		MembershipProfileOrderFieldEmailAddress,
		MembershipProfileOrderFieldKind,
		MembershipProfileOrderFieldOrganizationName,
		MembershipProfileOrderFieldState:
		return true
	}

	return false
}

func (v MembershipProfileOrderField) String() string {
	return string(v)
}

func (v MembershipProfileOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *MembershipProfileOrderField) UnmarshalText(text []byte) error {
	val := MembershipProfileOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid MembershipProfileOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p MembershipProfileOrderField) Column() string {
	return string(p)
}
