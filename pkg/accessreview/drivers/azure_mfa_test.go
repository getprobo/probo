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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cloudazure "go.probo.inc/probo/pkg/cloud/azure"
	"go.probo.inc/probo/pkg/coredata"
)

func azureMFARecords() []AccountRecord {
	return []AccountRecord{
		{
			Email:       "alice@probo-azure.test",
			ExternalID:  vcrAzureAliceID,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodSSO,
			MFAStatus:   coredata.MFAStatusUnknown,
		},
		{
			Email:       "bob@probo-azure.test",
			ExternalID:  vcrAzureBobID,
			AccountType: coredata.AccessReviewEntryAccountTypeUser,
			AuthMethod:  coredata.AccessReviewEntryAuthMethodSSO,
			MFAStatus:   coredata.MFAStatusUnknown,
		},
		{
			Email:       "dana@probo-azure.test",
			ExternalID:  vcrAzureDanaID,
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

func TestFetchAzureMFA_ReadsCapableFlag(t *testing.T) {
	t.Parallel()

	rec := newAzureRecorder(t, "testdata/azure_mfa")
	session := newAzureTestSession(t, rec, cloudazure.EnvironmentPublic)

	mfa, err := fetchAzureMFA(context.Background(), session, azureMFARecords())
	require.NoError(t, err)
	assert.Equal(t, coredata.MFAStatusEnabled, mfa[vcrAzureAliceID])
	assert.Equal(t, coredata.MFAStatusDisabled, mfa[vcrAzureBobID])
	assert.NotContains(t, mfa, vcrAzureDanaID)
	assert.NotContains(t, mfa, vcrAzureEngID)
	assert.NotContains(t, mfa, vcrAzureCIID)
}

func TestFetchAzureMFA_DegradesWhenDenied(t *testing.T) {
	t.Parallel()

	rec := newAzureRecorder(t, "testdata/azure_mfa_denied")
	session := newAzureTestSession(t, rec, cloudazure.EnvironmentPublic)

	mfa, err := fetchAzureMFA(context.Background(), session, azureMFARecords())
	require.Error(t, err)
	assert.True(t, cloudazure.As[cloudazure.ErrPermissionDenied](err))
	assert.Empty(t, mfa)
}

func TestFetchAzureMFA_DegradesWhenNotFound(t *testing.T) {
	t.Parallel()

	rec := newAzureRecorder(t, "testdata/azure_mfa_not_found")
	session := newAzureTestSession(t, rec, cloudazure.EnvironmentPublic)

	mfa, err := fetchAzureMFA(context.Background(), session, azureMFARecords())
	require.Error(t, err)
	assert.True(t, cloudazure.As[cloudazure.ErrNotFound](err))
	assert.Empty(t, mfa)
}

func TestFetchAzureMFA_DegradesWhenUnlicensed(t *testing.T) {
	t.Parallel()

	rec := newAzureRecorder(t, "testdata/azure_mfa_unlicensed")
	session := newAzureTestSession(t, rec, cloudazure.EnvironmentPublic)

	mfa, err := fetchAzureMFA(context.Background(), session, azureMFARecords())
	require.Error(t, err)
	assert.True(t, azureGraphLicenceError(err))
	assert.Empty(t, mfa)
}

func TestFetchAzureMFA_SkipsGroupsAndServicePrincipals(t *testing.T) {
	t.Parallel()

	mfa, err := fetchAzureMFA(
		context.Background(),
		cloudazure.NewSessionFromToken(
			vcrAzureSubscriptionID,
			vcrAzureAccessToken,
			cloudazure.WithHTTPClient(
				&http.Client{
					Transport: roundTripFunc(
						func(*http.Request) (*http.Response, error) {
							t.Fatal("fetchAzureMFA must not call the network for groups or service principals")
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
	assert.Empty(t, mfa)
}

func TestAzureUserRegistrationDetailsURL(t *testing.T) {
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
				got, err := azureUserRegistrationDetailsURL(session.GraphBaseURL())
				require.NoError(t, err)
				assert.True(t, strings.HasPrefix(got, session.GraphBaseURL()))
				assert.True(
					t,
					strings.HasSuffix(got, "/v1.0/reports/authenticationMethods/userRegistrationDetails"),
				)
				assert.NotContains(t, got, "graph.microsoft.com/"+env.String())
			},
		)
	}
}

func TestAzureRegistrationMFAStatus(t *testing.T) {
	t.Parallel()

	status, ok := azureRegistrationMFAStatus(azureUserRegistrationDetails{IsMFACapable: new(true)})
	assert.True(t, ok)
	assert.Equal(t, coredata.MFAStatusEnabled, status)

	status, ok = azureRegistrationMFAStatus(azureUserRegistrationDetails{IsMFACapable: new(false)})
	assert.True(t, ok)
	assert.Equal(t, coredata.MFAStatusDisabled, status)

	status, ok = azureRegistrationMFAStatus(azureUserRegistrationDetails{})
	assert.False(t, ok)
	assert.Equal(t, coredata.MFAStatusUnknown, status)
}
