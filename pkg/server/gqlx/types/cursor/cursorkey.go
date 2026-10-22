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

func UnmarshalCursorKeyScalar(v any) (page.CursorKey, error) {
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
