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

package authn

import (
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/kit/httpserver"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

func NewSessionMiddleware(svc *iam.Service, cookieConfig securecookie.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()

				cookieValue, err := securecookie.Get(r, cookieConfig)
				if err != nil {
					next.ServeHTTP(w, r)
					return
				}

				sessionID, err := gid.ParseGID(cookieValue)
				if err != nil {
					securecookie.Clear(w, cookieConfig)
					next.ServeHTTP(w, r)
					return
				}

				apiKey := APIKeyFromContext(ctx)
				if sessionID != gid.Nil && apiKey != nil {
					httpserver.RenderJSON(
						w,
						http.StatusUnauthorized,
						&graphql.Response{
							Errors: gqlerror.List{
								gqlutils.Conflictf(ctx, "session authentication cannot be used with API key authentication"),
							},
						},
					)
					return
				}

				session, err := svc.SessionService.GetSession(ctx, sessionID)
				if err != nil {
					var errSessionNotFound *iam.ErrSessionNotFound
					var errSessionExpired *iam.ErrSessionExpired

					if errors.As(err, &errSessionNotFound) || errors.As(err, &errSessionExpired) {
						securecookie.Clear(w, cookieConfig)
						next.ServeHTTP(w, r)
						return
					}

					panic(fmt.Errorf("cannot get session: %w", err))
				}

				identity, err := svc.AccountService.GetIdentity(ctx, session.IdentityID)
				if err != nil {
					var errIdentityNotFound *iam.ErrIdentityNotFound
					if errors.As(err, &errIdentityNotFound) {
						securecookie.Clear(w, cookieConfig)
						next.ServeHTTP(w, r)
						return
					}

					panic(fmt.Errorf("cannot get identity: %w", err))
				}

				userAgent := r.UserAgent()
				// TODO: will work well when no layer 7 proxy is in front of the server
				var ipAddress net.IP
				if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
					ipAddress = net.ParseIP(host)
				} else {
					ipAddress = net.ParseIP(r.RemoteAddr)
				}

				err = svc.SessionService.UpdateSessionInfo(ctx, session.ID, userAgent, ipAddress)
				if err != nil {
					panic(fmt.Errorf("cannot update session info: %w", err))
				}

				ctx = ContextWithSession(ctx, session)
				ctx = ContextWithIdentity(ctx, identity)

				next.ServeHTTP(w, r.WithContext(ctx))

				err = svc.SessionService.UpdateSessionData(ctx, session.ID, session.Data)
				if err != nil {
					panic(fmt.Errorf("cannot update session data: %w", err))
				}
			},
		)
	}
}
