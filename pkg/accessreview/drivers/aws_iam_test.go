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
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/awsx/arn"
)

func TestParseCredentialReport_ReadsColumnsByHeaderName(t *testing.T) {
	t.Parallel()

	// Extra columns sit between the ones we read so a positional parser would
	// pick the wrong fields.
	csv := "" +
		"user,arn,extra,mfa_active,password_enabled,password_last_used," +
		"access_key_1_active,access_key_2_active," +
		"access_key_1_last_used_date,access_key_2_last_used_date\n" +
		"alice,arn:aws:iam::123456789012:user/alice,ignored,true,true,2026-06-01T00:00:00Z," +
		"false,false,N/A,N/A\n" +
		"ci-deploy,arn:aws:iam::123456789012:user/ci-deploy,ignored,false,false,N/A," +
		"true,false,2026-05-01T12:00:00Z,not_supported\n"

	rows, err := parseCredentialReport([]byte(csv))
	require.NoError(t, err)
	require.Len(t, rows, 2)

	alice := rows["alice"]
	assert.True(t, alice.mfaActive)
	assert.True(t, alice.passwordEnabled)
	assert.False(t, alice.accessKeyActive)
	require.NotNil(t, alice.passwordLastUsed)
	assert.True(t, alice.passwordLastUsed.Equal(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)))
	assert.Nil(t, alice.accessKeyLastUse)

	ci := rows["ci-deploy"]
	assert.False(t, ci.mfaActive)
	assert.False(t, ci.passwordEnabled)
	assert.True(t, ci.accessKeyActive)
	assert.Nil(t, ci.passwordLastUsed)
	require.NotNil(t, ci.accessKeyLastUse)
	assert.True(t, ci.accessKeyLastUse.Equal(time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)))
}

func TestParseCredentialReportTime_TreatsUnusedAsNil(t *testing.T) {
	t.Parallel()

	assert.Nil(t, parseCredentialReportTime(""))
	assert.Nil(t, parseCredentialReportTime("N/A"))
	assert.Nil(t, parseCredentialReportTime("not_supported"))
	assert.Nil(t, parseCredentialReportTime("not-a-time"))

	got := parseCredentialReportTime("2026-01-02T03:04:05Z")
	require.NotNil(t, got)
	assert.True(t, got.Equal(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)))
}

func TestJoinGroupPolicies_FoldsAndCompactsGroupPolicies(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	users := joinGroupPolicies(
		[]iamtypes.UserDetail{
			{
				Arn:        aws.String("arn:aws:iam::123456789012:user/alice"),
				UserName:   aws.String("alice"),
				GroupList:  []string{"Admins", "Developers"},
				CreateDate: &createdAt,
				AttachedManagedPolicies: []iamtypes.AttachedPolicy{
					{PolicyName: aws.String("AdministratorAccess")},
				},
				UserPolicyList: []iamtypes.PolicyDetail{
					{PolicyName: aws.String("inline-admin")},
				},
			},
			{
				Arn:      aws.String(""),
				UserName: aws.String("skipped"),
			},
		},
		map[string][]string{
			"Admins":     {"AdministratorAccess", "ForceMFA"},
			"Developers": {"ReadOnlyAccess", "ForceMFA"},
		},
	)

	require.Len(t, users, 1)
	alice := users[0]
	assert.Equal(t, "arn:aws:iam::123456789012:user/alice", alice.ARN)
	assert.Equal(t, "alice", alice.Name)
	assert.Equal(t, []string{"Admins", "Developers"}, alice.Groups)
	assert.Equal(t, []string{"inline-admin"}, alice.InlinePolicies)
	assert.Equal(t, &createdAt, alice.CreatedAt)
	assert.Equal(
		t,
		[]string{"AdministratorAccess", "ForceMFA", "ReadOnlyAccess"},
		alice.AttachedPolicies,
	)
	assert.Nil(t, alice.MFAEnabled)
	assert.Nil(t, alice.ConsoleAccess)
	assert.Nil(t, alice.AccessKeyActive)
}

