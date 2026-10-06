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

type InternalControlImplementationStatus string

const (
	InternalControlImplementationStatusNotImplemented InternalControlImplementationStatus = "NOT_IMPLEMENTED"
	InternalControlImplementationStatusInProgress     InternalControlImplementationStatus = "IN_PROGRESS"
	InternalControlImplementationStatusImplemented    InternalControlImplementationStatus = "IMPLEMENTED"
	InternalControlImplementationStatusOperating      InternalControlImplementationStatus = "OPERATING"
)

var (
	_ fmt.Stringer             = InternalControlImplementationStatus("")
	_ encoding.TextMarshaler   = InternalControlImplementationStatus("")
	_ encoding.TextUnmarshaler = (*InternalControlImplementationStatus)(nil)
)

func InternalControlImplementationStatuses() []InternalControlImplementationStatus {
	return []InternalControlImplementationStatus{
		InternalControlImplementationStatusNotImplemented,
		InternalControlImplementationStatusInProgress,
		InternalControlImplementationStatusImplemented,
		InternalControlImplementationStatusOperating,
	}
}

func (v InternalControlImplementationStatus) IsValid() bool {
	switch v {
	case
		InternalControlImplementationStatusNotImplemented,
		InternalControlImplementationStatusInProgress,
		InternalControlImplementationStatusImplemented,
		InternalControlImplementationStatusOperating:
		return true
	}

	return false
}

func (v InternalControlImplementationStatus) String() string {
	return string(v)
}

func (v InternalControlImplementationStatus) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *InternalControlImplementationStatus) UnmarshalText(text []byte) error {
	val := InternalControlImplementationStatus(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid InternalControlImplementationStatus value: %q", string(text))
	}

	*v = val

	return nil
}

// InternalControlStateForImplementationStatus maps the operational status onto the
// legacy internal control state used by treatment-plan progress. Operating counts as
// implemented because the control is in place and running.
func InternalControlStateForImplementationStatus(
	status InternalControlImplementationStatus,
) (InternalControlState, bool) {
	switch status {
	case InternalControlImplementationStatusNotImplemented:
		return InternalControlStateNotImplemented, true
	case InternalControlImplementationStatusInProgress:
		return InternalControlStateInProgress, true
	case InternalControlImplementationStatusImplemented,
		InternalControlImplementationStatusOperating:
		return InternalControlStateImplemented, true
	default:
		return "", false
	}
}

// ImplementationStatusForInternalControlState maps a legacy internal control state onto the
// operational status. States with no operational equivalent stay not implemented.
func ImplementationStatusForInternalControlState(state InternalControlState) InternalControlImplementationStatus {
	switch state {
	case InternalControlStateInProgress:
		return InternalControlImplementationStatusInProgress
	case InternalControlStateImplemented:
		return InternalControlImplementationStatusImplemented
	default:
		return InternalControlImplementationStatusNotImplemented
	}
}
