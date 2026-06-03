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

type TrustCenterAccessState string

const (
	TrustCenterAccessStateActive   TrustCenterAccessState = "ACTIVE"
	TrustCenterAccessStateInactive TrustCenterAccessState = "INACTIVE"
)

var (
	_ fmt.Stringer             = TrustCenterAccessState("")
	_ encoding.TextMarshaler   = TrustCenterAccessState("")
	_ encoding.TextUnmarshaler = (*TrustCenterAccessState)(nil)
)

func TrustCenterAccessStates() []TrustCenterAccessState {
	return []TrustCenterAccessState{
		TrustCenterAccessStateActive,
		TrustCenterAccessStateInactive,
	}
}

func (v TrustCenterAccessState) IsValid() bool {
	switch v {
	case
		TrustCenterAccessStateActive,
		TrustCenterAccessStateInactive:
		return true
	}

	return false
}

func (v TrustCenterAccessState) String() string {
	return string(v)
}

func (v TrustCenterAccessState) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *TrustCenterAccessState) UnmarshalText(text []byte) error {
	val := TrustCenterAccessState(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid TrustCenterAccessState value: %q", string(text))
	}

	*v = val

	return nil
}
