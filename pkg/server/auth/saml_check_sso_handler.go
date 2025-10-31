// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"

	authsvc "go.probo.inc/probo/pkg/auth"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
)

type (
	CheckSSORequest struct {
		Email string `json:"email"`
	}

	CheckSSOResponse struct {
		SSOAvailable     bool    `json:"ssoAvailable"`
		SAMLConfigID     *string `json:"samlConfigId,omitempty"`
		OrganizationID   *string `json:"organizationId,omitempty"`
		EnforcementPolicy *string `json:"enforcementPolicy,omitempty"`
	}
)

func SAMLCheckSSOHandler(authSvc *authsvc.Service, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req CheckSSORequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot decode body: %w", err))
			return
		}

		if req.Email == "" {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("email is required"))
			return
		}

		if _, err := mail.ParseAddress(req.Email); err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("invalid email format"))
			return
		}

		configs, err := authSvc.CheckSSOAvailabilityByEmail(ctx, req.Email)
		if err != nil {
			logger.ErrorCtx(ctx, "cannot check SSO availability", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("cannot check SSO availability"))
			return
		}

		// No SAML configs found for this domain
		if len(configs) == 0 {
			httpserver.RenderJSON(w, http.StatusOK, CheckSSOResponse{
				SSOAvailable: false,
			})
			return
		}

		// Multiple SAML configs found - ambiguous, user must use organization-specific SSO URL
		if len(configs) > 1 {
			logger.WarnCtx(ctx, "multiple SAML configurations found for domain", log.Int("count", len(configs)))
			httpserver.RenderError(w, http.StatusConflict, fmt.Errorf("multiple SSO configurations found for this domain. Please use your organization-specific SSO login URL"))
			return
		}

		// Single SAML config found - return it
		config := configs[0]
		configIDStr := config.ID.String()
		orgIDStr := config.OrganizationID.String()
		enforcementPolicy := string(config.EnforcementPolicy)

		httpserver.RenderJSON(w, http.StatusOK, CheckSSOResponse{
			SSOAvailable:      true,
			SAMLConfigID:      &configIDStr,
			OrganizationID:    &orgIDStr,
			EnforcementPolicy: &enforcementPolicy,
		})
	}
}
