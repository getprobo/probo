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

package coredata

import (
	"github.com/jackc/pgx/v5"
)

type (
	DocumentVersionApprovalDecisionFilter struct {
		states DocumentVersionApprovalDecisionStates
		query  *string
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

// WithQuery keeps the decisions whose approver's name contains query.
func (f *DocumentVersionApprovalDecisionFilter) WithQuery(query *string) *DocumentVersionApprovalDecisionFilter {
	f.query = query
	return f
}

func (f *DocumentVersionApprovalDecisionFilter) SQLArguments() pgx.StrictNamedArgs {
	var filterQuery *string

	if f.query != nil && *f.query != "" {
		escaped := escapeLikePattern(*f.query)
		filterQuery = &escaped
	}

	return pgx.StrictNamedArgs{
		"filter_states": f.states,
		"filter_query":  filterQuery,
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
    AND
    CASE
        WHEN @filter_query::text IS NOT NULL THEN
            approver_id IN (
                SELECT id
                FROM iam_membership_profiles
                WHERE full_name ILIKE '%' || @filter_query::text || '%' ESCAPE '\'
            )
        ELSE TRUE
    END
)`
}
