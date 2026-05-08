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

package cloudaccount_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.probo.inc/probo/pkg/cloudaccount"
	"go.probo.inc/probo/pkg/coredata"
)

func TestAWSActionsForModules_AccessReview(t *testing.T) {
	t.Parallel()

	got := cloudaccount.AWSActionsForModules([]coredata.CloudAccountAuditModule{
		coredata.CloudAccountAuditModuleAccessReview,
	})

	assert.Contains(t, got, "iam:ListUsers")
	assert.Contains(t, got, "iam:ListMFADevices")
	assert.Contains(t, got, "sts:GetCallerIdentity")
	assert.Contains(t, got, "iam:GenerateCredentialReport")
}

func TestGCPRolesForModules_ProjectVsOrgScope(t *testing.T) {
	t.Parallel()

	t.Run("project scope grants only securityReviewer", func(t *testing.T) {
		t.Parallel()

		got := cloudaccount.GCPRolesForModules(
			coredata.CloudAccountScopeKindGCPProject,
			[]coredata.CloudAccountAuditModule{coredata.CloudAccountAuditModuleAccessReview},
		)
		assert.Equal(t, []string{"roles/iam.securityReviewer"}, got)
	})

	t.Run("org scope additionally grants cloudasset.viewer", func(t *testing.T) {
		t.Parallel()

		got := cloudaccount.GCPRolesForModules(
			coredata.CloudAccountScopeKindGCPOrganization,
			[]coredata.CloudAccountAuditModule{coredata.CloudAccountAuditModuleAccessReview},
		)
		assert.Contains(t, got, "roles/cloudasset.viewer")
		assert.Contains(t, got, "roles/iam.securityReviewer")
	})
}

func TestGCPAPIsForModules_ProjectScopeOmitsCloudAsset(t *testing.T) {
	t.Parallel()

	got := cloudaccount.GCPAPIsForModules(
		coredata.CloudAccountScopeKindGCPProject,
		[]coredata.CloudAccountAuditModule{coredata.CloudAccountAuditModuleAccessReview},
	)
	assert.NotContains(t, got, "cloudasset.googleapis.com")
}

func TestGCPAPIsForModules_OrgScopeIncludesCloudAsset(t *testing.T) {
	t.Parallel()

	got := cloudaccount.GCPAPIsForModules(
		coredata.CloudAccountScopeKindGCPOrganization,
		[]coredata.CloudAccountAuditModule{coredata.CloudAccountAuditModuleAccessReview},
	)
	assert.Contains(t, got, "cloudasset.googleapis.com")
}

func TestAzureGraphPermissions_RequiresDirectoryRead(t *testing.T) {
	t.Parallel()

	got := cloudaccount.AzureGraphPermissionsForModules(
		[]coredata.CloudAccountAuditModule{coredata.CloudAccountAuditModuleAccessReview},
	)
	assert.Contains(t, got, "Directory.Read.All")
}

func TestRequiredPermissionsFor_PerProvider(t *testing.T) {
	t.Parallel()

	t.Run("aws populates only AWSActions", func(t *testing.T) {
		t.Parallel()

		req := cloudaccount.RequiredPermissionsFor(
			coredata.CloudAccountProviderAWS,
			coredata.CloudAccountScopeKindAWSAccount,
			[]coredata.CloudAccountAuditModule{coredata.CloudAccountAuditModuleAccessReview},
		)
		assert.NotEmpty(t, req.AWSActions)
		assert.Empty(t, req.GCPRoles)
		assert.Empty(t, req.AzureRBACRoles)
	})

	t.Run("gcp populates only GCP fields", func(t *testing.T) {
		t.Parallel()

		req := cloudaccount.RequiredPermissionsFor(
			coredata.CloudAccountProviderGCP,
			coredata.CloudAccountScopeKindGCPOrganization,
			[]coredata.CloudAccountAuditModule{coredata.CloudAccountAuditModuleAccessReview},
		)
		assert.Empty(t, req.AWSActions)
		assert.NotEmpty(t, req.GCPRoles)
		assert.NotEmpty(t, req.GCPAPIs)
		assert.Empty(t, req.AzureRBACRoles)
	})

	t.Run("azure populates only Azure fields", func(t *testing.T) {
		t.Parallel()

		req := cloudaccount.RequiredPermissionsFor(
			coredata.CloudAccountProviderAzure,
			coredata.CloudAccountScopeKindAzureManagementGroup,
			[]coredata.CloudAccountAuditModule{coredata.CloudAccountAuditModuleAccessReview},
		)
		assert.Empty(t, req.AWSActions)
		assert.Empty(t, req.GCPRoles)
		assert.NotEmpty(t, req.AzureRBACRoles)
		assert.NotEmpty(t, req.AzureGraphPermissions)
	})
}
