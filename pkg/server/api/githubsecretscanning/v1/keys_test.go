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

package githubsecretscanning_v1

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingHTTPClient struct {
	body  []byte
	calls int
}

func (c *recordingHTTPClient) Do(req *http.Request) (*http.Response, error) {
	c.calls++

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(c.body)),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func TestGitHubKeyProvider_CachesPublicKeys(t *testing.T) {
	t.Parallel()

	privateKey := newTestPrivateKey(t)
	body := publicKeysBody(t, "test-key", &privateKey.PublicKey)
	client := &recordingHTTPClient{body: body}
	provider := NewGitHubKeyProvider(client)

	first, err := provider.PublicKey(context.Background(), "test-key")
	require.NoError(t, err)
	second, err := provider.PublicKey(context.Background(), "test-key")
	require.NoError(t, err)

	assert.Equal(t, &privateKey.PublicKey, first)
	assert.Equal(t, first, second)
	assert.Equal(t, 1, client.calls)
}

func TestGitHubKeyProvider_RejectsUnknownKey(t *testing.T) {
	t.Parallel()

	privateKey := newTestPrivateKey(t)
	client := &recordingHTTPClient{
		body: publicKeysBody(t, "another-key", &privateKey.PublicKey),
	}
	provider := NewGitHubKeyProvider(client)

	_, err := provider.PublicKey(context.Background(), "missing-key")

	assert.ErrorIs(t, err, ErrPublicKeyNotFound)
	assert.Equal(t, 1, client.calls)
}

func publicKeysBody(t *testing.T, identifier string, key *ecdsa.PublicKey) []byte {
	t.Helper()

	der, err := x509.MarshalPKIXPublicKey(key)
	require.NoError(t, err)

	body, err := json.Marshal(
		publicKeyResponse{
			PublicKeys: []publicKeyEntry{
				{
					Identifier: identifier,
					Key: string(
						pem.EncodeToMemory(
							&pem.Block{
								Type:  "PUBLIC KEY",
								Bytes: der,
							},
						),
					),
				},
			},
		},
	)
	require.NoError(t, err)

	return body
}
