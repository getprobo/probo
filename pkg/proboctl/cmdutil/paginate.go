// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package cmdutil

import (
	"context"

	"go.probo.inc/probo/pkg/page"
)

// paginatePageSize is the per-request fetch size used to walk a cursor
// connection. It is independent of the caller's --limit, which only
// bounds the total returned.
const paginatePageSize = 100

// Paginate walks a cursor-paginated coredata Load into a single slice,
// following the keyset cursor until the caller's limit is reached or the
// connection is exhausted. A limit <= 0 returns every matching row. This
// is the proboctl-side counterpart of the API's connection resolvers: it
// drives page.NewCursor + page.NewPage against the same coredata Load
// methods, so a future proboctl API can reuse the data layer untouched.
func Paginate[E page.Paginable[F], F page.OrderField](
	ctx context.Context,
	orderBy page.OrderBy[F],
	limit int,
	load func(ctx context.Context, cursor *page.Cursor[F]) ([]E, error),
) ([]E, error) {
	var (
		result []E
		from   *page.CursorKey
	)

	for {
		size := paginatePageSize
		if limit > 0 {
			remaining := limit - len(result)
			if remaining <= 0 {
				break
			}

			if remaining < size {
				size = remaining
			}
		}

		cursor := page.NewCursor(size, from, page.Head, orderBy)

		rows, err := load(ctx, cursor)
		if err != nil {
			return nil, err
		}

		p := page.NewPage(rows, cursor)
		result = append(result, p.Data...)

		if !p.Info.HasNext || len(p.Data) == 0 {
			break
		}

		key := p.Data[len(p.Data)-1].CursorKey(orderBy.Field)
		from = &key
	}

	return result, nil
}
