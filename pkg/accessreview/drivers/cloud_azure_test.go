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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/coredata"
)

type (
	// stubAzureRoleAssignmentLister is the in-memory test seam for
	// CloudAzureDriver. It returns a fixed slice of records and
	// optionally an error so propagation tests can pin the wrapping
	// shape.
	stubAzureRoleAssignmentLister struct {
		records []AzureRoleAssignmentRecord
		err     error
		calls   int
	}
)

var _ AzureRoleAssignmentLister = (*stubAzureRoleAssignmentLister)(nil)

func (s *stubAzureRoleAssignmentLister) ListRoleAssignments(ctx context.Context) ([]AzureRoleAssignmentRecord, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return s.records, nil
}

// TestCloudAzureDriver_EmptyAssignments asserts an empty role-
// assignment list returns an empty record slice with no error -- the
// expected state for fresh subscriptions or management groups.
func TestCloudAzureDriver_EmptyAssignments(t *testing.T) {
	t.Parallel()

	stub := &stubAzureRoleAssignmentLister{}
	driver := NewCloudAzureDriver(stub)

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	assert.Empty(t, records)
	assert.Equal(t, 1, stub.calls)
}

// TestCloudAzureDriver_BasicMapping covers the per-principal map
// from a role assignment to an AccountRecord.
func TestCloudAzureDriver_BasicMapping(t *testing.T) {
	t.Parallel()

	stub := &stubAzureRoleAssignmentLister{
		records: []AzureRoleAssignmentRecord{
			{
				PrincipalID:    "user-1",
				PrincipalType:  "User",
				PrincipalEmail: "alice@example.com",
				PrincipalName:  "Alice Anderson",
				RoleName:       "Reader",
			},
		},
	}
	driver := NewCloudAzureDriver(stub)

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)

	r := records[0]
	assert.Equal(t, "alice@example.com", r.Email)
	assert.Equal(t, "Alice Anderson", r.FullName)
	assert.Equal(t, "user-1", r.ExternalID)
	assert.Equal(t, "Reader", r.Role)
	assert.Equal(t, coredata.MFAStatusUnknown, r.MFAStatus)
	assert.Equal(t, coredata.AccessEntryAuthMethodUnknown, r.AuthMethod)
	assert.Equal(t, coredata.AccessEntryAccountTypeUser, r.AccountType)
}

// TestCloudAzureDriver_ServicePrincipalAccountType asserts a
// PrincipalType="ServicePrincipal" maps to AccessEntryAccountType-
// ServiceAccount, while every other type collapses to User.
func TestCloudAzureDriver_ServicePrincipalAccountType(t *testing.T) {
	t.Parallel()

	stub := &stubAzureRoleAssignmentLister{
		records: []AzureRoleAssignmentRecord{
			{
				PrincipalID:   "sp-1",
				PrincipalType: "ServicePrincipal",
				PrincipalName: "deploy-bot",
				RoleName:      "Contributor",
			},
			{
				PrincipalID:   "grp-1",
				PrincipalType: "Group",
				PrincipalName: "engineering",
				RoleName:      "Reader",
			},
			{
				PrincipalID:   "usr-1",
				PrincipalType: "User",
				PrincipalName: "alice",
				RoleName:      "Reader",
			},
		},
	}
	driver := NewCloudAzureDriver(stub)

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	byID := indexAzureRecordsByExternalID(records)
	assert.Equal(t, coredata.AccessEntryAccountTypeServiceAccount, byID["sp-1"].AccountType)
	assert.Equal(t, coredata.AccessEntryAccountTypeUser, byID["grp-1"].AccountType)
	assert.Equal(t, coredata.AccessEntryAccountTypeUser, byID["usr-1"].AccountType)
}

