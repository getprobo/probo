// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	ExpireReason string
)

const (
	ExpireReasonIdleTimeout ExpireReason = "idle_timeout"
	ExpireReasonRevoked     ExpireReason = "revoked"
	ExpireReasonClosed      ExpireReason = "closed"
)

var (
	_ fmt.Stringer             = ExpireReason("")
	_ encoding.TextMarshaler   = ExpireReason("")
	_ encoding.TextUnmarshaler = (*ExpireReason)(nil)
)

func ExpireReasons() []ExpireReason {
	return []ExpireReason{
		ExpireReasonIdleTimeout,
		ExpireReasonRevoked,
		ExpireReasonClosed,
	}
}

func (v ExpireReason) IsValid() bool {
	switch v {
	case
		ExpireReasonIdleTimeout,
		ExpireReasonRevoked,
		ExpireReasonClosed:
		return true
	}

	return false
}

func (v ExpireReason) String() string {
	return string(v)
}

func (v ExpireReason) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ExpireReason) UnmarshalText(text []byte) error {
	val := ExpireReason(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ExpireReason value: %q", string(text))
	}

	*v = val

	return nil
}
