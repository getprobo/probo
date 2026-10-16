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

// planetScaleFixtureOrganization is Probo's own recording organization, kept as recorded.
const planetScaleFixtureOrganization = "oauth-service-accounts"

func TestPlanetScaleDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/planetscale", "PLANETSCALE_TOKEN", sanitizePlanetScale, dropResponseHeaders("Etag"))
	client := newVCRClient(rec, bearerAuth(os.Getenv("PLANETSCALE_TOKEN")))
	driver := NewPlanetScaleDriver(client, planetScaleFixtureOrganization, "https://api.planetscale.com/v1")

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)

	admin := records[0]
	assert.Equal(t, "member1@example.com", admin.Email)
	assert.Equal(t, "user00000001", admin.ExternalID)
	// The recorded user has no name, and display_name repeats the email.
	assert.Empty(t, admin.FullName)
	assert.Equal(t, []string{"admin"}, admin.Roles)
	assert.Equal(t, new(true), admin.IsAdmin)
	// Two-factor is off and the organization has no SSO.
	assert.Equal(t, coredata.MFAStatusDisabled, admin.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, admin.AccountType)
	require.NotNil(t, admin.CreatedAt)
	assert.False(t, admin.CreatedAt.IsZero())

	// Membership carries no status, and PlanetScale exposes no sign-in time.
	assert.Nil(t, admin.Active)
	assert.Nil(t, admin.LastLogin)
}

func TestListPlanetScaleOrganizations(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/planetscale_organizations", "PLANETSCALE_TOKEN", sanitizePlanetScale, dropResponseHeaders("Etag"))
	client := newVCRClient(rec, bearerAuth(os.Getenv("PLANETSCALE_TOKEN")))

	orgs, err := ListPlanetScaleOrganizations(context.Background(), client, "https://api.planetscale.com/v1")
	require.NoError(t, err)
	assert.Equal(
		t,
		[]Organization{{Slug: planetScaleFixtureOrganization, DisplayName: planetScaleFixtureOrganization}},
		orgs,
	)
}

func TestPlanetScaleDriverPaginates(t *testing.T) {
	t.Parallel()

	var pages []string

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch r.URL.Path {
			case "/v1/organizations/acme":
				_, _ = w.Write([]byte(`{"name":"acme","sso":false}`))
			case "/v1/organizations/acme/members":
				assert.Equal(t, "100", r.URL.Query().Get("per_page"))

				page := r.URL.Query().Get("page")
				pages = append(pages, page)

				switch page {
				case "1":
					_, _ = w.Write([]byte(`{"current_page":1,"next_page":2,"data":[
						{"id":"m1","role":"admin","created_at":"2026-01-02T03:04:05.000Z","user":{"id":"u1","email":"ada@example.com","display_name":"ada@example.com","name":"Ada Lovelace","two_factor_auth_configured":true}},
						{"id":"m2","role":"member","user":{"id":"u2","email":"bob@example.com","display_name":"Bob","name":null,"two_factor_auth_configured":false}}
					]}`))
				case "2":
					_, _ = w.Write([]byte(`{"current_page":2,"next_page":null,"data":[
						{"id":"m3","role":"analyst","user":{"id":"u3","email":"cy@example.com","display_name":"cy@example.com","name":null}}
					]}`))
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}),
	)
	t.Cleanup(srv.Close)

	records, err := NewPlanetScaleDriver(srv.Client(), "acme", srv.URL+"/v1").ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)
	assert.Equal(t, []string{"1", "2"}, pages)

	assert.Equal(t, "u1", records[0].ExternalID)
	assert.Equal(t, "Ada Lovelace", records[0].FullName)
	assert.Equal(t, []string{"admin"}, records[0].Roles)
	assert.Equal(t, new(true), records[0].IsAdmin)
	assert.Equal(t, coredata.MFAStatusEnabled, records[0].MFAStatus)
	require.NotNil(t, records[0].CreatedAt)

	assert.Equal(t, "u2", records[1].ExternalID)
	assert.Equal(t, "Bob", records[1].FullName)
	assert.Equal(t, []string{"member"}, records[1].Roles)
	assert.Equal(t, new(false), records[1].IsAdmin)
	assert.Equal(t, coredata.MFAStatusDisabled, records[1].MFAStatus)
	assert.Nil(t, records[1].CreatedAt)

	// A missing two-factor flag is no evidence either way.
	assert.Equal(t, "u3", records[2].ExternalID)
	assert.Empty(t, records[2].FullName)
	assert.Equal(t, []string{"analyst"}, records[2].Roles)
	assert.Equal(t, new(false), records[2].IsAdmin)
	assert.Equal(t, coredata.MFAStatusUnknown, records[2].MFAStatus)
}

