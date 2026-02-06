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
		ActionIdentityGet,
		ActionIdentityUpdate,
		ActionIdentityDelete,
	).
		WithSID("manage-own-identity").
		When(policy.Equals("principal.id", "resource.identity_id")),

	// Users can list their own memberships, invitations, sessions, and API keys
	policy.Allow(
		ActionMembershipList,
		ActionInvitationList,
		ActionSessionList,
		ActionPersonalAPIKeyList,
	).
		WithSID("list-own-associations").
		When(policy.Equals("principal.id", "resource.identity_id")),
).
	WithDescription("Allows users to manage their own identity, sessions, API keys, and view their memberships")

// IAMSelfManageSessionPolicy allows users to manage their own sessions.
var IAMSelfManageSessionPolicy = policy.NewPolicy(
	"iam:self-manage-session",
	"Self-Manage Sessions",

	// Users can view and revoke their own sessions
	policy.Allow(
		ActionSessionGet,
		ActionSessionRevoke,
		ActionSessionRevokeAll,
	).
		WithSID("manage-own-sessions").
		When(policy.Equals("principal.id", "resource.identity_id")),
).
	WithDescription("Allows users to view and revoke their own sessions")

// IAMSelfManageInvitationPolicy allows users to manage invitations sent to them.
var IAMSelfManageInvitationPolicy = policy.NewPolicy(
	"iam:self-manage-invitation",
	"Self-Manage Invitations",

	// Users can view and accept invitations sent to their email
	policy.Allow(
		ActionInvitationGet,
		ActionInvitationAccept,
	).
		WithSID("manage-own-invitations").
		When(policy.Equals("principal.email", "resource.email")),

	// Users can view organization details for organizations they have pending invitations to
	policy.Allow(ActionOrganizationGet).
		WithSID("view-org-for-pending-invitation").
		When(policy.In("resource.organization_id", "principal.pending_invitation_organization_ids")),
).
	WithDescription("Allows users to view and accept invitations sent to them")

// IAMSelfManageProfilePolicy allows users to view their own profiles.
var IAMSelfManageProfilePolicy = policy.NewPolicy(
	"iam:self-manage-profile",
	"Self-Manage Profiles",

	// Users can view their own profiles
	policy.Allow(ActionMembershipProfileGet).
		WithSID("view-own-profiles").
		When(policy.Equals("principal.id", "resource.identity_id")),
).
	WithDescription("Allows users to view their organization profiles")

// IAMSelfManageMembershipPolicy allows users to view their own memberships.
var IAMSelfManageMembershipPolicy = policy.NewPolicy(
	"iam:self-manage-membership",
	"Self-Manage Memberships",

	// Users can view their own memberships
	policy.Allow(ActionMembershipGet).
		WithSID("view-own-memberships").
		When(policy.Equals("principal.id", "resource.identity_id")),
).
	WithDescription("Allows users to view their organization memberships")

// IAMSelfManagePersonalAPIKeyPolicy allows users to manage their own API keys.
var IAMSelfManagePersonalAPIKeyPolicy = policy.NewPolicy(
	"iam:self-manage-personal-api-key",
	"Self-Manage Personal API Keys",

	// Users can create, view, update, and delete their own API keys
	policy.Allow(
		ActionPersonalAPIKeyCreate,
		ActionPersonalAPIKeyGet,
		ActionPersonalAPIKeyUpdate,
		ActionPersonalAPIKeyDelete,
	).
		WithSID("manage-own-api-keys").
		When(policy.Equals("principal.id", "resource.identity_id")),
).
	WithDescription("Allows users to manage their own personal API keys")

// IAMOwnerPolicy defines permissions for organization owners.
var IAMOwnerPolicy = policy.NewPolicy(
	"iam:owner",
	"Organization Owner",

	// Full access to organization management
	policy.Allow("iam:organization:*").
		WithSID("full-org-access").
		When(policy.Equals("principal.organization_id", "resource.organization_id")),

	// Full access to member management (scoped to own organization), except deletion on SCIM sourced memberships
	policy.Allow(
		ActionMembershipGet,
		ActionMembershipList,
		ActionMembershipUpdate,
	).
		WithSID("membership-owner-access").
		When(policy.Equals("principal.organization_id", "resource.organization_id")),
	policy.Allow(
		ActionMembershipDelete,
	).
		WithSID("membership-deletion-owner-access").
		When(
			policy.Equals("principal.organization_id", "resource.organization_id"),
			policy.NotEquals("resource.source", "SCIM"),
		),

	// Can set other members OWNER
	policy.Allow(ActionMembershipRoleSetOwner).
		WithSID("membership-role-owner-access").
		When(policy.Equals("principal.organization_id", "resource.organization_id")),

	// Full access to membership profiles (scoped to own organization)
	policy.Allow(
		ActionMembershipProfileGet,
		ActionMembershipProfileList,
		ActionMembershipProfileUpdate,
	).
		WithSID("full-membership-profile-access").
		When(policy.Equals("principal.organization_id", "resource.organization_id")),

	// Can view identities of members in the same organization
	policy.Allow(ActionIdentityGet).
		WithSID("view-member-identity").
		When(policy.Equals("principal.organization_id", "resource.organization_id")),

	// Can manage invitations (scoped to own organization)
	policy.Allow(
		ActionInvitationList,
		ActionInvitationCreate,
		ActionInvitationGet,
		ActionInvitationList,
		ActionInvitationDelete,
	).
		WithSID("manage-invitations").
		When(policy.Equals("principal.organization_id", "resource.organization_id")),

	// Full access to SAML configuration management (scoped to own organization)
	policy.Allow("iam:saml-configuration:*").
		WithSID("full-saml-access").
		When(policy.Equals("principal.organization_id", "resource.organization_id")),

	// Full access to SCIM configuration management (scoped to own organization)
	policy.Allow("iam:scim-configuration:*").
		WithSID("full-scim-configuration-access").
		When(policy.Equals("principal.organization_id", "resource.organization_id")),

	// Full access to SCIM event viewing (scoped to own organization)
	policy.Allow("iam:scim-event:*").
		WithSID("full-scim-event-access").
		When(policy.Equals("principal.organization_id", "resource.organization_id")),

	// Allow updating SCIM bridge settings (scoped to own organization)
	policy.Allow(ActionSCIMBridgeUpdate).
		WithSID("scim-bridge-update-access").
		When(policy.Equals("principal.organization_id", "resource.organization_id")),
).
	WithDescription("Full IAM access for organization owners")

