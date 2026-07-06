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
	"go.probo.inc/probo/pkg/server/api/clientip"
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

				err = svc.SessionService.UpdateSessionData(ctx, session.ID, session.Data)
				if err != nil {
					panic(fmt.Errorf("cannot update session data: %w", err))
				}
			},
		)
	}
}
