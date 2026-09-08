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
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cloudazure "go.probo.inc/probo/pkg/cloud/azure"
	"go.probo.inc/probo/pkg/coredata"
)

func azureTestRecords() []AccountRecord {
	return []AccountRecord{
		{
			Email:       "alice@probo-azure.test",
			ExternalID:  vcrAzureAliceID,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodSSO,
			MFAStatus:   coredata.MFAStatusUnknown,
		},
		{
			Email:       "eng@probo-azure.test",
			ExternalID:  vcrAzureEngID,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			MFAStatus:   coredata.MFAStatusUnknown,
		},
		{
			ExternalID:  vcrAzureCIID,
			AccountType: coredata.AccessReviewEntryAccountTypeServiceAccount,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
			MFAStatus:   coredata.MFAStatusUnknown,
		},
	}
}

func TestApplyAzureActivity(t *testing.T) {
	t.Parallel()

	fallback := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	records := azureTestRecords()

	applyAzureActivity(
		records,
		map[string]time.Time{
			vcrAzureAliceID: azureAliceLastLogin,
			vcrAzureEngID:   fallback,
			vcrAzureCIID:    fallback,
		},
		map[string]coredata.MFAStatus{
			vcrAzureAliceID: coredata.MFAStatusEnabled,
			vcrAzureEngID:   coredata.MFAStatusDisabled,
			vcrAzureCIID:    coredata.MFAStatusEnabled,
		},
	)

	require.NotNil(t, records[0].LastLogin)
	assert.True(t, records[0].LastLogin.Equal(azureAliceLastLogin))
	assert.Equal(t, coredata.MFAStatusEnabled, records[0].MFAStatus)

	assert.Nil(t, records[1].LastLogin)
	assert.Equal(t, coredata.MFAStatusUnknown, records[1].MFAStatus)

	assert.Nil(t, records[2].LastLogin)
	assert.Equal(t, coredata.MFAStatusUnknown, records[2].MFAStatus)
}

func TestAzureSignInTime(t *testing.T) {
	t.Parallel()

	t.Run(
		"prefers last successful sign-in",
		func(t *testing.T) {
			t.Parallel()

			at, ok := azureSignInTime(
				azureSignInActivity{
					LastSuccessfulSignInDateTime: "2026-08-15T12:00:00Z",
					LastSignInDateTime:           "2026-07-01T00:00:00Z",
				},
			)
			require.True(t, ok)
			assert.True(t, at.Equal(azureAliceLastLogin))
		},
	)

	t.Run(
		"falls back to last sign-in",
		func(t *testing.T) {
			t.Parallel()

			at, ok := azureSignInTime(
				azureSignInActivity{
					LastSignInDateTime: "2026-07-01T00:00:00Z",
				},
			)
			require.True(t, ok)
			assert.True(t, at.Equal(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)))
		},
	)

	t.Run(
		"ignores zero and empty timestamps",
		func(t *testing.T) {
			t.Parallel()

			_, ok := azureSignInTime(
				azureSignInActivity{
					LastSuccessfulSignInDateTime: "0001-01-01T00:00:00Z",
				},
			)
			assert.False(t, ok)
		},
	)
}

func TestAzureUsersIDFilter(t *testing.T) {
	t.Parallel()

	assert.Equal(
		t,
		"id in ('"+vcrAzureAliceID+"')",
		azureUsersIDFilter([]string{vcrAzureAliceID}),
	)
	assert.Equal(
		t,
		"id in ('"+vcrAzureAliceID+"','"+vcrAzureBobID+"')",
		azureUsersIDFilter([]string{vcrAzureAliceID, vcrAzureBobID}),
	)
	assert.Equal(t, "id in ('O''Brien')", azureUsersIDFilter([]string{"O'Brien"}))
}

func TestAzureUsersSignInActivityURL(t *testing.T) {
	t.Parallel()

	for _, env := range cloudazure.Environments() {
		t.Run(
			env.String(),
			func(t *testing.T) {
				t.Parallel()

				session := cloudazure.NewSessionFromToken(
					vcrAzureSubscriptionID,
					vcrAzureAccessToken,
					cloudazure.WithEnvironment(env),
				)
				got, err := azureUsersSignInActivityURL(session.GraphBaseURL(), []string{vcrAzureAliceID})
				require.NoError(t, err)
				assert.True(t, strings.HasPrefix(got, session.GraphBaseURL()))
				assert.Contains(t, got, "/v1.0/users")
				assert.Contains(t, got, vcrAzureAliceID)
				assert.NotContains(t, got, "graph.microsoft.com/"+env.String())
			},
		)
	}
}

func TestEnrichAzureIdentities_FillsLastLoginAndMFA(t *testing.T) {
	t.Parallel()

	rec := newAzureRecorder(t, "testdata/azure_activity")
	session := newAzureTestSession(t, rec, cloudazure.EnvironmentPublic)
	records := azureTestRecords()

	err := enrichAzureIdentities(context.Background(), session, records)
	require.NoError(t, err)

	require.NotNil(t, records[0].LastLogin)
	assert.True(t, records[0].LastLogin.Equal(azureAliceLastLogin))
	assert.Equal(t, coredata.MFAStatusEnabled, records[0].MFAStatus)

	assert.Nil(t, records[1].LastLogin)
	assert.Equal(t, coredata.MFAStatusUnknown, records[1].MFAStatus)

	assert.Nil(t, records[2].LastLogin)
	assert.Equal(t, coredata.MFAStatusUnknown, records[2].MFAStatus)
}

