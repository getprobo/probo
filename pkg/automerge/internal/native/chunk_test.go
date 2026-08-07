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

package native_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/automerge/internal/native"
)

func TestParseChunks_OfficialDocument(t *testing.T) {
	t.Parallel()

	var actor automerge.ActorID
	actor[0] = 1
	document, err := automerge.New(context.Background(), actor)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, document.Close(context.Background()))
	})

	text, err := document.CreateText(context.Background(), "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(context.Background(), 0, 0, "Hello"))
	_, err = document.Commit(
		context.Background(),
		"Create document",
		time.Unix(1_786_104_000, 0),
	)
	require.NoError(t, err)
	data, err := document.Save(context.Background())
	require.NoError(t, err)

	chunks, err := native.ParseChunks(data)

	require.NoError(t, err)
	require.Len(t, chunks, 1)
	assert.Equal(t, native.ChunkTypeDocument, chunks[0].Type)
	assert.Equal(t, data, chunks[0].Raw)
	assert.NotEmpty(t, chunks[0].Data)

	parsed, err := native.ParseDocument(data)
	require.NoError(t, err)
	require.Len(t, parsed.Actors, 1)
	assert.Equal(t, actor[:], parsed.Actors[0])
	assert.Equal(t, chunks[0].Hash, parsed.Hash)
	assert.NotEmpty(t, parsed.Heads)
	assert.NotEmpty(t, parsed.ChangeColumns)
	assert.NotEmpty(t, parsed.OpColumns)
}

func TestParseChunks_RejectsCorruption(t *testing.T) {
	t.Parallel()

	data := []byte{
		0x85, 0x6f, 0x4a, 0x83,
		0x00, 0x00, 0x00, 0x00,
		0x01, 0x01, 0x00,
	}
	_, err := native.ParseChunks(data)
	assert.ErrorContains(t, err, "checksum mismatch")

	_, err = native.ParseChunks(data[:9])
	assert.ErrorContains(t, err, "truncated chunk header")
}
