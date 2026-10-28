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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
)

const (
	daytonaFixtureOrganizationID = "00000000-0000-4000-8000-000000000001"
	daytonaFixtureUserID         = "user_00000000000000000000000001"
)

func TestDaytonaDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/daytona", "DAYTONA_TOKEN", sanitizeDaytona, dropResponseHeaders("Etag"))
	client := newVCRClient(rec, bearerAuth(os.Getenv("DAYTONA_TOKEN")))
	driver := NewDaytonaDriver(client, "https://app.daytona.io/api")

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)

	owner := records[0]
	assert.Equal(t, "member1@example.com", owner.Email)
	assert.Equal(t, "Member 1", owner.FullName)
	assert.Equal(t, daytonaFixtureUserID, owner.ExternalID)
	// An owner holds every permission, so Daytona lists no assignments.
	assert.Equal(t, []string{"Owner"}, owner.Roles)
	assert.Equal(t, new(true), owner.IsAdmin)
	assert.Equal(t, coredata.MFAStatusUnknown, owner.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, owner.AuthMethod)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, owner.AccountType)
	require.NotNil(t, owner.CreatedAt)
	assert.False(t, owner.CreatedAt.IsZero())

	// Membership carries no status, and Daytona exposes no sign-in time.
	assert.Nil(t, owner.Active)
	assert.Nil(t, owner.LastLogin)
}

func TestDaytonaDriverMember(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/api-keys/current":
				_, _ = w.Write([]byte(`{"organizationId":"org-1"}`))
			case "/api/organizations/org-1/users":
				_, _ = w.Write([]byte(`[{"organizationId":"org-1","userId":"user_2","role":"member","assignedRoles":[{"name":"Viewer"},{"name":"Developer"}],"name":"John Smith","email":"john.smith@example.com","createdAt":"2026-10-01T09:00:00.000Z"}]`))
			default:
				http.NotFound(w, r)
			}
		}),
	)
	t.Cleanup(srv.Close)

	records, err := NewDaytonaDriver(srv.Client(), srv.URL+"/api").ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)

	assert.Equal(t, "user_2", records[0].ExternalID)
	assert.Equal(t, []string{"Member", "Viewer", "Developer"}, records[0].Roles)
	assert.Equal(t, new(false), records[0].IsAdmin)
}

func TestDaytonaDriverRejectedKey(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"statusCode":401,"error":"Unauthorized","message":"Invalid credentials"}`))
		}),
	)
	t.Cleanup(srv.Close)

	_, err := NewDaytonaDriver(srv.Client(), srv.URL).ListAccounts(context.Background())
	require.ErrorContains(t, err, "unexpected status 401")
}

func TestDaytonaDriverRefusesKeyWithoutOrganization(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"name":"probo"}`))
		}),
	)
	t.Cleanup(srv.Close)

	_, err := NewDaytonaDriver(srv.Client(), srv.URL).ListAccounts(context.Background())
	require.ErrorContains(t, err, "no organization id")
}

func TestDaytonaDriverRefusesMemberWithoutUserID(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api-keys/current" {
				_, _ = w.Write([]byte(`{"organizationId":"org-1"}`))

				return
			}

			_, _ = w.Write([]byte(`[{"role":"member","email":"john.smith@example.com"}]`))
		}),
	)
	t.Cleanup(srv.Close)

	_, err := NewDaytonaDriver(srv.Client(), srv.URL).ListAccounts(context.Background())
	require.ErrorContains(t, err, "member has no user id")
}

