// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

func TestQoveryNameResolver(t *testing.T) {
	t.Parallel()

	t.Run("empty organization id returns nothing without HTTP call", func(t *testing.T) {
		t.Parallel()

		client := &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			t.Fatalf("resolver should not make an HTTP call for an empty organization id")
			return nil, nil
		})}

		got, err := NewQoveryNameResolver(client, "").ResolveInstanceName(context.Background())
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	cases := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{
			name:   "200 returns name",
			status: http.StatusOK,
			body:   `{"id":"26ac87db-ae79-4be4-bd33-7f839f0e1647","name":"Acme Inc"}`,
			want:   "Acme Inc",
		},
		{
			name:   "401 is terminal (no error, no name)",
			status: http.StatusUnauthorized,
			body:   `{"error":"unauthorized"}`,
			want:   "",
		},
		{
			name:   "404 is terminal (no error, no name)",
			status: http.StatusNotFound,
			body:   `{"error":"not found"}`,
			want:   "",
		},
		{
			name:   "500 is terminal (no error, no name)",
			status: http.StatusInternalServerError,
			body:   `{"error":"boom"}`,
			want:   "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/organization/26ac87db-ae79-4be4-bd33-7f839f0e1647", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

			got, err := NewQoveryNameResolver(client, "26ac87db-ae79-4be4-bd33-7f839f0e1647").ResolveInstanceName(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestRenderNameResolver(t *testing.T) {
	t.Parallel()

	t.Run("empty owner id returns nothing without HTTP call", func(t *testing.T) {
		t.Parallel()

		client := &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			t.Fatalf("resolver should not make an HTTP call for an empty owner id")
			return nil, nil
		})}

		got, err := NewRenderNameResolver(client, "").ResolveInstanceName(context.Background())
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	cases := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{
			name:   "200 returns name",
			status: http.StatusOK,
			body:   `{"id":"tea-test","name":"Acme Workspace","email":"ops@example.com","type":"team"}`,
			want:   "Acme Workspace",
		},
		{
			name:   "401 is terminal (no error, no name)",
			status: http.StatusUnauthorized,
			body:   `{"message":"unauthorized"}`,
			want:   "",
		},
		{
			name:   "404 is terminal (no error, no name)",
			status: http.StatusNotFound,
			body:   `{"message":"not found"}`,
			want:   "",
		},
		{
			name:   "500 is terminal (no error, no name)",
			status: http.StatusInternalServerError,
			body:   `{"message":"boom"}`,
			want:   "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/v1/owners/tea-test", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

			got, err := NewRenderNameResolver(client, "tea-test").ResolveInstanceName(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNeonNameResolver(t *testing.T) {
	t.Parallel()

	t.Run("empty organization id returns nothing without HTTP call", func(t *testing.T) {
		t.Parallel()

		client := &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			t.Fatalf("resolver should not make an HTTP call for an empty organization id")
			return nil, nil
		})}

		got, err := NewNeonNameResolver(client, "").ResolveInstanceName(context.Background())
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	cases := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{
			name:   "200 returns name",
			status: http.StatusOK,
			body:   `{"id":"org-cool-breeze-12345678","name":"Acme Inc","handle":"acme-inc-org-cool-breeze-12345678","plan":"launch"}`,
			want:   "Acme Inc",
		},
		{
			name:   "401 is terminal (no error, no name)",
			status: http.StatusUnauthorized,
			body:   `{"error":"unauthorized"}`,
			want:   "",
		},
		{
			name:   "404 is terminal (no error, no name)",
			status: http.StatusNotFound,
			body:   `{"error":"not found"}`,
			want:   "",
		},
		{
			name:   "500 is terminal (no error, no name)",
			status: http.StatusInternalServerError,
			body:   `{"error":"boom"}`,
			want:   "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/api/v2/organizations/org-cool-breeze-12345678", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

			got, err := NewNeonNameResolver(client, "org-cool-breeze-12345678").ResolveInstanceName(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestTailscaleNameResolver(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		status  int
		body    string
		want    string
		wantErr bool
	}{
		{
			name:   "custom domain tailnet",
			status: http.StatusOK,
			body:   `{"users":[{"loginName":"jane@acme.example.com"},{"loginName":"bob@acme.example.com"}]}`,
			want:   "acme.example.com",
		},
		{
			name:   "most common domain wins",
			status: http.StatusOK,
			body:   `{"users":[{"loginName":"a@one.com"},{"loginName":"b@two.com"},{"loginName":"c@two.com"}]}`,
			want:   "two.com",
		},
		{
			name:   "no usable login names",
			status: http.StatusOK,
			body:   `{"users":[{"loginName":""},{"loginName":"tagged-device"}]}`,
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
				assert.Equal(t, "/api/v2/tailnet/-/users", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

			got, err := NewTailscaleNameResolver(client).ResolveInstanceName(context.Background())
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestHerokuNameResolver(t *testing.T) {
	t.Parallel()

	t.Run("personal-account slug returns a name without an HTTP call", func(t *testing.T) {
		t.Parallel()

		client := &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			t.Fatalf("resolver should not make an HTTP call for a personal account")
			return nil, nil
		})}

		got, err := NewHerokuNameResolver(client, herokuPersonalAccountSlug).ResolveInstanceName(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "Personal account", got)
	})

	t.Run("team slug resolves the team name", func(t *testing.T) {
		t.Parallel()

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/teams/acme", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"name":"Acme Inc"}`))
		}))
		defer srv.Close()

		client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

		got, err := NewHerokuNameResolver(client, "acme").ResolveInstanceName(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "Acme Inc", got)
	})
}

func TestGitHubNameResolver(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		org       string
		status    int
		body      string
		want      string
		wantErr   bool
		wantNoReq bool
	}{
		{
			name:      "empty org returns empty name without HTTP call",
			org:       "",
			wantNoReq: true,
			want:      "",
		},
		{
			name:   "200 with name",
			org:    "acme",
			status: http.StatusOK,
			body:   `{"name":"Acme Inc"}`,
			want:   "Acme Inc",
		},
		{
			name:   "200 with empty name falls back to org slug",
			org:    "acme",
			status: http.StatusOK,
			body:   `{"name":""}`,
			want:   "acme",
		},
		{
			name:    "404 errors",
			org:     "missing",
			status:  http.StatusNotFound,
			body:    `{"message":"Not Found"}`,
			wantErr: true,
		},
		{
			name:    "500 errors",
			org:     "acme",
			status:  http.StatusInternalServerError,
			body:    `{"message":"boom"}`,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var called bool

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true

				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/orgs/"+tc.org, r.URL.Path)
				assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

			got, err := NewGitHubNameResolver(client, tc.org).ResolveInstanceName(context.Background())
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)

			if tc.wantNoReq {
				assert.False(t, called, "expected no HTTP call when org is empty")
			} else {
				assert.True(t, called, "expected HTTP call when org is non-empty")
			}
		})
	}
}

