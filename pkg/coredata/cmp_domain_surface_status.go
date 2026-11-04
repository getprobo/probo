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

// CMPDomainSurfaceStatus is the user's decision about a discovered
// surface. COVERED is not stored: it is derived from a banner that uses
// the same origin.
type CMPDomainSurfaceStatus string

const (
	CMPDomainSurfaceStatusUnconfirmed CMPDomainSurfaceStatus = "UNCONFIRMED"
	CMPDomainSurfaceStatusIgnored     CMPDomainSurfaceStatus = "IGNORED"
	CMPDomainSurfaceStatusVendorRun   CMPDomainSurfaceStatus = "VENDOR_RUN"
)

var (
	_ fmt.Stringer             = CMPDomainSurfaceStatus("")
	_ encoding.TextMarshaler   = CMPDomainSurfaceStatus("")
	_ encoding.TextUnmarshaler = (*CMPDomainSurfaceStatus)(nil)
)

func CMPDomainSurfaceStatuses() []CMPDomainSurfaceStatus {
	return []CMPDomainSurfaceStatus{
		CMPDomainSurfaceStatusUnconfirmed,
		CMPDomainSurfaceStatusIgnored,
		CMPDomainSurfaceStatusVendorRun,
	}
}

func (v CMPDomainSurfaceStatus) IsValid() bool {
	switch v {
	case CMPDomainSurfaceStatusUnconfirmed,
		CMPDomainSurfaceStatusIgnored,
		CMPDomainSurfaceStatusVendorRun:
		return true
	}

	return false
}

func (v CMPDomainSurfaceStatus) String() string {
	return string(v)
}

func (v CMPDomainSurfaceStatus) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CMPDomainSurfaceStatus) UnmarshalText(text []byte) error {
	val := CMPDomainSurfaceStatus(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CMPDomainSurfaceStatus value: %q", string(text))
	}

	*v = val

	return nil
}

// CMPDomainSurfaceState is the status a caller sees. COVERED wins over
// the stored status while a banner in the domain uses the surface origin.
type CMPDomainSurfaceState string

const (
	CMPDomainSurfaceStateCovered     CMPDomainSurfaceState = "COVERED"
	CMPDomainSurfaceStateUnconfirmed CMPDomainSurfaceState = "UNCONFIRMED"
	CMPDomainSurfaceStateIgnored     CMPDomainSurfaceState = "IGNORED"
	CMPDomainSurfaceStateVendorRun   CMPDomainSurfaceState = "VENDOR_RUN"
)

var (
	_ fmt.Stringer             = CMPDomainSurfaceState("")
	_ encoding.TextMarshaler   = CMPDomainSurfaceState("")
	_ encoding.TextUnmarshaler = (*CMPDomainSurfaceState)(nil)
)

func CMPDomainSurfaceStates() []CMPDomainSurfaceState {
	return []CMPDomainSurfaceState{
		CMPDomainSurfaceStateCovered,
		CMPDomainSurfaceStateUnconfirmed,
		CMPDomainSurfaceStateIgnored,
		CMPDomainSurfaceStateVendorRun,
	}
}

func (v CMPDomainSurfaceState) IsValid() bool {
	switch v {
	case CMPDomainSurfaceStateCovered,
		CMPDomainSurfaceStateUnconfirmed,
		CMPDomainSurfaceStateIgnored,
		CMPDomainSurfaceStateVendorRun:
		return true
	}

	return false
}

func (v CMPDomainSurfaceState) String() string {
	return string(v)
}

func (v CMPDomainSurfaceState) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CMPDomainSurfaceState) UnmarshalText(text []byte) error {
	val := CMPDomainSurfaceState(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CMPDomainSurfaceState value: %q", string(text))
	}

	*v = val

	return nil
}
