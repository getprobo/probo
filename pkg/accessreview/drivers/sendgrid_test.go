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
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
)

func TestSendGridDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/sendgrid", "SENDGRID_API_KEY")
	client := newVCRClient(rec, bearerAuth(os.Getenv("SENDGRID_API_KEY")))
	driver := NewSendGridDriver(client, log.NewLogger(log.WithName("test")))

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	// Two records: the owner + a restricted teammate. A third list row with an
	// empty email is skipped.
	require.Len(t, records, 2)

	// The owner record is the real recording. The list endpoint carries no
	// scopes, so the driver fetches the teammate detail to read them.
	owner := records[0]
	assert.Equal(t, "owner@example.com", owner.Email)
	assert.Empty(t, owner.FullName)
	assert.Equal(t, []string{"Owner"}, owner.Roles)
	assert.True(t, owner.IsAdmin)
	assert.Equal(t, "owner@example.com", owner.ExternalID)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, owner.AccountType)
	// is_sso=false on the owner -> authenticates with SendGrid credentials.
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodPassword, owner.AuthMethod)
	// The owner is a full-access user whose scope catalog contains BOTH
	// 2fa_exempt and 2fa_required, so the MFA signal is ambiguous and the
	// driver reports Unknown rather than guessing from scope ordering.
	assert.Equal(t, coredata.MFAStatusUnknown, owner.MFAStatus)

	// A restricted teammate, synthetic (the trial account has only the owner)
	// but modelled on the real detail shape: a BARE object whose scopes carry
	// a single 2fa flag. This makes the N+1 detail fetch load-bearing — an
	// Enabled MFA here is reachable ONLY by correctly decoding the detail
	// response, so it guards against the {"result":...}-envelope regression.
	teammate := records[1]
	assert.Equal(t, "taylor@example.com", teammate.Email)
	assert.Equal(t, "Taylor Teammate", teammate.FullName)
	assert.Equal(t, []string{"Teammate"}, teammate.Roles)
	assert.False(t, teammate.IsAdmin)
	// Non-unified teammate: username is a handle distinct from the email.
	assert.Equal(t, "taylor-teammate", teammate.ExternalID)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodSSO, teammate.AuthMethod)
	assert.Equal(t, coredata.MFAStatusEnabled, teammate.MFAStatus)
}

func TestSendGridRoles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		userType string
		isAdmin  bool
		want     []string
	}{
		{name: "owner", userType: "owner", isAdmin: true, want: []string{"Owner"}},
		{name: "admin", userType: "admin", isAdmin: true, want: []string{"Admin"}},
		{name: "teammate", userType: "teammate", isAdmin: false, want: []string{"Teammate"}},
		{name: "empty admin", userType: "", isAdmin: true, want: []string{"Admin"}},
		{name: "empty teammate", userType: "", isAdmin: false, want: []string{"Teammate"}},
		{name: "unknown", userType: "custom-role", isAdmin: false, want: []string{"custom-role"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, sendGridRoles(tt.userType, tt.isAdmin))
		})
	}
}

func TestSendGridResponseItems(t *testing.T) {
	t.Parallel()

	t.Run("prefers result", func(t *testing.T) {
		t.Parallel()

		items := sendGridResponseItems(&sendGridTeammatesResponse{
			Result: []sendGridTeammate{
				{Email: "owner@example.com"},
			},
			Results: []sendGridTeammate{
				{Email: "fallback@example.com"},
			},
		})

		require.Len(t, items, 1)
		assert.Equal(t, "owner@example.com", items[0].Email)
	})

	t.Run("falls back to results", func(t *testing.T) {
		t.Parallel()

		items := sendGridResponseItems(&sendGridTeammatesResponse{
			Results: []sendGridTeammate{
				{Email: "fallback@example.com"},
			},
		})

		require.Len(t, items, 1)
		assert.Equal(t, "fallback@example.com", items[0].Email)
	})
}

func TestSendGridMFAStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		scopes []string
		want   coredata.MFAStatus
	}{
		{name: "required", scopes: []string{"mail.send", "2fa_required"}, want: coredata.MFAStatusEnabled},
		{name: "exempt", scopes: []string{"mail.send", "2fa_exempt"}, want: coredata.MFAStatusDisabled},
		{name: "both is ambiguous", scopes: []string{"2fa_exempt", "2fa_required", "mail.send"}, want: coredata.MFAStatusUnknown},
		{name: "neither", scopes: []string{"mail.send"}, want: coredata.MFAStatusUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, sendGridMFAStatus(tt.scopes))
		})
	}
}
