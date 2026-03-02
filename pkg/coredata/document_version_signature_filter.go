// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	DocumentVersionSignatureFilter struct {
		states         DocumentVersionSignatureStates
		activeContract *bool
	}
)

func NewDocumentVersionSignatureFilter(states []DocumentVersionSignatureState, activeContract *bool) *DocumentVersionSignatureFilter {
	return &DocumentVersionSignatureFilter{
		states:         DocumentVersionSignatureStates(states),
		activeContract: activeContract,
	}
}

func (f *DocumentVersionSignatureFilter) SQLArguments() pgx.StrictNamedArgs {
	return pgx.StrictNamedArgs{
		"states":          f.states,
		"active_contract": f.activeContract,
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
                    p.contract_end_date IS NULL OR p.contract_end_date > NOW()
                )
            ) OR (
                @active_contract::boolean = FALSE
                AND (
                    p.contract_end_date IS NOT NULL AND p.contract_end_date <= NOW()
                )
            )
        )
    )
    END
)`
}
