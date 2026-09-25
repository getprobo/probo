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
	"net/url"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
)

const mongoDBAtlasTestBaseURL = "https://cloud.mongodb.com/api/atlas/v2"

// Synthetic identity the sanitizer writes over the recorded organization, and
// which the replay assertions below then read back. The organization id is
// rewritten in request URLs as well as response bodies: the driver discovers
// it from GET /orgs and then builds every later URL from it, so a cassette
// that sanitized only the bodies would stop matching its own recorded
// requests.
const (
	mongoDBAtlasFixtureOrgID        = "000000000000000000000001"
	mongoDBAtlasFixtureOrgName      = "Example Organization"
	mongoDBAtlasFixtureUserID       = "000000000000000000000101"
	mongoDBAtlasFixtureUserEmail    = "member1@example.com"
	mongoDBAtlasFixtureSAClientID   = "mdb_sa_id_000000000000000000000201"
	mongoDBAtlasFixtureSAName       = "example-service-account"
	mongoDBAtlasFixtureSACreatedAt  = "2026-01-01T00:00:00Z"
	mongoDBAtlasFixtureSAExpiresAt  = "2027-01-01T00:00:00Z"
	mongoDBAtlasFixtureSALastUsedAt = "2026-01-02T03:04:05Z"
)

// TestMongoDBAtlasDriver replays a recording taken against a live Atlas
// organization, so the field names and the response envelope are the
// provider's own rather than this package's idea of them. The recorded
// organization holds one active owner and one service account; the mapping of
// the statuses and roles it does not contain is covered exhaustively by the
// table tests below, which need no fixture.
func TestMongoDBAtlasDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/mongodb_atlas", "MONGODB_ATLAS_TOKEN", sanitizeMongoDBAtlas)
	client := newVCRClient(rec, bearerAuth(os.Getenv("MONGODB_ATLAS_TOKEN")))
	driver := NewMongoDBAtlasDriver(client, mongoDBAtlasTestBaseURL)

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 2)

	owner := records[0]
	assert.Equal(t, mongoDBAtlasFixtureUserEmail, owner.Email)
	assert.Equal(t, mongoDBAtlasFixtureUserID, owner.ExternalID)
	assert.Equal(t, []string{"Organization Owner", "Project Owner"}, owner.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, owner.AccountType)
	require.NotNil(t, owner.Active)
	assert.True(t, *owner.Active)
	require.NotNil(t, owner.IsAdmin)
	assert.True(t, *owner.IsAdmin)
	require.NotNil(t, owner.LastLogin)
	require.NotNil(t, owner.CreatedAt)

	// Atlas exposes no MFA or authentication-method signal anywhere on this
	// API, on any version date. Asserting it keeps a future mapping change
	// from quietly inventing one.
	assert.Equal(t, coredata.MFAStatusUnknown, owner.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, owner.AuthMethod)

	// The service account is a non-human principal holding organization roles.
	// Listing only the human members would hide it from the review entirely.
	serviceAccount := records[1]
	assert.Equal(t, mongoDBAtlasFixtureSAClientID, serviceAccount.ExternalID)
	assert.Equal(t, mongoDBAtlasFixtureSAName, serviceAccount.FullName)
	assert.Empty(t, serviceAccount.Email)
	assert.Equal(t, []string{"Organization Read Only"}, serviceAccount.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, serviceAccount.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodServiceAccount, serviceAccount.AuthMethod)

	// Atlas states no enabled/disabled status for a service account, so Active
	// stays unknown rather than being inferred from secret expiry.
	assert.Nil(t, serviceAccount.Active)
	require.NotNil(t, serviceAccount.IsAdmin)
	assert.False(t, *serviceAccount.IsAdmin)

	// The most recent secret use is the only activity signal a service account
	// has; it stands in for a last login.
	require.NotNil(t, serviceAccount.LastLogin)
	assert.Equal(t, mongoDBAtlasFixtureSALastUsedAt, serviceAccount.LastLogin.UTC().Format(time.RFC3339))
}

