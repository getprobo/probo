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

package compliancepage

import (
	"errors"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/server/gqlutils"
	"go.probo.inc/probo/pkg/trust"
)

func NewNDAMiddleware(trustSvc *trust.Service, esignSvc *esign.Service, logger *log.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()

				membership := ComplianceMembershipFromContext(ctx)

				if membership == nil {
					next.ServeHTTP(w, r)
					return
				}

				compliancePage := CompliancePageFromContext(ctx)

				if _, err := trustSvc.GetNDAFile(ctx, compliancePage.ID); err != nil {
					if errors.Is(err, trust.ErrNDAFileNotFound) {
						next.ServeHTTP(w, r)
						return
					}

					logger.ErrorCtx(ctx, "cannot get NDA file", log.Error(err))
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

				if membership.ElectronicSignatureID == nil {
					next.ServeHTTP(w, r)
					return
				}

				sig, err := esignSvc.GetSignatureByID(ctx, *membership.ElectronicSignatureID)
				if err != nil {
					logger.ErrorCtx(ctx, "cannot get NDA signature", log.Error(err))
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

				if sig.Status != coredata.ElectronicSignatureStatusCompleted {
					httpserver.RenderJSON(
						w,
						http.StatusForbidden,
						&graphql.Response{
							Errors: gqlerror.List{
								gqlutils.NDASignatureRequiredf(ctx, "NDA signature required"),
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
