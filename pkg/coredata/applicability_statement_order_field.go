// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

	"go.probo.inc/probo/pkg/page"
)

type ApplicabilityStatementOrderField string

const (
	ApplicabilityStatementOrderFieldCreatedAt           ApplicabilityStatementOrderField = "CREATED_AT"
	ApplicabilityStatementOrderFieldControlSectionTitle ApplicabilityStatementOrderField = "CONTROL_SECTION_TITLE"
)

var (
	_ page.OrderField          = ApplicabilityStatementOrderField("")
	_ fmt.Stringer             = ApplicabilityStatementOrderField("")
	_ encoding.TextMarshaler   = ApplicabilityStatementOrderField("")
	_ encoding.TextUnmarshaler = (*ApplicabilityStatementOrderField)(nil)
)

func ApplicabilityStatementOrderFields() []ApplicabilityStatementOrderField {
	return []ApplicabilityStatementOrderField{
		ApplicabilityStatementOrderFieldCreatedAt,
		ApplicabilityStatementOrderFieldControlSectionTitle,
	}
}

func (v ApplicabilityStatementOrderField) IsValid() bool {
	switch v {
	case
		ApplicabilityStatementOrderFieldCreatedAt,
		ApplicabilityStatementOrderFieldControlSectionTitle:
		return true
	}

	return false
}

func (v ApplicabilityStatementOrderField) String() string {
	return string(v)
}

func (v ApplicabilityStatementOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ApplicabilityStatementOrderField) UnmarshalText(text []byte) error {
	val := ApplicabilityStatementOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ApplicabilityStatementOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p ApplicabilityStatementOrderField) Column() string {
	switch p {
	case ApplicabilityStatementOrderFieldCreatedAt:
		return "created_at"
	case ApplicabilityStatementOrderFieldControlSectionTitle:
		return "section_title_sort_key(section_title)"
	}

	panic("unknown ApplicabilityStatementOrderField")
}
