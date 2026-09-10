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

package azure

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubCredential struct {
	err   error
	errN  int
	calls int
}

func (c *stubCredential) GetToken(
	_ context.Context,
	_ policy.TokenRequestOptions,
) (azcore.AccessToken, error) {
	c.calls++
	if c.calls <= c.errN {
		return azcore.AccessToken{}, c.err
	}

	return azcore.AccessToken{
		Token:     "arm-token",
		ExpiresOn: time.Now().Add(time.Hour),
	}, nil
}

func entraAuthError(codes ...int) error {
	body, err := json.Marshal(map[string]any{"error_codes": codes})
	if err != nil {
		panic(err)
	}

	return &azidentity.AuthenticationFailedError{
		RawResponse: &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(bytes.NewReader(body)),
		},
	}
}

func afterImmediately(time.Duration) <-chan time.Time {
	ch := make(chan time.Time)
	close(ch)

	return ch
}

func TestCheckAccess_RetriesFederatedCredentialNotFound(t *testing.T) {
	t.Parallel()

	cred := &stubCredential{err: entraAuthError(entraFederatedCredentialNotFound), errN: 2}
	session := &Session{
		credential: cred,
		armScope:   "https://management.azure.com/.default",
		after:      afterImmediately,
	}

	err := session.CheckAccess(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 3, cred.calls)
}

func TestCheckAccess_GivesUpAfterFederatedCredentialRetries(t *testing.T) {
	t.Parallel()

	cred := &stubCredential{err: entraAuthError(entraFederatedCredentialNotFound), errN: 10}
	session := &Session{
		credential: cred,
		armScope:   "https://management.azure.com/.default",
		after:      afterImmediately,
	}

	err := session.CheckAccess(context.Background())
	require.Error(t, err)
	assert.Equal(t, 3, cred.calls)
}

func TestCheckAccess_DoesNotRetryOtherEntraError(t *testing.T) {
	t.Parallel()

	cred := &stubCredential{err: entraAuthError(700016), errN: 5}
	session := &Session{
		credential: cred,
		armScope:   "https://management.azure.com/.default",
		after:      afterImmediately,
	}

	err := session.CheckAccess(context.Background())
	require.Error(t, err)
	assert.Equal(t, 1, cred.calls)
}

func TestCheckAccess_ReturnsImmediatelyWhenCanceledDuringBackoff(t *testing.T) {
	t.Parallel()

	cred := &stubCredential{err: entraAuthError(entraFederatedCredentialNotFound), errN: 10}
	entered := make(chan struct{})
	session := &Session{
		credential: cred,
		armScope:   "https://management.azure.com/.default",
		after: func(time.Duration) <-chan time.Time {
			close(entered)

			return make(chan time.Time)
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- session.CheckAccess(ctx)
	}()

	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("CheckAccess did not enter backoff")
	}

	cancel()

	select {
	case err := <-done:
		require.ErrorIs(t, err, context.Canceled)
		assert.Equal(t, 1, cred.calls)
	case <-time.After(time.Second):
		t.Fatal("CheckAccess blocked after cancel")
	}
}
