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
