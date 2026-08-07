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
	"go.probo.inc/probo/pkg/automerge/internal/native"
)

type (
	oracleRequest struct {
		Action    string `json:"action"`
		Actor     string `json:"actor,omitempty"`
		ActorB    string `json:"actorB,omitempty"`
		ActorC    string `json:"actorC,omitempty"`
		Document  string `json:"document,omitempty"`
		Message   string `json:"message,omitempty"`
		Text      string `json:"text,omitempty"`
		Timestamp int64  `json:"timestamp,omitempty"`
	}

	oracleResponse struct {
		Body     string   `json:"body"`
		Change   string   `json:"change"`
		Changes  []string `json:"changes"`
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

func TestConformance_GoReadsJavaScriptRichTextSpans(t *testing.T) {
	t.Parallel()

	actorID := actor(11)
	response := runOracle(
		t,
		oracleRequest{
			Action: "createRichText",
			Actor:  hex.EncodeToString(actorID[:]),
		},
	)

	data, err := base64.StdEncoding.DecodeString(response.Document)
	require.NoError(t, err)
	document, err := automerge.Load(context.Background(), data, actor(12))
	require.NoError(t, err)
	closeDocument(t, document)

	text, err := document.Text(context.Background(), "body")
	require.NoError(t, err)
	spans, err := text.Spans(context.Background())
	require.NoError(t, err)
	require.Len(t, spans, 2)
	assert.Equal(t, automerge.SpanTypeBlock, spans[0].Type)
	assert.Equal(t, "heading", spans[0].Block["type"])
	attrs, ok := spans[0].Block["attrs"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, float64(2), attrs["level"])
	assert.Equal(t, automerge.SpanTypeText, spans[1].Type)
	assert.Equal(t, "Policy", spans[1].Text)
	assert.Equal(t, true, spans[1].Marks["strong"])
}

func TestConformance_NativeParsesJavaScriptChange(t *testing.T) {
	t.Parallel()

	actorID := actor(13)
	response := runOracle(
		t,
		oracleRequest{
			Action:    "createChange",
			Actor:     hex.EncodeToString(actorID[:]),
			Message:   "Create policy",
			Timestamp: commitTime.Unix(),
		},
	)

	data, err := base64.StdEncoding.DecodeString(response.Change)
	require.NoError(t, err)
	change, err := native.ParseChange(data)
	require.NoError(t, err)
	assert.Equal(t, actorID[:], change.Actor)
	assert.Equal(t, uint64(1), change.Sequence)
	assert.Equal(t, uint64(1), change.StartOp)
	assert.Equal(t, commitTime.Unix(), change.Timestamp)
	assert.Equal(t, "Create policy", change.Message)
	assert.Empty(t, change.Dependencies)
	assert.NotEmpty(t, change.Columns)
	assert.Equal(t, response.Heads[0], hex.EncodeToString(change.Hash[:]))

	operations, err := change.Operations()
	require.NoError(t, err)
	require.Len(t, operations, 7)
	assert.Equal(t, native.ActionMakeText, operations[0].Action)
	assert.True(t, operations[0].Object.Root)
	assert.True(t, operations[0].Key.Map)
	assert.Equal(t, "title", operations[0].Key.Property)
	assert.Equal(t, uint64(1), operations[0].ID.Counter)
	assert.Equal(t, uint64(0), operations[0].ID.ActorIndex)
	assert.Equal(t, native.ActionSet, operations[1].Action)
	assert.Equal(t, uint64(1), operations[1].Object.OpID.Counter)
	assert.True(t, operations[1].Key.Head)
	assert.Equal(t, "P", operations[1].Value.Value)
	assert.Equal(t, uint64(6), operations[6].Key.Element.Counter)
	assert.Equal(t, "y", operations[6].Value.Value)

	state := native.NewState()
	require.NoError(t, state.ApplyChange(change))
	title, err := state.Text("title")
	require.NoError(t, err)
	assert.Equal(t, "Policy", title)
}

func TestConformance_NativeConcurrentChangesConverge(t *testing.T) {
	t.Parallel()

	actorA := actor(20)
	actorB := actor(21)
	actorC := actor(22)
	response := runOracle(
		t,
		oracleRequest{
			Action: "createConcurrentChanges",
			Actor:  hex.EncodeToString(actorA[:]),
			ActorB: hex.EncodeToString(actorB[:]),
			ActorC: hex.EncodeToString(actorC[:]),
		},
	)
	require.Len(t, response.Changes, 3)

	changes := make([]*native.Change, len(response.Changes))
	for i, encoded := range response.Changes {
		data, err := base64.StdEncoding.DecodeString(encoded)
		require.NoError(t, err)
		changes[i], err = native.ParseChange(data)
		require.NoError(t, err)
	}

	leftFirst := native.NewState()
	require.NoError(t, leftFirst.ApplyChange(changes[0]))
	require.NoError(t, leftFirst.ApplyChange(changes[1]))
	require.NoError(t, leftFirst.ApplyChange(changes[2]))

	rightFirst := native.NewState()
	require.NoError(t, rightFirst.ApplyChange(changes[0]))
	require.NoError(t, rightFirst.ApplyChange(changes[2]))
	require.NoError(t, rightFirst.ApplyChange(changes[1]))

	leftText, err := leftFirst.Text("body")
	require.NoError(t, err)
	rightText, err := rightFirst.Text("body")
	require.NoError(t, err)
	assert.Equal(t, response.Body, leftText)
	assert.Equal(t, response.Body, rightText)

	leftHeads := leftFirst.Heads()
	rightHeads := rightFirst.Heads()
	require.Len(t, leftHeads, 2)
	require.Len(t, rightHeads, 2)
	assert.ElementsMatch(
		t,
		response.Heads,
		[]string{
			hex.EncodeToString(leftHeads[0][:]),
			hex.EncodeToString(leftHeads[1][:]),
		},
	)
	assert.Equal(t, leftHeads, rightHeads)
}
