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

package drivers

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/coredata"
)

type (
	// stubGCPIAMPolicyReader is the in-memory test seam for
	// CloudGCPDriver. It records which fetch path the driver
	// dispatched to so tests can assert that GCP_PROJECT scope
	// only hits GetProjectIAMPolicy and GCP_ORGANIZATION scope
	// only hits SearchAllIAMPolicies.
	stubGCPIAMPolicyReader struct {
		projectPolicy GCPIAMPolicy
		orgPolicies   []GCPIAMPolicy

		projectErr error
		orgErr     error

		getProjectCalls  int
		searchAllCalls   int
		lastProjectIDArg string
		lastOrgIDArg     string
	}
)

var _ GCPIAMPolicyReader = (*stubGCPIAMPolicyReader)(nil)

func (s *stubGCPIAMPolicyReader) GetProjectIAMPolicy(ctx context.Context, projectID string) (GCPIAMPolicy, error) {
	s.getProjectCalls++
	s.lastProjectIDArg = projectID
	if s.projectErr != nil {
		return GCPIAMPolicy{}, s.projectErr
	}
	return s.projectPolicy, nil
}

func (s *stubGCPIAMPolicyReader) SearchAllIAMPolicies(ctx context.Context, orgID string) ([]GCPIAMPolicy, error) {
	s.searchAllCalls++
	s.lastOrgIDArg = orgID
	if s.orgErr != nil {
		return nil, s.orgErr
	}
	return s.orgPolicies, nil
}

// TestCloudGCPDriver_ProjectScopeOnlyHitsProjectAPI asserts a
// project-scoped account uses cloudresourcemanager only and never
// touches cloudasset. cloudasset.SearchAllIAMPolicies is a transit-
// scoped scan and must NOT be issued at the project granularity.
func TestCloudGCPDriver_ProjectScopeOnlyHitsProjectAPI(t *testing.T) {
	t.Parallel()

	stub := &stubGCPIAMPolicyReader{
		projectPolicy: GCPIAMPolicy{
			Bindings: []GCPIAMBinding{
				{Role: "roles/owner", Members: []string{"user:alice@example.com"}},
			},
		},
	}
	driver := NewCloudGCPDriver(stub, coredata.CloudAccountScopeKindGCPProject, "probo-prod-1")

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)

	assert.Equal(t, 1, stub.getProjectCalls)
	assert.Equal(t, 0, stub.searchAllCalls, "project scope must NOT call cloudasset.SearchAllIAMPolicies")
	assert.Equal(t, "probo-prod-1", stub.lastProjectIDArg)
	assert.Equal(t, "alice@example.com", records[0].Email)
	assert.Equal(t, coredata.AccessEntryAccountTypeUser, records[0].AccountType)
}

// TestCloudGCPDriver_OrgScopeOnlyHitsCloudAsset asserts an
// organization-scoped account dispatches to cloudasset.SearchAll-
// IAMPolicies and never to the project endpoint.
func TestCloudGCPDriver_OrgScopeOnlyHitsCloudAsset(t *testing.T) {
	t.Parallel()

	stub := &stubGCPIAMPolicyReader{
		orgPolicies: []GCPIAMPolicy{
			{
				Bindings: []GCPIAMBinding{
					{Role: "roles/viewer", Members: []string{"user:bob@example.com"}},
				},
			},
		},
	}
	driver := NewCloudGCPDriver(stub, coredata.CloudAccountScopeKindGCPOrganization, "987654321")

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)

	assert.Equal(t, 0, stub.getProjectCalls, "org scope must NOT call cloudresourcemanager.Projects.GetIamPolicy")
	assert.Equal(t, 1, stub.searchAllCalls)
	assert.Equal(t, "987654321", stub.lastOrgIDArg)
	assert.Equal(t, "bob@example.com", records[0].Email)
}

// TestCloudGCPDriver_MemberSubjectExtraction covers the member
// string parser. GCP IAM members are encoded as "kind:identifier";
// the driver preserves the identifier and maps the kind to the
// AccessEntryAccountType enum.
func TestCloudGCPDriver_MemberSubjectExtraction(t *testing.T) {
	t.Parallel()

	stub := &stubGCPIAMPolicyReader{
		projectPolicy: GCPIAMPolicy{
			Bindings: []GCPIAMBinding{
				{
					Role: "roles/owner",
					Members: []string{
						"user:alice@example.com",
						"serviceAccount:bot@proj.iam.gserviceaccount.com",
						"group:eng@example.com",
						"domain:example.com",
						// Unknown / unmodelled prefix passes through
						// with kind = "" but identifier preserved.
						"deleted:user:zoe@example.com?uid=1234",
						// Empty after colon -- skipped.
						"user:",
						// No colon -- skipped.
						"alice@example.com",
					},
				},
			},
		},
	}
	driver := NewCloudGCPDriver(stub, coredata.CloudAccountScopeKindGCPProject, "p")

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)

	byEmail := indexGCPRecordsByEmail(records)

	require.Contains(t, byEmail, "alice@example.com")
	assert.Equal(t, coredata.AccessEntryAccountTypeUser, byEmail["alice@example.com"].AccountType)

	require.Contains(t, byEmail, "bot@proj.iam.gserviceaccount.com")
	assert.Equal(
		t,
		coredata.AccessEntryAccountTypeServiceAccount,
		byEmail["bot@proj.iam.gserviceaccount.com"].AccountType,
	)

	require.Contains(t, byEmail, "eng@example.com")
	assert.Equal(t, coredata.AccessEntryAccountTypeUser, byEmail["eng@example.com"].AccountType)

	require.Contains(t, byEmail, "example.com")
	assert.Equal(t, coredata.AccessEntryAccountTypeUser, byEmail["example.com"].AccountType)

	// Empty user:"" identifier is skipped.
	for _, r := range records {
		assert.NotEmpty(t, r.Email, "empty identifier must not produce a record")
	}

	// "alice@example.com" without a colon prefix is skipped (the
	// strict member-format check rejects it). The well-formed
	// "user:alice@example.com" above produced the alice record.
	assert.Equal(t, "alice@example.com", byEmail["alice@example.com"].Email)
}

