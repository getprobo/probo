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

// This file reproduces test_compressed_doc_cols (rust/automerge/tests/test.rs):
// a document large enough to trigger DEFLATE compression must save smaller with
// compression than without, and the compressed save must load back to the same
// value on both engines.

package automerge_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

func TestRustTest_CompressedDocCols(t *testing.T) {
	t.Parallel()

	const items = 200

	values := make([]automerge.Value, items)
	for i := range values {
		values[i] = automerge.Value{
			Type:   automerge.ValueTypeScalar,
			Scalar: automerge.Scalar{Type: automerge.ScalarTypeUint, Uint: uint64(i)},
		}
	}

	for _, engine := range rustParityEngines() {
		t.Run(
			engine.name,
			func(t *testing.T) {
				t.Parallel()

				document, err := engine.open(actor(0x01))
				require.NoError(t, err)
				closeDocument(t, document)

				require.NoError(
					t,
					document.Root().PutValue(

						"list",
						automerge.Value{
							Type: automerge.ValueTypeList,
							List: values,
						},
					),
				)

				_, err = document.Commit("list", commitTime)
				require.NoError(t, err)

				uncompressed, err := document.Save(automerge.NoCompress())
				require.NoError(t, err)

				compressed, err := document.Save()
				require.NoError(t, err)

				assert.Less(
					t,
					len(compressed),
					len(uncompressed),
					"compressed save should be smaller than uncompressed",
				)

				loaded, err := engine.load(compressed, actor(0x02))
				require.NoError(t, err)
				closeDocument(t, loaded)

				list, err := loaded.Root().Object("list")
				require.NoError(t, err)

				length, err := list.Len()
				require.NoError(t, err)
				require.Equal(t, uint64(items), length)

				for i := range items {
					value, err := list.ScalarAt(uint64(i))
					require.NoError(t, err)
					assert.Equal(t, uint64(i), value.Uint)
				}
			},
		)
	}
}
