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

package iam

type (
	Action string
)

const (
	ActionIAMOrganizationCreate          Action = "iam:organization:create"
	ActionIAMOrganizationUpdate          Action = "iam:organization:update"
	ActionIAMOrganizationGet             Action = "iam:organization:get"
	ActionIAMOrganizationDelete          Action = "iam:organization:delete"
	ActionIAMOrganizationList            Action = "iam:organization:list"
	ActionIAMOrganizationInviteMember    Action = "iam:organization:invite-member"
	ActionIAMOrganizationRemoveMember    Action = "iam:organization:remove-member"
	ActionIAMOrganizationListMembers     Action = "iam:organization:list-members"
	ActionIAMOrganizationListInvitations Action = "iam:organization:list-invitations"

	ActionIAMIdentityListMemberships Action = "iam:identity:list-memberships"
	ActionIAMIdentityListInvitations Action = "iam:identity:list-invitations"
	ActionIAMIdentityListSessions    Action = "iam:identity:list-sessions"

	ActionIAMSessionClose     Action = "iam:identity:close-session"
	ActionIAMSessionRevoke    Action = "iam:identity:revoke-session"
	ActionIAMSessionRevokeAll Action = "iam:identity:revoke-all-sessions"

	ActionIAMInvitationAccept Action = "iam:identity:accept-invitation"
)
