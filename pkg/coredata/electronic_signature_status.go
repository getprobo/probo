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

type (
	ElectronicSignatureStatus string
)

const (
	ElectronicSignatureStatusPending    ElectronicSignatureStatus = "PENDING"
	ElectronicSignatureStatusAccepted   ElectronicSignatureStatus = "ACCEPTED"
	ElectronicSignatureStatusProcessing ElectronicSignatureStatus = "PROCESSING"
	ElectronicSignatureStatusCompleted  ElectronicSignatureStatus = "COMPLETED"
	ElectronicSignatureStatusFailed     ElectronicSignatureStatus = "FAILED"
)

var (
	_ fmt.Stringer             = ElectronicSignatureStatus("")
	_ encoding.TextMarshaler   = ElectronicSignatureStatus("")
	_ encoding.TextUnmarshaler = (*ElectronicSignatureStatus)(nil)
)

func ElectronicSignatureStatuses() []ElectronicSignatureStatus {
	return []ElectronicSignatureStatus{
		ElectronicSignatureStatusPending,
		ElectronicSignatureStatusAccepted,
		ElectronicSignatureStatusProcessing,
		ElectronicSignatureStatusCompleted,
		ElectronicSignatureStatusFailed,
	}
}

func (v ElectronicSignatureStatus) IsValid() bool {
	switch v {
	case
		ElectronicSignatureStatusPending,
		ElectronicSignatureStatusAccepted,
		ElectronicSignatureStatusProcessing,
		ElectronicSignatureStatusCompleted,
		ElectronicSignatureStatusFailed:
		return true
	}

	return false
}

func (v ElectronicSignatureStatus) String() string {
	return string(v)
}

func (v ElectronicSignatureStatus) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ElectronicSignatureStatus) UnmarshalText(text []byte) error {
	val := ElectronicSignatureStatus(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ElectronicSignatureStatus value: %q", string(text))
	}

	*v = val

	return nil
}
