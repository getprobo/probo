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

	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/coredata"
)

type (
	// stubAWSIAMReader is the in-memory test seam for CloudAWSDriver.
	// It records every call so tests can assert pagination, MFA
	// lookups, and credential-report dispatch without spinning up
	// AWS SDK transport.
	stubAWSIAMReader struct {
		// users is a per-page slice. The driver issues ListUsers
		// once per page and follows IsTruncated/Marker. The stub
		// returns users[i] on the i-th call.
		users [][]iamtypes.User

		// truncated controls the IsTruncated bit per page. nil means
		// no pagination -- the stub returns IsTruncated=false on
		// every call.
		truncated []bool

		// listUsersErr is returned (if non-nil) on every ListUsers
		// call. Used by the propagation test.
		listUsersErr error

		// mfaDevices keys MFA device count by user name. Absent
		// users default to zero MFA devices.
		mfaDevices map[string]int

		// mfaErr is returned (if non-nil) on every ListMFADevices
		// call. Used by the MFA-error propagation test.
		mfaErr error

		// Recorded calls for inspection.
		listUsersCalls         int
		listMFADevicesCalls    int
		listMFADevicesUserArgs []string
	}
)

var _ AWSIAMReader = (*stubAWSIAMReader)(nil)

func (s *stubAWSIAMReader) ListUsers(
	ctx context.Context,
	in *iam.ListUsersInput,
	opts ...func(*iam.Options),
) (*iam.ListUsersOutput, error) {
	if s.listUsersErr != nil {
		return nil, s.listUsersErr
	}

	page := s.listUsersCalls
	s.listUsersCalls++

	if page >= len(s.users) {
		return &iam.ListUsersOutput{}, nil
	}

	out := &iam.ListUsersOutput{Users: s.users[page]}
	if page < len(s.truncated) && s.truncated[page] {
		out.IsTruncated = true
		nextMarker := "page-" + stringFromInt(page+1)
		out.Marker = &nextMarker
	}
	return out, nil
}

func (s *stubAWSIAMReader) ListMFADevices(
	ctx context.Context,
	in *iam.ListMFADevicesInput,
	opts ...func(*iam.Options),
) (*iam.ListMFADevicesOutput, error) {
	s.listMFADevicesCalls++
	if in.UserName != nil {
		s.listMFADevicesUserArgs = append(s.listMFADevicesUserArgs, *in.UserName)
	}

	if s.mfaErr != nil {
		return nil, s.mfaErr
	}

	count := 0
	if in.UserName != nil {
		count = s.mfaDevices[*in.UserName]
	}
	devices := make([]iamtypes.MFADevice, count)
	return &iam.ListMFADevicesOutput{MFADevices: devices}, nil
}

func (s *stubAWSIAMReader) GenerateCredentialReport(
	ctx context.Context,
	in *iam.GenerateCredentialReportInput,
	opts ...func(*iam.Options),
) (*iam.GenerateCredentialReportOutput, error) {
	return &iam.GenerateCredentialReportOutput{}, nil
}

func (s *stubAWSIAMReader) GetCredentialReport(
	ctx context.Context,
	in *iam.GetCredentialReportInput,
	opts ...func(*iam.Options),
) (*iam.GetCredentialReportOutput, error) {
	return &iam.GetCredentialReportOutput{}, nil
}

// stringFromInt is a tiny helper to keep the stub free of fmt
// imports (which would otherwise drag the seam into more allocator
// surface than necessary for a unit test).
func stringFromInt(i int) string {
	switch i {
	case 0:
		return "0"
	case 1:
		return "1"
	case 2:
		return "2"
	case 3:
		return "3"
	default:
		return "n"
	}
}

func awsUser(name, id string) iamtypes.User {
	return iamtypes.User{
		UserName: stringPtr(name),
		UserId:   stringPtr(id),
	}
}

func stringPtr(s string) *string { return &s }

// TestCloudAWSDriver_EmptyAccount asserts an empty IAM tenant
// returns an empty record slice with no error -- this is a
// legitimate state for fresh accounts and must not surface as a
// failure.
func TestCloudAWSDriver_EmptyAccount(t *testing.T) {
	t.Parallel()

	stub := &stubAWSIAMReader{}
	driver := NewCloudAWSDriver(stub)

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	assert.Empty(t, records)
	assert.Equal(t, 1, stub.listUsersCalls, "one ListUsers call drains the empty page")
	assert.Equal(t, 0, stub.listMFADevicesCalls, "no MFA lookup when there are no users")
}

