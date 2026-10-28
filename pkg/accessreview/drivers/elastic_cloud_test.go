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
	"io"
	"net/http"
	"net/url"
	"os"
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

const elasticCloudTestBaseURL = "https://api.elastic-cloud.com/api/v1"

// Synthetic identity the sanitizer writes over the recorded organization. The
// organization id is rewritten in request URLs as well as bodies: the driver
// discovers it from GET /organizations and builds the members URL from it, so
// a cassette sanitized only in its bodies would stop matching its own
// requests.
const (
	elasticCloudFixtureOrgID   = "00000000000000000000000000000001"
	elasticCloudFixtureOrgName = "Example Organization"

	// GET /organizations returns a cursor even with a single organization.
	// It is opaque and may encode what it points at, so the sanitizer writes
	// this over it in the response and in the request that follows it.
	elasticCloudFixtureCursor = "example-organizations-cursor"
)

var elasticCloudOrgIDInPath = regexp.MustCompile(`/organizations/[^/]+`)

// The fields the sanitizer knows how to rewrite or keep. Anything else in a
// recording is refused: a field the provider adds later may carry identity.
var (
	elasticCloudOrganizationFields = []string{
		"id",
		"name",
		"force_mfa",
		"custom_bundles_allowed",
		"custom_plugins_allowed",
		"enforce_authentication_method",
		"default_disk_usage_alerts_enabled",
		"notifications_allowed_email_domains",
		"billing_contacts",
		"operational_contacts",
		"sso_login_identifier",
	}
	elasticCloudMemberFields = []string{
		"organization_id",
		"user_id",
		"name",
		"email",
		"member_since",
		"role_assignments",
	}
	elasticCloudAPIKeyFields = []string{
		"id",
		"user_id",
		"organization_id",
		"description",
		"creation_date",
		"expiration_date",
		"role_assignments",
	}
	elasticCloudAssignmentFields = []string{
		"role_id",
		"organization_id",
		"all",
		"deployment_ids",
		"project_ids",
		"application_roles",
	}
)

// elasticCloudAuth is the header value the connection transport sends.
func elasticCloudAuth(key string) string {
	if key == "" {
		return ""
	}

	return "ApiKey " + key
}

// TestElasticCloudDriver replays a recording of a live organization, so the
// field names and envelopes are Elastic's own rather than this package's idea
// of them. The organization holds its owner and two API keys: the Organization
// owner key the recording authenticated with, and a Billing admin key that
// also views every deployment. The mappings the recording does not exercise
// are covered by the table tests, which need no fixture.
func TestElasticCloudDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/elastic_cloud", "ELASTIC_CLOUD_API_KEY", sanitizeElasticCloud, dropResponseHeaders("Set-Cookie"))
	client := newVCRClient(rec, elasticCloudAuth(os.Getenv("ELASTIC_CLOUD_API_KEY")))
	driver := NewElasticCloudDriver(client, elasticCloudTestBaseURL)

	// The recorded keys expire a month after the recording; the clock stays
	// on the recording day so the replay does not age.
	driver.now = func() time.Time { return time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC) }

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)

	// While recording, the records hold the live identities the sanitizer
	// only rewrites on save; a failed assertion would print them.
	if os.Getenv("ELASTIC_CLOUD_API_KEY") != "" {
		return
	}

	require.Len(t, records, 3)

	owner := records[0]
	assert.Equal(t, "member1@example.com", owner.Email)
	assert.Equal(t, "ec-user-1", owner.ExternalID)
	assert.Equal(t, []string{"Organization owner"}, owner.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, owner.AccountType)
	require.NotNil(t, owner.IsAdmin)
	assert.True(t, *owner.IsAdmin)
	require.NotNil(t, owner.CreatedAt)

	// No status, last login or MFA anywhere on the members API. Asserting it
	// keeps a future mapping change from quietly inventing one.
	assert.Nil(t, owner.Active)
	assert.Nil(t, owner.LastLogin)
	assert.Equal(t, coredata.MFAStatusUnknown, owner.MFAStatus)

	// The key the recording authenticated with is itself a principal holding
	// Organization owner; listing only the members would hide it.
	key := records[1]
	assert.Equal(t, "ec-key-1", key.ExternalID)
	assert.Equal(t, "Example key 1", key.FullName)
	assert.Empty(t, key.Email)
	assert.Equal(t, []string{"Organization owner"}, key.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, key.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodAPIKey, key.AuthMethod)
	require.NotNil(t, key.IsAdmin)
	assert.True(t, *key.IsAdmin)
	require.NotNil(t, key.Active)
	assert.True(t, *key.Active)
	require.NotNil(t, key.CreatedAt)

	// Billing admin and Viewer on every deployment are both real reach, and
	// neither makes the key an administrator of the organization.
	narrow := records[2]
	assert.Equal(t, "ec-key-2", narrow.ExternalID)
	assert.Equal(t, []string{"Billing admin", "Deployment viewer (all deployments)"}, narrow.Roles)
	require.NotNil(t, narrow.IsAdmin)
	assert.False(t, *narrow.IsAdmin)
}

func TestElasticCloudNameResolverReplay(t *testing.T) {
	t.Parallel()

	// A cassette of its own, not the driver's: two recorders writing one path
	// race and silently drop interactions when the package is re-recorded.
	rec := newRecorder(t, "testdata/elastic_cloud_name", "ELASTIC_CLOUD_API_KEY", sanitizeElasticCloud, dropResponseHeaders("Set-Cookie"))
	client := newVCRClient(rec, elasticCloudAuth(os.Getenv("ELASTIC_CLOUD_API_KEY")))

	name, err := NewElasticCloudNameResolver(client, elasticCloudTestBaseURL).ResolveInstanceName(context.Background())
	require.NoError(t, err)

	if os.Getenv("ELASTIC_CLOUD_API_KEY") != "" {
		return
	}

	assert.Equal(t, elasticCloudFixtureOrgName, name)
}

