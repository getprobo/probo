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

package html

import (
	"fmt"
	"strconv"

	"go.probo.inc/probo/pkg/prosemirror"
)

func renderTableCellHTML(w *Writer, n prosemirror.Node, tag string) error {
	if err := writeTableCellOpen(w, n, tag); err != nil {
		return err
	}

	if err := w.RenderChildren(n.Content); err != nil {
		return err
	}

	writeTableCellClose(w, tag)

	return nil
}

func writeTableCellOpen(w *Writer, n prosemirror.Node, tag string) error {
	attrs, err := n.TableCellAttrs()
	if err != nil {
		return fmt.Errorf("cannot render %s node: %w", tag, err)
	}

	w.WriteByte('<')
	w.WriteString(tag)

	if attrs.Colspan > 1 {
		w.WriteAttr("colspan", strconv.Itoa(attrs.Colspan))
	}

	if attrs.Rowspan > 1 {
		w.WriteAttr("rowspan", strconv.Itoa(attrs.Rowspan))
	}

	if len(attrs.Colwidth) > 0 {
		total := 0
		for _, width := range attrs.Colwidth {
			total += width
		}

		w.WriteAttr("style", "min-width: "+strconv.Itoa(total)+"px")
	}

	w.WriteByte('>')

	return nil
}

func writeTableCellClose(w *Writer, tag string) {
	w.WriteString("</")
	w.WriteString(tag)
	w.WriteByte('>')
}
