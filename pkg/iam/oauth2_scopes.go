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

package iam

import "go.probo.inc/probo/pkg/coredata"

const (
	ScopeV1IAMRead coredata.OAuth2Scope = "v1:iam:read"
	ScopeV1IAM     coredata.OAuth2Scope = "v1:iam"
)

var IAMOAuth2ScopeMappings = map[coredata.OAuth2Scope][]string{
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
		ActionOAuth2AccessTokenGet,
		ActionOAuth2AccessTokenList,
	},
	ScopeV1IAM: {
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
		ActionOAuth2AccessTokenGet,
		ActionOAuth2AccessTokenList,
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
}
