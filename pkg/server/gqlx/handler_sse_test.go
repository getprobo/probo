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

package gqlx_test

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/server/gqlx"
)

func TestHandler_SSEStreamsDeferredPayloads(t *testing.T) {
	t.Parallel()

	gate := make(chan struct{})
	server := httptest.NewServer(newStreamingHandler(gate))
	t.Cleanup(server.Close)

	req, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		server.URL,
		strings.NewReader(`{"query":"{ name }"}`),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream, application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
	assert.Equal(t, "no", resp.Header.Get("X-Accel-Buffering"))

	reader := bufio.NewReader(resp.Body)

	first := readUntil(t, reader, `"ready"`)
	assert.Contains(t, first, "event: next")
	assert.NotContains(t, first, `"deferred"`)

	close(gate)

	rest := readUntil(t, reader, "event: complete")
	assert.Contains(t, rest, `"deferred"`)
}

func TestHandler_JSONPostStaysASingleResponse(t *testing.T) {
	t.Parallel()

	gate := make(chan struct{})
	close(gate)

	server := httptest.NewServer(newStreamingHandler(gate))
	t.Cleanup(server.Close)

	req, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		server.URL,
		strings.NewReader(`{"query":"{ name }"}`),
	)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.NotContains(t, string(body), "event:")
	assert.JSONEq(t, `{"data":{"name":"ready"},"hasNext":true}`, string(body))
}

func readUntil(t *testing.T, reader *bufio.Reader, marker string) string {
	t.Helper()

	type result struct {
		text string
		err  error
	}

	done := make(chan result, 1)

	go func() {
		var b strings.Builder
		for {
			line, err := reader.ReadString('\n')
			b.WriteString(line)
			if strings.Contains(b.String(), marker) || err != nil {
				done <- result{text: b.String(), err: err}
				return
			}
		}
	}()

	select {
	case r := <-done:
		require.Contains(t, r.text, marker, "stream ended early: %v", r.err)
		return r.text
	case <-time.After(5 * time.Second):
		require.FailNow(t, "timed out waiting for "+marker)
		return ""
	}
}

func newStreamingHandler(gate <-chan struct{}) http.Handler {
	schema := gqlparser.MustLoadSchema(
		&ast.Source{Input: "type Query { name: String! }"},
	)

	executable := &graphql.ExecutableSchemaMock{
		ComplexityFunc: func(context.Context, string, string, int, map[string]any) (int, bool) {
			return 0, true
		},
		SchemaFunc: func() *ast.Schema { return schema },
		ExecFunc: func(ctx context.Context) graphql.ResponseHandler {
			step := 0

			return func(ctx context.Context) *graphql.Response {
				step++

				switch step {
				case 1:
					hasNext := true
					return &graphql.Response{
						Data:    json.RawMessage(`{"name":"ready"}`),
						HasNext: &hasNext,
					}
				case 2:
					select {
					case <-gate:
					case <-ctx.Done():
						return nil
					}

					hasNext := false
					return &graphql.Response{
						Data:    json.RawMessage(`{"name":"deferred"}`),
						HasNext: &hasNext,
					}
				default:
					return nil
				}
			}
		},
	}

	return gqlx.NewHandler(executable, log.NewLogger(log.WithOutput(io.Discard)), gqlx.Limits{})
}
