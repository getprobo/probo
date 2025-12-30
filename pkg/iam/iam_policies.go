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

import "go.probo.inc/probo/pkg/iam/policy"

// IAM Policies
//
// These policies define access control for the IAM service.
// They are applied in addition to role-based organization policies.

// IAMSelfManageIdentityPolicy allows users to manage their own identity.
// This is applied to all authenticated users regardless of organization membership.
var IAMSelfManageIdentityPolicy = policy.NewPolicy(
	"iam:self-manage-identity",
	"Self-Manage Identity",
	// Users can view and update their own identity
	policy.Allow(
		ActionIAMIdentityGet,
		ActionIAMIdentityUpdate,
		ActionIAMIdentityDelete,
	).WithSID("manage-own-identity").
		When(policy.Equals("principal.id", "resource.id")),

	// Users can list their own memberships, invitations, sessions, and API keys
	policy.Allow(
		ActionIAMIdentityListMemberships,
		ActionIAMIdentityListInvitations,
		ActionIAMIdentityListSessions,
		ActionIAMPersonalAPIKeyList,
	).WithSID("list-own-associations").
		When(policy.Equals("principal.id", "resource.id")),
).WithDescription("Allows users to manage their own identity, sessions, API keys, and view their memberships")

// IAMSelfManageSessionPolicy allows users to manage their own sessions.
var IAMSelfManageSessionPolicy = policy.NewPolicy(
	"iam:self-manage-session",
	"Self-Manage Sessions",
	// Users can view and revoke their own sessions
	policy.Allow(
		ActionIAMSessionGet,
		ActionIAMSessionRevoke,
		ActionIAMSessionRevokeAll,
	).WithSID("manage-own-sessions").
		When(policy.Equals("principal.id", "resource.user_id")),
).WithDescription("Allows users to view and revoke their own sessions")

// IAMSelfManageInvitationPolicy allows users to manage invitations sent to them.
var IAMSelfManageInvitationPolicy = policy.NewPolicy(
	"iam:self-manage-invitation",
	"Self-Manage Invitations",
	// Users can view and accept invitations sent to their email
	policy.Allow(
		ActionIAMInvitationGet,
		ActionIAMInvitationAccept,
	).WithSID("manage-own-invitations").
		When(policy.Equals("principal.id", "resource.user_id")),
).WithDescription("Allows users to view and accept invitations sent to them")

// IAMSelfManageMembershipPolicy allows users to view their own memberships.
var IAMSelfManageMembershipPolicy = policy.NewPolicy(
	"iam:self-manage-membership",
	"Self-Manage Memberships",
	// Users can view their own memberships
	policy.Allow(
		ActionIAMMembershipGet,
	).WithSID("view-own-memberships").
		When(policy.Equals("principal.id", "resource.user_id")),
).WithDescription("Allows users to view their organization memberships")

// IAMSelfManagePersonalAPIKeyPolicy allows users to manage their own API keys.
var IAMSelfManagePersonalAPIKeyPolicy = policy.NewPolicy(
	"iam:self-manage-personal-api-key",
	"Self-Manage Personal API Keys",
	// Users can create, view, update, and delete their own API keys
	policy.Allow(
		ActionIAMPersonalAPIKeyCreate,
		ActionIAMPersonalAPIKeyGet,
		ActionIAMPersonalAPIKeyUpdate,
		ActionIAMPersonalAPIKeyDelete,
	).WithSID("manage-own-api-keys").
		When(policy.Equals("principal.id", "resource.user_id")),
).WithDescription("Allows users to manage their own personal API keys")

// IAMOwnerPolicy defines permissions for organization owners.
var IAMOwnerPolicy = policy.NewPolicy(
	"iam:owner",
	"Organization Owner",
	// Full access to organization management
	policy.Allow("iam:organization:*").WithSID("full-org-access"),
	// Full access to member management
	policy.Allow("iam:membership:*").WithSID("full-membership-access"),
	// Can manage invitations
	policy.Allow(
		ActionIAMInvitationCreate,
		ActionIAMInvitationGet,
		ActionIAMInvitationDelete,
	).WithSID("manage-invitations"),
	// Full access to SAML configuration management
	policy.Allow("iam:saml-configuration:*").WithSID("full-saml-access"),
).WithDescription("Full IAM access for organization owners")

// IAMAdminPolicy defines permissions for organization admins.
var IAMAdminPolicy = policy.NewPolicy(
	"iam:admin",
	"Organization Admin",
	// Can view and update organization (but not delete)
	policy.Allow(
		ActionIAMOrganizationGet,
		ActionIAMOrganizationUpdate,
		ActionIAMOrganizationListMembers,
		ActionIAMOrganizationListInvitations,
		ActionIAMOrganizationInviteMember,
	).WithSID("org-admin-access"),
	// Can manage memberships (but not remove owner)
	policy.Allow(
		ActionIAMMembershipGet,
		ActionIAMMembershipUpdate,
	).WithSID("membership-admin-access"),
	// Can manage invitations
	policy.Allow(
		ActionIAMInvitationGet,
		ActionIAMInvitationDelete,
	).WithSID("invitation-admin-access"),
	// Can view and update SAML configurations
	policy.Allow(ActionIAMSAMLConfigurationGet).WithSID("saml-configuration-admin-access"),
	// Cannot delete organization
	policy.Deny(ActionIAMOrganizationDelete).WithSID("deny-org-delete"),
	// Cannot remove members (only owner can)
	policy.Deny(ActionIAMOrganizationRemoveMember).WithSID("deny-remove-member"),
	// Cannot manage SAML configurations (only owner can)
	policy.Deny(
		ActionIAMSAMLConfigurationCreate,
		ActionIAMSAMLConfigurationUpdate,
		ActionIAMSAMLConfigurationDelete,
	).WithSID("deny-saml-management"),
).WithDescription("IAM admin access - can manage members but cannot delete organization or manage SAML")

// IAMViewerPolicy defines permissions for organization viewers.
var IAMViewerPolicy = policy.NewPolicy(
	"iam:viewer",
	"Organization Viewer",
	// Read-only access to organization
	policy.Allow(
		ActionIAMOrganizationGet,
		ActionIAMOrganizationListMembers,
	).WithSID("org-viewer-access"),
	// Can view memberships
	policy.Allow(ActionIAMMembershipGet).WithSID("membership-viewer-access"),
).WithDescription("Read-only IAM access for organization viewers")
