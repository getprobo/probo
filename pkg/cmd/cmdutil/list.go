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

package cmdutil

import (
	"fmt"

	"go.probo.inc/probo/pkg/cmd/iostreams"
)

// ListView describes how a list command prints one record type.
// Empty is the full sentence written when the list is empty.
// Noun is the plural name used in the truncation line.
// NextCursor is printed when another page exists. An empty cursor is omitted.
type ListView[T any] struct {
	Empty      string
	Noun       string
	Headers    []string
	Row        func(T) ([]string, error)
	NextCursor string
}

// PrintList writes items as JSON or as a table.
// JSON null slices are printed as []. When the server total is larger than
// the page, or a next cursor is set, stderr names the window and the cursor.
// Stdout stays the JSON array or the table.
func PrintList[T any](ios *iostreams.IOStreams, format string, items []T, total int, view ListView[T]) error {
	if format == OutputJSON {
		if items == nil {
			items = []T{}
		}

		if err := PrintJSON(ios.Out, items); err != nil {
			return err
		}

		WritePageFooter(ios, len(items), total, view.Noun, view.NextCursor)

		return nil
	}

	if len(items) == 0 {
		_, _ = fmt.Fprintln(ios.Out, view.Empty)

		WritePageFooter(ios, 0, total, view.Noun, view.NextCursor)

		return nil
	}

	rows := make([][]string, 0, len(items))

	for _, item := range items {
		row, err := view.Row(item)
		if err != nil {
			return err
		}

		rows = append(rows, row)
	}

	_, _ = fmt.Fprintln(ios.Out, NewTable(view.Headers...).Rows(rows...))

	WritePageFooter(ios, len(items), total, view.Noun, view.NextCursor)

	return nil
}

// WritePageFooter writes the truncation and next-cursor lines to stderr.
// A leading blank line separates them from the list. Nothing is written
// when the page is complete and nextCursor is empty.
func WritePageFooter(ios *iostreams.IOStreams, shown, total int, noun, nextCursor string) {
	if total <= shown && nextCursor == "" {
		return
	}

	_, _ = fmt.Fprintln(ios.ErrOut)

	if total > shown {
		_, _ = fmt.Fprintf(ios.ErrOut, "Showing %d of %d %s\n", shown, total, noun)
	}

	if nextCursor != "" {
		_, _ = fmt.Fprintf(ios.ErrOut, "Next cursor: %s\n", nextCursor)
	}
}
