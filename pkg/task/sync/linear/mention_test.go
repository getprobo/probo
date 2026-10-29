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

package linear

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileSlug(t *testing.T) {
	t.Parallel()

	t.Run("user profile", func(t *testing.T) {
		t.Parallel()

		slug, ok := ProfileSlug("https://linear.app/acme/profiles/jane")
		require.True(t, ok)
		assert.Equal(t, "jane", slug)
	})

	t.Run("encoded display name", func(t *testing.T) {
		t.Parallel()

		slug, ok := ProfileSlug("https://linear.app/acme/profiles/Jane%20Doe")
		require.True(t, ok)
		assert.Equal(t, "Jane Doe", slug)
	})

	t.Run("issue url", func(t *testing.T) {
		t.Parallel()

		_, ok := ProfileSlug("https://linear.app/acme/issue/ENG-1/title")
		assert.False(t, ok)
	})

	t.Run("other host", func(t *testing.T) {
		t.Parallel()

		_, ok := ProfileSlug("https://example.com/acme/profiles/jane")
		assert.False(t, ok)
	})
}

func TestClient_MentionUserByEmail(t *testing.T) {
	t.Parallel()

	var handlerErr error

	server := newLinearGraphQLServer(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			req, ok := serveDecodedGraphQL(w, r, &handlerErr)
			if !ok {
				return
			}

			vars, varErr := graphQLObjectVariables(req)
			if varErr != nil {
				handlerErr = varErr
				http.Error(w, varErr.Error(), http.StatusBadRequest)

				return
			}

			assert.Contains(t, req.Query, "TaskSyncLinearMentionUserByEmail")
			assert.Equal(t, "jane@example.com", vars["email"])

			_, _ = w.Write([]byte(`{
				"data": {
					"users": {
						"nodes": [{
							"id": "user-1",
							"email": "jane@example.com",
							"name": "Jane Doe",
							"displayName": "jane",
							"url": "https://linear.app/acme/profiles/jane",
							"active": true,
							"isMentionable": true
						}]
					}
				}
			}`))
		},
	)

	user, err := NewClient(server.Client(), server.URL).MentionUserByEmail(
		context.Background(),
		"jane@example.com",
	)
	require.NoError(t, err)
	require.NoError(t, handlerErr)
	require.NotNil(t, user)
	assert.Equal(t, "user-1", user.ID)
	assert.Equal(t, "https://linear.app/acme/profiles/jane", user.URL)
	assert.True(t, user.Mentionable)
}

