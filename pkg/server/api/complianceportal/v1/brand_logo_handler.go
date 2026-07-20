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

package complianceportal_v1

import (
	"errors"
	"net/http"

	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/complianceportal"
)

type brandLogoVariant int

const (
	brandLogoVariantLogo brandLogoVariant = iota
	brandLogoVariantDarkLogo
)

type brandLogoHandler struct {
	logger      *log.Logger
	fileManager *filemanager.Service
	variant     brandLogoVariant
}

func NewBrandLogoHandler(logger *log.Logger, fileManager *filemanager.Service) http.Handler {
	return &brandLogoHandler{
		logger:      logger,
		fileManager: fileManager,
		variant:     brandLogoVariantLogo,
	}
}

func NewBrandDarkLogoHandler(logger *log.Logger, fileManager *filemanager.Service) http.Handler {
	return &brandLogoHandler{
		logger:      logger,
		fileManager: fileManager,
		variant:     brandLogoVariantDarkLogo,
	}
}

func (h *brandLogoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	compliancePage := complianceportal.CompliancePortalFromContext(r.Context())
	if compliancePage == nil {
		httpserver.RenderError(w, http.StatusNotFound, errNotFound)
		return
	}

	var fileID *gid.GID

	switch h.variant {
	case brandLogoVariantLogo:
		fileID = compliancePage.LogoFileID
	case brandLogoVariantDarkLogo:
		fileID = compliancePage.DarkLogoFileID
	}

	if fileID == nil {
		httpserver.RenderError(w, http.StatusNotFound, errNotFound)
		return
	}

	file, err := h.fileManager.GetPublicFile(r.Context(), *fileID)
	if errors.Is(err, coredata.ErrResourceNotFound) {
		httpserver.RenderError(w, http.StatusNotFound, errNotFound)
		return
	}

	if err != nil {
		h.logger.ErrorCtx(r.Context(), "cannot load compliance page brand logo", log.Error(err))
		httpserver.RenderError(w, http.StatusInternalServerError, errInternal)

		return
	}

	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, h.fileManager.GenerateFileURL(file), http.StatusTemporaryRedirect)
}
