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

type AuditOrderField string

const (
	AuditOrderFieldCreatedAt  AuditOrderField = "CREATED_AT"
	AuditOrderFieldValidFrom  AuditOrderField = "VALID_FROM"
	AuditOrderFieldValidUntil AuditOrderField = "VALID_UNTIL"
	AuditOrderFieldState      AuditOrderField = "STATE"
)

var (
	_ page.OrderField          = AuditOrderField("")
	_ fmt.Stringer             = AuditOrderField("")
	_ encoding.TextMarshaler   = AuditOrderField("")
	_ encoding.TextUnmarshaler = (*AuditOrderField)(nil)
)

func AuditOrderFields() []AuditOrderField {
	return []AuditOrderField{
		AuditOrderFieldCreatedAt,
		AuditOrderFieldValidFrom,
		AuditOrderFieldValidUntil,
		AuditOrderFieldState,
	}
}

func (v AuditOrderField) IsValid() bool {
	switch v {
	case
		AuditOrderFieldCreatedAt,
		AuditOrderFieldValidFrom,
		AuditOrderFieldValidUntil,
		AuditOrderFieldState:
		return true
	}

	return false
}

func (v AuditOrderField) String() string {
	return string(v)
}

func (v AuditOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AuditOrderField) UnmarshalText(text []byte) error {
	val := AuditOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid AuditOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p AuditOrderField) Column() string {
	return string(p)
}
