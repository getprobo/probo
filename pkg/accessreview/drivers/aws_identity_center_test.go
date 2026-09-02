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

package drivers

import (
	"context"
	"slices"
	"testing"
	"time"

	istypes "github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	cloudaws "go.probo.inc/probo/pkg/cloud/aws"
	"go.probo.inc/probo/pkg/coredata"
)

func TestIdentityCenterRegions_PutsPreferredFirst(t *testing.T) {
	t.Parallel()

	t.Run(
		"commercial preferred is first and unique",
		func(t *testing.T) {
			t.Parallel()

			regions := identityCenterRegions(cloudaws.CommercialPartition, "eu-west-1")
			require.NotEmpty(t, regions)
			assert.Equal(t, "eu-west-1", regions[0])

			sorted := slices.Clone(regions)
			slices.Sort(sorted)
			assert.Equal(t, sorted, slices.Compact(slices.Clone(sorted)))
			assert.Contains(t, regions, "us-east-1")
			assert.Contains(t, regions, "eu-central-1")
		},
	)

	t.Run(
		"govcloud preferred is first",
		func(t *testing.T) {
			t.Parallel()

			regions := identityCenterRegions(cloudaws.GovPartition, cloudaws.DefaultGovRegion)
			require.Equal(t, []string{cloudaws.DefaultGovRegion, "us-gov-east-1"}, regions)
		},
	)

	t.Run(
		"china is preferred only",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				[]string{cloudaws.DefaultChinaRegion},
				identityCenterRegions(cloudaws.ChinaPartition, cloudaws.DefaultChinaRegion),
			)
			assert.Empty(t, identityCenterRegions(cloudaws.ChinaPartition, ""))
		},
	)
}

func TestListIdentityCenterUsers_FindsInstanceOutsideSessionRegion(t *testing.T) {
	t.Parallel()

	rec := newAWSRecorder(t, "testdata/aws_identity_center_eu_west_1")
	session := newAWSTestSession(t, rec)

	users, err := listIdentityCenterUsersInRegions(
		context.Background(),
		session,
		log.NewLogger(log.WithName("test")),
		[]string{cloudaws.DefaultCommercialRegion, "eu-west-1"},
	)
	require.NoError(t, err)
	require.Len(t, users, 3)

	records := make([]AccountRecord, 0, len(users))
	for _, user := range users {
		records = append(records, identityCenterUserRecord(user))
	}

	byID := make(map[string]AccountRecord, len(records))
	for _, record := range records {
		byID[record.ExternalID] = record
	}

	bob := byID["arn:aws:identitystore::123456789012:user/11111111-1111-1111-1111-111111111111"]
	assert.Equal(t, "Bob Admin", bob.FullName)
	assert.Equal(t, []string{"AdministratorAccess"}, bob.Roles)
	require.NotNil(t, bob.IsAdmin)
	assert.True(t, *bob.IsAdmin)

	carol := byID["arn:aws:identitystore::123456789012:user/22222222-2222-2222-2222-222222222222"]
	assert.Equal(t, "Carol Engineer", carol.FullName)
	assert.ElementsMatch(t, []string{"Engineers", "ReadOnlyAccess"}, carol.Roles)
	assert.Equal(t, coredata.MFAStatusUnknown, carol.MFAStatus)
	require.NotNil(t, carol.LastLogin)
	assert.True(t, carol.LastLogin.Equal(time.Unix(1782864000, 0).UTC()))

	assert.Equal(t, coredata.MFAStatusEnabled, bob.MFAStatus)
	require.NotNil(t, bob.LastLogin)
	assert.True(t, bob.LastLogin.Equal(time.Unix(1786752000, 0).UTC()))

	dave := byID["arn:aws:identitystore::123456789012:user/33333333-3333-3333-3333-333333333333"]
	assert.Equal(t, "Dave Unused", dave.FullName)
	assert.Empty(t, dave.Roles)
	assert.Nil(t, dave.IsAdmin)
}

