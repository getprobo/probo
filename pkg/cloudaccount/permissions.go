// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package cloudaccount

import (
	"sort"

	"go.probo.inc/probo/pkg/coredata"
)

type (
	// scopeModuleKey indexes the GCP and Azure permission tables by
	// (scopeKind, module). GCP is scope-dependent: project scope
	// only needs roles/iam.securityReviewer; org scope additionally
	// needs roles/cloudasset.viewer.
	scopeModuleKey struct {
		Scope  coredata.CloudAccountScopeKind
		Module coredata.CloudAccountAuditModule
	}
)

// awsActionsByModule lists the IAM actions a Probo cross-account
// role must be granted for each audit module. v1 only ACCESS_REVIEW
// is populated; v2 adds CSPM_S3_PUBLIC and CSPM_IAM_AUDIT.
var awsActionsByModule = map[coredata.CloudAccountAuditModule][]string{
	coredata.CloudAccountAuditModuleAccessReview: {
		"iam:ListUsers",
		"iam:ListGroups",
		"iam:ListGroupsForUser",
		"iam:ListAttachedUserPolicies",
		"iam:ListAttachedGroupPolicies",
		"iam:ListUserPolicies",
		"iam:ListGroupPolicies",
		"iam:ListMFADevices",
		"iam:GetCredentialReport",
		"iam:GenerateCredentialReport",
		"iam:GetLoginProfile",
		"iam:GetUser",
		"sts:GetCallerIdentity",
	},
}

// gcpRolesByModule maps (scopeKind, module) to the predefined GCP
// IAM roles the install script must grant. Project scope omits
// cloudasset.viewer because cloudasset.searchAllResources requires
// org/folder scope and granting it at the project would mislead
// the customer about what enumeration paths v1 actually uses.
var gcpRolesByModule = map[scopeModuleKey][]string{
	{Scope: coredata.CloudAccountScopeKindGCPProject, Module: coredata.CloudAccountAuditModuleAccessReview}: {
		"roles/iam.securityReviewer",
	},
	{Scope: coredata.CloudAccountScopeKindGCPOrganization, Module: coredata.CloudAccountAuditModuleAccessReview}: {
		"roles/iam.securityReviewer",
		"roles/cloudasset.viewer",
	},
}

// gcpAPIsByModule lists the Google APIs the install script must
// enable for each (scope, module) combination. Project scope skips
// cloudasset.googleapis.com (no transitive enumeration on project
// scope -- see gcpRolesByModule rationale).
var gcpAPIsByModule = map[scopeModuleKey][]string{
	{Scope: coredata.CloudAccountScopeKindGCPProject, Module: coredata.CloudAccountAuditModuleAccessReview}: {
		"cloudresourcemanager.googleapis.com",
		"iam.googleapis.com",
	},
	{Scope: coredata.CloudAccountScopeKindGCPOrganization, Module: coredata.CloudAccountAuditModuleAccessReview}: {
		"cloudresourcemanager.googleapis.com",
		"iam.googleapis.com",
		"cloudasset.googleapis.com",
	},
}

// azureRBACRolesByModule lists the Azure RBAC roles a Probo App
// Registration must be granted at the chosen scope. v1 ACCESS_REVIEW
// asks for Reader at the customer's scope (Subscription / MG /
// Tenant). The customer assigns this; Probo never grants RBAC on
// its own.
var azureRBACRolesByModule = map[coredata.CloudAccountAuditModule][]string{
	coredata.CloudAccountAuditModuleAccessReview: {
		"Reader",
	},
}

// azureGraphPermissionsByModule lists the Microsoft Graph
// application permissions the App Registration must hold. v1
// requires Directory.Read.All -- an admin-consent permission, the
// install walkthrough must include an explicit "Grant admin consent"
// step before the user can paste credentials back.
var azureGraphPermissionsByModule = map[coredata.CloudAccountAuditModule][]string{
	coredata.CloudAccountAuditModuleAccessReview: {
		"Directory.Read.All",
	},
}

// AWSActionsForModules aggregates the AWS IAM actions required by
// the supplied audit modules. The result is de-duplicated and
// sorted for deterministic install-template output.
func AWSActionsForModules(modules []coredata.CloudAccountAuditModule) []string {
	return collectFlat(awsActionsByModule, modules)
}

// GCPRolesForModules aggregates the GCP roles required by the
// supplied (scope, modules) combination.
func GCPRolesForModules(scope coredata.CloudAccountScopeKind, modules []coredata.CloudAccountAuditModule) []string {
	return collectScoped(gcpRolesByModule, scope, modules)
}

// GCPAPIsForModules aggregates the Google APIs to enable for the
// supplied (scope, modules) combination.
func GCPAPIsForModules(scope coredata.CloudAccountScopeKind, modules []coredata.CloudAccountAuditModule) []string {
	return collectScoped(gcpAPIsByModule, scope, modules)
}

// AzureRBACRolesForModules aggregates the Azure RBAC roles required
// by the supplied audit modules.
func AzureRBACRolesForModules(modules []coredata.CloudAccountAuditModule) []string {
	return collectFlat(azureRBACRolesByModule, modules)
}

// AzureGraphPermissionsForModules aggregates the Microsoft Graph
// application permissions required by the supplied audit modules.
func AzureGraphPermissionsForModules(modules []coredata.CloudAccountAuditModule) []string {
	return collectFlat(azureGraphPermissionsByModule, modules)
}

// RequiredPermissions is the polymorphic aggregator the GraphQL
// install-assets resolver and the CLI both call to populate the
// "what the customer must grant" surface.
type RequiredPermissions struct {
	AWSActions            []string
	GCPRoles              []string
	GCPAPIs               []string
	AzureRBACRoles        []string
	AzureGraphPermissions []string
}

// RequiredPermissionsFor returns the permission set the customer
// must grant for the (provider, scope, modules) tuple. Fields not
// applicable to the provider remain nil so callers can render only
// the relevant subset.
func RequiredPermissionsFor(
	provider coredata.CloudAccountProvider,
	scope coredata.CloudAccountScopeKind,
	modules []coredata.CloudAccountAuditModule,
) RequiredPermissions {
	switch provider {
	case coredata.CloudAccountProviderAWS:
		return RequiredPermissions{AWSActions: AWSActionsForModules(modules)}
	case coredata.CloudAccountProviderGCP:
		return RequiredPermissions{
			GCPRoles: GCPRolesForModules(scope, modules),
			GCPAPIs:  GCPAPIsForModules(scope, modules),
		}
	case coredata.CloudAccountProviderAzure:
		return RequiredPermissions{
			AzureRBACRoles:        AzureRBACRolesForModules(modules),
			AzureGraphPermissions: AzureGraphPermissionsForModules(modules),
		}
	default:
		return RequiredPermissions{}
	}
}

func collectFlat(table map[coredata.CloudAccountAuditModule][]string, modules []coredata.CloudAccountAuditModule) []string {
	seen := make(map[string]struct{})
	for _, m := range modules {
		for _, v := range table[m] {
			seen[v] = struct{}{}
		}
	}

	return sortedKeys(seen)
}

func collectScoped(table map[scopeModuleKey][]string, scope coredata.CloudAccountScopeKind, modules []coredata.CloudAccountAuditModule) []string {
	seen := make(map[string]struct{})
	for _, m := range modules {
		for _, v := range table[scopeModuleKey{Scope: scope, Module: m}] {
			seen[v] = struct{}{}
		}
	}

	return sortedKeys(seen)
}

func sortedKeys(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)

	return out
}
