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
	"strings"

	"go.gearno.de/kit/httpserver"
	"go.probo.inc/probo/pkg/iam/oauth2server"
)

// oauth2RedirectContext wraps an OAuth2 error with redirect context for the
// authorization endpoint. When handled, redirectable errors are sent back
// to the client via query parameters instead of rendered as JSON.
type oauth2RedirectContext struct {
	err         error
	redirectURI string
	state       string
}

func (e *oauth2RedirectContext) Error() string { return e.err.Error() }
func (e *oauth2RedirectContext) Unwrap() error { return e.err }

// withRedirect wraps an error with redirect context so that writeOAuth2Error
// can redirect back to the client when appropriate.
func withRedirect(err error, redirectURI, state string) error {
	return &oauth2RedirectContext{
		err:         err,
		redirectURI: redirectURI,
		state:       state,
	}
}

func writeOAuth2Error(w http.ResponseWriter, r *http.Request, err error) {
	rc, ok := errors.AsType[*oauth2RedirectContext](err)
	if ok {
		if isRedirectableError(err) && rc.redirectURI != "" {
			redirectWithError(w, r, rc.redirectURI, rc.state, rc.err)
			return
		}

		err = rc.err
	}

	code := oauth2server.OAuth2ErrorCode(err)
	if code == "server_error" && !errors.Is(err, oauth2server.ErrServerError) {
		httpserver.RenderError(w, http.StatusInternalServerError, err)
		return
	}

	description := oauth2ErrorDescription(err, code)
	statusCode := oauth2ErrorStatusCode(err)

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	httpserver.RenderJSON(w, statusCode, &struct {
		Code        string `json:"error"`
		Description string `json:"error_description,omitempty"`
	}{
		Code:        code,
		Description: description,
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

func oauth2ErrorStatusCode(err error) int {
	switch {
	case errors.Is(err, oauth2server.ErrAccessDenied):
		return http.StatusForbidden
	case errors.Is(err, oauth2server.ErrInvalidClient):
		return http.StatusUnauthorized
	case errors.Is(err, oauth2server.ErrServerError):
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}

func oauth2ErrorDescription(err error, code string) string {
	msg := err.Error()

	if msg == code {
		return ""
	}

	prefix := code + ": "
	if strings.HasPrefix(msg, prefix) {
		return msg[len(prefix):]
	}

	return msg
}

func redirectWithError(w http.ResponseWriter, r *http.Request, redirectURI, state string, err error) {
	u, parseErr := url.Parse(redirectURI)
	if parseErr != nil {
		httpserver.RenderError(w, http.StatusInternalServerError, errors.New("internal server error"))
		return
	}

	code := oauth2server.OAuth2ErrorCode(err)
	description := oauth2ErrorDescription(err, code)

	q := u.Query()
	q.Set("error", code)
	if description != "" {
		q.Set("error_description", description)
	}
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()

	http.Redirect(w, r, u.String(), http.StatusFound)
}