func TestSanitizeElasticCloudFailsClosed(t *testing.T) {
	t.Parallel()

	members := elasticCloudTestBaseURL + "/organizations/" + elasticCloudFixtureOrgID + "/members"

	for _, tc := range []struct {
		name   string
		url    string
		status int
		body   string
	}{
		{
			name:   "error response",
			url:    members,
			status: http.StatusForbidden,
			body:   `{"errors":[]}`,
		},
		{
			name:   "unknown field on a member",
			url:    members,
			status: http.StatusOK,
			body:   `{"members":[{"user_id":"a","email":"real@example.org","surprise":"real@example.org"}]}`,
		},
		{
			name:   "unknown field on a role assignment",
			url:    members,
			status: http.StatusOK,
			body:   `{"members":[{"user_id":"a","role_assignments":{"organization":[{"role_id":"billing-admin","surprise":"x"}]}}]}`,
		},
		{
			name:   "unknown field on the envelope",
			url:    elasticCloudTestBaseURL + "/organizations",
			status: http.StatusOK,
			body:   `{"organizations":[],"surprise":"x"}`,
		},
		{
			name:   "an api key secret",
			url:    elasticCloudTestBaseURL + "/users/auth/keys",
			status: http.StatusOK,
			body:   `{"keys":[{"id":"a","description":"b","creation_date":"c","key":"secret"}]}`,
		},
		{
			// A cursor is opaque and may encode what it points at.
			name:   "a paginated request",
			url:    elasticCloudTestBaseURL + "/users/auth/keys?next_page=abc",
			status: http.StatusOK,
			body:   `{"keys":[]}`,
		},
		{
			name:   "a page that is not the last",
			url:    elasticCloudTestBaseURL + "/users/auth/keys",
			status: http.StatusOK,
			body:   `{"keys":[],"next_page":"abc"}`,
		},
		{
			name:   "unrecognised endpoint",
			url:    elasticCloudTestBaseURL + "/organizations/" + elasticCloudFixtureOrgID + "/invitations",
			status: http.StatusOK,
			body:   `{"invitations":[]}`,
		},
		{
			name:   "known path on another host",
			url:    "https://evil.example.com/api/v1/organizations",
			status: http.StatusOK,
			body:   `{"organizations":[]}`,
		},
	} {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				i := &cassette.Interaction{}
				i.Request.URL = tc.url
				i.Response.Code = tc.status
				i.Response.Body = tc.body

				assert.Error(t, sanitizeElasticCloud(i))
			},
		)
	}
}

// TestSanitizeElasticCloudRewritesIdentity checks the happy path removes what
// it rewrote, rather than only not erroring, and keeps what the driver reads.
func TestSanitizeElasticCloudRewritesIdentity(t *testing.T) {
	t.Parallel()

	i := &cassette.Interaction{}
	i.Request.URL = elasticCloudTestBaseURL + "/organizations/real-org-4242/members"
	i.Response.Code = http.StatusOK
	i.Response.Body = `{"members":[` +
		`{"organization_id":"real-org-4242","user_id":"real-user-1","name":"Real Person","email":"real.person@corp.example.org","member_since":"2026-01-02T03:04:05Z",` +
		`"role_assignments":{"organization":[{"organization_id":"real-org-4242","role_id":"organization-admin"}],` +
		`"deployment":[{"organization_id":"real-org-4242","role_id":"deployment-viewer","deployment_ids":["real-dep-a","real-dep-b"],"application_roles":["acme-payments-readers"]},` +
		`{"organization_id":"real-org-4242","role_id":"deployment-viewer","deployment_ids":["real-dep-b"],"application_roles":["acme-payments-readers"]}]}}]}`

	require.NoError(t, sanitizeElasticCloud(i))

	assert.Equal(t, elasticCloudTestBaseURL+"/organizations/"+elasticCloudFixtureOrgID+"/members", i.Request.URL)

	for _, recorded := range []string{"real-org-4242", "real-user-1", "Real Person", "real.person", "real-dep-a", "real-dep-b", "acme-payments-readers"} {
		assert.NotContains(t, i.Response.Body, recorded)
	}

	var body elasticCloudMembersResponse
	require.NoError(t, json.Unmarshal([]byte(i.Response.Body), &body))
	require.Len(t, body.Members, 1)
	assert.Equal(t, "ec-user-1", body.Members[0].UserID)
	assert.Equal(t, "member1@example.com", body.Members[0].Email)
	assert.Equal(t, "2026-01-02T03:04:05Z", body.Members[0].MemberSince)

	// The same deployment and the same custom role named twice must still
	// read as one each, or the sanitized labels would disagree with the live
	// ones.
	grants, err := elasticCloudGrants(body.Members[0].RoleAssignments)
	require.NoError(t, err)
	assert.Equal(
		t,
		[]string{"Organization owner", "Deployment viewer (2 deployments; application roles: application-role-1)"},
		elasticCloudGrantLabels(grants),
	)
}

