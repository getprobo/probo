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
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
)

const (
	frontTestBaseURL       = "https://api2.frontapp.com"
	frontFixtureHost       = "example.api.frontapp.com"
	frontFixtureCompanyID  = "cmp_fixture"
	frontFixtureCompany    = "Example Company"
	frontFixtureTeammateID = "tea_fixture%d"
)

// TestFrontDriver replays a recording of a live Front company holding an admin
// and an invited teammate who never signed in. A blocked teammate also existed
// at record time: GET /teammates leaves it out.
func TestFrontDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/front", "FRONT_TOKEN", sanitizeFront)
	client := newVCRClient(rec, bearerAuth(os.Getenv("FRONT_TOKEN")))

	records, err := NewFrontDriver(client, frontTestBaseURL).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 2)

	// Record mode asserts against the live company; only the sanitized
	// cassette may reach a failure message.
	if os.Getenv("FRONT_TOKEN") != "" {
		return
	}

	admin := records[0]
	assert.Equal(t, "tea_fixture1", admin.ExternalID)
	assert.Equal(t, "teammate1@example.com", admin.Email)
	assert.Equal(t, "Teammate 1", admin.FullName)
	assert.Equal(t, []string{"Admin"}, admin.Roles)
	assert.Equal(t, new(true), admin.IsAdmin)
	assert.Equal(t, new(true), admin.Active)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, admin.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodUnknown, admin.AuthMethod)
	assert.Equal(t, coredata.MFAStatusUnknown, admin.MFAStatus)
	assert.Nil(t, admin.LastLogin)
	assert.Nil(t, admin.CreatedAt)

	invited := records[1]
	assert.Equal(t, "tea_fixture2", invited.ExternalID)
	assert.Equal(t, "teammate2@example.com", invited.Email)
	assert.Equal(t, "teammate2", invited.FullName)
	assert.Equal(t, []string{"Teammate"}, invited.Roles)
	assert.Equal(t, new(false), invited.IsAdmin)
	assert.Equal(t, new(true), invited.Active)
}

func TestFrontNameResolver(t *testing.T) {
	t.Parallel()

	// Its own cassette: two recorders on one path race on Stop.
	rec := newRecorder(t, "testdata/front_me", "FRONT_TOKEN", sanitizeFront)
	client := newVCRClient(rec, bearerAuth(os.Getenv("FRONT_TOKEN")))

	name, err := NewFrontNameResolver(client, frontTestBaseURL).ResolveInstanceName(context.Background())
	require.NoError(t, err)

	if os.Getenv("FRONT_TOKEN") == "" {
		assert.Equal(t, frontFixtureCompany, name)
	}
}

func TestFrontNameResolverTerminalOnClientError(t *testing.T) {
	t.Parallel()

	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Body:       io.NopCloser(strings.NewReader(`{"_error":{"status":401}}`)),
		}, nil
	})}

	name, err := NewFrontNameResolver(client, frontTestBaseURL).ResolveInstanceName(context.Background())
	require.ErrorIs(t, err, ErrTerminalNameResolution)
	assert.Empty(t, name)
}

// TestFrontDriverRefusesPartialRoster pins the completeness guarantee: a short
// answer returned as a success would mark every missing teammate removed.
func TestFrontDriverRefusesPartialRoster(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		status int
		body   string
	}{
		{name: "server error", status: http.StatusInternalServerError, body: `{"_error":{"status":500}}`},
		{name: "malformed body", status: http.StatusOK, body: `{"_results":`},
		{name: "no teammates array", status: http.StatusOK, body: `{}`},
		{name: "null teammates array", status: http.StatusOK, body: `{"_results":null}`},
		{name: "teammate without id", status: http.StatusOK, body: `{"_results":[{"email":"jane@example.com","type":"user"}]}`},
		{
			name:   "paginated",
			status: http.StatusOK,
			body:   `{"_pagination":{"next":"https://api2.frontapp.com/teammates?page_token=abc"},"_results":[{"id":"tea_1","type":"user"}]}`,
		},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: tc.status,
						Body:       io.NopCloser(strings.NewReader(tc.body)),
					}, nil
				})}

				records, err := NewFrontDriver(client, frontTestBaseURL).ListAccounts(context.Background())
				require.Error(t, err)
				assert.Nil(t, records)
			},
		)
	}
}

func TestFrontDriverContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		cancel()

		return nil, ctx.Err()
	})}

	records, err := NewFrontDriver(client, frontTestBaseURL).ListAccounts(ctx)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
	assert.Nil(t, records)
}

func TestFrontAccountType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		teammateType string
		want         coredata.AccessReviewEntryAccountType
	}{
		{teammateType: "user", want: coredata.AccessReviewEntryAccountTypeUser},
		{teammateType: "visitor", want: coredata.AccessReviewEntryAccountTypeUser},
		{teammateType: "", want: coredata.AccessReviewEntryAccountTypeUser},
		{teammateType: "something_new", want: coredata.AccessReviewEntryAccountTypeUser},
		{teammateType: "ai", want: coredata.AccessReviewEntryAccountTypeServiceAccount},
		{teammateType: "api", want: coredata.AccessReviewEntryAccountTypeServiceAccount},
		{teammateType: "application", want: coredata.AccessReviewEntryAccountTypeServiceAccount},
		{teammateType: "bulk_reply", want: coredata.AccessReviewEntryAccountTypeServiceAccount},
		{teammateType: "csat", want: coredata.AccessReviewEntryAccountTypeServiceAccount},
		{teammateType: "integration", want: coredata.AccessReviewEntryAccountTypeServiceAccount},
		{teammateType: "macro", want: coredata.AccessReviewEntryAccountTypeServiceAccount},
		{teammateType: "rule", want: coredata.AccessReviewEntryAccountTypeServiceAccount},
		{teammateType: "smart_csat", want: coredata.AccessReviewEntryAccountTypeServiceAccount},
	}

	for _, tc := range cases {
		t.Run(
			tc.teammateType,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tc.want, frontAccountType(tc.teammateType))
			},
		)
	}
}

func TestSanitizeFront(t *testing.T) {
	t.Parallel()

	interaction := func(path, body string) *cassette.Interaction {
		return &cassette.Interaction{
			Request:  cassette.Request{URL: frontTestBaseURL + path},
			Response: cassette.Response{Code: http.StatusOK, Body: body},
		}
	}

	t.Run(
		"rewrites every identity",
		func(t *testing.T) {
			t.Parallel()

			i := interaction(
				frontTeammatesPath,
				`{"_links":{"self":"https://acme.api.frontapp.com/teammates"},"_pagination":{"next":null},"_results":[{"_links":{"self":"https://acme.api.frontapp.com/teammates/tea_real"},"custom_fields":{"Phone":"+33600000000"},"email":"jane@acme.io","first_name":"Jane","id":"tea_real","is_admin":true,"is_available":true,"is_blocked":false,"last_name":"Doe","type":"user","username":"jane.doe"}]}`,
			)

			require.NoError(t, sanitizeFront(i))

			for _, leaked := range []string{"acme", "tea_real", "jane", "Jane", "Doe", "+33600000000"} {
				assert.NotContains(t, i.Response.Body, leaked)
			}
		},
	)

	t.Run(
		"refuses an unclassified teammate field",
		func(t *testing.T) {
			t.Parallel()

			i := interaction(
				frontTeammatesPath,
				`{"_results":[{"id":"tea_real","email":"jane@acme.io","phone_number":"+33600000000"}]}`,
			)

			require.Error(t, sanitizeFront(i))
		},
	)

	t.Run(
		"refuses a next-page cursor",
		func(t *testing.T) {
			t.Parallel()

			i := interaction(
				frontTeammatesPath,
				`{"_pagination":{"next":"https://acme.api.frontapp.com/teammates?page_token=abc"},"_results":[{"id":"tea_real"}]}`,
			)

			require.Error(t, sanitizeFront(i))
		},
	)

	t.Run(
		"refuses an unclassified company field",
		func(t *testing.T) {
			t.Parallel()

			i := interaction(frontMePath, `{"id":"cmp_real","name":"Acme","domain":"acme.io"}`)

			require.Error(t, sanitizeFront(i))
		},
	)
}

