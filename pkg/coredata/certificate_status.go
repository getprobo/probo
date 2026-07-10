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

type CertificateStatus string

const (
	CertificateStatusPending      CertificateStatus = "PENDING"
	CertificateStatusProvisioning CertificateStatus = "PROVISIONING"
	CertificateStatusActive       CertificateStatus = "ACTIVE"
	CertificateStatusRenewing     CertificateStatus = "RENEWING"
	CertificateStatusExpired      CertificateStatus = "EXPIRED"
	CertificateStatusFailed       CertificateStatus = "FAILED"
)

var (
	_ fmt.Stringer             = CertificateStatus("")
	_ encoding.TextMarshaler   = CertificateStatus("")
	_ encoding.TextUnmarshaler = (*CertificateStatus)(nil)
)

func CertificateStatuses() []CertificateStatus {
	return []CertificateStatus{
		CertificateStatusPending,
		CertificateStatusProvisioning,
		CertificateStatusActive,
		CertificateStatusRenewing,
		CertificateStatusExpired,
		CertificateStatusFailed,
	}
}

func (v CertificateStatus) IsValid() bool {
	switch v {
	case
		CertificateStatusPending,
		CertificateStatusProvisioning,
		CertificateStatusActive,
		CertificateStatusRenewing,
		CertificateStatusExpired,
		CertificateStatusFailed:
		return true
	}

	return false
}

func (v CertificateStatus) String() string {
	return string(v)
}

func (v CertificateStatus) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *CertificateStatus) UnmarshalText(text []byte) error {
	val := CertificateStatus(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CertificateStatus value: %q", string(text))
	}

	*v = val

	return nil
}
