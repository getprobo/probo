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

package connector

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingRoundTripper captures the last request it sees and returns a
// canned 200 response without touching the network, so transport
// behaviour can be asserted without tripping SSRF protection.
type recordingRoundTripper struct {
	lastRequest *http.Request
}

func (rt *recordingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.lastRequest = req

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       http.NoBody,
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func TestBasicAuthTransport_RoundTrip(t *testing.T) {
	t.Parallel()

	rec := &recordingRoundTripper{}
	transport := &basicAuthTransport{username: "key_secret", underlying: rec}

	req := httptest.NewRequest(http.MethodGet, "https://api.cursor.com/teams/members", nil)
	_, err := transport.RoundTrip(req)
	require.NoError(t, err)

	require.NotNil(t, rec.lastRequest)
	user, pass, ok := rec.lastRequest.BasicAuth()
	require.True(t, ok, "expected a Basic auth header")
	assert.Equal(t, "key_secret", user)
	assert.Empty(t, pass, "password must be empty")

	// The original request must be left untouched (RoundTrip clones it).
	_, _, originalHasAuth := req.BasicAuth()
	assert.False(t, originalHasAuth, "original request must not be mutated")
}

func TestAPIKeyConnection_Client_BasicAuth(t *testing.T) {
	t.Parallel()

	conn := &APIKeyConnection{APIKey: "key_secret", BasicAuth: true}

	client, err := conn.Client(context.Background())
	require.NoError(t, err)

	transport, ok := client.Transport.(*basicAuthTransport)
	require.Truef(t, ok, "expected *basicAuthTransport, got %T", client.Transport)
	assert.Equal(t, "key_secret", transport.username)
}

func TestAPIKeyConnection_Client_Header(t *testing.T) {
	t.Parallel()

	conn := &APIKeyConnection{APIKey: "sk-ant-admin", Header: "x-api-key"}

	client, err := conn.Client(context.Background())
	require.NoError(t, err)

	transport, ok := client.Transport.(*apiKeyHeaderTransport)
	require.Truef(t, ok, "expected *apiKeyHeaderTransport, got %T", client.Transport)
	assert.Equal(t, "x-api-key", transport.header)
	assert.Equal(t, "sk-ant-admin", transport.value)
}

func TestAPIKeyConnection_Client_BearerDefault(t *testing.T) {
	t.Parallel()

	conn := &APIKeyConnection{APIKey: "token"}

	client, err := conn.Client(context.Background())
	require.NoError(t, err)

	transport, ok := client.Transport.(*oauth2Transport)
	require.Truef(t, ok, "expected *oauth2Transport, got %T", client.Transport)
	assert.Equal(t, "token", transport.token)
}
