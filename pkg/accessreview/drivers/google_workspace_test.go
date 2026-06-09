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
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleWorkspaceDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/google_workspace", "GOOGLE_WORKSPACE_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("GOOGLE_WORKSPACE_TOKEN")))

	driver := NewGoogleWorkspaceDriver(client)
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, records)

	r := records[0]
	assert.NotEmpty(t, r.Email)
	assert.NotEmpty(t, r.FullName)
	assert.NotEmpty(t, r.ExternalID)
	assert.NotEmpty(t, r.Role)
}

func TestGoogleWorkspaceDriverSuspendedAndArchivedUsers(t *testing.T) {
	t.Parallel()

	client := &http.Client{
		Transport: roundTripFunc(
			func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/admin/directory/v1/users" {
					return googleWorkspaceResponse(http.StatusNotFound, `{"error":"not found"}`), nil
				}

				return googleWorkspaceResponse(
					http.StatusOK,
					`{"kind":"admin#directory#users","users":[
						{"id":"user-1","primaryEmail":"active@example.com","name":{"fullName":"Active User"},"suspended":false,"archived":false},
						{"id":"user-2","primaryEmail":"suspended@example.com","name":{"fullName":"Suspended User"},"suspended":true,"archived":false},
						{"id":"user-3","primaryEmail":"archived@example.com","name":{"fullName":"Archived User"},"suspended":false,"archived":true}
					]}`,
				), nil
			},
		),
	}

	driver := NewGoogleWorkspaceDriver(client)

	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 3)

	assert.Equal(t, "active@example.com", records[0].Email)
	require.NotNil(t, records[0].Active)
	assert.True(t, *records[0].Active)

	assert.Equal(t, "suspended@example.com", records[1].Email)
	require.NotNil(t, records[1].Active)
	assert.False(t, *records[1].Active)

	assert.Equal(t, "archived@example.com", records[2].Email)
	require.NotNil(t, records[2].Active)
	assert.False(t, *records[2].Active)
}

func googleWorkspaceResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}
