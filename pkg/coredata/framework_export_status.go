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

type (
	FrameworkExportStatus string
)

const (
	FrameworkExportStatusPending    FrameworkExportStatus = "pending"
	FrameworkExportStatusProcessing FrameworkExportStatus = "processing"
	FrameworkExportStatusCompleted  FrameworkExportStatus = "completed"
	FrameworkExportStatusFailed     FrameworkExportStatus = "failed"
)

func (pvs FrameworkExportStatus) MarshalText() ([]byte, error) {
	return []byte(pvs.String()), nil
}

func (pvs *FrameworkExportStatus) UnmarshalText(data []byte) error {
	val := string(data)

	switch val {
	case FrameworkExportStatusPending.String():
		*pvs = FrameworkExportStatusPending
	case FrameworkExportStatusProcessing.String():
		*pvs = FrameworkExportStatusProcessing
	case FrameworkExportStatusCompleted.String():
		*pvs = FrameworkExportStatusCompleted
	case FrameworkExportStatusFailed.String():
		*pvs = FrameworkExportStatusFailed
	default:
		return fmt.Errorf("invalid FrameworkExportStatus value: %q", val)
	}

	return nil
}

func (pvs FrameworkExportStatus) String() string {
	var val string

	switch pvs {
	case FrameworkExportStatusPending:
		val = "pending"
	case FrameworkExportStatusProcessing:
		val = "processing"
	case FrameworkExportStatusCompleted:
		val = "completed"
	case FrameworkExportStatusFailed:
		val = "failed"
	default:
		panic(fmt.Errorf("invalid FrameworkExportStatus value: %q", string(pvs)))
	}

	return val
}

func (pvs *FrameworkExportStatus) Scan(value any) error {
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid scan source for FrameworkExportStatus, expected string got %T", value)
	}

	return pvs.UnmarshalText([]byte(val))
}

func (pvs FrameworkExportStatus) Value() (driver.Value, error) {
	return pvs.String(), nil
}
