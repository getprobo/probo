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

package prosemirror

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pm "go.probo.inc/probo/pkg/prosemirror"
)

type schemaMappingLedger struct {
	Blocks []struct {
		Automerge   string `json:"automerge"`
		ProseMirror string `json:"prosemirror"`
		Outer       string `json:"outer"`
		IsEmbed     bool   `json:"isEmbed"`
	} `json:"blocks"`
	Marks []struct {
		Automerge   string `json:"automerge"`
		ProseMirror string `json:"prosemirror"`
	} `json:"marks"`
}

func loadSchemaMappingLedger(t *testing.T) schemaMappingLedger {
	t.Helper()

	raw, err := os.ReadFile("testdata/schema-mapping.json")
	require.NoError(t, err)

	var ledger schemaMappingLedger
	require.NoError(t, json.Unmarshal(raw, &ledger))
	require.NotEmpty(t, ledger.Blocks)
	require.NotEmpty(t, ledger.Marks)

	return ledger
}

// TestSchemaMappingLedger keeps the codec registry in lockstep with the shared
// ledger. Its counterpart schemaMappingParity.test.ts holds the frontend adapter
// to the same ledger, so the two implementations cannot drift apart without a
// test failing.
func TestSchemaMappingLedger(t *testing.T) {
	t.Parallel()

	ledger := loadSchemaMappingLedger(t)

	blockCodecs := make(map[string]nodeCodec)
	for _, codec := range nodeCodecs {
		if codec.Block() != "" {
			blockCodecs[codec.Block()] = codec
		}
	}

	t.Run(
		"blocks match",
		func(t *testing.T) {
			t.Parallel()

			want := make(map[string]string, len(ledger.Blocks))
			for _, entry := range ledger.Blocks {
				want[entry.Automerge] = entry.ProseMirror
			}

			got := make(map[string]string, len(blockCodecs))
			for blockType, codec := range blockCodecs {
				got[blockType] = string(codec.Node())
			}

			assert.Equal(t, want, got)
			assert.Equal(t, want, blockNodeNames)
		},
	)

	t.Run(
		"block list wrappers match",
		func(t *testing.T) {
			t.Parallel()

			for _, entry := range ledger.Blocks {
				codec, ok := blockCodecs[entry.Automerge]
				require.Truef(t, ok, "ledger block %q has no codec", entry.Automerge)
				assert.Equalf(
					t,
					entry.Outer,
					string(codec.Outer()),
					"block %q must map to list wrapper %q",
					entry.Automerge,
					entry.Outer,
				)
			}
		},
	)

	t.Run(
		"block embeds match",
		func(t *testing.T) {
			t.Parallel()

			for _, entry := range ledger.Blocks {
				codec, ok := blockCodecs[entry.Automerge]
				require.Truef(t, ok, "ledger block %q has no codec", entry.Automerge)
				assert.Equalf(
					t,
					entry.IsEmbed,
					codec.Role() == roleInlineEmbed,
					"block %q embed flag must match the ledger",
					entry.Automerge,
				)
			}
		},
	)

	t.Run(
		"marks match",
		func(t *testing.T) {
			t.Parallel()

			want := make(map[string]string, len(ledger.Marks))
			for _, entry := range ledger.Marks {
				want[entry.Automerge] = entry.ProseMirror
			}

			got := make(map[string]string, len(markCodecs))
			for _, codec := range markCodecs {
				got[codec.Automerge()] = string(codec.Mark())
			}

			assert.Equal(t, want, got)
			assert.Equal(t, want, markNodeNames)
		},
	)

	t.Run(
		"mark order matches schema rank",
		func(t *testing.T) {
			t.Parallel()

			for index, entry := range ledger.Marks {
				assert.Equalf(
					t,
					index,
					markRenderOrder[entry.Automerge],
					"mark %q must render at schema rank %d",
					entry.Automerge,
					index,
				)
				assert.Equalf(
					t,
					entry.Automerge,
					markCodecs[index].Automerge(),
					"mark codec at rank %d must be %q",
					index,
					entry.Automerge,
				)
			}

			assert.Len(t, markRenderOrder, len(ledger.Marks))
			assert.Len(t, markCodecs, len(ledger.Marks))
		},
	)
}

// TestCodecsCoverDocumentSchema fails if a node or mark the console editor can
// produce (prosemirror.DocumentSchema) has no bridge codec, which would silently
// drop that content when seeding or rendering a collaboration document. The doc
// root and the inline text node are structural and never need a codec.
//
// TODO: mention has no codec yet; the next commit adds one and drops it from this
// exclusion list.
func TestCodecsCoverDocumentSchema(t *testing.T) {
	t.Parallel()

	t.Run(
		"nodes",
		func(t *testing.T) {
			t.Parallel()

			covered := make(map[pm.NodeType]struct{}, len(nodeCodecs))
			for _, codec := range nodeCodecs {
				covered[codec.Node()] = struct{}{}
				if codec.Outer() != "" {
					covered[codec.Outer()] = struct{}{}
				}
			}

			excluded := map[pm.NodeType]struct{}{
				pm.NodeDoc:     {},
				pm.NodeText:    {},
				pm.NodeMention: {},
			}

			for _, nodeType := range pm.DocumentSchema.NodeTypes() {
				if _, skip := excluded[nodeType]; skip {
					continue
				}

				_, ok := covered[nodeType]
				assert.Truef(t, ok, "node %q in DocumentSchema has no codec", nodeType)
			}
		},
	)

	t.Run(
		"marks",
		func(t *testing.T) {
			t.Parallel()

			covered := make(map[pm.MarkType]struct{}, len(markCodecs))
			for _, codec := range markCodecs {
				covered[codec.Mark()] = struct{}{}
			}

			for _, markType := range pm.DocumentSchema.MarkTypes() {
				_, ok := covered[markType]
				assert.Truef(t, ok, "mark %q in DocumentSchema has no codec", markType)
			}
		},
	)
}
