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
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

type (
	interopScenario struct {
		Name       string                     `json:"name"`
		Actor      string                     `json:"actor"`
		Operations []interopScenarioOperation `json:"operations"`
	}

	interopScenarioOperation struct {
		Action      string                `json:"action"`
		Path        []string              `json:"path"`
		Key         string                `json:"key"`
		ObjectType  automerge.ObjectType  `json:"objectType"`
		Scalar      interopScenarioScalar `json:"scalar"`
		Index       uint64                `json:"index"`
		DeleteCount int32                 `json:"deleteCount"`
		Text        string                `json:"text"`
		Delta       int64                 `json:"delta"`
		Message     string                `json:"message"`
		Timestamp   int64                 `json:"timestamp"`
	}

	interopScenarioScalar struct {
		Type      automerge.ScalarType `json:"type"`
		Bool      bool                 `json:"bool"`
		Uint      uint64               `json:"uint"`
		Int       int64                `json:"int"`
		FloatBits string               `json:"floatBits"`
		String    string               `json:"string"`
		Bytes     string               `json:"bytes"`
	}
)

func TestInteropScenario_CoreDataModel(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("testdata/scenarios/core-data-model.json")
	require.NoError(t, err)

	var scenario interopScenario
	require.NoError(t, json.Unmarshal(data, &scenario))
	actorBytes, err := hex.DecodeString(scenario.Actor)
	require.NoError(t, err)
	require.Len(t, actorBytes, 16)

	var actorID automerge.ActorID
	copy(actorID[:], actorBytes)

	ctx := context.Background()
	nativeDocument := runInteropScenario(
		t,
		ctx,
		scenario,
		func(ctx context.Context, actorID automerge.ActorID) (*automerge.Document, error) {
			return automerge.New(ctx, actorID)
		},
		actorID,
	)
	closeDocument(t, nativeDocument)
	referenceDocument := runInteropScenario(
		t,
		ctx,
		scenario,
		automerge.NewReference,
		actorID,
	)
	closeDocument(t, referenceDocument)

	nativeHeads, err := nativeDocument.Heads(ctx)
	require.NoError(t, err)
	referenceHeads, err := referenceDocument.Heads(ctx)
	require.NoError(t, err)
	nativeData, err := nativeDocument.Save(ctx)
	require.NoError(t, err)
	referenceData, err := referenceDocument.Save(ctx)
	require.NoError(t, err)
	assertInteropScenarioResult(t, ctx, nativeDocument)
	assertInteropScenarioResult(t, ctx, referenceDocument)

	nativeFromReference, err := automerge.Load(
		ctx,
		referenceData,
		actor(187),
	)
	require.NoError(t, err)
	closeDocument(t, nativeFromReference)
	assertInteropScenarioResult(t, ctx, nativeFromReference)
	referenceFromNative, err := automerge.LoadReference(
		ctx,
		nativeData,
		actor(188),
	)
	require.NoError(t, err)
	closeDocument(t, referenceFromNative)
	assertInteropScenarioResult(t, ctx, referenceFromNative)

	response := runOracle(
		t,
		oracleRequest{
			Action:   "runScenario",
			Scenario: data,
		},
	)
	nativeInspection := runOracle(
		t,
		oracleRequest{
			Action:   "inspectScenario",
			Document: base64.StdEncoding.EncodeToString(nativeData),
		},
	)
	assert.Equal(t, response.Data, nativeInspection.Data)
	assert.Equal(
		t,
		[]string{nativeHeads[0].String()},
		nativeInspection.Heads,
	)
	referenceInspection := runOracle(
		t,
		oracleRequest{
			Action:   "inspectScenario",
			Document: base64.StdEncoding.EncodeToString(referenceData),
		},
	)
	assert.Equal(t, response.Data, referenceInspection.Data)
	assert.Equal(
		t,
		[]string{referenceHeads[0].String()},
		referenceInspection.Heads,
	)

	javaScriptData, err := base64.StdEncoding.DecodeString(response.Document)
	require.NoError(t, err)
	javaScriptDocument, err := automerge.Load(ctx, javaScriptData, actor(186))
	require.NoError(t, err)
	closeDocument(t, javaScriptDocument)
	assertInteropScenarioResult(t, ctx, javaScriptDocument)
	javaScriptReference, err := automerge.LoadReference(
		ctx,
		javaScriptData,
		actor(189),
	)
	require.NoError(t, err)
	closeDocument(t, javaScriptReference)
	assertInteropScenarioResult(t, ctx, javaScriptReference)
	javaScriptHeads, err := javaScriptDocument.Heads(ctx)
	require.NoError(t, err)
	require.Equal(t, response.Heads, []string{javaScriptHeads[0].String()})
}

