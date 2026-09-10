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
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
)

// Synthetic identity the sanitizer writes over the recorded workspace, and
// which the replay assertions below then read back.
const (
	attioFixtureWorkspaceID   = "14beef7a-0000-4000-8000-000000000001"
	attioFixtureClientID      = "6f0b9b44-0000-4000-8000-000000000009"
	attioFixtureWorkspaceName = "Example Workspace"
	attioFixtureWorkspaceSlug = "example-workspace"
)

func attioFixtureMemberID(idx int) string {
	return fmt.Sprintf("50cf242c-0000-4000-8000-%012d", idx+1)
}

// TestAttioDriver replays a recording taken against a live Attio workspace, so
// the field names and the response envelope are the provider's own rather than
// this package's idea of them. The recorded workspace holds a single admin
// seat; the member and suspended levels are covered exhaustively by
// TestAttioAccessLevelMapping, which needs no fixture.
func TestAttioDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/attio", "ATTIO_TOKEN", sanitizeAttio)
	client := newVCRClient(rec, bearerAuth(os.Getenv("ATTIO_TOKEN")))
	driver := NewAttioDriver(client, "https://api.attio.com/v2")

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)

	admin := records[0]
	assert.Equal(t, "member1@example.com", admin.Email)
	assert.Equal(t, "Member 1", admin.FullName)
	assert.Equal(t, attioFixtureMemberID(0), admin.ExternalID)
	assert.Equal(t, []string{"Admin"}, admin.Roles)
	require.NotNil(t, admin.Active)
	assert.True(t, *admin.Active)
	require.NotNil(t, admin.IsAdmin)
	assert.True(t, *admin.IsAdmin)

	// Attio stamps created_at with nine fractional digits, which the RFC3339
	// layout does not itself spell out — Go accepts the fraction on parse.
	require.NotNil(t, admin.CreatedAt)
	assert.False(t, admin.CreatedAt.IsZero())

	// Nothing Attio returns carries a last login or an MFA signal.
	assert.Nil(t, admin.LastLogin)
	assert.Equal(t, coredata.MFAStatusUnknown, admin.MFAStatus)
}

func TestAttioAccessLevelMapping(t *testing.T) {
	t.Parallel()

	cases := []struct {
		accessLevel string
		wantRoles   []string
		wantActive  *bool
		wantIsAdmin *bool
	}{
		{accessLevel: "admin", wantRoles: []string{"Admin"}, wantActive: new(true), wantIsAdmin: new(true)},
		{accessLevel: "member", wantRoles: []string{"Member"}, wantActive: new(true), wantIsAdmin: new(false)},
		// Attio never deletes a workspace member; it overwrites access_level
		// with "suspended". That is a status, so it lands on Active, and the
		// role the member held is gone rather than replaced by a fake one.
		{accessLevel: "suspended", wantRoles: []string{}, wantActive: new(false), wantIsAdmin: new(false)},
		// Matched case-insensitively so a cased value cannot silently downgrade
		// a known admin to unknown.
		{accessLevel: "  Admin ", wantRoles: []string{"Admin"}, wantActive: new(true), wantIsAdmin: new(true)},
		// A level Attio adds later is not evidence the seat is live, so both
		// three-valued fields stay unknown rather than guessing.
		{accessLevel: "guest", wantRoles: []string{"guest"}, wantActive: nil, wantIsAdmin: nil},
		{accessLevel: "", wantRoles: []string{}, wantActive: nil, wantIsAdmin: nil},
	}

	for _, tc := range cases {
		t.Run(
			tc.accessLevel,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tc.wantRoles, attioRoles(tc.accessLevel))

				active, isAdmin := attioAccountStatus(tc.accessLevel)
				assert.Equal(t, tc.wantActive, active)
				assert.Equal(t, tc.wantIsAdmin, isAdmin)
			},
		)
	}
}

