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
	"fmt"
)

type (
	OrderField interface {
		Column() string
		fmt.Stringer
	}

	// CursorValuer maps the bound cursor placeholder into the same expression
	// as Column. The cursor key keeps the raw value; Postgres applies the
	// expression.
	CursorValuer interface {
		OrderField
		CursorValue(placeholder string) string
	}

	// NullsOrderField sets NULL placement (FIRST or LAST) for both sort
	// directions. NullsDefault leaves placement to the database default.
	NullsOrderField interface {
		OrderField
		Nulls() Nulls
	}

	// Nulls is where NULL values sort relative to every other value.
	Nulls string
)

const (
	NullsDefault Nulls = ""
	NullsFirst   Nulls = "FIRST"
	NullsLast    Nulls = "LAST"
)
