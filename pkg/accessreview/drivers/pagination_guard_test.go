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
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
)

// crossHostPaginationHost is the host a spoofed response points its next-page
// cursor at. It must never receive a request: the connection's bearer token is
// attached to every request the drivers make.
const crossHostPaginationHost = "evil.example.com"

// TestDriversRefuseCrossHostPagination covers every driver that follows a
// next-page URL supplied by the provider (response body or RFC 5988 Link
// header). Each case answers the first page with a cursor pointing off-host
// and asserts the driver refuses it instead of forwarding the token, with an
// error that never echoes the attacker-controlled host.
func TestDriversRefuseCrossHostPagination(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		baseURL string
		body    string
		header  http.Header
		// wantErr defaults to the shared guard's message; drivers with
		// their own stricter guard (Okta) set their own.
		wantErr  string
		listFunc func(ctx context.Context, client *http.Client, baseURL string) error
	}{
		{
			name:    "asana members",
			baseURL: "https://app.asana.com/api/1.0",
			body:    `{"data":[{"gid":"1","name":"Alice","email":"alice@example.com"}],"next_page":{"uri":"https://evil.example.com/api/1.0/workspaces/12345/users?offset=abc"}}`,
			listFunc: func(ctx context.Context, client *http.Client, baseURL string) error {
				_, err := NewAsanaDriver(client, "12345", baseURL).ListAccounts(ctx)

				return err
			},
		},
		{
			name:    "bitbucket workspace members",
			baseURL: "https://api.bitbucket.org/2.0",
			body:    `{"values":[{"user":{"account_id":"1","display_name":"Alice"}}],"next":"https://evil.example.com/2.0/workspaces/acme/members?page=2"}`,
			listFunc: func(ctx context.Context, client *http.Client, baseURL string) error {
				_, err := NewBitbucketDriver(client, "acme", baseURL).ListAccounts(ctx)

				return err
			},
		},
		{
			name:    "bitbucket organizations",
			baseURL: "https://api.bitbucket.org/2.0",
			body:    `{"values":[{"slug":"acme","name":"Acme"}],"next":"https://evil.example.com/2.0/user/workspaces?page=2"}`,
			listFunc: func(ctx context.Context, client *http.Client, baseURL string) error {
				_, err := ListBitbucketOrganizations(ctx, client, baseURL)

				return err
			},
		},
		{
			name:    "github organization members",
			baseURL: "https://api.github.com",
			body:    `[]`,
			header: http.Header{
				"Link": []string{`<https://evil.example.com/organizations/1/members?page=2>; rel="next"`},
			},
			listFunc: func(ctx context.Context, client *http.Client, baseURL string) error {
				_, err := NewGitHubDriver(client, "acme", log.NewLogger(log.WithName("test")), baseURL).ListAccounts(ctx)

				return err
			},
		},
		{
			name:    "gitlab group members",
			baseURL: "https://gitlab.com/api/v4",
			body:    `[]`,
			header: http.Header{
				"Link": []string{`<https://evil.example.com/api/v4/groups/12345/members/all?page=2>; rel="next"`},
			},
			listFunc: func(ctx context.Context, client *http.Client, baseURL string) error {
				_, err := NewGitLabDriver(client, "12345", baseURL).ListAccounts(ctx)

				return err
			},
		},
		{
			name:    "microsoft 365 directory roles",
			baseURL: "https://graph.microsoft.com/v1.0",
			body:    `{"value":[],"@odata.nextLink":"https://evil.example.com/v1.0/directoryRoles?$skiptoken=abc"}`,
			listFunc: func(ctx context.Context, client *http.Client, baseURL string) error {
				_, err := NewMicrosoft365Driver(client, log.NewLogger(log.WithName("test")), baseURL).ListAccounts(ctx)

				return err
			},
		},
		{
			name:    "netlify account members",
			baseURL: "https://api.netlify.com/api/v1",
			body:    `[]`,
			header: http.Header{
				"Link": []string{`<https://evil.example.com/api/v1/acme/members?page=2>; rel="next"`},
			},
			listFunc: func(ctx context.Context, client *http.Client, baseURL string) error {
				_, err := NewNetlifyDriver(client, "acme", baseURL).ListAccounts(ctx)

				return err
			},
		},
		{
			// Okta targets the org's own host rather than a shared API
			// base, so it pins pagination to that host with its own
			// stricter guard (scheme and port are checked too).
			name:    "okta users",
			baseURL: "acme.okta.com",
			body:    `[]`,
			header: http.Header{
				"Link": []string{`<https://evil.example.com/api/v1/users?after=abc&limit=200>; rel="next"`},
			},
			wantErr: "cannot follow okta next-page link: invalid target",
			listFunc: func(ctx context.Context, client *http.Client, domain string) error {
				_, err := NewOktaDriver(client, domain).ListAccounts(ctx)

				return err
			},
		},
		{
			name:    "posthog organization members",
			baseURL: "https://us.posthog.com",
			body:    `{"count":0,"next":"https://evil.example.com/api/organizations/@current/members/?offset=100","results":[]}`,
			listFunc: func(ctx context.Context, client *http.Client, baseURL string) error {
				_, err := NewPostHogDriver(client, baseURL).ListAccounts(ctx)

				return err
			},
		},
		{
			name:    "sentry organization members",
			baseURL: "https://sentry.io/api/0",
			body:    `[]`,
			header: http.Header{
				"Link": []string{`<https://evil.example.com/api/0/organizations/acme/members/?&cursor=100:1:0>; rel="next"; results="true"; cursor="100:1:0"`},
			},
			listFunc: func(ctx context.Context, client *http.Client, baseURL string) error {
				_, err := NewSentryDriver(client, "acme", baseURL).ListAccounts(ctx)

				return err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var crossHostRequests int

			client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if strings.EqualFold(req.URL.Hostname(), crossHostPaginationHost) {
					crossHostRequests++
				}

				header := http.Header{}
				for k, values := range tc.header {
					header[k] = values
				}

				header.Set("Content-Type", "application/json")

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     header,
					Body:       io.NopCloser(strings.NewReader(tc.body)),
					Request:    req,
				}, nil
			})}

			wantErr := tc.wantErr
			if wantErr == "" {
				wantErr = "cross-host pagination is not allowed"
			}

			err := tc.listFunc(context.Background(), client, tc.baseURL)
			require.Error(t, err, "a cross-host next page must be refused")
			assert.Contains(t, err.Error(), wantErr)
			assert.NotContains(
				t,
				err.Error(),
				crossHostPaginationHost,
				"the error must not echo the attacker-supplied host",
			)
			assert.Zero(t, crossHostRequests, "no request may reach the cross-host next page")
		})
	}
}

