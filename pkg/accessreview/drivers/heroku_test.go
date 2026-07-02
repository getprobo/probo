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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestHerokuDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/heroku", "HEROKU_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("HEROKU_TOKEN")))

	teamID := os.Getenv("HEROKU_TEAM_ID")
	if teamID == "" {
		teamID = "acme"
	}

	driver := NewHerokuDriver(client, teamID)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, records)

	r := records[0]
	assert.NotEmpty(t, r.Email)
	assert.NotEmpty(t, r.ExternalID)
	assert.NotEmpty(t, r.FullName)
	assert.NotEmpty(t, r.Roles)
	assert.Equal(t, coredata.MFAStatusEnabled, r.MFAStatus)
	assert.True(t, r.IsAdmin)
	require.NotNil(t, r.CreatedAt)
}

// TestHerokuDriverPersonalAccount exercises personal mode (empty teamID): a
// solo account with no Team is reviewed via its apps' owner + collaborators.
// It verifies the owner is always included, collaborators are deduped across
// apps, and team-owned apps are skipped.
func TestHerokuDriverPersonalAccount(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/apps":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{"id":"app-1","name":"one","owner":{"id":"u-alice","email":"alice@example.com"},"team":null},
				{"id":"app-2","name":"two","owner":{"id":"u-alice","email":"alice@example.com"},"team":null},
				{"id":"app-3","name":"teamed","owner":{"id":"u-x","email":"x@example.com"},"team":{"id":"team-1"}}
			]`))
		case "/apps/app-1/collaborators":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{"id":"c-1","role":"owner","user":{"id":"u-alice","email":"alice@example.com"}},
				{"id":"c-2","role":"member","user":{"id":"u-bob","email":"bob@example.com"}}
			]`))
		case "/apps/app-2/collaborators":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{"id":"c-3","role":"member","user":{"id":"u-bob","email":"bob@example.com"}},
				{"id":"c-4","role":"member","user":{"id":"u-carol","email":"carol@example.com"}}
			]`))
		default:
			t.Errorf("unexpected request to %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

	records, err := NewHerokuDriver(client, "").ListAccounts(context.Background())
	require.NoError(t, err)

	byEmail := make(map[string]AccountRecord, len(records))
	for _, r := range records {
		byEmail[r.Email] = r
	}

	require.Len(t, records, 3)
	assert.Contains(t, byEmail, "alice@example.com")
	assert.Contains(t, byEmail, "bob@example.com")
	assert.Contains(t, byEmail, "carol@example.com")

	assert.True(t, byEmail["alice@example.com"].IsAdmin)
	assert.Equal(t, []string{"owner"}, byEmail["alice@example.com"].Roles)
	assert.False(t, byEmail["bob@example.com"].IsAdmin)
}

// TestHerokuDriverPersonalAccountSlug verifies the reserved personal-account
// slug routes to personal mode just like an empty teamID.
func TestHerokuDriverPersonalAccountSlug(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/apps":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"id":"app-1","name":"one","owner":{"id":"u-alice","email":"alice@example.com"},"team":null}]`))
		case "/apps/app-1/collaborators":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[]`))
		default:
			t.Errorf("unexpected request to %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

	records, err := NewHerokuDriver(client, herokuPersonalAccountSlug).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, "alice@example.com", records[0].Email)
}

// TestHerokuDriverPersonalAccountErrors verifies non-2xx responses on the
// personal-mode endpoints propagate as errors rather than empty results.
func TestHerokuDriverPersonalAccountErrors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		appsCode  int
		collabFor string
	}{
		{name: "apps list fails", appsCode: http.StatusInternalServerError},
		{name: "collaborators fail", appsCode: http.StatusOK, collabFor: "app-1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				switch r.URL.Path {
				case "/apps":
					w.WriteHeader(tc.appsCode)
					_, _ = w.Write([]byte(`[{"id":"app-1","name":"one","owner":{"id":"u-alice","email":"alice@example.com"},"team":null}]`))
				case "/apps/app-1/collaborators":
					w.WriteHeader(http.StatusForbidden)
					_, _ = w.Write([]byte(`{"id":"forbidden"}`))
				default:
					t.Errorf("unexpected request to %s", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer srv.Close()

			client := &http.Client{Transport: &hostRewriter{target: srv.URL}}

			_, err := NewHerokuDriver(client, "").ListAccounts(context.Background())
			require.Error(t, err)
		})
	}
}
