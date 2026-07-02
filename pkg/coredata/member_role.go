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
)

type MembershipRole string

const (
	MembershipRoleOwner    MembershipRole = "OWNER"
	MembershipRoleAdmin    MembershipRole = "ADMIN"
	MembershipRoleEmployee MembershipRole = "EMPLOYEE"
	MembershipRoleViewer   MembershipRole = "VIEWER"
	MembershipRoleAuditor  MembershipRole = "AUDITOR"
)

var (
	_ fmt.Stringer             = MembershipRole("")
	_ encoding.TextMarshaler   = MembershipRole("")
	_ encoding.TextUnmarshaler = (*MembershipRole)(nil)
)

func MembershipRoles() []MembershipRole {
	return []MembershipRole{
		MembershipRoleOwner,
		MembershipRoleAdmin,
		MembershipRoleEmployee,
		MembershipRoleViewer,
		MembershipRoleAuditor,
	}
}

func (v MembershipRole) IsValid() bool {
	switch v {
	case
		MembershipRoleOwner,
		MembershipRoleAdmin,
		MembershipRoleEmployee,
		MembershipRoleViewer,
		MembershipRoleAuditor:
		return true
	}

	return false
}

func (v MembershipRole) String() string {
	return string(v)
}

func (v MembershipRole) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *MembershipRole) UnmarshalText(text []byte) error {
	val := MembershipRole(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid MembershipRole value: %q", string(text))
	}

	*v = val

	return nil
}
