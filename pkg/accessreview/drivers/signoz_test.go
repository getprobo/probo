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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
)

// TestSigNozDriver records with SIGNOZ_API_KEY (a signoz-admin service
// account) and SIGNOZ_BASE_URL set, then replays the sanitized cassette.
func TestSigNozDriver(t *testing.T) {
	t.Parallel()

	sanitizer := newSigNozCassetteSanitizer()
	rec := newRecorder(t, "testdata/signoz", "SIGNOZ_API_KEY", sanitizer.sanitize)
	client := newVCRClientWithHeader(rec, "SIGNOZ-API-KEY", os.Getenv("SIGNOZ_API_KEY"))

	baseURL := os.Getenv("SIGNOZ_BASE_URL")
	if baseURL == "" {
		baseURL = "https://" + sigNozCassetteHost
	}

	records, err := NewSigNozDriver(client, baseURL).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)

	assert.Equal(t, "member1@example.com", records[0].Email)
	assert.Equal(t, "Member 1", records[0].FullName)
	assert.Equal(t, []string{"Admin"}, records[0].Roles)
	assert.Equal(t, new(true), records[0].IsAdmin)
	assert.Equal(t, "00000000-0000-4000-8000-000000000001", records[0].ExternalID)
	assert.Equal(t, coredata.MFAStatusUnknown, records[0].MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, records[0].AccountType)
	require.NotNil(t, records[0].Active)
	assert.True(t, *records[0].Active)
	require.NotNil(t, records[0].CreatedAt)

	name, err := NewSigNozNameResolver(client, baseURL).ResolveInstanceName(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Example Org", name)
}

const (
	sigNozCassetteHost  = "signoz.example.com"
	sigNozCassetteOrgID = "00000000-0000-4000-8000-0000000000aa"
)

// sigNozCassetteSanitizer replaces tenant identity in bodies and URLs. It relies
// on the users response being saved before the roles requests that address it.
type sigNozCassetteSanitizer struct {
	rewrites map[string]string
	recorded []string
	users    int
	roles    int
}

func newSigNozCassetteSanitizer() *sigNozCassetteSanitizer {
	return &sigNozCassetteSanitizer{rewrites: map[string]string{}}
}

