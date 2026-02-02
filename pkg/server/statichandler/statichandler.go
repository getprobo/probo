// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

// Package statichandler provides functionality for serving SPA (Single Page Application) frontends.
package statichandler

import (
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"
)

type GzipOptions struct {
	EnableFileTypeCheck bool
	FileTypes           []string
}

func DefaultGzipOptions() GzipOptions {
	return GzipOptions{
		EnableFileTypeCheck: true,
		FileTypes:           []string{".js", ".css", ".html"},
	}
}

type Server struct {
	spaFS        http.FileSystem
	etags        map[string]string
	indexETag    string
	indexContent []byte
	gzipOptions  GzipOptions
}

func NewServer(staticFiles fs.FS, distPath string, gzipOptions GzipOptions) (*Server, error) {
	subFS, err := fs.Sub(staticFiles, distPath)
	if err != nil {
		return nil, err
	}

	etags := make(map[string]string)

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

	return &Server{
		spaFS:        http.FS(subFS),
		indexETag:    indexETag,
		indexContent: indexContent,
		etags:        etags,
		gzipOptions:  gzipOptions,
	}, nil
}

func (s *Server) ServeSPA(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	f, err := s.spaFS.Open(path)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("ETag", `"`+s.indexETag+`"`)

		if r.Header.Get("If-None-Match") == `"`+s.indexETag+`"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(s.indexContent)
		return
	}

	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if info.IsDir() {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("ETag", `"`+s.indexETag+`"`)

		if r.Header.Get("If-None-Match") == `"`+s.indexETag+`"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(s.indexContent)
		return
	}

	etag, ok := s.etags[path]
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	quotedETag := `"` + etag + `"`
	w.Header().Set("ETag", quotedETag)

	if strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".css") ||
		strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".jpg") ||
		strings.HasSuffix(path, ".svg") || strings.HasSuffix(path, ".woff2") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}

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

func (s *Server) shouldCompressWithGzip(r *http.Request) bool {
	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		return false
	}

	if !s.gzipOptions.EnableFileTypeCheck {
		return true
	}

	path := r.URL.Path
	for _, fileType := range s.gzipOptions.FileTypes {
		if strings.HasSuffix(path, fileType) {
			return true
		}
	}

	return false
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.shouldCompressWithGzip(r) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer func() { _ = gz.Close() }()

		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		s.ServeSPA(gzw, r)
		return
	}

	s.ServeSPA(w, r)
}
