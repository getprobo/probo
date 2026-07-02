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

type ProcessingActivityOrderField string

const (
	ProcessingActivityOrderFieldCreatedAt ProcessingActivityOrderField = "CREATED_AT"
	ProcessingActivityOrderFieldName      ProcessingActivityOrderField = "NAME"
)

var (
	_ page.OrderField          = ProcessingActivityOrderField("")
	_ fmt.Stringer             = ProcessingActivityOrderField("")
	_ encoding.TextMarshaler   = ProcessingActivityOrderField("")
	_ encoding.TextUnmarshaler = (*ProcessingActivityOrderField)(nil)
)

func ProcessingActivityOrderFields() []ProcessingActivityOrderField {
	return []ProcessingActivityOrderField{
		ProcessingActivityOrderFieldCreatedAt,
		ProcessingActivityOrderFieldName,
	}
}

func (v ProcessingActivityOrderField) IsValid() bool {
	switch v {
	case
		ProcessingActivityOrderFieldCreatedAt,
		ProcessingActivityOrderFieldName:
		return true
	}

	return false
}

func (v ProcessingActivityOrderField) String() string {
	return string(v)
}

func (v ProcessingActivityOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ProcessingActivityOrderField) UnmarshalText(text []byte) error {
	val := ProcessingActivityOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ProcessingActivityOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p ProcessingActivityOrderField) Column() string {
	return string(p)
}
