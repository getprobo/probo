// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	"database/sql/driver"
	"fmt"
	"strings"
)

type (
	InvitationStatus   string
	InvitationStatuses []InvitationStatus
)

const (
	InvitationStatusPending  InvitationStatus = "PENDING"
	InvitationStatusAccepted InvitationStatus = "ACCEPTED"
	InvitationStatusExpired  InvitationStatus = "EXPIRED"
)

func (tcv InvitationStatus) String() string {
	return string(tcv)
}

func (tcv *InvitationStatus) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("unsupported type for TrustCenterVisibility: %T", value)
	}

	switch s {
	case "PENDING":
		*tcv = InvitationStatusPending
	case "ACCEPTED":
		*tcv = InvitationStatusAccepted
	case "EXPIRED":
		*tcv = InvitationStatusExpired
	default:
		return fmt.Errorf("invalid InvitationStatus value: %q", s)
	}
	return nil
}

func (tcv InvitationStatus) Value() (driver.Value, error) {
	return tcv.String(), nil
}

func (statuses InvitationStatuses) Value() (driver.Value, error) {
	if len(statuses) == 0 {
		return nil, nil
	}

	var result strings.Builder
	result.WriteString("{")
	for i, status := range statuses {
		if i > 0 {
			result.WriteString(",")
		}
		fmt.Fprintf(&result, "%q", status.String())
	}
	result.WriteString("}")
	return result.String(), nil
}
