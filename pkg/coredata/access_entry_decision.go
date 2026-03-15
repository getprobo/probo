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
	"database/sql/driver"
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

func AccessEntryDecisions() []AccessEntryDecision {
	return []AccessEntryDecision{
		AccessEntryDecisionPending,
		AccessEntryDecisionApproved,
		AccessEntryDecisionRevoke,
		AccessEntryDecisionDefer,
		AccessEntryDecisionEscalate,
	}
}

func (d AccessEntryDecision) String() string {
	return string(d)
}

func (d *AccessEntryDecision) Scan(value any) error {
	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("cannot scan AccessEntryDecision: unsupported type %T", value)
	}

	switch str {
	case "PENDING":
		*d = AccessEntryDecisionPending
	case "APPROVED":
		*d = AccessEntryDecisionApproved
	case "REVOKE":
		*d = AccessEntryDecisionRevoke
	case "DEFER":
		*d = AccessEntryDecisionDefer
	case "ESCALATE":
		*d = AccessEntryDecisionEscalate
	default:
		return fmt.Errorf("cannot parse AccessEntryDecision: invalid value %q", str)
	}
	return nil
}

func (d AccessEntryDecision) Value() (driver.Value, error) {
	return d.String(), nil
}
