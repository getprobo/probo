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

// RegulationSource records whether the regulation applied to a consent
// was detected from the visitor's geolocation or defaulted because
// geolocation was unresolved or mapped to no known regulation.
type RegulationSource string

const (
	RegulationSourceDetected RegulationSource = "DETECTED"
	RegulationSourceDefault  RegulationSource = "DEFAULT"
)

var (
	_ fmt.Stringer             = RegulationSource("")
	_ encoding.TextMarshaler   = RegulationSource("")
	_ encoding.TextUnmarshaler = (*RegulationSource)(nil)
)

func RegulationSources() []RegulationSource {
	return []RegulationSource{
		RegulationSourceDetected,
		RegulationSourceDefault,
	}
}

func (v RegulationSource) IsValid() bool {
	switch v {
	case
		RegulationSourceDetected,
		RegulationSourceDefault:
		return true
	}

	return false
}

func (v RegulationSource) String() string {
	return string(v)
}

func (v RegulationSource) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *RegulationSource) UnmarshalText(text []byte) error {
	val := RegulationSource(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid RegulationSource value: %q", string(text))
	}

	*v = val

	return nil
}

func ParseRegulationSource(s string) (RegulationSource, error) {
	switch RegulationSource(s) {
	case RegulationSourceDetected:
		return RegulationSourceDetected, nil
	case RegulationSourceDefault:
		return RegulationSourceDefault, nil
	default:
		return "", fmt.Errorf("invalid RegulationSource value: %q", s)
	}
}