func (s *sigNozCassetteSanitizer) sanitize(i *cassette.Interaction) error {
	if i.Response.Code != http.StatusOK {
		return fmt.Errorf("refusing to sanitize signoz response with status %d", i.Response.Code)
	}

	u, err := url.Parse(i.Request.URL)
	if err != nil {
		return fmt.Errorf("cannot parse recorded signoz URL: %w", err)
	}

	if u.Host != sigNozCassetteHost {
		s.recorded = append(s.recorded, u.Host)
	}

	var body struct {
		Status string          `json:"status"`
		Data   json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal([]byte(i.Response.Body), &body); err != nil {
		return fmt.Errorf("cannot decode recorded signoz response: %w", err)
	}

	var data any

	switch segments := strings.Split(strings.Trim(u.Path, "/"), "/"); {
	case u.Path == "/api/v2/users":
		var users []map[string]any
		if err := json.Unmarshal(body.Data, &users); err != nil {
			return fmt.Errorf("cannot decode recorded signoz users: %w", err)
		}

		if len(users) == 0 {
			return fmt.Errorf("recorded signoz response lists no users")
		}

		for _, user := range users {
			s.users++
			for _, r := range []struct {
				field, replacement string
				id                 bool
			}{
				{"id", fmt.Sprintf("00000000-0000-4000-8000-%012d", s.users), true},
				{"orgId", sigNozCassetteOrgID, true},
				{"email", fmt.Sprintf("member%d@example.com", s.users), false},
				{"displayName", fmt.Sprintf("Member %d", s.users), false},
			} {
				if err := s.register(user, r.field, r.replacement, r.id); err != nil {
					return err
				}
			}
		}

		data = users
	case len(segments) == 5 && segments[0] == "api" && segments[1] == "v2" && segments[2] == "users" && segments[4] == "roles":
		if _, ok := s.rewrites[segments[3]]; !ok {
			return fmt.Errorf("recorded signoz roles request addresses a user absent from the users response")
		}

		var roles []map[string]any
		if err := json.Unmarshal(body.Data, &roles); err != nil {
			return fmt.Errorf("cannot decode recorded signoz roles: %w", err)
		}

		for _, role := range roles {
			replacement, known := s.rewrites[fmt.Sprint(role["id"])]
			if !known {
				s.roles++
				replacement = fmt.Sprintf("00000000-0000-4000-9000-%012d", s.roles)
			}

			if err := s.register(role, "id", replacement, true); err != nil {
				return err
			}

			if err := s.register(role, "orgId", sigNozCassetteOrgID, true); err != nil {
				return err
			}
		}

		data = roles
	case u.Path == "/api/v2/orgs/me":
		var org map[string]any
		if err := json.Unmarshal(body.Data, &org); err != nil {
			return fmt.Errorf("cannot decode recorded signoz organization: %w", err)
		}

		for _, r := range []struct {
			field, replacement string
			id                 bool
		}{
			{"id", sigNozCassetteOrgID, true},
			{"name", "example-org", false},
			{"displayName", "Example Org", false},
			{"alias", "example-alias", false},
		} {
			if err := s.register(org, r.field, r.replacement, r.id); err != nil {
				return err
			}
		}

		if _, ok := org["key"]; ok {
			org["key"] = 12345678
		}

		data = org
	default:
		return fmt.Errorf("refusing to record unexpected signoz path %q", u.Path)
	}

	rewritten, err := json.Marshal(map[string]any{
		"status": body.Status,
		"data":   data,
	})
	if err != nil {
		return fmt.Errorf("cannot encode sanitized signoz response: %w", err)
	}

	segments := strings.Split(u.Path, "/")
	for idx, segment := range segments {
		if replacement, ok := s.rewrites[segment]; ok {
			segments[idx] = replacement
		}
	}

	u.Host = sigNozCassetteHost
	u.Path = strings.Join(segments, "/")
	i.Request.URL = u.String()
	i.Request.Host = sigNozCassetteHost
	replaceCassetteBody(i, string(rewritten))

	for _, recorded := range s.recorded {
		if strings.Contains(i.Request.URL, recorded) || strings.Contains(i.Response.Body, recorded) {
			return fmt.Errorf("sanitized signoz interaction still carries a recorded identity value")
		}
	}

	return nil
}

// register replaces a non-blank string field, keeping its recorded value for
// the leak check and, for an id, for the URL rewrite.
func (s *sigNozCassetteSanitizer) register(object map[string]any, field, replacement string, id bool) error {
	raw, ok := object[field]
	if !ok || raw == nil {
		return nil
	}

	value, ok := raw.(string)
	if !ok {
		return fmt.Errorf("recorded signoz field %q is not a string", field)
	}

	if strings.TrimSpace(value) == "" {
		return nil
	}

	object[field] = replacement

	if value == replacement {
		return nil
	}

	s.recorded = append(s.recorded, value)

	if id {
		s.rewrites[value] = replacement
	}

	return nil
}

func TestSigNozDriverRoleStatusMatrix(t *testing.T) {
	t.Parallel()

	roles := map[string]string{
		"u1": `[{"name":"signoz-admin"}]`,
		"u2": `[{"name":"signoz-viewer"}]`,
		"u3": `[{"name":"signoz-editor"}]`,
		"u4": `[{"name":"signoz-viewer"}]`,
		"u5": `[{"name":"signoz-editor"},{"name":"billing-reader"}]`,
		"u6": `[]`,
		"u8": `[{"name":"admin"},{"name":"signoz-viewer"},{"name":"signoz-viewer"}]`,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/api/v2/users" {
			_, _ = w.Write([]byte(`{"status":"success","data":[
				{"id":"u1","email":"admin@example.com","displayName":"Admin User","status":"active","isRoot":false,"createdAt":"2026-05-01T10:20:30Z"},
				{"id":"u2","email":"owner@example.com","displayName":"Owner User","status":"active","isRoot":true},
				{"id":"u3","email":"editor@example.com","displayName":"Editor User","status":"active","isRoot":false},
				{"id":"u4","email":"invited@example.com","displayName":"Invited User","status":"pending_invite","isRoot":false},
				{"id":"u5","email":"removed@example.com","displayName":"Removed User","status":"deleted","isRoot":false},
				{"id":"u6","email":"norole@example.com","displayName":"No Role","status":"active","isRoot":false},
				{"id":"u7","email":"","displayName":"No Email","status":"active","isRoot":false},
				{"id":"u8","email":"custom@example.com","displayName":"Custom Admin","status":"active","isRoot":false},
				{"id":"","email":"noid@example.com","displayName":"No ID","status":"active","isRoot":false},
				{"id":"u9","email":"gone@example.com","displayName":"Deleted Since Listing","status":"active","isRoot":false}
			]}`))

			return
		}

		rest, prefixed := strings.CutPrefix(r.URL.Path, "/api/v2/users/")
		id, suffixed := strings.CutSuffix(rest, "/roles")

		payload, ok := roles[id]
		if !prefixed || !suffixed || !ok {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		_, _ = w.Write([]byte(`{"status":"success","data":` + payload + `}`))
	}))
	defer srv.Close()

	records, err := NewSigNozDriver(srv.Client(), srv.URL).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 8) // users with no email or no id are skipped, their roles never fetched

	// signoz-admin role -> admin.
	assert.Equal(t, "admin@example.com", records[0].Email)
	assert.Equal(t, "Admin User", records[0].FullName)
	assert.Equal(t, []string{"Admin"}, records[0].Roles)
	assert.Equal(t, new(true), records[0].IsAdmin)
	assert.Equal(t, "u1", records[0].ExternalID)
	require.NotNil(t, records[0].CreatedAt)

	// isRoot -> admin even with a non-admin role.
	assert.Equal(t, []string{"Viewer"}, records[1].Roles)
	assert.Equal(t, new(true), records[1].IsAdmin)

	assert.Equal(t, []string{"Editor"}, records[2].Roles)
	assert.Equal(t, new(false), records[2].IsAdmin)
	require.NotNil(t, records[2].Active)
	assert.True(t, *records[2].Active)

	// pending_invite -> inactive.
	require.NotNil(t, records[3].Active)
	assert.False(t, *records[3].Active)

	// deleted -> inactive; a custom role is kept verbatim next to a managed one.
	assert.Equal(t, []string{"Editor", "billing-reader"}, records[4].Roles)
	require.NotNil(t, records[4].Active)
	assert.False(t, *records[4].Active)

	// No role at all -> empty, not admin.
	assert.Empty(t, records[5].Roles)
	assert.Equal(t, new(false), records[5].IsAdmin)

	// A custom role named "admin" is not the managed admin role; duplicates
	// collapse.
	assert.Equal(t, []string{"admin", "Viewer"}, records[6].Roles)
	assert.Equal(t, new(false), records[6].IsAdmin)

	// Roles route answers 404 (user deleted since the listing) -> kept, no roles.
	assert.Equal(t, "gone@example.com", records[7].Email)
	assert.Empty(t, records[7].Roles)
	assert.Equal(t, new(false), records[7].IsAdmin)
}

