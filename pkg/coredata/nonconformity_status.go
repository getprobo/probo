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

type NonconformityStatus string

const (
	NonconformityStatusOpen       NonconformityStatus = "OPEN"
	NonconformityStatusInProgress NonconformityStatus = "IN_PROGRESS"
	NonconformityStatusClosed     NonconformityStatus = "CLOSED"
)

func NonconformityStatuses() []NonconformityStatus {
	return []NonconformityStatus{
		NonconformityStatusOpen,
		NonconformityStatusInProgress,
		NonconformityStatusClosed,
	}
}

func (ncs NonconformityStatus) String() string {
	return string(ncs)
}

func (ncs *NonconformityStatus) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("unsupported type for NonconformityStatus: %T", value)
	}

	switch s {
	case "OPEN":
		*ncs = NonconformityStatusOpen
	case "IN_PROGRESS":
		*ncs = NonconformityStatusInProgress
	case "CLOSED":
		*ncs = NonconformityStatusClosed
	default:
		return fmt.Errorf("invalid NonconformityStatus value: %q", s)
	}
	return nil
}

func (ncs NonconformityStatus) Value() (driver.Value, error) {
	return ncs.String(), nil
}
