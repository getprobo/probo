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

type AuditState string

const (
	AuditStateNotStarted AuditState = "NOT_STARTED"
	AuditStateInProgress AuditState = "IN_PROGRESS"
	AuditStateCompleted  AuditState = "COMPLETED"
	AuditStateRejected   AuditState = "REJECTED"
	AuditStateOutdated   AuditState = "OUTDATED"
)

var (
	_ fmt.Stringer             = AuditState("")
	_ encoding.TextMarshaler   = AuditState("")
	_ encoding.TextUnmarshaler = (*AuditState)(nil)
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

func (v AuditState) IsValid() bool {
	switch v {
	case
		AuditStateNotStarted,
		AuditStateInProgress,
		AuditStateCompleted,
		AuditStateRejected,
		AuditStateOutdated:
		return true
	}

	return false
}

func (v AuditState) String() string {
	return string(v)
}

func (v AuditState) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AuditState) UnmarshalText(text []byte) error {
	val := AuditState(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid AuditState value: %q", string(text))
	}

	*v = val

	return nil
}
