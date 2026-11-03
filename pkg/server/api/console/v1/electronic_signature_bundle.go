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
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/bearertoken"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/jsonx"
)

func handleElectronicSignatureBundle(
	logger *log.Logger,
	iamSvc *iam.Service,
	esignSvc *esign.Service,
	baseURL *baseurl.BaseURL,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		signatureID, err := gid.ParseGID(chi.URLParam(r, "signatureID"))
		if err != nil {
			jsonx.RenderNotFound(w, fmt.Errorf("signature bundle not found"))
			return
		}

		ctx := r.Context()
		if authn.IdentityFromContext(ctx) == nil {
			jsonx.RenderUnauthorized(w, fmt.Errorf("authentication required"))
			return
		}

		scope, ok := authorizeBundleResource(
			w,
			r,
			iamSvc,
			baseURL,
			signatureID,
			probo.ActionElectronicSignatureGet,
		)
		if !ok {
			return
		}

		signature, err := esignSvc.GetSignatureByID(ctx, scope, signatureID)
		if err != nil {
			if errors.Is(err, esign.ErrElectronicSignatureNotFound) {
				jsonx.RenderNotFound(w, fmt.Errorf("signature bundle not found"))
				return
			}

			logger.ErrorCtx(
				ctx,
				"cannot load electronic signature",
				log.Error(err),
				log.String("signature_id", signatureID.String()),
			)
			jsonx.RenderInternalServerError(w)

			return
		}

		if signature.CertificateFileID == nil {
			jsonx.RenderNotFound(w, fmt.Errorf("signature bundle not found"))
			return
		}

		if _, ok := authorizeBundleResource(
			w,
			r,
			iamSvc,
			baseURL,
			signature.FileID,
			probo.ActionFileGet,
		); !ok {
			return
		}

		if _, ok := authorizeBundleResource(
			w,
			r,
			iamSvc,
			baseURL,
			*signature.CertificateFileID,
			probo.ActionFileGet,
		); !ok {
			return
		}

		var body bytes.Buffer

		filename, err := esignSvc.WriteBundle(ctx, signatureID, &body)
		if err != nil {
			if errors.Is(err, esign.ErrElectronicSignatureNotFound) ||
				errors.Is(err, esign.ErrBundleNotReady) {
				jsonx.RenderNotFound(w, fmt.Errorf("signature bundle not found"))
				return
			}

			logger.ErrorCtx(
				ctx,
				"cannot build electronic signature bundle",
				log.Error(err),
				log.String("signature_id", signatureID.String()),
			)
			jsonx.RenderInternalServerError(w)

			return
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", bundleContentDisposition(filename))
		w.Header().Set("Content-Length", strconv.Itoa(body.Len()))
		w.Header().Set("Cache-Control", "private, no-store")
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write(body.Bytes()); err != nil {
			logger.ErrorCtx(
				ctx,
				"cannot write electronic signature bundle",
				log.Error(err),
				log.String("signature_id", signatureID.String()),
			)
		}
	}
}

func authorizeBundleResource(
	w http.ResponseWriter,
	r *http.Request,
	iamSvc *iam.Service,
	baseURL *baseurl.BaseURL,
	resource gid.GID,
	action string,
) (*coredata.Scope, bool) {
	ctx := r.Context()
	identity := authn.IdentityFromContext(ctx)
	session := authn.SessionFromContext(ctx)

	params := iam.AuthorizeParams{
		Principal:          identity.ID,
		Resource:           resource,
		Action:             action,
		ResourceAttributes: make(map[string]string),
	}
	if session != nil {
		params.Session = &session.ID
	}

	scope, err := iamSvc.Authorizer.Authorize(ctx, params)
	if err != nil {
		writeBundleAuthError(w, baseURL, err)
		return nil, false
	}

	return scope, true
}

func writeBundleAuthError(
	w http.ResponseWriter,
	baseURL *baseurl.BaseURL,
	err error,
) {
	if scopeErr, ok := errors.AsType[*iam.ErrInsufficientOAuth2Scope](err); ok {
		bearertoken.SetBearerInsufficientScope(w, baseURL, scopeErr.Scopes...)
		jsonx.RenderForbidden(w)

		return
	}

	if _, ok := errors.AsType[*iam.ErrInsufficientPermissions](err); ok {
		jsonx.RenderForbidden(w)
		return
	}

	jsonx.RenderNotFound(w, fmt.Errorf("signature bundle not found"))
}

func bundleContentDisposition(filename string) string {
	return fmt.Sprintf(
		"attachment; filename=%q; filename*=UTF-8''%s",
		asciiFilename(filename),
		url.PathEscape(filename),
	)
}

func asciiFilename(filename string) string {
	var b strings.Builder
	b.Grow(len(filename))

	for _, r := range filename {
		if r < 0x20 || r > 0x7e {
			b.WriteByte('_')
			continue
		}

		b.WriteRune(r)
	}

	return b.String()
}
