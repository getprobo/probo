// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	DocumentVersionOrderField string
)

const (
	DocumentVersionOrderFieldCreatedAt DocumentVersionOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = DocumentVersionOrderField("")
	_ fmt.Stringer             = DocumentVersionOrderField("")
	_ encoding.TextMarshaler   = DocumentVersionOrderField("")
	_ encoding.TextUnmarshaler = (*DocumentVersionOrderField)(nil)
)

func DocumentVersionOrderFields() []DocumentVersionOrderField {
	return []DocumentVersionOrderField{
		DocumentVersionOrderFieldCreatedAt,
	}
}

func (v DocumentVersionOrderField) IsValid() bool {
	switch v {
	case
		DocumentVersionOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v DocumentVersionOrderField) String() string {
	return string(v)
}

func (v DocumentVersionOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DocumentVersionOrderField) UnmarshalText(text []byte) error {
	val := DocumentVersionOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DocumentVersionOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p DocumentVersionOrderField) Column() string {
	return string(p)
}
