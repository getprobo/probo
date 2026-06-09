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

type ProfileSource string

const (
	ProfileSourceManual ProfileSource = "MANUAL"
	ProfileSourceSAML   ProfileSource = "SAML"
	ProfileSourceSCIM   ProfileSource = "SCIM"
)

var (
	_ fmt.Stringer             = ProfileSource("")
	_ encoding.TextMarshaler   = ProfileSource("")
	_ encoding.TextUnmarshaler = (*ProfileSource)(nil)
)

func ProfileSources() []ProfileSource {
	return []ProfileSource{
		ProfileSourceManual,
		ProfileSourceSAML,
		ProfileSourceSCIM,
	}
}

func (v ProfileSource) IsValid() bool {
	switch v {
	case
		ProfileSourceManual,
		ProfileSourceSAML,
		ProfileSourceSCIM:
		return true
	}

	return false
}

func (v ProfileSource) String() string {
	return string(v)
}

func (v ProfileSource) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ProfileSource) UnmarshalText(text []byte) error {
	val := ProfileSource(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ProfileSource value: %q", string(text))
	}

	*v = val

	return nil
}
