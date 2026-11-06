// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

// Package statichandler provides functionality for serving SPA (Single Page Application) frontends.
package statichandler

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
)

// minGzipRatio is the largest compressed/original size ratio worth serving:
// already-compressed formats (png, woff2) barely shrink and are served as is.
const minGzipRatio = 0.9

type (
	// FileRenderer renders dynamic content for a given file path. When
	// registered via WithFileRenderer, the server calls it instead of serving
	// the static embedded bytes. The renderer writes the response body to w.
	FileRenderer func(w io.Writer, r *http.Request) error

	Option func(*Server)

	// IndexData is the template data for the Faro collector meta tags in a
	// portal index.html.
	IndexData struct {
		FaroEnabled      bool
		FaroCollectorURL string
		FaroAPIKey       string
	}

	GzipOptions struct {
		EnableFileTypeCheck bool
		FileTypes           []string
	}

	Server struct {
		spaFS         http.FileSystem
		etags         map[string]string
		indexETag     string
		indexContent  []byte
		gzipOptions   GzipOptions
		fileRenderers map[string]FileRenderer
		gzipped       map[string]gzippedFile
	}

	// gzippedFile is a static file compressed once at startup so requests do
	// not pay for gzip on every hit.
	gzippedFile struct {
		content     []byte
		etag        string
		contentType string
	}
)

// WithFileRenderer registers a dynamic renderer for the given path (e.g.
// "/index.html"). When the server would serve that file, it calls the
// renderer instead. ETag-based caching is disabled for rendered files.
func WithFileRenderer(path string, renderer FileRenderer) Option {
	return func(s *Server) {
		s.fileRenderers[path] = renderer
	}
}

// NewIndexRenderer parses index.html under distPath as an html/template and
// returns a renderer that executes it with data.
func NewIndexRenderer(staticFiles fs.FS, distPath string, data any) (FileRenderer, error) {
	tmpl, err := parseIndexTemplate(staticFiles, distPath)
	if err != nil {
		return nil, err
	}

	return func(w io.Writer, _ *http.Request) error {
		return tmpl.Execute(w, data)
	}, nil
}

// NewRequestIndexRenderer parses index.html under distPath as an html/template
// and returns a renderer that executes it with the value from data for each
// request.
func NewRequestIndexRenderer(
	staticFiles fs.FS,
	distPath string,
	data func(*http.Request) any,
) (FileRenderer, error) {
	tmpl, err := parseIndexTemplate(staticFiles, distPath)
	if err != nil {
		return nil, err
	}

	return func(w io.Writer, r *http.Request) error {
		return tmpl.Execute(w, data(r))
	}, nil
}

func parseIndexTemplate(staticFiles fs.FS, distPath string) (*template.Template, error) {
	subFS, err := fs.Sub(staticFiles, distPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", distPath, err)
	}

	indexBytes, err := fs.ReadFile(subFS, "index.html")
	if err != nil {
		return nil, fmt.Errorf("cannot read index.html: %w", err)
	}

	tmpl, err := template.New("index").Parse(string(indexBytes))
	if err != nil {
		return nil, fmt.Errorf("cannot parse index.html template: %w", err)
	}

	return tmpl, nil
}

