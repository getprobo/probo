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
	"time"
)

type (
	InternalControlType                 string
	InternalControlNature               string
	InternalControlCadence              string
	InternalControlImplementationStatus string
)

const (
	InternalControlTypePreventive InternalControlType = "PREVENTIVE"
	InternalControlTypeDetective  InternalControlType = "DETECTIVE"
	InternalControlTypeCorrective InternalControlType = "CORRECTIVE"

	InternalControlNatureManual    InternalControlNature = "MANUAL"
	InternalControlNatureAutomated InternalControlNature = "AUTOMATED"
	InternalControlNatureHybrid    InternalControlNature = "HYBRID"

	InternalControlCadenceContinuous   InternalControlCadence = "CONTINUOUS"
	InternalControlCadenceDaily        InternalControlCadence = "DAILY"
	InternalControlCadenceWeekly       InternalControlCadence = "WEEKLY"
	InternalControlCadenceMonthly      InternalControlCadence = "MONTHLY"
	InternalControlCadenceQuarterly    InternalControlCadence = "QUARTERLY"
	InternalControlCadenceSemiannually InternalControlCadence = "SEMIANNUALLY"
	InternalControlCadenceAnnually     InternalControlCadence = "ANNUALLY"
	InternalControlCadenceAdHoc        InternalControlCadence = "AD_HOC"

	InternalControlImplementationStatusNotImplemented InternalControlImplementationStatus = "NOT_IMPLEMENTED"
	InternalControlImplementationStatusInProgress     InternalControlImplementationStatus = "IN_PROGRESS"
	InternalControlImplementationStatusImplemented    InternalControlImplementationStatus = "IMPLEMENTED"
	InternalControlImplementationStatusOperating      InternalControlImplementationStatus = "OPERATING"
)

var (
	_ fmt.Stringer             = InternalControlType("")
	_ encoding.TextMarshaler   = InternalControlType("")
	_ encoding.TextUnmarshaler = (*InternalControlType)(nil)

	_ fmt.Stringer             = InternalControlNature("")
	_ encoding.TextMarshaler   = InternalControlNature("")
	_ encoding.TextUnmarshaler = (*InternalControlNature)(nil)

	_ fmt.Stringer             = InternalControlCadence("")
	_ encoding.TextMarshaler   = InternalControlCadence("")
	_ encoding.TextUnmarshaler = (*InternalControlCadence)(nil)

	_ fmt.Stringer             = InternalControlImplementationStatus("")
	_ encoding.TextMarshaler   = InternalControlImplementationStatus("")
	_ encoding.TextUnmarshaler = (*InternalControlImplementationStatus)(nil)
)

func InternalControlTypes() []InternalControlType {
	return []InternalControlType{
		InternalControlTypePreventive,
		InternalControlTypeDetective,
		InternalControlTypeCorrective,
	}
}

func (v InternalControlType) IsValid() bool {
	switch v {
	case InternalControlTypePreventive,
		InternalControlTypeDetective,
		InternalControlTypeCorrective:
		return true
	}

	return false
}

func (v InternalControlType) String() string {
	return string(v)
}

func (v InternalControlType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *InternalControlType) UnmarshalText(text []byte) error {
	val := InternalControlType(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid InternalControlType value: %q", string(text))
	}

	*v = val

	return nil
}

func InternalControlNatures() []InternalControlNature {
	return []InternalControlNature{
		InternalControlNatureManual,
		InternalControlNatureAutomated,
		InternalControlNatureHybrid,
	}
}

func (v InternalControlNature) IsValid() bool {
	switch v {
	case InternalControlNatureManual,
		InternalControlNatureAutomated,
		InternalControlNatureHybrid:
		return true
	}

	return false
}

func (v InternalControlNature) String() string {
	return string(v)
}

func (v InternalControlNature) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *InternalControlNature) UnmarshalText(text []byte) error {
	val := InternalControlNature(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid InternalControlNature value: %q", string(text))
	}

	*v = val

	return nil
}

func InternalControlCadences() []InternalControlCadence {
	return []InternalControlCadence{
		InternalControlCadenceContinuous,
		InternalControlCadenceDaily,
		InternalControlCadenceWeekly,
		InternalControlCadenceMonthly,
		InternalControlCadenceQuarterly,
		InternalControlCadenceSemiannually,
		InternalControlCadenceAnnually,
		InternalControlCadenceAdHoc,
	}
}

func (v InternalControlCadence) IsValid() bool {
	switch v {
	case InternalControlCadenceContinuous,
		InternalControlCadenceDaily,
		InternalControlCadenceWeekly,
		InternalControlCadenceMonthly,
		InternalControlCadenceQuarterly,
		InternalControlCadenceSemiannually,
		InternalControlCadenceAnnually,
		InternalControlCadenceAdHoc:
		return true
	}

	return false
}

func (v InternalControlCadence) String() string {
	return string(v)
}

func (v InternalControlCadence) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *InternalControlCadence) UnmarshalText(text []byte) error {
	val := InternalControlCadence(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid InternalControlCadence value: %q", string(text))
	}

	*v = val

	return nil
}

// NextInternalControlDue returns the next evidence or test due instant for a
// cadence measured from from. Continuous and ad-hoc cadences have no schedule.
func NextInternalControlDue(from time.Time, cadence InternalControlCadence) *time.Time {
	var next time.Time

	switch cadence {
	case InternalControlCadenceDaily:
		next = from.AddDate(0, 0, 1)
	case InternalControlCadenceWeekly:
		next = from.AddDate(0, 0, 7)
	case InternalControlCadenceMonthly:
		next = from.AddDate(0, 1, 0)
	case InternalControlCadenceQuarterly:
		next = from.AddDate(0, 3, 0)
	case InternalControlCadenceSemiannually:
		next = from.AddDate(0, 6, 0)
	case InternalControlCadenceAnnually:
		next = from.AddDate(1, 0, 0)
	default:
		return nil
	}

	return &next
}

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
	case InternalControlImplementationStatusNotImplemented,
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

// MeasureStateForImplementationStatus maps the operational status onto the
// legacy measure state used by treatment-plan progress. Operating counts as
// implemented because the control is in place and running.
func MeasureStateForImplementationStatus(
	status InternalControlImplementationStatus,
) (MeasureState, bool) {
	switch status {
	case InternalControlImplementationStatusNotImplemented:
		return MeasureStateNotImplemented, true
	case InternalControlImplementationStatusInProgress:
		return MeasureStateInProgress, true
	case InternalControlImplementationStatusImplemented,
		InternalControlImplementationStatusOperating:
		return MeasureStateImplemented, true
	default:
		return "", false
	}
}

// ImplementationStatusForMeasureState maps a legacy measure state onto the
// operational status. States with no operational equivalent stay not implemented.
func ImplementationStatusForMeasureState(state MeasureState) InternalControlImplementationStatus {
	switch state {
	case MeasureStateInProgress:
		return InternalControlImplementationStatusInProgress
	case MeasureStateImplemented:
		return InternalControlImplementationStatusImplemented
	default:
		return InternalControlImplementationStatusNotImplemented
	}
}
