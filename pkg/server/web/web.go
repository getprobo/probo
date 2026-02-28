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
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"go.probo.inc/probo/apps/console"
	"go.probo.inc/probo/pkg/server/statichandler"
)

type Server struct {
	*statichandler.Server
	devProxy *httputil.ReverseProxy
}

func NewServer() (*Server, error) {
	return NewServerWithDevAddr("")
}

// NewServerWithDevAddr creates a new Server with optional dev server support.
// If devAddr is provided, requests will be proxied to the Vite dev server.
// If devAddr is empty, the embedded static files will be served.
// The devAddr should be in the format "http://localhost:5173" or be read from environment.
func NewServerWithDevAddr(devAddr string) (*Server, error) {
	gzipOptions := statichandler.GzipOptions{
		EnableFileTypeCheck: false,
	}

	spaServer, err := statichandler.NewServer(console.StaticFiles, "dist", gzipOptions)
	if err != nil {
		return nil, err
	}

	server := &Server{
		Server: spaServer,
	}

	// Support VITE_DEV_SERVER environment variable for automatic dev mode
	if devAddr == "" {
		devAddr = os.Getenv("VITE_DEV_SERVER_CONSOLE")
	}

	if devAddr != "" {
		// Validate and set up reverse proxy to dev server
		if err := server.setupDevProxy(devAddr); err != nil {
			return nil, err
		}
	}

	return server, nil
}

func (s *Server) setupDevProxy(devAddr string) error {
	devURL, err := url.Parse(devAddr)
	if err != nil {
		return err
	}

	s.devProxy = httputil.NewSingleHostReverseProxy(devURL)

	// Customize the reverse proxy to handle WebSocket and other dev server features
	s.devProxy.Director = func(req *http.Request) {
		req.URL.Scheme = devURL.Scheme
		req.URL.Host = devURL.Host
		// Preserve the original host for request headers if needed
		req.Host = devURL.Host
	}

	// Handle upgrade requests for WebSocket (HMR)
	originalTransport := s.devProxy.Transport
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}

	s.devProxy.Transport = &http.Transport{
		Dial:                (&net.Dialer{}).Dial,
		TLSHandshakeTimeout: originalTransport.(*http.Transport).TLSHandshakeTimeout,
	}

	return nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// If dev proxy is configured, use it
	if s.devProxy != nil {
		s.devProxy.ServeHTTP(w, r)
		return
	}

	// Otherwise, use the static file handler
	s.Server.ServeHTTP(w, r)
}

func (s *Server) ServeSPA(w http.ResponseWriter, r *http.Request) {
	// If dev proxy is configured, use it
	if s.devProxy != nil {
		s.devProxy.ServeHTTP(w, r)
		return
	}

	// Otherwise, use the static SPA handler
	s.Server.ServeSPA(w, r)
}
