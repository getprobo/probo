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
	DocumentStatus uint8
)

const (
	DocumentStatusDraft DocumentStatus = iota
	DocumentStatusPublished
)

func (ps DocumentStatus) MarshalText() ([]byte, error) {
	return []byte(ps.String()), nil
}

func (ps *DocumentStatus) UnmarshalText(data []byte) error {
	val := string(data)

	switch val {
	case DocumentStatusDraft.String():
		*ps = DocumentStatusDraft
	case DocumentStatusPublished.String():
		*ps = DocumentStatusPublished
	default:
		return fmt.Errorf("invalid DocumentStatus value: %q", val)
	}

	return nil
}

func (ps DocumentStatus) String() string {
	var val string

	switch ps {
	case DocumentStatusDraft:
		val = "DRAFT"
	case DocumentStatusPublished:
		val = "PUBLISHED"
	}

	return val
}

func (ps *DocumentStatus) Scan(value any) error {
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid scan source for DocumentStatus, expected string got %T", value)
	}

	return ps.UnmarshalText([]byte(val))
}

func (ps DocumentStatus) Value() (driver.Value, error) {
	return ps.String(), nil
}
