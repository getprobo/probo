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

package console_v1

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/getprobo/probo/pkg/crypto/cipher"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/probo"
	"github.com/getprobo/probo/pkg/securecookie"
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

	TrustCenterTokenData struct {
		TrustCenterID gid.GID      `json:"trust_center_id"`
		Email         string       `json:"email"`
		TenantID      gid.TenantID `json:"tenant_id"`
		Scope         string       `json:"scope"`
		ExpiresAt     time.Time    `json:"expires_at"`
	}
)

const (
	TokenCookieName = "trust_center_token"
)

func authTokenHandler(proboSvc *probo.Service, trustSvc *trust.Service, authCfg AuthConfig, encryptionKey cipher.EncryptionKey) http.HandlerFunc {
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

		accessData, err := validateTrustCenterAccessToken(r.Context(), proboSvc, authCfg, req.Token)
		if err != nil {
			httpserver.RenderJSON(w, http.StatusUnauthorized, AuthTokenResponse{
				Success: false,
				Message: "Invalid or expired token",
			})
			return
		}

		tokenData := TrustCenterTokenData{
			TrustCenterID: accessData.TrustCenterID,
			Email:         accessData.Email,
			TenantID:      accessData.TrustCenterID.TenantID(),
			Scope:         TokenScopeTrustCenterReadOnly,
			ExpiresAt:     time.Now().Add(24 * time.Hour),
		}

		tokenBytes, err := json.Marshal(tokenData)
		if err != nil {
			httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("failed to serialize token data: %w", err))
			return
		}

		encryptedTokenData, err := cipher.Encrypt(tokenBytes, encryptionKey)
		if err != nil {
			httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("failed to encrypt token data: %w", err))
			return
		}

		encryptedTokenString := base64.StdEncoding.EncodeToString(encryptedTokenData)

		cookieConfig := securecookie.Config{
			Name:     TokenCookieName,
			Secret:   authCfg.CookieSecret,
			Domain:   authCfg.CookieDomain,
			Path:     "/",
			MaxAge:   int(24 * time.Hour / time.Second), // 24 hours
			Secure:   true,
			HTTPOnly: true,
			SameSite: http.SameSiteStrictMode,
		}

		if err := securecookie.Set(w, cookieConfig, encryptedTokenString); err != nil {
			httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("failed to set cookie: %w", err))
			return
		}

		httpserver.RenderJSON(w, http.StatusOK, AuthTokenResponse{
			Success:       true,
			TrustCenterID: accessData.TrustCenterID.String(),
			Message:       "Authentication successful",
		})
	}
}

func validateTrustCenterAccessToken(ctx context.Context, proboSvc *probo.Service, authCfg AuthConfig, tokenString string) (*probo.TrustCenterAccessData, error) {
	token, err := statelesstoken.ValidateToken[probo.TrustCenterAccessData](
		authCfg.CookieSecret,
		probo.TokenTypeTrustCenterAccess,
		tokenString,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot validate trust center access token: %w", err)
	}

	tenantID := token.Data.TrustCenterID.TenantID()
	tenantSvc := proboSvc.WithTenant(tenantID)

	return tenantSvc.TrustCenterAccesses.ValidateToken(ctx, tokenString)
}

func trustCenterLogoutHandler(authCfg AuthConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookieConfig := securecookie.Config{
			Name:     TokenCookieName,
			Secret:   authCfg.CookieSecret,
			Domain:   authCfg.CookieDomain,
			Path:     "/",
			MaxAge:   -1,
			Secure:   true,
			HTTPOnly: true,
			SameSite: http.SameSiteStrictMode,
		}

		securecookie.Clear(w, cookieConfig)

		w.Header().Set("Clear-Site-Data", "*")

		httpserver.RenderJSON(w, http.StatusOK, map[string]bool{"success": true})
	}
}