// TestCloudAWSDriver_MFAEnabled asserts users with at least one
// MFA device come back with MFAStatus = enabled.
func TestCloudAWSDriver_MFAEnabled(t *testing.T) {
	t.Parallel()

	stub := &stubAWSIAMReader{
		users: [][]iamtypes.User{
			{
				awsUser("alice", "AIDA00000000ALICE"),
			},
		},
		mfaDevices: map[string]int{"alice": 1},
	}
	driver := NewCloudAWSDriver(stub)

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)

	r := records[0]
	assert.Equal(t, coredata.MFAStatusEnabled, r.MFAStatus)
	assert.Equal(t, coredata.AccessEntryAuthMethodUnknown, r.AuthMethod)
	assert.Equal(t, coredata.AccessEntryAccountTypeUser, r.AccountType)
	assert.Equal(t, "alice", r.Email)
	assert.Equal(t, "alice", r.FullName)
	assert.Equal(t, "AIDA00000000ALICE", r.ExternalID)

	assert.Equal(t, 1, stub.listMFADevicesCalls)
	assert.Equal(t, []string{"alice"}, stub.listMFADevicesUserArgs)
}

// TestCloudAWSDriver_MFADisabled asserts users with zero MFA
// devices come back with MFAStatus = disabled (NOT unknown -- the
// driver successfully listed the devices, the answer was zero).
func TestCloudAWSDriver_MFADisabled(t *testing.T) {
	t.Parallel()

	stub := &stubAWSIAMReader{
		users: [][]iamtypes.User{
			{
				awsUser("bob", "AIDA00000000BOB"),
			},
		},
		mfaDevices: map[string]int{}, // no devices
	}
	driver := NewCloudAWSDriver(stub)

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, coredata.MFAStatusDisabled, records[0].MFAStatus)
}

// TestCloudAWSDriver_Pagination asserts the driver follows the
// ListUsers Marker chain across two pages and aggregates the records.
func TestCloudAWSDriver_Pagination(t *testing.T) {
	t.Parallel()

	stub := &stubAWSIAMReader{
		users: [][]iamtypes.User{
			{awsUser("alice", "id1"), awsUser("bob", "id2")},
			{awsUser("carol", "id3")},
		},
		truncated: []bool{true, false},
	}
	driver := NewCloudAWSDriver(stub)

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)
	assert.Equal(t, 2, stub.listUsersCalls, "two ListUsers calls follow the Marker chain")
}

// TestCloudAWSDriver_ListUsersErrorPropagates asserts a ListUsers
// error surfaces wrapped with the canonical "cannot" prefix.
func TestCloudAWSDriver_ListUsersErrorPropagates(t *testing.T) {
	t.Parallel()

	stubErr := errors.New("ThrottlingException")
	stub := &stubAWSIAMReader{listUsersErr: stubErr}
	driver := NewCloudAWSDriver(stub)

	_, err := driver.ListAccounts(context.Background())
	require.Error(t, err)
	assert.True(t, errors.Is(err, stubErr), "error chain must wrap original SDK error")
	assert.True(
		t,
		strings.HasPrefix(err.Error(), "cannot list aws iam users"),
		"error message must use cannot prefix; got %q",
		err.Error(),
	)
}

// TestCloudAWSDriver_ListMFADevicesErrorPropagates asserts an MFA
// lookup error short-circuits the whole list and is wrapped with
// the canonical "cannot" prefix. The driver must not silently
// downgrade MFA fetches.
func TestCloudAWSDriver_ListMFADevicesErrorPropagates(t *testing.T) {
	t.Parallel()

	stubErr := errors.New("AccessDenied")
	stub := &stubAWSIAMReader{
		users: [][]iamtypes.User{
			{awsUser("alice", "id")},
		},
		mfaErr: stubErr,
	}
	driver := NewCloudAWSDriver(stub)

	_, err := driver.ListAccounts(context.Background())
	require.Error(t, err)
	assert.True(t, errors.Is(err, stubErr))
	assert.True(t, strings.HasPrefix(err.Error(), "cannot list aws iam mfa devices"))
}

// TestCloudAWSDriver_DefaultsForMissingFields asserts users coming
// back with nil UserName / nil UserId still produce an emit-able
// record (with empty strings) -- the driver must never panic on
// a sparse SDK response.
func TestCloudAWSDriver_DefaultsForMissingFields(t *testing.T) {
	t.Parallel()

	stub := &stubAWSIAMReader{
		users: [][]iamtypes.User{
			{
				// Both nil pointers -- e.g. SDK returned a row
				// without populating the fields.
				{UserName: nil, UserId: nil},
			},
		},
	}
	driver := NewCloudAWSDriver(stub)

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)

	r := records[0]
	assert.Equal(t, "", r.Email)
	assert.Equal(t, "", r.FullName)
	assert.Equal(t, "", r.ExternalID)
	assert.Equal(t, coredata.MFAStatusUnknown, r.MFAStatus, "no UserName means no MFA lookup ran, so status stays Unknown")
	assert.Equal(t, 0, stub.listMFADevicesCalls)
}