// TestSanitizeElasticCloudKeepsBuiltInRoleIDs replays Elastic's documented
// key example: application role "admin" beside role id "elasticsearch-admin".
// Rewriting the application role must not make the role id read as a leak.
func TestSanitizeElasticCloudKeepsBuiltInRoleIDs(t *testing.T) {
	t.Parallel()

	i := &cassette.Interaction{}
	i.Request.URL = elasticCloudTestBaseURL + "/users/auth/keys"
	i.Response.Code = http.StatusOK
	i.Response.Body = `{"keys":[{"id":"real-key-1","description":"ci","creation_date":"2026-01-01T00:00:00Z",` +
		`"role_assignments":{"project":{"elasticsearch":[{"role_id":"elasticsearch-admin","organization_id":"real-org-4242","all":true,"application_roles":["admin"]}]}}}]}`

	require.NoError(t, sanitizeElasticCloud(i))
	assert.Contains(t, i.Response.Body, `"role_id":"elasticsearch-admin"`)
	assert.Contains(t, i.Response.Body, `"application_roles":["application-role-1"]`)
	assert.NotContains(t, i.Response.Body, "real-org-4242")
}

// TestSanitizeElasticCloudRewritesTheOrganizationsCursor checks the cursor is
// rewritten to the same value in the page that returns it and in the request
// that follows it, which is what lets the driver replay its own recording.
func TestSanitizeElasticCloudRewritesTheOrganizationsCursor(t *testing.T) {
	t.Parallel()

	const recorded = "cmVhbC1vcmdhbml6YXRpb24tY3Vyc29yLTQyNDI="

	first := &cassette.Interaction{}
	first.Request.URL = elasticCloudTestBaseURL + "/organizations"
	first.Response.Code = http.StatusOK
	first.Response.Body = `{"organizations":[{"id":"real-org-4242","name":"Real Org","force_mfa":false,"billing_contacts":["billing@corp.example.org"],"operational_contacts":[]}],"next_page":"` + recorded + `"}`

	require.NoError(t, sanitizeElasticCloud(first))
	assert.NotContains(t, first.Response.Body, recorded)
	assert.NotContains(t, first.Response.Body, "billing@corp.example.org")
	assert.Contains(t, first.Response.Body, elasticCloudFixtureCursor)

	second := &cassette.Interaction{}
	second.Request.URL = elasticCloudTestBaseURL + "/organizations?next_page=" + url.QueryEscape(recorded)
	second.Request.Form = url.Values{"next_page": {recorded}}
	second.Response.Code = http.StatusOK
	second.Response.Body = `{"organizations":[],"next_page":""}`

	require.NoError(t, sanitizeElasticCloud(second))
	assert.Equal(t, elasticCloudTestBaseURL+"/organizations?next_page="+elasticCloudFixtureCursor, second.Request.URL)
	assert.Equal(t, url.Values{"next_page": {elasticCloudFixtureCursor}}, second.Request.Form)

	extra := &cassette.Interaction{}
	extra.Request.URL = elasticCloudTestBaseURL + "/organizations?next_page=x&size=10"
	extra.Response.Code = http.StatusOK
	extra.Response.Body = `{"organizations":[]}`

	require.Error(t, sanitizeElasticCloud(extra))
}

func sanitizeElasticCloud(i *cassette.Interaction) error {
	if i.Response.Code != http.StatusOK {
		return fmt.Errorf("refusing to sanitize elastic cloud response with status %d", i.Response.Code)
	}

	parsed, err := url.Parse(i.Request.URL)
	if err != nil {
		return fmt.Errorf("cannot parse recorded elastic cloud URL: %w", err)
	}

	parsed.Path = elasticCloudOrgIDInPath.ReplaceAllString(parsed.Path, "/organizations/"+elasticCloudFixtureOrgID)

	// The organizations cursor is the one query a recording may carry. Any
	// other means more API keys than fit on a page, and a cursor this
	// sanitizer does not rewrite consistently would break replay.
	if query := parsed.Query(); len(query) > 0 {
		if len(query) != 1 || !query.Has("next_page") || parsed.Path != elasticCloudTestBasePath()+"/organizations" {
			return fmt.Errorf("refusing to sanitize a paginated elastic cloud request: record against fewer api keys")
		}

		parsed.RawQuery = url.Values{"next_page": {elasticCloudFixtureCursor}}.Encode()

		// The recorder keeps the parsed query beside the URL, and the matcher
		// compares it: rewriting only the URL would commit the real cursor and
		// fail the replay.
		i.Request.Form = parsed.Query()
	}

	i.Request.URL = parsed.String()

	ids := &elasticCloudSyntheticIDs{seen: map[string]string{}, next: map[string]int{}}

	// Dispatch on the whole URL, so a known path behind another host or
	// prefix is refused rather than taken for the roster.
	switch i.Request.URL {
	case elasticCloudTestBaseURL + "/organizations",
		elasticCloudTestBaseURL + "/organizations?next_page=" + elasticCloudFixtureCursor:
		return sanitizeElasticCloudBody(i, ids, sanitizeElasticCloudOrganizations)
	case elasticCloudTestBaseURL + "/organizations/" + elasticCloudFixtureOrgID + "/members":
		return sanitizeElasticCloudBody(i, ids, sanitizeElasticCloudMembers)
	case elasticCloudTestBaseURL + "/users/auth/keys":
		return sanitizeElasticCloudBody(i, ids, sanitizeElasticCloudAPIKeys)
	default:
		return fmt.Errorf("refusing to sanitize an unrecognised elastic cloud interaction")
	}
}

// elasticCloudTestBasePath is the path the driver under test prefixes every
// request with.
func elasticCloudTestBasePath() string {
	base, err := url.Parse(elasticCloudTestBaseURL)
	if err != nil {
		return ""
	}

	return base.Path
}

