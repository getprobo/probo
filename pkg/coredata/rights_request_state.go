// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
	RightsRequestStateRejected   RightsRequestState = "REJECTED"
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
		RightsRequestStateRejected,
	}
}

func (v RightsRequestState) IsValid() bool {
	switch v {
	case
		RightsRequestStateTodo,
		RightsRequestStateInProgress,
		RightsRequestStateDone,
		RightsRequestStateRejected:
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
