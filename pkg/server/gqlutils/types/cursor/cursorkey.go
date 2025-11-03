// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package cursor

import (
	"errors"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"go.probo.inc/probo/pkg/page"
)

type CursorKeyScalar = page.CursorKey

func NewCursor[O page.OrderField](
	first *int,
	after *page.CursorKey,
	last *int,
	before *page.CursorKey,
	orderBy page.OrderBy[O],
) *page.Cursor[O] {
	var (
		size      int
		from      *page.CursorKey
		direction = page.Head
	)

	if first != nil {
		size = *first
		direction = page.Head
		from = after
	} else if last != nil {
		size = *last
		direction = page.Tail
		from = before
	}

	return page.NewCursor(size, from, direction, orderBy)
}

func MarshalCursorKeyScalar(ck page.CursorKey) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = w.Write([]byte(strconv.Quote(ck.String())))
	})
}

func UnmarshalCursorKeyScalar(v interface{}) (page.CursorKey, error) {
	s, ok := v.(string)
	if !ok {
		return page.CursorKeyNil, errors.New("must be a string")
	}

	ck, err := page.ParseCursorKey(s)
	if err != nil {
		return page.CursorKeyNil, err
	}

	return ck, nil
}