// elasticCloudSyntheticIDs hands out one synthetic id per recorded id, so a
// deployment named in two assignments still reads as one after rewriting. It
// lives for one interaction and is never written anywhere.
type elasticCloudSyntheticIDs struct {
	seen map[string]string
	next map[string]int
}

func (s *elasticCloudSyntheticIDs) get(prefix, recorded string) string {
	if id, ok := s.seen[prefix+"\x00"+recorded]; ok {
		return id
	}

	s.next[prefix]++
	id := fmt.Sprintf("%s-%d", prefix, s.next[prefix])
	s.seen[prefix+"\x00"+recorded] = id

	return id
}

func sanitizeElasticCloudBody(
	i *cassette.Interaction,
	ids *elasticCloudSyntheticIDs,
	rewrite func(map[string]json.RawMessage, *elasticCloudSyntheticIDs) ([]string, error),
) error {
	var body map[string]json.RawMessage
	if err := json.Unmarshal([]byte(i.Response.Body), &body); err != nil {
		return fmt.Errorf("cannot decode recorded elastic cloud response: %w", err)
	}

	recorded, err := rewrite(body, ids)
	if err != nil {
		return err
	}

	sanitized, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("cannot re-encode elastic cloud response: %w", err)
	}

	// A second guard behind the declared-field check, for a value rewritten in
	// one place and repeated somewhere the rewriter does not treat as identity.
	for _, value := range recorded {
		if strings.Contains(string(sanitized), value) {
			return fmt.Errorf("sanitized elastic cloud response still contains a recorded identity value")
		}
	}

	replaceCassetteBody(i, string(sanitized))

	return nil
}

func sanitizeElasticCloudOrganizations(body map[string]json.RawMessage, _ *elasticCloudSyntheticIDs) ([]string, error) {
	if err := elasticCloudRejectUnknownFields(body, []string{"organizations", "next_page"}, "organizations envelope"); err != nil {
		return nil, err
	}

	orgs, err := elasticCloudList(body, "organizations")
	if err != nil {
		return nil, err
	}

	var recorded []string

	if cursor := elasticCloudString(body, "next_page"); cursor != "" {
		recorded = append(recorded, cursor)

		elasticCloudSetIfPresent(body, "next_page", elasticCloudFixtureCursor)
	}

	for _, org := range orgs {
		if err := elasticCloudRejectUnknownFields(org, elasticCloudOrganizationFields, "organization"); err != nil {
			return nil, err
		}

		recorded = append(recorded, elasticCloudRecorded(org, "id", "name", "sso_login_identifier")...)

		for _, field := range []string{"notifications_allowed_email_domains", "billing_contacts", "operational_contacts"} {
			values, err := elasticCloudRecordedList(org, field)
			if err != nil {
				return nil, err
			}

			recorded = append(recorded, values...)

			elasticCloudSetIfPresent(org, field, []string{})
		}

		elasticCloudSetIfPresent(org, "id", elasticCloudFixtureOrgID)
		elasticCloudSetIfPresent(org, "name", elasticCloudFixtureOrgName)
		elasticCloudSetIfPresent(org, "sso_login_identifier", "example-organization")
	}

	return recorded, elasticCloudSetList(body, "organizations", orgs)
}

func sanitizeElasticCloudMembers(body map[string]json.RawMessage, ids *elasticCloudSyntheticIDs) ([]string, error) {
	if err := elasticCloudRejectUnknownFields(body, []string{"members"}, "members envelope"); err != nil {
		return nil, err
	}

	members, err := elasticCloudList(body, "members")
	if err != nil {
		return nil, err
	}

	var recorded []string

	for n, member := range members {
		if err := elasticCloudRejectUnknownFields(member, elasticCloudMemberFields, "member"); err != nil {
			return nil, err
		}

		recorded = append(recorded, elasticCloudRecorded(member, "organization_id", "user_id", "name", "email")...)

		elasticCloudSetIfPresent(member, "organization_id", elasticCloudFixtureOrgID)
		elasticCloudSetIfPresent(member, "user_id", fmt.Sprintf("ec-user-%d", n+1))
		elasticCloudSetIfPresent(member, "name", fmt.Sprintf("Member %d", n+1))
		elasticCloudSetIfPresent(member, "email", fmt.Sprintf("member%d@example.com", n+1))

		values, err := sanitizeElasticCloudRoleAssignments(member, ids)
		if err != nil {
			return nil, err
		}

		recorded = append(recorded, values...)
	}

	return recorded, elasticCloudSetList(body, "members", members)
}

func sanitizeElasticCloudAPIKeys(body map[string]json.RawMessage, ids *elasticCloudSyntheticIDs) ([]string, error) {
	keys, err := elasticCloudLastPage(body, "keys")
	if err != nil {
		return nil, err
	}

	var recorded []string

	for n, key := range keys {
		if _, ok := key["key"]; ok {
			return nil, fmt.Errorf("refusing to record an elastic cloud api key secret")
		}

		if err := elasticCloudRejectUnknownFields(key, elasticCloudAPIKeyFields, "api key"); err != nil {
			return nil, err
		}

		recorded = append(recorded, elasticCloudRecorded(key, "id", "user_id", "organization_id", "description")...)

		elasticCloudSetIfPresent(key, "id", fmt.Sprintf("ec-key-%d", n+1))
		elasticCloudSetIfPresent(key, "user_id", ids.get("ec-key-creator", elasticCloudString(key, "user_id")))
		elasticCloudSetIfPresent(key, "organization_id", elasticCloudFixtureOrgID)
		elasticCloudSetIfPresent(key, "description", fmt.Sprintf("Example key %d", n+1))

		values, err := sanitizeElasticCloudRoleAssignments(key, ids)
		if err != nil {
			return nil, err
		}

		recorded = append(recorded, values...)
	}

	return recorded, elasticCloudSetList(body, "keys", keys)
}

