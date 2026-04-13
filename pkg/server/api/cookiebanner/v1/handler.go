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
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/cookiebanner"
)

type Handler struct {
	logger          *log.Logger
	cookieBannerSvc *cookiebanner.Service
}

func NewMux(
	logger *log.Logger,
	cookieBannerSvc *cookiebanner.Service,
) *chi.Mux {
	h := &Handler{
		logger:          logger,
		cookieBannerSvc: cookieBannerSvc,
	}

	r := chi.NewMux()
	r.Use(newCORSMiddleware(logger, cookieBannerSvc))
	r.Get("/{bannerID}/config", h.handleGetConfig)

	return r
}

func (h *Handler) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	httpserver.RenderJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
