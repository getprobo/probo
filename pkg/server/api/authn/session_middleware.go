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
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/jackc/puddle/v2"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/clientip"
	"go.probo.inc/probo/pkg/server/gqlx"
)

// ignoreSessionDataUpdateError reports errors that happen because the request is
// already over. The database client does not wrap the request context error, so
// a failed write can only be attributed to cancellation when the context was
// already canceled before the write started; otherwise a real session-not-found
// or write error must surface. Shutdown closes the pool while a hijacked socket
// is still leaving the handler, so a closed pool is always ignored.
func ignoreSessionDataUpdateError(canceledBeforeWrite bool, err error) bool {
	if canceledBeforeWrite {
		return true
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	return errors.Is(err, puddle.ErrClosedPool)
}

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
								gqlx.Conflictf(ctx, "session authentication cannot be used with API key authentication"),
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
				ipAddress := net.ParseIP(clientip.Extract(r))

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

				// A collaboration socket is hijacked, so a panic here is written
				// onto a connection that can no longer carry an HTTP response.
				// Shutdown closes that socket and then the database pool.
				canceledBeforeWrite := ctx.Err() != nil

				err = svc.SessionService.UpdateSessionData(ctx, session.ID, session.Data)
				if err != nil && !ignoreSessionDataUpdateError(canceledBeforeWrite, err) {
					panic(fmt.Errorf("cannot update session data: %w", err))
				}
			},
		)
	}
}
