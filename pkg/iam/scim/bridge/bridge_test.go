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

package bridge_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	scimbridge "go.probo.inc/probo/pkg/iam/scim/bridge"
	scimclient "go.probo.inc/probo/pkg/iam/scim/bridge/client"
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
