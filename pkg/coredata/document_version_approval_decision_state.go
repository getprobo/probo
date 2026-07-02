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
	"database/sql/driver"
	"encoding"
	"fmt"
	"strings"
)

type (
	DocumentVersionApprovalDecisionState  string
	DocumentVersionApprovalDecisionStates []DocumentVersionApprovalDecisionState
)

const (
	DocumentVersionApprovalDecisionStatePending  DocumentVersionApprovalDecisionState = "PENDING"
	DocumentVersionApprovalDecisionStateApproved DocumentVersionApprovalDecisionState = "APPROVED"
	DocumentVersionApprovalDecisionStateRejected DocumentVersionApprovalDecisionState = "REJECTED"
	DocumentVersionApprovalDecisionStateVoided   DocumentVersionApprovalDecisionState = "VOIDED"
)

var (
	_ fmt.Stringer             = DocumentVersionApprovalDecisionState("")
	_ encoding.TextMarshaler   = DocumentVersionApprovalDecisionState("")
	_ encoding.TextUnmarshaler = (*DocumentVersionApprovalDecisionState)(nil)
)

func (v DocumentVersionApprovalDecisionState) IsValid() bool {
	switch v {
	case
		DocumentVersionApprovalDecisionStatePending,
		DocumentVersionApprovalDecisionStateApproved,
		DocumentVersionApprovalDecisionStateRejected,
		DocumentVersionApprovalDecisionStateVoided:
		return true
	}

	return false
}

func (v DocumentVersionApprovalDecisionState) String() string {
	return string(v)
}

func (v DocumentVersionApprovalDecisionState) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DocumentVersionApprovalDecisionState) UnmarshalText(text []byte) error {
	val := DocumentVersionApprovalDecisionState(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DocumentVersionApprovalDecisionState value: %q", string(text))
	}

	*v = val

	return nil
}

func (states DocumentVersionApprovalDecisionStates) Value() (driver.Value, error) {
	if len(states) == 0 {
		return nil, nil
	}

	var result strings.Builder
	result.WriteString("{")

	for i, state := range states {
		if i > 0 {
			result.WriteString(",")
		}

		fmt.Fprintf(&result, "%q", state.String())
	}

	result.WriteString("}")

	return result.String(), nil
}
