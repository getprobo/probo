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
type ListView[T any] struct {
	Empty   string
	Noun    string
	Headers []string
	Row     func(T) ([]string, error)
}

// PrintList writes items as JSON or as a table.
// JSON null slices are printed as []. A total larger than len(items) writes
// "Showing n of m <noun>" to stderr.
func PrintList[T any](ios *iostreams.IOStreams, format string, items []T, total int, view ListView[T]) error {
	if format == OutputJSON {
		if items == nil {
			items = []T{}
		}

		return PrintJSON(ios.Out, items)
	}

	if len(items) == 0 {
		_, _ = fmt.Fprintln(ios.Out, view.Empty)

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

	if total > len(items) {
		_, _ = fmt.Fprintf(ios.ErrOut, "\nShowing %d of %d %s\n", len(items), total, view.Noun)
	}

	return nil
}
