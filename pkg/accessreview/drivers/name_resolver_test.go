// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package drivers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hostRewriter redirects requests to the configured target host so that
// resolvers with hardcoded production URLs (api.notion.com, etc.) can be
// pointed at an httptest server.
type hostRewriter struct {
	target string
}

func (h *hostRewriter) RoundTrip(r *http.Request) (*http.Response, error) {
	u, err := url.Parse(h.target)
	if err != nil {
		return nil, err
	}

	r2 := r.Clone(r.Context())
	r2.URL.Scheme = u.Scheme
	r2.URL.Host = u.Host

	return http.DefaultTransport.RoundTrip(r2)
}

func TestNotionNameResolver(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		status  int
		body    string
		want    string
		wantErr bool
	}{
		{
			name:   "bot with workspace_name",
			status: http.StatusOK,
			body:   `{"type":"bot","bot":{"workspace_name":"Acme Inc"}}`,
			want:   "Acme Inc",
		},
		{
			name:   "user token (no bot field)",
			status: http.StatusOK,
			body:   `{"type":"person"}`,
			want:   "",
		},
		{
			name:    "server error",
			status:  http.StatusInternalServerError,
			body:    `{"message":"boom"}`,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/v1/users/me", r.URL.Path)
				assert.Equal(t, notionAPIVersion, r.Header.Get("Notion-Version"))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

			got, err := NewNotionNameResolver(client).ResolveInstanceName(context.Background())
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestSentryNameResolver(t *testing.T) {
	t.Parallel()

	t.Run("empty slug returns nothing without HTTP call", func(t *testing.T) {
		t.Parallel()

		client := &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			t.Fatalf("resolver should not make an HTTP call for an empty slug")
			return nil, nil
		})}

		got, err := NewSentryNameResolver(client, "").ResolveInstanceName(context.Background())
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	cases := []struct {
		name    string
		status  int
		body    string
		want    string
		wantErr bool
	}{
		{
			name:   "200 returns name",
			status: http.StatusOK,
			body:   `{"slug":"acme","name":"Acme Inc"}`,
			want:   "Acme Inc",
		},
		{
			name:   "404 is terminal (no error, no name)",
			status: http.StatusNotFound,
			body:   `{"detail":"The requested resource does not exist"}`,
			want:   "",
		},
		{
			name:    "401 is retryable",
			status:  http.StatusUnauthorized,
			body:    `{"detail":"Authentication credentials were not provided."}`,
			wantErr: true,
		},
		{
			name:    "403 is retryable",
			status:  http.StatusForbidden,
			body:    `{"detail":"You do not have permission to perform this action."}`,
			wantErr: true,
		},
		{
			name:    "500 is retryable",
			status:  http.StatusInternalServerError,
			body:    `{"detail":"Internal Server Error"}`,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/api/0/organizations/acme", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

			got, err := NewSentryNameResolver(client, "acme").ResolveInstanceName(context.Background())
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// roundTripperFunc adapts a function into an http.RoundTripper, useful for
// asserting that a resolver short-circuits before making any HTTP call.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