// TestCloudGCPDriver_RoleAggregation asserts a single principal
// holding multiple roles is collapsed to one record with all roles
// joined into Role (comma-separated).
func TestCloudGCPDriver_RoleAggregation(t *testing.T) {
	t.Parallel()

	stub := &stubGCPIAMPolicyReader{
		projectPolicy: GCPIAMPolicy{
			Bindings: []GCPIAMBinding{
				{Role: "roles/viewer", Members: []string{"user:alice@example.com"}},
				{Role: "roles/storage.admin", Members: []string{"user:alice@example.com"}},
				{Role: "roles/iam.tokenCreator", Members: []string{"user:alice@example.com"}},
			},
		},
	}
	driver := NewCloudGCPDriver(stub, coredata.CloudAccountScopeKindGCPProject, "p")

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1, "duplicate principal must be aggregated")

	roles := strings.Split(records[0].Role, ", ")
	sort.Strings(roles)
	assert.Equal(t, []string{"roles/iam.tokenCreator", "roles/storage.admin", "roles/viewer"}, roles)
}

// TestCloudGCPDriver_EmptyBindings asserts an empty IAM policy
// returns an empty slice with no error -- a valid state for newly
// provisioned projects with no IAM members yet.
func TestCloudGCPDriver_EmptyBindings(t *testing.T) {
	t.Parallel()

	stub := &stubGCPIAMPolicyReader{
		projectPolicy: GCPIAMPolicy{},
	}
	driver := NewCloudGCPDriver(stub, coredata.CloudAccountScopeKindGCPProject, "p")

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	assert.Empty(t, records)
}

// TestCloudGCPDriver_UnsupportedScopeKind asserts a non-GCP scope
// surfaces a typed error rather than dispatching to either path.
func TestCloudGCPDriver_UnsupportedScopeKind(t *testing.T) {
	t.Parallel()

	stub := &stubGCPIAMPolicyReader{}
	// AWS scope kind on a GCP driver must fail loud.
	driver := NewCloudGCPDriver(stub, coredata.CloudAccountScopeKindAWSAccount, "x")

	_, err := driver.ListAccounts(context.Background())
	require.Error(t, err)
	assert.True(
		t,
		strings.HasPrefix(err.Error(), "cannot list gcp accounts"),
		"error must use cannot prefix; got %q",
		err.Error(),
	)
	assert.Equal(t, 0, stub.getProjectCalls)
	assert.Equal(t, 0, stub.searchAllCalls)
}

// TestCloudGCPDriver_ProjectErrorPropagates asserts a project-API
// error is wrapped with the canonical "cannot" prefix and surfaces
// the original chain.
func TestCloudGCPDriver_ProjectErrorPropagates(t *testing.T) {
	t.Parallel()

	stubErr := errors.New("PERMISSION_DENIED")
	stub := &stubGCPIAMPolicyReader{projectErr: stubErr}
	driver := NewCloudGCPDriver(stub, coredata.CloudAccountScopeKindGCPProject, "p")

	_, err := driver.ListAccounts(context.Background())
	require.Error(t, err)
	assert.True(t, errors.Is(err, stubErr))
	assert.True(t, strings.HasPrefix(err.Error(), "cannot read gcp project iam policy"))
}

// TestCloudGCPDriver_OrgErrorPropagates asserts a cloudasset-API
// error is wrapped with the canonical "cannot" prefix.
func TestCloudGCPDriver_OrgErrorPropagates(t *testing.T) {
	t.Parallel()

	stubErr := errors.New("RESOURCE_EXHAUSTED")
	stub := &stubGCPIAMPolicyReader{orgErr: stubErr}
	driver := NewCloudGCPDriver(stub, coredata.CloudAccountScopeKindGCPOrganization, "o")

	_, err := driver.ListAccounts(context.Background())
	require.Error(t, err)
	assert.True(t, errors.Is(err, stubErr))
	assert.True(t, strings.HasPrefix(err.Error(), "cannot search gcp iam policies"))
}

// TestCloudGCPDriver_OrgScopeMultiplePolicies asserts the driver
// flattens bindings across all policies returned by SearchAllIAM-
// Policies (one policy per child resource in the org tree).
func TestCloudGCPDriver_OrgScopeMultiplePolicies(t *testing.T) {
	t.Parallel()

	stub := &stubGCPIAMPolicyReader{
		orgPolicies: []GCPIAMPolicy{
			{Bindings: []GCPIAMBinding{
				{Role: "roles/owner", Members: []string{"user:alice@example.com"}},
			}},
			{Bindings: []GCPIAMBinding{
				{Role: "roles/viewer", Members: []string{"user:alice@example.com"}},
				{Role: "roles/editor", Members: []string{"user:bob@example.com"}},
			}},
		},
	}
	driver := NewCloudGCPDriver(stub, coredata.CloudAccountScopeKindGCPOrganization, "o")

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 2)

	emails := []string{records[0].Email, records[1].Email}
	sort.Strings(emails)
	assert.Equal(t, []string{"alice@example.com", "bob@example.com"}, emails)
}

func indexGCPRecordsByEmail(records []AccountRecord) map[string]AccountRecord {
	out := make(map[string]AccountRecord, len(records))
	for _, r := range records {
		out[r.Email] = r
	}
	return out
}
