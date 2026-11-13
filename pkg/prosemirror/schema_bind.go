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

package prosemirror

func checkNodeSpec(spec NodeSpec, n *Node, mode CheckMode) error {
	switch s := spec.(type) {
	case docNode:
		return s.Check(n, mode)
	case paragraphNode:
		return s.Check(n, mode)
	case textNode:
		return s.Check(n, mode)
	case headingNode:
		return s.Check(n, mode)
	case blockquoteNode:
		return s.Check(n, mode)
	case codeBlockNode:
		return s.Check(n, mode)
	case horizontalRuleNode:
		return s.Check(n, mode)
	case hardBreakNode:
		return s.Check(n, mode)
	case bulletListNode:
		return s.Check(n, mode)
	case orderedListNode:
		return s.Check(n, mode)
	case listItemNode:
		return s.Check(n, mode)
	case taskListNode:
		return s.Check(n, mode)
	case taskItemNode:
		return s.Check(n, mode)
	case tableNode:
		return s.Check(n, mode)
	case tableRowNode:
		return s.Check(n, mode)
	case tableCellNode:
		return s.Check(n, mode)
	case tableHeaderNode:
		return s.Check(n, mode)
	default:
		return spec.Check(n, mode)
	}
}

func checkMarkSpec(spec MarkSpec, m *Mark, mode CheckMode) error {
	switch s := spec.(type) {
	case boldMark:
		return s.Check(m, mode)
	case italicMark:
		return s.Check(m, mode)
	case underlineMark:
		return s.Check(m, mode)
	case strikeMark:
		return s.Check(m, mode)
	case codeMark:
		return s.Check(m, mode)
	case linkMark:
		return s.Check(m, mode)
	default:
		return spec.Check(m, mode)
	}
}