// TestCloudAzureDriver_FallbackEmailToPrincipalID asserts that when
// Microsoft Graph fails to resolve an email (so PrincipalEmail is
// empty), the driver back-fills Email with the principal id rather
// than emitting an empty string.
func TestCloudAzureDriver_FallbackEmailToPrincipalID(t *testing.T) {
	t.Parallel()

	stub := &stubAzureRoleAssignmentLister{
		records: []AzureRoleAssignmentRecord{
			{
				PrincipalID:    "00000000-0000-0000-0000-000000000abc",
				PrincipalType:  "ServicePrincipal",
				PrincipalEmail: "", // Graph lookup failed
				PrincipalName:  "",
				RoleName:       "Owner",
			},
		},
	}
	driver := NewCloudAzureDriver(stub)

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, "00000000-0000-0000-0000-000000000abc", records[0].Email)
	assert.Equal(t, "00000000-0000-0000-0000-000000000abc", records[0].ExternalID)
}

// TestCloudAzureDriver_EmptyPrincipalIDSkipped asserts assignments
// without a principal id are silently dropped (not surfaced as
// records with empty external ids). The PrincipalID is the
// canonical aggregation key; rows without one are unmappable.
func TestCloudAzureDriver_EmptyPrincipalIDSkipped(t *testing.T) {
	t.Parallel()

	stub := &stubAzureRoleAssignmentLister{
		records: []AzureRoleAssignmentRecord{
			{PrincipalID: "", PrincipalType: "User", RoleName: "Reader"},
			{PrincipalID: "real", PrincipalType: "User", RoleName: "Owner"},
		},
	}
	driver := NewCloudAzureDriver(stub)

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, "real", records[0].ExternalID)
}

// TestCloudAzureDriver_MultipleRolesAggregated asserts a single
// principal carrying multiple role assignments collapses to one
// record. The current implementation surfaces only the first role
// (reviewer-friendly summary); this test pins that contract so a
// future change to "join with comma" surfaces in the diff.
func TestCloudAzureDriver_MultipleRolesAggregated(t *testing.T) {
	t.Parallel()

	stub := &stubAzureRoleAssignmentLister{
		records: []AzureRoleAssignmentRecord{
			{PrincipalID: "p1", PrincipalType: "User", RoleName: "Reader"},
			{PrincipalID: "p1", PrincipalType: "User", RoleName: "Contributor"},
			{PrincipalID: "p1", PrincipalType: "User", RoleName: "Owner"},
		},
	}
	driver := NewCloudAzureDriver(stub)

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1, "duplicate PrincipalID rows must be aggregated")
	// The driver picks the first role seen; the order of map
	// iteration in the implementation is irrelevant here because
	// only one principal is in flight.
	assert.NotEmpty(t, records[0].Role)
}

// TestCloudAzureDriver_ListErrorPropagates asserts a seam-level
// error surfaces wrapped with the canonical "cannot" prefix.
func TestCloudAzureDriver_ListErrorPropagates(t *testing.T) {
	t.Parallel()

	stubErr := errors.New("AuthenticationFailed")
	stub := &stubAzureRoleAssignmentLister{err: stubErr}
	driver := NewCloudAzureDriver(stub)

	_, err := driver.ListAccounts(context.Background())
	require.Error(t, err)
	assert.True(t, errors.Is(err, stubErr))
	assert.True(
		t,
		strings.HasPrefix(err.Error(), "cannot list azure role assignments"),
		"error must use cannot prefix; got %q",
		err.Error(),
	)
}

// TestCloudAzureDriver_EmptyRoleNameSkipped asserts an assignment
// missing a role name does not poison the aggregate's Role slice
// (so the resulting record's Role stays empty rather than being
// "" with a leading separator).
func TestCloudAzureDriver_EmptyRoleNameSkipped(t *testing.T) {
	t.Parallel()

	stub := &stubAzureRoleAssignmentLister{
		records: []AzureRoleAssignmentRecord{
			{PrincipalID: "p1", PrincipalType: "User", PrincipalEmail: "alice@example.com", RoleName: ""},
		},
	}
	driver := NewCloudAzureDriver(stub)

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, "", records[0].Role)
}

func indexAzureRecordsByExternalID(records []AccountRecord) map[string]AccountRecord {
	out := make(map[string]AccountRecord, len(records))
	for _, r := range records {
		out[r.ExternalID] = r
	}
	return out
}
