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

type OAuth2CodeChallengeMethod string

const (
	OAuth2CodeChallengeMethodS256 OAuth2CodeChallengeMethod = "S256"
)

var (
	_ fmt.Stringer             = OAuth2CodeChallengeMethod("")
	_ encoding.TextMarshaler   = OAuth2CodeChallengeMethod("")
	_ encoding.TextUnmarshaler = (*OAuth2CodeChallengeMethod)(nil)
)

func OAuth2CodeChallengeMethods() []OAuth2CodeChallengeMethod {
	return []OAuth2CodeChallengeMethod{
		OAuth2CodeChallengeMethodS256,
	}
}

func (v OAuth2CodeChallengeMethod) IsValid() bool {
	switch v {
	case
		OAuth2CodeChallengeMethodS256:
		return true
	}

	return false
}

func (v OAuth2CodeChallengeMethod) String() string {
	return string(v)
}

func (v OAuth2CodeChallengeMethod) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *OAuth2CodeChallengeMethod) UnmarshalText(text []byte) error {
	val := OAuth2CodeChallengeMethod(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid OAuth2CodeChallengeMethod value: %q", string(text))
	}

	*v = val

	return nil
}
