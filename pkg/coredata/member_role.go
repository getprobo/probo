// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