func TestSanitizeDaytonaMembers(t *testing.T) {
	t.Parallel()

	newInteraction := func(body string) *cassette.Interaction {
		i := &cassette.Interaction{}
		i.Request.URL = "https://app.daytona.io/api/organizations/recorded-org/users"
		i.Response.Code = http.StatusOK
		i.Response.Body = body

		return i
	}

	t.Run(
		"rewrites identity and custom roles",
		func(t *testing.T) {
			t.Parallel()

			i := newInteraction(`[{"organizationId":"recorded-org","userId":"user_recorded","role":"member","name":"Recorded Person","email":"recorded.person@acme.test","createdAt":"2026-10-01T09:00:00.000Z","updatedAt":"2026-10-01T09:00:00.000Z","assignedRoles":[` +
				`{"id":"00000000-0000-0000-0000-000000000001","name":"Viewer","description":"Grants read access","permissions":[],"isGlobal":true,"createdAt":"2026-01-01T00:00:00.000Z","updatedAt":"2026-01-01T00:00:00.000Z"},` +
				`{"id":"recorded-role","name":"Recorded Team","description":"Owned by recorded.person@acme.test","permissions":["write:sandboxes"],"isGlobal":false,"createdAt":"2026-01-01T00:00:00.000Z","updatedAt":"2026-01-01T00:00:00.000Z"}]}]`)

			require.NoError(t, sanitizeDaytona(i))

			saved := i.Request.URL + i.Response.Body
			assert.NotContains(t, saved, "recorded")
			assert.NotContains(t, saved, "Recorded")
			assert.NotContains(t, saved, "acme")
			assert.Contains(t, i.Response.Body, `"name":"Viewer"`)
			assert.Contains(t, i.Response.Body, `"name":"Custom Role 2"`)
		},
	)

	t.Run(
		"refuses an unreviewed assigned role field",
		func(t *testing.T) {
			t.Parallel()

			i := newInteraction(`[{"organizationId":"recorded-org","userId":"user_recorded","role":"member","name":"Recorded Person","email":"recorded.person@acme.test","createdAt":"2026-10-01T09:00:00.000Z","updatedAt":"2026-10-01T09:00:00.000Z","assignedRoles":[` +
				`{"id":"recorded-role","name":"Recorded Team","description":"","permissions":[],"isGlobal":false,"createdBy":"recorded.person@acme.test","createdAt":"2026-01-01T00:00:00.000Z","updatedAt":"2026-01-01T00:00:00.000Z"}]}]`)

			require.ErrorContains(t, sanitizeDaytona(i), `unreviewed field "createdBy"`)
		},
	)
}

// sanitizeDaytona rewrites recorded identity on save and refuses any endpoint
// it has no rewriter for. Re-record, then rerun without DAYTONA_TOKEN.
func sanitizeDaytona(i *cassette.Interaction) error {
	if i.Response.Code != http.StatusOK {
		return fmt.Errorf("refusing to sanitize daytona response with status %d", i.Response.Code)
	}

	u, err := url.Parse(i.Request.URL)
	if err != nil {
		return fmt.Errorf("cannot parse recorded daytona URL: %w", err)
	}

	segments := strings.Split(strings.TrimPrefix(u.Path, "/api/"), "/")

	switch {
	case len(segments) == 2 && segments[0] == "api-keys" && segments[1] == "current":
		return sanitizeDaytonaAPIKey(i)
	case len(segments) == 3 && segments[0] == "organizations" && segments[2] == daytonaUsersSegment:
		// The organization id is a path segment, so the request URL is rewritten too.
		u.Path = "/api/organizations/" + daytonaFixtureOrganizationID + "/users"
		i.Request.URL = u.String()

		return sanitizeDaytonaMembers(i, segments[1])
	default:
		return fmt.Errorf("refusing to sanitize an unrecognised daytona interaction")
	}
}

func sanitizeDaytonaAPIKey(i *cassette.Interaction) error {
	var key map[string]json.RawMessage
	if err := json.Unmarshal([]byte(i.Response.Body), &key); err != nil {
		return fmt.Errorf("cannot decode recorded daytona api key: %w", err)
	}

	recorded, err := sanitizeDaytonaObject(
		key,
		[]string{"createdAt", "expiresAt", "lastUsedAt", "name", "organizationId", "permissions", "userId", "value"},
		map[string]string{
			"name":           "probo-access-review",
			"value":          "dtn********************000",
			"organizationId": daytonaFixtureOrganizationID,
			"userId":         daytonaFixtureUserID,
		},
	)
	if err != nil {
		return fmt.Errorf("recorded daytona api key: %w", err)
	}

	return replaceDaytonaBody(i, key, recorded)
}

