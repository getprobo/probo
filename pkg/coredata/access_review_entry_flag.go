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
