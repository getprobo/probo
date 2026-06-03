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
