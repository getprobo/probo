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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
)

// sanitizeTwingateUsers replaces member identity in a recorded response with
// synthetic values.
//
// It walks the decoded JSON rather than re-marshalling twingateUsersResponse,
// which would rewrite the body as a serialization of the type under test and
// drop whatever the driver does not model. Decoding into map[string]any makes
// the walk safe: the nested maps are references, so a rewrite reaches the body
// without a repack chain. Numbers decode as json.Number so re-encoding cannot
// reshape one.
//
// role, isAdmin, state, type and createdAt are left as recorded: they are the
// authority and lifecycle signals the assertions below are about, and none of
// them identifies a person.
//
// The hook runs on save, so a recording run still sees live data and its
// assertions fail. Re-record, then run again without TWINGATE_API_KEY.
func sanitizeTwingateUsers(i *cassette.Interaction) error {
	if i.Response.Code != http.StatusOK {
		return fmt.Errorf("refusing to sanitize twingate response with status %d", i.Response.Code)
	}

	decoder := json.NewDecoder(strings.NewReader(i.Response.Body))
	decoder.UseNumber()

	var body map[string]any
	if err := decoder.Decode(&body); err != nil {
		return fmt.Errorf("cannot decode recorded twingate response: %w", err)
	}

	// Twingate reports a rejected query with 200 and an errors array, so a
	// recording that failed would otherwise be saved as a usable cassette.
	if _, ok := body["errors"]; ok {
		return fmt.Errorf("refusing to sanitize a twingate response carrying graphql errors")
	}

	data, ok := body["data"].(map[string]any)
	if !ok {
		return fmt.Errorf("recorded twingate response has no data object")
	}

	users, ok := data["users"].(map[string]any)
	if !ok {
		return fmt.Errorf("recorded twingate response has no users object")
	}

	edges, ok := users["edges"].([]any)
	if !ok {
		return fmt.Errorf("recorded twingate response has no edges array")
	}

	if len(edges) == 0 {
		return fmt.Errorf("recorded twingate response lists no users")
	}

	for idx, rawEdge := range edges {
		edge, ok := rawEdge.(map[string]any)
		if !ok {
			return fmt.Errorf("recorded twingate edge %d is not an object", idx)
		}

		node, ok := edge["node"].(map[string]any)
		if !ok {
			return fmt.Errorf("recorded twingate edge %d has no node object", idx)
		}

		// Decoded rather than merely present: rewriting a value that came back
		// as null or a number would paper over the struct mismatch this
		// cassette exposes.
		replacements := map[string]string{
			"id":        fmt.Sprintf("twingate-user-%04d", idx+1),
			"email":     fmt.Sprintf("member%d@example.com", idx+1),
			"firstName": fmt.Sprintf("Member%d", idx+1),
			"lastName":  "Example",
		}

		for field, replacement := range replacements {
			value, ok := node[field]
			if !ok {
				return fmt.Errorf("recorded twingate user %d has no %s field", idx, field)
			}

			if _, ok := value.(string); !ok {
				return fmt.Errorf("recorded twingate user %d has a non-string %s", idx, field)
			}

			node[field] = replacement
		}
	}

	// The cursor is derived from the tenant's own record ordering; it is not
	// identity, but it is replayed verbatim so it must stay a valid string.
	if pageInfo, ok := users["pageInfo"].(map[string]any); ok {
		if _, ok := pageInfo["endCursor"].(string); ok {
			pageInfo["endCursor"] = "YXJyYXljb25uZWN0aW9uOjA="
		}
	}

	sanitized, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("cannot re-encode sanitized twingate response: %w", err)
	}

	replaceCassetteBody(i, string(sanitized))

	return nil
}

func TestTwingateDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/twingate", "TWINGATE_API_KEY", sanitizeTwingateUsers)
	// Twingate authenticates via the X-API-KEY header, not Authorization.
	client := newVCRClientWithHeader(rec, "X-API-KEY", os.Getenv("TWINGATE_API_KEY"))

	driver := NewTwingateDriver(client, "https://probo.twingate.com/api/graphql/")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)

	// Cassette recorded live, then anonymized: one active ADMIN, created
	// manually rather than synced from an identity provider.
	admin := records[0]
	assert.Equal(t, "twingate-user-0001", admin.ExternalID)
	assert.Equal(t, "member1@example.com", admin.Email)
	assert.Equal(t, "Member1 Example", admin.FullName)
	assert.Equal(t, []string{"Admin"}, admin.Roles)
	assert.Equal(t, new(true), admin.IsAdmin)
	require.NotNil(t, admin.Active)
	assert.True(t, *admin.Active)
	require.NotNil(t, admin.CreatedAt)
	assert.Equal(t, coredata.MFAStatusUnknown, admin.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, admin.AccountType)
}

func TestTwingateEndpoint(t *testing.T) {
	t.Parallel()

	endpoint, err := TwingateEndpoint("acme")
	require.NoError(t, err)
	assert.Equal(t, "https://acme.twingate.com/api/graphql/", endpoint)

	// Surrounding whitespace is the shape a pasted setting arrives in.
	trimmed, err := TwingateEndpoint("  acme  ")
	require.NoError(t, err)
	assert.Equal(t, "https://acme.twingate.com/api/graphql/", trimmed)

	// The network lands in the host, so anything that is not a DNS label is
	// refused here rather than reaching a URL.
	for _, invalid := range []string{
		"",
		"-acme",
		"acme-",
		"acme.evil.com",
		"acme/../evil",
		"acme:8080",
		"ac me",
		"acme@evil",
		strings.Repeat("a", 64),
	} {
		_, err := TwingateEndpoint(invalid)
		require.Errorf(t, err, "expected %q to be refused", invalid)
	}

	// 63 characters is the longest a DNS label may be.
	_, err = TwingateEndpoint(strings.Repeat("a", 63))
	require.NoError(t, err)
}

func TestTwingateRoles(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"Admin"}, twingateRoles(twingateUser{Role: twingateRoleAdmin}))
	assert.Equal(t, []string{"DevOps"}, twingateRoles(twingateUser{Role: twingateRoleDevops}))
	assert.Equal(t, []string{"Support"}, twingateRoles(twingateUser{Role: twingateRoleSupport}))
	assert.Equal(t, []string{"Helpdesk"}, twingateRoles(twingateUser{Role: twingateRoleHelpdesk}))
	assert.Equal(t, []string{"Access Reviewer"}, twingateRoles(twingateUser{Role: twingateRoleAccessReviewer}))
	assert.Equal(t, []string{"Billing"}, twingateRoles(twingateUser{Role: twingateRoleBilling}))
	assert.Equal(t, []string{"Member"}, twingateRoles(twingateUser{Role: twingateRoleMember}))
	// An unrecognised future role is preserved rather than dropped.
	assert.Equal(t, []string{"AUDITOR"}, twingateRoles(twingateUser{Role: "AUDITOR"}))
	// No role at all is no role, never a nil slice.
	assert.Equal(t, []string{}, twingateRoles(twingateUser{}))
}

func TestTwingateActive(t *testing.T) {
	t.Parallel()

	assert.Equal(t, new(true), twingateActive(twingateUser{State: twingateStateActive}))
	// An invited user has not joined and a disabled one has been revoked;
	// neither is a live seat.
	assert.Equal(t, new(false), twingateActive(twingateUser{State: twingateStatePending}))
	assert.Equal(t, new(false), twingateActive(twingateUser{State: twingateStateDisabled}))
	// An unrecognised future state reports no signal rather than guessing.
	assert.Nil(t, twingateActive(twingateUser{State: "ARCHIVED"}))
	assert.Nil(t, twingateActive(twingateUser{}))
}

func TestTwingateFullName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Ada Lovelace", twingateFullName(twingateUser{FirstName: "Ada", LastName: "Lovelace"}, "x@example.com"))
	assert.Equal(t, "Ada", twingateFullName(twingateUser{FirstName: " Ada "}, "x@example.com"))
	assert.Equal(t, "Lovelace", twingateFullName(twingateUser{LastName: "Lovelace"}, "x@example.com"))
	// A nameless record falls back to the email rather than reviewing blank.
	assert.Equal(t, "x@example.com", twingateFullName(twingateUser{}, "x@example.com"))
}