func sanitizeDaytonaMembers(i *cassette.Interaction, organizationID string) error {
	var members []map[string]json.RawMessage
	if err := json.Unmarshal([]byte(i.Response.Body), &members); err != nil {
		return fmt.Errorf("cannot decode recorded daytona members: %w", err)
	}

	if len(members) == 0 {
		return fmt.Errorf("recorded daytona members response lists nobody")
	}

	recorded := []string{organizationID}

	for idx, member := range members {
		values, err := sanitizeDaytonaObject(
			member,
			[]string{"assignedRoles", "createdAt", "email", "name", "organizationId", "role", "updatedAt", "userId"},
			map[string]string{
				"organizationId": daytonaFixtureOrganizationID,
				"userId":         fmt.Sprintf("user_%026d", idx+1),
				"name":           fmt.Sprintf("Member %d", idx+1),
				"email":          fmt.Sprintf("member%d@example.com", idx+1),
			},
		)
		if err != nil {
			return fmt.Errorf("recorded daytona member: %w", err)
		}

		recorded = append(recorded, values...)

		values, err = sanitizeDaytonaAssignedRoles(member)
		if err != nil {
			return fmt.Errorf("recorded daytona member: %w", err)
		}

		recorded = append(recorded, values...)
	}

	return replaceDaytonaBody(i, members, recorded)
}

// sanitizeDaytonaAssignedRoles keeps Daytona's built-in roles and rewrites
// the custom ones, whose name and description the customer writes.
func sanitizeDaytonaAssignedRoles(member map[string]json.RawMessage) ([]string, error) {
	var roles []map[string]json.RawMessage
	if err := json.Unmarshal(member["assignedRoles"], &roles); err != nil {
		return nil, fmt.Errorf("cannot decode assigned roles: %w", err)
	}

	var recorded []string

	for idx, role := range roles {
		var global bool
		if err := json.Unmarshal(role["isGlobal"], &global); err != nil {
			return nil, fmt.Errorf("assigned role has no isGlobal flag: %w", err)
		}

		replacements := map[string]string{
			"id":          fmt.Sprintf("00000000-0000-4000-8000-%012d", idx+1),
			"name":        fmt.Sprintf("Custom Role %d", idx+1),
			"description": "",
		}
		if global {
			replacements = nil
		}

		values, err := sanitizeDaytonaObject(
			role,
			[]string{"createdAt", "description", "id", "isGlobal", "name", "permissions", "updatedAt"},
			replacements,
		)
		if err != nil {
			return nil, fmt.Errorf("assigned role: %w", err)
		}

		recorded = append(recorded, values...)
	}

	encoded, err := json.Marshal(roles)
	if err != nil {
		return nil, fmt.Errorf("cannot encode assigned roles: %w", err)
	}

	member["assignedRoles"] = encoded

	return recorded, nil
}

// sanitizeDaytonaObject replaces identity fields in place and refuses a field
// nobody has reviewed. It returns the recorded values it replaced.
func sanitizeDaytonaObject(
	obj map[string]json.RawMessage,
	reviewed []string,
	replacements map[string]string,
) ([]string, error) {
	for field := range obj {
		if !slices.Contains(reviewed, field) {
			return nil, fmt.Errorf("unreviewed field %q", field)
		}
	}

	var recorded []string

	for field, replacement := range replacements {
		var value string
		if err := json.Unmarshal(obj[field], &value); err != nil {
			return nil, fmt.Errorf("field %q is not a string: %w", field, err)
		}

		encoded, err := json.Marshal(replacement)
		if err != nil {
			return nil, fmt.Errorf("cannot encode replacement for %q: %w", field, err)
		}

		obj[field] = encoded

		if value != "" && value != replacement {
			recorded = append(recorded, value)
		}
	}

	return recorded, nil
}

// replaceDaytonaBody re-encodes the sanitized body and fails if a replaced
// value survives anywhere in it.
func replaceDaytonaBody(i *cassette.Interaction, body any, recorded []string) error {
	sanitized, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("cannot re-encode sanitized daytona response: %w", err)
	}

	for _, value := range recorded {
		if bytes.Contains(sanitized, []byte(value)) || strings.Contains(i.Request.URL, value) {
			return fmt.Errorf("sanitized daytona interaction still carries a recorded identity value")
		}
	}

	replaceCassetteBody(i, string(sanitized))

	return nil
}
