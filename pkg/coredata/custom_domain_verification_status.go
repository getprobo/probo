// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

type CustomDomainVerificationStatus string

const (
	CustomDomainVerificationStatusPending  CustomDomainVerificationStatus = "PENDING"
	CustomDomainVerificationStatusVerified CustomDomainVerificationStatus = "VERIFIED"
	CustomDomainVerificationStatusFailed   CustomDomainVerificationStatus = "FAILED"
)

var (
	_ fmt.Stringer             = CustomDomainVerificationStatus("")
	_ encoding.TextMarshaler   = CustomDomainVerificationStatus("")
	_ encoding.TextUnmarshaler = (*CustomDomainVerificationStatus)(nil)
)

func CustomDomainVerificationStatuses() []CustomDomainVerificationStatus {
	return []CustomDomainVerificationStatus{
		CustomDomainVerificationStatusPending,
		CustomDomainVerificationStatusVerified,
		CustomDomainVerificationStatusFailed,
	}
}

func (v CustomDomainVerificationStatus) IsValid() bool {
	switch v {
	case
		CustomDomainVerificationStatusPending,
		CustomDomainVerificationStatusVerified,
		CustomDomainVerificationStatusFailed:
		return true
	}

	return false
}

func (v CustomDomainVerificationStatus) String() string {
	return string(v)
}

func (v CustomDomainVerificationStatus) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CustomDomainVerificationStatus) UnmarshalText(text []byte) error {
	val := CustomDomainVerificationStatus(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CustomDomainVerificationStatus value: %q", string(text))
	}

	*v = val

	return nil
}
