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
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
)

const testTokenType = "probo_cloud_api_token_us"

type (
	staticKeyProvider struct {
		key *ecdsa.PublicKey
		err error
	}

	recordingRevoker struct {
		tokenType string
		tokens    []string
		err       error
	}
)

func (p *staticKeyProvider) PublicKey(context.Context, string) (*ecdsa.PublicKey, error) {
	return p.key, p.err
}

func (r *recordingRevoker) SecretScanningTokenType() string {
	return r.tokenType
}

func (r *recordingRevoker) RevokeLeakedManualAccessToken(_ context.Context, token string) (bool, error) {
	r.tokens = append(r.tokens, token)

	return r.err == nil, r.err
}

func TestHandler_AcceptsSignedAlerts(t *testing.T) {
	t.Parallel()

	const body = `[{"token":"prb_a1u_example","type":"probo_cloud_api_token_us","url":"https://github.com/example/repo","source":"content"}]`

	privateKey := newTestPrivateKey(t)
	revoker := &recordingRevoker{tokenType: testTokenType}
	response := serveAlert(t, body, privateKey, revoker)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, []string{"prb_a1u_example"}, revoker.tokens)
	assert.JSONEq(t, `{"accepted":true}`, response.Body.String())
}

func TestHandler_IgnoresOtherTokenTypes(t *testing.T) {
	t.Parallel()

	const body = `[{"token":"other-secret","type":"another_provider_token","url":"","source":"content"}]`

	privateKey := newTestPrivateKey(t)
	revoker := &recordingRevoker{tokenType: testTokenType}
	response := serveAlert(t, body, privateKey, revoker)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Empty(t, revoker.tokens)
}

func TestHandler_RejectsMissingSignature(t *testing.T) {
	t.Parallel()

	revoker := &recordingRevoker{tokenType: testTokenType}
	handler := NewMux(log.NewLogger(), revoker, &staticKeyProvider{})
	req := httptest.NewRequest(http.MethodPost, "/alerts", strings.NewReader(`[]`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, req)

	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Empty(t, revoker.tokens)
}

func TestHandler_RejectsInvalidSignature(t *testing.T) {
	t.Parallel()

	privateKey := newTestPrivateKey(t)
	revoker := &recordingRevoker{tokenType: testTokenType}
	req := signedRequest(t, `[{"token":"secret","type":"`+testTokenType+`"}]`, privateKey)
	req.Header.Set(publicKeySignatureHeader, base64.StdEncoding.EncodeToString([]byte("invalid")))
	response := httptest.NewRecorder()

	NewMux(
		log.NewLogger(),
		revoker,
		&staticKeyProvider{key: &privateKey.PublicKey},
	).ServeHTTP(response, req)

	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Empty(t, revoker.tokens)
}

func TestHandler_RejectsMalformedSignedPayload(t *testing.T) {
	t.Parallel()

	privateKey := newTestPrivateKey(t)
	revoker := &recordingRevoker{tokenType: testTokenType}
	response := serveAlert(t, `{`, privateKey, revoker)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Empty(t, revoker.tokens)
}

func TestHandler_RejectsOversizedPayload(t *testing.T) {
	t.Parallel()

	revoker := &recordingRevoker{tokenType: testTokenType}
	handler := NewMux(log.NewLogger(), revoker, &staticKeyProvider{})
	req := httptest.NewRequest(
		http.MethodPost,
		"/alerts",
		strings.NewReader(strings.Repeat("a", maxAlertBodyBytes+1)),
	)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, response.Code)
	assert.Empty(t, revoker.tokens)
}

func TestHandler_ReturnsUnavailableWhenKeysCannotBeLoaded(t *testing.T) {
	t.Parallel()

	revoker := &recordingRevoker{tokenType: testTokenType}
	handler := NewMux(
		log.NewLogger(),
		revoker,
		&staticKeyProvider{err: errors.New("unavailable")},
	)
	req := httptest.NewRequest(http.MethodPost, "/alerts", strings.NewReader(`[]`))
	req.Header.Set(publicKeyIdentifierHeader, "test-key")
	req.Header.Set(publicKeySignatureHeader, "signature")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, req)

	assert.Equal(t, http.StatusServiceUnavailable, response.Code)
	assert.Empty(t, revoker.tokens)
}

func TestHandler_ReturnsErrorWhenRevocationFails(t *testing.T) {
	t.Parallel()

	const body = `[{"token":"prb_a1u_example","type":"probo_cloud_api_token_us","url":"","source":"content"}]`

	privateKey := newTestPrivateKey(t)
	revoker := &recordingRevoker{
		tokenType: testTokenType,
		err:       errors.New("database unavailable"),
	}
	response := serveAlert(t, body, privateKey, revoker)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, []string{"prb_a1u_example"}, revoker.tokens)
}

func serveAlert(
	t *testing.T,
	body string,
	privateKey *ecdsa.PrivateKey,
	revoker TokenRevoker,
) *httptest.ResponseRecorder {
	t.Helper()

	req := signedRequest(t, body, privateKey)
	response := httptest.NewRecorder()
	handler := NewMux(
		log.NewLogger(),
		revoker,
		&staticKeyProvider{key: &privateKey.PublicKey},
	)
	handler.ServeHTTP(response, req)

	return response
}

func signedRequest(t *testing.T, body string, privateKey *ecdsa.PrivateKey) *http.Request {
	t.Helper()

	digest := sha256.Sum256([]byte(body))
	signature, err := ecdsa.SignASN1(cryptorand.Reader, privateKey, digest[:])
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/alerts", strings.NewReader(body))
	req.Header.Set(publicKeyIdentifierHeader, "test-key")
	req.Header.Set(publicKeySignatureHeader, base64.StdEncoding.EncodeToString(signature))

	return req
}

func newTestPrivateKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), cryptorand.Reader)
	require.NoError(t, err)

	return privateKey
}
