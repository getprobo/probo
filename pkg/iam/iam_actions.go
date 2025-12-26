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

type Action = string

const (
	// Organization actions
	ActionIAMOrganizationCreate          = "iam:organization:create"
	ActionIAMOrganizationGet             = "iam:organization:get"
	ActionIAMOrganizationUpdate          = "iam:organization:update"
	ActionIAMOrganizationDelete          = "iam:organization:delete"
	ActionIAMOrganizationList            = "iam:organization:list"
	ActionIAMOrganizationInviteMember    = "iam:organization:invite-member"
	ActionIAMOrganizationRemoveMember    = "iam:organization:remove-member"
	ActionIAMOrganizationListMembers     = "iam:organization:list-members"
	ActionIAMOrganizationListInvitations = "iam:organization:list-invitations"

	// Identity actions
	ActionIAMIdentityGet                 = "iam:identity:get"
	ActionIAMIdentityUpdate              = "iam:identity:update"
	ActionIAMIdentityDelete              = "iam:identity:delete"
	ActionIAMIdentityListMemberships     = "iam:identity:list-memberships"
	ActionIAMIdentityListInvitations     = "iam:identity:list-invitations"
	ActionIAMIdentityListSessions        = "iam:identity:list-sessions"
	ActionIAMIdentityListPersonalAPIKeys = "iam:identity:list-personal-api-keys"

	// Session actions
	ActionIAMSessionGet       = "iam:session:get"
	ActionIAMSessionRevoke    = "iam:session:revoke"
	ActionIAMSessionRevokeAll = "iam:session:revoke-all"

	// Invitation actions
	ActionIAMInvitationGet    = "iam:invitation:get"
	ActionIAMInvitationAccept = "iam:invitation:accept"
	ActionIAMInvitationDelete = "iam:invitation:delete"

	// Membership actions
	ActionIAMMembershipGet    = "iam:membership:get"
	ActionIAMMembershipLit    = "iam:membership:list"
	ActionIAMMembershipUpdate = "iam:membership:update"
	ActionIAMMembershipDelete = "iam:membership:delete"

	// Personal API Key actions
	ActionIAMPersonalAPIKeyCreate = "iam:personal-api-key:create"
	ActionIAMPersonalAPIKeyGet    = "iam:personal-api-key:get"
	ActionIAMPersonalAPIKeyUpdate = "iam:personal-api-key:update"
	ActionIAMPersonalAPIKeyDelete = "iam:personal-api-key:delete"

	// SAML Configuration actions
	ActionIAMSAMLConfigurationCreate = "iam:saml-configuration:create"
	ActionIAMSAMLConfigurationGet    = "iam:saml-configuration:get"
	ActionIAMSAMLConfigurationUpdate = "iam:saml-configuration:update"
	ActionIAMSAMLConfigurationDelete = "iam:saml-configuration:delete"
	ActionIAMSAMLConfigurationList   = "iam:saml-configuration:list"
)