func TestMongoDBAtlasNameResolver(t *testing.T) {
	t.Parallel()

	// A cassette of its own, not the driver's: two recorders writing one path
	// race and silently drop interactions when the package is re-recorded.
	rec := newRecorder(t, "testdata/mongodb_atlas_name", "MONGODB_ATLAS_TOKEN", sanitizeMongoDBAtlas)
	client := newVCRClient(rec, bearerAuth(os.Getenv("MONGODB_ATLAS_TOKEN")))
	resolver := NewMongoDBAtlasNameResolver(client, mongoDBAtlasTestBaseURL)

	name, err := resolver.ResolveInstanceName(context.Background())
	require.NoError(t, err)
	assert.Equal(t, mongoDBAtlasFixtureOrgName, name)
}

// TestMongoDBAtlasUserMapping covers the membership statuses and payload
// shapes the recorded organization does not contain. A pending invitation is
// a different variant of the response: it carries no name and no createdAt,
// and the invitation timestamp stands in for the latter.
func TestMongoDBAtlasUserMapping(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		user          mongoDBAtlasUser
		wantActive    bool
		wantIsAdmin   bool
		wantFullName  string
		wantRoles     []string
		wantCreatedAt string
	}{
		{
			name: "active owner",
			user: mongoDBAtlasUser{
				Username:            "owner@example.com",
				FirstName:           "Ada",
				LastName:            "Lovelace",
				OrgMembershipStatus: "ACTIVE",
				CreatedAt:           "2026-01-01T00:00:00Z",
				Roles:               mongoDBAtlasRoleAssignments{OrgRoles: []string{"ORG_OWNER"}},
			},
			wantActive:    true,
			wantIsAdmin:   true,
			wantFullName:  "Ada Lovelace",
			wantRoles:     []string{"Organization Owner"},
			wantCreatedAt: "2026-01-01T00:00:00Z",
		},
		{
			// A member who owns a project holds far more access than the
			// organization role alone shows, so the project role is listed
			// too — but it does not make them an organization admin.
			name: "member owning a project",
			user: mongoDBAtlasUser{
				Username:            "dev@example.com",
				OrgMembershipStatus: "ACTIVE",
				Roles:               mongoDBAtlasRoleAssignments{OrgRoles: []string{"ORG_MEMBER"}, GroupRoleAssignments: []mongoDBAtlasGroupRoleAssignment{{GroupID: "000000000000000000000301", GroupRoles: []string{"GROUP_OWNER"}}}},
			},
			wantActive:  true,
			wantIsAdmin: false,
			wantRoles:   []string{"Organization Member", "Project Owner"},
		},
		{
			// PENDING selects the other payload variant: no name, no
			// createdAt, and invitationCreatedAt in its place.
			name: "pending invitation",
			user: mongoDBAtlasUser{
				Username:            "invited@example.com",
				OrgMembershipStatus: "PENDING",
				InvitationCreatedAt: "2026-02-03T04:05:06Z",
				Roles:               mongoDBAtlasRoleAssignments{OrgRoles: []string{"ORG_MEMBER"}},
			},
			wantActive:    false,
			wantIsAdmin:   false,
			wantRoles:     []string{"Organization Member"},
			wantCreatedAt: "2026-02-03T04:05:06Z",
		},
		{
			// The driver asks for ACTIVE and PENDING only, so no other status
			// should arrive. If one ever does — a new Atlas status, or a change
			// to mongoDBAtlasReviewedStatuses — only ACTIVE may read as active.
			name: "any other status is inactive",
			user: mongoDBAtlasUser{
				Username:            "lapsed@example.com",
				OrgMembershipStatus: "INVITATION_EXPIRED",
				InvitationCreatedAt: "2026-02-03T04:05:06Z",
				Roles:               mongoDBAtlasRoleAssignments{OrgRoles: []string{"ORG_READ_ONLY"}},
			},
			wantActive:    false,
			wantIsAdmin:   false,
			wantRoles:     []string{"Organization Read Only"},
			wantCreatedAt: "2026-02-03T04:05:06Z",
		},
		{
			// A role Atlas ships after this map was written still has to read
			// as something: the console already offers an Organization Model
			// Owner the 2025-02-19 schema does not list.
			name: "unmapped role is humanized, not dropped",
			user: mongoDBAtlasUser{
				Username:            "ml@example.com",
				OrgMembershipStatus: "ACTIVE",
				Roles:               mongoDBAtlasRoleAssignments{OrgRoles: []string{"ORG_MODEL_OWNER"}, GroupRoleAssignments: []mongoDBAtlasGroupRoleAssignment{{GroupID: "000000000000000000000301", GroupRoles: []string{"GROUP_DATA_ACCESS_ADMIN"}}}},
			},
			wantActive:  true,
			wantIsAdmin: false,
			wantRoles:   []string{"Organization Model Owner", "Project Data Access Admin"},
		},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				got := mongoDBAtlasUserRecord(tc.user)

				require.NotNil(t, got.Active)
				assert.Equal(t, tc.wantActive, *got.Active)
				require.NotNil(t, got.IsAdmin)
				assert.Equal(t, tc.wantIsAdmin, *got.IsAdmin)
				assert.Equal(t, tc.wantFullName, got.FullName)
				assert.Equal(t, tc.wantRoles, got.Roles)
				assert.Equal(t, tc.user.Username, got.Email)

				if tc.wantCreatedAt == "" {
					assert.Nil(t, got.CreatedAt)

					return
				}

				require.NotNil(t, got.CreatedAt)
				assert.Equal(t, tc.wantCreatedAt, got.CreatedAt.UTC().Format(time.RFC3339))
			},
		)
	}
}

