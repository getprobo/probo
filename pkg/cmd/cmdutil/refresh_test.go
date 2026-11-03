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

package cmdutil_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cli/config"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

func TestAuthenticatedCommandsRefreshTokens(t *testing.T) {
	t.Parallel()

	var missing []string

	err := filepath.WalkDir("..", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		if filepath.Base(path) == "login.go" && strings.Contains(path, "/auth/login/") {
			return nil
		}

		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		text := string(body)
		if strings.Contains(text, "api.NewClient(") && !strings.Contains(text, "TokenRefreshOption(") {
			missing = append(missing, path)
		}

		return nil
	})
	require.NoError(t, err)
	assert.Empty(t, missing)
}

func TestTokenRefreshOption_OverlappingCommandsRefreshOnce(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("PROBO_HOST", "")
	t.Setenv("PROBO_TOKEN", "")

	var refreshes atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/token") {
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

	cfg := &config.Config{
		ActiveHost: server.URL,
		Hosts: map[string]*config.HostConfig{
			server.URL: {
				Token:         "expired",
				RefreshToken:  "refresh-1",
				TokenEndpoint: server.URL + "/token",
				ExpiresAt:     time.Now().Add(-time.Hour),
			},
		},
	}
	require.NoError(t, cfg.Save())

	loaded, err := config.Load()
	require.NoError(t, err)

	host, hc, err := loaded.DefaultHost()
	require.NoError(t, err)

	clients := []*api.Client{
		api.NewClient(
			host,
			"expired",
			"/graphql",
			time.Second,
			cmdutil.TokenRefreshOption(loaded, host, hc),
		),
		api.NewClient(
			host,
			"expired",
			"/graphql",
			time.Second,
			cmdutil.TokenRefreshOption(loaded, host, hc),
		),
	}

	var wg sync.WaitGroup
	wg.Add(len(clients))

	for _, client := range clients {
		go func(client *api.Client) {
			defer wg.Done()

			body, err := client.DoRaw("query { ok }", nil)
			assert.NoError(t, err)
			assert.JSONEq(t, `{"data":{"ok":true}}`, string(body))
		}(client)
	}

	wg.Wait()
	assert.Equal(t, int32(1), refreshes.Load())

	fresh, err := config.Load()
	require.NoError(t, err)
	saved := fresh.Hosts[host]
	require.NotNil(t, saved)
	assert.Equal(t, "access-2", saved.Token)
	assert.Equal(t, "refresh-2", saved.RefreshToken)
	assert.True(t, saved.ExpiresAt.After(time.Now()))
}
