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

// DevicePostureValueKind classifies the observed value of a posture check so
// clients can localize it. It says nothing about whether the observation is
// acceptable — rulesets own that verdict.
type DevicePostureValueKind string

const (
	// DevicePostureValueKindOn and DevicePostureValueKindOff are boolean
	// observations: the feature is turned on, or it is turned off.
	DevicePostureValueKindOn  DevicePostureValueKind = "ON"
	DevicePostureValueKindOff DevicePostureValueKind = "OFF"

	// DevicePostureValueKindImmediate is a screen lock with no grace period.
	DevicePostureValueKindImmediate DevicePostureValueKind = "IMMEDIATE"

	// DevicePostureValueKindSeconds carries a delay in Number.
	DevicePostureValueKindSeconds DevicePostureValueKind = "SECONDS"

	// DevicePostureValueKindMinPasswordLength carries a character count in
	// Number.
	DevicePostureValueKindMinPasswordLength DevicePostureValueKind = "MIN_PASSWORD_LENGTH"

	// DevicePostureValueKindConfigured means a policy exists but its content
	// could not be reduced to a single figure.
	DevicePostureValueKindConfigured DevicePostureValueKind = "CONFIGURED"

	// DevicePostureValueKindNone means the host positively reported the
	// absence of the thing being checked.
	DevicePostureValueKindNone DevicePostureValueKind = "NONE"

	// DevicePostureValueKindText carries a literal in Text that needs no
	// translation: an OS version, an engine name, a list of agents.
	DevicePostureValueKindText DevicePostureValueKind = "TEXT"

	// DevicePostureValueKindUnknown means the evidence did not answer the
	// question. It is never a guess.
	DevicePostureValueKindUnknown DevicePostureValueKind = "UNKNOWN"
)

var (
	_ fmt.Stringer             = DevicePostureValueKind("")
	_ encoding.TextMarshaler   = DevicePostureValueKind("")
	_ encoding.TextUnmarshaler = (*DevicePostureValueKind)(nil)
)

func DevicePostureValueKinds() []DevicePostureValueKind {
	return []DevicePostureValueKind{
		DevicePostureValueKindOn,
		DevicePostureValueKindOff,
		DevicePostureValueKindImmediate,
		DevicePostureValueKindSeconds,
		DevicePostureValueKindMinPasswordLength,
		DevicePostureValueKindConfigured,
		DevicePostureValueKindNone,
		DevicePostureValueKindText,
		DevicePostureValueKindUnknown,
	}
}

func (v DevicePostureValueKind) IsValid() bool {
	switch v {
	case
		DevicePostureValueKindOn,
		DevicePostureValueKindOff,
		DevicePostureValueKindImmediate,
		DevicePostureValueKindSeconds,
		DevicePostureValueKindMinPasswordLength,
		DevicePostureValueKindConfigured,
		DevicePostureValueKindNone,
		DevicePostureValueKindText,
		DevicePostureValueKindUnknown:
		return true
	}

	return false
}

func (v DevicePostureValueKind) String() string {
	return string(v)
}

func (v DevicePostureValueKind) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DevicePostureValueKind) UnmarshalText(text []byte) error {
	val := DevicePostureValueKind(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DevicePostureValueKind value: %q", string(text))
	}

	*v = val

	return nil
}
