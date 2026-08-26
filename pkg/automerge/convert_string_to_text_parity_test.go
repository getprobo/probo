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

// The tests in this file reproduce the string-to-text load migration from
// upstream Rust automerge 0.10 (rust/automerge/tests/convert_string_to_text.rs),
// asserting the native Go and Rust/WASM reference engines both convert string
// scalars into text objects when loading with the migration enabled.

package automerge_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
)

// TestRustCurrentState_LoadChanges reproduces test_load_changes: loading a stored
// document and materializing its current state yields a single put of the
// counter's summed value.
func TestRustCurrentState_LoadChanges(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("testdata/fixtures/counter_value_is_ok.automerge")
	require.NoError(t, err)

	result := make(map[string][]automerge.Patch)

	for _, engine := range rustParityEngines() {
		document, err := engine.load(data, actor(1))
		require.NoError(t, err)
		closeDocument(t, document)

		patches, err := document.CurrentState()
		require.NoError(t, err)

		result[engine.name] = patches
	}

	reference := result["reference"]
	require.Len(t, reference, 1)
	assert.Equal(t, automerge.PatchPutMap, reference[0].Action)
	assert.Equal(t, "a", reference[0].Key)
	require.NotNil(t, reference[0].Value.Scalar)
	assert.Equal(t, automerge.ScalarTypeCounter, reference[0].Value.Scalar.Type)
	assert.Equal(t, int64(2000), reference[0].Value.Scalar.Int)
	assert.Equal(t, result["reference"], result["native"])
}

func loadConvertingEngines() []struct {
	name string
	load func([]byte, automerge.ActorID) (*automerge.Document, error)
} {
	return []struct {
		name string
		load func([]byte, automerge.ActorID) (*automerge.Document, error)
	}{
		{
			"native",
			func(data []byte, actorID automerge.ActorID) (*automerge.Document, error) {
				return automerge.Load(data, actorID, automerge.ConvertStringsToText())
			},
		},
		{
			"reference",
			func(data []byte, actorID automerge.ActorID) (*automerge.Document, error) {
				return automerge.LoadReference(data, actorID, automerge.ConvertStringsToText())
			},
		},
	}
}

// TestRustConvert_StringsInMapsAreConvertedToText reproduces
// test_strings_in_maps_are_converted_to_text.
func TestRustConvert_StringsInMapsAreConvertedToText(t *testing.T) {
	t.Parallel()

	source, err := automerge.New(actor(0xaa))
	require.NoError(t, err)
	require.NoError(
		t,
		source.Root().PutScalar(

			"somestring",
			automerge.Scalar{Type: automerge.ScalarTypeString, String: "hello"},
		),
	)
	_, err = source.Commit("seed", commitTime)
	require.NoError(t, err)

	saved, err := source.Save()
	require.NoError(t, err)
	require.NoError(t, source.Close())

	for _, engine := range loadConvertingEngines() {
		document, err := engine.load(saved, actor(0xbb))
		require.NoError(t, err)
		closeDocument(t, document)

		object, err := document.Root().Object("somestring")
		require.NoError(t, err)
		assert.Equal(t, automerge.ObjectTypeText, object.Type)

		text, err := object.Text()
		require.NoError(t, err)
		value, err := text.String()
		require.NoError(t, err)
		assert.Equal(t, "hello", value)
	}
}

// TestRustConvert_StringsInListsAreConvertedToText reproduces
// test_strings_in_lists_are_converted_to_text.
func TestRustConvert_StringsInListsAreConvertedToText(t *testing.T) {
	t.Parallel()

	source, err := automerge.New(actor(0xaa))
	require.NoError(t, err)
	list, err := source.Root().CreateObject("list", automerge.ObjectTypeList)
	require.NoError(t, err)
	require.NoError(t, list.InsertScalar(0, automerge.Scalar{Type: automerge.ScalarTypeString, String: "hello"}))
	_, err = source.Commit("seed", commitTime)
	require.NoError(t, err)

	saved, err := source.Save()
	require.NoError(t, err)
	require.NoError(t, source.Close())

	for _, engine := range loadConvertingEngines() {
		document, err := engine.load(saved, actor(0xbb))
		require.NoError(t, err)
		closeDocument(t, document)

		listObject, err := document.Root().Object("list")
		require.NoError(t, err)
		element, err := listObject.ObjectAt(0)
		require.NoError(t, err)
		assert.Equal(t, automerge.ObjectTypeText, element.Type)

		text, err := element.Text()
		require.NoError(t, err)
		value, err := text.String()
		require.NoError(t, err)
		assert.Equal(t, "hello", value)
	}
}

// TestRustConvert_DoesNotAddSizeWhenStringsAreNotConverted reproduces
// test_does_not_add_size_when_strings_are_not_converted.
func TestRustConvert_DoesNotAddSizeWhenStringsAreNotConverted(t *testing.T) {
	t.Parallel()

	empty, err := automerge.New(actor(0xaa))
	require.NoError(t, err)

	saved, err := empty.Save()
	require.NoError(t, err)
	require.NoError(t, empty.Close())

	for _, engine := range loadConvertingEngines() {
		document, err := engine.load(saved, actor(0xbb))
		require.NoError(t, err)
		closeDocument(t, document)

		resaved, err := document.Save()
		require.NoError(t, err)
		assert.Equal(t, len(saved), len(resaved))
	}
}
