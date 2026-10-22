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

package collaboration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

func TestParseAutomergeHashes_RoundTrip(t *testing.T) {
	t.Parallel()

	var first, second automerge.Hash

	first[0] = 1
	second[31] = 2

	encoded := EncodeAutomergeHashes([]automerge.Hash{first, second})
	require.Len(t, encoded, 2)

	decoded, err := ParseAutomergeHashes(encoded)
	require.NoError(t, err)
	assert.Equal(t, []automerge.Hash{first, second}, decoded)
}

func TestWriteSpans_ReplacesSeededContent(t *testing.T) {
	t.Parallel()

	const (
		first = `{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"Vendors"}]},{"type":"paragraph","content":[{"type":"text","text":"Legal Name: ","marks":[{"type":"bold"}]},{"type":"text","text":"Acme"}]},{"type":"horizontalRule"}]}`
		same  = first
		next  = `{"type":"doc","content":[{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"Vendors"}]},{"type":"paragraph","content":[{"type":"text","text":"Legal Name: ","marks":[{"type":"bold"}]},{"type":"text","text":"Beta"}]},{"type":"horizontalRule"}]}`
	)

	actorID, err := automerge.NewActorID()
	require.NoError(t, err)

	document, err := automerge.New(actorID)
	require.NoError(t, err)

	require.NoError(t, writeSpans(document, first, true))
	_, err = document.Commit("Set document content", time.Now())
	require.NoError(t, err)

	snapshot, err := document.Save()
	require.NoError(t, err)
	require.NoError(t, document.Close())

	reloaded, err := automerge.Load(snapshot, actorID)
	require.NoError(t, err)

	defer func() { _ = reloaded.Close() }()

	require.NoError(t, writeSpans(reloaded, same, false))
	_, err = reloaded.Commit("Set document content", time.Now())
	require.ErrorIs(t, err, automerge.ErrNoOperations)

	require.NoError(t, writeSpans(reloaded, next, false))
	_, err = reloaded.Commit("Set document content", time.Now())
	require.NoError(t, err)

	rendered, err := materialize(context.Background(), reloaded)
	require.NoError(t, err)
	assert.Contains(t, rendered, "Beta")
	assert.NotContains(t, rendered, "Acme")

	const table = `{"type":"doc","content":[{"type":"table","content":[{"type":"tableRow","content":[{"type":"tableHeader","content":[{"type":"paragraph","content":[{"type":"text","text":"Name"}]}]},{"type":"tableHeader","content":[{"type":"paragraph","content":[{"type":"text","text":"Owner"}]}]}]},{"type":"tableRow","content":[{"type":"tableCell","content":[{"type":"paragraph","content":[{"type":"text","text":"Laptop"}]}]},{"type":"tableCell","content":[{"type":"paragraph","content":[{"type":"text","text":"Ada"}]}]}]}]}]}`

	tableDoc, err := automerge.New(actorID)
	require.NoError(t, err)

	require.NoError(t, writeSpans(tableDoc, table, true))
	_, err = tableDoc.Commit("Set document content", time.Now())
	require.NoError(t, err)

	tableSnapshot, err := tableDoc.Save()
	require.NoError(t, err)
	require.NoError(t, tableDoc.Close())

	reloadedTable, err := automerge.Load(tableSnapshot, actorID)
	require.NoError(t, err)

	defer func() { _ = reloadedTable.Close() }()

	require.NoError(t, writeSpans(reloadedTable, table, false))
	_, err = reloadedTable.Commit("Set document content", time.Now())
	require.ErrorIs(t, err, automerge.ErrNoOperations)
}
