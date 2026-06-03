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

	"go.probo.inc/probo/pkg/page"
)

type ObligationOrderField string

const (
	ObligationOrderFieldCreatedAt      ObligationOrderField = "CREATED_AT"
	ObligationOrderFieldLastReviewDate ObligationOrderField = "LAST_REVIEW_DATE"
	ObligationOrderFieldDueDate        ObligationOrderField = "DUE_DATE"
	ObligationOrderFieldStatus         ObligationOrderField = "STATUS"
)

var (
	_ page.OrderField          = ObligationOrderField("")
	_ fmt.Stringer             = ObligationOrderField("")
	_ encoding.TextMarshaler   = ObligationOrderField("")
	_ encoding.TextUnmarshaler = (*ObligationOrderField)(nil)
)

func ObligationOrderFields() []ObligationOrderField {
	return []ObligationOrderField{
		ObligationOrderFieldCreatedAt,
		ObligationOrderFieldLastReviewDate,
		ObligationOrderFieldDueDate,
		ObligationOrderFieldStatus,
	}
}

func (v ObligationOrderField) IsValid() bool {
	switch v {
	case
		ObligationOrderFieldCreatedAt,
		ObligationOrderFieldLastReviewDate,
		ObligationOrderFieldDueDate,
		ObligationOrderFieldStatus:
		return true
	}

	return false
}

func (v ObligationOrderField) String() string {
	return string(v)
}

func (v ObligationOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ObligationOrderField) UnmarshalText(text []byte) error {
	val := ObligationOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ObligationOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p ObligationOrderField) Column() string {
	return string(p)
}
