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
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.gearno.de/kit/log"
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
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/some-valid-looking-id", nil)
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
