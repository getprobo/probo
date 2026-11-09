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

package markdown

import "go.probo.inc/probo/pkg/prosemirror"

func renderNode(w *Writer, r NodeRenderer, n prosemirror.Node) error {
	switch s := r.(type) {
	case docRenderer:
		return s.Render(w, n)
	case paragraphRenderer:
		return s.Render(w, n)
	case textRenderer:
		return s.Render(w, n)
	case headingRenderer:
		return s.Render(w, n)
	case blockquoteRenderer:
		return s.Render(w, n)
	case codeBlockRenderer:
		return s.Render(w, n)
	case horizontalRuleRenderer:
		return s.Render(w, n)
	case hardBreakRenderer:
		return s.Render(w, n)
	case bulletListRenderer:
		return s.Render(w, n)
	case orderedListRenderer:
		return s.Render(w, n)
	case listItemRenderer:
		return s.Render(w, n)
	case taskListRenderer:
		return s.Render(w, n)
	case taskItemRenderer:
		return s.Render(w, n)
	case tableRenderer:
		return s.Render(w, n)
	case tableRowRenderer:
		return s.Render(w, n)
	case tableCellRenderer:
		return s.Render(w, n)
	case tableHeaderRenderer:
		return s.Render(w, n)
	default:
		return r.Render(w, n)
	}
}

func openMark(w *Writer, r MarkRenderer, m prosemirror.Mark) error {
	switch s := r.(type) {
	case boldRenderer:
		return s.Open(w, m)
	case italicRenderer:
		return s.Open(w, m)
	case underlineRenderer:
		return s.Open(w, m)
	case strikeRenderer:
		return s.Open(w, m)
	case codeRenderer:
		return s.Open(w, m)
	case linkRenderer:
		return s.Open(w, m)
	default:
		return r.Open(w, m)
	}
}

func closeMark(w *Writer, r MarkRenderer, m prosemirror.Mark) error {
	switch s := r.(type) {
	case boldRenderer:
		return s.Close(w, m)
	case italicRenderer:
		return s.Close(w, m)
	case underlineRenderer:
		return s.Close(w, m)
	case strikeRenderer:
		return s.Close(w, m)
	case codeRenderer:
		return s.Close(w, m)
	case linkRenderer:
		return s.Close(w, m)
	default:
		return r.Close(w, m)
	}
}
