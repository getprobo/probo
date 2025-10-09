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

// CustomDomainSSLStatus represents the SSL status of a custom domain
type CustomDomainSSLStatus string

const (
	CustomDomainSSLStatusPending      CustomDomainSSLStatus = "PENDING"
	CustomDomainSSLStatusProvisioning CustomDomainSSLStatus = "PROVISIONING"
	CustomDomainSSLStatusActive       CustomDomainSSLStatus = "ACTIVE"
	CustomDomainSSLStatusRenewing     CustomDomainSSLStatus = "RENEWING"
	CustomDomainSSLStatusExpired      CustomDomainSSLStatus = "EXPIRED"
	CustomDomainSSLStatusFailed       CustomDomainSSLStatus = "FAILED"
)

func (s CustomDomainSSLStatus) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *CustomDomainSSLStatus) UnmarshalText(data []byte) error {
	val := string(data)

	switch val {
	case CustomDomainSSLStatusPending.String():
		*s = CustomDomainSSLStatusPending
	case CustomDomainSSLStatusProvisioning.String():
		*s = CustomDomainSSLStatusProvisioning
	case CustomDomainSSLStatusActive.String():
		*s = CustomDomainSSLStatusActive
	case CustomDomainSSLStatusRenewing.String():
		*s = CustomDomainSSLStatusRenewing
	case CustomDomainSSLStatusExpired.String():
		*s = CustomDomainSSLStatusExpired
	case CustomDomainSSLStatusFailed.String():
		*s = CustomDomainSSLStatusFailed
	default:
		return fmt.Errorf("invalid CustomDomainSSLStatus value: %q", val)
	}

	return nil
}

func (s CustomDomainSSLStatus) String() string {
	return string(s)
}

func (s *CustomDomainSSLStatus) Scan(value any) error {
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid scan source for CustomDomainSSLStatus, expected string got %T", value)
	}

	return s.UnmarshalText([]byte(val))
}

func (s CustomDomainSSLStatus) Value() (driver.Value, error) {
	return s.String(), nil
}