func TestListIdentityCenterUsers_ListsDirectAndGroupAssignments(t *testing.T) {
	t.Parallel()

	rec := newAWSRecorder(t, "testdata/aws_identity_center")
	session := newAWSTestSession(t, rec)

	users, err := listIdentityCenterUsers(context.Background(), session, log.NewLogger(log.WithName("test")))
	require.NoError(t, err)
	require.Len(t, users, 3)

	records := make([]AccountRecord, 0, len(users))
	for _, user := range users {
		records = append(records, identityCenterUserRecord(user))
	}

	byID := make(map[string]AccountRecord, len(records))
	for _, record := range records {
		byID[record.ExternalID] = record
	}

	bob := byID["arn:aws:identitystore::123456789012:user/11111111-1111-1111-1111-111111111111"]
	assert.Equal(t, "Bob Admin", bob.FullName)
	assert.Equal(t, "bob@example.com", bob.Email)
	assert.Equal(t, "Security Lead", bob.JobTitle)
	assert.Equal(t, []string{"AdministratorAccess"}, bob.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodSSO, bob.AuthMethod)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, bob.AccountType)
	assert.Equal(t, coredata.MFAStatusEnabled, bob.MFAStatus)
	require.NotNil(t, bob.IsAdmin)
	assert.True(t, *bob.IsAdmin)
	require.NotNil(t, bob.Active)
	assert.True(t, *bob.Active)
	require.NotNil(t, bob.CreatedAt)
	assert.True(t, bob.CreatedAt.Equal(time.Unix(1767225600, 0).UTC()))
	require.NotNil(t, bob.LastLogin)
	assert.True(t, bob.LastLogin.Equal(time.Unix(1786752000, 0).UTC()))

	carol := byID["arn:aws:identitystore::123456789012:user/22222222-2222-2222-2222-222222222222"]
	assert.Equal(t, "Carol Engineer", carol.FullName)
	assert.Equal(t, "carol@example.com", carol.Email)
	assert.ElementsMatch(t, []string{"Engineers", "ReadOnlyAccess"}, carol.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodSSO, carol.AuthMethod)
	assert.Equal(t, coredata.MFAStatusUnknown, carol.MFAStatus)
	assert.Nil(t, carol.IsAdmin)
	require.NotNil(t, carol.Active)
	assert.True(t, *carol.Active)
	require.NotNil(t, carol.LastLogin)
	assert.True(t, carol.LastLogin.Equal(time.Unix(1782864000, 0).UTC()))

	dave := byID["arn:aws:identitystore::123456789012:user/33333333-3333-3333-3333-333333333333"]
	assert.Equal(t, "Dave Unused", dave.FullName)
	assert.Equal(t, "dave@example.com", dave.Email)
	assert.Empty(t, dave.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodSSO, dave.AuthMethod)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, dave.AccountType)
	assert.Equal(t, coredata.MFAStatusUnknown, dave.MFAStatus)
	assert.Nil(t, dave.IsAdmin)
	assert.Nil(t, dave.Active)
	assert.Nil(t, dave.LastLogin)
}

func TestListIdentityCenterUsers_DegradesWhenActivityDenied(t *testing.T) {
	t.Parallel()

	rec := newAWSRecorder(t, "testdata/aws_identity_center_activity_denied")
	session := newAWSTestSession(t, rec)

	users, err := listIdentityCenterUsers(context.Background(), session, log.NewLogger(log.WithName("test")))
	require.NoError(t, err)
	require.Len(t, users, 3)

	records := make([]AccountRecord, 0, len(users))
	for _, user := range users {
		records = append(records, identityCenterUserRecord(user))
	}

	for _, record := range records {
		assert.Equal(t, coredata.MFAStatusUnknown, record.MFAStatus)
		assert.Nil(t, record.LastLogin)
	}
}

func TestListIdentityCenterUsers_DegradesWhenLaterSSOAdminDenied(t *testing.T) {
	t.Parallel()

	rec := newAWSRecorder(t, "testdata/aws_identity_center_permission_sets_denied")
	session := newAWSTestSession(t, rec)

	users, err := listIdentityCenterUsers(context.Background(), session, log.NewLogger(log.WithName("test")))
	require.NoError(t, err)
	assert.Empty(t, users)
}

func TestListIdentityCenterUsers_FailsOnIdentityStoreDenied(t *testing.T) {
	t.Parallel()

	rec := newAWSRecorder(t, "testdata/aws_identity_center_users_denied")
	session := newAWSTestSession(t, rec)

	users, err := listIdentityCenterUsers(context.Background(), session, log.NewLogger(log.WithName("test")))
	require.Error(t, err)
	assert.ErrorContains(t, err, "cannot list iam identity store users")
	assert.Empty(t, users)
}

func TestListIdentityCenterUsers_FailsOnCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	rec := newAWSRecorder(t, "testdata/aws_identity_center")
	session := newAWSTestSession(t, rec)

	users, err := listIdentityCenterUsers(ctx, session, log.NewLogger(log.WithName("test")))
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, users)
}

func TestIdentityCenterUserRecord_MapsPrimaryEmailAndAdminName(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	record := identityCenterUserRecord(
		identityCenterUser{
			ARN:         "arn:aws:identitystore::123456789012:user/bob",
			UserName:    "bob",
			DisplayName: "Bob Admin",
			Email:       "bob@example.com",
			Title:       "Security Lead",
			Grants:      []string{"AdministratorAccess"},
			Admin:       true,
			Status:      istypes.UserStatusEnabled,
			CreatedAt:   &createdAt,
		},
	)

	assert.Equal(t, "bob@example.com", record.Email)
	assert.Equal(t, "Bob Admin", record.FullName)
	assert.Equal(t, "Security Lead", record.JobTitle)
	assert.Equal(t, []string{"AdministratorAccess"}, record.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodSSO, record.AuthMethod)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, record.AccountType)
	assert.Equal(t, coredata.MFAStatusUnknown, record.MFAStatus)
	require.NotNil(t, record.IsAdmin)
	assert.True(t, *record.IsAdmin)
	require.NotNil(t, record.Active)
	assert.True(t, *record.Active)
	assert.Equal(t, &createdAt, record.CreatedAt)
	assert.Nil(t, record.LastLogin)
	assert.Equal(t, "arn:aws:identitystore::123456789012:user/bob", record.ExternalID)
}

