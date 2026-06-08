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

package agentrun

import (
	"errors"
)

var (
	// ErrAgentRunNotFound is returned when the target agent run does not
	// exist. It wraps coredata.ErrResourceNotFound so callers depend on the
	// agentrun API rather than the underlying data layer.
	ErrAgentRunNotFound = errors.New("agent run not found")

	// ErrNotAwaitingApproval is returned when an approval decision is
	// submitted for a run that is not currently parked in AWAITING_APPROVAL.
	ErrNotAwaitingApproval = errors.New("agent run is not awaiting approval")

	// ErrApprovalDecisionsMismatch is returned when the submitted decisions
	// do not cover exactly the run's pending approvals. It shields callers
	// from the agent package's internal mismatch error.
	ErrApprovalDecisionsMismatch = errors.New("approval decisions do not match the run's pending approvals")
)
