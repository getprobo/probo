// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import "go.probo.inc/probo/pkg/coredata"

const (
	ScopeV1IAMRead coredata.OAuth2Scope = "v1:iam:read"
	ScopeV1IAM     coredata.OAuth2Scope = "v1:iam"
)

// IAMOAuth2ScopeSet returns OAuth2 scope mappings for IAM actions.
func IAMOAuth2ScopeSet() *ScopeSet {
	return CreateScopeSet(
		map[coredata.OAuth2Scope][]Action{
			ScopeV1IAMRead: {
				ActionOrganizationGet,
				ActionOrganizationList,
				ActionIdentityGet,
				ActionSessionList,
				ActionSessionGet,
				ActionInvitationList,
				ActionInvitationGet,
				ActionMembershipGet,
				ActionMembershipList,
				ActionMembershipProfileGet,
				ActionMembershipProfileList,
				ActionPersonalAPIKeyGet,
				ActionPersonalAPIKeyList,
				ActionSAMLConfigurationGet,
				ActionSAMLConfigurationList,
				ActionSCIMConfigurationGet,
				ActionSCIMEventList,
				ActionSCIMEventGet,
				ActionSCIMBridgeGet,
				ActionOAuth2ConsentGet,
				ActionAuditLogEntryGet,
				ActionAuditLogEntryList,
			},
			ScopeV1IAM: {
				ActionOrganizationCreate,
				ActionOrganizationUpdate,
				ActionOrganizationDelete,
				ActionIdentityUpdate,
				ActionIdentityDelete,
				ActionSessionRevoke,
				ActionSessionRevokeAll,
				ActionInvitationCreate,
				ActionInvitationAccept,
				ActionInvitationDelete,
				ActionMembershipUpdate,
				ActionMembershipDelete,
				ActionMembershipRoleSetOwner,
				ActionMembershipProfileCreate,
				ActionMembershipProfileUpdate,
				ActionMembershipProfileDelete,
				ActionMembershipProfileActivate,
				ActionMembershipProfileDeactivate,
				ActionPersonalAPIKeyCreate,
				ActionPersonalAPIKeyUpdate,
				ActionPersonalAPIKeyDelete,
				ActionSAMLConfigurationCreate,
				ActionSAMLConfigurationUpdate,
				ActionSAMLConfigurationDelete,
				ActionSCIMConfigurationCreate,
				ActionSCIMConfigurationUpdate,
				ActionSCIMConfigurationDelete,
				ActionSCIMBridgeCreate,
				ActionSCIMBridgeUpdate,
				ActionSCIMBridgeDelete,
				ActionOAuth2ConsentApprove,
			},
		},
	)
}
