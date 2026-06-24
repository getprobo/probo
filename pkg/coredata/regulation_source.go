// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