// IAMAdminPolicy defines permissions for organization admins.
var IAMAdminPolicy = policy.NewPolicy(
	"iam:admin",
	"Organization Admin",

	// Can view and update organization (but not delete)
	policy.Allow(
		ActionOrganizationGet,
		ActionOrganizationUpdate,
		ActionMembershipList,
		ActionInvitationList,
		ActionInvitationCreate,
	).
		WithSID("org-admin-access").
		When(policy.Equals("principal.organization_id", "resource.organization_id")),

	// Can manage memberships (scoped to own organization)
	policy.Allow(
		ActionMembershipGet,
	).
		WithSID("membership-admin-access").
		When(policy.Equals("principal.organization_id", "resource.organization_id")),

	policy.Allow(
		ActionMembershipUpdate,
	).
		WithSID("membership-role-admin-access").
		When(
			policy.Equals("principal.organization_id", "resource.organization_id"),
			policy.NotEquals("resource.role", "OWNER"),
		),

	// Can view membership profiles (scoped to own organization)
	policy.Allow(
		ActionMembershipProfileGet,
		ActionMembershipProfileList,
		ActionMembershipProfileUpdate,
	).
		WithSID("membership-profile-admin-access").
		When(policy.Equals("principal.organization_id", "resource.organization_id")),

	// Can view identities of members in the same organization
	policy.Allow(ActionIdentityGet).
		WithSID("view-member-identity").
		When(policy.Equals("principal.organization_id", "resource.organization_id")),

	// Can manage invitations (scoped to own organization)
	policy.Allow(
		ActionInvitationGet,
		ActionInvitationDelete,
	).
		WithSID("invitation-admin-access").
		When(policy.Equals("principal.organization_id", "resource.organization_id")),

	// Can view SAML configurations (scoped to own organization)
	policy.Allow(ActionSAMLConfigurationGet).
		WithSID("saml-configuration-admin-access").
		When(policy.Equals("principal.organization_id", "resource.organization_id")),

	// Cannot delete organization
	policy.Deny(ActionOrganizationDelete).
		WithSID("deny-org-delete"),

	// Cannot remove members (only owner can)
	policy.Deny(ActionMembershipDelete).
		WithSID("deny-remove-member"),

	// Cannot manage SAML configurations (only owner can)
	policy.Deny(
		ActionSAMLConfigurationCreate,
		ActionSAMLConfigurationUpdate,
		ActionSAMLConfigurationDelete,
	).
		WithSID("deny-saml-management"),

	// Can view SCIM configuration and events (scoped to own organization)
	policy.Allow(
		ActionSCIMConfigurationGet,
		ActionSCIMEventList,
		ActionSCIMEventGet,
	).
		WithSID("scim-admin-view-access").
		When(policy.Equals("principal.organization_id", "resource.organization_id")),

	// Cannot manage SCIM configurations (only owner can)
	policy.Deny(
		ActionSCIMConfigurationCreate,
		ActionSCIMConfigurationUpdate,
		ActionSCIMConfigurationDelete,
	).
		WithSID("deny-scim-management"),
).
	WithDescription("IAM admin access - can manage members but cannot delete organization or manage SAML/SCIM")

// IAMViewerPolicy defines permissions for organization viewers.
var IAMViewerPolicy = policy.NewPolicy(
	"iam:viewer",
	"Organization Viewer",

	// Read-only access to organization
	policy.Allow(
		ActionOrganizationGet,
		ActionMembershipList,
	).
		WithSID("org-viewer-access").
		When(policy.Equals("principal.organization_id", "resource.organization_id")),

	// Can view memberships (scoped to own organization)
	policy.Allow(ActionMembershipGet).
		WithSID("membership-viewer-access").
		When(policy.Equals("principal.organization_id", "resource.organization_id")),

	// Can view membership profiles (scoped to own organization)
	policy.Allow(
		ActionMembershipProfileGet,
		ActionMembershipProfileList,
	).
		WithSID("membership-profile-viewer-access").
		When(policy.Equals("principal.organization_id", "resource.organization_id")),

	// Can view identities of members in the same organization
	policy.Allow(ActionIdentityGet).
		WithSID("view-member-identity").
		When(policy.Equals("principal.organization_id", "resource.organization_id")),
).
	WithDescription("Read-only IAM access for organization viewers")
