// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package page

import (
	"context"
	"fmt"
)

// MaxLoadAllPages caps how many pages LoadAll walks, bounding a single
// call to MaxLoadAllPages*MaxCursorSize rows. Past that, LoadAll errors
// rather than materialising an unbounded set.
const MaxLoadAllPages = 20

// Loader runs one paginated query for the given cursor and returns the
// rows it loaded. Callers bind the connection, scope, parent key and
// filter in a closure, exposing only ctx and cursor (typically a coredata
// LoadBy* on a fresh receiver).
type Loader[T Paginable[U], U OrderField] func(ctx context.Context, cursor *Cursor[U]) ([]T, error)

// LoadAll walks every matching row via keyset pagination, advancing a
// MaxCursorSize forward cursor until no rows remain, and returns them
// concatenated. fetch runs the paginated query for the cursor. It errors
// past MaxLoadAllPages pages.
func LoadAll[T Paginable[U], U OrderField](
	ctx context.Context,
	orderBy OrderBy[U],
	fetch Loader[T, U],
) ([]T, error) {
	var (
		all []T
		key *CursorKey
	)

	for page := 0; ; page++ {
		if page >= MaxLoadAllPages {
			return nil, fmt.Errorf(
				"cannot load all rows: result set exceeds %d rows (%d pages of %d)",
				MaxLoadAllPages*MaxCursorSize,
				MaxLoadAllPages,
				MaxCursorSize,
			)
		}

		cursor := NewCursor(MaxCursorSize, key, Head, orderBy)

		rows, err := fetch(ctx, cursor)
		if err != nil {
			return nil, fmt.Errorf("cannot load all rows: %w", err)
		}

		p := NewPage(rows, cursor)
		all = append(all, p.Data...)

		if !p.Info.HasNext {
			break
		}

		k := p.Last().CursorKey(orderBy.Field)
		key = &k
	}

	return all, nil
}