func TestPlanetScaleDriverRefusesNonAdvancingPage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			if strings.HasSuffix(r.URL.Path, "/members") {
				_, _ = w.Write([]byte(`{"next_page":1,"data":[]}`))
				return
			}

			_, _ = w.Write([]byte(`{"name":"acme","sso":false}`))
		}),
	)
	t.Cleanup(srv.Close)

	_, err := NewPlanetScaleDriver(srv.Client(), "acme", srv.URL).ListAccounts(context.Background())
	require.ErrorContains(t, err, "does not follow page")
}

func TestPlanetScaleDriverRefusesMemberWithoutUserID(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			if strings.HasSuffix(r.URL.Path, "/members") {
				_, _ = w.Write([]byte(`{"next_page":null,"data":[{"id":"m1","role":"member","user":{"email":"ada@example.com"}}]}`))
				return
			}

			_, _ = w.Write([]byte(`{"name":"acme","sso":false}`))
		}),
	)
	t.Cleanup(srv.Close)

	_, err := NewPlanetScaleDriver(srv.Client(), "acme", srv.URL).ListAccounts(context.Background())
	require.ErrorContains(t, err, `membership "m1" has no user id`)
}

func TestPlanetScaleDriverSSOOrganization(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			if strings.HasSuffix(r.URL.Path, "/members") {
				_, _ = w.Write([]byte(`{"next_page":null,"data":[
					{"id":"m1","role":"member","user":{"id":"u1","email":"ada@example.com","two_factor_auth_configured":false}},
					{"id":"m2","role":"member","user":{"id":"u2","email":"bob@example.com","two_factor_auth_configured":true}}
				]}`))

				return
			}

			_, _ = w.Write([]byte(`{"name":"acme","sso":true}`))
		}),
	)
	t.Cleanup(srv.Close)

	records, err := NewPlanetScaleDriver(srv.Client(), "acme", srv.URL).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 2)
	assert.Equal(t, coredata.MFAStatusUnknown, records[0].MFAStatus)
	assert.Equal(t, coredata.MFAStatusUnknown, records[1].MFAStatus)
}

func TestPlanetScaleRoleMapping(t *testing.T) {
	t.Parallel()

	cases := []struct {
		role        string
		wantRoles   []string
		wantIsAdmin *bool
	}{
		{role: "admin", wantRoles: []string{"admin"}, wantIsAdmin: new(true)},
		{role: "Admin", wantRoles: []string{"Admin"}, wantIsAdmin: new(true)},
		{role: "member", wantRoles: []string{"member"}, wantIsAdmin: new(false)},
		{role: "analyst", wantRoles: []string{"analyst"}, wantIsAdmin: new(false)},
		// A role PlanetScale adds later reaches the reviewer unchanged, and
		// one that merely contains "admin" is not the admin role.
		{role: "billing_admin", wantRoles: []string{"billing_admin"}, wantIsAdmin: new(false)},
		{role: "", wantRoles: []string{}, wantIsAdmin: nil},
	}

	for _, tc := range cases {
		t.Run(
			tc.role,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tc.wantRoles, planetScaleRoles(tc.role))
				assert.Equal(t, tc.wantIsAdmin, planetScaleIsAdmin(tc.role))
			},
		)
	}
}

