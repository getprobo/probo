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

	"go.probo.inc/probo/pkg/llm"
)

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
