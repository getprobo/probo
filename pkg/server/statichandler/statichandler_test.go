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
