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

type (
	ExportJobStatus string
)

const (
	ExportJobStatusPending    ExportJobStatus = "PENDING"
	ExportJobStatusProcessing ExportJobStatus = "PROCESSING"
	ExportJobStatusCompleted  ExportJobStatus = "COMPLETED"
	ExportJobStatusFailed     ExportJobStatus = "FAILED"
)

func (ejs ExportJobStatus) String() string {
	return string(ejs)
}

func (ejs *ExportJobStatus) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("unsupported type for ExportJobStatus: %T", value)
	}

	switch s {
	case ExportJobStatusPending.String():
		*ejs = ExportJobStatusPending
	case ExportJobStatusProcessing.String():
		*ejs = ExportJobStatusProcessing
	case ExportJobStatusCompleted.String():
		*ejs = ExportJobStatusCompleted
	case ExportJobStatusFailed.String():
		*ejs = ExportJobStatusFailed
	default:
		return fmt.Errorf("invalid ExportJobStatus value: %q", s)
	}
	return nil
}

func (ejs ExportJobStatus) Value() (driver.Value, error) {
	return ejs.String(), nil
}
