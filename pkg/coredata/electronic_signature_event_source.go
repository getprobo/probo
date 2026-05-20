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
)

type (
	ElectronicSignatureEventSource string
)

const (
	ElectronicSignatureEventSourceClient ElectronicSignatureEventSource = "CLIENT"
	ElectronicSignatureEventSourceServer ElectronicSignatureEventSource = "SERVER"
)

var (
	_ fmt.Stringer             = ElectronicSignatureEventSource("")
	_ encoding.TextMarshaler   = ElectronicSignatureEventSource("")
	_ encoding.TextUnmarshaler = (*ElectronicSignatureEventSource)(nil)
)

func ElectronicSignatureEventSources() []ElectronicSignatureEventSource {
	return []ElectronicSignatureEventSource{
		ElectronicSignatureEventSourceClient,
		ElectronicSignatureEventSourceServer,
	}
}

func (v ElectronicSignatureEventSource) IsValid() bool {
	switch v {
	case
		ElectronicSignatureEventSourceClient,
		ElectronicSignatureEventSourceServer:
		return true
	}

	return false
}

func (v ElectronicSignatureEventSource) String() string {
	return string(v)
}

func (v ElectronicSignatureEventSource) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ElectronicSignatureEventSource) UnmarshalText(text []byte) error {
	val := ElectronicSignatureEventSource(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ElectronicSignatureEventSource value: %q", string(text))
	}

	*v = val

	return nil
}
