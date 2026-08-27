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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

func TestDocument_HydrateReturnsTypedRootValue(t *testing.T) {
	t.Parallel()

	document, err := automerge.NewFrom(
		actor(201),
		map[string]automerge.Value{
			"counter": {
				Type:   automerge.ValueTypeScalar,
				Scalar: automerge.CounterScalar(3),
			},
			"nested": {
				Type: automerge.ValueTypeList,
				List: []automerge.Value{
					{Type: automerge.ValueTypeText, Text: "A😀B"},
					{
						Type:   automerge.ValueTypeScalar,
						Scalar: automerge.BytesScalar([]byte{1, 2, 3}),
					},
				},
			},
		},
		"hydrate",
		commitTime,
	)
	require.NoError(t, err)
	closeDocument(t, document)

	value, err := document.Hydrate()
	require.NoError(t, err)
	require.Equal(t, automerge.ValueTypeMap, value.Type)
	assert.Equal(t, automerge.CounterScalar(3), value.Map["counter"].Scalar)
	assert.Equal(t, "A😀B", value.Map["nested"].List[0].Text)
	assert.Equal(
		t,
		automerge.BytesScalar([]byte{1, 2, 3}),
		value.Map["nested"].List[1].Scalar,
	)
}

func TestRescue_RecoversInvalidMarkOrderSnapshot(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(
		"testdata/fixtures/broken_zero_width_mark.automerge",
	)
	require.NoError(t, err)

	document, err := automerge.Load(data, actor(202))
	assert.Error(t, err)
	assert.Nil(t, document)

	value, err := automerge.Rescue(data)
	require.NoError(t, err)
	require.Equal(t, automerge.ValueTypeMap, value.Type)
	assert.Equal(
		t,
		automerge.Value{Type: automerge.ValueTypeText, Text: ""},
		value.Map["text"],
	)
	assert.Equal(
		t,
		automerge.Value{Type: automerge.ValueTypeText, Text: "tsxs"},
		value.Map["o1"],
	)
}

func TestRescue_RejectsMalformedData(t *testing.T) {
	t.Parallel()

	_, err := automerge.Rescue([]byte("not an Automerge document"))
	require.Error(t, err)
}
