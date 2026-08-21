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

// TestSchemaMappingLedger keeps the Go renderer's block and mark tables in lockstep
// with the shared ledger. Its counterpart schemaMappingParity.test.ts holds the
// frontend adapter to the same ledger, so the two implementations cannot drift
// apart without a test failing.
func TestSchemaMappingLedger(t *testing.T) {
	t.Parallel()

	ledger := loadSchemaMappingLedger(t)

	t.Run("blocks match", func(t *testing.T) {
		t.Parallel()

		want := make(map[string]string, len(ledger.Blocks))
		for _, entry := range ledger.Blocks {
			want[entry.Automerge] = entry.ProseMirror
		}

		got := make(map[string]string, len(blockMappings))
		for _, mapping := range blockMappings {
			got[mapping.Automerge] = mapping.ProseMirror
		}

		assert.Equal(t, want, got)
		assert.Equal(t, want, blockNodeNames)
	})

	t.Run("marks match", func(t *testing.T) {
		t.Parallel()

		want := make(map[string]string, len(ledger.Marks))
		for _, entry := range ledger.Marks {
			want[entry.Automerge] = entry.ProseMirror
		}

		got := make(map[string]string, len(markMappings))
		for _, mapping := range markMappings {
			got[mapping.Automerge] = mapping.ProseMirror
		}

		assert.Equal(t, want, got)
		assert.Equal(t, want, markNodeNames)
	})

	t.Run("mark order matches schema rank", func(t *testing.T) {
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
		}

		assert.Len(t, markRenderOrder, len(ledger.Marks))
	})
}
