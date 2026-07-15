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

package complianceportal

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
)

func TestTrustedRequestHost_PrefersTLSServerName(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "https://evil.example.com/graphql", nil)
	req.Host = "evil.example.com"
	req.TLS = &tls.ConnectionState{ServerName: "portal.example.com"}

	host, ok := TrustedRequestHost(req)

	require.True(t, ok)
	assert.Equal(t, "portal.example.com", host)
}

func TestSessionHostMiddleware_RejectsMismatchedHost(t *testing.T) {
	t.Parallel()

	var authenticated bool

	handler := NewSessionHostMiddleware(securecookie.Config{Name: "ssid"})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authenticated = authn.IdentityFromContext(r.Context()) != nil
			w.WriteHeader(http.StatusOK)
		}),
	)

	identity := &coredata.Identity{ID: gid.New(gid.NilTenant, coredata.IdentityEntityType)}
	session := &coredata.Session{
		ID:   identity.ID,
		Data: coredata.SessionDataForHost("portal-a.example.com"),
	}

	req := httptest.NewRequest(http.MethodGet, "/graphql", nil)
	req.Host = "portal-b.example.com"
	req.TLS = &tls.ConnectionState{ServerName: "portal-b.example.com"}
	req = req.WithContext(authn.ContextWithIdentity(req.Context(), identity))
	req = req.WithContext(authn.ContextWithSession(req.Context(), session))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.False(t, authenticated)
}

func TestSessionHostMiddleware_AllowsMatchingTLSHost(t *testing.T) {
	t.Parallel()

	var authenticated bool

	handler := NewSessionHostMiddleware(securecookie.Config{Name: "ssid"})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authenticated = authn.IdentityFromContext(r.Context()) != nil
			w.WriteHeader(http.StatusOK)
		}),
	)

	identity := &coredata.Identity{ID: gid.New(gid.NilTenant, coredata.IdentityEntityType)}
	session := &coredata.Session{
		ID:   identity.ID,
		Data: coredata.SessionDataForHost("portal.example.com"),
	}

	req := httptest.NewRequest(http.MethodGet, "/graphql", nil)
	req.Host = "evil.example.com"
	req.TLS = &tls.ConnectionState{ServerName: "portal.example.com"}
	req = req.WithContext(authn.ContextWithIdentity(req.Context(), identity))
	req = req.WithContext(authn.ContextWithSession(req.Context(), session))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, authenticated)
}

func TestSessionHostMiddleware_RejectsSpoofedHostHeader(t *testing.T) {
	t.Parallel()

	var authenticated bool

	handler := NewSessionHostMiddleware(securecookie.Config{Name: "ssid"})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authenticated = authn.IdentityFromContext(r.Context()) != nil
			w.WriteHeader(http.StatusOK)
		}),
	)

	identity := &coredata.Identity{ID: gid.New(gid.NilTenant, coredata.IdentityEntityType)}
	session := &coredata.Session{
		ID:   identity.ID,
		Data: coredata.SessionDataForHost("portal.example.com"),
	}

	req := httptest.NewRequest(http.MethodGet, "/graphql", nil)
	req.Host = "portal.example.com"
	req.TLS = &tls.ConnectionState{ServerName: "other.example.com"}
	req = req.WithContext(authn.ContextWithIdentity(req.Context(), identity))
	req = req.WithContext(authn.ContextWithSession(req.Context(), session))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.False(t, authenticated)
}
