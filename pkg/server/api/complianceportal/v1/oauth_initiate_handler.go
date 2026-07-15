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
	"net/http"

	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/complianceportal/visitor"
	"go.probo.inc/probo/pkg/iam/oauth2"
	"go.probo.inc/probo/pkg/saferedirect"
	"go.probo.inc/probo/pkg/server/api/complianceportal"
)

type OAuthInitiateHandler struct {
	proboBaseURL *baseurl.BaseURL
	visitor      *visitor.Service
	safeRedirect *saferedirect.SafeRedirect
	httpClient   *http.Client
	logger       *log.Logger
}

func NewOAuthInitiateHandler(
	proboBaseURL *baseurl.BaseURL,
	visitorSvc *visitor.Service,
	allowedHost saferedirect.AllowedHostFunc,
	logger *log.Logger,
) *OAuthInitiateHandler {
	return &OAuthInitiateHandler{
		proboBaseURL: proboBaseURL,
		visitor:      visitorSvc,
		safeRedirect: saferedirect.New(allowedHost),
		httpClient: httpclient.DefaultClient(
			httpclient.WithLogger(logger),
			httpclient.WithSSRFProtection(),
		),
		logger: logger,
	}
}

func (h *OAuthInitiateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	portalBaseURL := complianceportal.CompliancePageBaseURLFromContext(ctx)
	if portalBaseURL == nil {
		httpserver.RenderError(w, http.StatusNotFound, errNotFound)
		return
	}

	continueURL := r.URL.Query().Get("continue")
	if continueURL == "" {
		continueURL = "/overview"
	}

	safeContinue, ok := h.safeRedirect.Validate(ctx, continueURL)
	if !ok {
		httpserver.RenderError(w, http.StatusBadRequest, errInvalidContinueURL)
		return
	}

	clientID, err := complianceportal.CIMDClientIDURL(*portalBaseURL)
	if err != nil {
		h.logger.ErrorCtx(ctx, "cannot build cimd client_id", log.Error(err))
		httpserver.RenderError(w, http.StatusInternalServerError, errInternal)

		return
	}

	redirectURI, err := complianceportal.OAuthCallbackURL(*portalBaseURL)
	if err != nil {
		h.logger.ErrorCtx(ctx, "cannot build oauth redirect_uri", log.Error(err))
		httpserver.RenderError(w, http.StatusInternalServerError, errInternal)

		return
	}

	metadata, err := oauth2.FetchServerMetadata(ctx, h.httpClient, h.proboBaseURL.String())
	if err != nil {
		h.logger.ErrorCtx(ctx, "cannot fetch discovery metadata", log.Error(err))
		httpserver.RenderError(w, http.StatusInternalServerError, errInternal)

		return
	}

	authorizeURL, err := h.visitor.InitiateOAuthAuthorizeURL(
		ctx,
		metadata.AuthorizationEndpoint.String(),
		clientID,
		redirectURI,
		[]string{complianceportal.VisitorOAuthScope},
		safeContinue,
	)
	if err != nil {
		h.logger.ErrorCtx(ctx, "cannot initiate oauth authorize", log.Error(err))
		httpserver.RenderError(w, http.StatusInternalServerError, errInternal)

		return
	}

	http.Redirect(w, r, authorizeURL, http.StatusFound)
}