// TestMongoDBAtlasRoleDeduplication pins that the same role held on several
// projects is listed once, and that a project role identical to an
// organization role does not appear twice.
func TestMongoDBAtlasRoleDeduplication(t *testing.T) {
	t.Parallel()

	user := mongoDBAtlasUser{
		OrgMembershipStatus: "ACTIVE",
		Roles: mongoDBAtlasRoleAssignments{
			OrgRoles: []string{"ORG_MEMBER"},
			GroupRoleAssignments: []mongoDBAtlasGroupRoleAssignment{
				{GroupID: "000000000000000000000301", GroupRoles: []string{"GROUP_OWNER"}},
				{GroupID: "000000000000000000000302", GroupRoles: []string{"GROUP_OWNER"}},
				{GroupID: "000000000000000000000303", GroupRoles: []string{"GROUP_READ_ONLY"}},
			},
		},
	}

	assert.Equal(
		t,
		[]string{"Organization Member", "Project Owner", "Project Read Only"},
		mongoDBAtlasUserRoles(user),
	)
}

func TestMongoDBAtlasCollectionURL(t *testing.T) {
	t.Parallel()

	users, err := mongoDBAtlasCollectionURL(
		mongoDBAtlasTestBaseURL,
		mongoDBAtlasFixtureOrgID,
		"users",
		2,
		url.Values{"orgMembershipStatuses": mongoDBAtlasReviewedStatuses},
	)
	require.NoError(t, err)

	// The reviewed population is stated on the wire rather than left to the
	// endpoint's default, which MongoDB is free to change under us.
	assert.Equal(
		t,
		mongoDBAtlasTestBaseURL+"/orgs/"+mongoDBAtlasFixtureOrgID+
			"/users?itemsPerPage=500&orgMembershipStatuses=ACTIVE&orgMembershipStatuses=PENDING&pageNum=2",
		users,
	)

	serviceAccounts, err := mongoDBAtlasCollectionURL(
		mongoDBAtlasTestBaseURL,
		mongoDBAtlasFixtureOrgID,
		"serviceAccounts",
		1,
		nil,
	)
	require.NoError(t, err)
	assert.Equal(
		t,
		mongoDBAtlasTestBaseURL+"/orgs/"+mongoDBAtlasFixtureOrgID+"/serviceAccounts?itemsPerPage=500&pageNum=1",
		serviceAccounts,
	)
}

// TestMongoDBAtlasAPIKeyMapping covers the programmatic API keys, the older
// non-human principal. The recorded organization holds none, so the cassette
// cannot exercise this.
func TestMongoDBAtlasAPIKeyMapping(t *testing.T) {
	t.Parallel()

	owner := mongoDBAtlasAPIKeyRecord(mongoDBAtlasAPIKey{
		ID:        "000000000000000000000401",
		Desc:      "terraform",
		PublicKey: "abcdefgh",
		Roles: []mongoDBAtlasAPIKeyRole{
			{OrgID: mongoDBAtlasFixtureOrgID, RoleName: "ORG_OWNER"},
			{GroupID: "000000000000000000000301", RoleName: "GROUP_READ_ONLY"},
		},
	})

	assert.Equal(t, "000000000000000000000401", owner.ExternalID)
	assert.Equal(t, "terraform", owner.FullName)
	assert.Empty(t, owner.Email)
	assert.Equal(t, []string{"Organization Owner", "Project Read Only"}, owner.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, owner.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodAPIKey, owner.AuthMethod)
	require.NotNil(t, owner.IsAdmin)
	assert.True(t, *owner.IsAdmin)

	// Atlas states no enabled flag, no creation date and no last use for a
	// programmatic API key.
	assert.Nil(t, owner.Active)
	assert.Nil(t, owner.CreatedAt)
	assert.Nil(t, owner.LastLogin)

	// A key with no description is identified by its public half rather than
	// showing a blank row.
	unnamed := mongoDBAtlasAPIKeyRecord(mongoDBAtlasAPIKey{ID: "x", PublicKey: "abcdefgh"})
	assert.Equal(t, "abcdefgh", unnamed.FullName)
	require.NotNil(t, unnamed.IsAdmin)
	assert.False(t, *unnamed.IsAdmin)
}

