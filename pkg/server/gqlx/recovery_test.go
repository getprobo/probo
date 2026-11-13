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

package gqlx_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/server/gqlx"
)

func TestRecoverFunc_Panic(t *testing.T) {
	t.Parallel()

	// The request logger lives on the HTTP server context, which is the
	// context RecoverFunc logs with in production.
	var recovered error

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recovered = gqlx.RecoverFunc(r.Context(), "boom")
		w.WriteHeader(http.StatusNoContent)
	})

	server := httpserver.NewServer(
		"127.0.0.1:0",
		handler,
		httpserver.WithLogger(log.NewLogger(log.WithOutput(io.Discard))),
		httpserver.WithRegisterer(prometheus.NewRegistry()),
	)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	rec := httptest.NewRecorder()
	server.Handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)

	gqlErr, ok := errors.AsType[*gqlerror.Error](recovered)
	require.True(t, ok)
	assert.Equal(t, "INTERNAL", gqlErr.Extensions["code"])
}
