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

type FindingOrderField string

const (
	FindingOrderFieldCreatedAt    FindingOrderField = "CREATED_AT"
	FindingOrderFieldIdentifiedOn FindingOrderField = "IDENTIFIED_ON"
	FindingOrderFieldDueDate      FindingOrderField = "DUE_DATE"
	FindingOrderFieldStatus       FindingOrderField = "STATUS"
	FindingOrderFieldPriority     FindingOrderField = "PRIORITY"
	FindingOrderFieldReferenceId  FindingOrderField = "REFERENCE_ID"
	FindingOrderFieldKind         FindingOrderField = "KIND"
)

var (
	_ page.OrderField          = FindingOrderField("")
	_ fmt.Stringer             = FindingOrderField("")
	_ encoding.TextMarshaler   = FindingOrderField("")
	_ encoding.TextUnmarshaler = (*FindingOrderField)(nil)
)

func FindingOrderFields() []FindingOrderField {
	return []FindingOrderField{
		FindingOrderFieldCreatedAt,
		FindingOrderFieldIdentifiedOn,
		FindingOrderFieldDueDate,
		FindingOrderFieldStatus,
		FindingOrderFieldPriority,
		FindingOrderFieldReferenceId,
		FindingOrderFieldKind,
	}
}

func (v FindingOrderField) IsValid() bool {
	switch v {
	case
		FindingOrderFieldCreatedAt,
		FindingOrderFieldIdentifiedOn,
		FindingOrderFieldDueDate,
		FindingOrderFieldStatus,
		FindingOrderFieldPriority,
		FindingOrderFieldReferenceId,
		FindingOrderFieldKind:
		return true
	}

	return false
}

func (v FindingOrderField) String() string {
	return string(v)
}

func (v FindingOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *FindingOrderField) UnmarshalText(text []byte) error {
	val := FindingOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid FindingOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p FindingOrderField) Column() string {
	return string(p)
}
