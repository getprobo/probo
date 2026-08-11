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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/automerge/internal/native"
	automergeprosemirror "go.probo.inc/probo/pkg/automerge/prosemirror"
)

type (
	oracleRequest struct {
		Action    string          `json:"action"`
		Actor     string          `json:"actor,omitempty"`
		ActorB    string          `json:"actorB,omitempty"`
		ActorC    string          `json:"actorC,omitempty"`
		Change    string          `json:"change,omitempty"`
		Document  string          `json:"document,omitempty"`
		Message   string          `json:"message,omitempty"`
		Scenario  json.RawMessage `json:"scenario,omitempty"`
		Text      string          `json:"text,omitempty"`
		Timestamp int64           `json:"timestamp,omitempty"`
	}

	oracleResponse struct {
		Body     string   `json:"body"`
		Change   string   `json:"change"`
		Changes  []string `json:"changes"`
		Data     any      `json:"data"`
		Document string   `json:"document"`
		Heads    []string `json:"heads"`
		Message  string   `json:"message"`
		Sync     string   `json:"sync"`
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

func TestConformance_DatesFlowBetweenDocuments(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	sourceActor := actor(20)
	created := runOracle(
		t,
		oracleRequest{
			Action:    "createTimestamps",
			Actor:     hex.EncodeToString(sourceActor[:]),
			Message:   "dates",
			Timestamp: commitTime.Unix(),
		},
	)

	createdData, ok := created.Data.(map[string]any)
	require.True(t, ok)
	iso, ok := createdData["iso"].(string)
	require.True(t, ok)

	source, err := base64.StdEncoding.DecodeString(created.Document)
	require.NoError(t, err)
	sourceDocument, err := automerge.Load(ctx, source, actor(21))
	require.NoError(t, err)
	closeDocument(t, sourceDocument)

	when, err := sourceDocument.Root().Scalar(ctx, "when")
	require.NoError(t, err)
	require.Equal(t, automerge.ScalarTypeTimestamp, when.Type)

	list, err := sourceDocument.Root().Object(ctx, "list")
	require.NoError(t, err)
	listWhen, err := list.ScalarAt(ctx, 0)
	require.NoError(t, err)
	require.Equal(t, automerge.ScalarTypeTimestamp, listWhen.Type)
	require.Equal(t, when.Int, listWhen.Int)

	// Reuse the timestamps read from the source document in a new document.
	target, err := automerge.New(ctx, actor(22))
	require.NoError(t, err)
	closeDocument(t, target)
	require.NoError(t, target.Root().PutScalar(ctx, "when", when))
	targetList, err := target.Root().CreateObject(ctx, "list", automerge.ObjectTypeList)
	require.NoError(t, err)
	require.NoError(t, targetList.InsertScalar(ctx, 0, listWhen))
	_, err = target.Commit(ctx, "reuse dates", commitTime.Add(time.Second))
	require.NoError(t, err)
	saved, err := target.Save(ctx)
	require.NoError(t, err)

	read := runOracle(
		t,
		oracleRequest{
			Action:   "readTimestamps",
			Document: base64.StdEncoding.EncodeToString(saved),
		},
	)

	readData, ok := read.Data.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, true, readData["whenIsDate"])
	assert.Equal(t, iso, readData["whenISO"])
	assert.Equal(t, true, readData["listIsDate"])
	assert.Equal(t, iso, readData["listISO"])
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

func TestConformance_JavaScriptPreservesGoChanges(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	document, err := automerge.New(ctx, actor(2))
	require.NoError(t, err)
	closeDocument(t, document)
	text, err := document.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "ABC"))
	hash, err := document.Commit(ctx, "Create in Go", commitTime)
	require.NoError(t, err)
	changes, err := document.ChangesSince(ctx, nil)
	require.NoError(t, err)
	require.Len(t, changes, 1)

	data, err := document.Save(ctx)
	require.NoError(t, err)

	response := runOracle(
		t,
		oracleRequest{
			Action:   "inspectChanges",
			Document: base64.StdEncoding.EncodeToString(data),
		},
	)

	require.Equal(t, []string{hash.String()}, response.Heads)
	require.Len(t, response.Changes, 1)
	forwarded, err := base64.StdEncoding.DecodeString(response.Changes[0])
	require.NoError(t, err)
	assert.Equal(t, changes[0].Bytes, forwarded)
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

	nativeDocument, err := native.Decode(data)
	require.NoError(t, err)
	nativeState, err := native.NewStateFromDocument(nativeDocument)
	require.NoError(t, err)
	nativeText, err := nativeState.Text("body")
	require.NoError(t, err)
	assert.Equal(t, "Hello from JavaScript 😀", nativeText)
}

