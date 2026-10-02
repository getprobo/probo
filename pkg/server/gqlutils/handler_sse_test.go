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

package gqlutils_test

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"go.gearno.de/kit/log"

	"go.probo.inc/probo/pkg/server/gqlutils"
)

func TestHandler_SSEStreamsDeferredPayloads(t *testing.T) {
	t.Parallel()

	gate := make(chan struct{})
	server := httptest.NewServer(newStreamingHandler(t, gate))
	t.Cleanup(server.Close)

	req, err := http.NewRequest(http.MethodPost, server.URL, strings.NewReader(`{"query":"{ name }"}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/event-stream")
	assert.Equal(t, "no", resp.Header.Get("X-Accel-Buffering"))

	reader := bufio.NewReader(resp.Body)
	first, err := readUntil(reader, `"name":"ready"`)
	require.NoError(t, err)
	assert.Contains(t, first, "event: next")
	assert.NotContains(t, first, `"name":"deferred"`)

	close(gate)

	rest, err := io.ReadAll(reader)
	require.NoError(t, err)
	body := string(rest)
	assert.Contains(t, body, `"name":"deferred"`)
	assert.Contains(t, body, "event: complete")
}

func TestHandler_JSONPostStaysASingleResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(newStreamingHandler(t, nil))
	t.Cleanup(server.Close)

	req, err := http.NewRequest(http.MethodPost, server.URL, strings.NewReader(`{"query":"{ name }"}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")
	assert.JSONEq(t, `{"data":{"name":"ready"},"hasNext":true}`, string(body))
	assert.NotContains(t, string(body), "event:")
}

func newStreamingHandler(t *testing.T, gate chan struct{}) http.Handler {
	t.Helper()

	schema := gqlparser.MustLoadSchema(&ast.Source{Input: `
		type Query {
			name: String!
		}
	`})

	executable := &graphql.ExecutableSchemaMock{
		ComplexityFunc: func(context.Context, string, string, int, map[string]any) (int, bool) {
			return 0, true
		},
		SchemaFunc: func() *ast.Schema {
			return schema
		},
		ExecFunc: func(ctx context.Context) graphql.ResponseHandler {
			next := 0

			return func(ctx context.Context) *graphql.Response {
				next++
				switch next {
				case 1:
					hasNext := true

					return &graphql.Response{
						Data:    []byte(`{"name":"ready"}`),
						HasNext: &hasNext,
					}
				case 2:
					if gate != nil {
						select {
						case <-ctx.Done():
							return nil
						case <-gate:
						}
					}
					hasNext := false

					return &graphql.Response{
						Data:    []byte(`{"name":"deferred"}`),
						HasNext: &hasNext,
					}
				default:
					return nil
				}
			}
		},
	}

	return gqlutils.NewHandler(executable, log.NewLogger(), gqlutils.Limits{})
}

func readUntil(reader *bufio.Reader, needle string) (string, error) {
	var body strings.Builder
	buf := make([]byte, 64)
	for !strings.Contains(body.String(), needle) {
		n, err := reader.Read(buf)
		if n > 0 {
			body.Write(buf[:n])
		}
		if err != nil {
			return body.String(), err
		}
	}

	return body.String(), nil
}
