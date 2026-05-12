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

package files_v1

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/file"
	"go.probo.inc/probo/pkg/gid"
)

const presignedURLExpiry = 1 * time.Hour

type Handler struct {
	logger  *log.Logger
	fileSvc *file.Service
}

func NewMux(logger *log.Logger, fileSvc *file.Service) *chi.Mux {
	h := &Handler{
		logger:  logger,
		fileSvc: fileSvc,
	}

	r := chi.NewRouter()

	r.Get("/{fileID}", h.handleGetPublicFile)

	return r
}

func (h *Handler) handleGetPublicFile(w http.ResponseWriter, r *http.Request) {
	fileIDStr := chi.URLParam(r, "fileID")

	fileID, err := gid.ParseGID(fileIDStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	presignedURL, err := h.fileSvc.GetPublicFileURL(r.Context(), fileID, presignedURLExpiry)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			http.NotFound(w, r)
			return
		}

		h.logger.ErrorCtx(
			r.Context(),
			"cannot get public file URL",
			log.Error(err),
			log.String("file_id", fileIDStr),
		)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, presignedURL, http.StatusTemporaryRedirect)
}