func TestConformance_NativePreservesJavaScriptDataModel(t *testing.T) {
	t.Parallel()

	actorID := actor(15)
	created := runOracle(
		t,
		oracleRequest{
			Action: "createDataModel",
			Actor:  hex.EncodeToString(actorID[:]),
		},
	)
	data, err := base64.StdEncoding.DecodeString(created.Document)
	require.NoError(t, err)

	document, err := automerge.Load(context.Background(), data, actor(16))
	require.NoError(t, err)
	closeDocument(t, document)
	saved, err := document.Save(context.Background())
	require.NoError(t, err)

	inspected := runOracle(
		t,
		oracleRequest{
			Action:   "inspectDataModel",
			Document: base64.StdEncoding.EncodeToString(saved),
		},
	)
	assert.Equal(t, created.Data, inspected.Data)
	assert.Equal(t, created.Heads, inspected.Heads)
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
	document, err := automerge.LoadReference(context.Background(), data, actor(12))
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

	nativeDocument, err := automerge.LoadPureGo(context.Background(), data, actor(14))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	nativeText, err := nativeDocument.Text(context.Background(), "body")
	require.NoError(t, err)
	nativeSpans, err := nativeText.Spans(context.Background())
	require.NoError(t, err)
	assert.Equal(t, spans, nativeSpans)
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
	decoded, err := native.Decode(data)
	require.NoError(t, err)
	require.Len(t, decoded.Changes, 1)
	change := &decoded.Changes[0]
	assert.Equal(t, actorID[:], change.Actor.Bytes())
	assert.Equal(t, uint64(1), change.Sequence)
	assert.Equal(t, uint64(1), change.StartOp)
	assert.Equal(t, commitTime.Unix(), change.Time)
	assert.Equal(t, "Create policy", change.Message)
	assert.Empty(t, change.Dependencies)
	require.NotNil(t, change.Hash)
	assert.Equal(t, response.Heads[0], change.Hash.String())

	operations := change.Operations
	require.Len(t, operations, 7)
	assert.Equal(t, native.ActionMakeText, operations[0].Action)
	assert.True(t, operations[0].Object.IsRoot)
	require.NotNil(t, operations[0].Key.Property)
	assert.Equal(t, "title", *operations[0].Key.Property)
	assert.Equal(t, uint64(1), operations[0].ID.Counter)
	assert.Equal(t, native.ActorID(string(actorID[:])), operations[0].ID.Actor)
	assert.Equal(t, native.ActionSet, operations[1].Action)
	assert.Equal(t, uint64(1), operations[1].Object.OpID.Counter)
	assert.True(t, operations[1].Key.IsHead)
	require.NotNil(t, operations[1].Value)
	assert.Equal(t, "P", operations[1].Value.String)
	require.NotNil(t, operations[6].Key.Element)
	assert.Equal(t, uint64(6), operations[6].Key.Element.Counter)
	require.NotNil(t, operations[6].Value)
	assert.Equal(t, "y", operations[6].Value.String)

	state := native.NewState()
	require.NoError(t, state.ApplyChange(change))
	title, err := state.Text("title")
	require.NoError(t, err)
	assert.Equal(t, "Policy", title)

	nativeEncoded, err := native.EncodeChange(change)
	require.NoError(t, err)
	inspection := runOracle(
		t,
		oracleRequest{
			Action: "inspectChange",
			Change: base64.StdEncoding.EncodeToString(nativeEncoded),
		},
	)
	assert.Equal(t, "Policy", inspection.Body)
	assert.Equal(t, "Create policy", inspection.Message)
	require.NotNil(t, change.Hash)
	assert.Equal(t, inspection.Heads[0], change.Hash.String())
}

