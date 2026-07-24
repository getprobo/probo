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

package deviceagent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadOrExchangeAPIKey(t *testing.T) {
	t.Parallel()

	const (
		persistedKey = "persisted-device-key"
		serverURL    = "https://us.probo.com"
	)

	t.Run(
		"reuses key when persisted server URL matches",
		func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			require.NoError(t, SaveAPIKey(dir, persistedKey))
			require.NoError(t, SaveConfig(dir, &Config{ServerURL: serverURL}))

			client := NewClient("https://wrong.example.com", "", "probo-agent/test")

			apiKey, err := LoadOrExchangeAPIKey(
				context.Background(),
				dir,
				client,
				serverURL,
				"unused-token",
			)
			require.NoError(t, err)
			assert.Equal(t, persistedKey, apiKey)
		},
	)

	t.Run(
		"rejects key reuse when persisted server URL differs",
		func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			require.NoError(t, SaveAPIKey(dir, persistedKey))
			require.NoError(t, SaveConfig(dir, &Config{ServerURL: serverURL}))

			client := NewClient(serverURL, "", "probo-agent/test")

			_, err := LoadOrExchangeAPIKey(
				context.Background(),
				dir,
				client,
				"https://eu.probo.com",
				"unused-token",
			)
			require.ErrorIs(t, err, ErrServerURLMismatch)
		},
	)

	t.Run(
		"reuses key when server URL differs only by hostname case",
		func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			require.NoError(t, SaveAPIKey(dir, persistedKey))
			require.NoError(t, SaveConfig(dir, &Config{ServerURL: serverURL}))

			client := NewClient("https://wrong.example.com", "", "probo-agent/test")

			apiKey, err := LoadOrExchangeAPIKey(
				context.Background(),
				dir,
				client,
				"https://US.probo.com",
				"unused-token",
			)
			require.NoError(t, err)
			assert.Equal(t, persistedKey, apiKey)
		},
	)

	t.Run(
		"rejects key reuse when config is missing",
		func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			require.NoError(t, SaveAPIKey(dir, persistedKey))

			client := NewClient(serverURL, "", "probo-agent/test")

			_, err := LoadOrExchangeAPIKey(
				context.Background(),
				dir,
				client,
				serverURL,
				"unused-token",
			)
			require.ErrorIs(t, err, ErrServerURLMismatch)
		},
	)

	t.Run(
		"rejects key reuse when config is missing and server differs",
		func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			require.NoError(t, SaveAPIKey(dir, persistedKey))

			client := NewClient(serverURL, "", "probo-agent/test")

			_, err := LoadOrExchangeAPIKey(
				context.Background(),
				dir,
				client,
				"https://eu.probo.com",
				"unused-token",
			)
			require.ErrorIs(t, err, ErrServerURLMismatch)
		},
	)

	t.Run(
		"exchanges token when key is missing",
		func(t *testing.T) {
			t.Parallel()

			const exchangedKey = "exchanged-device-key"

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/agent/v1/enroll", r.URL.Path)

				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"api_key":"` + exchangedKey + `"}`))
			}))
			t.Cleanup(srv.Close)

			dir := t.TempDir()
			client := NewClient(srv.URL, "", "probo-agent/test")

			apiKey, err := LoadOrExchangeAPIKey(
				context.Background(),
				dir,
				client,
				srv.URL,
				"enroll-token",
			)
			require.NoError(t, err)
			assert.Equal(t, exchangedKey, apiKey)

			loaded, err := LoadAPIKey(dir)
			require.NoError(t, err)
			assert.Equal(t, exchangedKey, loaded)

			cfg, err := LoadConfig(dir)
			require.NoError(t, err)
			assert.Equal(t, srv.URL, cfg.ServerURL)
		},
	)

	t.Run(
		"rejects invalid server URL before exchange",
		func(t *testing.T) {
			t.Parallel()

			var hits atomic.Int32

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits.Add(1)
				w.WriteHeader(http.StatusInternalServerError)
			}))
			t.Cleanup(srv.Close)

			dir := t.TempDir()
			client := NewClient(srv.URL, "", "probo-agent/test")

			_, err := LoadOrExchangeAPIKey(
				context.Background(),
				dir,
				client,
				"ftp://example.com",
				"enroll-token",
			)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid server URL")
			assert.Equal(t, int32(0), hits.Load())

			_, err = LoadAPIKey(dir)
			require.ErrorIs(t, err, ErrKeyNotFound)
		},
	)

	t.Run(
		"rolls back key when config save fails",
		func(t *testing.T) {
			t.Parallel()

			const exchangedKey = "exchanged-device-key"

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"api_key":"` + exchangedKey + `"}`))
			}))
			t.Cleanup(srv.Close)

			dir := t.TempDir()
			require.NoError(t, os.Mkdir(filepath.Join(dir, ConfigFileName), 0o700))

			client := NewClient(srv.URL, "", "probo-agent/test")

			_, err := LoadOrExchangeAPIKey(
				context.Background(),
				dir,
				client,
				srv.URL,
				"enroll-token",
			)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot save device config")

			_, err = LoadAPIKey(dir)
			require.ErrorIs(t, err, ErrKeyNotFound)
		},
	)
}
