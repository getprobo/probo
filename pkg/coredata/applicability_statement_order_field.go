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
	"fmt"
)

type ApplicabilityStatementOrderField string

const (
	ApplicabilityStatementOrderFieldCreatedAt            ApplicabilityStatementOrderField = "CREATED_AT"
	ApplicabilityStatementOrderFieldControlSectionTitle  ApplicabilityStatementOrderField = "CONTROL_SECTION_TITLE"
)

func (p ApplicabilityStatementOrderField) Column() string {
	switch p {
	case ApplicabilityStatementOrderFieldCreatedAt:
		return "created_at"
	case ApplicabilityStatementOrderFieldControlSectionTitle:
		return "section_title_sort_key(section_title)"
	}

	panic("unknown ApplicabilityStatementOrderField")
}

func (p ApplicabilityStatementOrderField) String() string {
	return string(p)
}

func (p ApplicabilityStatementOrderField) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *ApplicabilityStatementOrderField) UnmarshalText(text []byte) error {
	val := string(text)
	switch val {
	case string(ApplicabilityStatementOrderFieldCreatedAt),
		string(ApplicabilityStatementOrderFieldControlSectionTitle):
		*p = ApplicabilityStatementOrderField(val)
		return nil
	}
	return fmt.Errorf("invalid ApplicabilityStatementOrderField value: %q", val)
}