func TestConformance_NativeParsesJavaScriptEmptyChange(t *testing.T) {
	t.Parallel()

	actorID := actor(46)
	response := runOracle(
		t,
		oracleRequest{
			Action:    "createEmptyChange",
			Actor:     hex.EncodeToString(actorID[:]),
			Message:   "Empty",
			Timestamp: 12_345,
		},
	)
	data, err := base64.StdEncoding.DecodeString(response.Change)
	require.NoError(t, err)
	decoded, err := native.Decode(data)
	require.NoError(t, err)
	require.Len(t, decoded.Changes, 1)
	change := decoded.Changes[0]
	assert.Equal(t, uint64(1), change.Sequence)
	assert.Equal(t, uint64(1), change.StartOp)
	assert.Equal(t, uint64(0), change.MaxOp)
	assert.Equal(t, int64(12_345), change.Time)
	assert.Equal(t, "Empty", change.Message)
	assert.Empty(t, change.Operations)
	require.NotNil(t, change.Hash)
	assert.Equal(t, response.Heads[0], change.Hash.String())
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

	var (
		combined   []byte
		rawChanges [][]byte
	)

	for _, encoded := range response.Changes {
		data, err := base64.StdEncoding.DecodeString(encoded)
		require.NoError(t, err)

		combined = append(combined, data...)
		rawChanges = append(rawChanges, data)
	}

	decoded, err := native.Decode(combined)
	require.NoError(t, err)
	require.Len(t, decoded.Changes, 3)
	changes := []*native.Change{
		&decoded.Changes[0],
		&decoded.Changes[1],
		&decoded.Changes[2],
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

	backend, err := native.LoadEngine(context.Background(), rawChanges[0])
	require.NoError(t, err)
	_, err = backend.Merge(context.Background(), rawChanges[1])
	require.NoError(t, err)

	before, err := backend.Heads(context.Background())
	require.NoError(t, err)

	saved, err := backend.Save(context.Background(), true, true)
	require.NoError(t, err)

	// Save now writes a compacted document rather than embedding change bytes, so
	// the guarantee is that reloading it reproduces the same frontier.
	reloaded, err := native.LoadEngine(context.Background(), saved)
	require.NoError(t, err)

	after, err := reloaded.Heads(context.Background())
	require.NoError(t, err)
	assert.Equal(t, before, after)
}

func TestConformance_NativeSyncMessageRoundTrip(t *testing.T) {
	t.Parallel()

	actorID := actor(30)
	response := runOracle(
		t,
		oracleRequest{
			Action: "createSyncMessage",
			Actor:  hex.EncodeToString(actorID[:]),
		},
	)
	data, err := base64.StdEncoding.DecodeString(response.Sync)
	require.NoError(t, err)

	message, err := native.ParseSyncMessage(data)
	require.NoError(t, err)
	assert.Contains(
		t,
		[]native.SyncMessageVersion{
			native.SyncMessageVersion1,
			native.SyncMessageVersion2,
		},
		message.Version,
	)
	require.Len(t, message.Heads, 1)
	assert.Equal(t, response.Heads[0], hex.EncodeToString(message.Heads[0][:]))

	encoded, err := message.Encode()
	require.NoError(t, err)
	assert.Equal(t, data, encoded)
}

func TestConformance_NativeComplexRichTextSpans(t *testing.T) {
	t.Parallel()

	actorID := actor(31)
	response := runOracle(
		t,
		oracleRequest{
			Action: "createComplexRichText",
			Actor:  hex.EncodeToString(actorID[:]),
		},
	)
	data, err := base64.StdEncoding.DecodeString(response.Document)
	require.NoError(t, err)

	referenceDocument, err := automerge.LoadReference(
		context.Background(),
		data,
		actor(32),
	)
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	referenceText, err := referenceDocument.Text(context.Background(), "body")
	require.NoError(t, err)
	referenceSpans, err := referenceText.Spans(context.Background())
	require.NoError(t, err)

	nativeDocument, err := automerge.Load(context.Background(), data, actor(33))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	nativeText, err := nativeDocument.Text(context.Background(), "body")
	require.NoError(t, err)
	nativeSpans, err := nativeText.Spans(context.Background())
	require.NoError(t, err)

	assert.Equal(t, referenceSpans, nativeSpans)
}

func TestConformance_NativeBoundaryMarks(t *testing.T) {
	t.Parallel()

	actorID := actor(37)
	response := runOracle(
		t,
		oracleRequest{
			Action: "createBoundaryMarks",
			Actor:  hex.EncodeToString(actorID[:]),
		},
	)
	data, err := base64.StdEncoding.DecodeString(response.Document)
	require.NoError(t, err)

	referenceDocument, err := automerge.LoadReference(
		context.Background(),
		data,
		actor(38),
	)
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	referenceText, err := referenceDocument.Text(context.Background(), "body")
	require.NoError(t, err)
	referenceSpans, err := referenceText.Spans(context.Background())
	require.NoError(t, err)

	nativeDocument, err := automerge.Load(
		context.Background(),
		data,
		actor(39),
	)
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	nativeText, err := nativeDocument.Text(context.Background(), "body")
	require.NoError(t, err)
	nativeSpans, err := nativeText.Spans(context.Background())
	require.NoError(t, err)

	assert.Equal(t, referenceSpans, nativeSpans)
}

func TestConformance_NativeSplitMarks(t *testing.T) {
	t.Parallel()

	actorID := actor(40)
	response := runOracle(
		t,
		oracleRequest{
			Action: "createSplitMarks",
			Actor:  hex.EncodeToString(actorID[:]),
		},
	)
	data, err := base64.StdEncoding.DecodeString(response.Document)
	require.NoError(t, err)

	referenceDocument, err := automerge.LoadReference(
		context.Background(),
		data,
		actor(41),
	)
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	referenceText, err := referenceDocument.Text(context.Background(), "body")
	require.NoError(t, err)
	referenceSpans, err := referenceText.Spans(context.Background())
	require.NoError(t, err)

	nativeDocument, err := automerge.Load(
		context.Background(),
		data,
		actor(42),
	)
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	nativeText, err := nativeDocument.Text(context.Background(), "body")
	require.NoError(t, err)
	nativeSpans, err := nativeText.Spans(context.Background())
	require.NoError(t, err)

	assert.Equal(t, referenceSpans, nativeSpans)
}

func TestConformance_NativeUnicodeMarks(t *testing.T) {
	t.Parallel()

	actorID := actor(43)
	response := runOracle(
		t,
		oracleRequest{
			Action: "createUnicodeMarks",
			Actor:  hex.EncodeToString(actorID[:]),
		},
	)
	data, err := base64.StdEncoding.DecodeString(response.Document)
	require.NoError(t, err)

	referenceDocument, err := automerge.LoadReference(
		context.Background(),
		data,
		actor(44),
	)
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	referenceText, err := referenceDocument.Text(context.Background(), "body")
	require.NoError(t, err)
	referenceSpans, err := referenceText.Spans(context.Background())
	require.NoError(t, err)

	nativeDocument, err := automerge.Load(
		context.Background(),
		data,
		actor(45),
	)
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	nativeText, err := nativeDocument.Text(context.Background(), "body")
	require.NoError(t, err)
	nativeSpans, err := nativeText.Spans(context.Background())
	require.NoError(t, err)

	assert.Equal(t, referenceSpans, nativeSpans)
}

func TestConformance_NativeTableRichTextSpans(t *testing.T) {
	t.Parallel()

	actorID := actor(34)
	response := runOracle(
		t,
		oracleRequest{
			Action: "createTableRichText",
			Actor:  hex.EncodeToString(actorID[:]),
		},
	)
	data, err := base64.StdEncoding.DecodeString(response.Document)
	require.NoError(t, err)

	referenceDocument, err := automerge.LoadReference(
		context.Background(),
		data,
		actor(35),
	)
	require.NoError(t, err)
	closeDocument(t, referenceDocument)
	referenceText, err := referenceDocument.Text(context.Background(), "body")
	require.NoError(t, err)
	referenceSpans, err := referenceText.Spans(context.Background())
	require.NoError(t, err)

	nativeDocument, err := automerge.Load(context.Background(), data, actor(36))
	require.NoError(t, err)
	closeDocument(t, nativeDocument)
	nativeText, err := nativeDocument.Text(context.Background(), "body")
	require.NoError(t, err)
	nativeSpans, err := nativeText.Spans(context.Background())
	require.NoError(t, err)
	assert.Equal(t, referenceSpans, nativeSpans)

	content, err := automergeprosemirror.Render(nativeSpans)
	require.NoError(t, err)

	var document struct {
		Content []struct {
			Type    string `json:"type"`
			Content []struct {
				Type    string            `json:"type"`
				Content []json.RawMessage `json:"content"`
			} `json:"content"`
		} `json:"content"`
	}
	require.NoError(t, json.Unmarshal([]byte(content), &document))
	require.Len(t, document.Content, 1)
	assert.Equal(t, "table", document.Content[0].Type)
	require.Len(t, document.Content[0].Content, 2)
	assert.Equal(t, "tableRow", document.Content[0].Content[0].Type)
	assert.Len(t, document.Content[0].Content[0].Content, 2)
}
