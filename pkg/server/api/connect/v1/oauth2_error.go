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
	"strings"

	"go.gearno.de/kit/httpserver"
	"go.probo.inc/probo/pkg/iam/oauth2server"
)

func writeOAuth2Error(w http.ResponseWriter, err error) {
	code := oauth2server.OAuth2ErrorCode(err)
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
