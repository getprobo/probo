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

package console_v1

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

// newMembershipAccessMiddleware refuses console requests from an authenticated
// identity that belongs to no organization, but only when public signup is
// disabled. It closes the gap where a self-service magic-link or OIDC sign-in
// creates an identity that can then reach the console (and create an
// organization) on an instance meant to stay private. Compliance-portal
// visitors authenticate against a different API and are unaffected.
//
// It must run after the identity-presence middleware so the identity is
// already resolved.
func newMembershipAccessMiddleware(iamSvc *iam.Service, logger *log.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()

				if iamSvc.IsSignUpEnabled() {
					next.ServeHTTP(w, r)
					return
				}

				identity := authn.IdentityFromContext(ctx)
				if identity == nil {
					next.ServeHTTP(w, r)
					return
				}

				hasMembership, err := iamSvc.IdentityHasMembership(ctx, identity.ID, nil)
				if err != nil {
					logger.ErrorCtx(ctx, "cannot check identity membership", log.Error(err))
					httpserver.RenderJSON(
						w,
						http.StatusInternalServerError,
						&graphql.Response{
							Errors: gqlerror.List{
								gqlutils.Internal(ctx),
							},
						},
					)

					return
				}

				if !hasMembership {
					httpserver.RenderJSON(
						w,
						http.StatusForbidden,
						&graphql.Response{
							Errors: gqlerror.List{
								gqlutils.MembershipRequiredf(
									ctx,
									"an organization membership is required to access this resource",
								),
							},
						},
					)

					return
				}

				next.ServeHTTP(w, r)
			},
		)
	}
}