// sanitizeFront rewrites every identity a Front response carries: teammate
// ids, emails, usernames and names, the company id and name, and the company
// subdomain Front puts in every _links URL.
func sanitizeFront(i *cassette.Interaction) error {
	if i.Response.Code != http.StatusOK {
		return fmt.Errorf("refusing to sanitize front response with status %d", i.Response.Code)
	}

	var body map[string]any
	if err := json.Unmarshal([]byte(i.Response.Body), &body); err != nil {
		return fmt.Errorf("cannot decode recorded front response: %w", err)
	}

	var recorded []string

	switch {
	case strings.HasSuffix(i.Request.URL, frontTeammatesPath):
		if err := frontRejectUnknownFields(body, "teammates response", "_links", "_pagination", "_results"); err != nil {
			return err
		}

		if pagination, ok := body["_pagination"].(map[string]any); ok && pagination["next"] != nil {
			return fmt.Errorf("refusing to record a front next-page cursor")
		}

		teammates, ok := body["_results"].([]any)
		if !ok || len(teammates) == 0 {
			return fmt.Errorf("recorded front response lists no teammates")
		}

		for idx, raw := range teammates {
			teammate, ok := raw.(map[string]any)
			if !ok {
				return fmt.Errorf("recorded front teammate %d is not an object", idx)
			}

			if err := frontRejectUnknownFields(
				teammate,
				fmt.Sprintf("teammate %d", idx),
				"_links", "custom_fields", "email", "first_name", "id", "is_admin",
				"is_available", "is_blocked", "last_name", "type", "username",
			); err != nil {
				return err
			}

			n := idx + 1

			replacements := map[string]string{
				"id":         fmt.Sprintf(frontFixtureTeammateID, n),
				"email":      fmt.Sprintf("teammate%d@example.com", n),
				"username":   fmt.Sprintf("teammate%d", n),
				"first_name": "",
				"last_name":  "",
			}
			firstName, _ := teammate["first_name"].(string)

			lastName, _ := teammate["last_name"].(string)
			if firstName != "" || lastName != "" {
				replacements["first_name"] = "Teammate"
				replacements["last_name"] = fmt.Sprintf("%d", n)
			}

			id, _ := teammate["id"].(string)
			for field, replacement := range replacements {
				if value, _ := teammate[field].(string); len(value) > 2 {
					recorded = append(recorded, value)
				}

				teammate[field] = replacement
			}

			teammate["custom_fields"] = map[string]any{}
			if err := sanitizeFrontLinks(teammate, id, fmt.Sprintf(frontFixtureTeammateID, n)); err != nil {
				return err
			}
		}
	case strings.HasSuffix(i.Request.URL, frontMePath):
		if err := frontRejectUnknownFields(body, "company response", "_links", "id", "name"); err != nil {
			return err
		}

		for field, replacement := range map[string]string{"id": frontFixtureCompanyID, "name": frontFixtureCompany} {
			value, ok := body[field].(string)
			if !ok || value == "" {
				return fmt.Errorf("recorded front company has no %s", field)
			}

			recorded = append(recorded, value)
			body[field] = replacement
		}
	default:
		return fmt.Errorf("refusing to sanitize an unrecognised front interaction")
	}

	if err := sanitizeFrontLinks(body, "", ""); err != nil {
		return err
	}

	sanitized, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("cannot re-encode front response: %w", err)
	}

	for _, value := range recorded {
		if strings.Contains(string(sanitized), value) {
			return fmt.Errorf("sanitized front response still contains a recorded identity value")
		}
	}

	replaceCassetteBody(i, string(sanitized))

	// The Etag hashes the body as recorded, before sanitizing.
	i.Response.Headers.Del("Etag")

	return nil
}

// frontRejectUnknownFields refuses a field the sanitizer does not classify, so
// an identity Front adds later is never recorded verbatim.
func frontRejectUnknownFields(object map[string]any, what string, known ...string) error {
	for field := range object {
		if !slices.Contains(known, field) {
			return fmt.Errorf("refusing to record unclassified field %q in front %s", field, what)
		}
	}

	return nil
}

// sanitizeFrontLinks points every _links URL at the fixture host and swaps the
// teammate id in its path.
func sanitizeFrontLinks(object map[string]any, id, fixtureID string) error {
	links, ok := object["_links"].(map[string]any)
	if !ok {
		return nil
	}

	var rewrite func(map[string]any) error

	rewrite = func(m map[string]any) error {
		for key, value := range m {
			switch v := value.(type) {
			case map[string]any:
				if err := rewrite(v); err != nil {
					return err
				}
			case string:
				u, err := url.Parse(v)
				if err != nil {
					return fmt.Errorf("cannot parse recorded front link: %w", err)
				}

				u.Host = frontFixtureHost
				if id != "" {
					u.Path = strings.ReplaceAll(u.Path, id, fixtureID)
				}

				m[key] = u.String()
			}
		}

		return nil
	}

	return rewrite(links)
}