func TestClient_MentionUserForProfileURL(t *testing.T) {
	t.Parallel()

	t.Run("matches the profile url", func(t *testing.T) {
		t.Parallel()

		user := mentionUserFromProfileQuery(
			t,
			`{
				"data": {
					"byDisplayName": {"nodes": [{
						"id": "user-1",
						"email": "jane@example.com",
						"name": "Jane Doe",
						"displayName": "jane",
						"url": "https://linear.app/acme/profiles/jane",
						"active": true,
						"isMentionable": true
					}]},
					"byName": {"nodes": []}
				}
			}`,
		)
		require.NotNil(t, user)
		assert.Equal(t, "jane@example.com", user.Email)
	})

	t.Run("ambiguous names in the workspace are skipped", func(t *testing.T) {
		t.Parallel()

		user := mentionUserForURL(
			t,
			"https://linear.app/acme/profiles/Jane%20Doe",
			`{
				"data": {
					"byDisplayName": {"nodes": []},
					"byName": {"nodes": [
						{
							"id": "user-1",
							"email": "jane@example.com",
							"name": "Jane Doe",
							"displayName": "jane",
							"url": "https://linear.app/acme/profiles/jane",
							"active": true,
							"isMentionable": true
						},
						{
							"id": "user-2",
							"email": "other@example.com",
							"name": "Jane Doe",
							"displayName": "jane2",
							"url": "https://linear.app/acme/profiles/jane2",
							"active": true,
							"isMentionable": true
						}
					]}
				}
			}`,
		)
		assert.Nil(t, user)
	})

	t.Run("rewritten slug stays in the same workspace", func(t *testing.T) {
		t.Parallel()

		user := mentionUserForURL(
			t,
			"https://linear.app/acme/profiles/Jane%20Doe",
			`{
				"data": {
					"byDisplayName": {"nodes": []},
					"byName": {"nodes": [{
						"id": "user-1",
						"email": "jane@example.com",
						"name": "Jane Doe",
						"displayName": "jane",
						"url": "https://linear.app/acme/profiles/jane",
						"active": true,
						"isMentionable": true
					}]}
				}
			}`,
		)
		require.NotNil(t, user)
		assert.Equal(t, "jane@example.com", user.Email)
	})

	t.Run("a deactivated namesake makes the match ambiguous", func(t *testing.T) {
		t.Parallel()

		user := mentionUserForURL(
			t,
			"https://linear.app/acme/profiles/Jane%20Doe",
			`{
				"data": {
					"byDisplayName": {"nodes": []},
					"byName": {"nodes": [
						{
							"id": "user-1",
							"email": "jane@example.com",
							"name": "Jane Doe",
							"displayName": "jane",
							"url": "https://linear.app/acme/profiles/jane",
							"active": false,
							"isMentionable": true
						},
						{
							"id": "user-2",
							"email": "other@example.com",
							"name": "Jane Doe",
							"displayName": "jane2",
							"url": "https://linear.app/acme/profiles/jane2",
							"active": true,
							"isMentionable": true
						}
					]}
				}
			}`,
		)
		assert.Nil(t, user)
	})

	t.Run("other workspace is skipped", func(t *testing.T) {
		t.Parallel()

		user := mentionUserForURL(
			t,
			"https://linear.app/other/profiles/jane",
			`{
				"data": {
					"byDisplayName": {"nodes": [{
						"id": "user-1",
						"email": "jane@example.com",
						"name": "Jane Doe",
						"displayName": "jane",
						"url": "https://linear.app/acme/profiles/jane",
						"active": true,
						"isMentionable": true
					}]},
					"byName": {"nodes": []}
				}
			}`,
		)
		assert.Nil(t, user)
	})
}

func TestClient_CommentBody(t *testing.T) {
	t.Parallel()

	var handlerErr error

	server := newLinearGraphQLServer(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			req, ok := serveDecodedGraphQL(w, r, &handlerErr)
			if !ok {
				return
			}

			vars, varErr := graphQLObjectVariables(req)
			if varErr != nil {
				handlerErr = varErr
				http.Error(w, varErr.Error(), http.StatusBadRequest)

				return
			}

			assert.Equal(t, "comment-1", vars["id"])

			_, _ = w.Write([]byte(`{"data":{"comment":{"id":"comment-1","body":"Hi"}}}`))
		},
	)

	body, err := NewClient(server.Client(), server.URL).CommentBody(context.Background(), "comment-1")
	require.NoError(t, err)
	require.NoError(t, handlerErr)
	assert.Equal(t, "Hi", body)
}

func mentionUserFromProfileQuery(t *testing.T, payload string) *MentionUser {
	t.Helper()

	return mentionUserForURL(t, "https://linear.app/acme/profiles/jane", payload)
}

func mentionUserForURL(t *testing.T, profileURL string, payload string) *MentionUser {
	t.Helper()

	var handlerErr error

	server := newLinearGraphQLServer(
		t,
		func(w http.ResponseWriter, r *http.Request) {
			req, ok := serveDecodedGraphQL(w, r, &handlerErr)
			if !ok {
				return
			}

			vars, varErr := graphQLObjectVariables(req)
			if varErr != nil {
				handlerErr = varErr
				http.Error(w, varErr.Error(), http.StatusBadRequest)

				return
			}

			slug, slugOK := ProfileSlug(profileURL)
			if !slugOK {
				handlerErr = fmt.Errorf("cannot read profile url")

				http.Error(w, "cannot read profile url", http.StatusBadRequest)

				return
			}

			assert.Contains(t, req.Query, "TaskSyncLinearMentionUserByProfile")
			assert.Equal(t, slug, vars["displayName"])
			assert.Equal(t, slug, vars["name"])

			_, _ = w.Write([]byte(payload))
		},
	)

	user, err := NewClient(server.Client(), server.URL).MentionUserForProfileURL(
		context.Background(),
		profileURL,
	)
	require.NoError(t, err)
	require.NoError(t, handlerErr)

	return user
}
