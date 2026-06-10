// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/brand"
)

const staticFileCacheControl = "public, max-age=31536000, immutable"

type (
	staticFileServer struct {
		files map[string]staticFile
	}

	staticFile struct {
		content     []byte
		contentType string
		etag        string
		name        string
	}
)

var defaultStaticFileServer = mustNewStaticFileServer(brand.Assets)

func mustNewStaticFileServer(files fs.FS) *staticFileServer {
	server, err := newStaticFileServer(files)
	if err != nil {
		panic(err)
	}

	return server
}

func newStaticFileServer(files fs.FS) (*staticFileServer, error) {
	server := &staticFileServer{
		files: make(map[string]staticFile),
	}

	err := fs.WalkDir(
		files,
		".",
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return fmt.Errorf("cannot walk static file: %w", err)
			}

			if d.IsDir() {
				return nil
			}

			content, err := fs.ReadFile(files, path)
			if err != nil {
				return fmt.Errorf("cannot read static file: %w", err)
			}

			hash := md5.Sum(content)

			contentType := mime.TypeByExtension(filepath.Ext(path))
			if contentType == "" {
				contentType = http.DetectContentType(content)
			}

			server.files[path] = staticFile{
				content:     content,
				contentType: contentType,
				etag:        `"` + hex.EncodeToString(hash[:]) + `"`,
				name:        path,
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot index static files: %w", err)
	}

	return server, nil
}

func (s *staticFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request, file string) bool {
	asset, ok := s.files[file]
	if !ok {
		return false
	}

	h := w.Header()
	h.Set("Cache-Control", staticFileCacheControl)
	h.Set("Content-Type", asset.contentType)
	h.Set("ETag", asset.etag)
	addVary(h, "Accept-Encoding")

	if ifNoneMatch(r.Header.Get("If-None-Match"), asset.etag) {
		w.WriteHeader(http.StatusNotModified)
		return true
	}

	if shouldCompressStaticFile(r) {
		h.Set("Content-Encoding", "gzip")
		w.WriteHeader(http.StatusOK)

		gz := gzip.NewWriter(w)
		defer func() { _ = gz.Close() }()

		_, _ = gz.Write(asset.content)

		return true
	}

	http.ServeContent(w, r, asset.name, time.Time{}, bytes.NewReader(asset.content))

	return true
}

func shouldCompressStaticFile(r *http.Request) bool {
	if r.Header.Get("Range") != "" {
		return false
	}

	gzipQ := -1.0
	wildcardQ := -1.0

	for encoding := range strings.SplitSeq(r.Header.Get("Accept-Encoding"), ",") {
		encoding, q := parseAcceptEncoding(encoding)
		switch encoding {
		case "gzip":
			gzipQ = q
		case "*":
			wildcardQ = q
		}
	}

	if gzipQ >= 0 {
		return gzipQ > 0
	}

	return wildcardQ > 0
}

func parseAcceptEncoding(encoding string) (string, float64) {
	parts := strings.Split(encoding, ";")
	token := strings.ToLower(strings.TrimSpace(parts[0]))
	q := 1.0

	for _, param := range parts[1:] {
		key, value, ok := strings.Cut(strings.TrimSpace(param), "=")
		if !ok || !strings.EqualFold(key, "q") {
			continue
		}

		parsedQ, err := strconv.ParseFloat(value, 64)
		if err == nil {
			q = parsedQ
		}
	}

	return token, q
}

func ifNoneMatch(header string, etag string) bool {
	for candidate := range strings.SplitSeq(header, ",") {
		candidate = strings.TrimSpace(candidate)
		if candidate == "*" {
			return true
		}

		candidate = strings.TrimPrefix(candidate, "W/")
		candidate = strings.TrimPrefix(candidate, "w/")

		if candidate == etag {
			return true
		}
	}

	return false
}

func addVary(h http.Header, value string) {
	for _, vary := range h.Values("Vary") {
		for field := range strings.SplitSeq(vary, ",") {
			if strings.EqualFold(strings.TrimSpace(field), value) {
				return
			}
		}
	}

	if h.Get("Vary") == "" {
		h.Set("Vary", value)
		return
	}

	h.Add("Vary", value)
}
