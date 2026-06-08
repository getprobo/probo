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

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"go.probo.inc/probo/pkg/llm"
)

// ErrApprovalDecisionsMismatch is returned by MergeApprovalDecisions when
// the supplied decisions do not cover exactly the checkpoint's pending
// approvals. A missing decision would resume as an implicit denial, so
// partial submissions are rejected.
var ErrApprovalDecisionsMismatch = errors.New("approval decisions do not match the checkpoint's pending approvals")

type (
	ApprovalConfig struct {
		ToolNames     []string
		ShouldApprove func(ctx context.Context, toolCall llm.ToolCall) bool

		toolNameSet map[string]struct{}
	}

	ApprovalResult struct {
		Approved bool
		Message  string
	}

	ResumeInput struct {
		Approvals map[string]ApprovalResult
	}
)

// MergeApprovalDecisions decodes an awaiting-approval checkpoint, records
// the human decisions into its ApprovalInput, and returns the re-encoded
// checkpoint ready to be persisted. decisions is keyed by pending
// tool-call ID and must cover exactly the checkpoint's pending approvals;
// otherwise ErrApprovalDecisionsMismatch is returned.
func MergeApprovalDecisions(
	raw json.RawMessage,
	decisions map[string]ApprovalResult,
) (json.RawMessage, error) {
	var cp Checkpoint
	if err := json.Unmarshal(raw, &cp); err != nil {
		return nil, fmt.Errorf("cannot unmarshal checkpoint: %w", err)
	}

	pending := make(map[string]struct{}, len(cp.PendingApprovals))
	for _, toolCall := range cp.PendingApprovals {
		pending[toolCall.ID] = struct{}{}
	}

	if len(decisions) != len(pending) {
		return nil, ErrApprovalDecisionsMismatch
	}

	for id := range decisions {
		if _, ok := pending[id]; !ok {
			return nil, ErrApprovalDecisionsMismatch
		}
	}

	if cp.ApprovalInput == nil {
		cp.ApprovalInput = make(map[string]ApprovalResult, len(decisions))
	}
	for id, decision := range decisions {
		cp.ApprovalInput[id] = decision
	}

	data, err := json.Marshal(&cp)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal checkpoint: %w", err)
	}

	return data, nil
}

func buildToolNameSet(names []string) map[string]struct{} {
	if len(names) == 0 {
		return nil
	}

	set := make(map[string]struct{}, len(names))
	for _, name := range names {
		set[name] = struct{}{}
	}

	return set
}

func (c *ApprovalConfig) requiresApproval(ctx context.Context, tc llm.ToolCall) bool {
	if c == nil {
		return false
	}

	if c.ShouldApprove != nil {
		return c.ShouldApprove(ctx, tc)
	}

	_, ok := c.toolNameSet[tc.Function.Name]

	return ok
}
