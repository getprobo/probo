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

package complianceportal_v1

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	complianceportalstatics "go.probo.inc/probo/apps/compliance-portal"
)

func TestSkipForStaticFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		path           string
		wantMiddleware bool
	}{
		{name: "hashed asset", path: "/assets/index-abc.js", wantMiddleware: false},
		{name: "favicon", path: "/favicons/favicon.ico", wantMiddleware: false},
		{name: "boot script", path: "/scripts/fouc-dark.js", wantMiddleware: false},
		{name: "portal page", path: "/en/documents", wantMiddleware: true},
		{name: "graphql", path: "/graphql", wantMiddleware: true},
		{name: "prefix without separator", path: "/assetsx/app.js", wantMiddleware: true},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				middlewareCalled := false
				middleware := func(next http.Handler) http.Handler {
					return http.HandlerFunc(
						func(w http.ResponseWriter, r *http.Request) {
							middlewareCalled = true
							next.ServeHTTP(w, r)
						},
					)
				}

				nextCalled := false
				next := http.HandlerFunc(
					func(_ http.ResponseWriter, _ *http.Request) {
						nextCalled = true
					},
				)

				handler := skipForStaticFiles(middleware)(next)
				handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, tt.path, nil))

				assert.True(t, nextCalled)
				assert.Equal(t, tt.wantMiddleware, middlewareCalled)
			},
		)
	}
}

func TestStaticFilesHandler(t *testing.T) {
	t.Parallel()

	webServer, err := NewServer(
		func(_ *http.Request) HeadData { return HeadData{} },
		"",
		false,
		"",
		"",
	)
	require.NoError(t, err)

	entries, err := fs.ReadDir(complianceportalstatics.StaticFiles, "dist/assets")
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	assetPath := "/assets/" + entries[0].Name()
	handler := staticFilesHandler(webServer)

	t.Run(
		"serves a build file",
		func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, assetPath, nil))

			assert.Equal(t, http.StatusOK, w.Code)
			assert.NotContains(t, w.Header().Get("Content-Type"), "text/html")
		},
	)

	t.Run(
		"returns not found for a missing file instead of the spa shell",
		func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/assets/missing-file.js", nil))

			assert.Equal(t, http.StatusNotFound, w.Code)
		},
	)
}
