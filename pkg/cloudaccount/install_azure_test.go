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

func TestBuildAzureInstallGuide_Subscription(t *testing.T) {
	t.Parallel()

	got, err := cloudaccount.BuildAzureInstallGuide(
		[]coredata.CloudAccountAuditModule{coredata.CloudAccountAuditModuleAccessReview},
		coredata.CloudAccountScopeKindAzureSubscription,
	)
	require.NoError(t, err)

	t.Run("each step has non-empty Title and Body", func(t *testing.T) {
		t.Parallel()
		require.NotEmpty(t, got.Steps)
		for i, s := range got.Steps {
			assert.NotEmptyf(t, s.Title, "step %d Title must be non-empty", i)
			assert.NotEmptyf(t, s.Body, "step %d Body must be non-empty", i)
			// Code is allowed to be empty for narrative-only steps.
		}
	})

	t.Run("includes a mandatory Grant admin consent step", func(t *testing.T) {
		t.Parallel()
		var found bool
		for _, s := range got.Steps {
			if strings.Contains(s.Title, "Grant admin consent") {
				found = true
				break
			}
		}
		assert.True(t, found, "an explicit 'Grant admin consent' step must be present (Directory.Read.All is admin-consent only)")
	})

	t.Run("required RBAC roles include Reader", func(t *testing.T) {
		t.Parallel()
		assert.Contains(t, got.RequiredRBACRoles, "Reader")
	})

	t.Run("required Graph permissions include Directory.Read.All", func(t *testing.T) {
		t.Parallel()
		assert.Contains(t, got.RequiredGraphPermissions, "Directory.Read.All")
	})

	t.Run("subscription guide labels the scope as 'Subscription'", func(t *testing.T) {
		t.Parallel()
		var labelHit bool
		for _, s := range got.Steps {
			if strings.Contains(s.Title, "Subscription") {
				labelHit = true
				break
			}
		}
		assert.True(t, labelHit, "at least one step title should mention the Subscription scope")
	})
}

func TestBuildAzureInstallGuide_ManagementGroup(t *testing.T) {
	t.Parallel()

	got, err := cloudaccount.BuildAzureInstallGuide(
		[]coredata.CloudAccountAuditModule{coredata.CloudAccountAuditModuleAccessReview},
		coredata.CloudAccountScopeKindAzureManagementGroup,
	)
	require.NoError(t, err)

	t.Run("includes Grant admin consent step", func(t *testing.T) {
		t.Parallel()
		var found bool
		for _, s := range got.Steps {
			if strings.Contains(s.Title, "Grant admin consent") {
				found = true
				break
			}
		}
		assert.True(t, found, "Grant admin consent step is mandatory regardless of scope")
	})

	t.Run("management group guide labels the scope as 'Management Group'", func(t *testing.T) {
		t.Parallel()
		var labelHit bool
		for _, s := range got.Steps {
			if strings.Contains(s.Title, "Management Group") {
				labelHit = true
				break
			}
		}
		assert.True(t, labelHit, "at least one step title should mention the Management Group scope")
	})
}

func TestBuildAzureInstallGuide_StepOrderingSnapshot(t *testing.T) {
	t.Parallel()

	// Snapshot the title-only ordering. A future refactor that reorders
	// steps will produce a visible diff here so the change is reviewed
	// rather than slipped in.
	got, err := cloudaccount.BuildAzureInstallGuide(
		[]coredata.CloudAccountAuditModule{coredata.CloudAccountAuditModuleAccessReview},
		coredata.CloudAccountScopeKindAzureSubscription,
	)
	require.NoError(t, err)

	want := []string{
		"Create the App Registration",
		"Generate a client secret",
		"Assign the Reader role at the Subscription scope",
		"Add the Microsoft Graph permissions",
		"Grant admin consent",
		"Paste the credentials back into Probo",
	}
	require.Len(t, got.Steps, len(want), "step count must match the snapshot")
	for i := range want {
		assert.Equalf(t, want[i], got.Steps[i].Title, "step %d title mismatch", i)
	}
}

func TestBuildAzureInstallGuide_RejectsInvalidScope(t *testing.T) {
	t.Parallel()

	_, err := cloudaccount.BuildAzureInstallGuide(
		[]coredata.CloudAccountAuditModule{coredata.CloudAccountAuditModuleAccessReview},
		coredata.CloudAccountScopeKindAWSAccount,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported scope kind")
}

func TestBuildAzureInstallGuide_RejectsEmptyModules(t *testing.T) {
	t.Parallel()

	_, err := cloudaccount.BuildAzureInstallGuide(
		nil,
		coredata.CloudAccountScopeKindAzureSubscription,
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no rbac or graph permissions")
}
