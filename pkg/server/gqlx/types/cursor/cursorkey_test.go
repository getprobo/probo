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

package cursor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type testOrderField string

func (f testOrderField) Column() string { return string(f) }
func (f testOrderField) String() string { return string(f) }

func TestNewCursor_BeforeWinsOverLeakedFirst(t *testing.T) {
	t.Parallel()

	first := 15
	last := 15
	before := page.NewCursorKey(gid.New(gid.NewTenantID(), 1), "cursor")
	orderBy := page.OrderBy[testOrderField]{
		Field:     "created_at",
		Direction: page.OrderDirectionDesc,
	}

	cursor := NewCursor(&first, nil, &last, &before, orderBy)

	require.NotNil(t, cursor.Key)
	assert.Equal(t, page.Tail, cursor.Position)
	assert.Equal(t, before.String(), cursor.Key.String())
	assert.Equal(t, last, cursor.Size)
}

func TestNewCursor_AfterWinsOverLeakedLast(t *testing.T) {
	t.Parallel()

	first := 15
	last := 15
	after := page.NewCursorKey(gid.New(gid.NewTenantID(), 1), "cursor")
	orderBy := page.OrderBy[testOrderField]{
		Field:     "created_at",
		Direction: page.OrderDirectionDesc,
	}

	cursor := NewCursor(&first, &after, &last, nil, orderBy)

	require.NotNil(t, cursor.Key)
	assert.Equal(t, page.Head, cursor.Position)
	assert.Equal(t, after.String(), cursor.Key.String())
	assert.Equal(t, first, cursor.Size)
}

func TestNewCursor_FirstPageIgnoresEmptyCursors(t *testing.T) {
	t.Parallel()

	first := 15
	orderBy := page.OrderBy[testOrderField]{
		Field:     "created_at",
		Direction: page.OrderDirectionDesc,
	}

	cursor := NewCursor(&first, nil, nil, nil, orderBy)

	assert.Nil(t, cursor.Key)
	assert.Equal(t, page.Head, cursor.Position)
	assert.Equal(t, first, cursor.Size)
}