func runInteropScenario(
	t *testing.T,
	ctx context.Context,
	scenario interopScenario,
	factory func(
		context.Context,
		automerge.ActorID,
	) (*automerge.Document, error),
	actorID automerge.ActorID,
) *automerge.Document {
	t.Helper()

	document, err := factory(ctx, actorID)
	require.NoError(t, err)

	objects := map[string]*automerge.Object{"": document.Root()}
	texts := make(map[string]*automerge.Text)

	for index, operation := range scenario.Operations {
		parent := objects[scenarioPath(operation.Path)]
		switch operation.Action {
		case "createObject":
			require.NotNil(t, parent, "operation %d parent", index)
			object, err := parent.CreateObject(
				ctx,
				operation.Key,
				operation.ObjectType,
			)
			require.NoError(t, err, "operation %d", index)

			objects[scenarioPath(
				append(operation.Path, operation.Key),
			)] = object
		case "putScalar":
			require.NotNil(t, parent, "operation %d parent", index)
			require.NoError(
				t,
				parent.PutScalar(ctx, operation.Key, operation.Scalar.value(t)),
				"operation %d",
				index,
			)
		case "insertScalar":
			require.NotNil(t, parent, "operation %d parent", index)
			require.NoError(
				t,
				parent.InsertScalar(
					ctx,
					operation.Index,
					operation.Scalar.value(t),
				),
				"operation %d",
				index,
			)
		case "putScalarAt":
			require.NotNil(t, parent, "operation %d parent", index)
			require.NoError(
				t,
				parent.PutScalarAt(
					ctx,
					operation.Index,
					operation.Scalar.value(t),
				),
				"operation %d",
				index,
			)
		case "deleteIndex":
			require.NotNil(t, parent, "operation %d parent", index)
			require.NoError(
				t,
				parent.DeleteIndex(ctx, operation.Index),
				"operation %d",
				index,
			)
		case "createText":
			require.NotNil(t, parent, "operation %d parent", index)
			require.Empty(t, operation.Path)
			text, err := document.CreateText(ctx, operation.Key)
			require.NoError(t, err, "operation %d", index)

			texts[scenarioPath(
				append(operation.Path, operation.Key),
			)] = text
		case "spliceText":
			text := texts[scenarioPath(operation.Path)]
			require.NotNil(t, text, "operation %d text", index)
			require.NoError(
				t,
				text.Splice(
					ctx,
					uint32(operation.Index),
					operation.DeleteCount,
					operation.Text,
				),
				"operation %d",
				index,
			)
		case "increment":
			require.NotNil(t, parent, "operation %d parent", index)
			require.NoError(
				t,
				parent.Increment(ctx, operation.Key, operation.Delta),
				"operation %d",
				index,
			)
		case "commit":
			_, err := document.Commit(
				ctx,
				operation.Message,
				time.Unix(operation.Timestamp, 0),
			)
			require.NoError(t, err, "operation %d", index)
		default:
			require.Failf(
				t,
				"unsupported scenario action",
				"operation %d: %q",
				index,
				operation.Action,
			)
		}
	}

	return document
}

func assertInteropScenarioResult(
	t *testing.T,
	ctx context.Context,
	document *automerge.Document,
) {
	t.Helper()

	config, err := document.Root().Object(ctx, "config")
	require.NoError(t, err)
	assertScenarioScalar(
		t,
		config,
		"name",
		automerge.Scalar{Type: automerge.ScalarTypeString, String: "Policy 😀"},
	)
	assertScenarioScalar(
		t,
		config,
		"enabled",
		automerge.Scalar{Type: automerge.ScalarTypeBoolean, Bool: true},
	)
	assertScenarioScalar(
		t,
		config,
		"nothing",
		automerge.Scalar{Type: automerge.ScalarTypeNull},
	)
	assertScenarioScalar(
		t,
		config,
		"int",
		automerge.Scalar{Type: automerge.ScalarTypeInt, Int: -42},
	)
	assertScenarioScalar(
		t,
		config,
		"uint",
		automerge.Scalar{Type: automerge.ScalarTypeUint, Uint: 42},
	)
	assertScenarioScalar(
		t,
		config,
		"float64",
		automerge.Scalar{Type: automerge.ScalarTypeFloat64, Float: 3.25},
	)
	assertScenarioScalar(
		t,
		config,
		"bytes",
		automerge.Scalar{
			Type:  automerge.ScalarTypeBytes,
			Bytes: []byte{0, 1, 254, 255},
		},
	)
	assertScenarioScalar(
		t,
		config,
		"timestamp",
		automerge.Scalar{
			Type: automerge.ScalarTypeTimestamp,
			Int:  1_786_147_200_000,
		},
	)
	assertScenarioScalar(
		t,
		config,
		"counter",
		automerge.Scalar{Type: automerge.ScalarTypeCounter, Int: 8},
	)

	items, err := document.Root().Object(ctx, "items")
	require.NoError(t, err)
	length, err := items.Len(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1), length)

	item, err := items.ScalarAt(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, "replaced", item.String)

	text, err := document.Text(ctx, "body")
	require.NoError(t, err)
	value, err := text.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, "AXB", value)
}

func assertScenarioScalar(
	t *testing.T,
	object *automerge.Object,
	key string,
	expected automerge.Scalar,
) {
	t.Helper()

	actual, err := object.Scalar(context.Background(), key)
	require.NoError(t, err)
	assertScalarEqual(t, expected, actual)
}

func (s interopScenarioScalar) value(t *testing.T) automerge.Scalar {
	t.Helper()

	bytes, err := hex.DecodeString(s.Bytes)
	require.NoError(t, err)

	var floatBits uint64
	if s.FloatBits != "" {
		floatBits, err = strconv.ParseUint(s.FloatBits, 10, 64)
		require.NoError(t, err)
	}

	return automerge.Scalar{
		Type:   s.Type,
		Bool:   s.Bool,
		Uint:   s.Uint,
		Int:    s.Int,
		Float:  math.Float64frombits(floatBits),
		String: s.String,
		Bytes:  bytes,
	}
}

func scenarioPath(path []string) string {
	return strings.Join(path, "\x00")
}
