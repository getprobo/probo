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
	"net/http"

	"go.probo.inc/probo/apps/console"
	"go.probo.inc/probo/pkg/server/statichandler"
)

type Server struct {
	*statichandler.Server
}

func NewServer() (*Server, error) {
	gzipOptions := statichandler.GzipOptions{
		EnableFileTypeCheck: false,
	}

	spaServer, err := statichandler.NewServer(console.StaticFiles, "dist", gzipOptions)
	if err != nil {
		return nil, err
	}

	return &Server{
		Server: spaServer,
	}, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Server.ServeHTTP(w, r)
}

func (s *Server) ServeSPA(w http.ResponseWriter, r *http.Request) {
	s.Server.ServeSPA(w, r)
}