func TestPlanetScaleMFAStatus(t *testing.T) {
	t.Parallel()

	assert.Equal(t, coredata.MFAStatusEnabled, planetScaleMFAStatus(new(true), false))
	assert.Equal(t, coredata.MFAStatusUnknown, planetScaleMFAStatus(new(true), true))
	assert.Equal(t, coredata.MFAStatusDisabled, planetScaleMFAStatus(new(false), false))
	assert.Equal(t, coredata.MFAStatusUnknown, planetScaleMFAStatus(new(false), true))
	assert.Equal(t, coredata.MFAStatusUnknown, planetScaleMFAStatus(nil, false))
}

func TestPlanetScaleFullName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Ada Lovelace", planetScaleFullName(" Ada Lovelace ", "ada@example.com", "ada@example.com"))
	assert.Equal(t, "Ada", planetScaleFullName("", "Ada", "ada@example.com"))
	assert.Empty(t, planetScaleFullName("", "ADA@example.com", "ada@example.com"))
	assert.Empty(t, planetScaleFullName("", "", "ada@example.com"))
}

func TestListPlanetScaleOrganizationsSkipsUngranted(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch r.URL.Path {
			case "/organizations":
				if r.URL.Query().Get("page") == "2" {
					_, _ = w.Write([]byte(`{"next_page":null,"data":[{"name":"gamma"}]}`))
					return
				}

				_, _ = w.Write([]byte(`{"next_page":2,"data":[{"name":"alpha"},{"name":"beta"}]}`))
			case "/organizations/alpha/members", "/organizations/gamma/members":
				assert.Equal(t, "1", r.URL.Query().Get("per_page"))

				_, _ = w.Write([]byte(`{"next_page":null,"data":[]}`))
			case "/organizations/beta/members":
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"code":"forbidden","message":"User does not have permission to perform this action."}`))
			default:
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"code":"not_found"}`))
			}
		}),
	)
	t.Cleanup(srv.Close)

	orgs, err := ListPlanetScaleOrganizations(context.Background(), srv.Client(), srv.URL)
	require.NoError(t, err)
	assert.Equal(
		t,
		[]Organization{
			{Slug: "alpha", DisplayName: "alpha"},
			{Slug: "gamma", DisplayName: "gamma"},
		},
		orgs,
	)
}

func TestListPlanetScaleOrganizationsNoneGranted(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			if r.URL.Path == "/organizations" {
				_, _ = w.Write([]byte(`{"next_page":null,"data":[{"name":"alpha"}]}`))
				return
			}

			w.WriteHeader(http.StatusForbidden)
		}),
	)
	t.Cleanup(srv.Close)

	orgs, err := ListPlanetScaleOrganizations(context.Background(), srv.Client(), srv.URL)
	require.NoError(t, err)
	assert.NotNil(t, orgs)
	assert.Empty(t, orgs)
}

func TestListPlanetScaleOrganizationsDeadToken(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			if r.URL.Path == "/organizations" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			w.WriteHeader(http.StatusNotFound)
		}),
	)
	t.Cleanup(srv.Close)

	_, err := ListPlanetScaleOrganizations(context.Background(), srv.Client(), srv.URL)
	require.ErrorContains(t, err, "unexpected status 401")
}

// sanitizePlanetScale rewrites recorded identity on save and refuses any
// endpoint it has no rewriter for. Re-record, then rerun without PLANETSCALE_TOKEN.
func sanitizePlanetScale(i *cassette.Interaction) error {
	if i.Response.Code != http.StatusOK {
		return fmt.Errorf("refusing to sanitize planetscale response with status %d", i.Response.Code)
	}

	u, err := url.Parse(i.Request.URL)
	if err != nil {
		return fmt.Errorf("cannot parse recorded planetscale URL: %w", err)
	}

	segments := strings.Split(strings.TrimPrefix(u.Path, "/v1/"), "/")

	switch {
	case len(segments) == 1 && segments[0] == "organizations":
		return sanitizePlanetScaleBody(i, sanitizePlanetScaleOrganizationList)
	case len(segments) == 2 && segments[0] == "organizations":
		return sanitizePlanetScaleBody(i, sanitizePlanetScaleOrganization)
	case len(segments) == 3 && segments[0] == "organizations" && segments[2] == planetScaleMembersSegment:
		return sanitizePlanetScaleBody(i, sanitizePlanetScaleMembers)
	default:
		return fmt.Errorf("refusing to sanitize an unrecognised planetscale interaction")
	}
}

