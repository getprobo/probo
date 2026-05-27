// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package oauth2server

import (
	"fmt"
	"net/url"
	"strconv"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/uri"
)

const (
	// WellKnownProtectedResourcePrefix is the default RFC 9728 path inserted
	// between the host and the resource path.
	WellKnownProtectedResourcePrefix = "/.well-known/oauth-protected-resource"
)

type (
	// ProtectedResourceMetadata is the RFC 9728 metadata document for an
	// OAuth 2.0 protected resource.
	ProtectedResourceMetadata struct {
		Resource               uri.URI                `json:"resource"`
		AuthorizationServers   []uri.URI              `json:"authorization_servers"`
		ScopesSupported        []coredata.OAuth2Scope `json:"scopes_supported,omitempty"`
		BearerMethodsSupported BearerMethods          `json:"bearer_methods_supported,omitempty"`
	}

	// ProtectedResourceConfig describes a protected resource served by Probo.
	ProtectedResourceConfig struct {
		Resource               uri.URI
		AuthorizationServers   []uri.URI
		ScopesSupported        coredata.OAuth2Scopes
		BearerMethodsSupported BearerMethods
	}
)

func WellKnownPath(resourcePath string) (string, error) {
	if resourcePath == "" {
		resourcePath = "/"
	}

	escapedPath, err := url.JoinPath(WellKnownProtectedResourcePrefix, resourcePath)
	if err != nil {
		return "", err
	}

	return url.PathUnescape(escapedPath)
}

func ProtectedResourceMetadataURL(resource uri.URI) (uri.URI, error) {
	u, err := url.Parse(resource.String())

	if err != nil {
		return "", fmt.Errorf("cannot parse resource URI: %w", err)
	}

	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("resource URI must be absolute: %q", resource)
	}

	wellKnownPath, err := WellKnownPath(u.Path)
	if err != nil {
		return "", fmt.Errorf("cannot build well-known path: %w", err)
	}

	meta := &url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   wellKnownPath,
	}

	return uri.Parse(meta.String())
}

func WWWAuthenticateHeader(resource uri.URI) (string, error) {
	metadataURL, err := ProtectedResourceMetadataURL(resource)
	if err != nil {
		return "", fmt.Errorf("cannot build protected resource metadata URL: %w", err)
	}

	return "Bearer resource_metadata=" + strconv.Quote(metadataURL.String()), nil
}

func NewProtectedResourceMetadata(cfg ProtectedResourceConfig) *ProtectedResourceMetadata {
	return &ProtectedResourceMetadata{
		Resource:               cfg.Resource,
		AuthorizationServers:   cfg.AuthorizationServers,
		ScopesSupported:        cfg.ScopesSupported,
		BearerMethodsSupported: cfg.BearerMethodsSupported,
	}
}