// sanitizeElasticCloudRoleAssignments walks every scope the way the driver
// does: a list of assignments, or one list per project type.
func sanitizeElasticCloudRoleAssignments(holder map[string]json.RawMessage, ids *elasticCloudSyntheticIDs) ([]string, error) {
	raw, ok := holder["role_assignments"]
	if !ok {
		return nil, nil
	}

	var scopes map[string]json.RawMessage
	if err := json.Unmarshal(raw, &scopes); err != nil {
		return nil, fmt.Errorf("cannot decode recorded elastic cloud role assignments: %w", err)
	}

	var recorded []string

	for scope, value := range scopes {
		var assignments []map[string]json.RawMessage
		if err := json.Unmarshal(value, &assignments); err == nil {
			values, err := sanitizeElasticCloudAssignments(assignments, ids)
			if err != nil {
				return nil, err
			}

			recorded = append(recorded, values...)

			if scopes[scope], err = json.Marshal(assignments); err != nil {
				return nil, fmt.Errorf("cannot re-encode elastic cloud role assignments: %w", err)
			}

			continue
		}

		var nested map[string][]map[string]json.RawMessage
		if err := json.Unmarshal(value, &nested); err != nil {
			return nil, fmt.Errorf("cannot decode recorded elastic cloud %s role assignments: %w", scope, err)
		}

		for _, assignments := range nested {
			values, err := sanitizeElasticCloudAssignments(assignments, ids)
			if err != nil {
				return nil, err
			}

			recorded = append(recorded, values...)
		}

		var err error
		if scopes[scope], err = json.Marshal(nested); err != nil {
			return nil, fmt.Errorf("cannot re-encode elastic cloud role assignments: %w", err)
		}
	}

	encoded, err := json.Marshal(scopes)
	if err != nil {
		return nil, fmt.Errorf("cannot re-encode elastic cloud role assignments: %w", err)
	}

	holder["role_assignments"] = encoded

	return recorded, nil
}

func sanitizeElasticCloudAssignments(assignments []map[string]json.RawMessage, ids *elasticCloudSyntheticIDs) ([]string, error) {
	var recorded []string

	for _, assignment := range assignments {
		if err := elasticCloudRejectUnknownFields(assignment, elasticCloudAssignmentFields, "role assignment"); err != nil {
			return nil, err
		}

		recorded = append(recorded, elasticCloudRecorded(assignment, "organization_id")...)
		elasticCloudSetIfPresent(assignment, "organization_id", elasticCloudFixtureOrgID)

		// Application roles include custom ones, whose names are the
		// customer's to choose.
		for field, prefix := range map[string]string{
			"deployment_ids":    "deployment",
			"project_ids":       "project",
			"application_roles": "application-role",
		} {
			values, err := elasticCloudRecordedList(assignment, field)
			if err != nil {
				return nil, err
			}

			synthetic := make([]string, 0, len(values))
			for _, v := range values {
				synthetic = append(synthetic, ids.get(prefix, v))
			}

			// Built-in application roles such as "admin" are substrings of
			// role ids the recording keeps, so the residual check would refuse
			// every recording that holds one. The field is rewritten whole.
			if field != "application_roles" {
				recorded = append(recorded, values...)
			}

			elasticCloudSetIfPresent(assignment, field, synthetic)
		}
	}

	return recorded, nil
}

// elasticCloudLastPage reads a paged envelope and refuses one that is not the
// last page: a cursor is opaque, and a recording that needs one cannot be
// sanitized with confidence.
func elasticCloudLastPage(body map[string]json.RawMessage, field string) ([]map[string]json.RawMessage, error) {
	if err := elasticCloudRejectUnknownFields(body, []string{field, "next_page"}, field+" envelope"); err != nil {
		return nil, err
	}

	if elasticCloudString(body, "next_page") != "" {
		return nil, fmt.Errorf("refusing to sanitize a paginated elastic cloud response: record against fewer %s", field)
	}

	return elasticCloudList(body, field)
}

func elasticCloudList(body map[string]json.RawMessage, field string) ([]map[string]json.RawMessage, error) {
	var list []map[string]json.RawMessage
	if raw, ok := body[field]; ok {
		if err := json.Unmarshal(raw, &list); err != nil {
			return nil, fmt.Errorf("cannot decode recorded elastic cloud %s: %w", field, err)
		}
	}

	return list, nil
}

func elasticCloudSetList(body map[string]json.RawMessage, field string, list []map[string]json.RawMessage) error {
	if _, ok := body[field]; !ok {
		return nil
	}

	encoded, err := json.Marshal(list)
	if err != nil {
		return fmt.Errorf("cannot re-encode elastic cloud %s: %w", field, err)
	}

	body[field] = encoded

	return nil
}

func elasticCloudRejectUnknownFields(object map[string]json.RawMessage, known []string, what string) error {
	for field := range object {
		if !slices.Contains(known, field) {
			return fmt.Errorf("recorded elastic cloud %s carries an unknown field %q", what, field)
		}
	}

	return nil
}

func elasticCloudString(object map[string]json.RawMessage, field string) string {
	var value string
	if raw, ok := object[field]; ok {
		_ = json.Unmarshal(raw, &value)
	}

	return value
}

