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
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"go.probo.inc/probo/pkg/mail"
)

type (
	MembershipProfileFilter struct {
		withMembership             *bool
		withCompliancePortalAccess *bool
		contractEnded              *bool
		currentDate                time.Time
		email                      *mail.Addr
		userName                   *string
		externalID                 *string
		states                     ProfileStateValues
		source                     *ProfileSource
		query                      *string
		role                       *MembershipRole
		kind                       *string
	}
)

func NewMembershipProfileFilter(contractEnded *bool) *MembershipProfileFilter {
	return &MembershipProfileFilter{
		contractEnded: contractEnded,
		currentDate:   time.Now(),
	}
}

func (f *MembershipProfileFilter) WithMembership() *MembershipProfileFilter {
	f.withMembership = new(true)
	return f
}

func (f *MembershipProfileFilter) WithCompliancePortalAccess() *MembershipProfileFilter {
	f.withCompliancePortalAccess = new(true)
	return f
}

func (f *MembershipProfileFilter) WithEmail(email *mail.Addr) *MembershipProfileFilter {
	f.email = email
	return f
}

func (f *MembershipProfileFilter) Email() *mail.Addr {
	return f.email
}

func (f *MembershipProfileFilter) WithUserName(userName string) *MembershipProfileFilter {
	f.userName = &userName
	return f
}

func (f *MembershipProfileFilter) WithExternalID(externalID string) *MembershipProfileFilter {
	f.externalID = &externalID
	return f
}

func (f *MembershipProfileFilter) WithState(state ProfileState) *MembershipProfileFilter {
	f.states = ProfileStateValues{state}
	return f
}

func (f *MembershipProfileFilter) WithStates(states ...ProfileState) *MembershipProfileFilter {
	f.states = ProfileStateValues(states)
	return f
}

func (f *MembershipProfileFilter) States() []ProfileState {
	return f.states
}

func (f *MembershipProfileFilter) WithSource(source ProfileSource) *MembershipProfileFilter {
	f.source = &source
	return f
}

func (f *MembershipProfileFilter) Source() *ProfileSource {
	return f.source
}

func (f *MembershipProfileFilter) WithQuery(query *string) *MembershipProfileFilter {
	f.query = query
	return f
}

func (f *MembershipProfileFilter) Query() *string {
	return f.query
}

func (f *MembershipProfileFilter) WithRole(role MembershipRole) *MembershipProfileFilter {
	f.role = &role
	return f
}

func (f *MembershipProfileFilter) Role() *MembershipRole {
	return f.role
}

func (f *MembershipProfileFilter) WithKind(kind *string) *MembershipProfileFilter {
	f.kind = kind
	return f
}

func (f *MembershipProfileFilter) Kind() *string {
	return f.kind
}

func escapeLikePattern(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func (f *MembershipProfileFilter) SQLArguments() pgx.StrictNamedArgs {
	var filterQuery *string

	if f.query != nil && *f.query != "" {
		escaped := escapeLikePattern(*f.query)
		filterQuery = &escaped
	}

	return pgx.StrictNamedArgs{
		"filter_email":             f.email,
		"filter_user_name":         f.userName,
		"filter_external_id":       f.externalID,
		"with_membership":          f.withMembership,
		"with_trust_center_access": f.withCompliancePortalAccess,
		"contract_ended":           f.contractEnded,
		"current_date":             f.currentDate,
		"filter_states":            f.states,
		"filter_source":            f.source,
		"filter_query":             filterQuery,
		"filter_role":              f.role,
		"filter_kind":              f.kind,
	}
}

func (f *MembershipProfileFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @with_trust_center_access::boolean IS NOT NULL AND @with_trust_center_access::boolean = TRUE THEN
			EXISTS (SELECT 1 FROM trust_center_accesses WHERE identity_id = p.identity_id AND organization_id = p.organization_id)
		WHEN @with_trust_center_access::boolean IS NOT NULL AND @with_trust_center_access::boolean = FALSE THEN
			NOT EXISTS (SELECT 1 FROM trust_center_accesses WHERE identity_id = p.identity_id AND organization_id = p.organization_id)
		ELSE TRUE
	END
)
AND (
	CASE
		WHEN @with_membership::boolean IS NOT NULL AND @with_membership::boolean = TRUE THEN
			EXISTS (SELECT 1 FROM iam_memberships WHERE identity_id = p.identity_id AND organization_id = p.organization_id)
		WHEN @with_membership::boolean IS NOT NULL AND @with_membership::boolean = FALSE THEN
			NOT EXISTS (SELECT 1 FROM iam_memberships WHERE identity_id = p.identity_id AND organization_id = p.organization_id)
		ELSE TRUE
	END
)
AND (
	CASE
		WHEN @filter_email::citext IS NOT NULL THEN
			i.email_address = @filter_email::citext
		ELSE TRUE
	END
)
AND (
    CASE
        WHEN @contract_ended::boolean IS NOT NULL AND @contract_ended::boolean = true THEN
            (p.contract_end_date IS NOT NULL AND p.contract_end_date < @current_date::date)
        WHEN @contract_ended::boolean IS NOT NULL AND @contract_ended::boolean = false THEN
            (p.contract_end_date IS NULL OR p.contract_end_date >= @current_date::date)
        ELSE TRUE
    END
)
AND (
	CASE
		WHEN @filter_states::membership_state[] IS NOT NULL THEN
			p.state = ANY(@filter_states::membership_state[])
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
AND (
	CASE
		WHEN @filter_user_name::citext IS NOT NULL THEN
			p.user_name = @filter_user_name::citext
		ELSE TRUE
	END
)
AND (
	CASE
		WHEN @filter_external_id::text IS NOT NULL THEN
			p.external_id = @filter_external_id::text
		ELSE TRUE
	END
)
AND (
	CASE
		WHEN @filter_query::text IS NOT NULL AND @filter_query::text <> '' THEN
			(
				p.full_name ILIKE '%' || @filter_query || '%' ESCAPE '\'
				OR i.email_address ILIKE '%' || @filter_query || '%' ESCAPE '\'
				OR p.position ILIKE '%' || @filter_query || '%' ESCAPE '\'
			)
		ELSE TRUE
	END
)
AND (
	CASE
		WHEN @filter_role::authz_role IS NOT NULL THEN
			p.identity_id IN (
				SELECT identity_id
				FROM iam_memberships
				WHERE
					organization_id = p.organization_id
					AND role = @filter_role::authz_role
			)
		ELSE TRUE
	END
)
AND (
	CASE
		WHEN @filter_kind::text IS NOT NULL AND @filter_kind::text <> '' THEN
			p.kind = @filter_kind::text
		ELSE TRUE
	END
)
`
}
