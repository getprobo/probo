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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
	"go.probo.inc/probo/pkg/coredata"
	cloudlogging "google.golang.org/api/logging/v2"
	policyanalyzer "google.golang.org/api/policyanalyzer/v1"
)

func gcpTestRecords() []AccountRecord {
	return []AccountRecord{
		{
			Email:       "alice@example.com",
			ExternalID:  "user:alice@example.com",
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodSSO,
			MFAStatus:   coredata.MFAStatusUnknown,
		},
		{
			Email:       "eng@example.com",
			ExternalID:  "group:eng@example.com",
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			MFAStatus:   coredata.MFAStatusUnknown,
		},
		{
			Email:       "sa@my-project.iam.gserviceaccount.com",
			ExternalID:  "111111111111",
			AccountType: coredata.AccessReviewEntryAccountTypeServiceAccount,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			MFAStatus:   coredata.MFAStatusUnknown,
		},
	}
}

func TestFoldGCPActivityTimestamp(t *testing.T) {
	t.Parallel()

	older := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)
	found := map[string]time.Time{}

	foldGCPActivityTimestamp(found, "Alice@example.com", older)
	foldGCPActivityTimestamp(found, "alice@example.com", newer)
	foldGCPActivityTimestamp(found, "alice@example.com", older)

	require.Contains(t, found, "alice@example.com")
	assert.True(t, found["alice@example.com"].Equal(newer))
}

func TestParseGCPAdminActivityEntry(t *testing.T) {
	t.Parallel()

	t.Run(
		"reads principal email and timestamp",
		func(t *testing.T) {
			t.Parallel()

			email, at, ok := parseGCPAdminActivityEntry(
				&cloudlogging.LogEntry{
					Timestamp:    "2026-08-15T12:00:00Z",
					ProtoPayload: []byte(`{"authenticationInfo":{"principalEmail":"Alice@example.com"}}`),
				},
			)
			require.True(t, ok)
			assert.Equal(t, "alice@example.com", email)
			assert.True(t, at.Equal(time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)))
		},
	)

	t.Run(
		"rejects missing email",
		func(t *testing.T) {
			t.Parallel()

			_, _, ok := parseGCPAdminActivityEntry(
				&cloudlogging.LogEntry{
					Timestamp:    "2026-08-15T12:00:00Z",
					ProtoPayload: []byte(`{"authenticationInfo":{}}`),
				},
			)
			assert.False(t, ok)
		},
	)

	t.Run(
		"rejects invalid json",
		func(t *testing.T) {
			t.Parallel()

			_, _, ok := parseGCPAdminActivityEntry(
				&cloudlogging.LogEntry{
					Timestamp:    "2026-08-15T12:00:00Z",
					ProtoPayload: []byte(`{`),
				},
			)
			assert.False(t, ok)
		},
	)
}

func TestParseGCPPolicyActivity(t *testing.T) {
	t.Parallel()

	t.Run(
		"reads last authenticated time and email",
		func(t *testing.T) {
			t.Parallel()

			id, at, ok := parseGCPPolicyActivity(
				&policyanalyzer.GoogleCloudPolicyanalyzerV1Activity{
					FullResourceName: "//iam.googleapis.com/projects/123456789012/serviceAccounts/sa@my-project.iam.gserviceaccount.com",
					Activity:         []byte(`{"lastAuthenticatedTime":"2026-08-10T07:00:00Z"}`),
				},
			)
			require.True(t, ok)
			assert.Equal(t, "sa@my-project.iam.gserviceaccount.com", id)
			assert.True(t, at.Equal(time.Date(2026, 8, 10, 7, 0, 0, 0, time.UTC)))
		},
	)

	t.Run(
		"reads unique id from resource name",
		func(t *testing.T) {
			t.Parallel()

			id, _, ok := parseGCPPolicyActivity(
				&policyanalyzer.GoogleCloudPolicyanalyzerV1Activity{
					FullResourceName: "//iam.googleapis.com/projects/123456789012/serviceAccounts/111111111111",
					Activity:         []byte(`{"lastAuthenticatedTime":"2026-08-10T07:00:00Z"}`),
				},
			)
			require.True(t, ok)
			assert.Equal(t, "111111111111", id)
		},
	)

	t.Run(
		"rejects missing last authenticated time",
		func(t *testing.T) {
			t.Parallel()

			_, _, ok := parseGCPPolicyActivity(
				&policyanalyzer.GoogleCloudPolicyanalyzerV1Activity{
					FullResourceName: "//iam.googleapis.com/projects/123456789012/serviceAccounts/sa@my-project.iam.gserviceaccount.com",
					Activity:         []byte(`{}`),
				},
			)
			assert.False(t, ok)
		},
	)
}

