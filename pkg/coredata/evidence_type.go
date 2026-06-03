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
)

type (
	EvidenceType string
)

const (
	EvidenceTypeFile EvidenceType = "FILE"
	EvidenceTypeLink EvidenceType = "LINK"
)

var (
	_ fmt.Stringer             = EvidenceType("")
	_ encoding.TextMarshaler   = EvidenceType("")
	_ encoding.TextUnmarshaler = (*EvidenceType)(nil)
)

func EvidenceTypes() []EvidenceType {
	return []EvidenceType{
		EvidenceTypeFile,
		EvidenceTypeLink,
	}
}

func (v EvidenceType) IsValid() bool {
	switch v {
	case
		EvidenceTypeFile,
		EvidenceTypeLink:
		return true
	}

	return false
}

func (v EvidenceType) String() string {
	return string(v)
}

func (v EvidenceType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *EvidenceType) UnmarshalText(text []byte) error {
	val := EvidenceType(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid EvidenceType value: %q", string(text))
	}

	*v = val

	return nil
}
