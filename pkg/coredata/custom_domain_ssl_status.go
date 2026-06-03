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

import (
	"encoding"
	"fmt"
)

type CustomDomainSSLStatus string

const (
	CustomDomainSSLStatusPending      CustomDomainSSLStatus = "PENDING"
	CustomDomainSSLStatusProvisioning CustomDomainSSLStatus = "PROVISIONING"
	CustomDomainSSLStatusActive       CustomDomainSSLStatus = "ACTIVE"
	CustomDomainSSLStatusRenewing     CustomDomainSSLStatus = "RENEWING"
	CustomDomainSSLStatusExpired      CustomDomainSSLStatus = "EXPIRED"
	CustomDomainSSLStatusFailed       CustomDomainSSLStatus = "FAILED"
)

var (
	_ fmt.Stringer             = CustomDomainSSLStatus("")
	_ encoding.TextMarshaler   = CustomDomainSSLStatus("")
	_ encoding.TextUnmarshaler = (*CustomDomainSSLStatus)(nil)
)

func CustomDomainSSLStatuses() []CustomDomainSSLStatus {
	return []CustomDomainSSLStatus{
		CustomDomainSSLStatusPending,
		CustomDomainSSLStatusProvisioning,
		CustomDomainSSLStatusActive,
		CustomDomainSSLStatusRenewing,
		CustomDomainSSLStatusExpired,
		CustomDomainSSLStatusFailed,
	}
}

func (v CustomDomainSSLStatus) IsValid() bool {
	switch v {
	case
		CustomDomainSSLStatusPending,
		CustomDomainSSLStatusProvisioning,
		CustomDomainSSLStatusActive,
		CustomDomainSSLStatusRenewing,
		CustomDomainSSLStatusExpired,
		CustomDomainSSLStatusFailed:
		return true
	}

	return false
}

func (v CustomDomainSSLStatus) String() string {
	return string(v)
}

func (v CustomDomainSSLStatus) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CustomDomainSSLStatus) UnmarshalText(text []byte) error {
	val := CustomDomainSSLStatus(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CustomDomainSSLStatus value: %q", string(text))
	}

	*v = val

	return nil
}
