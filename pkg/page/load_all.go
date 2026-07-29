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
// rather than materialising an unbounded set. WalkAll is uncapped: it
// streams page-by-page and is meant for callers that can process rows
// without holding the full set in memory.
const MaxLoadAllPages = 20

// Loader runs one paginated query for the given cursor and returns the
// rows it loaded. Callers bind the connection, scope, parent key and
// filter in a closure, exposing only ctx and cursor (typically a coredata
// LoadBy* on a fresh receiver).
type Loader[T Paginable[U], U OrderField] func(ctx context.Context, cursor *Cursor[U]) ([]T, error)

// WalkAll walks every matching row via keyset pagination, advancing a
// MaxCursorSize forward cursor until no rows remain, and invokes walk with
// every page of rows. Unlike LoadAll, it does not apply MaxLoadAllPages.
func WalkAll[T Paginable[U], U OrderField](
	ctx context.Context,
	orderBy OrderBy[U],
	fetch Loader[T, U],
	walk func(rows []T) error,
) error {
	var key *CursorKey

	for {
		cursor := NewCursor(MaxCursorSize, key, Head, orderBy)

		rows, err := fetch(ctx, cursor)
		if err != nil {
			return fmt.Errorf("cannot load all rows: %w", err)
		}

		p := NewPage(rows, cursor)
		if err := walk(p.Data); err != nil {
			return err
		}

		if !p.Info.HasNext {
			return nil
		}

		k := p.Last().CursorKey(orderBy.Field)
		key = &k
	}
}

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
		all   []T
		pages int
	)

	err := WalkAll(
		ctx,
		orderBy,
		func(ctx context.Context, cursor *Cursor[U]) ([]T, error) {
			if pages >= MaxLoadAllPages {
				return nil, fmt.Errorf(
					"cannot load all rows: result set exceeds %d rows (%d pages of %d)",
					MaxLoadAllPages*MaxCursorSize,
					MaxLoadAllPages,
					MaxCursorSize,
				)
			}

			pages++

			return fetch(ctx, cursor)
		},
		func(rows []T) error {
			all = append(all, rows...)
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return all, nil
}
