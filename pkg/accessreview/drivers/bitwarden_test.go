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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
)

// TestBitwardenDriver replays a recording of a Teams organization on the EU
// cloud: a confirmed owner, an invited admin, a revoked user and an invited
// Custom member allowed to manage users. Identities are rewritten by
// sanitizeBitwarden; enums and the envelope are Bitwarden's own.
func TestBitwardenDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/bitwarden", "BITWARDEN_TOKEN", sanitizeBitwarden)
	client := newVCRClient(rec, bearerAuth(os.Getenv("BITWARDEN_TOKEN")))

	baseURL, err := BitwardenAPIBaseURL("EU")
	require.NoError(t, err)

	records, err := NewBitwardenDriver(client, baseURL).ListAccounts(context.Background())
	require.NoError(t, err)

	// While recording, the records are the live roster: the sanitizer only
	// rewrites what is saved, and a failed assertion would print them.
	if os.Getenv("BITWARDEN_TOKEN") != "" {
		return
	}

	require.Len(t, records, 4)

	byEmail := map[string]AccountRecord{}
	for _, r := range records {
		byEmail[r.Email] = r
	}

	require.Len(t, byEmail, 4)

	for _, r := range records {
		assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, r.AccountType)
		assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, r.AuthMethod)
		assert.NotEmpty(t, r.ExternalID)
	}

	admin := byEmail["member1@example.com"]
	assert.Equal(t, []string{"Admin"}, admin.Roles)
	assert.Equal(t, new(true), admin.IsAdmin)
	assert.Equal(t, new(false), admin.Active)
	assert.Empty(t, admin.FullName, "an invitee has no name yet")
	assert.Equal(t, coredata.MFAStatusUnknown, admin.MFAStatus, "an invitee has no account to hold 2FA")

	revoked := byEmail["member2@example.com"]
	assert.Equal(t, []string{"User"}, revoked.Roles)
	assert.Equal(t, new(false), revoked.IsAdmin)
	assert.Equal(t, new(false), revoked.Active)

	custom := byEmail["member3@example.com"]
	assert.Equal(t, []string{"Custom"}, custom.Roles)
	assert.Equal(t, new(true), custom.IsAdmin, "manageUsers opens the Admin Console")
	assert.Equal(t, new(false), custom.Active)

	owner := byEmail["member4@example.com"]
	assert.Equal(t, []string{"Owner"}, owner.Roles)
	assert.Equal(t, new(true), owner.IsAdmin)
	assert.Equal(t, new(true), owner.Active)
	assert.Equal(t, "Member 4", owner.FullName)
	assert.Equal(t, coredata.MFAStatusDisabled, owner.MFAStatus)
}

