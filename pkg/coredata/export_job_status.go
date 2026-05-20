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

type (
	ExportJobStatus string
)

const (
	ExportJobStatusPending    ExportJobStatus = "PENDING"
	ExportJobStatusProcessing ExportJobStatus = "PROCESSING"
	ExportJobStatusCompleted  ExportJobStatus = "COMPLETED"
	ExportJobStatusFailed     ExportJobStatus = "FAILED"
)

var (
	_ fmt.Stringer             = ExportJobStatus("")
	_ encoding.TextMarshaler   = ExportJobStatus("")
	_ encoding.TextUnmarshaler = (*ExportJobStatus)(nil)
)

func ExportJobStatuses() []ExportJobStatus {
	return []ExportJobStatus{
		ExportJobStatusPending,
		ExportJobStatusProcessing,
		ExportJobStatusCompleted,
		ExportJobStatusFailed,
	}
}

func (v ExportJobStatus) IsValid() bool {
	switch v {
	case
		ExportJobStatusPending,
		ExportJobStatusProcessing,
		ExportJobStatusCompleted,
		ExportJobStatusFailed:
		return true
	}

	return false
}

func (v ExportJobStatus) String() string {
	return string(v)
}

func (v ExportJobStatus) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ExportJobStatus) UnmarshalText(text []byte) error {
	val := ExportJobStatus(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ExportJobStatus value: %q", string(text))
	}

	*v = val

	return nil
}
