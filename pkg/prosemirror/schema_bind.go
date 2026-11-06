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

import "fmt"

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
	case imageNode:
		return s.Check(n, mode)
	case fileNode:
		return s.Check(n, mode)
	case mentionNodeSpec:
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

// renderHTMLByType dispatches by node type string so *HTMLWriter does not
// escape through an interface method table (map lookup + interface call).
func renderHTMLByType(w *HTMLWriter, n Node) error {
	switch n.Type {
	case NodeDoc:
		return docNode{}.RenderHTML(w, n)
	case NodeParagraph:
		return paragraphNode{}.RenderHTML(w, n)
	case NodeText:
		return textNode{}.RenderHTML(w, n)
	case NodeHeading:
		return headingNode{}.RenderHTML(w, n)
	case NodeBlockquote:
		return blockquoteNode{}.RenderHTML(w, n)
	case NodeCodeBlock:
		return codeBlockNode{}.RenderHTML(w, n)
	case NodeHorizontalRule:
		return horizontalRuleNode{}.RenderHTML(w, n)
	case NodeHardBreak:
		return hardBreakNode{}.RenderHTML(w, n)
	case NodeImage:
		return imageNode{}.RenderHTML(w, n)
	case NodeFile:
		return fileNode{}.RenderHTML(w, n)
	case NodeMention:
		return mentionNodeSpec{}.RenderHTML(w, n)
	case NodeBulletList:
		return bulletListNode{}.RenderHTML(w, n)
	case NodeOrderedList:
		return orderedListNode{}.RenderHTML(w, n)
	case NodeListItem:
		return listItemNode{}.RenderHTML(w, n)
	case NodeTaskList:
		return taskListNode{}.RenderHTML(w, n)
	case NodeTaskItem:
		return taskItemNode{}.RenderHTML(w, n)
	case NodeTable:
		return tableNode{}.RenderHTML(w, n)
	case NodeTableRow:
		return tableRowNode{}.RenderHTML(w, n)
	case NodeTableCell:
		return tableCellNode{}.RenderHTML(w, n)
	case NodeTableHeader:
		return tableHeaderNode{}.RenderHTML(w, n)
	default:
		return fmt.Errorf("cannot render node: unknown type %q", n.Type)
	}
}

func renderMarkdownByType(w *MarkdownWriter, n Node) error {
	switch n.Type {
	case NodeDoc:
		return docNode{}.RenderMarkdown(w, n)
	case NodeParagraph:
		return paragraphNode{}.RenderMarkdown(w, n)
	case NodeText:
		return textNode{}.RenderMarkdown(w, n)
	case NodeHeading:
		return headingNode{}.RenderMarkdown(w, n)
	case NodeBlockquote:
		return blockquoteNode{}.RenderMarkdown(w, n)
	case NodeCodeBlock:
		return codeBlockNode{}.RenderMarkdown(w, n)
	case NodeHorizontalRule:
		return horizontalRuleNode{}.RenderMarkdown(w, n)
	case NodeHardBreak:
		return hardBreakNode{}.RenderMarkdown(w, n)
	case NodeImage:
		return imageNode{}.RenderMarkdown(w, n)
	case NodeFile:
		return fileNode{}.RenderMarkdown(w, n)
	case NodeMention:
		return mentionNodeSpec{}.RenderMarkdown(w, n)
	case NodeBulletList:
		return bulletListNode{}.RenderMarkdown(w, n)
	case NodeOrderedList:
		return orderedListNode{}.RenderMarkdown(w, n)
	case NodeListItem:
		return listItemNode{}.RenderMarkdown(w, n)
	case NodeTaskList:
		return taskListNode{}.RenderMarkdown(w, n)
	case NodeTaskItem:
		return taskItemNode{}.RenderMarkdown(w, n)
	case NodeTable:
		return tableNode{}.RenderMarkdown(w, n)
	case NodeTableRow:
		return tableRowNode{}.RenderMarkdown(w, n)
	case NodeTableCell:
		return tableCellNode{}.RenderMarkdown(w, n)
	case NodeTableHeader:
		return tableHeaderNode{}.RenderMarkdown(w, n)
	default:
		return fmt.Errorf("cannot render node: unknown type %q", n.Type)
	}
}

func openHTMLMarkByType(w *HTMLWriter, m Mark) error {
	switch m.Type {
	case MarkStrong:
		return boldMark{}.OpenHTML(w, m)
	case MarkEm:
		return italicMark{}.OpenHTML(w, m)
	case MarkUnderline:
		return underlineMark{}.OpenHTML(w, m)
	case MarkStrike:
		return strikeMark{}.OpenHTML(w, m)
	case MarkCode:
		return codeMark{}.OpenHTML(w, m)
	case MarkLink:
		return linkMark{}.OpenHTML(w, m)
	default:
		return fmt.Errorf("cannot render mark: unknown type %q", m.Type)
	}
}

func closeHTMLMarkByType(w *HTMLWriter, m Mark) {
	switch m.Type {
	case MarkStrong:
		boldMark{}.CloseHTML(w, m)
	case MarkEm:
		italicMark{}.CloseHTML(w, m)
	case MarkUnderline:
		underlineMark{}.CloseHTML(w, m)
	case MarkStrike:
		strikeMark{}.CloseHTML(w, m)
	case MarkCode:
		codeMark{}.CloseHTML(w, m)
	case MarkLink:
		linkMark{}.CloseHTML(w, m)
	}
}

func openMarkdownMarkByType(w *MarkdownWriter, m Mark) error {
	switch m.Type {
	case MarkStrong:
		return boldMark{}.OpenMarkdown(w, m)
	case MarkEm:
		return italicMark{}.OpenMarkdown(w, m)
	case MarkUnderline:
		return underlineMark{}.OpenMarkdown(w, m)
	case MarkStrike:
		return strikeMark{}.OpenMarkdown(w, m)
	case MarkCode:
		return codeMark{}.OpenMarkdown(w, m)
	case MarkLink:
		return linkMark{}.OpenMarkdown(w, m)
	default:
		return fmt.Errorf("cannot render mark: unknown type %q", m.Type)
	}
}

func closeMarkdownMarkByType(w *MarkdownWriter, m Mark) error {
	switch m.Type {
	case MarkStrong:
		return boldMark{}.CloseMarkdown(w, m)
	case MarkEm:
		return italicMark{}.CloseMarkdown(w, m)
	case MarkUnderline:
		return underlineMark{}.CloseMarkdown(w, m)
	case MarkStrike:
		return strikeMark{}.CloseMarkdown(w, m)
	case MarkCode:
		return codeMark{}.CloseMarkdown(w, m)
	case MarkLink:
		return linkMark{}.CloseMarkdown(w, m)
	default:
		return fmt.Errorf("cannot render mark: unknown type %q", m.Type)
	}
}
