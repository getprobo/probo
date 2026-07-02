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

package files_v1

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/bearertoken"
	"go.probo.inc/probo/pkg/brand"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/jsonx"
)

const presignedURLExpiry = 1 * time.Hour

type Handler struct {
	logger  *log.Logger
	fileSvc *filemanager.Service
	probo   *probo.Service
	iamSvc  *iam.Service
	assets  *brand.Assets
	baseURL *baseurl.BaseURL
}

func NewMux(
	logger *log.Logger,
	fileSvc *filemanager.Service,
	proboSvc *probo.Service,
	iamSvc *iam.Service,
	cookieConfig securecookie.Config,
	tokenSecret string,
	baseURL *baseurl.BaseURL,
) *chi.Mux {
	h := &Handler{
		logger:  logger,
		fileSvc: fileSvc,
		probo:   proboSvc,
		iamSvc:  iamSvc,
		assets:  brand.NewAssets(),
		baseURL: baseURL,
	}

	r := chi.NewRouter()

	r.Get("/static/{file}", h.handleGetStaticFile)
	r.Get("/public/{fileID}", h.handleGetPublicFile)

	r.Group(func(r chi.Router) {
		r.Use(authn.NewSessionMiddleware(iamSvc, cookieConfig))
		r.Use(authn.NewAPIKeyMiddleware(iamSvc, tokenSecret))
		r.Use(authn.NewOAuth2AccessTokenMiddleware(iamSvc))
		r.Use(authn.NewIdentityPresenceMiddleware(baseURL))
		r.Get("/{fileID}", h.handleGetFile)
	})

	return r
}

func (h *Handler) handleGetStaticFile(w http.ResponseWriter, r *http.Request) {
	file := chi.URLParam(r, "file")

	if _, statErr := h.assets.Stat(file); statErr != nil {
		jsonx.RenderNotFound(w, fmt.Errorf("file not found"))
		return
	}

	// ServeAssets honors the ETag header we set above for If-None-Match (and
	// If-Range), so it emits 304 Not Modified and handles range requests without
	// any extra conditional logic here.
	h.assets.ServeAssets(w, r, file)
}

func (h *Handler) handleGetPublicFile(w http.ResponseWriter, r *http.Request) {
	fileIDStr := chi.URLParam(r, "fileID")

	fileID, err := gid.ParseGID(fileIDStr)
	if err != nil {
		jsonx.RenderNotFound(w, fmt.Errorf("file not found"))
		return
	}

	file, err := h.fileSvc.GetPublicFile(r.Context(), fileID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			jsonx.RenderNotFound(w, fmt.Errorf("file not found"))
			return
		}

		h.logger.ErrorCtx(
			r.Context(),
			"cannot get public file URL",
			log.Error(err),
			log.String("file_id", fileIDStr),
		)
		jsonx.RenderInternalServerError(w)

		return
	}

	conds := filemanager.FileConditions{IfNoneMatch: r.Header.Get("If-None-Match")}
	if ifModifiedSince := r.Header.Get("If-Modified-Since"); ifModifiedSince != "" {
		if t, parseErr := http.ParseTime(ifModifiedSince); parseErr == nil {
			conds.IfModifiedSince = t
		}
	}

	obj, err := h.fileSvc.OpenFile(r.Context(), file, conds)
	if err != nil {
		h.logger.ErrorCtx(
			r.Context(),
			"cannot open public file",
			log.Error(err),
			log.String("file_id", fileIDStr),
		)
		jsonx.RenderInternalServerError(w)

		return
	}

	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

	if obj.ETag != "" {
		w.Header().Set("ETag", obj.ETag)
	}

	if obj.NotModified {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	defer func() { _ = obj.Body.Close() }()

	w.Header().Set("Content-Type", file.MimeType)
	w.Header().Set("Content-Length", strconv.FormatInt(file.FileSize, 10))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; sandbox")

	if !obj.LastModified.IsZero() {
		w.Header().Set("Last-Modified", obj.LastModified.UTC().Format(http.TimeFormat))
	}

	if _, err := io.Copy(w, obj.Body); err != nil {
		h.logger.ErrorCtx(
			r.Context(),
			"cannot stream public file",
			log.Error(err),
			log.String("file_id", fileIDStr),
		)

		return
	}
}

func (h *Handler) handleGetFile(w http.ResponseWriter, r *http.Request) {
	fileIDStr := chi.URLParam(r, "fileID")

	fileID, err := gid.ParseGID(fileIDStr)
	if err != nil {
		jsonx.RenderNotFound(w, fmt.Errorf("file not found"))
		return
	}

	ctx := r.Context()
	identity := authn.IdentityFromContext(ctx)
	session := authn.SessionFromContext(ctx)

	params := iam.AuthorizeParams{
		Principal:          identity.ID,
		Resource:           fileID,
		Action:             probo.ActionFileGet,
		ResourceAttributes: make(map[string]string),
	}
	if session != nil {
		params.Session = &session.ID
	}

	scope, err := h.iamSvc.Authorizer.Authorize(ctx, params)
	if err != nil {
		if scopeErr, ok := errors.AsType[*iam.ErrInsufficientOAuth2Scope](err); ok {
			bearertoken.SetBearerInsufficientScope(w, h.baseURL, scopeErr.Scopes...)
			jsonx.RenderForbidden(w)

			return
		}

		if _, ok := errors.AsType[*iam.ErrInsufficientPermissions](err); ok {
			jsonx.RenderForbidden(w)
			return
		}

		jsonx.RenderNotFound(w, fmt.Errorf("file not found"))

		return
	}

	f, err := h.probo.Files.Get(ctx, scope, fileID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			jsonx.RenderNotFound(w, fmt.Errorf("file not found"))
			return
		}

		h.logger.ErrorCtx(ctx, "cannot get file", log.Error(err), log.String("file_id", fileIDStr))
		jsonx.RenderInternalServerError(w)

		return
	}

	presignedURL, err := h.fileSvc.GeneratePresignedURL(ctx, f, presignedURLExpiry)
	if err != nil {
		h.logger.ErrorCtx(ctx, "cannot generate file URL", log.Error(err), log.String("file_id", fileIDStr))
		jsonx.RenderInternalServerError(w)

		return
	}

	http.Redirect(w, r, presignedURL, http.StatusTemporaryRedirect)
}
