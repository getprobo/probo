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
	"database/sql/driver"
	"fmt"
)

type FindingStatus string

const (
	FindingStatusOpen          FindingStatus = "OPEN"
	FindingStatusInProgress    FindingStatus = "IN_PROGRESS"
	FindingStatusClosed        FindingStatus = "CLOSED"
	FindingStatusRiskAccepted  FindingStatus = "RISK_ACCEPTED"
	FindingStatusMitigated     FindingStatus = "MITIGATED"
	FindingStatusFalsePositive FindingStatus = "FALSE_POSITIVE"
)

func FindingStatuses() []FindingStatus {
	return []FindingStatus{
		FindingStatusOpen,
		FindingStatusInProgress,
		FindingStatusClosed,
		FindingStatusRiskAccepted,
		FindingStatusMitigated,
		FindingStatusFalsePositive,
	}
}

func (fs FindingStatus) String() string {
	return string(fs)
}

func (fs *FindingStatus) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("unsupported type for FindingStatus: %T", value)
	}

	switch s {
	case "OPEN":
		*fs = FindingStatusOpen
	case "IN_PROGRESS":
		*fs = FindingStatusInProgress
	case "CLOSED":
		*fs = FindingStatusClosed
	case "RISK_ACCEPTED":
		*fs = FindingStatusRiskAccepted
	case "MITIGATED":
		*fs = FindingStatusMitigated
	case "FALSE_POSITIVE":
		*fs = FindingStatusFalsePositive
	default:
		return fmt.Errorf("invalid FindingStatus value: %q", s)
	}
	return nil
}

func (fs FindingStatus) Value() (driver.Value, error) {
	return fs.String(), nil
}
