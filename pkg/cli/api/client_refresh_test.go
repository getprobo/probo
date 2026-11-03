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

package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cli/api"
)

func TestDoRaw_RefreshesExpiredAccessToken(t *testing.T) {
	t.Parallel()

	var refreshes atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			refreshes.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "access-2",
				"refresh_token": "refresh-2",
				"expires_in":    3600,
				"token_type":    "Bearer",
			})

			return
		}

		if r.Header.Get("Authorization") != "Bearer access-2" {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"ok":true}}`))
	}))
	t.Cleanup(server.Close)

	var gotAccess, gotRefresh string
	var gotExpiry time.Time

	client := api.NewClient(
		server.URL,
		"access-1",
		"/graphql",
		time.Second,
		api.WithTokenRefresher(&api.TokenRefresher{
			RefreshToken:  "refresh-1",
			TokenEndpoint: server.URL + "/token",
			ClientID:      "client",
			OnRefresh: func(accessToken, refreshToken string, expiresAt time.Time) error {
				gotAccess = accessToken
				gotRefresh = refreshToken
				gotExpiry = expiresAt

				return nil
			},
		}),
	)

	body, err := client.DoRaw("query { ok }", nil)
	require.NoError(t, err)
	assert.JSONEq(t, `{"data":{"ok":true}}`, string(body))
	assert.Equal(t, int32(1), refreshes.Load())
	assert.Equal(t, "access-2", gotAccess)
	assert.Equal(t, "refresh-2", gotRefresh)
	assert.False(t, gotExpiry.IsZero())
}

func TestDoRaw_KeepsRefreshTokenWhenResponseOmitsIt(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "access-2",
				"expires_in":   3600,
				"token_type":   "Bearer",
			})

			return
		}

		if r.Header.Get("Authorization") != "Bearer access-2" {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		_, _ = w.Write([]byte(`{"data":{"ok":true}}`))
	}))
	t.Cleanup(server.Close)

	var gotRefresh string

	client := api.NewClient(
		server.URL,
		"access-1",
		"/graphql",
		time.Second,
		api.WithTokenRefresher(&api.TokenRefresher{
			RefreshToken:  "refresh-1",
			TokenEndpoint: server.URL + "/token",
			ClientID:      "client",
			OnRefresh: func(_, refreshToken string, _ time.Time) error {
				gotRefresh = refreshToken

				return nil
			},
		}),
	)

	_, err := client.DoRaw("query { ok }", nil)
	require.NoError(t, err)
	assert.Equal(t, "refresh-1", gotRefresh)
}

func TestDoRaw_OverlappingRefreshesUseTheSavedToken(t *testing.T) {
	t.Parallel()

	var (
		mu        sync.Mutex
		stored    = api.StoredCredentials{AccessToken: "expired", RefreshToken: "refresh-1"}
		refreshes atomic.Int32
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			refreshes.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "access-2",
				"refresh_token": "refresh-2",
				"expires_in":    3600,
				"token_type":    "Bearer",
			})

			return
		}

		if r.Header.Get("Authorization") != "Bearer access-2" {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		_, _ = w.Write([]byte(`{"data":{"ok":true}}`))
	}))
	t.Cleanup(server.Close)

	newClient := func() *api.Client {
		return api.NewClient(
			server.URL,
			"expired",
			"/graphql",
			time.Second,
			api.WithTokenRefresher(&api.TokenRefresher{
				RefreshToken:  "refresh-1",
				TokenEndpoint: server.URL + "/token",
				ClientID:      "client",
				ExpiresAt:     time.Now().Add(-time.Hour),
				Lock: func() (func(), error) {
					mu.Lock()

					return mu.Unlock, nil
				},
				Current: func() (api.StoredCredentials, error) {
					return stored, nil
				},
				OnRefresh: func(accessToken, refreshToken string, expiresAt time.Time) error {
					stored = api.StoredCredentials{
						AccessToken:  accessToken,
						RefreshToken: refreshToken,
						ExpiresAt:    expiresAt,
					}

					return nil
				},
			}),
		)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	for range 2 {
		go func() {
			defer wg.Done()

			body, err := newClient().DoRaw("query { ok }", nil)
			assert.NoError(t, err)
			assert.JSONEq(t, `{"data":{"ok":true}}`, string(body))
		}()
	}

	wg.Wait()
	assert.Equal(t, int32(1), refreshes.Load())
}
