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
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

// TestRetoolDriver replays a hand-authored cassette.
//
// It is hand-authored rather than recorded because listing users needs an API
// token carrying the users:read scope, and Retool refuses to widen a token's
// scopes through the API — GET /access_tokens answers "this API endpoint is
// not available on your plan", so the scope can only be granted in the web
// console. The body is built from Retool's own published OpenAPI
// (https://api.retool.com/api/v2/spec): every field, type and enum value is
// taken from the User schema there. Re-record it against a scoped token when
// one exists — an authored fixture can only prove the decoder matches its
// author's reading of the spec, which is how a hand-edited cassette once hid
// a real bug here for two months.
//
// The cassette carries two pages so the next_token/has_more loop is exercised
// rather than only its terminal case. The second page also holds every
// nullable field at once, which is where a decoder tends to break.
func TestRetoolDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/retool", "RETOOL_API_KEY")
	// The key is only consulted when RETOOL_API_KEY is set, which is also what
	// puts the recorder in record mode: passing "" here would make a re-record
	// capture an authentication failure instead of the roster.
	client := newVCRClient(rec, os.Getenv("RETOOL_API_KEY"))

	driver := NewRetoolDriver(client, "https://api.retool.com/api/v2")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	// Two records means the driver forwarded next_token and read page two;
	// stopping early would return one.
	require.Len(t, records, 2)

	// An active builder who administers the organization and has 2FA on.
	admin := records[0]
	assert.Equal(t, "01920000-0000-7000-8000-000000000001", admin.ExternalID)
	assert.Equal(t, "admin@example.com", admin.Email)
	assert.Equal(t, "Ada Admin", admin.FullName)
	assert.Equal(t, []string{"Admin", "Builder"}, admin.Roles)
	assert.Equal(t, new(true), admin.IsAdmin)
	require.NotNil(t, admin.Active)
	assert.True(t, *admin.Active)
	// Retool reports the flag for every user, so the status is never unknown.
	assert.Equal(t, coredata.MFAStatusEnabled, admin.MFAStatus)
	require.NotNil(t, admin.LastLogin)
	assert.Equal(t, time.Date(2026, 9, 1, 8, 30, 0, 0, time.UTC), admin.LastLogin.UTC())
	require.NotNil(t, admin.CreatedAt)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, admin.AccountType)

	// A deactivated external embed user who never signed in and has no name:
	// the nullable fields the schema allows, all at once.
	external := records[1]
	assert.Equal(t, "01920000-0000-7000-8000-000000000002", external.ExternalID)
	assert.Equal(t, "viewer@example.com", external.Email)
	// No first or last name falls back to the email rather than reviewing blank.
	assert.Equal(t, "viewer@example.com", external.FullName)
	assert.Equal(t, []string{"External User", "Embed"}, external.Roles)
	assert.Equal(t, new(false), external.IsAdmin)
	require.NotNil(t, external.Active)
	assert.False(t, *external.Active)
	assert.Equal(t, coredata.MFAStatusDisabled, external.MFAStatus)
	// A null last_active is no signal, not the zero time.
	assert.Nil(t, external.LastLogin)
}

func TestRetoolRoles(t *testing.T) {
	t.Parallel()

	builder := "builder"
	internal := "internalUser"
	external := "externalUser"
	unknown := "seatOfTheFuture"

	// Administering, the seat, and the account kind are three separate things
	// a reviewer acts on differently, so each is its own entry.
	assert.Equal(
		t,
		[]string{"Admin", "Builder"},
		retoolRoles(retoolUser{IsAdmin: true, SeatType: &builder, UserType: retoolUserTypeDefault}),
	)
	assert.Equal(
		t,
		[]string{"Internal User"},
		retoolRoles(retoolUser{SeatType: &internal, UserType: retoolUserTypeDefault}),
	)
	assert.Equal(
		t,
		[]string{"External User", "Embed"},
		retoolRoles(retoolUser{SeatType: &external, UserType: retoolUserTypeEmbed}),
	)
	assert.Equal(
		t,
		[]string{"Mobile"},
		retoolRoles(retoolUser{UserType: retoolUserTypeMobile}),
	)
	// Unrecognised future values are preserved rather than dropped.
	assert.Equal(
		t,
		[]string{"seatOfTheFuture", "kiosk"},
		retoolRoles(retoolUser{SeatType: &unknown, UserType: "kiosk"}),
	)
	// A seatless default user has no role, and never a nil slice.
	assert.Equal(t, []string{}, retoolRoles(retoolUser{UserType: retoolUserTypeDefault}))
	assert.Equal(t, []string{}, retoolRoles(retoolUser{}))
}

func TestRetoolMFAStatus(t *testing.T) {
	t.Parallel()

	assert.Equal(t, coredata.MFAStatusEnabled, retoolMFAStatus(retoolUser{TwoFactorAuthEnabled: true}))
	// Retool reports the flag for every user, so false means disabled rather
	// than unknown — this is one of the few rosters that can say so.
	assert.Equal(t, coredata.MFAStatusDisabled, retoolMFAStatus(retoolUser{}))
}

func TestRetoolFullName(t *testing.T) {
	t.Parallel()

	first := "Ada"
	last := "Lovelace"
	blank := "   "

	assert.Equal(t, "Ada Lovelace", retoolFullName(retoolUser{FirstName: &first, LastName: &last}, "x@example.com"))
	assert.Equal(t, "Ada", retoolFullName(retoolUser{FirstName: &first}, "x@example.com"))
	assert.Equal(t, "Lovelace", retoolFullName(retoolUser{LastName: &last}, "x@example.com"))
	// Both nullable, which the schema allows: fall back to the email.
	assert.Equal(t, "x@example.com", retoolFullName(retoolUser{}, "x@example.com"))
	assert.Equal(t, "x@example.com", retoolFullName(retoolUser{FirstName: &blank, LastName: &blank}, "x@example.com"))
}

// TestRetoolDriverRefusesNextPageWithoutToken covers a page that advertises
// more users without saying where to resume. Returning what we have would be a
// short roster with no error, and a member missing from a campaign is reviewed
// by nobody.
func TestRetoolDriverRefusesNextPageWithoutToken(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":[{"id":"u1","legacy_id":1,"email":"one@example.com",
          "active":true,"created_at":"2026-01-01T00:00:00.000Z","last_active":null,
          "first_name":null,"last_name":null,"metadata":{},"is_admin":false,
          "user_type":"default","seat_type":null,"two_factor_auth_enabled":false}],
          "total_count":2,"next_token":null,"has_more":true}`))
	}))
	defer server.Close()

	driver := NewRetoolDriver(server.Client(), server.URL)

	_, err := driver.ListAccounts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no token")
}