// TestBitwardenMemberRecord covers what the recording cannot: the remaining
// statuses, a Custom member with and without permissions, and values Bitwarden
// may add later.
func TestBitwardenMemberRecord(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		member  bitwardenMember
		roles   []string
		active  *bool
		isAdmin *bool
		mfa     coredata.MFAStatus
	}{
		{
			name:    "confirmed user with 2FA",
			member:  bitwardenMember{UserID: "u", Status: bitwardenStatusConfirmed, Type: bitwardenTypeUser, TwoFactorEnabled: true},
			roles:   []string{"User"},
			active:  new(true),
			isAdmin: new(false),
			mfa:     coredata.MFAStatusEnabled,
		},
		{
			name:    "accepted but not yet confirmed",
			member:  bitwardenMember{UserID: "u", Status: bitwardenStatusAccepted, Type: bitwardenTypeUser},
			roles:   []string{"User"},
			active:  new(false),
			isAdmin: new(false),
			mfa:     coredata.MFAStatusDisabled,
		},
		{
			name:    "staged by directory sync",
			member:  bitwardenMember{Status: bitwardenStatusStaged, Type: bitwardenTypeUser},
			roles:   []string{"User"},
			active:  new(false),
			isAdmin: new(false),
			mfa:     coredata.MFAStatusUnknown,
		},
		{
			name: "custom with no permission granted",
			member: bitwardenMember{
				UserID:      "u",
				Status:      bitwardenStatusConfirmed,
				Type:        bitwardenTypeCustom,
				Permissions: map[string]bool{"manageUsers": false, "accessReports": false},
			},
			roles:   []string{"Custom"},
			active:  new(true),
			isAdmin: new(false),
			mfa:     coredata.MFAStatusDisabled,
		},
		{
			name: "custom allowed only to create collections",
			member: bitwardenMember{
				UserID:      "u",
				Status:      bitwardenStatusConfirmed,
				Type:        bitwardenTypeCustom,
				Permissions: map[string]bool{"createNewCollections": true},
			},
			roles:   []string{"Custom"},
			active:  new(true),
			isAdmin: new(false),
			mfa:     coredata.MFAStatusDisabled,
		},
		{
			name:    "custom without a permissions object",
			member:  bitwardenMember{UserID: "u", Status: bitwardenStatusConfirmed, Type: bitwardenTypeCustom},
			roles:   []string{"Custom"},
			active:  new(true),
			isAdmin: nil,
			mfa:     coredata.MFAStatusDisabled,
		},
		{
			name:    "unknown type and status",
			member:  bitwardenMember{UserID: "u", Status: 9, Type: 9},
			roles:   []string{},
			active:  nil,
			isAdmin: nil,
			mfa:     coredata.MFAStatusDisabled,
		},
	} {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				r := bitwardenMemberRecord(tc.member)
				assert.Equal(t, tc.roles, r.Roles)
				assert.Equal(t, tc.active, r.Active)
				assert.Equal(t, tc.isAdmin, r.IsAdmin)
				assert.Equal(t, tc.mfa, r.MFAStatus)
			},
		)
	}
}

