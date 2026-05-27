// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package mcp_v1

import (
	"errors"
	"fmt"
	"net/http"

	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/iam/oauth2server"
	"go.probo.inc/probo/pkg/server/api/authn"
)

func RequireIdentityHandler(
	logger *log.Logger,
	protectedResource oauth2server.ProtectedResourceConfig,
	next http.Handler,
) http.Handler {
	wwwAuthenticate, err := oauth2server.WWWAuthenticateHeader(protectedResource.Resource)
	if err != nil {
		panic(fmt.Errorf("cannot build MCP WWW-Authenticate header: %w", err))
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		correlationID := r.Header.Get("X-Request-ID")
		if correlationID == "" {
			correlationID = r.Header.Get("X-Correlation-ID")
		}

		logger.InfoCtx(
			ctx,
			"MCP authentication attempt",
			log.String("correlation_id", correlationID),
			log.String("path", r.URL.Path),
		)

		identity := authn.IdentityFromContext(ctx)
		if identity == nil {
			w.Header().Set("WWW-Authenticate", wwwAuthenticate)
			httpserver.RenderError(w, http.StatusUnauthorized, errors.New("authentication required"))

			return
		}

		if apiKey := authn.APIKeyFromContext(ctx); apiKey != nil {
			logger.InfoCtx(
				ctx,
				"MCP authentication successful",
				log.String("correlation_id", correlationID),
				log.String("identity_id", identity.ID.String()),
				log.String("api_key_id", apiKey.ID.String()),
			)
		} else {
			logger.InfoCtx(
				ctx,
				"MCP authentication successful",
				log.String("correlation_id", correlationID),
				log.String("identity_id", identity.ID.String()),
			)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
