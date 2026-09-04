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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
	"go.probo.inc/probo/pkg/coredata"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

const (
	vcrGCPProjectNumber = "123456789012"
	vcrGCPAccessToken   = "vcr-gcp-access-token"
	vcrGCPServiceEmail  = "probo-audit@my-project.iam.gserviceaccount.com"
)

func newGCPTestSession(t *testing.T, rec *recorder.Recorder) *cloudgcp.Session {
	t.Helper()

	return cloudgcp.NewSessionFromToken(
		vcrGCPProjectNumber,
		vcrGCPAccessToken,
		cloudgcp.WithHTTPClient(newVCRClient(rec, "")),
	)
}

func TestGCPDriver(t *testing.T) {
	t.Parallel()

	rec := newGCPRecorder(t, "testdata/gcp")
	session := newGCPTestSession(t, rec)

	records, err := NewGCPDriver(session).
		ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 5)

	byID := make(map[string]AccountRecord, len(records))
	for _, record := range records {
		byID[record.ExternalID] = record
	}

	alice := byID["user:alice@example.com"]
	assert.Equal(t, "alice@example.com", alice.Email)
	assert.Equal(t, []string{gcpRoleOwner}, alice.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, alice.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodSSO, alice.AuthMethod)
	assert.Equal(t, coredata.MFAStatusUnknown, alice.MFAStatus)
	require.NotNil(t, alice.IsAdmin)
	assert.True(t, *alice.IsAdmin)
	assert.Nil(t, alice.Active)
	assert.Nil(t, alice.LastLogin)

	eng := byID["group:eng@example.com"]
	assert.Equal(t, "eng@example.com", eng.Email)
	assert.Equal(t, []string{"roles/viewer"}, eng.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, eng.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, eng.AuthMethod)
	assert.Nil(t, eng.IsAdmin)
	assert.Nil(t, eng.Active)

	ci := byID["111111111111111111111"]
	assert.Equal(t, "ci@my-project.iam.gserviceaccount.com", ci.Email)
	assert.Equal(t, "CI Deploy", ci.FullName)
	assert.Equal(t, []string{"roles/editor"}, ci.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, ci.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodAPIKey, ci.AuthMethod)
	assert.Nil(t, ci.IsAdmin)
	require.NotNil(t, ci.Active)
	assert.True(t, *ci.Active)

	unbound := byID["222222222222222222222"]
	assert.Equal(t, "unbound@my-project.iam.gserviceaccount.com", unbound.Email)
	assert.Empty(t, unbound.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, unbound.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, unbound.AuthMethod)
	require.NotNil(t, unbound.Active)
	assert.True(t, *unbound.Active)

	disabled := byID["333333333333333333333"]
	assert.Equal(t, "disabled-bot@my-project.iam.gserviceaccount.com", disabled.Email)
	assert.Equal(t, []string{"projects/my-project/roles/customAdmin"}, disabled.Roles)
	assert.Nil(t, disabled.IsAdmin)
	require.NotNil(t, disabled.Active)
	assert.False(t, *disabled.Active)
}

func TestGCPDriver_FailsWhenServiceAccountsDenied(t *testing.T) {
	t.Parallel()

	rec := newGCPRecorder(t, "testdata/gcp_service_accounts_denied")
	session := newGCPTestSession(t, rec)

	_, err := NewGCPDriver(session).
		ListAccounts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot list service accounts of the gcp project")
}

func TestGCPDriver_FailsWhenIAMPolicyDenied(t *testing.T) {
	t.Parallel()

	rec := newGCPRecorder(t, "testdata/gcp_iam_denied")
	session := newGCPTestSession(t, rec)

	_, err := NewGCPDriver(session).
		ListAccounts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot get gcp project iam policy")
}

func TestGCPNameResolver_UsesDisplayName(t *testing.T) {
	t.Parallel()

	rec := newGCPRecorder(t, "testdata/gcp_project_name")
	session := newGCPTestSession(t, rec)

	name, err := NewGCPNameResolver(
		session,
		vcrGCPServiceEmail,
		log.NewLogger(log.WithName("test")),
	).ResolveInstanceName(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Acme Prod", name)
}

func TestGCPNameResolver_UsesProjectIDWhenDisplayNameEmpty(t *testing.T) {
	t.Parallel()

	rec := newGCPRecorder(t, "testdata/gcp_project_id_no_display_name")
	session := newGCPTestSession(t, rec)

	name, err := NewGCPNameResolver(
		session,
		vcrGCPServiceEmail,
		log.NewLogger(log.WithName("test")),
	).ResolveInstanceName(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "connected-project", name)
}

func TestGCPNameResolver_FallsBackToServiceAccountProjectID(t *testing.T) {
	t.Parallel()

	rec := newGCPRecorder(t, "testdata/gcp_project_id")
	session := newGCPTestSession(t, rec)

	name, err := NewGCPNameResolver(
		session,
		vcrGCPServiceEmail,
		log.NewLogger(log.WithName("test")),
	).ResolveInstanceName(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "my-project", name)
}

func TestGCPNameResolver_FallsBackToProjectNumber(t *testing.T) {
	t.Parallel()

	rec := newGCPRecorder(t, "testdata/gcp_project_number")
	session := newGCPTestSession(t, rec)

	name, err := NewGCPNameResolver(
		session,
		"",
		log.NewLogger(log.WithName("test")),
	).ResolveInstanceName(context.Background())
	require.NoError(t, err)
	assert.Equal(t, vcrGCPProjectNumber, name)
}
