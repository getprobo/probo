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
	"github.com/jackc/pgx/v5"
)

type (
	DocumentVersionApprovalDecisionFilter struct {
		states DocumentVersionApprovalDecisionStates
	}
)

func NewDocumentVersionApprovalDecisionFilter(states []DocumentVersionApprovalDecisionState) *DocumentVersionApprovalDecisionFilter {
	if len(states) == 0 {
		states = nil
	}
	return &DocumentVersionApprovalDecisionFilter{
		states: DocumentVersionApprovalDecisionStates(states),
	}
}

func (f *DocumentVersionApprovalDecisionFilter) SQLArguments() pgx.StrictNamedArgs {
	return pgx.StrictNamedArgs{
		"filter_states": f.states,
	}
}

func (f *DocumentVersionApprovalDecisionFilter) SQLFragment() string {
	return `
(
    CASE
        WHEN @filter_states::document_version_approval_decision_state[] IS NOT NULL THEN
            state = ANY(@filter_states::document_version_approval_decision_state[])
        ELSE TRUE
    END
)`
}
