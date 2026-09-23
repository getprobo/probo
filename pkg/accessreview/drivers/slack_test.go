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
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
)

// sanitizeSlackUsers replaces member identity in a recorded users.list page
// with synthetic values. The profile is replaced wholesale: it carries phone
// numbers, avatars and status text, none of which the driver reads.
//
// The hook runs on save, so a recording run still sees live data and its
// assertions fail. Re-record, then run again without SLACK_TOKEN.
func sanitizeSlackUsers(i *cassette.Interaction) error {
	if i.Response.Code != http.StatusOK {
		return fmt.Errorf("refusing to sanitize slack response with status %d", i.Response.Code)
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal([]byte(i.Response.Body), &body); err != nil {
		return fmt.Errorf("cannot decode recorded slack response: %w", err)
	}

	var members []map[string]json.RawMessage
	if err := json.Unmarshal(body["members"], &members); err != nil {
		return fmt.Errorf("cannot decode recorded slack members: %w", err)
	}

	if len(members) == 0 {
		return fmt.Errorf("recorded slack response lists no members")
	}

	for idx, member := range members {
		var profile struct {
			Email string `json:"email"`
		}
		if err := json.Unmarshal(member["profile"], &profile); err != nil {
			return fmt.Errorf("cannot decode recorded slack member %d profile: %w", idx, err)
		}

		name := fmt.Sprintf("Member %d", idx+1)
		synthetic := map[string]string{"real_name": name, "display_name": name, "title": ""}

		if profile.Email != "" {
			synthetic["email"] = fmt.Sprintf("member%d@example.com", idx+1)
		}

		rawProfile, err := json.Marshal(synthetic)
		if err != nil {
			return fmt.Errorf("cannot encode synthetic slack profile: %w", err)
		}

		if id, _ := strconv.Unquote(string(member["id"])); id != "USLACKBOT" {
			member["id"] = json.RawMessage(fmt.Sprintf(`"U%010d"`, idx+1))
		}

		member["team_id"] = json.RawMessage(`"T0000000000"`)
		member["name"] = json.RawMessage(fmt.Sprintf(`"member%d"`, idx+1))
		member["real_name"] = json.RawMessage(strconv.Quote(name))
		member["profile"] = rawProfile
	}

	rawMembers, err := json.Marshal(members)
	if err != nil {
		return fmt.Errorf("cannot re-encode sanitized slack members: %w", err)
	}

	body["members"] = rawMembers

	sanitized, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("cannot re-encode sanitized slack response: %w", err)
	}

	replaceCassetteBody(i, string(sanitized))

	return nil
}

// Recorded with an admin's user token, the only token Slack returns has_2fa
// to.
func TestSlackDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/slack", "SLACK_TOKEN", sanitizeSlackUsers)
	client := newVCRClient(rec, bearerAuth(os.Getenv("SLACK_TOKEN")))

	driver := NewSlackDriver(client, "https://slack.com/api")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, records)

	var active []AccountRecord

	for _, r := range records {
		assert.NotEmpty(t, r.Email)
		assert.NotEmpty(t, r.ExternalID)
		assert.NotEmpty(t, r.Roles)

		if *r.Active {
			active = append(active, r)
		}
	}

	require.NotEmpty(t, active, "expected at least one active member")

	for _, r := range active {
		assert.NotEqual(t, coredata.MFAStatusUnknown, r.MFAStatus, "active member %s has no has_2fa", r.ExternalID)
	}
}

func TestSlackDriverMFAStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"members":[
			{"id":"U1","profile":{"email":"on@example.com"},"has_2fa":true},
			{"id":"U2","profile":{"email":"off@example.com"},"has_2fa":false},
			{"id":"U3","profile":{"email":"hidden@example.com"}}
		]}`))
	}))
	t.Cleanup(server.Close)

	records, err := NewSlackDriver(server.Client(), server.URL).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	assert.Equal(t, coredata.MFAStatusEnabled, records[0].MFAStatus)
	assert.Equal(t, coredata.MFAStatusDisabled, records[1].MFAStatus)
	// Slack only returns has_2fa to an admin caller; a bot token never sees it.
	assert.Equal(t, coredata.MFAStatusUnknown, records[2].MFAStatus)
}

func TestCheckSlackInstallerIsAdmin(t *testing.T) {
	t.Parallel()

	newServer := func(t *testing.T, user string) *httptest.Server {
		t.Helper()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch r.URL.Path {
			case "/auth.test":
				_, _ = w.Write([]byte(`{"ok":true,"user_id":"U1","team":"Acme"}`))
			case "/users.info":
				assert.Equal(t, "U1", r.URL.Query().Get("user"))

				_, _ = w.Write([]byte(`{"ok":true,"user":` + user + `}`))
			default:
				http.NotFound(w, r)
			}
		}))
		t.Cleanup(server.Close)

		return server
	}

	t.Run("admin", func(t *testing.T) {
		t.Parallel()

		server := newServer(t, `{"id":"U1","is_admin":true}`)
		require.NoError(t, CheckSlackInstallerIsAdmin(context.Background(), server.Client(), server.URL))
	})

	t.Run("primary owner", func(t *testing.T) {
		t.Parallel()

		server := newServer(t, `{"id":"U1","is_primary_owner":true}`)
		require.NoError(t, CheckSlackInstallerIsAdmin(context.Background(), server.Client(), server.URL))
	})

	t.Run("member", func(t *testing.T) {
		t.Parallel()

		server := newServer(t, `{"id":"U1"}`)
		err := CheckSlackInstallerIsAdmin(context.Background(), server.Client(), server.URL)
		_, ok := errors.AsType[*InstallRejectedError](err)
		require.True(t, ok, "expected an install rejection, got %v", err)
	})
}
