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
	"database/sql/driver"
	"fmt"
)

type TrustCenterDocumentAccessStatus string

const (
	TrustCenterDocumentAccessStatusRequested TrustCenterDocumentAccessStatus = "REQUESTED"
	TrustCenterDocumentAccessStatusGranted   TrustCenterDocumentAccessStatus = "GRANTED"
	TrustCenterDocumentAccessStatusRejected  TrustCenterDocumentAccessStatus = "REJECTED"
	TrustCenterDocumentAccessStatusRevoked   TrustCenterDocumentAccessStatus = "REVOKED"
)

func (tcdas TrustCenterDocumentAccessStatus) String() string {
	return string(tcdas)
}

func (tcdas *TrustCenterDocumentAccessStatus) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("unsupported type for TrustCenterDocumentAccessStatus: %T", value)
	}

	switch s {
	case "REQUESTED":
		*tcdas = TrustCenterDocumentAccessStatusRequested
	case "GRANTED":
		*tcdas = TrustCenterDocumentAccessStatusGranted
	case "REJECTED":
		*tcdas = TrustCenterDocumentAccessStatusRejected
	case "REVOKED":
		*tcdas = TrustCenterDocumentAccessStatusRevoked
	default:
		return fmt.Errorf("invalid TrustCenterDocumentAccessStatus value: %q", s)
	}
	return nil
}

func (tcdas TrustCenterDocumentAccessStatus) Value() (driver.Value, error) {
	return tcdas.String(), nil
}
