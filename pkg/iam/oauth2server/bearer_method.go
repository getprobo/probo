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

package oauth2server

import (
	"encoding"
	"fmt"
)

type (
	// BearerMethod is an RFC 6750 bearer token presentation method named in
	// RFC 9728 bearer_methods_supported.
	BearerMethod string

	// BearerMethods is a list of bearer token presentation methods.
	BearerMethods []BearerMethod
)

const (
	BearerMethodHeader BearerMethod = "header"
	BearerMethodBody   BearerMethod = "body"
	BearerMethodQuery  BearerMethod = "query"
)

var (
	_ encoding.TextMarshaler   = BearerMethod("")
	_ encoding.TextUnmarshaler = (*BearerMethod)(nil)
)

func (m BearerMethod) IsValid() bool {
	switch m {
	case BearerMethodHeader, BearerMethodBody, BearerMethodQuery:
		return true
	}

	return false
}

func (m BearerMethod) String() string {
	return string(m)
}

func (m BearerMethod) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

func (m *BearerMethod) UnmarshalText(text []byte) error {
	val := BearerMethod(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid BearerMethod value: %q", string(text))
	}

	*m = val

	return nil
}