// elasticCloudRecorded collects the string values about to be overwritten, for
// the post-rewrite check that none of them survived elsewhere in the body.
func elasticCloudRecorded(object map[string]json.RawMessage, fields ...string) []string {
	var values []string

	for _, field := range fields {
		if value := elasticCloudString(object, field); value != "" {
			values = append(values, value)
		}
	}

	return values
}

func elasticCloudRecordedList(object map[string]json.RawMessage, field string) ([]string, error) {
	raw, ok := object[field]
	if !ok {
		return nil, nil
	}

	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, fmt.Errorf("cannot decode recorded elastic cloud %s: %w", field, err)
	}

	return slices.DeleteFunc(values, func(v string) bool { return v == "" }), nil
}

func elasticCloudSetIfPresent(object map[string]json.RawMessage, field string, value any) {
	if _, ok := object[field]; !ok {
		return
	}

	if encoded, err := json.Marshal(value); err == nil {
		object[field] = encoded
	}
}

// elasticCloudStub answers each request path with a fixed status and body, and
// fails the test on any path it was not given.
func elasticCloudStub(t *testing.T, responses map[string]func(*http.Request) (int, string)) *http.Client {
	t.Helper()

	return &http.Client{
		Transport: roundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				respond, ok := responses[req.URL.Path]
				if !ok {
					t.Errorf("unexpected elastic cloud request to %s", req.URL.Path)

					return nil, errors.New("unexpected request")
				}

				assert.Equal(t, "application/json", req.Header.Get("Accept"))

				status, body := respond(req)

				return &http.Response{
					StatusCode: status,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(body)),
				}, nil
			},
		),
	}
}

func elasticCloudFixed(status int, body string) func(*http.Request) (int, string) {
	return func(*http.Request) (int, string) {
		return status, body
	}
}

func elasticCloudRoleAssignments(t *testing.T, raw string) map[string]json.RawMessage {
	t.Helper()

	var assignments map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(raw), &assignments))

	return assignments
}

// elasticCloudOwnerMembers is a roster as an Organization owner key sees it:
// the owner's role assignments are visible.
const elasticCloudOwnerMembers = `{"members":[{"organization_id":"org-1","user_id":"u-1","name":"Alice","email":"alice@example.com","member_since":"2026-01-02T03:04:05Z","role_assignments":{"organization":[{"organization_id":"org-1","role_id":"organization-admin"}]}}]}`

// elasticCloudOneOrganization answers GET /organizations the way Elastic does:
// the only organization on the first page, with a cursor that leads to an
// empty one.
func elasticCloudOneOrganization(req *http.Request) (int, string) {
	if req.URL.Query().Get("next_page") == "" {
		return http.StatusOK, `{"organizations":[{"id":"org-1","name":" Acme "}],"next_page":"org-cursor"}`
	}

	return http.StatusOK, `{"organizations":[],"next_page":""}`
}

func TestElasticCloudDriverListsMembersThenAPIKeys(t *testing.T) {
	t.Parallel()

	var cursors []string

	client := elasticCloudStub(
		t,
		map[string]func(*http.Request) (int, string){
			"/api/v1/organizations":               elasticCloudOneOrganization,
			"/api/v1/organizations/org-1/members": elasticCloudFixed(http.StatusOK, elasticCloudOwnerMembers),
			"/api/v1/users/auth/keys": func(req *http.Request) (int, string) {
				cursor := req.URL.Query().Get("next_page")
				cursors = append(cursors, cursor)

				if cursor == "" {
					return http.StatusOK, `{"keys":[{"id":"k-1","description":"terraform","creation_date":"2026-02-01T00:00:00Z","role_assignments":{"organization":[{"organization_id":"org-1","role_id":"billing-admin"}]}}],"next_page":"c-1"}`
				}

				return http.StatusOK, `{"keys":[{"id":"k-2","description":"probo","creation_date":"2026-03-01T00:00:00Z","expiration_date":"2027-03-01T00:00:00Z","role_assignments":{"deployment":[{"organization_id":"org-1","role_id":"deployment-viewer","all":true}]}}]}`
			},
		},
	)

	driver := NewElasticCloudDriver(client, elasticCloudTestBaseURL)

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)
	assert.Equal(t, []string{"", "c-1"}, cursors)

	member := records[0]
	assert.Equal(t, "alice@example.com", member.Email)
	assert.Equal(t, "Alice", member.FullName)
	assert.Equal(t, "u-1", member.ExternalID)
	assert.Equal(t, []string{"Organization owner"}, member.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, member.AccountType)
	require.NotNil(t, member.IsAdmin)
	assert.True(t, *member.IsAdmin)
	assert.Nil(t, member.Active)
	require.NotNil(t, member.CreatedAt)
	assert.Equal(t, "2026-01-02T03:04:05Z", member.CreatedAt.UTC().Format(time.RFC3339))

	terraformKey := records[1]
	assert.Equal(t, "k-1", terraformKey.ExternalID)
	assert.Equal(t, "terraform", terraformKey.FullName)
	assert.Empty(t, terraformKey.Email)
	assert.Equal(t, []string{"Billing admin"}, terraformKey.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, terraformKey.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodAPIKey, terraformKey.AuthMethod)
	require.NotNil(t, terraformKey.IsAdmin)
	assert.False(t, *terraformKey.IsAdmin)

	assert.Equal(t, "k-2", records[2].ExternalID)
	assert.Equal(t, []string{"Deployment viewer (all deployments)"}, records[2].Roles)
}