// TestSanitizeMongoDBAtlasFailsClosed pins the property the cassette's safety
// rests on: a field this sanitizer has not been taught about aborts the
// recording instead of being written through. Re-recording is the mandated way
// to maintain a cassette, so a fail-open sanitizer would commit whatever the
// next organization happens to carry.
func TestSanitizeMongoDBAtlasFailsClosed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		url  string
		body string
	}{
		{
			name: "unknown field on a user",
			url:  "https://cloud.mongodb.com/api/atlas/v2/orgs/" + mongoDBAtlasFixtureOrgID + "/users?pageNum=1",
			body: `{"results":[{"id":"a","username":"real@example.org","surpriseField":"real@example.org"}]}`,
		},
		{
			name: "unknown field on a service account",
			url:  "https://cloud.mongodb.com/api/atlas/v2/orgs/" + mongoDBAtlasFixtureOrgID + "/serviceAccounts?pageNum=1",
			body: `{"results":[{"clientId":"a","name":"b","surpriseField":"c"}]}`,
		},
		{
			name: "unknown field on the envelope",
			url:  "https://cloud.mongodb.com/api/atlas/v2/orgs",
			body: `{"results":[],"surpriseField":"c"}`,
		},
		{
			name: "unrecognised endpoint",
			url:  "https://cloud.mongodb.com/api/atlas/v2/orgs/" + mongoDBAtlasFixtureOrgID + "/invoices",
			body: `{"results":[]}`,
		},
		{
			// A substring match would have treated this as the roster.
			name: "unrelated path whose query mentions a known collection",
			url:  "https://cloud.mongodb.com/api/atlas/v2/orgs/" + mongoDBAtlasFixtureOrgID + "/invoices?next=/users",
			body: `{"results":[]}`,
		},
		{
			// A suffix match would have accepted this.
			name: "known collection path behind an unrelated prefix",
			url:  "https://evil.example/proxy/api/atlas/v2/orgs/" + mongoDBAtlasFixtureOrgID + "/users",
			body: `{"results":[]}`,
		},
		{
			name: "deeper path below a known collection",
			url:  "https://cloud.mongodb.com/api/atlas/v2/orgs/" + mongoDBAtlasFixtureOrgID + "/users/aaaaaaaaaaaaaaaaaaaaaaaa/settings",
			body: `{"results":[]}`,
		},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				i := &cassette.Interaction{}
				i.Request.URL = tc.url
				i.Response.Code = http.StatusOK
				i.Response.Body = tc.body

				assert.Error(t, sanitizeMongoDBAtlas(i))
			},
		)
	}
}

