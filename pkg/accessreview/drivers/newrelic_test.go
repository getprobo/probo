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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
)

// newRelicSanitizer replaces the identity in recorded NerdGraph responses with
// synthetic values, keeping one mapping across every interaction in the
// cassette so a user id in a group's member list still resolves to the same
// user in the roster — which is exactly the join the driver performs.
//
// It walks the decoded JSON rather than re-marshalling the response types,
// which would rewrite the body as a serialization of the types under test and
// drop whatever the driver does not model. Decoding into map[string]any is
// what makes that walk safe: the nested maps are references, so a rewrite
// reaches the body without a repack chain that could silently miss a level.
// Numbers are decoded as json.Number so re-encoding cannot turn an integer id
// into 1e+06.
//
// Group display names are left alone: they are New Relic's own built-in group
// names, they carry no identity, and admin detection is asserted off them.
//
// The hook runs on save, so a recording run still sees live data and its
// assertions fail. Re-record, then run again without NEW_RELIC_API_KEY.
type newRelicSanitizer struct {
	userIDs   map[string]string
	domainIDs map[string]string
	groupIDs  map[string]string
}

func newNewRelicSanitizer() *newRelicSanitizer {
	return &newRelicSanitizer{
		userIDs:   map[string]string{},
		domainIDs: map[string]string{},
		groupIDs:  map[string]string{},
	}
}

// synthetic maps one real identifier to a stable stand-in, so the same domain
// or group keeps one id across every interaction in the cassette. Numbering by
// position instead would renumber it per page, and a targeted follow-up whose
// domain came back as domain-0001 on one page and domain-0002 on another can
// never be matched back to the page that asked for it.
func synthetic(seen map[string]string, prefix, real string) string {
	if id, ok := seen[real]; ok {
		return id
	}

	id := fmt.Sprintf("%s-%04d", prefix, len(seen)+1)
	seen[real] = id

	return id
}

func (s *newRelicSanitizer) userID(real string) string {
	if id, ok := s.userIDs[real]; ok {
		return id
	}

	id := fmt.Sprintf("100000%04d", len(s.userIDs)+1)
	s.userIDs[real] = id

	return id
}

// newRelicRewrite replaces a field that must be present and must be a string.
// Decoded rather than merely present: rewriting a value that came back as null
// or a number would paper over the struct mismatch this cassette exposes.
func newRelicRewrite(object map[string]any, field, replacement, what string) error {
	value, ok := object[field]
	if !ok {
		return fmt.Errorf("recorded new relic %s has no %s field", what, field)
	}

	if _, ok := value.(string); !ok {
		return fmt.Errorf("recorded new relic %s has a non-string %s", what, field)
	}

	object[field] = replacement

	return nil
}

// newRelicChild returns a nested object, erroring when the shape is not the
// one the driver decodes.
func newRelicChild(parent map[string]any, field, what string) (map[string]any, error) {
	raw, ok := parent[field]
	if !ok {
		return nil, fmt.Errorf("recorded new relic %s has no %s field", what, field)
	}

	child, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("recorded new relic %s has a non-object %s", what, field)
	}

	return child, nil
}

// newRelicList returns a nested array of objects.
func newRelicList(parent map[string]any, field, what string) ([]map[string]any, error) {
	raw, ok := parent[field]
	if !ok {
		return nil, fmt.Errorf("recorded new relic %s has no %s field", what, field)
	}

	items, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("recorded new relic %s has a non-array %s", what, field)
	}

	list := make([]map[string]any, 0, len(items))

	for _, item := range items {
		object, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("recorded new relic %s has a non-object entry in %s", what, field)
		}

		list = append(list, object)
	}

	return list, nil
}

func (s *newRelicSanitizer) sanitize(i *cassette.Interaction) error {
	if i.Response.Code != http.StatusOK {
		return fmt.Errorf("refusing to sanitize new relic response with status %d", i.Response.Code)
	}

	decoder := json.NewDecoder(strings.NewReader(i.Response.Body))
	decoder.UseNumber()

	var body map[string]any
	if err := decoder.Decode(&body); err != nil {
		return fmt.Errorf("cannot decode recorded new relic response: %w", err)
	}

	// NerdGraph reports a rejected query with 200 and an errors array, so a
	// recording that failed would otherwise be saved as a usable cassette.
	if _, ok := body["errors"]; ok {
		return fmt.Errorf("refusing to sanitize a new relic response carrying graphql errors")
	}

	data, err := newRelicChild(body, "data", "response")
	if err != nil {
		return err
	}

	actor, err := newRelicChild(data, "actor", "data")
	if err != nil {
		return err
	}

	organization, err := newRelicChild(actor, "organization", "actor")
	if err != nil {
		return err
	}

	// The two queries in this cassette ask for different halves of the
	// organization — the name resolver for its name, the driver for its
	// user management — so each field is rewritten only where it was asked
	// for. A response carrying neither is one neither query would produce.
	_, hasName := organization["name"]
	_, hasUserManagement := organization["userManagement"]

	if !hasName && !hasUserManagement {
		return fmt.Errorf("recorded new relic organization carries neither name nor userManagement")
	}

	if hasName {
		if err := newRelicRewrite(organization, "name", "Example Organization", "organization"); err != nil {
			return err
		}
	}

	if hasUserManagement {
		if err := s.sanitizeUserManagement(organization); err != nil {
			return err
		}
	}

	sanitized, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("cannot re-encode sanitized new relic response: %w", err)
	}

	replaceCassetteBody(i, string(sanitized))

	return nil
}

