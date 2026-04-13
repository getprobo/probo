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

package cookiebanner_v1

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/gid"
)

func newCORSMiddleware(logger *log.Logger, cookieBannerSvc *cookiebanner.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			bannerIDStr := chi.URLParam(r, "bannerID")
			if bannerIDStr == "" {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			bannerID, err := gid.ParseGID(bannerIDStr)
			if err != nil {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			banner, err := cookieBannerSvc.GetActiveCookieBanner(r.Context(), bannerID)
			if err != nil {
				if errors.Is(err, cookiebanner.ErrBannerNotFound) {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
				logger.ErrorCtx(r.Context(), "cannot load cookie banner for CORS check", log.Error(err))
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			canonicalOrigin := cookiebanner.CanonicalizeOrigin(origin)
			if banner.Origin != canonicalOrigin {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Max-Age", "600")
			w.Header().Set("Vary", "Origin")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
