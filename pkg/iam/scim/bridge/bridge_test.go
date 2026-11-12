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

package bridge_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	scimbridge "go.probo.inc/probo/pkg/iam/scim/bridge"
	scimclient "go.probo.inc/probo/pkg/iam/scim/bridge/client"
	"go.probo.inc/probo/pkg/iam/scim/bridge/provider"
)

type mockProvider struct {
	users scimclient.Users
}

func (p *mockProvider) Name() string {
	return "mock"
}

func (p *mockProvider) ListUsers(_ context.Context) (scimclient.Users, error) {
	return p.users, nil
}

func TestBridge_Run_DeletesInactiveExcludedUsers(t *testing.T) {
	t.Parallel()

	deleteCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/Users":
			w.Header().Set("Content-Type", "application/scim+json")
			_, _ = w.Write([]byte(`{
				"schemas": ["urn:ietf:params:scim:api:messages:2.0:ListResponse"],
				"totalResults": 1,
				"startIndex": 1,
				"itemsPerPage": 100,
				"Resources": [{
					"id": "gid://probo/MembershipProfile/abc",
					"userName": "excluded@example.com",
					"displayName": "Excluded User",
					"active": false,
					"externalId": "ext-1"
				}]
			}`))
		case r.Method == http.MethodDelete:
			deleteCalled = true

			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	provider := &mockProvider{users: scimclient.Users{}}
	client := scimclient.NewClient(server.Client(), server.URL, "token")
	bridge := scimbridge.NewBridge(
		provider,
		client,
		scimbridge.WithExcludedUserNames([]string{"excluded@example.com"}),
	)

	created, updated, deleted, deactivated, skipped, err := bridge.Run(context.Background())

	require.NoError(t, err)
	assert.True(t, deleteCalled)
	assert.Equal(t, 0, created)
	assert.Equal(t, 0, updated)
	assert.Equal(t, 1, deleted)
	assert.Equal(t, 0, deactivated)
	assert.Equal(t, 0, skipped)
}

type photoProvider struct {
	users scimclient.Users
	photo provider.Photo
	err   error
	mu    sync.Mutex
	calls int
}

func (p *photoProvider) Name() string {
	return "photo"
}

func (p *photoProvider) ListUsers(_ context.Context) (scimclient.Users, error) {
	return p.users, nil
}

func (p *photoProvider) UserPhoto(_ context.Context, user scimclient.User) (provider.Photo, error) {
	p.mu.Lock()
	p.calls++
	p.mu.Unlock()

	photo := p.photo
	if photo.ETag == "" {
		photo.ETag = user.PhotoETag
	}

	return photo, p.err
}

func TestBridge_Run_SyncsPhotoWhenETagChanges(t *testing.T) {
	t.Parallel()

	var (
		mu       sync.Mutex
		stored   map[string]any
		requests []string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/Users":
			mu.Lock()
			resources := []any{}
			if stored != nil {
				resources = []any{stored}
			}
			mu.Unlock()

			w.Header().Set("Content-Type", "application/scim+json")
			if err := json.NewEncoder(w).Encode(map[string]any{
				"schemas":      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
				"totalResults": len(resources),
				"startIndex":   1,
				"itemsPerPage": 100,
				"Resources":    resources,
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case r.Method == http.MethodPost && r.URL.Path == "/Users":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			var created map[string]any
			if err := json.Unmarshal(body, &created); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			created["id"] = "profile-1"

			mu.Lock()
			stored = created
			requests = append(requests, string(body))
			mu.Unlock()

			w.WriteHeader(http.StatusCreated)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	directory := &photoProvider{
		users: scimclient.Users{{
			UserName:    "ada@example.com",
			DisplayName: "Ada",
			Active:      true,
			ExternalID:  "ada",
			PhotoETag:   "v1",
		}},
		photo: provider.Photo{Body: []byte("face"), MIME: "image/png"},
	}
	client := scimclient.NewClient(server.Client(), server.URL, "token")
	syncer := scimbridge.NewBridge(directory, client)

	created, updated, _, _, skipped, err := syncer.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 1, created)
	assert.Equal(t, 0, updated)
	assert.Equal(t, 0, skipped)
	assert.Equal(t, 1, directory.calls)

	mu.Lock()
	require.Len(t, requests, 1)
	assert.Contains(t, requests[0], `"photoETag":"v1"`)
	assert.Contains(t, requests[0], "data:image/png;base64,")
	mu.Unlock()

	created, updated, _, _, skipped, err = syncer.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 0, created)
	assert.Equal(t, 0, updated)
	assert.Equal(t, 1, skipped)
	assert.Equal(t, 1, directory.calls)
}
