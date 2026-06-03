// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

type AccessEntryFlag string

const (
	AccessEntryFlagNone                    AccessEntryFlag = "NONE"
	AccessEntryFlagOrphaned                AccessEntryFlag = "ORPHANED"
	AccessEntryFlagInactive                AccessEntryFlag = "INACTIVE"
	AccessEntryFlagExcessive               AccessEntryFlag = "EXCESSIVE"
	AccessEntryFlagRoleMismatch            AccessEntryFlag = "ROLE_MISMATCH"
	AccessEntryFlagNew                     AccessEntryFlag = "NEW"
	AccessEntryFlagDormant                 AccessEntryFlag = "DORMANT"
	AccessEntryFlagTerminatedUser          AccessEntryFlag = "TERMINATED_USER"
	AccessEntryFlagContractorExpired       AccessEntryFlag = "CONTRACTOR_EXPIRED"
	AccessEntryFlagSoDConflict             AccessEntryFlag = "SOD_CONFLICT"
	AccessEntryFlagPrivilegedAccess        AccessEntryFlag = "PRIVILEGED_ACCESS"
	AccessEntryFlagRoleCreep               AccessEntryFlag = "ROLE_CREEP"
	AccessEntryFlagNoBusinessJustification AccessEntryFlag = "NO_BUSINESS_JUSTIFICATION"
	AccessEntryFlagOutOfDepartment         AccessEntryFlag = "OUT_OF_DEPARTMENT"
	AccessEntryFlagSharedAccount           AccessEntryFlag = "SHARED_ACCOUNT"
)

var (
	_ fmt.Stringer             = AccessEntryFlag("")
	_ encoding.TextMarshaler   = AccessEntryFlag("")
	_ encoding.TextUnmarshaler = (*AccessEntryFlag)(nil)
)

func AccessEntryFlags() []AccessEntryFlag {
	return []AccessEntryFlag{
		AccessEntryFlagNone,
		AccessEntryFlagOrphaned,
		AccessEntryFlagInactive,
		AccessEntryFlagExcessive,
		AccessEntryFlagRoleMismatch,
		AccessEntryFlagNew,
		AccessEntryFlagDormant,
		AccessEntryFlagTerminatedUser,
		AccessEntryFlagContractorExpired,
		AccessEntryFlagSoDConflict,
		AccessEntryFlagPrivilegedAccess,
		AccessEntryFlagRoleCreep,
		AccessEntryFlagNoBusinessJustification,
		AccessEntryFlagOutOfDepartment,
		AccessEntryFlagSharedAccount,
	}
}

func (v AccessEntryFlag) IsValid() bool {
	switch v {
	case
		AccessEntryFlagNone,
		AccessEntryFlagOrphaned,
		AccessEntryFlagInactive,
		AccessEntryFlagExcessive,
		AccessEntryFlagRoleMismatch,
		AccessEntryFlagNew,
		AccessEntryFlagDormant,
		AccessEntryFlagTerminatedUser,
		AccessEntryFlagContractorExpired,
		AccessEntryFlagSoDConflict,
		AccessEntryFlagPrivilegedAccess,
		AccessEntryFlagRoleCreep,
		AccessEntryFlagNoBusinessJustification,
		AccessEntryFlagOutOfDepartment,
		AccessEntryFlagSharedAccount:
		return true
	}

	return false
}

func (v AccessEntryFlag) String() string {
	return string(v)
}

func (v AccessEntryFlag) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AccessEntryFlag) UnmarshalText(text []byte) error {
	val := AccessEntryFlag(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid AccessEntryFlag value: %q", string(text))
	}

	*v = val

	return nil
}