// TestSameHostNextPageURLRejectsDisguisedHosts covers the references that do
// not look absolute but resolve off-host. url.URL.IsAbs() only reports whether
// a scheme is present, so a guard that consults it before resolving lets a
// protocol-relative reference through — and the connection's bearer token
// follows it.
func TestSameHostNextPageURLRejectsDisguisedHosts(t *testing.T) {
	t.Parallel()

	const base = "https://api.example.com/v1"

	tests := []struct {
		name    string
		next    string
		want    string
		wantErr bool
	}{
		{name: "protocol relative off-host", next: "//evil.example.com/members", wantErr: true},
		{name: "protocol relative same host", next: "//api.example.com/v1/members", want: "https://api.example.com/v1/members"},
		{name: "scheme downgrade", next: "http://api.example.com/v1/members", wantErr: true},
		{name: "absolute off-host", next: "https://evil.example.com/members", wantErr: true},
		{name: "embedded credentials stripped", next: "https://u" + ":p@api.example.com/v1/members", want: "https://api.example.com/v1/members"},
		// Resolving this against the base replaces the last segment, so it
		// leaves /v1 behind — see TestSameHostNextPageURLStaysOnCollection.
		{name: "relative reference re-roots off the base path", next: "members?page=2", wantErr: true},
		{name: "absolute same host", next: "https://api.example.com/v1/members?page=2", want: "https://api.example.com/v1/members?page=2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := sameHostNextPageURL("example", base, tt.next)
			if tt.wantErr {
				require.Error(t, err)
				assert.NotContains(t, err.Error(), "evil.example.com", "error must not echo the attacker-supplied host")
				assert.Empty(t, got)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestSameHostNextPageURLStaysOnCollection covers the same-host escapes that a
// host check alone lets through: traversal to another endpoint, and a relative
// reference re-rooting off a versioned base. The connection's bearer token
// rides on every one of these requests.
func TestSameHostNextPageURLStaysOnCollection(t *testing.T) {
	t.Parallel()

	const versioned = "https://graph.microsoft.com/v1.0"

	tests := []struct {
		name    string
		base    string
		next    string
		want    string
		wantErr bool
	}{
		{
			name:    "traversal to another endpoint",
			base:    versioned,
			next:    "../../../oauth2/token",
			wantErr: true,
		},
		{
			// Resolving this against the base yields /users, silently dropping
			// the version segment and paging a different API.
			name:    "relative reference re-roots off the version",
			base:    versioned,
			next:    "users?$skiptoken=X",
			wantErr: true,
		},
		{
			name: "absolute next on the collection",
			base: versioned,
			next: versioned + "/users?$skiptoken=X",
			want: versioned + "/users?$skiptoken=X",
		},
		{
			// A base with no path constrains nothing beyond the host, which is
			// the shape most providers use.
			name: "pathless base still allows any path",
			base: "https://api.github.com",
			next: "https://api.github.com/organizations/1/members?page=2",
			want: "https://api.github.com/organizations/1/members?page=2",
		},
		{
			name: "query-only next on a collection base",
			base: "https://us.posthog.com/api/organizations/@current/members",
			next: "?offset=100",
			want: "https://us.posthog.com/api/organizations/@current/members?offset=100",
		},
		{
			// A sibling path sharing a textual prefix must not pass as though
			// it were under the base.
			name:    "sibling path sharing a prefix",
			base:    "https://api.example.com/v1",
			next:    "https://api.example.com/v10/secrets",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := sameHostNextPageURL("example", tt.base, tt.next)
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, got)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