func TestIdentityCenterUserRecord_MapsActivitySignals(t *testing.T) {
	t.Parallel()

	lastLogin := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)

	record := identityCenterUserRecord(
		identityCenterUser{
			ARN:        "arn:aws:identitystore::123456789012:user/bob",
			UserName:   "bob",
			MFAEnabled: new(true),
			LastLogin:  &lastLogin,
		},
	)

	assert.Equal(t, coredata.MFAStatusEnabled, record.MFAStatus)
	assert.Equal(t, &lastLogin, record.LastLogin)
}

func TestIdentityCenterUserRecord_FallsBackToUserNameEmailAndDisplayName(t *testing.T) {
	t.Parallel()

	record := identityCenterUserRecord(
		identityCenterUser{
			ARN:      "arn:aws:identitystore::123456789012:user/carol",
			UserName: "carol@example.com",
			Email:    identityCenterEmail(nil, "carol@example.com"),
			Grants:   []string{"ReadOnlyAccess", "Engineers"},
			Status:   istypes.UserStatusDisabled,
		},
	)

	assert.Equal(t, "carol@example.com", record.Email)
	assert.Equal(t, "carol@example.com", record.FullName)
	assert.Nil(t, record.IsAdmin)
	require.NotNil(t, record.Active)
	assert.False(t, *record.Active)
}

func TestIdentityCenterUserRecord_LeavesUnknownSignalsNil(t *testing.T) {
	t.Parallel()

	record := identityCenterUserRecord(
		identityCenterUser{
			ARN:      "arn:aws:identitystore::111111111111:user/mystery",
			UserName: "mystery",
			Grants:   []string{"CustomPowerUsers"},
		},
	)

	assert.Empty(t, record.Email)
	assert.Equal(t, "mystery", record.FullName)
	assert.Equal(t, coredata.MFAStatusUnknown, record.MFAStatus)
	assert.Nil(t, record.Active)
	assert.Nil(t, record.IsAdmin)
}

func TestIdentityCenterAssignedUsers_DropsEmptyGrants(t *testing.T) {
	t.Parallel()

	assigned := identityCenterAssignedUsers(
		[]identityCenterUser{
			{ID: "bob", Grants: []string{"AdministratorAccess"}},
			{ID: "dave"},
			{ID: "carol", Grants: []string{"ReadOnlyAccess"}},
		},
	)

	require.Len(t, assigned, 2)
	assert.Equal(t, "bob", assigned[0].ID)
	assert.Equal(t, "carol", assigned[1].ID)
}

func TestIdentityCenterIsAdmin_TreatsSSOAdministratorAccessAsAdmin(t *testing.T) {
	t.Parallel()

	require.NotNil(t, identityCenterIsAdmin(identityCenterUser{Grants: []string{awsSSOAdministratorAccess}}))
	assert.True(t, *identityCenterIsAdmin(identityCenterUser{Grants: []string{awsSSOAdministratorAccess}}))
	assert.Nil(t, identityCenterIsAdmin(identityCenterUser{Grants: []string{"ReadOnlyAccess"}}))
	require.NotNil(t, identityCenterIsAdmin(identityCenterUser{Admin: true, Grants: []string{"CustomAdmin"}}))
}

func TestIdentityCenterEmail_PrefersPrimaryThenUserName(t *testing.T) {
	t.Parallel()

	primary := "primary@example.com"
	other := "other@example.com"

	assert.Equal(
		t,
		primary,
		identityCenterEmail(
			[]istypes.Email{
				{Value: &other},
				{Primary: true, Value: &primary},
			},
			"alice",
		),
	)
	assert.Equal(t, other, identityCenterEmail([]istypes.Email{{Value: &other}}, "alice"))
	assert.Equal(t, "alice@example.com", identityCenterEmail(nil, "alice@example.com"))
	assert.Empty(t, identityCenterEmail(nil, "alice"))
}

func TestIdentityCenterUserARN_UsesPartitionAndAccount(t *testing.T) {
	t.Parallel()

	assert.Equal(
		t,
		"arn:aws:identitystore::123456789012:user/11111111-1111-1111-1111-111111111111",
		identityCenterUserARN("aws", "123456789012", "11111111-1111-1111-1111-111111111111"),
	)
	assert.Equal(
		t,
		"arn:aws-us-gov:identitystore::123456789012:user/bob",
		identityCenterUserARN("aws-us-gov", "123456789012", "bob"),
	)
}