// TestSanitizeMongoDBAtlasRewritesIdentity checks the happy path actually
// removes what it collected, rather than only not erroring.
func TestSanitizeMongoDBAtlasRewritesIdentity(t *testing.T) {
	t.Parallel()

	i := &cassette.Interaction{}
	i.Request.URL = "https://cloud.mongodb.com/api/atlas/v2/orgs/aaaaaaaaaaaaaaaaaaaaaaaa/users?pageNum=1"
	i.Response.Code = http.StatusOK
	i.Response.Body = `{"links":[{"href":"https://cloud.mongodb.com/api/atlas/v2/orgs/aaaaaaaaaaaaaaaaaaaaaaaa/users"}],` +
		`"results":[{"links":[{"href":"https://cloud.mongodb.com/api/atlas/v2/orgs/aaaaaaaaaaaaaaaaaaaaaaaa/users/bbbbbbbbbbbbbbbbbbbbbbbb"}],` +
		`"id":"bbbbbbbbbbbbbbbbbbbbbbbb","username":"real.person@acme.example",` +
		`"firstName":"Real","lastName":"Person","country":"FR","mobileNumber":"+33123456789",` +
		`"orgMembershipStatus":"ACTIVE","teamIds":["cccccccccccccccccccccccc"],` +
		`"roles":{"orgRoles":["ORG_OWNER"],"groupRoleAssignments":[{"groupId":"dddddddddddddddddddddddd","groupRoles":["GROUP_OWNER"]}]}}]}`

	require.NoError(t, sanitizeMongoDBAtlas(i))

	for _, leaked := range []string{
		"real.person@acme.example", "Real", "Person", "+33123456789",
		"aaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbb",
		"cccccccccccccccccccccccc", "dddddddddddddddddddddddd",
	} {
		assert.NotContains(t, i.Response.Body, leaked)
		assert.NotContains(t, i.Request.URL, leaked)
	}

	// Per-result self-links carry the resource id and are dropped entirely.
	assert.NotContains(t, i.Response.Body, `"results":[{"links"`)

	// The shape the driver reads is preserved.
	assert.Contains(t, i.Response.Body, `"orgRoles":["ORG_OWNER"]`)
	assert.Contains(t, i.Response.Body, mongoDBAtlasFixtureUserEmail)
}

// mongoDBAtlasOrgIDInURL matches the organization id Atlas puts in a path. The
// driver reads that id out of one response and builds the next request from it,
// so it is rewritten in URLs as well as bodies or the cassette stops matching
// its own recorded requests.
var mongoDBAtlasOrgIDInURL = regexp.MustCompile(`/orgs/[0-9a-fA-F]{24}`)

// sanitizeMongoDBAtlas rewrites the recorded organization's identity into the
// synthetic fixture above.
//
// It is fail-closed, the way sanitizeAttio is: an unrecognised interaction or
// an unrecognised field aborts the recording rather than passing the value
// through. Re-recording is the mandated way to maintain a cassette, and the
// organization recorded next may hold the things this one happens not to — a
// pending invitation carrying inviterUsername, a service account with a
// non-null owner object, a populated teamIds. A list of fields to scrub would
// commit those verbatim to a public repository the first time one appeared.
func sanitizeMongoDBAtlas(i *cassette.Interaction) error {
	if i.Response.Code != http.StatusOK {
		return fmt.Errorf("refusing to sanitize mongodb atlas response with status %d", i.Response.Code)
	}

	i.Request.URL = mongoDBAtlasOrgIDInURL.ReplaceAllString(
		i.Request.URL,
		"/orgs/"+mongoDBAtlasFixtureOrgID,
	)

	// Dispatch on the whole parsed path, not on a substring of the URL and not
	// on a suffix of it: a query value mentioning "/users", and equally a
	// longer path ending in one, must not be taken for the roster.
	parsed, err := url.Parse(i.Request.URL)
	if err != nil {
		return fmt.Errorf("cannot parse recorded mongodb atlas URL: %w", err)
	}

	switch parsed.Path {
	case mongoDBAtlasCollectionPath("serviceAccounts"):
		return sanitizeMongoDBAtlasBody(i, sanitizeMongoDBAtlasServiceAccounts)
	case mongoDBAtlasCollectionPath("apiKeys"):
		return sanitizeMongoDBAtlasBody(i, sanitizeMongoDBAtlasAPIKeys)
	case mongoDBAtlasCollectionPath("users"):
		return sanitizeMongoDBAtlasBody(i, sanitizeMongoDBAtlasUsers)
	case mongoDBAtlasOrgsListPath():
		return sanitizeMongoDBAtlasBody(i, sanitizeMongoDBAtlasOrgs)
	default:
		return fmt.Errorf("refusing to sanitize an unrecognised mongodb atlas interaction")
	}
}

// mongoDBAtlasOrgsListPath is the full path of GET /orgs, built from the same
// base URL the driver under test is constructed with.
func mongoDBAtlasOrgsListPath() string {
	base, err := url.Parse(mongoDBAtlasTestBaseURL)
	if err != nil {
		return ""
	}

	return path.Join(base.Path, mongoDBAtlasOrgsPath)
}

// mongoDBAtlasCollectionPath is the full path of one collection under the
// fixture organization, which the caller has already rewritten into the URL.
func mongoDBAtlasCollectionPath(collection string) string {
	return path.Join(mongoDBAtlasOrgsListPath(), mongoDBAtlasFixtureOrgID, collection)
}

