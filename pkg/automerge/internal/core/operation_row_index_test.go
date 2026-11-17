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

package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func requireRowsMatchOperations(t *testing.T, engine *Engine) {
	t.Helper()

	rows := engine.columns.operationRows
	assert.LessOrEqual(t, rows.chainDepth(), maxOperationRowIndexDepth)

	for row, operation := range engine.columns.operations {
		found, ok := rows.lookup(operation.ID)
		require.True(t, ok, "operation %v is missing from the row index", operation.ID)
		require.Equal(t, row, found, "operation %v has the wrong row", operation.ID)
	}
}

func TestOperationRowIndex_StaysBoundedAcrossCommits(t *testing.T) {
	t.Parallel()

	engine, err := NewEngine()
	require.NoError(t, err)

	text, err := engine.PutObject(0, "body", "text")
	require.NoError(t, err)
	require.NoError(t, engine.SpliceText(text, 0, 0, "hello"))
	_, err = engine.Commit("init", time.Time{})
	require.NoError(t, err)

	for range 3 * maxOperationRowIndexDepth {
		require.NoError(t, engine.SpliceText(text, 0, 0, "e"))
		_, err = engine.Commit("typing", time.Time{})
		require.NoError(t, err)

		requireRowsMatchOperations(t, engine)
	}
}

func TestOperationRowIndex_StaysBoundedAcrossApplies(t *testing.T) {
	t.Parallel()

	base, err := NewEngine()
	require.NoError(t, err)

	text, err := base.PutObject(0, "body", "text")
	require.NoError(t, err)
	require.NoError(t, base.SpliceText(text, 0, 0, "hello"))
	_, err = base.Commit("init", time.Time{})
	require.NoError(t, err)

	snapshot, err := base.Save(false, false)
	require.NoError(t, err)

	baseHeads, err := base.Heads()
	require.NoError(t, err)

	client, err := LoadEngine(snapshot)
	require.NoError(t, err)

	clientText, _, err := client.GetObject(0, "body")
	require.NoError(t, err)

	for range 3 * maxOperationRowIndexDepth {
		require.NoError(t, client.SpliceText(clientText, 0, 0, "e"))
		_, err = client.Commit("typing", time.Time{})
		require.NoError(t, err)
	}

	changes, _, err := client.ChangesSince(baseHeads)
	require.NoError(t, err)
	require.Len(t, changes, 3*maxOperationRowIndexDepth)

	server, err := LoadEngine(snapshot)
	require.NoError(t, err)

	for _, change := range changes {
		require.NoError(t, server.ApplyChanges([][]byte{change}))

		requireRowsMatchOperations(t, server)
	}
}
