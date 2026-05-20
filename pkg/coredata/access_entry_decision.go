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

type AccessEntryDecision string

const (
	AccessEntryDecisionPending  AccessEntryDecision = "PENDING"
	AccessEntryDecisionApproved AccessEntryDecision = "APPROVED"
	AccessEntryDecisionRevoke   AccessEntryDecision = "REVOKE"
	AccessEntryDecisionDefer    AccessEntryDecision = "DEFER"
	AccessEntryDecisionEscalate AccessEntryDecision = "ESCALATE"
)

var (
	_ fmt.Stringer             = AccessEntryDecision("")
	_ encoding.TextMarshaler   = AccessEntryDecision("")
	_ encoding.TextUnmarshaler = (*AccessEntryDecision)(nil)
)

func AccessEntryDecisions() []AccessEntryDecision {
	return []AccessEntryDecision{
		AccessEntryDecisionPending,
		AccessEntryDecisionApproved,
		AccessEntryDecisionRevoke,
		AccessEntryDecisionDefer,
		AccessEntryDecisionEscalate,
	}
}

func (v AccessEntryDecision) IsValid() bool {
	switch v {
	case
		AccessEntryDecisionPending,
		AccessEntryDecisionApproved,
		AccessEntryDecisionRevoke,
		AccessEntryDecisionDefer,
		AccessEntryDecisionEscalate:
		return true
	}

	return false
}

func (v AccessEntryDecision) String() string {
	return string(v)
}

func (v AccessEntryDecision) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AccessEntryDecision) UnmarshalText(text []byte) error {
	val := AccessEntryDecision(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid AccessEntryDecision value: %q", string(text))
	}

	*v = val

	return nil
}
