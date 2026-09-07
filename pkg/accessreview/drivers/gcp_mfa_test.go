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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
	"go.probo.inc/probo/pkg/coredata"
)

func TestFetchGCPMFA_ReadsEnrollment(t *testing.T) {
	t.Parallel()

	rec := newGCPRecorder(t, "testdata/gcp_mfa")
	session := newGCPTestSession(t, rec)
	records := gcpTestRecords()

	mfa, err := fetchGCPMFA(context.Background(), session, records)
	require.NoError(t, err)
	require.Equal(t, coredata.MFAStatusEnabled, mfa["alice@example.com"])
	assert.NotContains(t, mfa, "eng@example.com")
	assert.NotContains(t, mfa, "sa@my-project.iam.gserviceaccount.com")
}

func TestFetchGCPMFA_DegradesWhenDenied(t *testing.T) {
	t.Parallel()

	rec := newGCPRecorder(t, "testdata/gcp_mfa_denied")
	session := newGCPTestSession(t, rec)
	records := gcpTestRecords()

	mfa, err := fetchGCPMFA(context.Background(), session, records)
	require.Error(t, err)
	assert.True(t, cloudgcp.As[cloudgcp.ErrPermissionDenied](err))
	assert.Empty(t, mfa)
}

func TestFetchGCPMFA_SkipsGroupsAndServiceAccounts(t *testing.T) {
	t.Parallel()

	mfa, err := fetchGCPMFA(
		context.Background(),
		cloudgcp.NewSessionFromToken(
			vcrGCPProjectNumber,
			vcrGCPAccessToken,
			cloudgcp.WithHTTPClient(
				&http.Client{
					Transport: roundTripFunc(
						func(*http.Request) (*http.Response, error) {
							t.Fatal("fetchGCPMFA must not call the network for groups or service accounts")
							return nil, nil
						},
					),
				},
			),
		),
		[]AccountRecord{
			{
				Email:       "eng@example.com",
				ExternalID:  "group:eng@example.com",
				AccountType: coredata.AccessReviewEntryAccountTypeUser,
				MFAStatus:   coredata.MFAStatusUnknown,
			},
			{
				Email:       "sa@my-project.iam.gserviceaccount.com",
				ExternalID:  "111111111111",
				AccountType: coredata.AccessReviewEntryAccountTypeServiceAccount,
				MFAStatus:   coredata.MFAStatusUnknown,
			},
		},
	)
	require.NoError(t, err)
	assert.Empty(t, mfa)
}
