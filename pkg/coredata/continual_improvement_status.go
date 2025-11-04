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
)

type ContinualImprovementStatus string

const (
	ContinualImprovementStatusOpen       ContinualImprovementStatus = "OPEN"
	ContinualImprovementStatusInProgress ContinualImprovementStatus = "IN_PROGRESS"
	ContinualImprovementStatusClosed     ContinualImprovementStatus = "CLOSED"
)

func ContinualImprovementStatuses() []ContinualImprovementStatus {
	return []ContinualImprovementStatus{
		ContinualImprovementStatusOpen,
		ContinualImprovementStatusInProgress,
		ContinualImprovementStatusClosed,
	}
}

func (cis ContinualImprovementStatus) String() string {
	return string(cis)
}

func (cis *ContinualImprovementStatus) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("unsupported type for ContinualImprovementStatus: %T", value)
	}

	switch s {
	case "OPEN":
		*cis = ContinualImprovementStatusOpen
	case "IN_PROGRESS":
		*cis = ContinualImprovementStatusInProgress
	case "CLOSED":
		*cis = ContinualImprovementStatusClosed
	default:
		return fmt.Errorf("invalid ContinualImprovementStatus value: %q", s)
	}
	return nil
}

func (cis ContinualImprovementStatus) Value() (driver.Value, error) {
	return cis.String(), nil
}
