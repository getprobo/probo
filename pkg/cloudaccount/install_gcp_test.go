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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/cloudaccount"
	"go.probo.inc/probo/pkg/coredata"
)

func TestBuildGCPInstallScript_ProjectScope(t *testing.T) {
	t.Parallel()

	got, err := cloudaccount.BuildGCPInstallScript(
		[]coredata.CloudAccountAuditModule{coredata.CloudAccountAuditModuleAccessReview},
		coredata.CloudAccountScopeKindGCPProject,
		"probo-target-project",
	)
	require.NoError(t, err)

	t.Run("required roles include only securityReviewer", func(t *testing.T) {
		t.Parallel()
		assert.Contains(t, got.RequiredRoles, "roles/iam.securityReviewer")
		assert.NotContains(t, got.RequiredRoles, "roles/cloudasset.viewer", "project scope must NOT include cloudasset.viewer")
	})

	t.Run("required APIs include cloudresourcemanager + iam, NOT cloudasset", func(t *testing.T) {
		t.Parallel()
		assert.Contains(t, got.RequiredAPIs, "cloudresourcemanager.googleapis.com")
		assert.Contains(t, got.RequiredAPIs, "iam.googleapis.com")
		assert.NotContains(t, got.RequiredAPIs, "cloudasset.googleapis.com", "project scope must NOT enable cloudasset.googleapis.com")
	})

	t.Run("script enables the project-scope APIs and grants only securityReviewer", func(t *testing.T) {
		t.Parallel()
		assert.Contains(t, got.SetupScript, "gcloud services enable cloudresourcemanager.googleapis.com")
		assert.Contains(t, got.SetupScript, "gcloud services enable iam.googleapis.com")
		assert.NotContains(t, got.SetupScript, "cloudasset.googleapis.com")
		assert.Contains(t, got.SetupScript, "roles/iam.securityReviewer")
		assert.NotContains(t, got.SetupScript, "roles/cloudasset.viewer")
	})

	t.Run("script contains the probo-scanner- project prefix and embeds the scope identifier", func(t *testing.T) {
		t.Parallel()
		assert.True(t,
			strings.HasPrefix(got.ScannerProjectID, "probo-scanner-"),
			"scanner project id %q must carry the probo-scanner- prefix",
			got.ScannerProjectID,
		)
		assert.Contains(t, got.SetupScript, "SCANNER_PROJECT_ID=\""+got.ScannerProjectID+"\"")
		assert.Contains(t, got.SetupScript, "gcloud projects create")
		assert.Contains(t, got.SetupScript, "probo-target-project", "scope identifier must be referenced in the script")
	})

	t.Run("script tail prints SA email + JSON key path", func(t *testing.T) {
		t.Parallel()
		assert.Contains(t, got.SetupScript, "Service account email:")
		assert.Contains(t, got.SetupScript, "JSON key file:")
		assert.Contains(t, got.SetupScript, "iam service-accounts keys create")
	})
}

func TestBuildGCPInstallScript_OrganizationScope(t *testing.T) {
	t.Parallel()

	got, err := cloudaccount.BuildGCPInstallScript(
		[]coredata.CloudAccountAuditModule{coredata.CloudAccountAuditModuleAccessReview},
		coredata.CloudAccountScopeKindGCPOrganization,
		"123456789",
	)
	require.NoError(t, err)

	t.Run("required roles include both securityReviewer and cloudasset.viewer", func(t *testing.T) {
		t.Parallel()
		assert.Contains(t, got.RequiredRoles, "roles/iam.securityReviewer")
		assert.Contains(t, got.RequiredRoles, "roles/cloudasset.viewer")
	})

	t.Run("required APIs additionally enable cloudasset.googleapis.com", func(t *testing.T) {
		t.Parallel()
		assert.Contains(t, got.RequiredAPIs, "cloudresourcemanager.googleapis.com")
		assert.Contains(t, got.RequiredAPIs, "iam.googleapis.com")
		assert.Contains(t, got.RequiredAPIs, "cloudasset.googleapis.com")
	})

	t.Run("script binds roles at the organization scope", func(t *testing.T) {
		t.Parallel()
		assert.Contains(t, got.SetupScript, "gcloud organizations add-iam-policy-binding")
		assert.Contains(t, got.SetupScript, "123456789")
		assert.Contains(t, got.SetupScript, "roles/cloudasset.viewer")
	})
}

func TestBuildGCPInstallScript_RejectsInvalidScope(t *testing.T) {
	t.Parallel()

	_, err := cloudaccount.BuildGCPInstallScript(
		[]coredata.CloudAccountAuditModule{coredata.CloudAccountAuditModuleAccessReview},
		coredata.CloudAccountScopeKindAWSAccount,
		"123456789012",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported scope kind")
}

func TestBuildGCPInstallScript_RejectsEmptyModules(t *testing.T) {
	t.Parallel()

	_, err := cloudaccount.BuildGCPInstallScript(
		nil,
		coredata.CloudAccountScopeKindGCPProject,
		"probo-target-project",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no roles or apis")
}
