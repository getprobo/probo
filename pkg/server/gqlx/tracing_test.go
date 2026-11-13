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
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/server/gqlx"
)

func TestInterceptOperation(t *testing.T) {
	t.Parallel()

	extension := gqlx.NewTracingExtension(log.NewLogger(log.WithOutput(io.Discard)))

	t.Run(
		"nil response from a canceled operation",
		func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(operationContext())
			t.Cleanup(cancel)

			var logs bytes.Buffer

			extension := gqlx.NewTracingExtension(log.NewLogger(log.WithOutput(&logs)))

			handler := extension.InterceptOperation(
				ctx,
				func(context.Context) graphql.ResponseHandler {
					return func(ctx context.Context) *graphql.Response {
						if ctx.Err() != nil {
							return nil
						}

						return &graphql.Response{}
					}
				},
			)

			require.NotNil(t, handler(ctx))

			cancel()

			assert.Nil(t, handler(ctx))

			entry := canceledLogEntry(t, logs.String())
			assert.Equal(t, "INFO", entry["level"])
			assert.Contains(t, entry["msg"], "query Viewer canceled")
			assert.Equal(t, "Viewer", entry["graphql_operation_name"])
			assert.Equal(t, "query", entry["graphql_operation_type"])
			assert.Equal(t, "context canceled", entry["error"])
		},
	)

	t.Run(
		"returns the operation response",
		func(t *testing.T) {
			t.Parallel()

			ctx := operationContext()
			want := &graphql.Response{
				Errors: gqlerror.List{gqlerror.Errorf("boom")},
			}
			handler := extension.InterceptOperation(
				ctx,
				func(context.Context) graphql.ResponseHandler {
					return func(context.Context) *graphql.Response {
						return want
					}
				},
			)

			assert.Equal(t, want, handler(ctx))
		},
	)

	t.Run(
		"deferred stream logs once at the final payload",
		func(t *testing.T) {
			t.Parallel()

			ctx := operationContext()

			var logs bytes.Buffer

			extension := gqlx.NewTracingExtension(log.NewLogger(log.WithOutput(&logs)))

			hasNext, last := true, false
			responses := []*graphql.Response{
				{HasNext: &hasNext},
				{HasNext: &hasNext, Errors: gqlerror.List{
					&gqlerror.Error{
						Message: "submitted value: api-key-secret",
						Extensions: map[string]any{
							"code":  "INVALID",
							"value": "api-key-secret",
						},
					},
				}},
				{HasNext: &last},
				nil,
			}

			handler := extension.InterceptOperation(
				ctx,
				func(context.Context) graphql.ResponseHandler {
					return func(context.Context) *graphql.Response {
						resp := responses[0]
						responses = responses[1:]

						return resp
					}
				},
			)

			require.NotNil(t, handler(ctx))
			require.NotNil(t, handler(ctx))
			assert.Empty(t, logs.String())

			require.NotNil(t, handler(ctx))
			assert.Nil(t, handler(ctx))

			lines := strings.Split(strings.TrimSpace(logs.String()), "\n")
			require.Len(t, lines, 1)

			var entry map[string]any
			require.NoError(t, json.Unmarshal([]byte(lines[0]), &entry))
			assert.Equal(t, "INFO", entry["level"])
			assert.Contains(t, entry["msg"], "query Viewer failed")
			assert.Equal(t, float64(1), entry["graphql_operation_error_count"])
			assert.Equal(t, "INVALID", entry["graphql_operation_error_codes"])
			assert.NotContains(t, lines[0], "api-key-secret")
		},
	)

	t.Run(
		"logs client errors at info and internal errors at error",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name  string
				errs  gqlerror.List
				level string
				codes string
			}{
				{
					name: "not found",
					errs: gqlerror.List{
						&gqlerror.Error{Extensions: map[string]any{"code": "NOT_FOUND"}},
					},
					level: "INFO",
				},
				{
					name: "forbidden",
					errs: gqlerror.List{
						&gqlerror.Error{Extensions: map[string]any{"code": "FORBIDDEN"}},
					},
					level: "INFO",
				},
				{
					name: "unauthenticated",
					errs: gqlerror.List{
						&gqlerror.Error{Extensions: map[string]any{"code": "UNAUTHENTICATED"}},
					},
					level: "INFO",
				},
				{
					name: "invalid",
					errs: gqlerror.List{
						&gqlerror.Error{Extensions: map[string]any{"code": "INVALID"}},
					},
					level: "INFO",
				},
				{
					name: "internal",
					errs: gqlerror.List{
						&gqlerror.Error{Extensions: map[string]any{"code": "INTERNAL"}},
					},
					level: "ERROR",
				},
				{
					name:  "no code",
					errs:  gqlerror.List{gqlerror.Errorf("boom")},
					level: "ERROR",
				},
				{
					name: "invalid and internal",
					errs: gqlerror.List{
						&gqlerror.Error{Extensions: map[string]any{"code": "INVALID"}},
						&gqlerror.Error{Extensions: map[string]any{"code": "INTERNAL"}},
					},
					level: "ERROR",
				},
				{
					name: "parse failed",
					errs: gqlerror.List{
						&gqlerror.Error{Extensions: map[string]any{"code": "GRAPHQL_PARSE_FAILED"}},
					},
					level: "INFO",
					codes: "GRAPHQL_PARSE_FAILED",
				},
				{
					name: "validation failed",
					errs: gqlerror.List{
						&gqlerror.Error{Extensions: map[string]any{"code": "GRAPHQL_VALIDATION_FAILED"}},
					},
					level: "INFO",
					codes: "GRAPHQL_VALIDATION_FAILED",
				},
				{
					name: "invalid credentials",
					errs: gqlerror.List{
						&gqlerror.Error{Extensions: map[string]any{"code": "INVALID_CREDENTIALS"}},
					},
					level: "INFO",
					codes: "INVALID_CREDENTIALS",
				},
				{
					name: "email not verified",
					errs: gqlerror.List{
						&gqlerror.Error{Extensions: map[string]any{"code": "EMAIL_NOT_VERIFIED"}},
					},
					level: "INFO",
					codes: "EMAIL_NOT_VERIFIED",
				},
				{
					name: "complexity limit",
					errs: gqlerror.List{
						&gqlerror.Error{Extensions: map[string]any{"code": "COMPLEXITY_LIMIT_EXCEEDED"}},
					},
					level: "INFO",
					codes: "COMPLEXITY_LIMIT_EXCEEDED",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ctx := operationContext()

						var logs bytes.Buffer

						extension := gqlx.NewTracingExtension(log.NewLogger(log.WithOutput(&logs)))

						handler := extension.InterceptOperation(
							ctx,
							func(context.Context) graphql.ResponseHandler {
								return func(context.Context) *graphql.Response {
									return &graphql.Response{Errors: tt.errs}
								}
							},
						)

						require.NotNil(t, handler(ctx))

						var entry map[string]any
						require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(logs.String())), &entry))
						assert.Equal(t, tt.level, entry["level"])
						assert.Contains(t, entry["msg"], "query Viewer failed")

						if tt.codes != "" {
							assert.Equal(t, tt.codes, entry["graphql_operation_error_codes"])
						}
					},
				)
			}
		},
	)

	t.Run(
		"uses the document name when the request omits it",
		func(t *testing.T) {
			t.Parallel()

			ctx := graphql.WithOperationContext(
				context.Background(),
				&graphql.OperationContext{
					Operation: &ast.OperationDefinition{
						Operation: ast.Query,
						Name:      "AssetList",
					},
				},
			)

			entry := failedLogEntry(t, ctx)
			assert.Contains(t, entry["msg"], "query AssetList failed")
			assert.Equal(t, "AssetList", entry["graphql_operation_name"])
		},
	)

	t.Run(
		"anonymous operation stays unnamed",
		func(t *testing.T) {
			t.Parallel()

			ctx := graphql.WithOperationContext(
				context.Background(),
				&graphql.OperationContext{
					Operation: &ast.OperationDefinition{
						Operation: ast.Query,
					},
				},
			)

			entry := failedLogEntry(t, ctx)
			assert.Contains(t, entry["msg"], "query unnamed failed")
			assert.Equal(t, "unnamed", entry["graphql_operation_name"])
		},
	)
}

