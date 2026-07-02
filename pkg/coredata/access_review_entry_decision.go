// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

type AccessReviewEntryDecision string

const (
	AccessReviewEntryDecisionPending  AccessReviewEntryDecision = "PENDING"
	AccessReviewEntryDecisionApproved AccessReviewEntryDecision = "APPROVED"
	AccessReviewEntryDecisionRevoke   AccessReviewEntryDecision = "REVOKE"
	AccessReviewEntryDecisionDefer    AccessReviewEntryDecision = "DEFER"
	AccessReviewEntryDecisionEscalate AccessReviewEntryDecision = "ESCALATE"
)

var (
	_ fmt.Stringer             = AccessReviewEntryDecision("")
	_ encoding.TextMarshaler   = AccessReviewEntryDecision("")
	_ encoding.TextUnmarshaler = (*AccessReviewEntryDecision)(nil)
)

func AccessReviewEntryDecisions() []AccessReviewEntryDecision {
	return []AccessReviewEntryDecision{
		AccessReviewEntryDecisionPending,
		AccessReviewEntryDecisionApproved,
		AccessReviewEntryDecisionRevoke,
		AccessReviewEntryDecisionDefer,
		AccessReviewEntryDecisionEscalate,
	}
}

func (v AccessReviewEntryDecision) IsValid() bool {
	switch v {
	case
		AccessReviewEntryDecisionPending,
		AccessReviewEntryDecisionApproved,
		AccessReviewEntryDecisionRevoke,
		AccessReviewEntryDecisionDefer,
		AccessReviewEntryDecisionEscalate:
		return true
	}

	return false
}

func (v AccessReviewEntryDecision) String() string {
	return string(v)
}

func (v AccessReviewEntryDecision) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AccessReviewEntryDecision) UnmarshalText(text []byte) error {
	val := AccessReviewEntryDecision(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid AccessReviewEntryDecision value: %q", string(text))
	}

	*v = val

	return nil
}