func (s *newRelicSanitizer) sanitizeUserManagement(organization map[string]any) error {
	userManagement, err := newRelicChild(organization, "userManagement", "organization")
	if err != nil {
		return err
	}

	collection, err := newRelicChild(userManagement, "authenticationDomains", "userManagement")
	if err != nil {
		return err
	}

	domains, err := newRelicList(collection, "authenticationDomains", "authentication domain collection")
	if err != nil {
		return err
	}

	for _, domain := range domains {
		realID, ok := domain["id"].(string)
		if !ok {
			return fmt.Errorf("recorded new relic domain has no string id")
		}

		if err := newRelicRewrite(domain, "id", synthetic(s.domainIDs, "domain", realID), "domain"); err != nil {
			return err
		}

		if _, ok := domain["name"]; ok {
			if err := newRelicRewrite(domain, "name", "Default", "domain"); err != nil {
				return err
			}
		}

		if _, ok := domain["users"]; ok {
			if err := s.sanitizeDomainUsers(domain); err != nil {
				return err
			}
		}

		if _, ok := domain["groups"]; ok {
			if err := s.sanitizeDomainGroups(domain); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *newRelicSanitizer) sanitizeDomainUsers(domain map[string]any) error {
	collection, err := newRelicChild(domain, "users", "domain")
	if err != nil {
		return err
	}

	users, err := newRelicList(collection, "users", "user collection")
	if err != nil {
		return err
	}

	for _, user := range users {
		realID, ok := user["id"].(string)
		if !ok {
			return fmt.Errorf("recorded new relic user has no string id")
		}

		synthetic := s.userID(realID)

		if err := newRelicRewrite(user, "id", synthetic, "user"); err != nil {
			return err
		}

		if err := newRelicRewrite(user, "name", "Member "+synthetic, "user"); err != nil {
			return err
		}

		if err := newRelicRewrite(user, "email", synthetic+"@example.com", "user"); err != nil {
			return err
		}
	}

	return nil
}

func (s *newRelicSanitizer) sanitizeDomainGroups(domain map[string]any) error {
	collection, err := newRelicChild(domain, "groups", "domain")
	if err != nil {
		return err
	}

	groups, err := newRelicList(collection, "groups", "group collection")
	if err != nil {
		return err
	}

	for _, group := range groups {
		realID, ok := group["id"].(string)
		if !ok {
			return fmt.Errorf("recorded new relic group has no string id")
		}

		if err := newRelicRewrite(group, "id", synthetic(s.groupIDs, "group", realID), "group"); err != nil {
			return err
		}

		members, err := newRelicChild(group, "users", "group")
		if err != nil {
			return err
		}

		memberList, err := newRelicList(members, "users", "group member collection")
		if err != nil {
			return err
		}

		for _, member := range memberList {
			realID, ok := member["id"].(string)
			if !ok {
				return fmt.Errorf("recorded new relic group member has no string id")
			}

			// The same mapping as the roster, so the join still holds.
			if err := newRelicRewrite(member, "id", s.userID(realID), "group member"); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestNewRelicDriver(t *testing.T) {
	t.Parallel()

	sanitizer := newNewRelicSanitizer()
	rec := newRecorder(t, "testdata/newrelic", "NEW_RELIC_API_KEY", sanitizer.sanitize)
	// New Relic authenticates via the API-Key header, not Authorization.
	client := newVCRClientWithHeader(rec, "API-Key", os.Getenv("NEW_RELIC_API_KEY"))

	driver := NewNewRelicDriver(client, newRelicEUEndpoint)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)

	// Cassette recorded live against an EU account, then anonymized: one user,
	// who belongs to the built-in Admin group.
	member := records[0]
	assert.Equal(t, "1000000001", member.ExternalID)
	assert.Equal(t, "1000000001@example.com", member.Email)
	assert.Equal(t, "Member 1000000001", member.FullName)
	assert.Equal(t, new(true), member.IsAdmin)
	// Group membership is the role.
	assert.Equal(t, []string{"Admin"}, member.Roles)
	assert.Equal(t, coredata.MFAStatusUnknown, member.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, member.AccountType)
}

func TestNewRelicNameResolver(t *testing.T) {
	t.Parallel()

	sanitizer := newNewRelicSanitizer()
	rec := newRecorder(t, "testdata/newrelic_name", "NEW_RELIC_API_KEY", sanitizer.sanitize)
	client := newVCRClientWithHeader(rec, "API-Key", os.Getenv("NEW_RELIC_API_KEY"))

	resolver := NewNewRelicNameResolver(client, newRelicEUEndpoint)
	name, err := resolver.ResolveInstanceName(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Example Organization", name)
}

func TestNewRelicEndpoint(t *testing.T) {
	t.Parallel()

	us, err := NewRelicEndpoint("us")
	require.NoError(t, err)
	assert.Equal(t, "https://api.newrelic.com/graphql", us)

	eu, err := NewRelicEndpoint("EU")
	require.NoError(t, err)
	assert.Equal(t, "https://api.eu.newrelic.com/graphql", eu)

	jp, err := NewRelicEndpoint("jp")
	require.NoError(t, err)
	assert.Equal(t, "https://api.jp.newrelic.com/graphql", jp)

	// Surrounding whitespace is the shape a pasted setting arrives in.
	trimmed, err := NewRelicEndpoint("  eu  ")
	require.NoError(t, err)
	assert.Equal(t, "https://api.eu.newrelic.com/graphql", trimmed)

	// An unknown region must not fall back to a default: it would send an EU
	// customer's key to the US endpoint, which answers 403.
	_, err = NewRelicEndpoint("ap")
	require.Error(t, err)

	_, err = NewRelicEndpoint("")
	require.Error(t, err)
}

func TestNewRelicRoles(t *testing.T) {
	t.Parallel()

	// Roles is group membership and nothing else: New Relic documents the user
	// type as a billing factor rather than a permission.
	assert.Equal(t, []string{"Admin"}, newRelicRoles([]string{"Admin"}))
	assert.Equal(t, []string{"Admin", "User"}, newRelicRoles([]string{"Admin", "User"}))
	// No groups is no role, and never a nil slice.
	assert.Equal(t, []string{}, newRelicRoles(nil))
}

func TestNewRelicIsAdminGroup(t *testing.T) {
	t.Parallel()

	assert.True(t, newRelicIsAdminGroup("Admin"))
	// Matched exactly, including case. New Relic's built-in group is "Admin";
	// a customer is free to create one called "admin" that grants nothing, and
	// a group whose name merely contains the word grants nothing either.
	assert.False(t, newRelicIsAdminGroup("admin"))
	assert.False(t, newRelicIsAdminGroup("Billing Admin"))
	assert.False(t, newRelicIsAdminGroup("Admins"))
	assert.False(t, newRelicIsAdminGroup("User"))
}

// newRelicStub dispatches on which cursor variable a request carries, which is
// how the driver distinguishes its four queries. It records every request so a
// test can assert the driver asked for the pages it claims to follow.
func newRelicStub(t *testing.T, responses map[string]string) (*httptest.Server, *[]string) {
	t.Helper()

	asked := &[]string{}

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

		key := "domains"

		for _, cursor := range []string{"userCursor", "groupCursor", "groupUserCursor", "domainCursor"} {
			if value, ok := req.Variables[cursor].(string); ok && value != "" {
				key = value

				break
			}
		}

		*asked = append(*asked, key)

		body, ok := responses[key]
		if !ok {
			t.Errorf("driver asked for an unexpected page: %s", key)
			http.Error(w, "{}", http.StatusInternalServerError)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))

	t.Cleanup(server.Close)

	return server, asked
}

func newRelicDomainsBody(domain, users, groups, nextCursor string) string {
	return fmt.Sprintf(`{"data":{"actor":{"organization":{"userManagement":{
      "authenticationDomains":{"nextCursor":%s,"authenticationDomains":[
        {"id":%q,"name":"Default","users":%s,"groups":%s}]}}}}}}`,
		nextCursor, domain, users, groups)
}

func newRelicUsersBody(nextCursor string, users ...string) string {
	return fmt.Sprintf(`{"nextCursor":%s,"users":[%s]}`, nextCursor, strings.Join(users, ","))
}

func newRelicUserJSON(id string) string {
	return fmt.Sprintf(
		`{"id":%q,"name":"Member %s","email":"%s@example.com","lastActive":null,"type":{"id":"0","displayName":"Full platform"}}`,
		id, id, id,
	)
}

// TestNewRelicDriverPaginatesEveryLevel walks all four cursors the driver can
// follow: the authentication-domain list, a domain's users, a domain's groups,
// and a group's members.
//
// The assertion that matters is the last one. u2 appears on the SECOND page of
// the Admin group's members, so it can only be flagged as an administrator if
// the driver actually follows the group-member cursor and joins the result
// back onto the roster by user id. That join is what the whole file exists for,
// and a cassette cannot reach it: the sanitizer rewrites response ids but not
// the request variables, so a recorded multi-page cassette never replays.
func TestNewRelicDriverPaginatesEveryLevel(t *testing.T) {
	t.Parallel()

	server, asked := newRelicStub(t, map[string]string{
		// Domain 1: one user now, one behind a cursor; the Admin group has no
		// members on its first page.
		"domains": newRelicDomainsBody(
			"d1",
			newRelicUsersBody(`"users-2"`, newRelicUserJSON("u1")),
			`{"nextCursor":"groups-2","groups":[{"id":"g1","displayName":"Admin","users":{"nextCursor":"members-2","users":[]}}]}`,
			`"domains-2"`,
		),
		"users-2":   `{"data":{"actor":{"organization":{"userManagement":{"authenticationDomains":{"authenticationDomains":[{"id":"d1","users":` + newRelicUsersBody("null", newRelicUserJSON("u2")) + `}]}}}}}}`,
		"members-2": `{"data":{"actor":{"organization":{"userManagement":{"authenticationDomains":{"authenticationDomains":[{"id":"d1","groups":{"groups":[{"id":"g1","users":{"nextCursor":null,"users":[{"id":"u2"}]}}]}}]}}}}}}`,
		"groups-2":  `{"data":{"actor":{"organization":{"userManagement":{"authenticationDomains":{"authenticationDomains":[{"id":"d1","groups":{"nextCursor":null,"groups":[]}}]}}}}}}`,
		// Domain 2, reached only by following the domain cursor.
		"domains-2": newRelicDomainsBody(
			"d2",
			newRelicUsersBody("null", newRelicUserJSON("u3")),
			`{"nextCursor":null,"groups":[]}`,
			"null",
		),
	})

	driver := NewNewRelicDriver(server.Client(), server.URL)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	byID := map[string]AccountRecord{}
	for _, record := range records {
		byID[record.ExternalID] = record
	}

	require.Contains(t, byID, "u1")
	require.Contains(t, byID, "u2")
	// u3 lives in the second authentication domain, so it proves the domain
	// cursor was followed too.
	require.Contains(t, byID, "u3")

	assert.Equal(t, new(false), byID["u1"].IsAdmin)
	assert.Equal(t, new(true), byID["u2"].IsAdmin)
	assert.Equal(t, []string{"Admin"}, byID["u2"].Roles)

	assert.Equal(t, []string{"domains", "users-2", "members-2", "groups-2", "domains-2"}, *asked)
}

// TestNewRelicDriverRefusesUnrelatedTargetedPage covers the guard on a provider
// that ignores the id filter. Returning the pages gathered so far would be a
// roster missing whoever was on the pages never read, with no error at all.
func TestNewRelicDriverRefusesUnrelatedTargetedPage(t *testing.T) {
	t.Parallel()

	server, _ := newRelicStub(t, map[string]string{
		"domains": newRelicDomainsBody(
			"d1",
			newRelicUsersBody(`"users-2"`, newRelicUserJSON("u1")),
			`{"nextCursor":null,"groups":[]}`,
			"null",
		),
		// The follow-up answers with a different domain than the one asked for.
		"users-2": `{"data":{"actor":{"organization":{"userManagement":{"authenticationDomains":{"authenticationDomains":[{"id":"other","users":` + newRelicUsersBody("null", newRelicUserJSON("u9")) + `}]}}}}}}`,
	})

	driver := NewNewRelicDriver(server.Client(), server.URL)
	_, err := driver.ListAccounts(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "domain missing from its own targeted page")
}

func TestNewRelicNameResolverClassifiesFailures(t *testing.T) {
	t.Parallel()

	// A refusal that answers the same way every time retires the name now,
	// so the worker does not spend its whole budget rediscovering it.
	for _, status := range []int{
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusBadRequest,
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(status)
		}))

		name, err := NewNewRelicNameResolver(server.Client(), server.URL).ResolveInstanceName(context.Background())
		require.Errorf(t, err, "status %d", status)
		assert.Empty(t, name)
		assert.Truef(t, errors.Is(err, ErrTerminalNameResolution), "status %d should be terminal", status)

		server.Close()
	}

	// A blip must stay retryable: marking it terminal would name the source
	// generically for good on a single 500.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := NewNewRelicNameResolver(server.Client(), server.URL).ResolveInstanceName(context.Background())
	require.Error(t, err)
	assert.False(t, errors.Is(err, ErrTerminalNameResolution))
}
