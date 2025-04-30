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

package web

import (
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/getprobo/probo/apps/console"
)

type Server struct {
	spaFS        http.FileSystem
	indexContent []byte
	indexETag    string
}

func NewServer() (*Server, error) {
	subFS, err := fs.Sub(console.StaticFiles, "dist")
	if err != nil {
		return nil, err
	}

	indexFile, err := subFS.Open("index.html")
	if err != nil {
		return nil, err
	}
	defer indexFile.Close()

	stat, err := indexFile.Stat()
	if err != nil {
		return nil, err
	}

	indexContent := make([]byte, stat.Size())
	_, err = indexFile.Read(indexContent)
	if err != nil {
		return nil, err
	}

	// Generate ETag for index.html
	hash := md5.Sum(indexContent)
	indexETag := hex.EncodeToString(hash[:])

	return &Server{
		spaFS:        http.FS(subFS),
		indexContent: indexContent,
		indexETag:    indexETag,
	}, nil
}

func (s *Server) ServeSPA(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	f, err := s.spaFS.Open(path)
	if err == nil {
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		etag := fmt.Sprintf(`"%x-%x"`, info.Size(), info.ModTime().Unix())

		lastModified := info.ModTime().Format(http.TimeFormat)

		w.Header().Set("ETag", etag)
		w.Header().Set("Last-Modified", lastModified)

		if matchETag := r.Header.Get("If-None-Match"); matchETag != "" {
			if matchETag == etag {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		if modifiedSince := r.Header.Get("If-Modified-Since"); modifiedSince != "" {
			if t, err := time.Parse(http.TimeFormat, modifiedSince); err == nil {
				if info.ModTime().Unix() <= t.Unix() {
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}
		}

		if strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".css") ||
			strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".jpg") ||
			strings.HasSuffix(path, ".svg") || strings.HasSuffix(path, ".woff2") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable") // 1 year
		} else {
			w.Header().Set("Cache-Control", "public, max-age=3600") // 1 hour
		}

		w.Header().Set("Transfer-Encoding", "chunked")

		// Serve the file
		http.FileServer(s.spaFS).ServeHTTP(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("ETag", `"`+s.indexETag+`"`)

	if r.Header.Get("If-None-Match") == `"`+s.indexETag+`"` {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	w.WriteHeader(http.StatusOK)
	w.Write(s.indexContent)
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		s.ServeSPA(gzipResponseWriter{Writer: gz, ResponseWriter: w}, r)
		return
	}

	s.ServeSPA(w, r)
}
