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

package automerge_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

type (
	oracleRequest struct {
		Action    string `json:"action"`
		Actor     string `json:"actor,omitempty"`
		Document  string `json:"document,omitempty"`
		Message   string `json:"message,omitempty"`
		Text      string `json:"text,omitempty"`
		Timestamp int64  `json:"timestamp,omitempty"`
	}

	oracleResponse struct {
		Body     string   `json:"body"`
		Document string   `json:"document"`
		Heads    []string `json:"heads"`
	}
)

func runOracle(t *testing.T, request oracleRequest) oracleResponse {
	t.Helper()

	oracle := os.Getenv("AUTOMERGE_JS_ORACLE")
	if oracle == "" {
		t.Skip("AUTOMERGE_JS_ORACLE is not configured")
	}

	input, err := json.Marshal(request)
	require.NoError(t, err)

	command := exec.CommandContext(context.Background(), "node", oracle)
	command.Stdin = bytes.NewReader(input)
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))

	var response oracleResponse
	require.NoError(t, json.Unmarshal(output, &response))
	return response
}

func TestConformance_JavaScriptLoadsGoDocument(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	document, err := automerge.New(ctx, actor(1))
	require.NoError(t, err)
	closeDocument(t, document)

	text, err := document.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "Hello 😀"))
	hash, err := document.Commit(ctx, "Create in Go", commitTime)
	require.NoError(t, err)
	data, err := document.Save(ctx)
	require.NoError(t, err)

	response := runOracle(
		t,
		oracleRequest{
			Action:   "inspect",
			Document: base64.StdEncoding.EncodeToString(data),
		},
	)

	assert.Equal(t, "Hello 😀", response.Body)
	assert.Equal(t, []string{hash.String()}, response.Heads)
}

func TestConformance_GoLoadsJavaScriptDocument(t *testing.T) {
	t.Parallel()

	actorID := actor(9)
	response := runOracle(
		t,
		oracleRequest{
			Action:    "create",
			Actor:     hex.EncodeToString(actorID[:]),
			Message:   "Create in JavaScript",
			Text:      "Hello from JavaScript 😀",
			Timestamp: commitTime.Unix(),
		},
	)

	data, err := base64.StdEncoding.DecodeString(response.Document)
	require.NoError(t, err)
	document, err := automerge.Load(context.Background(), data, actor(10))
	require.NoError(t, err)
	closeDocument(t, document)

	text, err := document.Text(context.Background(), "body")
	require.NoError(t, err)
	value, err := text.String(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Hello from JavaScript 😀", value)

	heads, err := document.Heads(context.Background())
	require.NoError(t, err)
	require.Len(t, heads, 1)
	assert.Equal(t, response.Heads[0], heads[0].String())
}
