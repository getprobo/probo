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

type OAuth2TokenTypeHint string

const (
	OAuth2TokenTypeHintAccessToken  OAuth2TokenTypeHint = "access_token"
	OAuth2TokenTypeHintRefreshToken OAuth2TokenTypeHint = "refresh_token"
)

func (h OAuth2TokenTypeHint) IsValid() bool {
	switch h {
	case OAuth2TokenTypeHintAccessToken,
		OAuth2TokenTypeHintRefreshToken:
		return true
	}

	return false
}

func (h OAuth2TokenTypeHint) String() string { return string(h) }

func (h *OAuth2TokenTypeHint) UnmarshalText(text []byte) error {
	*h = OAuth2TokenTypeHint(text)
	if !h.IsValid() {
		return fmt.Errorf("%s is not a valid OAuth2TokenTypeHint", string(text))
	}

	return nil
}

func (h OAuth2TokenTypeHint) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}
