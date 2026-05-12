// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

type (
	Paginable[T OrderField] interface {
		CursorKey(orderBy T) CursorKey
	}

	PageInfo struct {
		HasNext bool
		HasPrev bool
	}

	Page[T Paginable[U], U OrderField] struct {
		Info   *PageInfo
		Cursor *Cursor[U]
		Data   []T
	}
)

func (p *Page[T, U]) First() T {
	if len(p.Data) == 0 {
		var zero T
		return zero
	}

	return p.Data[0]
}

func (p *Page[T, U]) Last() T {
	if len(p.Data) == 0 {
		var zero T
		return zero
	}

	return p.Data[len(p.Data)-1]
}

func NewPage[T Paginable[U], U OrderField](data []T, c *Cursor[U]) *Page[T, U] {
	pi := &PageInfo{}

	if len(data) == 0 {
		return &Page[T, U]{
			Info: pi,
			Data: data,
		}
	}

	edges := data
	firstFromData := data[0]

	switch c.Position {
	case Head:
		if c.Key != nil {
			if len(edges) == c.Size+2 {
				edges = edges[1 : len(edges)-1]
			} else {
				edges = edges[1:]
			}
		} else if c.Key == nil && len(edges) == c.Size+1 {
			edges = edges[0 : len(edges)-1]
		}

		if c.Key != nil && c.Key.String() == firstFromData.CursorKey(c.OrderBy.Field).String() {
			pi.HasPrev = true
		}

		if c.Key != nil && c.Size+2 == len(data) {
			pi.HasNext = true
		} else if c.Key == nil && c.Size+1 == len(data) {
			pi.HasNext = true
		}
	case Tail:
		for i, j := 0, len(edges)-1; i < j; i, j = i+1, j-1 {
			edges[i], edges[j] = edges[j], edges[i]
		}

		if c.Key != nil {
			if len(edges) == c.Size+2 {
				edges = edges[1 : len(edges)-1]
			} else {
				edges = edges[0 : len(edges)-1]
			}
		} else if c.Key == nil && len(edges) == c.Size+1 {
			edges = edges[1:]
		}

		if c.Key != nil && c.Key.String() == firstFromData.CursorKey(c.OrderBy.Field).String() {
			pi.HasNext = true
		}

		if c.Key != nil && c.Size+2 == len(data) {
			pi.HasPrev = true
		} else if c.Key == nil && c.Size+1 == len(data) {
			pi.HasPrev = true
		}
	}

	return &Page[T, U]{
		Info:   pi,
		Cursor: c,
		Data:   edges,
	}
}
