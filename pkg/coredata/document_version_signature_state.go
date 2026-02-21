// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	"database/sql/driver"
	"fmt"
	"strings"
)

type (
	DocumentVersionSignatureState  string
	DocumentVersionSignatureStates []DocumentVersionSignatureState
)

const (
	DocumentVersionSignatureStateRequested DocumentVersionSignatureState = "REQUESTED"
	DocumentVersionSignatureStateSigned    DocumentVersionSignatureState = "SIGNED"
)

func (pvs DocumentVersionSignatureState) MarshalText() ([]byte, error) {
	return []byte(pvs.String()), nil
}

func (pvs *DocumentVersionSignatureState) UnmarshalText(data []byte) error {
	val := string(data)

	switch val {
	case DocumentVersionSignatureStateRequested.String():
		*pvs = DocumentVersionSignatureStateRequested
	case DocumentVersionSignatureStateSigned.String():
		*pvs = DocumentVersionSignatureStateSigned
	default:
		return fmt.Errorf("invalid DocumentVersionSignatureState value: %q", val)
	}

	return nil
}

func (pvs DocumentVersionSignatureState) String() string {
	var val string

	switch pvs {
	case DocumentVersionSignatureStateRequested:
		val = "REQUESTED"
	case DocumentVersionSignatureStateSigned:
		val = "SIGNED"
	default:
		panic(fmt.Errorf("invalid DocumentVersionSignatureState value: %q", string(pvs)))
	}

	return val
}

func (pvs *DocumentVersionSignatureState) Scan(value any) error {
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid scan source for DocumentVersionSignatureState, expected string got %T", value)
	}

	return pvs.UnmarshalText([]byte(val))
}

func (pvs DocumentVersionSignatureState) Value() (driver.Value, error) {
	return pvs.String(), nil
}

func (states DocumentVersionSignatureStates) Value() (driver.Value, error) {
	if len(states) == 0 {
		return nil, nil
	}

	var result strings.Builder
	result.WriteString("{")
	for i, state := range states {
		if i > 0 {
			result.WriteString(",")
		}
		fmt.Fprintf(&result, "%q", state.String())
	}
	result.WriteString("}")
	return result.String(), nil
}
