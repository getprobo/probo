// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	DocumentVersionSignatureFilter struct {
		states         DocumentVersionSignatureStates
		activeContract *bool
		profileState   *ProfileState
	}
)

func NewDocumentVersionSignatureFilter(states []DocumentVersionSignatureState, activeContract *bool, profileState *ProfileState) *DocumentVersionSignatureFilter {
	return &DocumentVersionSignatureFilter{
		states:         DocumentVersionSignatureStates(states),
		activeContract: activeContract,
		profileState:   profileState,
	}
}

func (f *DocumentVersionSignatureFilter) SQLArguments() pgx.StrictNamedArgs {
	return pgx.StrictNamedArgs{
		"states":          f.states,
		"active_contract": f.activeContract,
		"profile_state":   f.profileState,
	}
}

func (f *DocumentVersionSignatureFilter) SQLFragment() string {
	return `
(
    CASE
        WHEN @states::policy_version_signature_state[] IS NOT NULL THEN
            state = ANY(@states::policy_version_signature_state[])
        ELSE TRUE
    END
    AND
    CASE
    WHEN @active_contract::boolean IS NULL
        THEN TRUE
    ELSE EXISTS (
        SELECT 1
        FROM iam_membership_profiles p
        WHERE p.id = signed_by_profile_id
        AND (
            (
                @active_contract::boolean = TRUE
                AND (
                    p.contract_end_date IS NULL OR p.contract_end_date >= CURRENT_DATE
                )
            ) OR (
                @active_contract::boolean = FALSE
                AND (
                    p.contract_end_date IS NOT NULL AND p.contract_end_date < CURRENT_DATE
                )
            )
        )
    )
    END
    AND
    CASE
    WHEN @profile_state::text IS NULL
        THEN TRUE
    ELSE EXISTS (
        SELECT 1
        FROM iam_membership_profiles p
        WHERE p.id = signed_by_profile_id
        AND p.state = @profile_state::membership_state
    )
    END
)`
}
