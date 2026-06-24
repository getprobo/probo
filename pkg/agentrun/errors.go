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