// sanitizePlanetScaleBody rewrites the decoded JSON in place and fails if a
// replaced value survives anywhere in the body.
func sanitizePlanetScaleBody(
	i *cassette.Interaction,
	rewrite func(map[string]json.RawMessage) ([]string, error),
) error {
	var body map[string]json.RawMessage
	if err := json.Unmarshal([]byte(i.Response.Body), &body); err != nil {
		return fmt.Errorf("cannot decode recorded planetscale response: %w", err)
	}

	recorded, err := rewrite(body)
	if err != nil {
		return err
	}

	sanitized, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("cannot re-encode sanitized planetscale response: %w", err)
	}

	for _, value := range recorded {
		if bytes.Contains(sanitized, []byte(value)) {
			return fmt.Errorf("sanitized planetscale response still carries a recorded identity value")
		}
	}

	replaceCassetteBody(i, string(sanitized))

	return nil
}

func sanitizePlanetScaleOrganizationList(body map[string]json.RawMessage) ([]string, error) {
	return sanitizePlanetScaleList(body, "organizations", func(idx int, org map[string]json.RawMessage) ([]string, error) {
		return sanitizePlanetScaleOrganization(org)
	})
}

func sanitizePlanetScaleMembers(body map[string]json.RawMessage) ([]string, error) {
	return sanitizePlanetScaleList(body, "members", sanitizePlanetScaleMembership)
}

func sanitizePlanetScaleList(
	body map[string]json.RawMessage,
	what string,
	rewrite func(int, map[string]json.RawMessage) ([]string, error),
) ([]string, error) {
	raw, ok := body["data"]
	if !ok {
		return nil, fmt.Errorf("recorded planetscale %s response has no data field", what)
	}

	var items []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("cannot decode recorded planetscale %s: %w", what, err)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("recorded planetscale %s response lists nothing", what)
	}

	var recorded []string

	for idx, item := range items {
		values, err := rewrite(idx, item)
		if err != nil {
			return nil, err
		}

		recorded = append(recorded, values...)
	}

	encoded, err := json.Marshal(items)
	if err != nil {
		return nil, fmt.Errorf("cannot re-encode planetscale %s: %w", what, err)
	}

	body["data"] = encoded

	return recorded, nil
}

// sanitizePlanetScaleOrganization keeps the name: it is the recorded path segment.
func sanitizePlanetScaleOrganization(org map[string]json.RawMessage) ([]string, error) {
	// The keys the rewriter has considered; any other key aborts the save.
	organizationFields := []string{
		"allow_llm_usage",
		"billing_email",
		"created_at",
		"database_count",
		"features",
		"flags",
		"has_card",
		"id",
		"idp_managed_roles",
		"idp_sso_managed_roles",
		"invoice_budget_alerts",
		"invoice_budget_amount",
		"keyspace_shard_limit",
		"managed_tenancy",
		"mcp_access_level",
		"name",
		"neki_router_replicas_per_cell_limit",
		"payment_info_required",
		"plan",
		"single_tenancy",
		"sso",
		"sso_directory",
		"type",
		"unrestricted_idp_organization_membership",
		"updated_at",
		"valid_billing_info",
	}

	for field := range org {
		if !slices.Contains(organizationFields, field) {
			return nil, fmt.Errorf("recorded planetscale organization has an unreviewed field %q", field)
		}
	}

	var recorded []string

	for field, replacement := range map[string]string{
		"id":            "org000000001",
		"billing_email": "billing@example.com",
	} {
		value, ok, err := planetScaleStringField(org, field)
		if err != nil {
			return nil, fmt.Errorf("recorded planetscale organization: %w", err)
		}

		if !ok {
			continue
		}

		recorded = append(recorded, value)
		org[field] = mustMarshalPlanetScaleJSON(replacement)
	}

	return recorded, nil
}

