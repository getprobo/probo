// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
