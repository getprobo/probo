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

package domainconnect_v1

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/saferedirect"
	"go.probo.inc/probo/pkg/statelesstoken"
)

const TokenType = "probo/domain-connect"

type State struct {
	OrganizationID string `json:"oid"`
	ContinueURL    string `json:"continue,omitempty"`
}

func NewMux(
	logger *log.Logger,
	baseURL *baseurl.BaseURL,
	tokenSecret string,
) *chi.Mux {
	r := chi.NewMux()

	safeRedirect := saferedirect.New(saferedirect.StaticHosts(baseURL.Host()))

	r.Get(
		"/callback",
		handleCallback(
			logger,
			baseURL,
			tokenSecret,
			safeRedirect,
		),
	)

	return r
}

func handleCallback(
	logger *log.Logger,
	baseURL *baseurl.BaseURL,
	tokenSecret string,
	safeRedirect *saferedirect.SafeRedirect,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		stateToken := query.Get("state")
		if stateToken == "" {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("missing state parameter"))
			return
		}

		payload, err := statelesstoken.ValidateToken[State](
			tokenSecret,
			TokenType,
			stateToken,
		)
		if err != nil {
			logger.WarnCtx(r.Context(), "invalid Domain Connect state token", log.Error(err))
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("invalid or expired state"))
			return
		}

		redirectURL := payload.Data.ContinueURL
		if redirectURL == "" {
			redirectURL = baseURL.WithPath("/organizations/" + payload.Data.OrganizationID).MustString()
		}

		parsedURL, err := url.Parse(redirectURL)
		if err != nil {
			logger.ErrorCtx(r.Context(), "cannot parse redirect URL", log.Error(err))
			parsedURL, _ = url.Parse(baseURL.String())
		}

		q := parsedURL.Query()

		if dcErr := query.Get("error"); dcErr != "" {
			q.Set("domain_connect", "error")
			if desc := query.Get("error_description"); desc != "" {
				q.Set("error_description", desc)
			}
		} else {
			q.Set("domain_connect", "success")
		}

		parsedURL.RawQuery = q.Encode()

		safeRedirect.Redirect(w, r, parsedURL.String(), "/", http.StatusSeeOther)
	}
}
