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

// Package complianceportalsourcemaps serves compliance portal source maps on the
// API server. Faro keeps the customer hostname in the request path and this
// handler ignores that segment.
package complianceportalsourcemaps

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"go.probo.inc/probo/pkg/server/statichandler"
)

const (
	// PathPrefix is the API route Faro fetches after stripping "https://".
	// The route sits on the public API server. The bytes are the same files
	// already served on each compliance portal hostname. Faro and CloudFront
	// both reach this router, so the handler does not try to tell them apart.
	PathPrefix = "/compliance-portal-maps"
)

type handler struct {
	files *statichandler.Server
}

// NewHandler serves compliance portal .js.map files from staticFiles. The
// request path, after PathPrefix is stripped, is /{host}/assets/{name}.js.map.
// host is ignored. Any other shape is a 404.
func NewHandler(staticFiles fs.FS) (http.Handler, error) {
	files, err := statichandler.NewServer(
		staticFiles,
		"dist",
		statichandler.GzipOptions{
			EnableFileTypeCheck: true,
			FileTypes:           []string{".js", ".css", ".html", ".map"},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create compliance portal source map handler: %w", err)
	}

	return &handler{files: files}, nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	assetPath, ok := sourceMapPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	cloned := r.Clone(r.Context())
	cloned.URL.Path = assetPath
	h.files.ServeExisting(w, cloned)
}

// sourceMapPath accepts /{host}/assets/{name}.js.map and returns the embedded
// file path. host is a single segment and is not used to choose the file.
func sourceMapPath(requestPath string) (string, bool) {
	path := strings.TrimPrefix(requestPath, "/")

	host, rest, ok := strings.Cut(path, "/")
	if !ok || !singleSegment(host) {
		return "", false
	}

	directory, name, ok := strings.Cut(rest, "/")
	if !ok || directory != "assets" || !singleSegment(name) {
		return "", false
	}

	if !strings.HasSuffix(name, ".js.map") {
		return "", false
	}

	return "/assets/" + name, true
}

func singleSegment(segment string) bool {
	return segment != "" &&
		segment != "." &&
		segment != ".." &&
		!strings.Contains(segment, "/")
}
