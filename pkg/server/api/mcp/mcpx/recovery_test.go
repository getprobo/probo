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

package mcpx_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/server/api/mcp/mcpx"
)

func TestNewRecoverFunc_RuntimePanicLogsStack(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	output := &bytes.Buffer{}
	logger := log.NewLogger(log.WithOutput(output))
	recoverFunc := mcpx.NewRecoverFunc(logger)

	var got error

	func() {
		defer func() {
			got = recoverFunc(ctx, recover())
		}()

		touchNilPointer()
	}()

	require.ErrorIs(t, got, mcpx.ErrInternal)

	entry := singleLogEntry(t, output.Bytes())
	assert.Equal(t, "internal error in MCP tool handler", entry["msg"])
	assert.Equal(t, "runtime error: invalid memory address or nil pointer dereference", entry["error"])
	stack, _ := entry["stack"].(string)
	assert.Contains(t, stack, "touchNilPointer")
	assert.NotContains(t, entry, "tool_name")
}

func TestNewRecoverFunc_DomainErrorStaysClassified(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	output := &bytes.Buffer{}
	logger := log.NewLogger(log.WithOutput(output))
	recoverFunc := mcpx.NewRecoverFunc(logger)

	got := recoverFunc(ctx, fmt.Errorf("cannot list cookie banners: %w", coredata.ErrResourceNotFound))

	require.ErrorIs(t, got, mcpx.ErrResourceNotFound)
	assert.Empty(t, output.String())
}

func TestNewRecoverFunc_UnknownErrorOmitsStack(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	output := &bytes.Buffer{}
	logger := log.NewLogger(log.WithOutput(output))
	recoverFunc := mcpx.NewRecoverFunc(logger)

	got := recoverFunc(ctx, errors.New("database is down"))

	require.ErrorIs(t, got, mcpx.ErrInternal)

	entry := singleLogEntry(t, output.Bytes())
	assert.Equal(t, "database is down", entry["error"])
	assert.NotContains(t, entry, "stack")
}

func TestLoggingMiddleware_NilParamsDoesNotPanic(t *testing.T) {
	t.Parallel()

	logger := log.NewLogger(log.WithOutput(&bytes.Buffer{}))
	called := false
	handler := mcpx.LoggingMiddleware(logger)(
		func(context.Context, string, mcp.Request) (mcp.Result, error) {
			called = true

			return nil, nil
		},
	)

	require.NotPanics(
		t,
		func() {
			_, err := handler(context.Background(), "tools/call", &mcp.CallToolRequest{})
			require.NoError(t, err)
		},
	)
	assert.True(t, called)
}

func TestLoggingMiddleware_RuntimePanicLogIncludesToolName(t *testing.T) {
	t.Parallel()

	output := &bytes.Buffer{}
	logger := log.NewLogger(log.WithOutput(output))
	recoverFunc := mcpx.NewRecoverFunc(logger)
	handler := mcpx.LoggingMiddleware(logger)(
		func(ctx context.Context, _ string, _ mcp.Request) (result mcp.Result, err error) {
			defer func() {
				if r := recover(); r != nil {
					err = recoverFunc(ctx, r)
				}
			}()

			touchNilPointer()

			return nil, nil
		},
	)

	_, err := handler(
		context.Background(),
		"tools/call",
		&mcp.CallToolRequest{
			Params: &mcp.CallToolParamsRaw{Name: "getDocument"},
		},
	)
	require.ErrorIs(t, err, mcpx.ErrInternal)

	var panicEntry map[string]any

	for line := range bytes.SplitSeq(bytes.TrimSpace(output.Bytes()), []byte("\n")) {
		if len(line) == 0 {
			continue
		}

		var entry map[string]any

		require.NoError(t, json.Unmarshal(line, &entry))

		if entry["msg"] == "internal error in MCP tool handler" {
			panicEntry = entry
		}
	}

	require.NotNil(t, panicEntry)
	assert.Equal(t, "getDocument", panicEntry["tool_name"])
	stack, _ := panicEntry["stack"].(string)
	assert.Contains(t, stack, "touchNilPointer")
}

func touchNilPointer() {
	var n *int

	_ = *n
}

func singleLogEntry(t *testing.T, raw []byte) map[string]any {
	t.Helper()

	var entry map[string]any
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(raw), &entry))

	return entry
}