func TestAttioNameResolver(t *testing.T) {
	t.Parallel()

	// Its own cassette rather than a second interaction in testdata/attio:
	// go-vcr writes the whole file on Stop, so two recorders sharing a path
	// race and the last one to finish drops the other's interaction.
	rec := newRecorder(t, "testdata/attio_self", "ATTIO_TOKEN", sanitizeAttio)
	client := newVCRClient(rec, bearerAuth(os.Getenv("ATTIO_TOKEN")))
	resolver := NewAttioNameResolver(client, "https://api.attio.com/v2")

	name, err := resolver.ResolveInstanceName(context.Background())
	require.NoError(t, err)
	assert.Equal(t, attioFixtureWorkspaceName, name)
}

// TestAttioNameResolverInactiveTokenIsTerminal covers the trap that sets Attio
// apart from the other providers here: /v2/self answers 200 with
// {"active": false} for a revoked, deleted or unknown token instead of a 4xx.
// A plain error there would leave the source-name worker re-claiming a dead
// connection until its attempt budget burned out.
func TestAttioNameResolverInactiveTokenIsTerminal(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"active":false}`))
		}),
	)
	t.Cleanup(srv.Close)

	_, err := NewAttioNameResolver(srv.Client(), srv.URL).ResolveInstanceName(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTerminalNameResolution)
}

// sanitizeAttio rewrites the identity in a recording of a live Attio workspace
// before it reaches disk. It dispatches on the endpoint because the two bodies
// carry different identity fields, and refuses anything else so a new
// interaction cannot be committed unsanitized. The same refusal applies one
// level down, to a response field the rewriters have never considered.
//
// The hook runs on save, so a recording run still sees live data and the
// assertions above fail against it. Re-record, then run again without
// ATTIO_TOKEN.
func sanitizeAttio(i *cassette.Interaction) error {
	if i.Response.Code != http.StatusOK {
		return fmt.Errorf("refusing to sanitize attio response with status %d", i.Response.Code)
	}

	switch {
	case strings.HasSuffix(i.Request.URL, attioWorkspaceMembersPath):
		return sanitizeAttioBody(i, sanitizeAttioWorkspaceMembers)
	case strings.HasSuffix(i.Request.URL, attioSelfPath):
		return sanitizeAttioBody(i, sanitizeAttioSelf)
	default:
		return fmt.Errorf("refusing to sanitize an unrecognised attio interaction")
	}
}

// sanitizeAttioBody decodes the recorded body into a raw field map, hands it to
// the endpoint's rewriter, and writes the result back.
//
// It edits the decoded JSON in place rather than re-marshalling the driver's
// own structs, which would rewrite the body as a serialization of the type
// under test and drop whatever the driver does not model — exactly the drift
// this cassette exists to catch.
func sanitizeAttioBody(
	i *cassette.Interaction,
	rewrite func(map[string]json.RawMessage) ([]string, error),
) error {
	var body map[string]json.RawMessage
	if err := json.Unmarshal([]byte(i.Response.Body), &body); err != nil {
		return fmt.Errorf("cannot decode recorded attio response: %w", err)
	}

	recorded, err := rewrite(body)
	if err != nil {
		return err
	}

	sanitized, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("cannot re-encode attio response: %w", err)
	}

	// A second guard behind the declared-field check: it catches identity that
	// was rewritten in one place and repeated in another the rewriter does not
	// treat as identity. Aborting writes no cassette, which is the right way to
	// fail: a short value may match by coincidence, and a spurious abort costs
	// a re-record while a miss would commit a real address.
	for _, value := range recorded {
		if strings.Contains(string(sanitized), value) {
			return fmt.Errorf("sanitized attio response still contains a recorded identity value")
		}
	}

	replaceCassetteBody(i, string(sanitized))

	return nil
}

func sanitizeAttioWorkspaceMembers(body map[string]json.RawMessage) ([]string, error) {
	if err := attioRejectUnknownFields(body, []string{"data"}, "members response"); err != nil {
		return nil, err
	}

	raw, ok := body["data"]
	if !ok {
		return nil, fmt.Errorf("recorded attio response has no data field")
	}

	var members []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &members); err != nil {
		return nil, fmt.Errorf("cannot decode recorded attio workspace members: %w", err)
	}

	if len(members) == 0 {
		return nil, fmt.Errorf("recorded attio response lists no workspace members")
	}

	knownMemberFields := []string{
		"access_level",
		"avatar_url",
		"created_at",
		"email_address",
		"first_name",
		"id",
		"last_name",
	}

	var recorded []string

	for idx, member := range members {
		if err := attioRejectUnknownFields(member, knownMemberFields, fmt.Sprintf("member %d", idx)); err != nil {
			return nil, err
		}

		ids, err := attioMemberIDs(member, idx)
		if err != nil {
			return nil, err
		}

		replacements := map[string]string{
			"first_name":    "Member",
			"last_name":     fmt.Sprintf("%d", idx+1),
			"email_address": fmt.Sprintf("member%d@example.com", idx+1),
		}

		for field, replacement := range replacements {
			value, err := attioRequiredString(member, field, idx)
			if err != nil {
				return nil, err
			}

			recorded = append(recorded, value)

			if err := attioSetString(member, field, replacement); err != nil {
				return nil, err
			}
		}

		// A Google-hosted avatar URL embeds the account it belongs to, so it
		// is identity too. Null is left as recorded — the driver ignores the
		// field, and substituting a value would change what replays.
		if value, present, err := attioOptionalString(member, "avatar_url", idx); err != nil {
			return nil, err
		} else if present {
			recorded = append(recorded, value)

			if err := attioSetString(member, "avatar_url", fmt.Sprintf("https://example.com/avatar/%d.png", idx+1)); err != nil {
				return nil, err
			}
		}

		recorded = append(recorded, ids...)

		sanitizedIDs, err := json.Marshal(map[string]string{
			"workspace_id":        attioFixtureWorkspaceID,
			"workspace_member_id": attioFixtureMemberID(idx),
		})
		if err != nil {
			return nil, fmt.Errorf("cannot encode sanitized attio member id: %w", err)
		}

		member["id"] = sanitizedIDs
	}

	sanitizedMembers, err := json.Marshal(members)
	if err != nil {
		return nil, fmt.Errorf("cannot re-encode attio workspace members: %w", err)
	}

	body["data"] = sanitizedMembers

	return recorded, nil
}

// attioRejectUnknownFields refuses a recorded object carrying a key the
// sanitizer has never considered.
//
// The scan in sanitizeAttioBody only catches a new field that repeats a value
// already rewritten elsewhere, so a new field holding a *different* identity —
// a phone number, a second person's address — would survive it. Listing the
// keys turns that from a silent leak into a failed re-record.
//
// When this trips, decide which kind the new key is: give it a rewrite if it
// carries identity, or add it here if it does not.
func attioRejectUnknownFields(
	fields map[string]json.RawMessage,
	known []string,
	what string,
) error {
	for field := range fields {
		if !slices.Contains(known, field) {
			return fmt.Errorf(
				"recorded attio %s carries the unhandled field %q: sanitize it, or list it as carrying no identity",
				what,
				field,
			)
		}
	}

	return nil
}

// attioMemberIDs returns the recorded workspace and member identifiers,
// decoded rather than merely present: a workspace_member_id that came back as
// null or a number would paper over the struct mismatch this cassette exists
// to expose.
func attioMemberIDs(member map[string]json.RawMessage, idx int) ([]string, error) {
	raw, ok := member["id"]
	if !ok {
		return nil, fmt.Errorf("recorded attio member %d has no id field", idx)
	}

	var id struct {
		WorkspaceID       string `json:"workspace_id"`
		WorkspaceMemberID string `json:"workspace_member_id"`
	}

	if err := json.Unmarshal(raw, &id); err != nil {
		return nil, fmt.Errorf("recorded attio member %d has a malformed id: %w", idx, err)
	}

	if strings.TrimSpace(id.WorkspaceID) == "" || strings.TrimSpace(id.WorkspaceMemberID) == "" {
		return nil, fmt.Errorf("recorded attio member %d has an empty identifier", idx)
	}

	return []string{id.WorkspaceID, id.WorkspaceMemberID}, nil
}

func sanitizeAttioSelf(body map[string]json.RawMessage) ([]string, error) {
	knownFields := []string{
		"active",
		"aud",
		"authorized_by_workspace_member_id",
		"client_id",
		"exp",
		"iat",
		"iss",
		"scope",
		"sub",
		"token_type",
		"workspace_id",
		"workspace_logo_url",
		"workspace_name",
		"workspace_slug",
	}

	if err := attioRejectUnknownFields(body, knownFields, "self response"); err != nil {
		return nil, err
	}

	var recorded []string

	replacements := map[string]string{
		"workspace_name": attioFixtureWorkspaceName,
		"workspace_slug": attioFixtureWorkspaceSlug,
		"workspace_id":   attioFixtureWorkspaceID,
		"sub":            attioFixtureWorkspaceID,
		"client_id":      attioFixtureClientID,
		"aud":            attioFixtureClientID,
	}

	for field, replacement := range replacements {
		value, err := attioRequiredString(body, field, 0)
		if err != nil {
			return nil, err
		}

		recorded = append(recorded, value)

		if err := attioSetString(body, field, replacement); err != nil {
			return nil, err
		}
	}

	// Absent for a workspace access token, which has no authorizing member.
	optional := map[string]string{
		"authorized_by_workspace_member_id": attioFixtureMemberID(0),
		"workspace_logo_url":                "https://example.com/logo.png",
	}

	for field, replacement := range optional {
		value, present, err := attioOptionalString(body, field, 0)
		if err != nil {
			return nil, err
		}

		if !present {
			continue
		}

		recorded = append(recorded, value)

		if err := attioSetString(body, field, replacement); err != nil {
			return nil, err
		}
	}

	return recorded, nil
}

// attioRequiredString decodes a field the recording must carry. Unmarshalling
// JSON null into a string is a no-op in Go rather than an error, so emptiness
// is the check that catches both null and "".
func attioRequiredString(fields map[string]json.RawMessage, field string, idx int) (string, error) {
	raw, ok := fields[field]
	if !ok {
		return "", fmt.Errorf("recorded attio object %d has no %s field", idx, field)
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", fmt.Errorf("recorded attio object %d has a non-string %s: %w", idx, field, err)
	}

	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("recorded attio object %d has an empty %s", idx, field)
	}

	return value, nil
}

// attioOptionalString decodes a nullable field, reporting whether it holds a
// value that has to be replaced. Absent, null or blank is left exactly as
// recorded.
func attioOptionalString(fields map[string]json.RawMessage, field string, idx int) (string, bool, error) {
	raw, ok := fields[field]
	if !ok {
		return "", false, nil
	}

	if string(raw) == "null" {
		return "", false, nil
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", false, fmt.Errorf("recorded attio object %d has a non-string %s: %w", idx, field, err)
	}

	return value, strings.TrimSpace(value) != "", nil
}

func attioSetString(fields map[string]json.RawMessage, field, value string) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cannot encode sanitized %s: %w", field, err)
	}

	fields[field] = encoded

	return nil
}

// TestSanitizeAttio feeds a body shaped like a live recording through the save
// hook and asserts that every identity value is gone while fields the *driver*
// does not model still reach the cassette. The hook only ever runs on a
// recording run, so without this test nothing exercises it.
//
// A field the *sanitizer* does not know is a different case and is refused
// outright — see TestSanitizeAttioRefusesUnexpectedResponses. Drift stays
// visible either way: the refusal names the new field, and once it is declared
// the recording carries it verbatim.
func TestSanitizeAttio(t *testing.T) {
	t.Parallel()

	t.Run("workspace members", func(t *testing.T) {
		t.Parallel()

		i := attioInteraction(
			"https://api.attio.com/v2/workspace_members",
			`{"data":[{"id":{"workspace_id":"11111111-1111-1111-1111-111111111111","workspace_member_id":"22222222-2222-2222-2222-222222222222"},`+
				`"first_name":"Ada","last_name":"Lovelace","avatar_url":"https://lh3.googleusercontent.com/a/ada","email_address":"ada@realcorp.example",`+
				`"created_at":"2022-11-21T13:22:49.061281000Z","access_level":"admin"}]}`,
		)

		require.NoError(t, sanitizeAttio(i))

		for _, secret := range []string{"Ada", "Lovelace", "ada@realcorp.example", "11111111", "22222222", "googleusercontent"} {
			assert.NotContains(t, i.Response.Body, secret)
		}

		assert.Contains(t, i.Response.Body, "member1@example.com")
		assert.Contains(t, i.Response.Body, attioFixtureMemberID(0))
		// The contract this cassette exists to police: a field the driver does
		// not model must still reach the recording, or drift stops being
		// visible. avatar_url is exactly that — sanitized, never read.
		assert.Contains(t, i.Response.Body, `"avatar_url":"https://example.com/avatar/1.png"`)
		assert.Contains(t, i.Response.Body, `"access_level":"admin"`)
		assert.Contains(t, i.Response.Body, `"created_at":"2022-11-21T13:22:49.061281000Z"`)
		assert.Equal(t, int64(len(i.Response.Body)), i.Response.ContentLength)
	})

	t.Run("self", func(t *testing.T) {
		t.Parallel()

		i := attioInteraction(
			"https://api.attio.com/v2/self",
			`{"active":true,"scope":"user_management:read","client_id":"33333333-3333-3333-3333-333333333333","token_type":"Bearer",`+
				`"exp":null,"iat":1789034543,"sub":"11111111-1111-1111-1111-111111111111","aud":"33333333-3333-3333-3333-333333333333",`+
				`"iss":"attio.com","workspace_id":"11111111-1111-1111-1111-111111111111","workspace_name":"Real Corp",`+
				`"workspace_slug":"real-corp","workspace_logo_url":null,`+
				`"authorized_by_workspace_member_id":"22222222-2222-2222-2222-222222222222"}`,
		)

		require.NoError(t, sanitizeAttio(i))

		for _, secret := range []string{"Real Corp", "real-corp", "11111111", "33333333", "22222222"} {
			assert.NotContains(t, i.Response.Body, secret)
		}

		assert.Contains(t, i.Response.Body, attioFixtureWorkspaceName)
		assert.Contains(t, i.Response.Body, `"scope":"user_management:read"`)
		// Null stays null: the rewriter must not invent a logo the recording
		// did not carry.
		assert.Contains(t, i.Response.Body, `"workspace_logo_url":null`)
	})

	// A workspace access token has no authorizing member, so the field is absent
	// rather than null. Requiring it would abort the cassette write.
	t.Run("self without an authorizing member", func(t *testing.T) {
		t.Parallel()

		i := attioInteraction(
			"https://api.attio.com/v2/self",
			`{"active":true,"scope":"user_management:read","client_id":"33333333-3333-3333-3333-333333333333","token_type":"Bearer",`+
				`"exp":null,"iat":1789034543,"sub":"11111111-1111-1111-1111-111111111111","aud":"33333333-3333-3333-3333-333333333333",`+
				`"iss":"attio.com","workspace_id":"11111111-1111-1111-1111-111111111111","workspace_name":"Real Corp",`+
				`"workspace_slug":"real-corp","workspace_logo_url":null}`,
		)

		require.NoError(t, sanitizeAttio(i))
		assert.NotContains(t, i.Response.Body, "authorized_by_workspace_member_id")
		assert.NotContains(t, i.Response.Body, "Real Corp")
	})
}

// TestSanitizeAttioRefusesUnexpectedResponses covers the aborts. Each one
// writes no cassette, which is the right way to fail: committing a recording
// the rewriter did not fully understand is how real identity leaks.
func TestSanitizeAttioRefusesUnexpectedResponses(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		url  string
		code int
		body string
	}{
		{
			name: "non-200 response",
			url:  "https://api.attio.com/v2/workspace_members",
			code: http.StatusForbidden,
			body: `{"status_code":403}`,
		},
		{
			name: "unrecognised endpoint",
			url:  "https://api.attio.com/v2/objects",
			code: http.StatusOK,
			body: `{"data":[]}`,
		},
		{
			name: "members response with no data field",
			url:  "https://api.attio.com/v2/workspace_members",
			code: http.StatusOK,
			body: `{}`,
		},
		{
			name: "members response listing nobody",
			url:  "https://api.attio.com/v2/workspace_members",
			code: http.StatusOK,
			body: `{"data":[]}`,
		},
		{
			name: "member with a non-string email",
			url:  "https://api.attio.com/v2/workspace_members",
			code: http.StatusOK,
			body: `{"data":[{"id":{"workspace_id":"a","workspace_member_id":"b"},"first_name":"Ada","last_name":"L","email_address":42,"created_at":"x","access_level":"admin"}]}`,
		},
		{
			name: "member with a null identifier",
			url:  "https://api.attio.com/v2/workspace_members",
			code: http.StatusOK,
			body: `{"data":[{"id":{"workspace_id":"a","workspace_member_id":null},"first_name":"Ada","last_name":"L","email_address":"ada@x.example","created_at":"x","access_level":"admin"}]}`,
		},
		{
			name: "self response missing the workspace name",
			url:  "https://api.attio.com/v2/self",
			code: http.StatusOK,
			body: `{"active":true,"client_id":"c","aud":"c","sub":"w","workspace_id":"w","workspace_slug":"s"}`,
		},
		// A field Attio did not return when this was written carries identity
		// the rewriters never see, and the value-repeat scan cannot catch it
		// because it matches nothing already rewritten.
		{
			name: "member carrying a field the sanitizer does not handle",
			url:  "https://api.attio.com/v2/workspace_members",
			code: http.StatusOK,
			body: `{"data":[{"id":{"workspace_id":"a","workspace_member_id":"b"},"first_name":"Ada","last_name":"L","email_address":"ada@x.example","created_at":"x","access_level":"admin","phone_number":"+33600000000"}]}`,
		},
		{
			name: "members response carrying an unhandled top-level field",
			url:  "https://api.attio.com/v2/workspace_members",
			code: http.StatusOK,
			body: `{"data":[{"id":{"workspace_id":"a","workspace_member_id":"b"},"first_name":"Ada","last_name":"L","email_address":"ada@x.example","created_at":"x","access_level":"admin"}],"billing_contact":"ada@x.example"}`,
		},
		{
			name: "self response carrying a field the sanitizer does not handle",
			url:  "https://api.attio.com/v2/self",
			code: http.StatusOK,
			body: `{"active":true,"client_id":"c","aud":"c","sub":"w","workspace_id":"w","workspace_slug":"s","workspace_name":"n","owner_email":"ada@x.example"}`,
		},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				i := attioInteraction(tc.url, tc.body)
				i.Response.Code = tc.code

				require.Error(t, sanitizeAttio(i))
			},
		)
	}
}

func attioInteraction(url, body string) *cassette.Interaction {
	i := &cassette.Interaction{}
	i.Request.URL = url
	i.Response.Code = http.StatusOK
	i.Response.Body = body
	i.Response.Headers = http.Header{}

	return i
}
