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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestScalewayDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/scaleway", "SCALEWAY_API_KEY")
	// Scaleway authenticates via the X-Auth-Token header, not Authorization.
	client := newVCRClientWithHeader(rec, "X-Auth-Token", os.Getenv("SCALEWAY_API_KEY"))

	driver := NewScalewayDriver(client, "11111111-2222-3333-4444-555555555555")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 2)

	owner := records[0]
	assert.Equal(t, "8a3f1b2c-9d4e-4a5f-8b6c-1d2e3f4a5b6c", owner.ExternalID)
	assert.Equal(t, "alice.martin@example.com", owner.Email)
	assert.Equal(t, "Alice Martin", owner.FullName)
	assert.Equal(t, []string{"Owner"}, owner.Roles)
	assert.True(t, owner.IsAdmin)
	require.NotNil(t, owner.Active)
	assert.True(t, *owner.Active)
	assert.Equal(t, coredata.MFAStatusEnabled, owner.MFAStatus)
	assert.NotNil(t, owner.CreatedAt)
	assert.NotNil(t, owner.LastLogin)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, owner.AccountType)

	// Owner is the only admin; an invitation-pending member is inactive with no
	// last-login timestamp.
	member := records[1]
	assert.Equal(t, "c7e2a9d4-5f6b-4c3a-9e8d-2b1c0a9f8e7d", member.ExternalID)
	assert.Equal(t, "bob.dupont@example.com", member.Email)
	assert.Equal(t, []string{"Member"}, member.Roles)
	assert.False(t, member.IsAdmin)
	require.NotNil(t, member.Active)
	assert.False(t, *member.Active)
	assert.Equal(t, coredata.MFAStatusDisabled, member.MFAStatus)
	assert.Nil(t, member.LastLogin)
}

func TestScalewayMFAStatus(t *testing.T) {
	t.Parallel()

	enabled := true
	disabled := false

	// The always-present mfa boolean is the fallback; the newer
	// two_factor_enabled pointer wins when set.
	assert.Equal(t, coredata.MFAStatusEnabled, scalewayMFAStatus(scalewayUser{MFA: true}))
	assert.Equal(t, coredata.MFAStatusDisabled, scalewayMFAStatus(scalewayUser{MFA: false}))
	assert.Equal(t, coredata.MFAStatusEnabled, scalewayMFAStatus(scalewayUser{MFA: false, TwoFactorEnabled: &enabled}))
	assert.Equal(t, coredata.MFAStatusDisabled, scalewayMFAStatus(scalewayUser{MFA: true, TwoFactorEnabled: &disabled}))
}

func TestScalewayActive(t *testing.T) {
	t.Parallel()

	mustBool := func(t *testing.T, want bool, got *bool) {
		t.Helper()
		require.NotNil(t, got)
		assert.Equal(t, want, *got)
	}

	mustBool(t, true, scalewayActive("activated", false))
	mustBool(t, false, scalewayActive("invitation_pending", false))
	// A locked account is inactive even when its status is "activated".
	mustBool(t, false, scalewayActive("activated", true))
	assert.Nil(t, scalewayActive("", false))
	assert.Nil(t, scalewayActive("unknown_status", false))
}