func TestElasticCloudDriverRequiresExactlyOneOrganization(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		pages map[string]string
	}{
		{
			name:  "no organization",
			pages: map[string]string{"": `{"organizations":[]}`},
		},
		{
			name:  "two organizations on one page",
			pages: map[string]string{"": `{"organizations":[{"id":"a","name":"A"},{"id":"b","name":"B"}]}`},
		},
		{
			name: "one organization on each of two pages",
			pages: map[string]string{
				"":  `{"organizations":[{"id":"a","name":"A"}],"next_page":"c"}`,
				"c": `{"organizations":[{"id":"b","name":"B"}],"next_page":""}`,
			},
		},
		{
			name:  "organization without an id",
			pages: map[string]string{"": `{"organizations":[{"id":"","name":"A"}]}`},
		},
	} {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				client := elasticCloudStub(
					t,
					map[string]func(*http.Request) (int, string){
						"/api/v1/organizations": func(req *http.Request) (int, string) {
							return http.StatusOK, tc.pages[req.URL.Query().Get("next_page")]
						},
					},
				)

				_, err := NewElasticCloudDriver(client, elasticCloudTestBaseURL).ListAccounts(context.Background())
				require.ErrorIs(t, err, errElasticCloudNoSoleOrg)
			},
		)
	}
}

// TestElasticCloudDriverRefusesARestrictedKey replays what Elastic answers a
// Billing admin key, observed live: the roster with no role_assignments at
// all. Syncing it would show the owner as a member without a role.
func TestElasticCloudDriverRefusesARestrictedKey(t *testing.T) {
	t.Parallel()

	client := elasticCloudStub(
		t,
		map[string]func(*http.Request) (int, string){
			"/api/v1/organizations": elasticCloudOneOrganization,
			"/api/v1/organizations/org-1/members": elasticCloudFixed(
				http.StatusOK,
				`{"members":[{"organization_id":"org-1","user_id":"u-1","name":"Alice","email":"alice@example.com","member_since":"2026-01-02T03:04:05Z"}]}`,
			),
		},
	)

	records, err := NewElasticCloudDriver(client, elasticCloudTestBaseURL).ListAccounts(context.Background())
	assert.Nil(t, records)

	// A SettingRejectedError, so the fetch failure shows the customer what to
	// fix rather than the generic message.
	rejected, ok := errors.AsType[*SettingRejectedError](err)
	require.True(t, ok)
	assert.Equal(t, ElasticCloudKeyNotOwner, rejected.Code)
	assert.Empty(t, rejected.Setting)
	assert.Contains(t, rejected.Message, "Organization owner")
}

func TestElasticCloudDriverFailsWhenAPIKeysAreRefused(t *testing.T) {
	t.Parallel()

	client := elasticCloudStub(
		t,
		map[string]func(*http.Request) (int, string){
			"/api/v1/organizations":               elasticCloudOneOrganization,
			"/api/v1/organizations/org-1/members": elasticCloudFixed(http.StatusOK, elasticCloudOwnerMembers),
			"/api/v1/users/auth/keys":             elasticCloudFixed(http.StatusForbidden, `{"errors":[]}`),
		},
	)

	records, err := NewElasticCloudDriver(client, elasticCloudTestBaseURL).ListAccounts(context.Background())
	require.Error(t, err)
	assert.Nil(t, records)
	assert.Contains(t, err.Error(), "unexpected status 403")
}

func TestElasticCloudDriverRefusesAStuckCursor(t *testing.T) {
	t.Parallel()

	client := elasticCloudStub(
		t,
		map[string]func(*http.Request) (int, string){
			"/api/v1/organizations":               elasticCloudOneOrganization,
			"/api/v1/organizations/org-1/members": elasticCloudFixed(http.StatusOK, elasticCloudOwnerMembers),
			"/api/v1/users/auth/keys":             elasticCloudFixed(http.StatusOK, `{"keys":[],"next_page":"same"}`),
		},
	)

	_, err := NewElasticCloudDriver(client, elasticCloudTestBaseURL).ListAccounts(context.Background())
	require.ErrorContains(t, err, "next_page cursor did not advance")
}

