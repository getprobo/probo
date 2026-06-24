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
	MembershipOrderField string
)

const (
	MembershipOrderFieldOrganizationName MembershipOrderField = "ORGANIZATION_NAME"
	MembershipOrderFieldFullName         MembershipOrderField = "FULL_NAME"
	MembershipOrderFieldEmailAddress     MembershipOrderField = "EMAIL_ADDRESS"
	MembershipOrderFieldRole             MembershipOrderField = "ROLE"
	MembershipOrderFieldCreatedAt        MembershipOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = MembershipOrderField("")
	_ fmt.Stringer             = MembershipOrderField("")
	_ encoding.TextMarshaler   = MembershipOrderField("")
	_ encoding.TextUnmarshaler = (*MembershipOrderField)(nil)
)

func MembershipOrderFields() []MembershipOrderField {
	return []MembershipOrderField{
		MembershipOrderFieldOrganizationName,
		MembershipOrderFieldFullName,
		MembershipOrderFieldEmailAddress,
		MembershipOrderFieldRole,
		MembershipOrderFieldCreatedAt,
	}
}

func (v MembershipOrderField) IsValid() bool {
	switch v {
	case
		MembershipOrderFieldOrganizationName,
		MembershipOrderFieldFullName,
		MembershipOrderFieldEmailAddress,
		MembershipOrderFieldRole,
		MembershipOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v MembershipOrderField) String() string {
	return string(v)
}

func (v MembershipOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *MembershipOrderField) UnmarshalText(text []byte) error {
	val := MembershipOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid MembershipOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p MembershipOrderField) Column() string {
	switch p {
	case MembershipOrderFieldOrganizationName:
		return "organization_name"
	case MembershipOrderFieldFullName:
		return "full_name"
	case MembershipOrderFieldEmailAddress:
		return "email_address"
	case MembershipOrderFieldRole:
		return "role"
	case MembershipOrderFieldCreatedAt:
		return "created_at"
	}

	return string(p)
}
