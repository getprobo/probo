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

// The tests in this file reproduce the cross-engine text behaviors from the
// upstream JavaScript suite (javascript/test/text_test.ts). Each runs on the
// native Go engine and the Rust/WASM reference engine and asserts they agree.

package automerge_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
)

func hydratedFromEngines() []struct {
	name string
	from func(automerge.ActorID, map[string]automerge.Value, string) (*automerge.Document, error)
} {
	return []struct {
		name string
		from func(automerge.ActorID, map[string]automerge.Value, string) (*automerge.Document, error)
	}{
		{"native", func(actorID automerge.ActorID, value map[string]automerge.Value, message string) (*automerge.Document, error) {
			return automerge.NewFrom(actorID, value, message, commitTime)
		}},
		{"reference", func(actorID automerge.ActorID, value map[string]automerge.Value, message string) (*automerge.Document, error) {
			return automerge.NewReferenceFrom(actorID, value, message, commitTime)
		}},
	}
}

// TestJSText_ImplicitAndExplicitDeletion reproduces the "implicit and explicit
// deletion" case: a delete splice removes a character and a zero-length splice
// is a no-op.
func TestJSText_ImplicitAndExplicitDeletion(t *testing.T) {
	t.Parallel()

	result := make(map[string]string)

	for _, engine := range rustParityEngines() {
		document, _, text := seedText(t, engine, "abc")
		require.NoError(t, text.Splice(1, 1, ""))
		require.NoError(t, text.Splice(1, 0, ""))
		_, err := document.Commit("edit", commitTime)
		require.NoError(t, err)

		value, err := text.String()
		require.NoError(t, err)

		result[engine.name] = value
	}

	assert.Equal(t, "ac", result["reference"])
	assert.Equal(t, result["reference"], result["native"])
}

// TestJSText_TextAndOtherOpsSameChange reproduces "text and other ops in the
// same change": a scalar put and a text splice committed together both apply.
func TestJSText_TextAndOtherOpsSameChange(t *testing.T) {
	t.Parallel()

	foos := make(map[string]string)
	texts := make(map[string]string)

	for _, engine := range rustParityEngines() {
		document, _, text := seedText(t, engine, "")
		require.NoError(
			t,
			document.Root().PutScalar(

				"foo",
				automerge.Scalar{Type: automerge.ScalarTypeString, String: "bar"},
			),
		)
		require.NoError(t, text.Splice(0, 0, "a"))
		_, err := document.Commit("mixed", commitTime)
		require.NoError(t, err)

		foo, err := document.Root().Scalar("foo")
		require.NoError(t, err)

		foos[engine.name] = foo.String

		value, err := text.String()
		require.NoError(t, err)

		texts[engine.name] = value
	}

	assert.Equal(t, "bar", foos["reference"])
	assert.Equal(t, "a", texts["reference"])
	assert.Equal(t, foos["reference"], foos["native"])
	assert.Equal(t, texts["reference"], texts["native"])
}

// TestJSText_InitializeTextInFrom reproduces "initialize text in
// Automerge.from()" and "encode the initial value as a change": a hydrated text
// value round-trips through save/load as a single change.
func TestJSText_InitializeTextInFrom(t *testing.T) {
	t.Parallel()

	loaded := make(map[string]string)
	changes := make(map[string]uint64)

	root := map[string]automerge.Value{
		"text": {Type: automerge.ValueTypeText, Text: "init"},
	}

	for _, engine := range hydratedFromEngines() {
		document, err := engine.from(actor(0xaa), root, "init")
		require.NoError(t, err)
		closeDocument(t, document)

		text, err := document.Text("text")
		require.NoError(t, err)

		value, err := text.String()
		require.NoError(t, err)
		assert.Equal(t, "init", value)

		stats, err := document.Stats()
		require.NoError(t, err)

		changes[engine.name] = stats.NumChanges

		saved, err := document.Save()
		require.NoError(t, err)

		load := automerge.Load
		if engine.name == "reference" {
			load = automerge.LoadReference
		}

		reloaded, err := load(saved, actor(0xcc))
		require.NoError(t, err)
		closeDocument(t, reloaded)

		reloadedText, err := reloaded.Text("text")
		require.NoError(t, err)
		loaded[engine.name], err = reloadedText.String()
		require.NoError(t, err)
	}

	assert.Equal(t, "init", loaded["reference"])
	assert.Equal(t, loaded["reference"], loaded["native"])
	assert.Equal(t, uint64(1), changes["reference"])
	assert.Equal(t, changes["reference"], changes["native"])
}

// TestJSText_SplicingIntoArrays reproduces "splicing into text in arrays": text
// nested inside lists can be spliced by descending into the nested objects.
func TestJSText_SplicingIntoArrays(t *testing.T) {
	t.Parallel()

	result := make(map[string]string)

	root := map[string]automerge.Value{
		"dom": {
			Type: automerge.ValueTypeList,
			List: []automerge.Value{{
				Type: automerge.ValueTypeList,
				List: []automerge.Value{{Type: automerge.ValueTypeText, Text: "world"}},
			}},
		},
	}

	for _, engine := range hydratedFromEngines() {
		document, err := engine.from(actor(0xaa), root, "init")
		require.NoError(t, err)
		closeDocument(t, document)

		outer, err := document.Root().Object("dom")
		require.NoError(t, err)
		inner, err := outer.ObjectAt(0)
		require.NoError(t, err)
		textObject, err := inner.ObjectAt(0)
		require.NoError(t, err)
		text, err := textObject.Text()
		require.NoError(t, err)

		require.NoError(t, text.Splice(0, 0, "Hello "))
		_, err = document.Commit("splice", commitTime)
		require.NoError(t, err)

		value, err := text.String()
		require.NoError(t, err)

		result[engine.name] = value
	}

	assert.Equal(t, "Hello world", result["reference"])
	assert.Equal(t, result["reference"], result["native"])
}
