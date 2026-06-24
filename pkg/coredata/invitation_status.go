// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"database/sql/driver"
	"encoding"
	"fmt"
	"strings"
)

type (
	InvitationStatus   string
	InvitationStatuses []InvitationStatus
)

const (
	InvitationStatusPending  InvitationStatus = "PENDING"
	InvitationStatusAccepted InvitationStatus = "ACCEPTED"
	InvitationStatusExpired  InvitationStatus = "EXPIRED"
)

var (
	_ fmt.Stringer             = InvitationStatus("")
	_ encoding.TextMarshaler   = InvitationStatus("")
	_ encoding.TextUnmarshaler = (*InvitationStatus)(nil)
)

func (v InvitationStatus) IsValid() bool {
	switch v {
	case
		InvitationStatusPending,
		InvitationStatusAccepted,
		InvitationStatusExpired:
		return true
	}

	return false
}

func (v InvitationStatus) String() string {
	return string(v)
}

func (v InvitationStatus) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *InvitationStatus) UnmarshalText(text []byte) error {
	val := InvitationStatus(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid InvitationStatus value: %q", string(text))
	}

	*v = val

	return nil
}

func (statuses InvitationStatuses) Value() (driver.Value, error) {
	if len(statuses) == 0 {
		return nil, nil
	}

	var result strings.Builder
	result.WriteString("{")

	for i, status := range statuses {
		if i > 0 {
			result.WriteString(",")
		}

		fmt.Fprintf(&result, "%q", status.String())
	}

	result.WriteString("}")

	return result.String(), nil
}
