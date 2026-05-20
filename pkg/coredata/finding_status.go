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

type FindingStatus string

const (
	FindingStatusOpen          FindingStatus = "OPEN"
	FindingStatusInProgress    FindingStatus = "IN_PROGRESS"
	FindingStatusClosed        FindingStatus = "CLOSED"
	FindingStatusRiskAccepted  FindingStatus = "RISK_ACCEPTED"
	FindingStatusMitigated     FindingStatus = "MITIGATED"
	FindingStatusFalsePositive FindingStatus = "FALSE_POSITIVE"
)

var (
	_ fmt.Stringer             = FindingStatus("")
	_ encoding.TextMarshaler   = FindingStatus("")
	_ encoding.TextUnmarshaler = (*FindingStatus)(nil)
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

func (v FindingStatus) IsValid() bool {
	switch v {
	case
		FindingStatusOpen,
		FindingStatusInProgress,
		FindingStatusClosed,
		FindingStatusRiskAccepted,
		FindingStatusMitigated,
		FindingStatusFalsePositive:
		return true
	}

	return false
}

func (v FindingStatus) String() string {
	return string(v)
}

func (v FindingStatus) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *FindingStatus) UnmarshalText(text []byte) error {
	val := FindingStatus(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid FindingStatus value: %q", string(text))
	}

	*v = val

	return nil
}
