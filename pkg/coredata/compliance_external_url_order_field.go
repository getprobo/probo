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

type (
	ComplianceExternalURLOrderField string
)

const (
	ComplianceExternalURLOrderFieldCreatedAt ComplianceExternalURLOrderField = "CREATED_AT"
	ComplianceExternalURLOrderFieldRank      ComplianceExternalURLOrderField = "RANK"
)

var (
	_ page.OrderField          = ComplianceExternalURLOrderField("")
	_ fmt.Stringer             = ComplianceExternalURLOrderField("")
	_ encoding.TextMarshaler   = ComplianceExternalURLOrderField("")
	_ encoding.TextUnmarshaler = (*ComplianceExternalURLOrderField)(nil)
)

func ComplianceExternalURLOrderFields() []ComplianceExternalURLOrderField {
	return []ComplianceExternalURLOrderField{
		ComplianceExternalURLOrderFieldCreatedAt,
		ComplianceExternalURLOrderFieldRank,
	}
}

func (v ComplianceExternalURLOrderField) IsValid() bool {
	switch v {
	case
		ComplianceExternalURLOrderFieldCreatedAt,
		ComplianceExternalURLOrderFieldRank:
		return true
	}

	return false
}

func (v ComplianceExternalURLOrderField) String() string {
	return string(v)
}

func (v ComplianceExternalURLOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ComplianceExternalURLOrderField) UnmarshalText(text []byte) error {
	val := ComplianceExternalURLOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ComplianceExternalURLOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p ComplianceExternalURLOrderField) Column() string {
	switch p {
	case ComplianceExternalURLOrderFieldCreatedAt:
		return "created_at"
	case ComplianceExternalURLOrderFieldRank:
		return "rank"
	default:
		return string(p)
	}
}
