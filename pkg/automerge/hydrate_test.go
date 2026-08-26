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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
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

	ctx := context.Background()
	nativeDocument, err := automerge.NewFrom(
		ctx,
		actor(163),
		value,
		"hydrate",
		commitTime,
	)
	require.NoError(t, err)
	closeDocument(t, nativeDocument)

	referenceDocument, err := automerge.NewReferenceFrom(
		ctx,
		actor(163),
		value,
		"hydrate",
		commitTime,
	)
	require.NoError(t, err)
	closeDocument(t, referenceDocument)

	nativeHeads, err := nativeDocument.Heads(ctx)
	require.NoError(t, err)
	referenceHeads, err := referenceDocument.Heads(ctx)
	require.NoError(t, err)
	assert.Equal(t, referenceHeads, nativeHeads)
	assertHydratedDocument(t, ctx, nativeDocument)
	assertHydratedDocument(t, ctx, referenceDocument)

	nativeData, err := nativeDocument.Save(ctx)
	require.NoError(t, err)
	referenceFromNative, err := automerge.LoadReference(
		ctx,
		nativeData,
		actor(164),
	)
	require.NoError(t, err)
	closeDocument(t, referenceFromNative)
	assertHydratedDocument(t, ctx, referenceFromNative)
}

func TestDocument_HydrateRollback(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	document, err := automerge.New(ctx, actor(165))
	require.NoError(t, err)
	closeDocument(t, document)
	require.NoError(
		t,
		document.Root().PutMap(
			ctx,
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
	cancelled, err := document.Rollback(ctx)
	require.NoError(t, err)
	assert.Positive(t, cancelled)

	_, err = document.Root().Object(ctx, "value")
	require.Error(t, err)
}

func TestDocument_HydrateSpliceMatchesReference(t *testing.T) {
	t.Parallel()

	factories := map[string]func(
		context.Context,
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

				ctx := context.Background()
				document, err := factory(
					ctx,
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
				list, err := document.Root().Object(ctx, "list")
				require.NoError(t, err)
				require.NoError(
					t,
					list.SpliceValues(
						ctx,
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
						ctx,
						3,
						automerge.Value{
							Type: automerge.ValueTypeList,
							List: []automerge.Value{hydratedInt(5)},
						},
					),
				)
				_, err = document.Commit(
					ctx,
					"splice",
					commitTime.Add(time.Second),
				)
				require.NoError(t, err)

				length, err := list.Len(ctx)
				require.NoError(t, err)
				assert.Equal(t, uint64(4), length)

				first, err := list.ScalarAt(ctx, 0)
				require.NoError(t, err)
				assert.Equal(t, int64(1), first.Int)

				nested, err := list.ObjectAt(ctx, 1)
				require.NoError(t, err)
				nestedValue, err := nested.Scalar(ctx, "value")
				require.NoError(t, err)
				assert.Equal(t, int64(4), nestedValue.Int)

				textObject, err := list.ObjectAt(ctx, 2)
				require.NoError(t, err)
				text, err := textObject.Text(ctx)
				require.NoError(t, err)
				textValue, err := text.String(ctx)
				require.NoError(t, err)
				assert.Equal(t, "text", textValue)

				nestedList, err := list.ObjectAt(ctx, 3)
				require.NoError(t, err)
				last, err := nestedList.ScalarAt(ctx, 0)
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
	ctx context.Context,
	document *automerge.Document,
) {
	t.Helper()

	config, err := document.Root().Object(ctx, "config")
	require.NoError(t, err)
	enabled, err := config.Scalar(ctx, "enabled")
	require.NoError(t, err)
	assert.True(t, enabled.Bool)

	name, err := config.Scalar(ctx, "name")
	require.NoError(t, err)
	assert.Equal(t, "Policy", name.String)

	items, err := document.Root().Object(ctx, "items")
	require.NoError(t, err)
	length, err := items.Len(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(3), length)

	first, err := items.ScalarAt(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), first.Int)

	nestedMap, err := items.ObjectAt(ctx, 1)
	require.NoError(t, err)
	nestedTextObject, err := nestedMap.Object(ctx, "nested")
	require.NoError(t, err)
	nestedText, err := nestedTextObject.Text(ctx)
	require.NoError(t, err)
	nestedValue, err := nestedText.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, "A😀B", nestedValue)

	text, err := document.Text(ctx, "text")
	require.NoError(t, err)
	value, err := text.String(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Hello", value)
}