func NewServer(staticFiles fs.FS, distPath string, gzipOptions GzipOptions, opts ...Option) (*Server, error) {
	subFS, err := fs.Sub(staticFiles, distPath)
	if err != nil {
		return nil, err
	}

	etags := make(map[string]string)
	compressible := make(map[string][]byte)

	err = fs.WalkDir(
		subFS,
		".",
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			info, err := d.Info()
			if err != nil {
				return err
			}

			content := make([]byte, info.Size())

			file, err := subFS.Open(path)
			if err != nil {
				return err
			}

			defer func() { _ = file.Close() }()

			_, err = file.Read(content)
			if err != nil {
				return err
			}

			hash := md5.Sum(content)
			etag := hex.EncodeToString(hash[:])
			etags["/"+path] = etag

			if gzipOptions.allows("/" + path) {
				compressible["/"+path] = content
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot generate etags: %w", err)
	}

	indexETag, ok := etags["/index.html"]
	if !ok {
		return nil, errors.New("index.html not found")
	}

	indexFile, err := subFS.Open("index.html")
	if err != nil {
		return nil, err
	}

	indexContent, err := io.ReadAll(indexFile)
	if err != nil {
		return nil, err
	}

	s := &Server{
		spaFS:         http.FS(subFS),
		indexETag:     indexETag,
		indexContent:  indexContent,
		etags:         etags,
		gzipOptions:   gzipOptions,
		fileRenderers: make(map[string]FileRenderer),
		gzipped:       make(map[string]gzippedFile),
	}

	for _, opt := range opts {
		opt(s)
	}

	for filePath, content := range compressible {
		// Rendered files change per request; they keep on-the-fly gzip.
		if _, ok := s.fileRenderers[filePath]; ok {
			continue
		}

		compressed, err := gzipBytes(content)
		if err != nil {
			return nil, fmt.Errorf("cannot gzip %s: %w", filePath, err)
		}

		if float64(len(compressed)) > float64(len(content))*minGzipRatio {
			continue
		}

		s.gzipped[filePath] = gzippedFile{
			content: compressed,
			// A strong ETag must differ per content encoding.
			etag:        etags[filePath] + "-gzip",
			contentType: contentTypeFor(filePath),
		}
	}

	return s, nil
}

func gzipBytes(content []byte) ([]byte, error) {
	var buf bytes.Buffer

	gz, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}

	if _, err := gz.Write(content); err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func contentTypeFor(filePath string) string {
	if contentType := mime.TypeByExtension(path.Ext(filePath)); contentType != "" {
		return contentType
	}

	return "application/octet-stream"
}

func cacheControlFor(filePath string) string {
	switch path.Ext(filePath) {
	case ".js", ".mjs", ".css", ".png", ".jpg", ".svg", ".woff2":
		return "public, max-age=31536000, immutable"
	default:
		return "public, max-age=3600"
	}
}

func (s *Server) serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	if renderer, ok := s.fileRenderers["/index.html"]; ok {
		var buf bytes.Buffer
		if err := renderer(&buf, r); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())

		return
	}

	w.Header().Set("ETag", `"`+s.indexETag+`"`)

	if r.Header.Get("If-None-Match") == `"`+s.indexETag+`"` {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(s.indexContent)
}

// HasFile reports whether filePath (e.g. "/assets/app.js") is a file in the
// embedded build, as opposed to a route the SPA fallback would answer with
// index.html.
func (s *Server) HasFile(filePath string) bool {
	_, ok := s.etags[filePath]
	return ok
}

func (s *Server) ServeSPA(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	f, err := s.spaFS.Open(path)
	if err != nil {
		s.serveIndex(w, r)
		return
	}

	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if info.IsDir() {
		s.serveIndex(w, r)
		return
	}

	if renderer, ok := s.fileRenderers[path]; ok {
		var buf bytes.Buffer
		if err := renderer(&buf, r); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())

		return
	}

	etag, ok := s.etags[path]
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	quotedETag := `"` + etag + `"`
	w.Header().Set("ETag", quotedETag)

	w.Header().Set("Cache-Control", cacheControlFor(path))

	if matchETag := r.Header.Get("If-None-Match"); matchETag != "" {
		if matchETag == quotedETag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	http.FileServer(s.spaFS).ServeHTTP(w, r)
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (o GzipOptions) allows(filePath string) bool {
	if !o.EnableFileTypeCheck {
		return true
	}

	for _, fileType := range o.FileTypes {
		if strings.HasSuffix(filePath, fileType) {
			return true
		}
	}

	return false
}

func acceptsGzip(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
}

func (s *Server) shouldCompressWithGzip(r *http.Request) bool {
	return acceptsGzip(r) && s.gzipOptions.allows(r.URL.Path)
}

// serveGzipped writes the precompressed variant of the requested file and
// reports whether it handled the request.
func (s *Server) serveGzipped(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return false
	}

	// Byte ranges address the identity encoding; let http.FileServer handle them.
	if r.Header.Get("Range") != "" || !acceptsGzip(r) {
		return false
	}

	file, ok := s.gzipped[r.URL.Path]
	if !ok {
		return false
	}

	quotedETag := `"` + file.etag + `"`

	w.Header().Set("ETag", quotedETag)
	w.Header().Set("Cache-Control", cacheControlFor(r.URL.Path))
	w.Header().Add("Vary", "Accept-Encoding")

	if r.Header.Get("If-None-Match") == quotedETag {
		w.WriteHeader(http.StatusNotModified)
		return true
	}

	w.Header().Set("Content-Type", file.contentType)
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Length", strconv.Itoa(len(file.content)))
	w.WriteHeader(http.StatusOK)

	if r.Method != http.MethodHead {
		_, _ = w.Write(file.content)
	}

	return true
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.serveGzipped(w, r) {
		return
	}

	if s.shouldCompressWithGzip(r) {
		w.Header().Add("Vary", "Accept-Encoding")
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)

		defer func() { _ = gz.Close() }()

		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		s.ServeSPA(gzw, r)

		return
	}

	s.ServeSPA(w, r)
}