func TestRailwayNameResolver(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{
			name:   "single workspace names the source",
			status: http.StatusOK,
			body:   `{"data":{"me":{"name":"Jane Doe","workspaces":[{"id":"w1","name":"Acme Workspace"}]}}}`,
			want:   "Acme Workspace",
		},
		{
			name:   "multiple workspaces fall back to account holder",
			status: http.StatusOK,
			body:   `{"data":{"me":{"name":"Jane Doe","workspaces":[{"id":"w1","name":"Acme"},{"id":"w2","name":"Beta"}]}}}`,
			want:   "Jane Doe",
		},
		{
			name:   "no workspaces fall back to account holder",
			status: http.StatusOK,
			body:   `{"data":{"me":{"name":"Jane Doe","workspaces":[]}}}`,
			want:   "Jane Doe",
		},
		{
			name:   "graphql error body is terminal (no name)",
			status: http.StatusOK,
			body:   `{"data":{"me":null},"errors":[{"message":"unauthorized"}]}`,
			want:   "",
		},
		{
			name:   "null me is terminal (no name)",
			status: http.StatusOK,
			body:   `{"data":{"me":null}}`,
			want:   "",
		},
		{
			name:   "non-2xx is terminal (no name)",
			status: http.StatusInternalServerError,
			body:   `{"message":"boom"}`,
			want:   "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/graphql/v2", r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

			got, err := NewRailwayNameResolver(client).ResolveInstanceName(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestCrispNameResolver(t *testing.T) {
	t.Parallel()

	t.Run("empty website id returns nothing without HTTP call", func(t *testing.T) {
		t.Parallel()

		client := &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			t.Fatalf("resolver should not make an HTTP call for an empty website id")
			return nil, nil
		})}

		got, err := NewCrispNameResolver(client, "").ResolveInstanceName(context.Background())
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	cases := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{
			name:   "200 returns website name",
			status: http.StatusOK,
			body:   `{"data":{"name":"Acme Support"}}`,
			want:   "Acme Support",
		},
		{
			name:   "401 is terminal (no name)",
			status: http.StatusUnauthorized,
			body:   `{"error":true,"reason":"not_allowed"}`,
			want:   "",
		},
		{
			name:   "404 is terminal (no name)",
			status: http.StatusNotFound,
			body:   `{"error":true,"reason":"website_not_found"}`,
			want:   "",
		},
		{
			name:   "500 is terminal (no name)",
			status: http.StatusInternalServerError,
			body:   `{"error":true,"reason":"boom"}`,
			want:   "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/v1/website/1a2b3c4d", r.URL.Path)
				assert.Equal(t, crispTierValue, r.Header.Get(crispTierHeader))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

			got, err := NewCrispNameResolver(client, "1a2b3c4d").ResolveInstanceName(context.Background())
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

func TestSquareNameResolver(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{
			name:   "200 returns business name",
			status: http.StatusOK,
			body:   `{"merchant":{"business_name":"Acme Coffee"}}`,
			want:   "Acme Coffee",
		},
		{
			name:   "401 is terminal (no name)",
			status: http.StatusUnauthorized,
			body:   `{"errors":[{"code":"UNAUTHORIZED"}]}`,
			want:   "",
		},
		{
			name:   "403 is terminal (no name)",
			status: http.StatusForbidden,
			body:   `{"errors":[{"code":"FORBIDDEN"}]}`,
			want:   "",
		},
		{
			name:   "500 is terminal (no name)",
			status: http.StatusInternalServerError,
			body:   `{"errors":[{"code":"INTERNAL_SERVER_ERROR"}]}`,
			want:   "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/v2/merchants/me", r.URL.Path)
				assert.Equal(t, squareAPIVersion, r.Header.Get("Square-Version"))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

			got, err := NewSquareNameResolver(client).ResolveInstanceName(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGoogleAnalyticsNameResolver(t *testing.T) {
	t.Parallel()

	t.Run("empty account id returns nothing without HTTP call", func(t *testing.T) {
		t.Parallel()

		client := &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			t.Fatalf("resolver should not make an HTTP call for an empty account id")
			return nil, nil
		})}

		got, err := NewGoogleAnalyticsNameResolver(client, "").ResolveInstanceName(context.Background())
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	cases := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{
			name:   "200 returns display name",
			status: http.StatusOK,
			body:   `{"displayName":"Acme Analytics"}`,
			want:   "Acme Analytics",
		},
		{
			name:   "401 is terminal (no name)",
			status: http.StatusUnauthorized,
			body:   `{"error":{"code":401}}`,
			want:   "",
		},
		{
			name:   "403 is terminal (no name)",
			status: http.StatusForbidden,
			body:   `{"error":{"code":403}}`,
			want:   "",
		},
		{
			name:   "404 is terminal (no name)",
			status: http.StatusNotFound,
			body:   `{"error":{"code":404}}`,
			want:   "",
		},
		{
			name:   "500 is terminal (no name)",
			status: http.StatusInternalServerError,
			body:   `{"error":{"code":500}}`,
			want:   "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/v1alpha/accounts/123456", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

			got, err := NewGoogleAnalyticsNameResolver(client, "123456").ResolveInstanceName(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
