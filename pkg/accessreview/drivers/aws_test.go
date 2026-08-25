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
	"testing"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	cloudaws "go.probo.inc/probo/pkg/cloud/aws"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	// vcrAWSAccountID matches the synthetic ARNs in testdata/aws.yaml.
	vcrAWSAccountID = "123456789012"

	// Dummy static keys so the SDK will SigV4-sign. They are not AWS
	// credentials: this test never reads AWS_ACCESS_KEY_ID,
	// AWS_SECRET_ACCESS_KEY, or AWS_SESSION_TOKEN.
	vcrAWSAccessKey = "TESTINGACCESSKEY"
	vcrAWSSecretKey = "testing-secret-not-a-credential"
)

func TestAWSDriver(t *testing.T) {
	t.Parallel()

	rec := newAWSRecorder(t, "testdata/aws")
	session := cloudaws.NewSessionFromConfig(
		vcrAWSAccountID,
		cloudaws.CommercialPartition,
		awssdk.Config{
			Region: cloudaws.DefaultCommercialRegion,
			Credentials: awssdk.NewCredentialsCache(
				credentials.NewStaticCredentialsProvider(vcrAWSAccessKey, vcrAWSSecretKey, ""),
			),
			HTTPClient:       newVCRClient(rec, ""),
			RetryMaxAttempts: 1,
		},
	)

	records, err := NewAWSDriver(session, log.NewLogger(log.WithName("test"))).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	byID := make(map[string]AccountRecord, len(records))
	for _, record := range records {
		byID[record.ExternalID] = record
	}

	alice := byID["arn:aws:iam::123456789012:user/alice"]
	assert.Equal(t, "alice", alice.FullName)
	assert.ElementsMatch(t, []string{"Admins", "AdministratorAccess"}, alice.Roles)
	assert.Equal(t, coredata.MFAStatusEnabled, alice.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodPassword, alice.AuthMethod)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, alice.AccountType)
	require.NotNil(t, alice.IsAdmin)
	assert.True(t, *alice.IsAdmin)
	require.NotNil(t, alice.Active)
	assert.True(t, *alice.Active)
	assert.Empty(t, alice.Email)
	require.NotNil(t, alice.CreatedAt)
	assert.True(t, alice.CreatedAt.Equal(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)))
	require.NotNil(t, alice.LastLogin)
	assert.True(t, alice.LastLogin.Equal(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)))

	ci := byID["arn:aws:iam::123456789012:user/ci-deploy"]
	assert.Equal(t, "ci-deploy", ci.FullName)
	assert.Equal(t, []string{"deploy"}, ci.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, ci.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodAPIKey, ci.AuthMethod)
	assert.Equal(t, coredata.MFAStatusDisabled, ci.MFAStatus)
	require.NotNil(t, ci.Active)
	assert.True(t, *ci.Active)
	require.NotNil(t, ci.LastLogin)
	assert.True(t, ci.LastLogin.Equal(time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)))

	root := byID["arn:aws:iam::123456789012:root"]
	assert.Equal(t, "<root_account>", root.FullName)
	assert.Equal(t, coredata.MFAStatusDisabled, root.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodPassword, root.AuthMethod)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, root.AccountType)
	require.NotNil(t, root.IsAdmin)
	assert.True(t, *root.IsAdmin)
	require.NotNil(t, root.Active)
	assert.True(t, *root.Active)
	require.NotNil(t, root.LastLogin)
	assert.True(t, root.LastLogin.Equal(time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)))
}

func TestIAMUserRecord_MapsAdminConsoleUser(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	lastActive := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	record := iamUserRecord(
		iamUser{
			ARN:              "arn:aws:iam::123456789012:user/alice",
			Name:             "alice",
			Groups:           []string{"Admins"},
			AttachedPolicies: []string{"AdministratorAccess"},
			CreatedAt:        &createdAt,
			MFAEnabled:       new(true),
			ConsoleAccess:    new(true),
			AccessKeyActive:  new(false),
			LastActiveAt:     &lastActive,
		},
	)

	assert.Equal(t, "alice", record.FullName)
	assert.ElementsMatch(t, []string{"Admins", "AdministratorAccess"}, record.Roles)
	assert.Equal(t, coredata.MFAStatusEnabled, record.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodPassword, record.AuthMethod)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, record.AccountType)
	require.NotNil(t, record.IsAdmin)
	assert.True(t, *record.IsAdmin)
	require.NotNil(t, record.Active)
	assert.True(t, *record.Active)
	assert.Equal(t, &createdAt, record.CreatedAt)
	assert.Equal(t, &lastActive, record.LastLogin)

	// An IAM user has no email attribute. Inventing one would match the wrong
	// human during reconciliation.
	assert.Empty(t, record.Email)
}