func TestGcpServiceAccountIDFromResource(t *testing.T) {
	t.Parallel()

	id, ok := gcpServiceAccountIDFromResource(
		"//iam.googleapis.com/projects/123456789012/serviceAccounts/sa@my-project.iam.gserviceaccount.com/keys/KEY1",
	)
	require.True(t, ok)
	assert.Equal(t, "sa@my-project.iam.gserviceaccount.com", id)

	_, ok = gcpServiceAccountIDFromResource("//compute.googleapis.com/projects/123/instances/one")
	assert.False(t, ok)
}

func TestGcpRecordKind(t *testing.T) {
	t.Parallel()

	assert.Equal(
		t,
		gcpPrincipalUser,
		gcpRecordKind(AccountRecord{
			ExternalID:  "user:alice@example.com",
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
		}),
	)
	assert.Equal(
		t,
		gcpPrincipalGroup,
		gcpRecordKind(AccountRecord{
			ExternalID:  "group:eng@example.com",
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
		}),
	)
	assert.Equal(
		t,
		gcpPrincipalServiceAccount,
		gcpRecordKind(AccountRecord{
			ExternalID:  "111111111111",
			AccountType: coredata.AccessReviewEntryAccountTypeServiceAccount,
		}),
	)
}

func TestApplyGCPActivity(t *testing.T) {
	t.Parallel()

	aliceLogin := time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)
	saLogin := time.Date(2026, 8, 10, 7, 0, 0, 0, time.UTC)
	records := gcpTestRecords()

	applyGCPActivity(
		records,
		map[string]time.Time{
			"alice@example.com": aliceLogin,
			"111111111111":      saLogin,
			"eng@example.com":   aliceLogin,
		},
		map[string]bool{"sa@my-project.iam.gserviceaccount.com": true},
		map[string]coredata.MFAStatus{"alice@example.com": coredata.MFAStatusEnabled},
	)

	require.NotNil(t, records[0].LastLogin)
	assert.True(t, records[0].LastLogin.Equal(aliceLogin))
	assert.Equal(t, coredata.MFAStatusEnabled, records[0].MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodSSO, records[0].AuthMethod)

	assert.Nil(t, records[1].LastLogin)
	assert.Equal(t, coredata.MFAStatusUnknown, records[1].MFAStatus)

	require.NotNil(t, records[2].LastLogin)
	assert.True(t, records[2].LastLogin.Equal(saLogin))
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodAPIKey, records[2].AuthMethod)
	assert.Equal(t, coredata.MFAStatusUnknown, records[2].MFAStatus)
}

func TestApplyGCPActivity_DoesNotOverwriteKnownAuthMethod(t *testing.T) {
	t.Parallel()

	records := []AccountRecord{
		{
			Email:       "sa@my-project.iam.gserviceaccount.com",
			ExternalID:  "111111111111",
			AccountType: coredata.AccessReviewEntryAccountTypeServiceAccount,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodAPIKey,
			MFAStatus:   coredata.MFAStatusUnknown,
		},
	}

	applyGCPActivity(
		records,
		nil,
		map[string]bool{"sa@my-project.iam.gserviceaccount.com": true},
		nil,
	)

	assert.Equal(t, coredata.AccessReviewEntryAuthMethodAPIKey, records[0].AuthMethod)
	assert.Nil(t, records[0].LastLogin)
}

func TestGcpEmailFilterBatches(t *testing.T) {
	t.Parallel()

	since := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	emails := []string{"alice@example.com", "bob@example.com"}

	t.Run(
		"keeps a short list in one filter",
		func(t *testing.T) {
			t.Parallel()

			batches := gcpEmailFilterBatches("123456789012", since, emails, gcpLoggingFilterMaxLen)
			require.Len(t, batches, 1)
			assert.Contains(t, batches[0], "cloudaudit.googleapis.com%2Factivity")
			assert.NotContains(t, batches[0], "data_access")
			assert.Contains(t, batches[0], "alice@example.com")
			assert.Contains(t, batches[0], "bob@example.com")
		},
	)

	t.Run(
		"splits when the filter would exceed the budget",
		func(t *testing.T) {
			t.Parallel()

			first := gcpAdminActivityFilter("123456789012", since, emails[:1])
			batches := gcpEmailFilterBatches("123456789012", since, emails, len(first)+1)
			require.Len(t, batches, 2)
			assert.Contains(t, batches[0], "alice@example.com")
			assert.NotContains(t, batches[0], "bob@example.com")
			assert.Contains(t, batches[1], "bob@example.com")
		},
	)
}

