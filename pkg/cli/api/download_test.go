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
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestParseFileID(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.NewTenantID(), coredata.FileEntityType).String()

	id, err := api.ParseFileID(fileID)
	require.NoError(t, err)
	assert.Equal(t, fileID, id)

	_, err = api.ParseFileID("/api/files/v1/" + fileID)
	require.Error(t, err)

	_, err = api.ParseFileID("1234")
	require.Error(t, err)
}

func TestDownloadFile_DropsTokenOnStorageRedirect(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.NewTenantID(), coredata.FileEntityType).String()

	var storageAuth string

	storage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		storageAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", `attachment; filename="evidence.pdf"; filename*=UTF-8''evidence.pdf`)
		_, _ = w.Write([]byte("pdf-bytes"))
	}))
	t.Cleanup(storage.Close)

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		if r.URL.Path != "/api/files/v1/"+fileID {
			http.NotFound(w, r)

			return
		}

		http.Redirect(w, r, storage.URL+"/object", http.StatusTemporaryRedirect)
	}))
	t.Cleanup(apiServer.Close)

	client := api.NewClient(apiServer.URL, "token", "/api/console/v1/graphql", time.Second)

	body, err := client.DownloadFile(context.Background(), fileID)
	require.NoError(t, err)
	t.Cleanup(func() { _ = body.Body.Close() })

	payload, err := io.ReadAll(body.Body)
	require.NoError(t, err)
	assert.Equal(t, "pdf-bytes", string(payload))
	assert.Equal(t, "evidence.pdf", body.FileName)
	assert.Equal(t, "application/pdf", body.ContentType)
	assert.Empty(t, storageAuth)
}

func TestDownloadFile_KeepsTokenOnSameHostRedirect(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.NewTenantID(), coredata.FileEntityType).String()

	var redirectedAuth string

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/files/v1/" + fileID:
			http.Redirect(w, r, "/blob", http.StatusTemporaryRedirect)
		case "/blob":
			redirectedAuth = r.Header.Get("Authorization")
			_, _ = w.Write([]byte("same-host"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(apiServer.Close)

	client := api.NewClient(apiServer.URL, "token", "/api/console/v1/graphql", time.Second)

	body, err := client.DownloadFile(context.Background(), fileID)
	require.NoError(t, err)
	t.Cleanup(func() { _ = body.Body.Close() })

	payload, err := io.ReadAll(body.Body)
	require.NoError(t, err)
	assert.Equal(t, "same-host", string(payload))
	assert.Equal(t, "Bearer token", redirectedAuth)
}

func TestDownloadFile_ReplacesControlCharactersInName(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.NewTenantID(), coredata.FileEntityType).String()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''evil%1B%5B2Jname%0A.pdf")
		_, _ = w.Write([]byte("bytes"))
	}))
	t.Cleanup(server.Close)

	client := api.NewClient(server.URL, "token", "/api/console/v1/graphql", time.Second)

	body, err := client.DownloadFile(context.Background(), fileID)
	require.NoError(t, err)
	t.Cleanup(func() { _ = body.Body.Close() })

	assert.Equal(t, "evil_[2Jname_.pdf", body.FileName)
}

func TestDownloadFile_RefreshesExpiredAccessToken(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.NewTenantID(), coredata.FileEntityType).String()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
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

		_, _ = w.Write([]byte("fresh"))
	}))
	t.Cleanup(server.Close)

	client := api.NewClient(
		server.URL,
		"access-1",
		"/api/console/v1/graphql",
		time.Second,
		api.WithTokenRefresher(&api.TokenRefresher{
			RefreshToken:  "refresh-1",
			TokenEndpoint: server.URL + "/token",
			ClientID:      "client",
			OnRefresh:     func(string, string, time.Time) error { return nil },
		}),
	)

	body, err := client.DownloadFile(context.Background(), fileID)
	require.NoError(t, err)
	t.Cleanup(func() { _ = body.Body.Close() })

	payload, err := io.ReadAll(body.Body)
	require.NoError(t, err)
	assert.Equal(t, "fresh", string(payload))
}

func TestDownloadFile_NotFound(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.NewTenantID(), coredata.FileEntityType).String()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	t.Cleanup(server.Close)

	client := api.NewClient(server.URL, "token", "/api/console/v1/graphql", time.Second)

	_, err := client.DownloadFile(context.Background(), fileID)
	require.Error(t, err)
	assert.ErrorContains(t, err, "file not found")
}