func TestIAMUserRecord_MapsBareRootAsAdminWithUnknownSignals(t *testing.T) {
	t.Parallel()

	record := iamUserRecord(rootIdentity("aws", "123456789012"))

	assert.Equal(t, rootUser, record.FullName)
	assert.Equal(t, "arn:aws:iam::123456789012:root", record.ExternalID)
	require.NotNil(t, record.IsAdmin)
	assert.True(t, *record.IsAdmin)
	assert.Nil(t, record.Active)
	assert.Equal(t, coredata.MFAStatusUnknown, record.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, record.AuthMethod)
}

func TestIAMUserRecord_MapsRootAsAdmin(t *testing.T) {
	t.Parallel()

	record := iamUserRecord(
		iamUser{
			ARN:           "arn:aws:iam::123456789012:root",
			Name:          rootUser,
			MFAEnabled:    new(false),
			ConsoleAccess: new(true),
		},
	)

	assert.Equal(t, rootUser, record.FullName)
	require.NotNil(t, record.IsAdmin)
	assert.True(t, *record.IsAdmin)
	require.NotNil(t, record.Active)
	assert.True(t, *record.Active)
	assert.Empty(t, record.Roles)
}

func TestAWSIsRoot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		user iamUser
		want bool
	}{
		{
			name: "credential-report sentinel",
			user: iamUser{Name: rootUser},
			want: true,
		},
		{
			name: "canonical root ARN",
			user: iamUser{ARN: "arn:aws:iam::123456789012:root"},
			want: true,
		},
		{
			name: "IAM user named root",
			user: iamUser{ARN: "arn:aws:iam::123456789012:user/root", Name: "root"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.want, awsIsRoot(tt.user))
			},
		)
	}
}

func TestIAMUserRecord_MapsKeyOnlyUserAsServiceAccount(t *testing.T) {
	t.Parallel()

	record := iamUserRecord(
		iamUser{
			ARN:             "arn:aws:iam::123456789012:user/ci-deploy",
			Name:            "ci-deploy",
			InlinePolicies:  []string{"deploy"},
			MFAEnabled:      new(false),
			ConsoleAccess:   new(false),
			AccessKeyActive: new(true),
		},
	)

	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, record.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodAPIKey, record.AuthMethod)
	assert.Equal(t, coredata.MFAStatusDisabled, record.MFAStatus)
	assert.Equal(t, []string{"deploy"}, record.Roles)
	require.NotNil(t, record.Active)
	assert.True(t, *record.Active)
	assert.Nil(t, record.LastLogin)
}

func TestIAMUserRecord_MarksUserInactiveWithoutCredentials(t *testing.T) {
	t.Parallel()

	record := iamUserRecord(
		iamUser{
			ARN:             "arn:aws:iam::123456789012:user/locked",
			Name:            "locked",
			ConsoleAccess:   new(false),
			AccessKeyActive: new(false),
		},
	)

	require.NotNil(t, record.Active)
	assert.False(t, *record.Active)
}

// Absent signals must stay absent. Reporting false would tell a reviewer
// something the evidence does not support.
func TestIAMUserRecord_LeavesUnknownSignalsNil(t *testing.T) {
	t.Parallel()

	record := iamUserRecord(
		iamUser{
			ARN:    "arn:aws:iam::111111111111:user/mystery",
			Name:   "mystery",
			Groups: []string{"CustomPowerUsers"},
		},
	)

	assert.Equal(t, coredata.MFAStatusUnknown, record.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, record.AuthMethod)
	assert.Nil(t, record.Active)
	assert.Nil(t, record.IsAdmin)
}