func TestSigNozDriverEscapesUserIDInRolesPath(t *testing.T) {
	t.Parallel()

	var rolesPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/api/v2/users" {
			_, _ = w.Write([]byte(`{"status":"success","data":[{"id":"a/b","email":"a@example.com","status":"active"}]}`))

			return
		}

		rolesPath = r.URL.EscapedPath()
		_, _ = w.Write([]byte(`{"status":"success","data":[]}`))
	}))
	defer srv.Close()

	_, err := NewSigNozDriver(srv.Client(), srv.URL).ListAccounts(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "/api/v2/users/a%2Fb/roles", rolesPath)
}

func TestSigNozDriverRetriesTransientRolesFailure(t *testing.T) {
	t.Parallel()

	var rolesCalls int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/api/v2/users" {
			_, _ = w.Write([]byte(`{"status":"success","data":[{"id":"u1","email":"a@example.com","status":"active"}]}`))

			return
		}

		rolesCalls++
		if rolesCalls == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)

			return
		}

		_, _ = w.Write([]byte(`{"status":"success","data":[{"name":"signoz-admin"}]}`))
	}))
	defer srv.Close()

	records, err := NewSigNozDriver(srv.Client(), srv.URL).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, []string{"Admin"}, records[0].Roles)
	assert.Equal(t, 2, rolesCalls)
}

func TestSigNozDriverListAccountsEmptyData(t *testing.T) {
	t.Parallel()

	for name, payload := range map[string]string{
		"null data":   `{"status":"success","data":null}`,
		"empty array": `{"status":"success","data":[]}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(payload))
			}))
			defer srv.Close()

			records, err := NewSigNozDriver(srv.Client(), srv.URL).ListAccounts(context.Background())
			require.NoError(t, err)
			assert.Empty(t, records)
		})
	}
}

func TestSigNozDriverListAccountsErrorStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"status":"error"}`))
	}))
	defer srv.Close()

	_, err := NewSigNozDriver(srv.Client(), srv.URL).ListAccounts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status 403")
}

func TestSigNozDriverListAccountsRolesErrorStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/users" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"success","data":[{"id":"u1","email":"a@example.com","status":"active"}]}`))

			return
		}

		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	_, err := NewSigNozDriver(srv.Client(), srv.URL).ListAccounts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot fetch signoz roles")
	assert.Contains(t, err.Error(), "unexpected status 403")
}

func TestNormalizeSigNozRole(t *testing.T) {
	t.Parallel()

	for in, want := range map[string]string{
		"signoz-admin":  "Admin",
		"signoz-editor": "Editor",
		"signoz-viewer": "Viewer",
		"":              "",
		"custom-role":   "custom-role", // unknown role preserved verbatim
		"superadmin":    "superadmin",  // contains "admin" but must NOT be promoted
		"admin":         "admin",       // a custom role, not the managed admin role
	} {
		assert.Equalf(t, want, normalizeSigNozRole(in), "role %q", in)
	}
}

func TestSigNozActiveStatus(t *testing.T) {
	t.Parallel()

	active := sigNozActiveStatus("active")
	require.NotNil(t, active)
	assert.True(t, *active)

	for _, status := range []string{"pending_invite", "deleted"} {
		v := sigNozActiveStatus(status)
		require.NotNilf(t, v, "status %q", status)
		assert.Falsef(t, *v, "status %q", status)
	}

	assert.Nil(t, sigNozActiveStatus("something_unexpected"))
}

func TestSigNozNameResolver(t *testing.T) {
	t.Parallel()

	t.Run("falls back to name when displayName is empty", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"success","data":{"displayName":"","name":"acme"}}`))
		}))
		defer srv.Close()

		name, err := NewSigNozNameResolver(srv.Client(), srv.URL).ResolveInstanceName(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "acme", name)
	})

	t.Run("returns empty without error on terminal failure", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer srv.Close()

		name, err := NewSigNozNameResolver(srv.Client(), srv.URL).ResolveInstanceName(context.Background())
		require.NoError(t, err)
		assert.Empty(t, name)
	})
}
