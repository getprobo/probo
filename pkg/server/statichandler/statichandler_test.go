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

package statichandler_test

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/server/statichandler"
)

type requestPage struct {
	Title string
	statichandler.IndexData
}

func TestNewIndexRenderer(t *testing.T) {
	t.Parallel()

	staticFiles := indexFiles(
		`{{if .FaroEnabled}}<meta content="{{.FaroCollectorURL}}">{{end}}`,
	)

	t.Run(
		"renders escaped faro values",
		func(t *testing.T) {
			t.Parallel()

			renderer, err := statichandler.NewIndexRenderer(
				staticFiles,
				"dist",
				statichandler.IndexData{
					FaroEnabled:      true,
					FaroCollectorURL: `https://example.com/c?a=1&b="x"`,
					FaroAPIKey:       "key",
				},
			)
			require.NoError(t, err)

			body := renderIndex(t, renderer, httptest.NewRequest(http.MethodGet, "/", nil))
			assert.Equal(
				t,
				`<meta content="https://example.com/c?a=1&amp;b=&#34;x&#34;">`,
				body,
			)
		},
	)

	t.Run(
		"omits faro tags when disabled",
		func(t *testing.T) {
			t.Parallel()

			renderer, err := statichandler.NewIndexRenderer(
				staticFiles,
				"dist",
				statichandler.IndexData{
					FaroEnabled:      false,
					FaroCollectorURL: "https://example.com/c",
					FaroAPIKey:       "key",
				},
			)
			require.NoError(t, err)

			body := renderIndex(t, renderer, httptest.NewRequest(http.MethodGet, "/", nil))
			assert.Empty(t, body)
		},
	)

	t.Run(
		"missing index.html",
		func(t *testing.T) {
			t.Parallel()

			_, err := statichandler.NewIndexRenderer(
				fstest.MapFS{
					"dist/app.js": &fstest.MapFile{Data: []byte("console.log(1)")},
				},
				"dist",
				statichandler.IndexData{},
			)
			require.Error(t, err)
			assert.ErrorContains(t, err, "cannot read index.html")
		},
	)
}

func TestNewRequestIndexRenderer_RendersPerRequest(t *testing.T) {
	t.Parallel()

	renderer, err := statichandler.NewRequestIndexRenderer(
		indexFiles(
			`<title>{{.Title}}</title>{{if .FaroEnabled}}<meta content="{{.FaroCollectorURL}}">{{end}}`,
		),
		"dist",
		func(r *http.Request) any {
			if r.URL.Query().Get("page") == "a" {
				return requestPage{
					Title:            "Alpha",
					FaroEnabled:      true,
					FaroCollectorURL: `https://a.example/?q=1&x="y"`,
				}
			}

			return requestPage{Title: "Beta"}
		},
	)
	require.NoError(t, err)

	alpha := renderIndex(
		t,
		renderer,
		httptest.NewRequest(http.MethodGet, "/?page=a", nil),
	)
	beta := renderIndex(
		t,
		renderer,
		httptest.NewRequest(http.MethodGet, "/?page=b", nil),
	)

	assert.Equal(
		t,
		`<title>Alpha</title><meta content="https://a.example/?q=1&amp;x=&#34;y&#34;">`,
		alpha,
	)
	assert.Equal(t, "<title>Beta</title>", beta)
}

