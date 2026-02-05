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
	"time"

	"github.com/jackc/pgx/v5"
)

type (
	MembershipProfileFilter struct {
		excludeContractEnded *bool
		currentDate          time.Time
	}
)

func NewMembershipProfileFilter(excludeContractEnded *bool) *MembershipProfileFilter {
	return &MembershipProfileFilter{
		excludeContractEnded: excludeContractEnded,
		currentDate:          time.Now(),
	}
}

func (f *MembershipProfileFilter) SQLArguments() pgx.StrictNamedArgs {
	return pgx.StrictNamedArgs{
		"exclude_contract_ended": f.excludeContractEnded,
		"current_date":           f.currentDate,
	}
}

func (f *MembershipProfileFilter) SQLFragment() string {
	return `
(
    CASE
        WHEN @exclude_contract_ended::boolean IS NOT NULL AND @exclude_contract_ended::boolean = true THEN
            (contract_end_date IS NULL OR contract_end_date >= @current_date::date)
        ELSE TRUE
    END
)`
}