// sanitizeMongoDBAtlasBody decodes the recorded envelope into a raw field map,
// hands each result to the endpoint's rewriter, and writes it back. It edits
// the decoded JSON rather than re-marshalling the driver's own structs, which
// would drop whatever the driver does not model — the drift this cassette
// exists to catch.
func sanitizeMongoDBAtlasBody(
	i *cassette.Interaction,
	rewrite func([]map[string]json.RawMessage) ([]string, error),
) error {
	var body map[string]json.RawMessage
	if err := json.Unmarshal([]byte(i.Response.Body), &body); err != nil {
		return fmt.Errorf("cannot decode recorded mongodb atlas response: %w", err)
	}

	if err := mongoDBAtlasRejectUnknownFields(
		body,
		[]string{"links", "results", "totalCount"},
		"response envelope",
	); err != nil {
		return err
	}

	var results []map[string]json.RawMessage
	if raw, ok := body["results"]; ok {
		if err := json.Unmarshal(raw, &results); err != nil {
			return fmt.Errorf("cannot decode recorded mongodb atlas results: %w", err)
		}
	}

	// An empty page is expected: the driver's pagination ends on one.
	recorded, err := rewrite(results)
	if err != nil {
		return err
	}

	if _, ok := body["results"]; ok {
		encoded, err := json.Marshal(results)
		if err != nil {
			return fmt.Errorf("cannot re-encode mongodb atlas results: %w", err)
		}

		body["results"] = encoded
	}

	sanitized, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("cannot re-encode mongodb atlas response: %w", err)
	}

	// Atlas echoes the requested URL back in links[].href, organization id and
	// all.
	out := mongoDBAtlasOrgIDInURL.ReplaceAllString(string(sanitized), "/orgs/"+mongoDBAtlasFixtureOrgID)

	// A second guard behind the declared-field check, for identity rewritten in
	// one place and repeated somewhere the rewriter does not treat as identity.
	for _, value := range recorded {
		if value != "" && strings.Contains(out, value) {
			return fmt.Errorf("sanitized mongodb atlas response still contains a recorded identity value")
		}
	}

	replaceCassetteBody(i, out)

	return nil
}

func sanitizeMongoDBAtlasOrgs(orgs []map[string]json.RawMessage) ([]string, error) {
	var recorded []string

	for _, org := range orgs {
		if err := mongoDBAtlasRejectUnknownFields(
			org,
			[]string{"id", "isDeleted", "links", "name", "skipDefaultAlertsSettings"},
			"organization",
		); err != nil {
			return nil, err
		}

		recorded = append(recorded, mongoDBAtlasRecorded(org, "name")...)

		delete(org, "links")

		mongoDBAtlasSet(org, "id", mongoDBAtlasFixtureOrgID)
		mongoDBAtlasSet(org, "name", mongoDBAtlasFixtureOrgName)
	}

	return recorded, nil
}

func sanitizeMongoDBAtlasUsers(users []map[string]json.RawMessage) ([]string, error) {
	var recorded []string

	for _, user := range users {
		// A self-link carries the resource's real id, which the org-id rewrite
		// does not touch and the post-rewrite guard does not look for. The
		// driver never reads them, so they go rather than get patched.
		delete(user, "links")

		if err := mongoDBAtlasRejectUnknownFields(
			user,
			[]string{
				"country", "createdAt", "firstName", "id", "invitationCreatedAt",
				"inviterUsername", "lastAuth", "lastName", "links", "mobileNumber",
				"orgMembershipStatus", "roles", "teamIds", "username",
			},
			"organization user",
		); err != nil {
			return nil, err
		}

		recorded = append(recorded, mongoDBAtlasRecorded(
			user, "username", "firstName", "lastName", "inviterUsername", "country", "mobileNumber",
		)...)

		mongoDBAtlasSet(user, "id", mongoDBAtlasFixtureUserID)
		mongoDBAtlasSet(user, "username", mongoDBAtlasFixtureUserEmail)
		mongoDBAtlasSetIfPresent(user, "firstName", "Member")
		mongoDBAtlasSetIfPresent(user, "lastName", "One")
		mongoDBAtlasSetIfPresent(user, "inviterUsername", "inviter@example.com")

		// Personal contact details the driver never reads. Blanked rather than
		// faked: an obviously empty value cannot be mistaken for real data.
		mongoDBAtlasSetIfPresent(user, "country", "")
		mongoDBAtlasSetIfPresent(user, "mobileNumber", "")

		// Team ids name real teams in the recorded organization.
		if _, ok := user["teamIds"]; ok {
			user["teamIds"] = json.RawMessage(`[]`)
		}

		if err := mongoDBAtlasSanitizeRoleAssignments(user); err != nil {
			return nil, err
		}
	}

	return recorded, nil
}

