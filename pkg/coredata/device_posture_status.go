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

import "fmt"

type DevicePostureStatus string

const (
	DevicePostureStatusPass          DevicePostureStatus = "PASS"
	DevicePostureStatusFail          DevicePostureStatus = "FAIL"
	DevicePostureStatusUnknown       DevicePostureStatus = "UNKNOWN"
	DevicePostureStatusNotApplicable DevicePostureStatus = "NOT_APPLICABLE"
)

func (s DevicePostureStatus) String() string {
	return string(s)
}

func (s DevicePostureStatus) IsValid() bool {
	switch s {
	case DevicePostureStatusPass,
		DevicePostureStatusFail,
		DevicePostureStatusUnknown,
		DevicePostureStatusNotApplicable:
		return true
	}
	return false
}

func (s DevicePostureStatus) MarshalText() ([]byte, error) {
	if !s.IsValid() {
		return nil, fmt.Errorf("invalid device posture status: %q", string(s))
	}
	return []byte(s), nil
}

func (s *DevicePostureStatus) UnmarshalText(text []byte) error {
	v := DevicePostureStatus(text)
	if !v.IsValid() {
		return fmt.Errorf("invalid device posture status: %q", string(text))
	}
	*s = v
	return nil
}
