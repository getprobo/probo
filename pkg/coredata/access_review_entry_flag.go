// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

type AccessReviewEntryFlag string

const (
	AccessReviewEntryFlagNone                    AccessReviewEntryFlag = "NONE"
	AccessReviewEntryFlagOrphaned                AccessReviewEntryFlag = "ORPHANED"
	AccessReviewEntryFlagInactive                AccessReviewEntryFlag = "INACTIVE"
	AccessReviewEntryFlagExcessive               AccessReviewEntryFlag = "EXCESSIVE"
	AccessReviewEntryFlagRoleMismatch            AccessReviewEntryFlag = "ROLE_MISMATCH"
	AccessReviewEntryFlagNew                     AccessReviewEntryFlag = "NEW"
	AccessReviewEntryFlagDormant                 AccessReviewEntryFlag = "DORMANT"
	AccessReviewEntryFlagTerminatedUser          AccessReviewEntryFlag = "TERMINATED_USER"
	AccessReviewEntryFlagContractorExpired       AccessReviewEntryFlag = "CONTRACTOR_EXPIRED"
	AccessReviewEntryFlagSoDConflict             AccessReviewEntryFlag = "SOD_CONFLICT"
	AccessReviewEntryFlagPrivilegedAccess        AccessReviewEntryFlag = "PRIVILEGED_ACCESS"
	AccessReviewEntryFlagRoleCreep               AccessReviewEntryFlag = "ROLE_CREEP"
	AccessReviewEntryFlagNoBusinessJustification AccessReviewEntryFlag = "NO_BUSINESS_JUSTIFICATION"
	AccessReviewEntryFlagOutOfDepartment         AccessReviewEntryFlag = "OUT_OF_DEPARTMENT"
	AccessReviewEntryFlagSharedAccount           AccessReviewEntryFlag = "SHARED_ACCOUNT"
)

var (
	_ fmt.Stringer             = AccessReviewEntryFlag("")
	_ encoding.TextMarshaler   = AccessReviewEntryFlag("")
	_ encoding.TextUnmarshaler = (*AccessReviewEntryFlag)(nil)
)

func AccessReviewEntryFlags() []AccessReviewEntryFlag {
	return []AccessReviewEntryFlag{
		AccessReviewEntryFlagNone,
		AccessReviewEntryFlagOrphaned,
		AccessReviewEntryFlagInactive,
		AccessReviewEntryFlagExcessive,
		AccessReviewEntryFlagRoleMismatch,
		AccessReviewEntryFlagNew,
		AccessReviewEntryFlagDormant,
		AccessReviewEntryFlagTerminatedUser,
		AccessReviewEntryFlagContractorExpired,
		AccessReviewEntryFlagSoDConflict,
		AccessReviewEntryFlagPrivilegedAccess,
		AccessReviewEntryFlagRoleCreep,
		AccessReviewEntryFlagNoBusinessJustification,
		AccessReviewEntryFlagOutOfDepartment,
		AccessReviewEntryFlagSharedAccount,
	}
}

func (v AccessReviewEntryFlag) IsValid() bool {
	switch v {
	case
		AccessReviewEntryFlagNone,
		AccessReviewEntryFlagOrphaned,
		AccessReviewEntryFlagInactive,
		AccessReviewEntryFlagExcessive,
		AccessReviewEntryFlagRoleMismatch,
		AccessReviewEntryFlagNew,
		AccessReviewEntryFlagDormant,
		AccessReviewEntryFlagTerminatedUser,
		AccessReviewEntryFlagContractorExpired,
		AccessReviewEntryFlagSoDConflict,
		AccessReviewEntryFlagPrivilegedAccess,
		AccessReviewEntryFlagRoleCreep,
		AccessReviewEntryFlagNoBusinessJustification,
		AccessReviewEntryFlagOutOfDepartment,
		AccessReviewEntryFlagSharedAccount:
		return true
	}

	return false
}

func (v AccessReviewEntryFlag) String() string {
	return string(v)
}

func (v AccessReviewEntryFlag) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AccessReviewEntryFlag) UnmarshalText(text []byte) error {
	val := AccessReviewEntryFlag(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid AccessReviewEntryFlag value: %q", string(text))
	}

	*v = val

	return nil
}
