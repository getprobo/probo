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

package files_v1

import (
	"compress/gzip"
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/brand"
	"go.probo.inc/probo/pkg/securecookie"
)

func testHandler(t *testing.T) *Handler {
	t.Helper()

	staticFiles, err := newStaticFileServer(brand.Assets)
	require.NoError(t, err)

	return &Handler{
		logger:      log.NewLogger(log.WithOutput(io.Discard)),
		staticFiles: staticFiles,
	}
}

func newStaticFileRequest(file string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/static/"+file, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("file", file)

	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func readBrandAsset(t *testing.T, file string) []byte {
	t.Helper()

	content, err := fs.ReadFile(brand.Assets, file)
	require.NoError(t, err)

	return content
}

func etagForContent(content []byte) string {
	hash := md5.Sum(content)

	return `"` + hex.EncodeToString(hash[:]) + `"`
}

func TestHandleGetStaticFile_SetsCachingHeaders(t *testing.T) {
	t.Parallel()

	h := testHandler(t)
	content := readBrandAsset(t, "probo.png")

	rec := httptest.NewRecorder()
	req := newStaticFileRequest("probo.png")

	h.handleGetStaticFile(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "image/png", rec.Header().Get("Content-Type"))
	assert.Equal(t, "public, max-age=31536000, immutable", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "Accept-Encoding", rec.Header().Get("Vary"))
	assert.Equal(t, etagForContent(content), rec.Header().Get("ETag"))
	assert.Empty(t, rec.Header().Get("Content-Encoding"))
	assert.Equal(t, content, rec.Body.Bytes())
}

func TestHandleGetStaticFile_ReturnsNotModifiedForMatchingETag(t *testing.T) {
	t.Parallel()

	t.Run(
		"exact match",
		func(t *testing.T) {
			t.Parallel()

			h := testHandler(t)
			content := readBrandAsset(t, "probo.png")

			rec := httptest.NewRecorder()
			req := newStaticFileRequest("probo.png")
			req.Header.Set("If-None-Match", etagForContent(content))

			h.handleGetStaticFile(rec, req)

			assert.Equal(t, http.StatusNotModified, rec.Code)
			assert.Equal(t, "public, max-age=31536000, immutable", rec.Header().Get("Cache-Control"))
			assert.Equal(t, etagForContent(content), rec.Header().Get("ETag"))
			assert.Empty(t, rec.Body.String())
		},
	)

	t.Run(
		"gzip request with etag list",
		func(t *testing.T) {
			t.Parallel()

			h := testHandler(t)
			content := readBrandAsset(t, "probo.png")

			rec := httptest.NewRecorder()
			req := newStaticFileRequest("probo.png")
			req.Header.Set("Accept-Encoding", "gzip")
			req.Header.Set("If-None-Match", `"other", `+etagForContent(content))

			h.handleGetStaticFile(rec, req)

			assert.Equal(t, http.StatusNotModified, rec.Code)
			assert.Equal(t, etagForContent(content), rec.Header().Get("ETag"))
			assert.Empty(t, rec.Header().Get("Content-Encoding"))
			assert.Empty(t, rec.Body.String())
		},
	)

	t.Run(
		"gzip request with weak etag",
		func(t *testing.T) {
			t.Parallel()

			h := testHandler(t)
			content := readBrandAsset(t, "probo.png")

			rec := httptest.NewRecorder()
			req := newStaticFileRequest("probo.png")
			req.Header.Set("Accept-Encoding", "gzip")
			req.Header.Set("If-None-Match", "W/"+etagForContent(content))

			h.handleGetStaticFile(rec, req)

			assert.Equal(t, http.StatusNotModified, rec.Code)
			assert.Equal(t, etagForContent(content), rec.Header().Get("ETag"))
			assert.Empty(t, rec.Header().Get("Content-Encoding"))
			assert.Empty(t, rec.Body.String())
		},
	)
}

func TestHandleGetStaticFile_CompressesGzipResponses(t *testing.T) {
	t.Parallel()

	h := testHandler(t)
	content := readBrandAsset(t, "probo.png")

	rec := httptest.NewRecorder()
	req := newStaticFileRequest("probo.png")
	req.Header.Set("Accept-Encoding", "br, GZip;q=0.5")

	h.handleGetStaticFile(rec, req)

	gz, err := gzip.NewReader(rec.Body)
	require.NoError(t, err)

	defer func() { _ = gz.Close() }()

	decompressed, err := io.ReadAll(gz)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
	assert.Equal(t, "Accept-Encoding", rec.Header().Get("Vary"))
	assert.Equal(t, etagForContent(content), rec.Header().Get("ETag"))
	assert.Equal(t, content, decompressed)
}

func TestHandleGetStaticFile_DoesNotCompressWhenGzipIsRefused(t *testing.T) {
	t.Parallel()

	h := testHandler(t)
	content := readBrandAsset(t, "probo.png")

	rec := httptest.NewRecorder()
	req := newStaticFileRequest("probo.png")
	req.Header.Set("Accept-Encoding", "br, gzip;q=0, *;q=1")

	h.handleGetStaticFile(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Content-Encoding"))
	assert.Equal(t, content, rec.Body.Bytes())
}

func TestHandleGetStaticFile_NotFound(t *testing.T) {
	t.Parallel()

	h := testHandler(t)

	rec := httptest.NewRecorder()
	req := newStaticFileRequest("missing.png")

	h.handleGetStaticFile(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "file not found")
}

func TestHandleGetPublicFile_InvalidGID(t *testing.T) {
	t.Parallel()

	h := testHandler(t)

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

	h := testHandler(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/not-a-valid-gid", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("fileID", "not-a-valid-gid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.handleGetFile(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandleGetFile_UnauthenticatedReturns401(t *testing.T) {
	t.Parallel()

	// NewMux with nil services — safe because auth middleware returns 401
	// before any service is called when no credentials are present.
	mux, err := NewMux(
		log.NewLogger(log.WithOutput(io.Discard)),
		nil, // fileSvc — not reached
		nil, // proboSvc — not reached
		nil, // iamSvc — not reached when no token/cookie present
		securecookie.Config{},
		"test-secret",
	)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/some-valid-looking-id", nil)
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
