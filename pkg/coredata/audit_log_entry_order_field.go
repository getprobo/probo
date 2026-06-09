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

type AuditLogEntryOrderField string

const (
	AuditLogEntryOrderFieldCreatedAt AuditLogEntryOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = AuditLogEntryOrderField("")
	_ fmt.Stringer             = AuditLogEntryOrderField("")
	_ encoding.TextMarshaler   = AuditLogEntryOrderField("")
	_ encoding.TextUnmarshaler = (*AuditLogEntryOrderField)(nil)
)

func AuditLogEntryOrderFields() []AuditLogEntryOrderField {
	return []AuditLogEntryOrderField{
		AuditLogEntryOrderFieldCreatedAt,
	}
}

func (v AuditLogEntryOrderField) IsValid() bool {
	switch v {
	case
		AuditLogEntryOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v AuditLogEntryOrderField) String() string {
	return string(v)
}

func (v AuditLogEntryOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AuditLogEntryOrderField) UnmarshalText(text []byte) error {
	val := AuditLogEntryOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid AuditLogEntryOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p AuditLogEntryOrderField) Column() string {
	switch p {
	case AuditLogEntryOrderFieldCreatedAt:
		return "created_at"
	}

	panic(fmt.Sprintf("unsupported order by: %s", p))
}
