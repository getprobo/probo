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

package cookiebanner_v1

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/jsonx"
)

func newCORSMiddleware(logger *log.Logger, cookieBannerSvc *cookiebanner.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				origin := r.Header.Get("Origin")
				if origin == "" {
					next.ServeHTTP(w, r)
					return
				}

				bannerIDStr := chi.URLParam(r, "bannerID")
				if bannerIDStr == "" {
					jsonx.RenderForbidden(w)
					return
				}

				bannerID, err := gid.ParseGID(bannerIDStr)
				if err != nil {
					jsonx.RenderForbidden(w)
					return
				}

				banner, err := cookieBannerSvc.GetActiveCookieBanner(r.Context(), bannerID)
				if err != nil {
					if errors.Is(err, cookiebanner.ErrBannerNotFound) {
						jsonx.RenderForbidden(w)
						return
					}

					logger.ErrorCtx(r.Context(), "cannot load cookie banner for CORS check", log.Error(err))
					jsonx.RenderInternalServerError(w)

					return
				}

				canonicalOrigin := cookiebanner.CanonicalizeOrigin(origin)
				if banner.Origin != canonicalOrigin {
					jsonx.RenderForbidden(w)
					return
				}

				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-SDK-Version")
				w.Header().Set("Access-Control-Max-Age", "600")
				w.Header().Set("Vary", "Origin")

				if r.Method == http.MethodOptions {
					w.WriteHeader(http.StatusNoContent)
					return
				}

				next.ServeHTTP(w, r)
			},
		)
	}
}
