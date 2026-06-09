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
)

type SCIMBridgeType string

const (
	SCIMBridgeTypeGoogleWorkspace SCIMBridgeType = "GOOGLE_WORKSPACE"
	SCIMBridgeTypeMicrosoft365    SCIMBridgeType = "MICROSOFT_365"
)

var (
	_ fmt.Stringer             = SCIMBridgeType("")
	_ encoding.TextMarshaler   = SCIMBridgeType("")
	_ encoding.TextUnmarshaler = (*SCIMBridgeType)(nil)
)

func SCIMBridgeTypes() []SCIMBridgeType {
	return []SCIMBridgeType{
		SCIMBridgeTypeGoogleWorkspace,
		SCIMBridgeTypeMicrosoft365,
	}
}

func (v SCIMBridgeType) IsValid() bool {
	switch v {
	case
		SCIMBridgeTypeGoogleWorkspace,
		SCIMBridgeTypeMicrosoft365:
		return true
	}

	return false
}

func (v SCIMBridgeType) String() string {
	return string(v)
}

func (v SCIMBridgeType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *SCIMBridgeType) UnmarshalText(text []byte) error {
	val := SCIMBridgeType(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid SCIMBridgeType value: %q", string(text))
	}

	*v = val

	return nil
}