func TestEnrichGCPIdentities_FillsLastLoginAndMFA(t *testing.T) {
	t.Parallel()

	rec := newGCPRecorder(t, "testdata/gcp_activity")
	session := newGCPTestSession(t, rec)
	records := gcpTestRecords()

	err := enrichGCPIdentities(context.Background(), session, records)
	require.NoError(t, err)

	require.NotNil(t, records[0].LastLogin)
	assert.True(t, records[0].LastLogin.Equal(time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)))
	assert.Equal(t, coredata.MFAStatusEnabled, records[0].MFAStatus)

	assert.Nil(t, records[1].LastLogin)
	assert.Equal(t, coredata.MFAStatusUnknown, records[1].MFAStatus)

	require.NotNil(t, records[2].LastLogin)
	assert.True(t, records[2].LastLogin.Equal(time.Date(2026, 8, 10, 7, 0, 0, 0, time.UTC)))
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodAPIKey, records[2].AuthMethod)
	assert.Equal(t, coredata.MFAStatusUnknown, records[2].MFAStatus)
}

func TestEnrichGCPIdentities_FallsBackToAdminActivityWhenPolicyAnalyzerDenied(t *testing.T) {
	t.Parallel()

	rec := newGCPRecorder(t, "testdata/gcp_activity_pa_denied")
	session := newGCPTestSession(t, rec)
	records := gcpTestRecords()

	err := enrichGCPIdentities(context.Background(), session, records)
	require.NoError(t, err)

	require.NotNil(t, records[0].LastLogin)
	assert.True(t, records[0].LastLogin.Equal(time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)))

	assert.Nil(t, records[1].LastLogin)

	require.NotNil(t, records[2].LastLogin)
	assert.True(t, records[2].LastLogin.Equal(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)))
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, records[2].AuthMethod)
}

func TestEnrichGCPIdentities_KeepsPolicyAnalyzerLoginWhenKeyActivityFails(t *testing.T) {
	t.Parallel()

	rec := newGCPRecorder(t, "testdata/gcp_activity_key_failed")
	session := newGCPTestSession(t, rec)
	records := gcpTestRecords()

	err := enrichGCPIdentities(context.Background(), session, records)
	require.NoError(t, err)

	require.NotNil(t, records[0].LastLogin)
	assert.True(t, records[0].LastLogin.Equal(time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)))
	assert.Equal(t, coredata.MFAStatusEnabled, records[0].MFAStatus)

	assert.Nil(t, records[1].LastLogin)
	assert.Equal(t, coredata.MFAStatusUnknown, records[1].MFAStatus)

	require.NotNil(t, records[2].LastLogin)
	assert.True(t, records[2].LastLogin.Equal(time.Date(2026, 8, 10, 7, 0, 0, 0, time.UTC)))
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, records[2].AuthMethod)
	assert.Equal(t, coredata.MFAStatusUnknown, records[2].MFAStatus)
}

func TestEnrichGCPIdentities_DegradesWhenActivityDenied(t *testing.T) {
	t.Parallel()

	rec := newGCPRecorder(t, "testdata/gcp_activity_denied")
	session := newGCPTestSession(t, rec)
	records := gcpTestRecords()

	err := enrichGCPIdentities(context.Background(), session, records)
	require.Error(t, err)
	assert.True(t, cloudgcp.As[cloudgcp.ErrPermissionDenied](err))

	for _, record := range records {
		assert.Nil(t, record.LastLogin)
		assert.Equal(t, coredata.MFAStatusUnknown, record.MFAStatus)
	}

	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, records[2].AuthMethod)
}

func TestEnrichGCPIdentities_FailsOnCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	rec := newGCPRecorder(t, "testdata/gcp_activity")
	session := newGCPTestSession(t, rec)
	records := gcpTestRecords()

	err := enrichGCPIdentities(ctx, session, records)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, records[0].LastLogin)
	assert.Equal(t, coredata.MFAStatusUnknown, records[0].MFAStatus)
}
