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

package microsoft365

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	scimclient "go.probo.inc/probo/pkg/iam/scim/bridge/client"
)

func TestProvider_ListUsers_PhotoETagWhenAvailable(t *testing.T) {
	t.Parallel()

	const picture = "png-bytes"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/users"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{
				"value": [
					{"id":"with-photo","mail":"ada@example.com","displayName":"Ada","accountEnabled":true},
					{"id":"without-photo","mail":"grace@example.com","displayName":"Grace","accountEnabled":true}
				]
			}`)
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/photo/$value"):
			w.Header().Set("Content-Type", "image/png")
			_, _ = io.WriteString(w, picture)
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/users/with-photo/photo"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{"@odata.mediaEtag":%q}`, `W/"ada"`)
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/users/without-photo/photo"):
			http.NotFound(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	provider := New(server.Client(), nil)
	provider.graphBaseURL = server.URL

	users, err := provider.ListUsers(context.Background())

	require.NoError(t, err)
	require.Len(t, users, 2)

	byEmail := map[string]scimclient.User{}
	for _, user := range users {
		byEmail[user.UserName] = user
	}

	assert.Equal(t, `W/"ada"`, byEmail["ada@example.com"].PhotoETag)
	assert.True(t, byEmail["ada@example.com"].PhotoETagKnown)
	assert.Empty(t, byEmail["grace@example.com"].PhotoETag)
	assert.True(t, byEmail["grace@example.com"].PhotoETagKnown)

	photo, err := provider.UserPhoto(context.Background(), byEmail["ada@example.com"])

	require.NoError(t, err)
	assert.False(t, photo.Absent)
	assert.Equal(t, picture, string(photo.Body))
	assert.Equal(t, "image/png", photo.MIME)
	assert.Equal(t, `W/"ada"`, photo.ETag)
}

func TestProvider_ListUsers_SkipsPhotosWhenForbidden(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/users") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{
				"value": [
					{"id":"user-1","mail":"ada@example.com","displayName":"Ada","accountEnabled":true}
				]
			}`)

			return
		}

		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	t.Cleanup(server.Close)

	provider := New(server.Client(), nil)
	provider.graphBaseURL = server.URL

	users, err := provider.ListUsers(context.Background())

	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "ada@example.com", users[0].UserName)
	assert.Empty(t, users[0].PhotoETag)
	assert.False(t, users[0].PhotoETagKnown)
}

func TestProvider_UserPhoto_AbsentWhenMissing(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	t.Cleanup(server.Close)

	provider := New(server.Client(), nil)
	provider.graphBaseURL = server.URL

	photo, err := provider.UserPhoto(
		context.Background(),
		scimclient.User{ExternalID: "missing"},
	)

	require.NoError(t, err)
	assert.True(t, photo.Absent)
	assert.Empty(t, photo.Body)
}

func TestProvider_ListUsers_KeepsPhotoVersionOnTransientError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/users"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{
				"value": [
					{"id":"flaky","mail":"ada@example.com","displayName":"Ada","accountEnabled":true},
					{"id":"steady","mail":"grace@example.com","displayName":"Grace","accountEnabled":true}
				]
			}`)
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/users/flaky/photo"):
			http.Error(w, "unavailable", http.StatusBadGateway)
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/users/steady/photo"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"@odata.mediaEtag":"W/\"grace\""}`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	provider := New(server.Client(), nil)
	provider.graphBaseURL = server.URL
	provider.photoLookupConcurrency = 1

	users, err := provider.ListUsers(context.Background())

	require.NoError(t, err)
	require.Len(t, users, 2)

	byEmail := map[string]scimclient.User{}
	for _, user := range users {
		byEmail[user.UserName] = user
	}

	assert.False(t, byEmail["ada@example.com"].PhotoETagKnown)
	assert.Empty(t, byEmail["ada@example.com"].PhotoETag)
	assert.True(t, byEmail["grace@example.com"].PhotoETagKnown)
	assert.Equal(t, `W/"grace"`, byEmail["grace@example.com"].PhotoETag)
}

func TestProvider_ListUsers_StopsPhotoLookupAfterRepeatedAuthFailures(t *testing.T) {
	t.Parallel()

	var photoRequests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/users") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{
				"value": [
					{"id":"user-1","mail":"ada@example.com","displayName":"Ada","accountEnabled":true},
					{"id":"user-2","mail":"grace@example.com","displayName":"Grace","accountEnabled":true}
				]
			}`)

			return
		}

		photoRequests.Add(1)
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	t.Cleanup(server.Close)

	provider := New(server.Client(), nil)
	provider.graphBaseURL = server.URL
	provider.photoLookupConcurrency = 1
	provider.photoAuthFailureLimit = 1

	users, err := provider.ListUsers(context.Background())

	require.NoError(t, err)
	require.Len(t, users, 2)
	assert.Equal(t, int32(1), photoRequests.Load())
	assert.False(t, users[0].PhotoETagKnown)
	assert.False(t, users[1].PhotoETagKnown)
	assert.Empty(t, users[0].PhotoETag)
	assert.Empty(t, users[1].PhotoETag)
}
