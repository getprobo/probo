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
	"encoding/json"
	"net/http"
)

type (
	oauth2Error struct {
		Code        string `json:"error"`
		Description string `json:"error_description,omitempty"`
		statusCode  int
	}
)

func (e *oauth2Error) writeResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(e.statusCode)
	_ = json.NewEncoder(w).Encode(e)
}

func writeOAuth2Error(w http.ResponseWriter, code, description string, statusCode int) {
	e := &oauth2Error{
		Code:        code,
		Description: description,
		statusCode:  statusCode,
	}
	e.writeResponse(w)
}

func writeOAuth2InvalidRequest(w http.ResponseWriter, description string) {
	writeOAuth2Error(w, "invalid_request", description, http.StatusBadRequest)
}

func writeOAuth2InvalidGrant(w http.ResponseWriter, description string) {
	writeOAuth2Error(w, "invalid_grant", description, http.StatusBadRequest)
}

func writeOAuth2InvalidClient(w http.ResponseWriter) {
	writeOAuth2Error(w, "invalid_client", "", http.StatusUnauthorized)
}

func writeOAuth2UnsupportedGrantType(w http.ResponseWriter) {
	writeOAuth2Error(w, "unsupported_grant_type", "", http.StatusBadRequest)
}

func writeOAuth2ServerError(w http.ResponseWriter, description string) {
	writeOAuth2Error(w, "server_error", description, http.StatusInternalServerError)
}
