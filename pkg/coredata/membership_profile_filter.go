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
	"go.probo.inc/probo/pkg/mail"
)

type (
	MembershipProfileFilter struct {
		excludeContractEnded *bool
		currentDate          time.Time
		email                *mail.Addr
		state                *ProfileState
		source               *ProfileSource
	}
)

func NewMembershipProfileFilter(excludeContractEnded *bool) *MembershipProfileFilter {
	return &MembershipProfileFilter{
		excludeContractEnded: excludeContractEnded,
		currentDate:          time.Now(),
	}
}

func (f *MembershipProfileFilter) WithEmail(email *mail.Addr) *MembershipProfileFilter {
	f.email = email
	return f
}

func (f *MembershipProfileFilter) Email() *mail.Addr {
	return f.email
}

func (f *MembershipProfileFilter) WithState(state ProfileState) *MembershipProfileFilter {
	f.state = &state
	return f
}

func (f *MembershipProfileFilter) State() *ProfileState {
	return f.state
}

func (f *MembershipProfileFilter) WithSource(source ProfileSource) *MembershipProfileFilter {
	f.source = &source
	return f
}

func (f *MembershipProfileFilter) Source() *ProfileSource {
	return f.source
}

func (f *MembershipProfileFilter) SQLArguments() pgx.StrictNamedArgs {
	return pgx.StrictNamedArgs{
		"filter_email":           f.email,
		"exclude_contract_ended": f.excludeContractEnded,
		"current_date":           f.currentDate,
		"filter_state":           f.state,
		"filter_source":          f.source,
	}
}

func (f *MembershipProfileFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @filter_email::text IS NOT NULL THEN
			i.email_address = @filter_email::text
		ELSE TRUE
	END
)
AND (
    CASE
        WHEN @exclude_contract_ended::boolean IS NOT NULL AND @exclude_contract_ended::boolean = true THEN
            (p.contract_end_date IS NULL OR p.contract_end_date >= @current_date::date)
        ELSE TRUE
    END
)
AND (
	CASE
		WHEN @filter_state::text IS NOT NULL THEN
			p.state = @filter_state::membership_state
		ELSE TRUE
	END
)
AND (
	CASE
		WHEN @filter_source::text IS NOT NULL THEN
			p.source = @filter_source::text
		ELSE TRUE
	END
)
`
}
