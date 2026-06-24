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

// Package trust provides functionality for serving the trust center SPA frontend.
package trust

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"

	truststatics "go.probo.inc/probo/apps/trust"
	"go.probo.inc/probo/pkg/server/statichandler"
)

type (
	HeadData struct {
		Title       string
		Description string
		OGURL       string
		FaviconURL  string
	}

	HeadDataFunc func(r *http.Request) HeadData

	Server struct {
		*statichandler.Server
	}
)

func NewServer(headDataFunc HeadDataFunc) (*Server, error) {
	renderer, err := buildIndexRenderer(headDataFunc)
	if err != nil {
		return nil, err
	}

	gzipOptions := statichandler.GzipOptions{
		EnableFileTypeCheck: true,
		FileTypes:           []string{".js", ".css", ".html"},
	}

	spaServer, err := statichandler.NewServer(
		truststatics.StaticFiles,
		"dist",
		gzipOptions,
		statichandler.WithFileRenderer("/index.html", renderer),
	)
	if err != nil {
		return nil, err
	}

	return &Server{Server: spaServer}, nil
}

func buildIndexRenderer(headDataFunc HeadDataFunc) (statichandler.FileRenderer, error) {
	subFS, err := fs.Sub(truststatics.StaticFiles, "dist")
	if err != nil {
		return nil, fmt.Errorf("cannot open dist: %w", err)
	}

	indexBytes, err := fs.ReadFile(subFS, "index.html")
	if err != nil {
		return nil, fmt.Errorf("cannot read index.html: %w", err)
	}

	tmpl, err := template.New("index").Parse(string(indexBytes))
	if err != nil {
		return nil, fmt.Errorf("cannot parse index.html template: %w", err)
	}

	return func(w io.Writer, r *http.Request) error {
		return tmpl.Execute(w, headDataFunc(r))
	}, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Server.ServeHTTP(w, r)
}

func (s *Server) ServeSPA(w http.ResponseWriter, r *http.Request) {
	s.Server.ServeSPA(w, r)
}
