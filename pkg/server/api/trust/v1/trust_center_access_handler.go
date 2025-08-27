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

package trust_v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/getprobo/probo/pkg/probo"
	console_v1 "github.com/getprobo/probo/pkg/server/api/console/v1"
	"github.com/getprobo/probo/pkg/server/session"
	"github.com/getprobo/probo/pkg/statelesstoken"
	"github.com/getprobo/probo/pkg/trust"
	"go.gearno.de/kit/httpserver"
)

type (
	AuthTokenRequest struct {
		Token string `json:"token"`
	}

	AuthTokenResponse struct {
		Success       bool   `json:"success"`
		TrustCenterID string `json:"trust_center_id,omitempty"`
		Message       string `json:"message,omitempty"`
	}
)

func authTokenHandler(trustSvc *trust.Service, trustAuthCfg TrustAuthConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AuthTokenRequest
		// Limit request body size to 1KB to prevent DoS attacks
		limitedReader := http.MaxBytesReader(w, r.Body, 1024)
		if err := json.NewDecoder(limitedReader).Decode(&req); err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot decode body: %w", err))
			return
		}

		if req.Token == "" {
			httpserver.RenderJSON(w, http.StatusBadRequest, AuthTokenResponse{
				Success: false,
				Message: "Token is required",
			})
			return
		}

		accessData, err := validateTrustCenterAccessToken(r.Context(), trustSvc, trustAuthCfg, req.Token)
		if err != nil {
			httpserver.RenderJSON(w, http.StatusUnauthorized, AuthTokenResponse{
				Success: false,
				Message: "Invalid or expired token",
			})
			return
		}

		tokenString, err := statelesstoken.NewToken(
			trustAuthCfg.TokenSecret,
			trustAuthCfg.TokenType,
			trustAuthCfg.TokenDuration,
			*accessData,
		)
		if err != nil {
			httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("failed to create token: %w", err))
			return
		}

		cookie := &http.Cookie{
			Name:     trustAuthCfg.CookieName,
			Value:    tokenString,
			Domain:   trustAuthCfg.CookieDomain,
			Path:     "/",
			MaxAge:   int(trustAuthCfg.CookieDuration / time.Second),
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, cookie)

		httpserver.RenderJSON(w, http.StatusOK, AuthTokenResponse{
			Success:       true,
			TrustCenterID: accessData.TrustCenterID.String(),
			Message:       "Authentication successful",
		})
	}
}

func validateTrustCenterAccessToken(ctx context.Context, trustSvc *trust.Service, trustAuthCfg TrustAuthConfig, tokenString string) (*probo.TrustCenterAccessData, error) {
	token, err := statelesstoken.ValidateToken[probo.TrustCenterAccessData](
		trustSvc.GetTokenSecret(),
		trustAuthCfg.TokenType,
		tokenString,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot validate trust center access token: %w", err)
	}

	tenantID := token.Data.TrustCenterID.TenantID()
	tenantSvc := trustSvc.WithTenant(tenantID)

	accessData, err := tenantSvc.TrustCenterAccesses.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, fmt.Errorf("cannot validate trust center access token: %w", err)
	}

	return accessData, nil
}

func trustCenterLogoutHandler(authCfg console_v1.AuthConfig, trustAuthCfg TrustAuthConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     trustAuthCfg.CookieName,
			Value:    "",
			Domain:   trustAuthCfg.CookieDomain,
			Path:     "/",
			MaxAge:   -1,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})

		session.ClearCookie(w, session.AuthConfig{
			CookieName:   authCfg.CookieName,
			CookieSecret: authCfg.CookieSecret,
		})

		httpserver.RenderJSON(w, http.StatusOK, map[string]string{
			"message": "Logged out successfully",
		})
	}
}
