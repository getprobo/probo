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
	ActionOrganizationCreate = "iam:organization:create"
	ActionOrganizationGet    = "iam:organization:get"
	ActionOrganizationUpdate = "iam:organization:update"
	ActionOrganizationDelete = "iam:organization:delete"
	ActionOrganizationList   = "iam:organization:list"

	// Identity actions
	ActionIdentityGet    = "iam:identity:get"
	ActionIdentityUpdate = "iam:identity:update"
	ActionIdentityDelete = "iam:identity:delete"

	// Session actions
	ActionSessionList      = "iam:session:list"
	ActionSessionGet       = "iam:session:get"
	ActionSessionRevoke    = "iam:session:revoke"
	ActionSessionRevokeAll = "iam:session:revoke-all"

	// Invitation actions
	ActionInvitationList   = "iam:invitation:list"
	ActionInvitationCreate = "iam:invitation:create"
	ActionInvitationGet    = "iam:invitation:get"
	ActionInvitationAccept = "iam:invitation:accept"
	ActionInvitationDelete = "iam:invitation:delete"

	// Membership actions
	ActionMembershipGet    = "iam:membership:get"
	ActionMembershipList   = "iam:membership:list"
	ActionMembershipUpdate = "iam:membership:update"
	ActionMembershipDelete = "iam:membership:delete"

	// Membership role actions
	ActionMembershipRoleSetOwner = "iam:membership-role:set-owner"

	// Membership Profile actions
	ActionMembershipProfileGet    = "iam:membership-profile:get"
	ActionMembershipProfileList   = "iam:membership-profile:list"
	ActionMembershipProfileUpdate = "iam:membership-profile:update"

	// Personal API Key actions
	ActionPersonalAPIKeyCreate = "iam:personal-api-key:create"
	ActionPersonalAPIKeyGet    = "iam:personal-api-key:get"
	ActionPersonalAPIKeyList   = "iam:personal-api-key:list"
	ActionPersonalAPIKeyUpdate = "iam:personal-api-key:update"
	ActionPersonalAPIKeyDelete = "iam:personal-api-key:delete"

	// SAML Configuration actions
	ActionSAMLConfigurationCreate = "iam:saml-configuration:create"
	ActionSAMLConfigurationGet    = "iam:saml-configuration:get"
	ActionSAMLConfigurationUpdate = "iam:saml-configuration:update"
	ActionSAMLConfigurationDelete = "iam:saml-configuration:delete"
	ActionSAMLConfigurationList   = "iam:saml-configuration:list"

	// SCIM Configuration actions
	ActionSCIMConfigurationCreate = "iam:scim-configuration:create"
	ActionSCIMConfigurationGet    = "iam:scim-configuration:get"
	ActionSCIMConfigurationUpdate = "iam:scim-configuration:update"
	ActionSCIMConfigurationDelete = "iam:scim-configuration:delete"

	// SCIM Event actions
	ActionSCIMEventList = "iam:scim-event:list"
	ActionSCIMEventGet  = "iam:scim-event:get"

	// SCIM Bridge actions
	ActionSCIMBridgeGet    = "iam:scim-bridge:get"
	ActionSCIMBridgeCreate = "iam:scim-bridge:create"
	ActionSCIMBridgeUpdate = "iam:scim-bridge:update"
	ActionSCIMBridgeDelete = "iam:scim-bridge:delete"

	// Connector actions
	ActionConnectorGet = "iam:connector:get"
)
