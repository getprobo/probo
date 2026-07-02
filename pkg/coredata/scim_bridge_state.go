// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

type SCIMBridgeState string

const (
	SCIMBridgeStatePending  SCIMBridgeState = "PENDING"
	SCIMBridgeStateActive   SCIMBridgeState = "ACTIVE"
	SCIMBridgeStateSyncing  SCIMBridgeState = "SYNCING"
	SCIMBridgeStateFailed   SCIMBridgeState = "FAILED"
	SCIMBridgeStateDisabled SCIMBridgeState = "DISABLED"
)

var (
	_ fmt.Stringer             = SCIMBridgeState("")
	_ encoding.TextMarshaler   = SCIMBridgeState("")
	_ encoding.TextUnmarshaler = (*SCIMBridgeState)(nil)
)

func SCIMBridgeStates() []SCIMBridgeState {
	return []SCIMBridgeState{
		SCIMBridgeStatePending,
		SCIMBridgeStateActive,
		SCIMBridgeStateSyncing,
		SCIMBridgeStateFailed,
		SCIMBridgeStateDisabled,
	}
}

func (v SCIMBridgeState) IsValid() bool {
	switch v {
	case
		SCIMBridgeStatePending,
		SCIMBridgeStateActive,
		SCIMBridgeStateSyncing,
		SCIMBridgeStateFailed,
		SCIMBridgeStateDisabled:
		return true
	}

	return false
}

func (v SCIMBridgeState) String() string {
	return string(v)
}

func (v SCIMBridgeState) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *SCIMBridgeState) UnmarshalText(text []byte) error {
	val := SCIMBridgeState(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid SCIMBridgeState value: %q", string(text))
	}

	*v = val

	return nil
}
