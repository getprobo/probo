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

type (
	MeasureState uint8
)

const (
	MeasureStateNotStarted MeasureState = iota
	MeasureStateInProgress
	MeasureStateNotApplicable
	MeasureStateImplemented
)

func MeasureStates() []MeasureState {
	return []MeasureState{
		MeasureStateNotStarted,
		MeasureStateInProgress,
		MeasureStateNotApplicable,
		MeasureStateImplemented,
	}
}

func (ms MeasureState) MarshalText() ([]byte, error) {
	return []byte(ms.String()), nil
}

func (ms *MeasureState) UnmarshalText(data []byte) error {
	val := string(data)

	switch val {
	case MeasureStateNotStarted.String():
		*ms = MeasureStateNotStarted
	case MeasureStateInProgress.String():
		*ms = MeasureStateInProgress
	case MeasureStateNotApplicable.String():
		*ms = MeasureStateNotApplicable
	case MeasureStateImplemented.String():
		*ms = MeasureStateImplemented
	default:
		return fmt.Errorf("invalid MeasureState value: %q", val)
	}

	return nil
}

func (ms MeasureState) String() string {
	var val string

	switch ms {
	case MeasureStateNotStarted:
		val = "NOT_STARTED"
	case MeasureStateInProgress:
		val = "IN_PROGRESS"
	case MeasureStateNotApplicable:
		val = "NOT_APPLICABLE"
	case MeasureStateImplemented:
		val = "IMPLEMENTED"
	}

	return val
}

func (ms *MeasureState) Scan(value any) error {
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid scan source for MeasureState, expected string got %T", value)
	}

	return ms.UnmarshalText([]byte(val))
}

func (ms MeasureState) Value() (driver.Value, error) {
	return ms.String(), nil
}
