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
	"go.probo.inc/probo/pkg/filesign"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/authz"
)

const presignedURLExpiry = 1 * time.Hour

type Handler struct {
	logger    *log.Logger
	fileSvc   *filesign.Service
	authorize authz.HTTPAuthorizeFunc
}

func NewMux(
	logger *log.Logger,
	fileSvc *filesign.Service,
	iamSvc *iam.Service,
	cookieConfig securecookie.Config,
	tokenSecret string,
) *chi.Mux {
	h := &Handler{
		logger:    logger,
		fileSvc:   fileSvc,
		authorize: authz.NewHTTPAuthorizeFunc(iamSvc, logger),
	}

	r := chi.NewRouter()

	r.Use(authn.NewAPIKeyMiddleware(iamSvc, tokenSecret))
	r.Use(authn.NewSessionMiddleware(iamSvc, cookieConfig))
	r.Use(authn.NewOAuth2AccessTokenMiddleware(iamSvc))

	r.Get("/{fileID}", h.handleGetFile)

	return r
}

func (h *Handler) handleGetFile(w http.ResponseWriter, r *http.Request) {
	fileIDStr := chi.URLParam(r, "fileID")

	fileID, err := gid.ParseGID(fileIDStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	file, err := h.fileSvc.LoadAnyActiveFile(r.Context(), fileID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			http.NotFound(w, r)
			return
		}

		h.logger.ErrorCtx(
			r.Context(),
			"cannot load file",
			log.Error(err),
			log.String("file_id", fileIDStr),
		)
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	if file.Visibility == coredata.FileVisibilityPrivate {
		// Return 404 on any auth failure — 401/403 would confirm the file
		// exists to an unauthorized caller, leaking private file existence.
		if authn.IdentityFromContext(r.Context()) == nil {
			http.NotFound(w, r)
			return
		}

		_, statusCode, err := h.authorize(r.Context(), fileID, probo.ActionFileDownloadUrl)
		if err != nil {
			if statusCode == http.StatusInternalServerError {
				h.logger.ErrorCtx(
					r.Context(),
					"cannot authorize file access",
					log.Error(err),
					log.String("file_id", fileIDStr),
				)
				http.Error(w, "internal server error", http.StatusInternalServerError)

				return
			}

			http.NotFound(w, r)

			return
		}
	}

	presignedURL, err := h.fileSvc.GeneratePresignedURLForFile(r.Context(), file, presignedURLExpiry)
	if err != nil {
		h.logger.ErrorCtx(
			r.Context(),
			"cannot generate presigned URL",
			log.Error(err),
			log.String("file_id", fileIDStr),
		)
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, presignedURL, http.StatusTemporaryRedirect)
}
