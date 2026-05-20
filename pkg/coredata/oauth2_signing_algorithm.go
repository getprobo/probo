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

type OAuth2SigningAlgorithm string

const (
	OAuth2SigningAlgorithmRS256 OAuth2SigningAlgorithm = "RS256"
)

var (
	_ fmt.Stringer             = OAuth2SigningAlgorithm("")
	_ encoding.TextMarshaler   = OAuth2SigningAlgorithm("")
	_ encoding.TextUnmarshaler = (*OAuth2SigningAlgorithm)(nil)
)

func OAuth2SigningAlgorithms() []OAuth2SigningAlgorithm {
	return []OAuth2SigningAlgorithm{
		OAuth2SigningAlgorithmRS256,
	}
}

func (v OAuth2SigningAlgorithm) IsValid() bool {
	switch v {
	case
		OAuth2SigningAlgorithmRS256:
		return true
	}

	return false
}

func (v OAuth2SigningAlgorithm) String() string {
	return string(v)
}

func (v OAuth2SigningAlgorithm) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *OAuth2SigningAlgorithm) UnmarshalText(text []byte) error {
	val := OAuth2SigningAlgorithm(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid OAuth2SigningAlgorithm value: %q", string(text))
	}

	*v = val

	return nil
}
