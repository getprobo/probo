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

package connect_v1

import (
	"errors"
	"net/http"
	"net/url"

	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/iam/oauth2server"
	"go.probo.inc/probo/pkg/server/api/connect/v1/types"
)

func (h *OAuth2Handler) handleAuthorizeError(w http.ResponseWriter, r *http.Request, err error, redirectURI, state string) {
	if isRedirectableError(err) && redirectURI != "" {
		redirectWithError(w, r, redirectURI, state, err)
		return
	}

	h.renderOAuth2ErrorResponse(w, r, err)
}

func (h *OAuth2Handler) renderOAuth2ErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	oauthErr, ok := errors.AsType[*oauth2server.OAuth2Error](err)
	if !ok {
		httpserver.RenderError(w, http.StatusInternalServerError, err)
		return
	}

	if errors.Is(err, oauth2server.ErrServerError) {
		h.logger.ErrorCtx(r.Context(), "oauth2 server error", log.Error(err))
	}

	NoCache(w)

	httpserver.RenderJSON(w, oauth2ErrorStatusCode(oauthErr), &types.OAuth2ErrorResponse{
		Code:        oauthErr.ErrorCode(),
		Description: oauthErr.Description(),
	})
}

func isRedirectableError(err error) bool {
	return errors.Is(err, oauth2server.ErrAccessDenied) ||
		errors.Is(err, oauth2server.ErrInvalidRequest) ||
		errors.Is(err, oauth2server.ErrInvalidScope) ||
		errors.Is(err, oauth2server.ErrUnauthorizedClient) ||
		errors.Is(err, oauth2server.ErrInvalidGrant) ||
		errors.Is(err, oauth2server.ErrUnsupportedGrantType)
}

func oauth2ErrorStatusCode(err *oauth2server.OAuth2Error) int {
	switch err.ErrorCode() {
	case "access_denied":
		return http.StatusForbidden
	case "invalid_client":
		return http.StatusUnauthorized
	case "server_error":
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}

func toOAuth2Error(err error) *oauth2server.OAuth2Error {
	switch {
	case errors.Is(err, oauth2server.ErrClientNotFound):
		return oauth2server.NewError(oauth2server.ErrInvalidClient, oauth2server.WithDescription("client not found"))
	case errors.Is(err, oauth2server.ErrInvalidRedirectURI):
		return oauth2server.ErrInvalidRedirectURI
	case errors.Is(err, oauth2server.ErrUnauthorizedMember):
		return oauth2server.NewError(oauth2server.ErrUnauthorizedClient, oauth2server.WithDescription("client is private and user is not a member of the organization"))
	case errors.Is(err, oauth2server.ErrDeviceCodeNotPending):
		return oauth2server.NewError(oauth2server.ErrInvalidGrant, oauth2server.WithDescription("device code is not pending"))
	default:
		if oauthErr, ok := errors.AsType[*oauth2server.OAuth2Error](err); ok {
			return oauthErr
		}
		return oauth2server.NewError(oauth2server.ErrServerError, oauth2server.WithDescription("internal error"))
	}
}

func redirectWithError(w http.ResponseWriter, r *http.Request, redirectURI, state string, err error) {
	u, parseErr := url.Parse(redirectURI)
	if parseErr != nil {
		httpserver.RenderError(w, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	oauthErr, ok := errors.AsType[*oauth2server.OAuth2Error](err)
	if !ok {
		httpserver.RenderError(w, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	q := u.Query()
	q.Set("error", oauthErr.ErrorCode())
	if desc := oauthErr.Description(); desc != "" {
		q.Set("error_description", desc)
	}
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()

	http.Redirect(w, r, u.String(), http.StatusFound)
}
