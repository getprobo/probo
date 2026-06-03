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

type RightsRequestState string

const (
	RightsRequestStateTodo       RightsRequestState = "TODO"
	RightsRequestStateInProgress RightsRequestState = "IN_PROGRESS"
	RightsRequestStateDone       RightsRequestState = "DONE"
)

var (
	_ fmt.Stringer             = RightsRequestState("")
	_ encoding.TextMarshaler   = RightsRequestState("")
	_ encoding.TextUnmarshaler = (*RightsRequestState)(nil)
)

func RightsRequestStates() []RightsRequestState {
	return []RightsRequestState{
		RightsRequestStateTodo,
		RightsRequestStateInProgress,
		RightsRequestStateDone,
	}
}

func (v RightsRequestState) IsValid() bool {
	switch v {
	case
		RightsRequestStateTodo,
		RightsRequestStateInProgress,
		RightsRequestStateDone:
		return true
	}

	return false
}

func (v RightsRequestState) String() string {
	return string(v)
}

func (v RightsRequestState) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *RightsRequestState) UnmarshalText(text []byte) error {
	val := RightsRequestState(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid RightsRequestState value: %q", string(text))
	}

	*v = val

	return nil
}
