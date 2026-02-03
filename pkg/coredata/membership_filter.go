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
	"go.probo.inc/probo/pkg/mail"
)

type MembershipFilter struct {
	email  *mail.Addr
	role   *MembershipRole
	state  *MembershipState
	source *MembershipSource
}

func NewMembershipFilter() *MembershipFilter {
	return &MembershipFilter{}
}

func (f *MembershipFilter) WithEmail(email *mail.Addr) *MembershipFilter {
	f.email = email
	return f
}

func (f *MembershipFilter) Email() *mail.Addr {
	return f.email
}

func (f *MembershipFilter) WithRole(role MembershipRole) *MembershipFilter {
	f.role = &role
	return f
}

func (f *MembershipFilter) Role() *MembershipRole {
	return f.role
}

func (f *MembershipFilter) WithState(state MembershipState) *MembershipFilter {
	f.state = &state
	return f
}

func (f *MembershipFilter) State() *MembershipState {
	return f.state
}

func (f *MembershipFilter) WithSource(source MembershipSource) *MembershipFilter {
	f.source = &source
	return f
}

func (f *MembershipFilter) Source() *MembershipSource {
	return f.source
}

func (f *MembershipFilter) SQLArguments() pgx.StrictNamedArgs {
	return pgx.StrictNamedArgs{
		"filter_email":  f.email,
		"filter_role":   f.role,
		"filter_state":  f.state,
		"filter_source": f.source,
	}
}

func (f *MembershipFilter) SQLFragment() string {
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
		WHEN @filter_role::text IS NOT NULL THEN
			m.role = @filter_role::authz_role
		ELSE TRUE
	END
)
AND (
	CASE
		WHEN @filter_state::text IS NOT NULL THEN
			m.state = @filter_state::membership_state
		ELSE TRUE
	END
)
AND (
	CASE
		WHEN @filter_source::text IS NOT NULL THEN
			m.source = @filter_source::text
		ELSE TRUE
	END
)`
}
