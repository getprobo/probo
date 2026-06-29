// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package gqlutils

import (
	"context"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/vektah/gqlparser/v2/ast"
	"go.gearno.de/kit/log"
)

type (
	Handler struct {
		gqlhandler *handler.Server
	}

	// Limits bounds the per-request cost of a GraphQL operation to protect
	// against alias-flooding and other application-layer denial-of-service
	// vectors. A zero value disables the corresponding guard.
	Limits struct {
		ParserTokenLimit  int
		ComplexityLimit   int
		QueryCacheSize    int
		DisableSuggestion bool
	}
)

var (
	mb int64 = 1024 * 1024

	postTransport      = transport.POST{}
	optionsTransport   = transport.Options{}
	multipartTransport = transport.MultipartForm{
		MaxMemory:     32 * mb,
		MaxUploadSize: 50 * mb,
	}

	introspectionExtension = extension.Introspection{}
)

func NewHandler[S graphql.ExecutableSchema](executableSchema S, logger *log.Logger, limits Limits) *Handler {
	handler := handler.New(executableSchema)

	handler.AddTransport(postTransport)
	handler.AddTransport(optionsTransport)
	handler.AddTransport(multipartTransport)

	if limits.QueryCacheSize > 0 {
		handler.SetQueryCache(lru.New[*ast.QueryDocument](limits.QueryCacheSize))
	}

	if limits.ParserTokenLimit > 0 {
		handler.SetParserTokenLimit(limits.ParserTokenLimit)
	}

	if limits.ComplexityLimit > 0 {
		handler.Use(extension.FixedComplexityLimit(limits.ComplexityLimit))
	}

	handler.SetDisableSuggestion(limits.DisableSuggestion)

	handler.Use(introspectionExtension)
	handler.Use(NewTracingExtension(logger))

	handler.SetRecoverFunc(RecoverFunc)

	return &Handler{gqlhandler: handler}
}

func (gqlh *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := WithHTTPContext(r.Context(), w, r)

	gqlh.gqlhandler.ServeHTTP(w, r.WithContext(ctx))
}

func (gqlh *Handler) Use(extension graphql.HandlerExtension) {
	gqlh.gqlhandler.Use(extension)
}

func (gqlh *Handler) AroundOperations(f func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler) {
	gqlh.gqlhandler.AroundOperations(f)
}