func TestEnrichAzureIdentities_FallsBackToLastSignIn(t *testing.T) {
	t.Parallel()

	rec := newAzureRecorder(t, "testdata/azure_activity_fallback")
	session := newAzureTestSession(t, rec, cloudazure.EnvironmentPublic)
	records := azureTestRecords()

	err := enrichAzureIdentities(context.Background(), session, records)
	require.NoError(t, err)

	require.NotNil(t, records[0].LastLogin)
	assert.True(t, records[0].LastLogin.Equal(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)))
	assert.Equal(t, coredata.MFAStatusEnabled, records[0].MFAStatus)
}

func TestEnrichAzureIdentities_KeepsMFAWhenActivityDenied(t *testing.T) {
	t.Parallel()

	rec := newAzureRecorder(t, "testdata/azure_activity_denied")
	session := newAzureTestSession(t, rec, cloudazure.EnvironmentPublic)
	records := azureTestRecords()

	err := enrichAzureIdentities(context.Background(), session, records)
	require.Error(t, err)
	assert.True(t, cloudazure.As[cloudazure.ErrPermissionDenied](err))

	assert.Nil(t, records[0].LastLogin)
	assert.Equal(t, coredata.MFAStatusEnabled, records[0].MFAStatus)
}

func TestEnrichAzureIdentities_KeepsLastLoginWhenMFADenied(t *testing.T) {
	t.Parallel()

	rec := newAzureRecorder(t, "testdata/azure_activity_mfa_denied")
	session := newAzureTestSession(t, rec, cloudazure.EnvironmentPublic)
	records := azureTestRecords()

	err := enrichAzureIdentities(context.Background(), session, records)
	require.Error(t, err)
	assert.True(t, cloudazure.As[cloudazure.ErrPermissionDenied](err))

	require.NotNil(t, records[0].LastLogin)
	assert.True(t, records[0].LastLogin.Equal(azureAliceLastLogin))
	assert.Equal(t, coredata.MFAStatusUnknown, records[0].MFAStatus)
}

func TestEnrichAzureIdentities_DegradesWhenActivityNotFound(t *testing.T) {
	t.Parallel()

	rec := newAzureRecorder(t, "testdata/azure_activity_not_found")
	session := newAzureTestSession(t, rec, cloudazure.EnvironmentPublic)
	records := azureTestRecords()

	err := enrichAzureIdentities(context.Background(), session, records)
	require.Error(t, err)
	assert.True(t, cloudazure.As[cloudazure.ErrNotFound](err))
	assert.Nil(t, records[0].LastLogin)
}

func TestEnrichAzureIdentities_DegradesWhenUnlicensed(t *testing.T) {
	t.Parallel()

	rec := newAzureRecorder(t, "testdata/azure_activity_unlicensed")
	session := newAzureTestSession(t, rec, cloudazure.EnvironmentPublic)
	records := azureTestRecords()

	err := enrichAzureIdentities(context.Background(), session, records)
	require.Error(t, err)
	assert.True(t, azureGraphLicenceError(err))
	assert.Nil(t, records[0].LastLogin)
}

func TestFetchAzureActivity_FallsBackToBatch(t *testing.T) {
	t.Parallel()

	rec := newAzureRecorder(t, "testdata/azure_activity_batch")
	session := newAzureTestSession(t, rec, cloudazure.EnvironmentPublic)

	logins, err := fetchAzureActivity(context.Background(), session, azureTestRecords())
	require.NoError(t, err)
	require.Contains(t, logins, vcrAzureAliceID)
	assert.True(t, logins[vcrAzureAliceID].Equal(azureAliceLastLogin))
}

func TestFetchAzureActivity_SkipsGroupsAndServicePrincipals(t *testing.T) {
	t.Parallel()

	logins, err := fetchAzureActivity(
		context.Background(),
		cloudazure.NewSessionFromToken(
			vcrAzureSubscriptionID,
			vcrAzureAccessToken,
			cloudazure.WithHTTPClient(
				&http.Client{
					Transport: roundTripFunc(
						func(*http.Request) (*http.Response, error) {
							t.Fatal("fetchAzureActivity must not call the network for groups or service principals")
							return nil, nil
						},
					),
				},
			),
		),
		[]AccountRecord{
			{
				Email:       "eng@probo-azure.test",
				ExternalID:  vcrAzureEngID,
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				MFAStatus:   coredata.MFAStatusUnknown,
			},
			{
				ExternalID:  vcrAzureCIID,
				AccountType: coredata.AccessReviewEntryAccountTypeServiceAccount,
				AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
				MFAStatus:   coredata.MFAStatusUnknown,
			},
		},
	)
	require.NoError(t, err)
	assert.Empty(t, logins)
}

func TestEnrichAzureIdentities_FailsOnCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	rec := newAzureRecorder(t, "testdata/azure_activity")
	session := newAzureTestSession(t, rec, cloudazure.EnvironmentPublic)
	records := azureTestRecords()

	err := enrichAzureIdentities(ctx, session, records)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, records[0].LastLogin)
	assert.Equal(t, coredata.MFAStatusUnknown, records[0].MFAStatus)
}
