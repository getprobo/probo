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

type CustomDomainVerificationStatus string

const (
	CustomDomainVerificationStatusPending  CustomDomainVerificationStatus = "PENDING"
	CustomDomainVerificationStatusVerified CustomDomainVerificationStatus = "VERIFIED"
	CustomDomainVerificationStatusFailed   CustomDomainVerificationStatus = "FAILED"
)

func (s CustomDomainVerificationStatus) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *CustomDomainVerificationStatus) UnmarshalText(data []byte) error {
	val := string(data)

	switch val {
	case CustomDomainVerificationStatusPending.String():
		*s = CustomDomainVerificationStatusPending
	case CustomDomainVerificationStatusVerified.String():
		*s = CustomDomainVerificationStatusVerified
	case CustomDomainVerificationStatusFailed.String():
		*s = CustomDomainVerificationStatusFailed
	default:
		return fmt.Errorf("invalid CustomDomainVerificationStatus value: %q", val)
	}

	return nil
}

func (s CustomDomainVerificationStatus) String() string {
	return string(s)
}

func (s *CustomDomainVerificationStatus) Scan(value any) error {
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid scan source for CustomDomainVerificationStatus, expected string got %T", value)
	}

	return s.UnmarshalText([]byte(val))
}

func (s CustomDomainVerificationStatus) Value() (driver.Value, error) {
	return s.String(), nil
}
