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

package mcputils_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/server/api/mcp/mcputils"
)

func TestNewSlogLogger_PreservesRecord(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	logger := log.NewLogger(
		log.WithLevel(log.LevelDebug),
		log.WithName("mcp.transport"),
		log.WithOutput(&output),
	)
	slogLogger := mcputils.NewSlogLogger(logger).
		With("component", "streamable").
		WithGroup("request").
		With("method", "POST").
		WithGroup("details")

	slogLogger.ErrorContext(context.Background(), "transport failed", "retryable", true)

	var entry map[string]any
	err := json.Unmarshal(output.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "ERROR", entry["level"])
	assert.Equal(t, "transport failed", entry["msg"])
	assert.Equal(t, "mcp.transport", entry["name"])
	assert.Equal(t, "streamable", entry["component"])
	assert.Equal(
		t,
		map[string]any{
			"method": "POST",
			"details": map[string]any{
				"retryable": true,
			},
		},
		entry["request"],
	)
}
