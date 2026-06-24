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

package files_v1

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/securecookie"
)

func testHandler() *Handler {
	return &Handler{
		logger: log.NewLogger(log.WithOutput(io.Discard)),
	}
}

func TestHandleGetPublicFile_InvalidGID(t *testing.T) {
	t.Parallel()

	h := testHandler()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/public/not-a-valid-gid", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("fileID", "not-a-valid-gid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.handleGetPublicFile(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandleGetFile_InvalidGID(t *testing.T) {
	t.Parallel()

	h := testHandler()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/not-a-valid-gid", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("fileID", "not-a-valid-gid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.handleGetFile(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandleGetStaticFile(t *testing.T) {
	t.Parallel()

	mux := NewMux(
		log.NewLogger(log.WithOutput(io.Discard)),
		nil,
		nil,
		nil,
		securecookie.Config{},
		"test-secret",
		baseurl.MustParse("https://example.com"),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/probo.png", nil)
	mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Header().Get("Cache-Control"), "max-age=3600")

	etag := rec.Header().Get("ETag")
	require.NotEmpty(t, etag)
	require.True(t, strings.HasPrefix(etag, `"`) && strings.HasSuffix(etag, `"`))

	recNotModified := httptest.NewRecorder()
	reqNotModified := httptest.NewRequest(http.MethodGet, "/static/probo.png", nil)
	reqNotModified.Header.Set("If-None-Match", etag)
	mux.ServeHTTP(recNotModified, reqNotModified)

	require.Equal(t, http.StatusNotModified, recNotModified.Code)

	recMissing := httptest.NewRecorder()
	reqMissing := httptest.NewRequest(http.MethodGet, "/static/does-not-exist.png", nil)
	mux.ServeHTTP(recMissing, reqMissing)

	require.Equal(t, http.StatusNotFound, recMissing.Code)
}

func TestHandleGetFile_UnauthenticatedReturns401(t *testing.T) {
	t.Parallel()

	// NewMux with nil services — safe because auth middleware returns 401
	// before any service is called when no credentials are present.
	mux := NewMux(
		log.NewLogger(log.WithOutput(io.Discard)),
		nil, // fileSvc — not reached
		nil, // proboSvc — not reached
		nil, // iamSvc — not reached when no token/cookie present
		securecookie.Config{},
		"test-secret",
		baseurl.MustParse("https://example.com"),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/some-valid-looking-id", nil)
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