// mongoDBAtlasSanitizeRoleAssignments rewrites the project ids inside a user's
// roles. The driver reads only groupRoles, so the ids cost the fixture nothing.
func mongoDBAtlasSanitizeRoleAssignments(user map[string]json.RawMessage) error {
	raw, ok := user["roles"]
	if !ok {
		return nil
	}

	var roles map[string]json.RawMessage
	if err := json.Unmarshal(raw, &roles); err != nil {
		return fmt.Errorf("cannot decode recorded mongodb atlas user roles: %w", err)
	}

	if err := mongoDBAtlasRejectUnknownFields(
		roles,
		[]string{"groupRoleAssignments", "orgRoles"},
		"user roles",
	); err != nil {
		return err
	}

	assignmentsRaw, ok := roles["groupRoleAssignments"]
	if !ok {
		return nil
	}

	var assignments []map[string]json.RawMessage
	if err := json.Unmarshal(assignmentsRaw, &assignments); err != nil {
		return fmt.Errorf("cannot decode recorded mongodb atlas group role assignments: %w", err)
	}

	for idx, assignment := range assignments {
		if err := mongoDBAtlasRejectUnknownFields(
			assignment,
			[]string{"groupId", "groupRoles"},
			"group role assignment",
		); err != nil {
			return err
		}

		mongoDBAtlasSet(assignment, "groupId", fmt.Sprintf("%024d", 301+idx))
	}

	encoded, err := json.Marshal(assignments)
	if err != nil {
		return fmt.Errorf("cannot re-encode mongodb atlas group role assignments: %w", err)
	}

	roles["groupRoleAssignments"] = encoded

	rolesEncoded, err := json.Marshal(roles)
	if err != nil {
		return fmt.Errorf("cannot re-encode mongodb atlas user roles: %w", err)
	}

	user["roles"] = rolesEncoded

	return nil
}

func sanitizeMongoDBAtlasServiceAccounts(serviceAccounts []map[string]json.RawMessage) ([]string, error) {
	var recorded []string

	for _, sa := range serviceAccounts {
		if err := mongoDBAtlasRejectUnknownFields(
			sa,
			[]string{
				"clientId", "createdAt", "description", "isSystemManaged",
				"name", "owner", "roles", "secrets",
			},
			"service account",
		); err != nil {
			return nil, err
		}

		recorded = append(recorded, mongoDBAtlasRecorded(sa, "clientId", "name", "description")...)

		mongoDBAtlasSet(sa, "clientId", mongoDBAtlasFixtureSAClientID)
		mongoDBAtlasSet(sa, "name", mongoDBAtlasFixtureSAName)

		// The whole timeline is made synthetic so it stays self-consistent: a
		// recorded createdAt with a fixture lastUsedAt would put the last use
		// before the account existed.
		mongoDBAtlasSetIfPresent(sa, "createdAt", mongoDBAtlasFixtureSACreatedAt)
		mongoDBAtlasSetIfPresent(sa, "description", "Example service account")

		// owner is null in the recording, and an object carrying the creating
		// user's identity when it is not.
		if _, ok := sa["owner"]; ok {
			sa["owner"] = json.RawMessage(`null`)
		}

		if err := mongoDBAtlasSanitizeSecrets(sa); err != nil {
			return nil, err
		}
	}

	return recorded, nil
}