// twingateStub serves the given response bodies in order, recording the
// decoded `after` variable of each request.
func twingateStub(t *testing.T, bodies ...string) (*httptest.Server, *[]any) {
	t.Helper()

	cursors := &[]any{}
	calls := 0

	// t.Errorf rather than require: this runs on the server's goroutine, where
	// require's FailNow unwinds only that goroutine, leaving the handler dead
	// without a response and the test hanging on it rather than failing.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Variables map[string]any `json:"variables"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("stub could not decode the driver's request: %v", err)
			http.Error(w, "{}", http.StatusBadRequest)

			return
		}

		*cursors = append(*cursors, req.Variables["after"])

		if calls >= len(bodies) {
			t.Errorf("driver asked for page %d but the stub serves %d", calls+1, len(bodies))
			http.Error(w, "{}", http.StatusInternalServerError)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(bodies[calls]))
		calls++
	}))

	t.Cleanup(server.Close)

	return server, cursors
}

func twingatePage(id, email string, hasNext bool, endCursor string) string {
	return fmt.Sprintf(`{"data":{"users":{
      "pageInfo":{"hasNextPage":%t,"endCursor":%q},
      "edges":[{"node":{"id":%q,"email":%q,"firstName":"A","lastName":"B",
        "role":"MEMBER","isAdmin":false,"state":"ACTIVE","type":"MANUAL",
        "createdAt":"2026-01-01T00:00:00Z"}}]}}}`, hasNext, endCursor, id, email)
}

func TestTwingateDriverPaginates(t *testing.T) {
	t.Parallel()

	server, cursors := twingateStub(t,
		twingatePage("u1", "one@example.com", true, "cursor-1"),
		twingatePage("u2", "two@example.com", false, ""),
	)

	driver := NewTwingateDriver(server.Client(), server.URL)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 2)

	assert.Equal(t, "one@example.com", records[0].Email)
	assert.Equal(t, "two@example.com", records[1].Email)

	// The first request opens the connection, the second resumes from the
	// cursor the first returned.
	require.Len(t, *cursors, 2)
	assert.Nil(t, (*cursors)[0])
	assert.Equal(t, "cursor-1", (*cursors)[1])
}

func TestTwingateDriverRefusesNextPageWithoutCursor(t *testing.T) {
	t.Parallel()

	// Twingate says there is more and will not say where to resume. Returning
	// what we have would be a short roster with no error, and a member missing
	// from a campaign is reviewed by nobody.
	server, _ := twingateStub(t, twingatePage("u1", "one@example.com", true, ""))

	driver := NewTwingateDriver(server.Client(), server.URL)
	_, err := driver.ListAccounts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no cursor")
}

func TestTwingateAuthMethod(t *testing.T) {
	t.Parallel()

	// SYNCED means an identity provider provisioned the account.
	assert.Equal(
		t,
		coredata.AccessReviewEntryAuthMethodSSO,
		twingateAuthMethod(twingateUser{Type: twingateTypeSynced}),
	)
	// MANUAL says how the record was created, not how the person signs in.
	assert.Equal(
		t,
		coredata.AccessReviewEntryAuthMethodUnknown,
		twingateAuthMethod(twingateUser{Type: "MANUAL"}),
	)
	assert.Equal(
		t,
		coredata.AccessReviewEntryAuthMethodUnknown,
		twingateAuthMethod(twingateUser{}),
	)
}

func TestTwingateNetworkCanonicalises(t *testing.T) {
	t.Parallel()

	// One network must not be storable as three different-looking sources.
	for _, raw := range []string{"acme", "ACME", "  Acme  "} {
		network, err := TwingateNetwork(raw)
		require.NoErrorf(t, err, "%q", raw)
		assert.Equal(t, "acme", network)
	}

	_, err := TwingateNetwork("acme.evil.example")
	require.Error(t, err)
}
