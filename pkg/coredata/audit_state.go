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

type AuditState string

const (
	AuditStateNotStarted AuditState = "NOT_STARTED"
	AuditStateInProgress AuditState = "IN_PROGRESS"
	AuditStateCompleted  AuditState = "COMPLETED"
	AuditStateRejected   AuditState = "REJECTED"
	AuditStateOutdated   AuditState = "OUTDATED"
)

func AuditStates() []AuditState {
	return []AuditState{
		AuditStateNotStarted,
		AuditStateInProgress,
		AuditStateCompleted,
		AuditStateRejected,
		AuditStateOutdated,
	}
}

func (as AuditState) String() string {
	return string(as)
}

func (as *AuditState) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("unsupported type for AuditState: %T", value)
	}

	switch s {
	case "NOT_STARTED":
		*as = AuditStateNotStarted
	case "IN_PROGRESS":
		*as = AuditStateInProgress
	case "COMPLETED":
		*as = AuditStateCompleted
	case "REJECTED":
		*as = AuditStateRejected
	case "OUTDATED":
		*as = AuditStateOutdated
	default:
		return fmt.Errorf("invalid AuditState value: %q", s)
	}
	return nil
}

func (as AuditState) Value() (driver.Value, error) {
	return as.String(), nil
}
