// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package authn

import (
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
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
					if _, ok := errors.AsType[*iam.ErrSessionNotFound](err); ok {
						securecookie.Clear(w, cookieConfig)
						next.ServeHTTP(w, r)

						return
					}

					if _, ok := errors.AsType[*iam.ErrSessionExpired](err); ok {
						securecookie.Clear(w, cookieConfig)
						next.ServeHTTP(w, r)

						return
					}

					panic(fmt.Errorf("cannot get session: %w", err))
				}

				identity, err := svc.AccountService.GetIdentity(ctx, session.IdentityID)
				if err != nil {
					if _, ok := errors.AsType[*iam.ErrIdentityNotFound](err); ok {
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

				httpserver.LoggerFromContext(ctx).InfoCtx(
					ctx,
					"session authenticated",
					log.String("identity_id", identity.ID.String()),
					log.String("session_id", session.ID.String()),
				)

				next.ServeHTTP(w, r.WithContext(ctx))

				err = svc.SessionService.UpdateSessionData(ctx, session.ID, session.Data)
				if err != nil {
					panic(fmt.Errorf("cannot update session data: %w", err))
				}
			},
		)
	}
}
