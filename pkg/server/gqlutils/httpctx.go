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

package gqlutils

import (
	"context"
	"net/http"
)

type (
	ctxKey struct{ name string }
)

var (
	httpResponseWriterKey = &ctxKey{name: "http_response_writer"}
	httpRequestKey        = &ctxKey{name: "http_request"}
)

func WithHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {

	ctx = context.WithValue(ctx, httpResponseWriterKey, w)
	ctx = context.WithValue(ctx, httpRequestKey, r)

	return ctx
}

func HTTPResponseWriterFromContext(ctx context.Context) http.ResponseWriter {
	return ctx.Value(httpResponseWriterKey).(http.ResponseWriter)
}

func HTTPRequestFromContext(ctx context.Context) *http.Request {
	return ctx.Value(httpRequestKey).(*http.Request)
}