func TestServer_ServeHTTP(t *testing.T) {
	t.Parallel()

	script := bytes.Repeat([]byte("console.log('probo');\n"), 200)
	incompressible := make([]byte, 4096)
	_, err := rand.Read(incompressible)
	require.NoError(t, err)

	staticFiles := fstest.MapFS{
		"dist/index.html":        &fstest.MapFile{Data: []byte("<html></html>")},
		"dist/assets/app.js":     &fstest.MapFile{Data: script},
		"dist/assets/worker.mjs": &fstest.MapFile{Data: script},
		"dist/assets/logo.png":   &fstest.MapFile{Data: incompressible},
	}

	newServer := func(t *testing.T) *statichandler.Server {
		t.Helper()

		server, err := statichandler.NewServer(
			staticFiles,
			"dist",
			statichandler.GzipOptions{
				EnableFileTypeCheck: true,
				FileTypes:           []string{".js", ".mjs", ".png"},
			},
		)
		require.NoError(t, err)

		return server
	}

	serve := func(t *testing.T, server *statichandler.Server, r *http.Request) *httptest.ResponseRecorder {
		t.Helper()

		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)

		return w
	}

	gzipRequest := func(target string) *http.Request {
		r := httptest.NewRequest(http.MethodGet, target, nil)
		r.Header.Set("Accept-Encoding", "gzip, br")

		return r
	}

	t.Run(
		"serves precompressed assets",
		func(t *testing.T) {
			t.Parallel()

			w := serve(t, newServer(t), gzipRequest("/assets/app.js"))

			require.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
			assert.Equal(t, "Accept-Encoding", w.Header().Get("Vary"))
			assert.Equal(t, "public, max-age=31536000, immutable", w.Header().Get("Cache-Control"))
			assert.Contains(t, w.Header().Get("Content-Type"), "javascript")
			assert.Contains(t, w.Header().Get("ETag"), "-gzip")

			gz, err := gzip.NewReader(w.Body)
			require.NoError(t, err)

			body, err := io.ReadAll(gz)
			require.NoError(t, err)
			assert.Equal(t, script, body)
		},
	)

	t.Run(
		"returns not modified for a matching gzip etag",
		func(t *testing.T) {
			t.Parallel()

			server := newServer(t)
			first := serve(t, server, gzipRequest("/assets/app.js"))

			r := gzipRequest("/assets/app.js")
			r.Header.Set("If-None-Match", first.Header().Get("ETag"))
			w := serve(t, server, r)

			assert.Equal(t, http.StatusNotModified, w.Code)
			assert.Empty(t, w.Body.Bytes())
		},
	)

	t.Run(
		"caches module scripts as immutable",
		func(t *testing.T) {
			t.Parallel()

			w := serve(t, newServer(t), gzipRequest("/assets/worker.mjs"))

			require.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
			assert.Equal(t, "public, max-age=31536000, immutable", w.Header().Get("Cache-Control"))
		},
	)

	t.Run(
		"serves identity encoding without gzip support",
		func(t *testing.T) {
			t.Parallel()

			w := serve(t, newServer(t), httptest.NewRequest(http.MethodGet, "/assets/app.js", nil))

			require.Equal(t, http.StatusOK, w.Code)
			assert.Empty(t, w.Header().Get("Content-Encoding"))
			assert.NotContains(t, w.Header().Get("ETag"), "-gzip")
			assert.Equal(t, script, w.Body.Bytes())
		},
	)

	t.Run(
		"serves byte ranges from the identity encoding",
		func(t *testing.T) {
			t.Parallel()

			r := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
			r.Header.Set("Range", "bytes=0-6")
			w := serve(t, newServer(t), r)

			require.Equal(t, http.StatusPartialContent, w.Code)
			assert.Equal(t, script[:7], w.Body.Bytes())
		},
	)

	t.Run(
		"skips precompression when gzip does not shrink the file",
		func(t *testing.T) {
			t.Parallel()

			w := serve(t, newServer(t), gzipRequest("/assets/logo.png"))

			require.Equal(t, http.StatusOK, w.Code)
			assert.NotContains(t, w.Header().Get("ETag"), "-gzip")
		},
	)

	t.Run(
		"omits body for head requests",
		func(t *testing.T) {
			t.Parallel()

			r := gzipRequest("/assets/app.js")
			r.Method = http.MethodHead
			w := serve(t, newServer(t), r)

			require.Equal(t, http.StatusOK, w.Code)
			assert.NotEmpty(t, w.Header().Get("Content-Length"))
			assert.Empty(t, w.Body.Bytes())
		},
	)
}

func indexFiles(indexHTML string) fs.FS {
	return fstest.MapFS{
		"dist/index.html": &fstest.MapFile{Data: []byte(indexHTML)},
	}
}

func renderIndex(t *testing.T, renderer statichandler.FileRenderer, r *http.Request) string {
	t.Helper()

	var buf bytes.Buffer

	err := renderer(&buf, r)
	require.NoError(t, err)

	return buf.String()
}
