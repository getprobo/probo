// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
