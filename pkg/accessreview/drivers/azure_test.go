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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	cloudazure "go.probo.inc/probo/pkg/cloud/azure"
	"go.probo.inc/probo/pkg/coredata"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

const (
	vcrAzureSubscriptionID = "11111111-1111-4111-8111-111111111111"
	vcrAzureAccessToken    = "vcr-azure-access-token"
	vcrAzureAliceID        = "22222222-2222-4222-8222-222222222222"
	vcrAzureEngID          = "33333333-3333-4333-8333-333333333333"
	vcrAzureCIID           = "44444444-4444-4444-8444-444444444444"
	vcrAzureDeviceID       = "55555555-5555-4555-8555-555555555555"
)

func newAzureTestSession(
	t *testing.T,
	rec *recorder.Recorder,
	environment cloudazure.Environment,
) *cloudazure.Session {
	t.Helper()

	return cloudazure.NewSessionFromToken(
		vcrAzureSubscriptionID,
		vcrAzureAccessToken,
		cloudazure.WithHTTPClient(newVCRClient(rec, "")),
		cloudazure.WithEnvironment(environment),
	)
}

func TestAzureDriver(t *testing.T) {
	t.Parallel()

	rec := newAzureRecorder(t, "testdata/azure")
	session := newAzureTestSession(t, rec, cloudazure.EnvironmentPublic)

	records, err := NewAzureDriver(session, log.NewLogger(log.WithName("test"))).
		ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 4)

	byID := make(map[string]AccountRecord, len(records))
	for _, record := range records {
		byID[record.ExternalID] = record
	}

	alice := byID[vcrAzureAliceID]
	assert.Equal(t, "alice@probo-azure.test", alice.Email)
	assert.Equal(t, "Alice Chen", alice.FullName)
	assert.Equal(t, []string{"Owner"}, alice.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, alice.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodSSO, alice.AuthMethod)
	assert.Equal(t, coredata.MFAStatusUnknown, alice.MFAStatus)
	require.NotNil(t, alice.IsAdmin)
	assert.True(t, *alice.IsAdmin)
	require.NotNil(t, alice.Active)
	assert.True(t, *alice.Active)
	assert.Nil(t, alice.LastLogin)
	assert.Nil(t, alice.CreatedAt)

	eng := byID[vcrAzureEngID]
	assert.Equal(t, "eng@probo-azure.test", eng.Email)
	assert.Equal(t, "Engineering", eng.FullName)
	assert.Equal(t, []string{"Reader"}, eng.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, eng.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, eng.AuthMethod)
	assert.Nil(t, eng.IsAdmin)
	assert.Nil(t, eng.Active)

	ci := byID[vcrAzureCIID]
	assert.Equal(t, "CI Deploy", ci.FullName)
	assert.Empty(t, ci.Email)
	assert.Equal(t, []string{"Custom Admin"}, ci.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, ci.AccountType)
	assert.Nil(t, ci.IsAdmin)
	require.NotNil(t, ci.Active)
	assert.False(t, *ci.Active)

	device := byID[vcrAzureDeviceID]
	assert.Empty(t, device.Email)
	assert.Empty(t, device.FullName)
	assert.Equal(t, []string{"Contributor"}, device.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, device.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, device.AuthMethod)
	assert.Nil(t, device.IsAdmin)
	assert.Nil(t, device.Active)
}

func TestAzureDriver_Government(t *testing.T) {
	t.Parallel()

	rec := newAzureRecorder(t, "testdata/azure_government")
	session := newAzureTestSession(t, rec, cloudazure.EnvironmentGovernment)

	records, err := NewAzureDriver(session, log.NewLogger(log.WithName("test"))).
		ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, vcrAzureAliceID, records[0].ExternalID)
	assert.Equal(t, "alice@probo-azure.test", records[0].Email)
}

func TestAzureDriver_FailsWhenRoleAssignmentsDenied(t *testing.T) {
	t.Parallel()

	rec := newAzureRecorder(t, "testdata/azure_rbac_denied")
	session := newAzureTestSession(t, rec, cloudazure.EnvironmentPublic)

	_, err := NewAzureDriver(session, log.NewLogger(log.WithName("test"))).
		ListAccounts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot list azure role assignments")
}

func TestAzureDriver_GraphForbiddenDegrades(t *testing.T) {
	t.Parallel()

	rec := newAzureRecorder(t, "testdata/azure_graph_forbidden")
	session := newAzureTestSession(t, rec, cloudazure.EnvironmentPublic)

	records, err := NewAzureDriver(session, log.NewLogger(log.WithName("test"))).
		ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, vcrAzureAliceID, records[0].ExternalID)
	assert.Empty(t, records[0].Email)
	assert.Empty(t, records[0].FullName)
	assert.Equal(t, []string{"Owner"}, records[0].Roles)
	require.NotNil(t, records[0].IsAdmin)
	assert.True(t, *records[0].IsAdmin)
}

func TestAzureDriver_GraphNotFoundDegrades(t *testing.T) {
	t.Parallel()

	rec := newAzureRecorder(t, "testdata/azure_graph_not_found")
	session := newAzureTestSession(t, rec, cloudazure.EnvironmentPublic)

	records, err := NewAzureDriver(session, log.NewLogger(log.WithName("test"))).
		ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, vcrAzureAliceID, records[0].ExternalID)
	assert.Empty(t, records[0].Email)
	assert.Empty(t, records[0].FullName)
}

func TestAzureNameResolver_UsesDisplayName(t *testing.T) {
	t.Parallel()

	rec := newAzureRecorder(t, "testdata/azure_subscription_name")
	session := newAzureTestSession(t, rec, cloudazure.EnvironmentPublic)

	name, err := NewAzureNameResolver(
		session,
		log.NewLogger(log.WithName("test")),
	).ResolveInstanceName(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Contoso Production", name)
}

func TestAzureNameResolver_FallsBackToSubscriptionID(t *testing.T) {
	t.Parallel()

	rec := newAzureRecorder(t, "testdata/azure_subscription_id")
	session := newAzureTestSession(t, rec, cloudazure.EnvironmentPublic)

	name, err := NewAzureNameResolver(
		session,
		log.NewLogger(log.WithName("test")),
	).ResolveInstanceName(context.Background())
	require.NoError(t, err)
	assert.Equal(t, vcrAzureSubscriptionID, name)
}

func TestAzureSources_DoNotHardcodePublicGraphHost(t *testing.T) {
	t.Parallel()

	matches, err := filepath.Glob("azure*.go")
	require.NoError(t, err)

	for _, path := range matches {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}

		t.Run(
			filepath.Base(path),
			func(t *testing.T) {
				t.Parallel()

				data, err := os.ReadFile(path)
				require.NoError(t, err)
				assert.NotContains(t, string(data), "graph.microsoft.com")
			},
		)
	}
}
