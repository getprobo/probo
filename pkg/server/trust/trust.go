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
