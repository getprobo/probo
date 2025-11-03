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
	DocumentType string
)

const (
	DocumentTypeOther     DocumentType = "OTHER"
	DocumentTypeISMS      DocumentType = "ISMS"
	DocumentTypePolicy    DocumentType = "POLICY"
	DocumentTypeProcedure DocumentType = "PROCEDURE"
)

func DocumentTypes() []DocumentType {
	return []DocumentType{
		DocumentTypeOther,
		DocumentTypeISMS,
		DocumentTypePolicy,
		DocumentTypeProcedure,
	}
}

func (dt DocumentType) MarshalText() ([]byte, error) {
	return []byte(dt.String()), nil
}

func (dt *DocumentType) UnmarshalText(data []byte) error {
	val := string(data)

	switch val {
	case DocumentTypeOther.String():
		*dt = DocumentTypeOther
	case DocumentTypeISMS.String():
		*dt = DocumentTypeISMS
	case DocumentTypePolicy.String():
		*dt = DocumentTypePolicy
	case DocumentTypeProcedure.String():
		*dt = DocumentTypeProcedure
	default:
		return fmt.Errorf("invalid DocumentType value: %q", val)
	}

	return nil
}

func (dt DocumentType) String() string {
	return string(dt)
}

func (dt *DocumentType) Scan(value any) error {
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid scan source for DocumentType, expected string got %T", value)
	}

	return dt.UnmarshalText([]byte(val))
}

func (dt DocumentType) Value() (driver.Value, error) {
	return dt.String(), nil
}
