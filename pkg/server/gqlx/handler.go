// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package gqlx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
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

const (
	sseKeepAlivePingInterval = 15 * time.Second
	maxSSERequestBytes       = 1 * 1024 * 1024
)

// eventStreamTransport delivers @defer payloads one SSE event at a time.
// It only answers requests that accept text/event-stream; everyone else
// stays on the single-response POST transport.
type eventStreamTransport struct {
	transport.SSE

	parserTokenLimit int
}

func (t eventStreamTransport) Do(w http.ResponseWriter, r *http.Request, exec graphql.GraphExecutor) {
	// SSE flushes headers before resolvers run, so a Set-Cookie written by a
	// mutation such as signIn would never reach the client.
	isMutation, err := t.isMutation(w, r)
	if err != nil {
		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)

		return
	}

	if isMutation {
		postTransport.Do(w, r, exec)
		return
	}

	// Reverse proxies buffer the body until the handler returns unless told
	// otherwise, which would hold the first payload until the last probe.
	w.Header().Set("X-Accel-Buffering", "no")
	t.SSE.Do(w, r, exec)
}

// isMutation peeks at the body and restores it, so whichever transport runs
// next reads it in full. Anything it cannot parse goes to SSE, which reports
// the error itself.
func (t eventStreamTransport) isMutation(w http.ResponseWriter, r *http.Request) (bool, error) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxSSERequestBytes))
	r.Body = io.NopCloser(bytes.NewReader(body))

	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			return false, err
		}

		return false, nil
	}

	var params struct {
		Query         string `json:"query"`
		OperationName string `json:"operationName"`
	}

	if err := json.Unmarshal(body, &params); err != nil {
		return false, nil
	}

	document, err := parser.ParseQueryWithTokenLimit(
		&ast.Source{Input: params.Query},
		t.parserTokenLimit,
	)
	if err != nil {
		return false, nil
	}

	operation := document.Operations.ForName(params.OperationName)

	return operation != nil && operation.Operation == ast.Mutation, nil
}

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

	// The first transport whose Supports matches wins, and SSE and POST both
	// accept application/json bodies, so SSE must be registered first.
	handler.AddTransport(
		eventStreamTransport{
			KeepAlivePingInterval: sseKeepAlivePingInterval,
			parserTokenLimit:      limits.ParserTokenLimit,
		},
	)
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

	ctx, operations := withOpenOperations(ctx)
	defer operations.close()

	gqlh.gqlhandler.ServeHTTP(w, r.WithContext(ctx))
}

func (gqlh *Handler) Use(extension graphql.HandlerExtension) {
	gqlh.gqlhandler.Use(extension)
}

func (gqlh *Handler) AroundOperations(f func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler) {
	gqlh.gqlhandler.AroundOperations(f)
}
