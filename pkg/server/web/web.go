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
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/getprobo/probo/apps/console"
)

type Server struct {
	spaFS        http.FileSystem
	etags        map[string]string
	indexETag    string
	indexContent []byte
}

func NewServer() (*Server, error) {
	if os.Getenv("APP_ENV") == "dev" {
		return NewDevServer()
	}
	subFS, err := fs.Sub(console.StaticFiles, "dist")
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
			defer file.Close()

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
	}, nil
}

func NewDevServer() (*Server, error) {
	publicFS := os.DirFS("./apps/console/public")

	etags := make(map[string]string)

	html, err := os.ReadFile("./apps/console/index.html")
	if err != nil {
		return nil, err
	}

	// Inject vite client code
	re := regexp.MustCompile(`(?m)^\s*<script.*src="([^"]*)".*$`)
	matches := re.FindStringSubmatch(string(html))
	if len(matches) < 2 {
		return nil, fmt.Errorf("no script tag found in the index.html template")
	}

	htmlStr := re.ReplaceAllString(string(html),
		fmt.Sprintf(`
		<script type="module" src="http://[::1]:%[1]d/@vite/client"></script>
		<script src="http://[::1]:%[1]d%[2]s" type="module"></script>
		`, 5173, matches[1]))

	// Add react preamble for react plugin
	htmlStr = strings.ReplaceAll(htmlStr, "<body", fmt.Sprintf(`
		<body>
		<script type="module">
			import RefreshRuntime from 'http://[::1]:%[1]d/@react-refresh'
			RefreshRuntime.injectIntoGlobalHook(window)
			window.$RefreshReg$ = () => {}
			window.$RefreshSig$ = () => (type) => type
			window.__vite_plugin_react_preamble_installed__ = true
		</script><body`, 5173))

	return &Server{
		spaFS:        http.FS(publicFS),
		indexETag:    "",
		indexContent: []byte(htmlStr),
		etags:        etags,
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
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		w.WriteHeader(http.StatusOK)
		w.Write(s.indexContent)
		return
	}

	defer f.Close()

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
		w.Write(s.indexContent)
		return
	}

	etag, ok := s.etags[path]
	if !ok {
		fmt.Printf("XXX %+#v\n", path)

		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("ETag", etag)

	if matchETag := r.Header.Get("If-None-Match"); matchETag != "" {
		if matchETag == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	if strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".css") ||
		strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".jpg") ||
		strings.HasSuffix(path, ".svg") || strings.HasSuffix(path, ".woff2") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}

	w.Header().Set("Transfer-Encoding", "chunked")

	http.FileServer(s.spaFS).ServeHTTP(w, r)
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
