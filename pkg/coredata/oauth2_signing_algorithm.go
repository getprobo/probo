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

import "fmt"

type OAuth2SigningAlgorithm string

const (
	OAuth2SigningAlgorithmRS256 OAuth2SigningAlgorithm = "RS256"
)

func (a OAuth2SigningAlgorithm) IsValid() bool {
	switch a {
	case OAuth2SigningAlgorithmRS256:
		return true
	}

	return false
}

func (a OAuth2SigningAlgorithm) String() string { return string(a) }

func (a *OAuth2SigningAlgorithm) UnmarshalText(text []byte) error {
	*a = OAuth2SigningAlgorithm(text)
	if !a.IsValid() {
		return fmt.Errorf("%s is not a valid OAuth2SigningAlgorithm", string(text))
	}

	return nil
}

func (a OAuth2SigningAlgorithm) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}