func TestRootUserFromReport(t *testing.T) {
	t.Parallel()

	lastUsed := time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)
	report := map[string]credentialReportRow{
		rootUser: {
			mfaActive:        false,
			passwordEnabled:  true,
			passwordLastUsed: &lastUsed,
			accessKeyActive:  true,
		},
	}

	t.Run(
		"builds the canonical root identity",
		func(t *testing.T) {
			t.Parallel()

			user, ok := rootUserFromReport(arn.Partition, "123456789012", report)
			require.True(t, ok)
			assert.Equal(t, arn.IAM(arn.Partition, "123456789012", "root"), user.ARN)
			assert.Equal(t, rootUser, user.Name)
			require.NotNil(t, user.MFAEnabled)
			assert.False(t, *user.MFAEnabled)
			require.NotNil(t, user.ConsoleAccess)
			assert.True(t, *user.ConsoleAccess)
			require.NotNil(t, user.AccessKeyActive)
			assert.True(t, *user.AccessKeyActive)
			assert.Equal(t, &lastUsed, user.LastActiveAt)
		},
	)

	t.Run(
		"uses the session partition",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name      string
				partition string
				wantARN   string
			}{
				{
					name:      "commercial",
					partition: arn.Partition,
					wantARN:   "arn:aws:iam::123456789012:root",
				},
				{
					name:      "govcloud",
					partition: "aws-us-gov",
					wantARN:   "arn:aws-us-gov:iam::123456789012:root",
				},
				{
					name:      "china",
					partition: "aws-cn",
					wantARN:   "arn:aws-cn:iam::123456789012:root",
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					t.Parallel()

					user, ok := rootUserFromReport(tt.partition, "123456789012", report)
					require.True(t, ok)
					assert.Equal(t, tt.wantARN, user.ARN)
				})
			}
		},
	)

	t.Run(
		"absent when the report has no root row",
		func(t *testing.T) {
			t.Parallel()

			_, ok := rootUserFromReport(arn.Partition, "123456789012", map[string]credentialReportRow{})
			assert.False(t, ok)
		},
	)

	t.Run(
		"absent when the report is nil",
		func(t *testing.T) {
			t.Parallel()

			_, ok := rootUserFromReport(arn.Partition, "123456789012", nil)
			assert.False(t, ok)
		},
	)
}

func TestRootIdentity(t *testing.T) {
	t.Parallel()

	user := rootIdentity(arn.Partition, "123456789012")

	assert.Equal(t, arn.IAM(arn.Partition, "123456789012", "root"), user.ARN)
	assert.Equal(t, rootUser, user.Name)
	assert.Nil(t, user.MFAEnabled)
	assert.Nil(t, user.ConsoleAccess)
	assert.Nil(t, user.AccessKeyActive)
	assert.Nil(t, user.LastActiveAt)
}

func TestCredentialReportAlreadyGenerated(t *testing.T) {
	t.Parallel()

	assert.True(t, credentialReportAlreadyGenerated(&iamtypes.LimitExceededException{}))
	assert.True(t, credentialReportAlreadyGenerated(&iamtypes.ReportGenerationLimitExceededException{}))
	assert.False(t, credentialReportAlreadyGenerated(&iamtypes.ServiceFailureException{}))
	assert.False(t, credentialReportAlreadyGenerated(fmt.Errorf("access denied")))
}

func TestCredentialReportBuilding(t *testing.T) {
	t.Parallel()

	assert.True(t, credentialReportBuilding(&iamtypes.CredentialReportNotReadyException{}))
	assert.True(t, credentialReportBuilding(&iamtypes.CredentialReportNotPresentException{}))
	assert.False(t, credentialReportBuilding(fmt.Errorf("access denied")))
}