func sanitizePlanetScaleMembership(idx int, membership map[string]json.RawMessage) ([]string, error) {
	// The keys the rewriter has considered; any other key aborts the save.
	membershipFields := []string{"id", "type", "user", "role", "created_at", "updated_at"}
	userFields := []string{
		"id",
		"type",
		"display_name",
		"name",
		"email",
		"created_at",
		"updated_at",
		"two_factor_auth_configured",
		"avatar_url",
	}

	for field := range membership {
		if !slices.Contains(membershipFields, field) {
			return nil, fmt.Errorf("recorded planetscale membership %d has an unreviewed field %q", idx, field)
		}
	}

	rawUser, ok := membership["user"]
	if !ok {
		return nil, fmt.Errorf("recorded planetscale membership %d has no user", idx)
	}

	var user map[string]json.RawMessage
	if err := json.Unmarshal(rawUser, &user); err != nil {
		return nil, fmt.Errorf("cannot decode recorded planetscale membership %d user: %w", idx, err)
	}

	for field := range user {
		if !slices.Contains(userFields, field) {
			return nil, fmt.Errorf("recorded planetscale user %d has an unreviewed field %q", idx, field)
		}
	}

	email, ok, err := planetScaleStringField(user, "email")
	if err != nil || !ok || email == "" {
		return nil, fmt.Errorf("recorded planetscale user %d has no email", idx)
	}

	syntheticEmail := fmt.Sprintf("member%d@example.com", idx+1)
	syntheticName := fmt.Sprintf("Member %d", idx+1)
	recorded := []string{email}

	membershipReplacements := map[string]string{"id": fmt.Sprintf("membership%02d", idx+1)}
	userReplacements := map[string]string{
		"id":           fmt.Sprintf("user%08d", idx+1),
		"email":        syntheticEmail,
		"display_name": syntheticName,
		"name":         syntheticName,
		"avatar_url":   fmt.Sprintf("https://avatars.example.com/%d.png", idx+1),
	}

	// display_name repeats the email when no name is set, and the driver acts
	// on that, so the repetition is kept.
	if displayName, ok, _ := planetScaleStringField(user, "display_name"); ok && strings.EqualFold(displayName, email) {
		userReplacements["display_name"] = syntheticEmail
	}

	for _, set := range []struct {
		object       map[string]json.RawMessage
		replacements map[string]string
	}{
		{object: membership, replacements: membershipReplacements},
		{object: user, replacements: userReplacements},
	} {
		for field, replacement := range set.replacements {
			value, ok, err := planetScaleStringField(set.object, field)
			if err != nil {
				return nil, fmt.Errorf("recorded planetscale membership %d: %w", idx, err)
			}

			// Absent or null stays as recorded: the driver reads a null name
			// as no name.
			if !ok {
				continue
			}

			if value != "" {
				recorded = append(recorded, value)
			}

			set.object[field] = mustMarshalPlanetScaleJSON(replacement)
		}
	}

	membership["user"] = mustMarshalPlanetScaleJSON(user)

	return recorded, nil
}

// planetScaleStringField reports false for an absent or null field and errors
// on any other non-string.
func planetScaleStringField(object map[string]json.RawMessage, field string) (string, bool, error) {
	raw, ok := object[field]
	if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return "", false, nil
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", false, fmt.Errorf("field %s is not a string: %w", field, err)
	}

	return value, true, nil
}

func mustMarshalPlanetScaleJSON(v any) json.RawMessage {
	encoded, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return encoded
}
