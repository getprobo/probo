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

package server

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/server/complianceportalsourcemaps"
)

func TestSetupRoutes_ServesCompliancePortalSourceMap(t *testing.T) {
	t.Parallel()

	maps, err := complianceportalsourcemaps.NewHandler(
		fstest.MapFS{
			"dist/index.html":        &fstest.MapFile{Data: []byte("index")},
			"dist/assets/app.js.map": &fstest.MapFile{Data: []byte(`{"version":3}`)},
		},
	)
	require.NoError(t, err)

	srv := &Server{
		router:                     chi.NewRouter(),
		compliancePortalSourceMaps: maps,
	}
	srv.setupRoutes()

	request := httptest.NewRequest(
		http.MethodGet,
		"/compliance-portal-maps/acme.example/assets/app.js.map",
		nil,
	)
	request.Header.Set("Accept-Encoding", "gzip")

	response := httptest.NewRecorder()
	srv.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "gzip", response.Header().Get("Content-Encoding"))
	assert.Equal(
		t,
		"public, max-age=31536000, immutable",
		response.Header().Get("Cache-Control"),
	)

	reader, err := gzip.NewReader(response.Body)
	require.NoError(t, err)

	body, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, []byte(`{"version":3}`), body)
}
