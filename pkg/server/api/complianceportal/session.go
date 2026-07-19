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
	"net/http"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
)

func TrustedRequestHost(r *http.Request) (string, bool) {
	if r.TLS != nil && r.TLS.ServerName != "" {
		return coredata.NormalizeBoundHost(r.TLS.ServerName), true
	}

	if r.Host == "" {
		return "", false
	}

	return coredata.NormalizeBoundHost(r.Host), true
}

func NewSessionHostMiddleware(cookieConfig securecookie.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()

				session := authn.SessionFromContext(ctx)
				if session == nil {
					next.ServeHTTP(w, r)
					return
				}

				host, ok := TrustedRequestHost(r)
				if ok && session.Data.MatchesBoundHost(host) {
					next.ServeHTTP(w, r)
					return
				}

				securecookie.Clear(w, cookieConfig)

				ctx = authn.ContextWithSession(ctx, nil)
				ctx = authn.ContextWithIdentity(ctx, nil)
				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}
