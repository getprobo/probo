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

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"

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

	maps.Copy(cp.ApprovalInput, decisions)

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
