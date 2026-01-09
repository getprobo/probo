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

package mcp_v1

import (
	"net/http"

	"errors"

	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/server/api/authn"
)

func RequireAPIKeyHandler(
	logger *log.Logger,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		correlationID := r.Header.Get("X-Request-ID")
		if correlationID == "" {
			correlationID = r.Header.Get("X-Correlation-ID")
		}

		logger.InfoCtx(ctx, "MCP authentication attempt",
			log.String("correlation_id", correlationID),
			log.String("path", r.URL.Path),
		)

		apiKey := authn.APIKeyFromContext(ctx)
		identity := authn.IdentityFromContext(ctx)
		if identity == nil {
			httpserver.RenderError(w, http.StatusUnauthorized, errors.New("authentication required"))
			return
		}

		logger.InfoCtx(ctx, "MCP authentication successful",
			log.String("correlation_id", correlationID),
			log.String("identity_id", identity.ID.String()),
			log.String("api_key_id", apiKey.ID.String()),
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
