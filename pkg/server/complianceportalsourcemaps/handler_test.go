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

package complianceportalsourcemaps_test

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/server/complianceportalsourcemaps"
)

const indexHTML = "compliance portal index"

func TestHandler_ServesSourceMap(t *testing.T) {
	t.Parallel()

	handler := newHandler(t)
	mounted := http.StripPrefix(complianceportalsourcemaps.PathPrefix, handler)

	for _, host := range []string{"acme.trust.example", "trust.acme.com"} {
		t.Run(
			host,
			func(t *testing.T) {
				t.Parallel()

				response := get(
					t,
					mounted,
					complianceportalsourcemaps.PathPrefix+"/"+host+"/assets/app.js.map",
				)

				assert.Equal(t, http.StatusOK, response.Code)
				assert.Equal(t, []byte(`{"version":3}`), response.Body.Bytes())
				assert.Equal(
					t,
					"public, max-age=31536000, immutable",
					response.Header().Get("Cache-Control"),
				)
				assert.NotContains(t, response.Body.String(), indexHTML)
			},
		)
	}
}

func TestHandler_GzipsSourceMap(t *testing.T) {
	t.Parallel()

	mounted := http.StripPrefix(complianceportalsourcemaps.PathPrefix, newHandler(t))
	request := httptest.NewRequest(
		http.MethodGet,
		complianceportalsourcemaps.PathPrefix+"/acme.example/assets/app.js.map",
		nil,
	)
	request.Header.Set("Accept-Encoding", "gzip")

	response := httptest.NewRecorder()
	mounted.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "gzip", response.Header().Get("Content-Encoding"))

	reader, err := gzip.NewReader(response.Body)
	require.NoError(t, err)

	body, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, []byte(`{"version":3}`), body)
}

func TestHandler_NotFound(t *testing.T) {
	t.Parallel()

	mounted := http.StripPrefix(complianceportalsourcemaps.PathPrefix, newHandler(t))

	tests := []struct {
		name string
		path string
	}{
		{
			name: "unknown file",
			path: "/acme.example/assets/missing.js.map",
		},
		{
			name: "extra segment",
			path: "/acme.example/assets/vendor/app.js.map",
		},
		{
			name: "parent segment",
			path: "/acme.example/assets/../index.html",
		},
		{
			name: "parent host",
			path: "/../assets/app.js.map",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				t.Parallel()

				response := get(
					t,
					mounted,
					complianceportalsourcemaps.PathPrefix+test.path,
				)

				assert.Equal(t, http.StatusNotFound, response.Code)
				assert.NotContains(t, response.Body.String(), indexHTML)
			},
		)
	}
}

func newHandler(t *testing.T) http.Handler {
	t.Helper()

	handler, err := complianceportalsourcemaps.NewHandler(
		fstest.MapFS{
			"dist/index.html":        &fstest.MapFile{Data: []byte(indexHTML)},
			"dist/assets/app.js.map": &fstest.MapFile{Data: []byte(`{"version":3}`)},
		},
	)
	require.NoError(t, err)

	return handler
}

func get(t *testing.T, handler http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))

	return response
}
