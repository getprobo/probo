// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	SCIMBridgeOrderField string
)

const (
	SCIMBridgeOrderFieldCreatedAt SCIMBridgeOrderField = "CREATED_AT"
)

var (
	_ page.OrderField          = SCIMBridgeOrderField("")
	_ fmt.Stringer             = SCIMBridgeOrderField("")
	_ encoding.TextMarshaler   = SCIMBridgeOrderField("")
	_ encoding.TextUnmarshaler = (*SCIMBridgeOrderField)(nil)
)

func SCIMBridgeOrderFields() []SCIMBridgeOrderField {
	return []SCIMBridgeOrderField{
		SCIMBridgeOrderFieldCreatedAt,
	}
}

func (v SCIMBridgeOrderField) IsValid() bool {
	switch v {
	case
		SCIMBridgeOrderFieldCreatedAt:
		return true
	}

	return false
}

func (v SCIMBridgeOrderField) String() string {
	return string(v)
}

func (v SCIMBridgeOrderField) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *SCIMBridgeOrderField) UnmarshalText(text []byte) error {
	val := SCIMBridgeOrderField(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid SCIMBridgeOrderField value: %q", string(text))
	}

	*v = val

	return nil
}

func (p SCIMBridgeOrderField) Column() string {
	return string(p)
}
