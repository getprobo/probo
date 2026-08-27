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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	automerge "go.probo.inc/probo/pkg/automerge/internal/testsupport"
)

func TestDocument_HydrateMatchesReference(t *testing.T) {
	t.Parallel()

	value := map[string]automerge.Value{
		"config": {
			Type: automerge.ValueTypeMap,
			Map: map[string]automerge.Value{
				"enabled": {
					Type: automerge.ValueTypeScalar,
					Scalar: automerge.Scalar{
						Type: automerge.ScalarTypeBoolean,
						Bool: true,
					},
				},
				"name": {
					Type: automerge.ValueTypeScalar,
					Scalar: automerge.Scalar{
						Type:   automerge.ScalarTypeString,
						String: "Policy",
					},
				},
			},
		},
		"items": {
			Type: automerge.ValueTypeList,
			List: []automerge.Value{
				{
					Type: automerge.ValueTypeScalar,
					Scalar: automerge.Scalar{
						Type: automerge.ScalarTypeInt,
						Int:  1,
					},
				},
				{
					Type: automerge.ValueTypeMap,
					Map: map[string]automerge.Value{
						"nested": {
							Type: automerge.ValueTypeText,
							Text: "A😀B",
						},
					},
				},
				{
					Type: automerge.ValueTypeList,
					List: []automerge.Value{
						{
							Type: automerge.ValueTypeScalar,
							Scalar: automerge.Scalar{
								Type:   automerge.ScalarTypeString,
								String: "deep",
							},
						},
					},
				},
			},
		},
		"text": {
			Type: automerge.ValueTypeText,
			Text: "Hello",
		},
	}

	nativeDocument, err := automerge.NewFrom(

		actor(163),
		value,
		"hydrate",
		commitTime,
	)
	require.NoError(t, err)
	closeDocument(t, nativeDocument)

	referenceDocument, err := automerge.NewReferenceFrom(

		actor(163),
		value,
		"hydrate",
		commitTime,
	)
	require.NoError(t, err)
	closeDocument(t, referenceDocument)

	nativeHeads, err := nativeDocument.Heads()
	require.NoError(t, err)
	referenceHeads, err := referenceDocument.Heads()
	require.NoError(t, err)
	assert.Equal(t, referenceHeads, nativeHeads)
	assertHydratedDocument(t, nativeDocument)
	assertHydratedDocument(t, referenceDocument)

	nativeData, err := nativeDocument.Save()
	require.NoError(t, err)
	referenceFromNative, err := automerge.LoadReference(

		nativeData,
		actor(164),
	)
	require.NoError(t, err)
	closeDocument(t, referenceFromNative)
	assertHydratedDocument(t, referenceFromNative)
}

func TestDocument_HydrateRollback(t *testing.T) {
	t.Parallel()

	document, err := automerge.New(actor(165))
	require.NoError(t, err)
	closeDocument(t, document)
	require.NoError(
		t,
		document.Root().PutMap(

			map[string]automerge.Value{
				"value": {
					Type: automerge.ValueTypeList,
					List: []automerge.Value{
						{
							Type: automerge.ValueTypeScalar,
							Scalar: automerge.Scalar{
								Type: automerge.ScalarTypeInt,
								Int:  1,
							},
						},
					},
				},
			},
		),
	)
	cancelled, err := document.Rollback()
	require.NoError(t, err)
	assert.Positive(t, cancelled)

	_, err = document.Root().Object("value")
	require.Error(t, err)
}

func TestDocument_HydrateSpliceMatchesReference(t *testing.T) {
	t.Parallel()

	factories := map[string]func(
		automerge.ActorID,
		map[string]automerge.Value,
		string,
		time.Time,
	) (*automerge.Document, error){
		"native":    automerge.NewFrom,
		"reference": automerge.NewReferenceFrom,
	}

	for name, factory := range factories {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				document, err := factory(

					actor(166),
					map[string]automerge.Value{
						"list": {
							Type: automerge.ValueTypeList,
							List: []automerge.Value{
								hydratedInt(1),
								hydratedInt(2),
								hydratedInt(3),
							},
						},
					},
					"initial",
					commitTime,
				)
				require.NoError(t, err)
				closeDocument(t, document)
				list, err := document.Root().Object("list")
				require.NoError(t, err)
				require.NoError(
					t,
					list.SpliceValues(

						1,
						1,
						[]automerge.Value{
							{
								Type: automerge.ValueTypeMap,
								Map: map[string]automerge.Value{
									"value": hydratedInt(4),
								},
							},
							{Type: automerge.ValueTypeText, Text: "text"},
						},
					),
				)
				require.NoError(
					t,
					list.PutValueAt(

						3,
						automerge.Value{
							Type: automerge.ValueTypeList,
							List: []automerge.Value{hydratedInt(5)},
						},
					),
				)

				_, err = document.Commit(

					"splice",
					commitTime.Add(time.Second),
				)
				require.NoError(t, err)

				length, err := list.Len()
				require.NoError(t, err)
				assert.Equal(t, uint64(4), length)

				first, err := list.ScalarAt(0)
				require.NoError(t, err)
				assert.Equal(t, int64(1), first.Int)

				nested, err := list.ObjectAt(1)
				require.NoError(t, err)
				nestedValue, err := nested.Scalar("value")
				require.NoError(t, err)
				assert.Equal(t, int64(4), nestedValue.Int)

				textObject, err := list.ObjectAt(2)
				require.NoError(t, err)
				text, err := textObject.Text()
				require.NoError(t, err)
				textValue, err := text.String()
				require.NoError(t, err)
				assert.Equal(t, "text", textValue)

				nestedList, err := list.ObjectAt(3)
				require.NoError(t, err)
				last, err := nestedList.ScalarAt(0)
				require.NoError(t, err)
				assert.Equal(t, int64(5), last.Int)
			},
		)
	}
}

func hydratedInt(value int64) automerge.Value {
	return automerge.Value{
		Type: automerge.ValueTypeScalar,
		Scalar: automerge.Scalar{
			Type: automerge.ScalarTypeInt,
			Int:  value,
		},
	}
}

func assertHydratedDocument(
	t *testing.T,
	document *automerge.Document,
) {
	t.Helper()

	config, err := document.Root().Object("config")
	require.NoError(t, err)
	enabled, err := config.Scalar("enabled")
	require.NoError(t, err)
	assert.True(t, enabled.Bool)

	name, err := config.Scalar("name")
	require.NoError(t, err)
	assert.Equal(t, "Policy", name.String)

	items, err := document.Root().Object("items")
	require.NoError(t, err)
	length, err := items.Len()
	require.NoError(t, err)
	assert.Equal(t, uint64(3), length)

	first, err := items.ScalarAt(0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), first.Int)

	nestedMap, err := items.ObjectAt(1)
	require.NoError(t, err)
	nestedTextObject, err := nestedMap.Object("nested")
	require.NoError(t, err)
	nestedText, err := nestedTextObject.Text()
	require.NoError(t, err)
	nestedValue, err := nestedText.String()
	require.NoError(t, err)
	assert.Equal(t, "A😀B", nestedValue)

	text, err := document.Text("text")
	require.NoError(t, err)
	value, err := text.String()
	require.NoError(t, err)
	assert.Equal(t, "Hello", value)
}