func mongoDBAtlasSanitizeSecrets(sa map[string]json.RawMessage) error {
	raw, ok := sa["secrets"]
	if !ok {
		return nil
	}

	var secrets []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &secrets); err != nil {
		return fmt.Errorf("cannot decode recorded mongodb atlas secrets: %w", err)
	}

	for _, secret := range secrets {
		if err := mongoDBAtlasRejectUnknownFields(
			secret,
			[]string{"createdAt", "expiresAt", "id", "lastUsedAt", "maskedSecretValue"},
			"service account secret",
		); err != nil {
			return err
		}

		mongoDBAtlasSet(secret, "id", "000000000000000000000202")
		mongoDBAtlasSetIfPresent(secret, "createdAt", mongoDBAtlasFixtureSACreatedAt)
		mongoDBAtlasSetIfPresent(secret, "expiresAt", mongoDBAtlasFixtureSAExpiresAt)

		// maskedSecretValue carries the live secret's trailing characters.
		mongoDBAtlasSetIfPresent(secret, "maskedSecretValue", "mdb_sa_sk_...0000")

		// Rewritten only when Atlas actually sent it, so the LastLogin
		// assertion tests the provider's payload rather than this sanitizer.
		mongoDBAtlasSetIfPresent(secret, "lastUsedAt", mongoDBAtlasFixtureSALastUsedAt)
	}

	encoded, err := json.Marshal(secrets)
	if err != nil {
		return fmt.Errorf("cannot re-encode mongodb atlas secrets: %w", err)
	}

	sa["secrets"] = encoded

	return nil
}

func sanitizeMongoDBAtlasAPIKeys(apiKeys []map[string]json.RawMessage) ([]string, error) {
	var recorded []string

	for idx, key := range apiKeys {
		delete(key, "links")

		if err := mongoDBAtlasRejectUnknownFields(
			key,
			[]string{"desc", "id", "links", "privateKey", "publicKey", "roles"},
			"organization api key",
		); err != nil {
			return nil, err
		}

		recorded = append(recorded, mongoDBAtlasRecorded(key, "desc", "publicKey", "privateKey")...)

		mongoDBAtlasSet(key, "id", fmt.Sprintf("%024d", 401+idx))
		mongoDBAtlasSetIfPresent(key, "desc", "Example API key")
		mongoDBAtlasSetIfPresent(key, "publicKey", "aaaaaaaa")
		mongoDBAtlasSetIfPresent(key, "privateKey", "********-****-****-redacted")

		if err := mongoDBAtlasSanitizeAPIKeyRoles(key); err != nil {
			return nil, err
		}
	}

	return recorded, nil
}

func mongoDBAtlasSanitizeAPIKeyRoles(key map[string]json.RawMessage) error {
	raw, ok := key["roles"]
	if !ok {
		return nil
	}

	var roles []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &roles); err != nil {
		return fmt.Errorf("cannot decode recorded mongodb atlas api key roles: %w", err)
	}

	for _, role := range roles {
		if err := mongoDBAtlasRejectUnknownFields(
			role,
			[]string{"groupId", "orgId", "roleName"},
			"api key role",
		); err != nil {
			return err
		}

		mongoDBAtlasSetIfPresent(role, "orgId", mongoDBAtlasFixtureOrgID)
		mongoDBAtlasSetIfPresent(role, "groupId", "000000000000000000000301")
	}

	encoded, err := json.Marshal(roles)
	if err != nil {
		return fmt.Errorf("cannot re-encode mongodb atlas api key roles: %w", err)
	}

	key["roles"] = encoded

	return nil
}

// mongoDBAtlasRejectUnknownFields fails the recording when the provider sends a
// field this sanitizer has not been taught about, so a new identity-bearing
// field cannot reach a committed cassette unreviewed.
func mongoDBAtlasRejectUnknownFields(
	object map[string]json.RawMessage,
	known []string,
	what string,
) error {
	for field := range object {
		if !slices.Contains(known, field) {
			return fmt.Errorf("recorded mongodb atlas %s carries an unknown field %q", what, field)
		}
	}

	return nil
}

// mongoDBAtlasRecorded collects the string values about to be overwritten, for
// the post-rewrite check that none of them survived elsewhere in the body.
func mongoDBAtlasRecorded(object map[string]json.RawMessage, fields ...string) []string {
	var values []string

	for _, field := range fields {
		raw, ok := object[field]
		if !ok {
			continue
		}

		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			continue
		}

		if value != "" {
			values = append(values, value)
		}
	}

	return values
}

func mongoDBAtlasSet(object map[string]json.RawMessage, field, value string) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return
	}

	object[field] = encoded
}

func mongoDBAtlasSetIfPresent(object map[string]json.RawMessage, field, value string) {
	if _, ok := object[field]; !ok {
		return
	}

	mongoDBAtlasSet(object, field, value)
}
