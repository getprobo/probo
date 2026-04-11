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

type (
	OAuth2ResponseType  string
	OAuth2ResponseTypes []OAuth2ResponseType
)

const (
	OAuth2ResponseTypeCode OAuth2ResponseType = "code"
)

func (r OAuth2ResponseType) IsValid() bool {
	switch r {
	case OAuth2ResponseTypeCode:
		return true
	}

	return false
}

func (r OAuth2ResponseType) String() string { return string(r) }

func (r *OAuth2ResponseType) UnmarshalText(text []byte) error {
	*r = OAuth2ResponseType(text)
	if !r.IsValid() {
		return fmt.Errorf("%s is not a valid OAuth2ResponseType", string(text))
	}

	return nil
}

func (r OAuth2ResponseType) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}