func TestElasticCloudGrantLabels(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		raw  string
		want []string
	}{
		{
			name: "no role assignments",
			raw:  `{}`,
			want: []string{},
		},
		{
			name: "organization roles keep the console names",
			raw:  `{"organization":[{"role_id":"organization-admin"},{"role_id":"billing-admin"}]}`,
			want: []string{"Organization owner", "Billing admin"},
		},
		{
			name: "one role on two sets of deployments is listed once, with the union",
			raw:  `{"deployment":[{"role_id":"deployment-editor","deployment_ids":["d-1","d-2"]},{"role_id":"deployment-editor","deployment_ids":["d-2","d-3"]}]}`,
			want: []string{"Deployment editor (3 deployments)"},
		},
		{
			name: "a single deployment reads in the singular",
			raw:  `{"deployment":[{"role_id":"deployment-viewer","deployment_ids":["d-1"]}]}`,
			want: []string{"Deployment viewer (1 deployment)"},
		},
		{
			name: "all wins over a list of ids",
			raw:  `{"deployment":[{"role_id":"deployment-admin","deployment_ids":["d-1"]},{"role_id":"deployment-admin","all":true}]}`,
			want: []string{"Deployment admin (all deployments)"},
		},
		{
			name: "project roles name their type whether or not the id carries it",
			raw:  `{"project":{"security":[{"role_id":"admin","project_ids":["p-1"]}],"elasticsearch":[{"role_id":"elasticsearch-viewer","all":true}]}}`,
			want: []string{"Elasticsearch viewer (all projects)", "Security admin (1 project)"},
		},
		{
			name: "application roles are part of what a role grants",
			raw:  `{"deployment":[{"role_id":"deployment-viewer","all":true,"application_roles":["reporting","ingest"]}]}`,
			want: []string{"Deployment viewer (all deployments; application roles: ingest, reporting)"},
		},
		{
			// Application roles are bound to the assignment's own resources,
			// so merging would claim admin on every project.
			name: "one role with different application roles is listed per set",
			raw:  `{"project":{"elasticsearch":[{"role_id":"elasticsearch-viewer","all":true},{"role_id":"elasticsearch-viewer","project_ids":["p-1"],"application_roles":["admin"]}]}}`,
			want: []string{"Elasticsearch viewer (all projects)", "Elasticsearch viewer (1 project; application roles: admin)"},
		},
		{
			name: "an empty application role list merges with an absent one",
			raw:  `{"deployment":[{"role_id":"deployment-viewer","deployment_ids":["d-1"],"application_roles":[]},{"role_id":"deployment-viewer","deployment_ids":["d-2"]}]}`,
			want: []string{"Deployment viewer (2 deployments)"},
		},
		{
			name: "an unknown scope keeps how far it reaches",
			raw:  `{"connected_cluster":[{"role_id":"viewer","all":true}]}`,
			want: []string{"Connected cluster viewer (all resources)"},
		},
		{
			name: "scopes sort from widest to narrowest, then unknown ones by name",
			raw:  `{"platform":[{"role_id":"platform-viewer"}],"project":{"observability":[{"role_id":"observability-editor","all":true}]},"connected_cluster":[{"role_id":"viewer"}],"organization":[{"role_id":"billing-admin"}],"deployment":[{"role_id":"deployment-viewer","all":true}]}`,
			want: []string{
				"Billing admin",
				"Deployment viewer (all deployments)",
				"Observability editor (all projects)",
				"Platform viewer",
				"Connected cluster viewer",
			},
		},
		{
			name: "an assignment without a role id is skipped",
			raw:  `{"organization":[{"role_id":" "}]}`,
			want: []string{},
		},
	} {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				grants, err := elasticCloudGrants(elasticCloudRoleAssignments(t, tc.raw))
				require.NoError(t, err)
				assert.Equal(t, tc.want, elasticCloudGrantLabels(grants))
			},
		)
	}
}

func TestElasticCloudGrantsRejectAnUndecodableScope(t *testing.T) {
	t.Parallel()

	_, err := elasticCloudGrants(elasticCloudRoleAssignments(t, `{"organization":"organization-admin"}`))
	require.ErrorContains(t, err, "cannot decode elastic cloud organization role assignments")
}

func TestElasticCloudOwnerDetection(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		raw  string
		want bool
	}{
		{name: "organization owner", raw: `{"organization":[{"role_id":"organization-admin"}]}`, want: true},
		{name: "billing admin is not an owner", raw: `{"organization":[{"role_id":"billing-admin"}]}`, want: false},
		{name: "deployment admin on every deployment is not an owner", raw: `{"deployment":[{"role_id":"deployment-admin","all":true}]}`, want: false},
		{name: "the owner role id outside the organization scope", raw: `{"deployment":[{"role_id":"organization-admin","all":true}]}`, want: false},
		{name: "no roles", raw: `{}`, want: false},
	} {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				grants, err := elasticCloudGrants(elasticCloudRoleAssignments(t, tc.raw))
				require.NoError(t, err)
				assert.Equal(t, tc.want, elasticCloudIsOwner(grants))
			},
		)
	}
}

// TestElasticCloudAPIKeyExpiry pins the clock: the expiry rule must not depend
// on when the suite runs.
func TestElasticCloudAPIKeyExpiry(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)

	for _, tc := range []struct {
		name       string
		expiration string
		want       bool
	}{
		{name: "never expires", expiration: "", want: true},
		{name: "expires later", expiration: "2026-10-28T12:00:01Z", want: true},
		{name: "expires now", expiration: "2026-10-28T12:00:00Z", want: false},
		{name: "expired", expiration: "2026-01-01T00:00:00Z", want: false},
	} {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				record, err := elasticCloudAPIKeyRecord(
					elasticCloudAPIKey{ID: "k-1", Description: "ci", ExpirationDate: tc.expiration},
					now,
				)
				require.NoError(t, err)
				require.NotNil(t, record.Active)
				assert.Equal(t, tc.want, *record.Active)
			},
		)
	}
}

func TestElasticCloudNameResolver(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name         string
		status       int
		body         string
		want         string
		wantErr      bool
		wantTerminal bool
	}{
		{name: "organization name", status: http.StatusOK, body: `{"organizations":[{"id":"org-1","name":" Acme "}]}`, want: "Acme"},
		{name: "no sole organization keeps the generic name", status: http.StatusOK, body: `{"organizations":[]}`, want: ""},
		{name: "rejected key", status: http.StatusUnauthorized, body: `{}`, wantErr: true, wantTerminal: true},
		{name: "refused key", status: http.StatusForbidden, body: `{}`, wantErr: true, wantTerminal: true},
		{name: "server error retries", status: http.StatusBadGateway, body: `{}`, wantErr: true},
	} {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				client := elasticCloudStub(
					t,
					map[string]func(*http.Request) (int, string){
						"/api/v1/organizations": elasticCloudFixed(tc.status, tc.body),
					},
				)

				name, err := NewElasticCloudNameResolver(client, elasticCloudTestBaseURL).ResolveInstanceName(context.Background())
				if !tc.wantErr {
					require.NoError(t, err)
					assert.Equal(t, tc.want, name)

					return
				}

				require.Error(t, err)
				assert.Equal(t, tc.wantTerminal, errors.Is(err, ErrTerminalNameResolution))
			},
		)
	}
}
