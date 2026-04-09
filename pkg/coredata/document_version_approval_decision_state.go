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

func (s DocumentVersionApprovalDecisionState) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *DocumentVersionApprovalDecisionState) UnmarshalText(data []byte) error {
	val := string(data)

	switch val {
	case DocumentVersionApprovalDecisionStatePending.String():
		*s = DocumentVersionApprovalDecisionStatePending
	case DocumentVersionApprovalDecisionStateApproved.String():
		*s = DocumentVersionApprovalDecisionStateApproved
	case DocumentVersionApprovalDecisionStateRejected.String():
		*s = DocumentVersionApprovalDecisionStateRejected
	case DocumentVersionApprovalDecisionStateVoided.String():
		*s = DocumentVersionApprovalDecisionStateVoided
	default:
		return fmt.Errorf("invalid DocumentVersionApprovalDecisionState value: %q", val)
	}

	return nil
}

func (s DocumentVersionApprovalDecisionState) String() string {
	var val string

	switch s {
	case DocumentVersionApprovalDecisionStatePending:
		val = "PENDING"
	case DocumentVersionApprovalDecisionStateApproved:
		val = "APPROVED"
	case DocumentVersionApprovalDecisionStateRejected:
		val = "REJECTED"
	case DocumentVersionApprovalDecisionStateVoided:
		val = "VOIDED"
	default:
		panic(fmt.Errorf("invalid DocumentVersionApprovalDecisionState value: %q", string(s)))
	}

	return val
}

func (s *DocumentVersionApprovalDecisionState) Scan(value any) error {
	val, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid scan source for DocumentVersionApprovalDecisionState, expected string got %T", value)
	}

	return s.UnmarshalText([]byte(val))
}

func (s DocumentVersionApprovalDecisionState) Value() (driver.Value, error) {
	return s.String(), nil
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