func TestBitwardenDriverRefusesPaginatedRoster(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"m","email":"a@example.com","status":2,"type":2}],"continuationToken":"next"}`))
			},
		),
	)
	t.Cleanup(srv.Close)

	_, err := NewBitwardenDriver(srv.Client(), srv.URL).ListAccounts(context.Background())
	require.Error(t, err)
}

func TestBitwardenHosts(t *testing.T) {
	t.Parallel()

	for region, want := range map[string][2]string{
		"US": {"https://api.bitwarden.com", "https://identity.bitwarden.com/connect/token"},
		"EU": {"https://api.bitwarden.eu", "https://identity.bitwarden.eu/connect/token"},
	} {
		apiBase, err := BitwardenAPIBaseURL(region)
		require.NoError(t, err)
		assert.Equal(t, want[0], apiBase)

		tokenURL, err := BitwardenTokenURL(region)
		require.NoError(t, err)
		assert.Equal(t, want[1], tokenURL)
	}

	// A stored region is never guessed: the key only exists on one cloud.
	for _, region := range []string{"", "us", "CA"} {
		_, err := BitwardenAPIBaseURL(region)
		assert.Error(t, err)

		_, err = BitwardenTokenURL(region)
		assert.Error(t, err)
	}
}

func TestSanitizeBitwarden(t *testing.T) {
	t.Parallel()

	interaction := func(body string) *cassette.Interaction {
		i := &cassette.Interaction{}
		i.Request.Method = http.MethodGet
		i.Request.URL = "https://api.bitwarden.eu/public/members"
		i.Response.Code = http.StatusOK
		i.Response.Body = body

		return i
	}

	t.Run(
		"rewrites every identity",
		func(t *testing.T) {
			t.Parallel()

			i := interaction(
				`{"object":"list","data":[{"object":"member","id":"real-member","userId":"real-user",` +
					`"name":"Real Person","email":"real.person@corp.example.org","externalId":"real-external",` +
					`"ssoExternalId":"real-sso","status":2,"type":0,"collections":[{"id":"real-collection","readOnly":true}]}]}`,
			)

			require.NoError(t, sanitizeBitwarden(i))

			for _, leaked := range []string{"real-", "Real Person", "corp.example.org"} {
				assert.NotContains(t, i.Response.Body, leaked)
			}

			assert.Contains(t, i.Response.Body, `"status":2`)
			assert.Contains(t, i.Response.Body, `"readOnly":true`)
		},
	)

	for name, body := range map[string]string{
		"an unknown member field":     `{"object":"list","data":[{"id":"a","email":"a@example.com","groups":["real-group"]}]}`,
		"an unknown collection field": `{"object":"list","data":[{"id":"a","email":"a@example.com","collections":[{"id":"c","name":"Real"}]}]}`,
		"a continuation token":        `{"object":"list","data":[],"continuationToken":"abc"}`,
	} {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				assert.Error(t, sanitizeBitwarden(interaction(body)))
			},
		)
	}
}

// sanitizeBitwarden rewrites every identifier and identity in the recorded
// roster. Members are numbered in response order; fields the driver does not
// read are kept so the cassette still shows Bitwarden's full shape, and a field
// it does not know is refused.
func sanitizeBitwarden(i *cassette.Interaction) error {
	if i.Response.Code != http.StatusOK {
		return fmt.Errorf("refusing to sanitize bitwarden response with status %d", i.Response.Code)
	}

	// Only the roster's shape is known here; anything else could carry
	// identities this sanitizer does not rewrite.
	parsed, err := url.Parse(i.Request.URL)
	if err != nil {
		return fmt.Errorf("cannot parse recorded bitwarden URL: %w", err)
	}

	if i.Request.Method != http.MethodGet || parsed.Path != "/public/members" {
		return fmt.Errorf("refusing to sanitize an unrecognised bitwarden interaction")
	}

	var body map[string]any
	if err := json.Unmarshal([]byte(i.Response.Body), &body); err != nil {
		return fmt.Errorf("cannot decode recorded bitwarden response: %w", err)
	}

	if err := bitwardenRejectUnknownFields(body, "envelope", "object", "data"); err != nil {
		return err
	}

	members, ok := body["data"].([]any)
	if !ok {
		return fmt.Errorf("recorded bitwarden response has no member list")
	}

	for n, raw := range members {
		k := n + 1

		m, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("recorded bitwarden member %d is not an object", k)
		}

		if err := bitwardenRejectUnknownFields(
			m,
			"member",
			"collections",
			"email",
			"externalId",
			"id",
			"name",
			"object",
			"permissions",
			"resetPasswordEnrolled",
			"ssoExternalId",
			"status",
			"twoFactorEnabled",
			"type",
			"userId",
		); err != nil {
			return err
		}

		m["id"] = fmt.Sprintf("00000000-0000-0000-0000-%012d", k)
		m["email"] = fmt.Sprintf("member%d@example.com", k)

		if m["userId"] != nil {
			m["userId"] = fmt.Sprintf("00000000-0000-0000-0001-%012d", k)
		}

		if m["name"] != nil {
			m["name"] = fmt.Sprintf("Member %d", k)
		}

		// Directory sync and SSO fill these with the customer's own
		// identifiers, often email addresses.
		for _, field := range []string{"externalId", "ssoExternalId"} {
			if m[field] != nil {
				m[field] = fmt.Sprintf("external-%d", k)
			}
		}

		collections, _ := m["collections"].([]any)
		for c, raw := range collections {
			collection, ok := raw.(map[string]any)
			if !ok {
				return fmt.Errorf("recorded bitwarden collection of member %d is not an object", k)
			}

			if err := bitwardenRejectUnknownFields(collection, "collection", "hidePasswords", "id", "manage", "readOnly"); err != nil {
				return err
			}

			collection["id"] = fmt.Sprintf("00000000-0000-0000-0002-%012d", c+1)
		}
	}

	encoded, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("cannot encode sanitized bitwarden response: %w", err)
	}

	replaceCassetteBody(i, string(encoded))

	return nil
}

// bitwardenRejectUnknownFields refuses a field the sanitizer does not classify,
// so an identity Bitwarden adds later is never recorded verbatim.
func bitwardenRejectUnknownFields(object map[string]any, what string, known ...string) error {
	for field := range object {
		if !slices.Contains(known, field) {
			return fmt.Errorf("refusing to record unclassified field %q in bitwarden %s", field, what)
		}
	}

	return nil
}
