// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package drivers

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestSendGridDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/sendgrid", "SENDGRID_API_KEY")
	client := newVCRClient(rec, bearerAuth(os.Getenv("SENDGRID_API_KEY")))
	driver := NewSendGridDriver(client)

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)

	// Recorded against a live SendGrid account that has only the owner. The
	// list endpoint carries no scopes, so the driver fetches the teammate
	// detail to read them.
	owner := records[0]
	assert.Equal(t, "owner@example.com", owner.Email)
	assert.Empty(t, owner.FullName)
	assert.Equal(t, "Owner", owner.Role)
	assert.True(t, owner.IsAdmin)
	assert.Equal(t, "owner@example.com", owner.ExternalID)
	assert.Equal(t, coredata.AccessEntryAccountTypeUser, owner.AccountType)
	// The owner is a full-access user whose scope catalog contains BOTH
	// 2fa_exempt and 2fa_required, so the MFA signal is ambiguous and the
	// driver reports Unknown rather than guessing from scope ordering.
	assert.Equal(t, coredata.MFAStatusUnknown, owner.MFAStatus)
}

func TestSendGridRole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		userType string
		isAdmin  bool
		want     string
	}{
		{name: "owner", userType: "owner", isAdmin: true, want: "Owner"},
		{name: "admin", userType: "admin", isAdmin: true, want: "Admin"},
		{name: "teammate", userType: "teammate", isAdmin: false, want: "Teammate"},
		{name: "empty admin", userType: "", isAdmin: true, want: "Admin"},
		{name: "empty teammate", userType: "", isAdmin: false, want: "Teammate"},
		{name: "unknown", userType: "custom-role", isAdmin: false, want: "custom-role"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, sendGridRole(tt.userType, tt.isAdmin))
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