func failedLogEntry(t *testing.T, ctx context.Context) map[string]any {
	t.Helper()

	var logs bytes.Buffer

	extension := gqlx.NewTracingExtension(log.NewLogger(log.WithOutput(&logs)))
	handler := extension.InterceptOperation(
		ctx,
		func(context.Context) graphql.ResponseHandler {
			return func(context.Context) *graphql.Response {
				return &graphql.Response{
					Errors: gqlerror.List{gqlerror.Errorf("boom")},
				}
			}
		},
	)

	require.NotNil(t, handler(ctx))

	var entry map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(logs.String())), &entry))

	return entry
}

func canceledLogEntry(t *testing.T, logs string) map[string]any {
	t.Helper()

	var canceled map[string]any

	for line := range strings.SplitSeq(strings.TrimSpace(logs), "\n") {
		var entry map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &entry))

		msg, _ := entry["msg"].(string)
		if strings.Contains(msg, "canceled") {
			canceled = entry
		}
	}

	require.NotNil(t, canceled)

	return canceled
}

func operationContext() context.Context {
	return graphql.WithOperationContext(
		context.Background(),
		&graphql.OperationContext{
			OperationName: "Viewer",
			Operation: &ast.OperationDefinition{
				Operation: ast.Query,
			},
		},
	)
}
