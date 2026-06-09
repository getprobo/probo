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

package googleworkspace_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/iam/scim/bridge/provider/googleworkspace"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestProvider_ListUsers_ActiveSuspendedArchivedAndExcluded(t *testing.T) {
	t.Parallel()

	client := &http.Client{
		Transport: roundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/admin/directory/v1/users" {
					return googleDirectoryResponse(http.StatusNotFound, `{"error":"not found"}`), nil
				}

				return googleDirectoryResponse(
					http.StatusOK,
					`{"kind":"admin#directory#users","users":[
						{"id":"user-1","primaryEmail":"active@example.com","name":{"givenName":"Active","familyName":"User","fullName":"Active User"},"suspended":false,"archived":false},
						{"id":"user-2","primaryEmail":"suspended@example.com","name":{"givenName":"Suspended","familyName":"User","fullName":"Suspended User"},"suspended":true,"archived":false},
						{"id":"user-3","primaryEmail":"archived@example.com","name":{"givenName":"Archived","familyName":"User","fullName":"Archived User"},"suspended":false,"archived":true},
						{"id":"user-4","primaryEmail":"excluded@example.com","name":{"givenName":"Excluded","familyName":"User","fullName":"Excluded User"},"suspended":false,"archived":false}
					]}`,
				), nil
			},
		),
	}

	provider := googleworkspace.New(client, []string{"excluded@example.com"})

	users, err := provider.ListUsers(context.Background())
	require.NoError(t, err)
	require.Len(t, users, 3)

	assert.Equal(t, "active@example.com", users[0].UserName)
	assert.True(t, users[0].Active)

	assert.Equal(t, "suspended@example.com", users[1].UserName)
	assert.False(t, users[1].Active)

	assert.Equal(t, "archived@example.com", users[2].UserName)
	assert.False(t, users[2].Active)
}

func googleDirectoryResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}
