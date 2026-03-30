// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package prosemirror

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"
)

// htmlBlockSanitizePolicy matches tags and attributes we can represent as
// ProseMirror/Tiptap JSON and strips scripts, event handlers, and unsafe URLs.
func htmlBlockSanitizePolicy() *bluemonday.Policy {
	return bluemonday.UGCPolicy()
}

func sanitizeHTMLBlockContent(s string) string {
	return htmlBlockSanitizePolicy().Sanitize(s)
}

// convertProseMirrorFromInlineHTML sanitizes inline raw HTML and maps it to
// ProseMirror paragraph-level children (text, hardBreak, image, marks).
func convertProseMirrorFromInlineHTML(raw string) ([]Node, error) {
	sanitized := strings.TrimSpace(sanitizeHTMLBlockContent(raw))
	if sanitized == "" {
		return nil, nil
	}

	roots, err := parseHTMLFragmentRoots(sanitized)
	if err != nil {
		return nil, fmt.Errorf("cannot convert inline html: %w", err)
	}

	c := &htmlBlockConverter{}
	var out []Node
	for _, root := range roots {
		nodes, err := c.convertInlineNode(root)
		if err != nil {
			return nil, err
		}
		out = append(out, nodes...)
	}
	if len(out) > 0 {
		return out, nil
	}

	plain := strings.TrimSpace(plainTextFromHTMLFragment(sanitized))
	if plain == "" {
		return nil, nil
	}
	return []Node{{Type: NodeText, Text: &plain}}, nil
}

func convertProseMirrorFromHTMLBlock(raw string) ([]Node, error) {
	sanitized := strings.TrimSpace(sanitizeHTMLBlockContent(raw))
	if sanitized == "" {
		return nil, nil
	}

	nodes, err := htmlFragmentToProseMirrorBlocks(sanitized)
	if err != nil {
		return nil, fmt.Errorf("cannot convert html block to prosemirror: %w", err)
	}
	if len(nodes) > 0 {
		return nodes, nil
	}

	plain := strings.TrimSpace(plainTextFromHTMLFragment(sanitized))
	if plain == "" {
		return nil, nil
	}
	return []Node{paragraphWithPlainText(plain)}, nil
}

func paragraphWithPlainText(s string) Node {
	return Node{
		Type: NodeParagraph,
		Content: []Node{{
			Type: NodeText,
			Text: &s,
		}},
	}
}

func parseHTMLFragmentRoots(htmlStr string) ([]*html.Node, error) {
	return html.ParseFragmentWithOptions(
		strings.NewReader(htmlStr),
		nil,
		html.ParseOptionEnableScripting(false),
	)
}

func plainTextFromHTMLFragment(htmlStr string) string {
	roots, err := parseHTMLFragmentRoots(htmlStr)
	if err != nil {
		return ""
	}
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	for _, root := range roots {
		walk(root)
	}
	return b.String()
}

func htmlFragmentToProseMirrorBlocks(htmlStr string) ([]Node, error) {
	roots, err := parseHTMLFragmentRoots(htmlStr)
	if err != nil {
		return nil, err
	}

	c := &htmlBlockConverter{}
	var out []Node
	for _, root := range roots {
		nodes, err := c.convertTopLevel(root)
		if err != nil {
			return nil, err
		}
		out = append(out, nodes...)
	}
	return out, nil
}

type htmlBlockConverter struct {
	marks []Mark
}

func (c *htmlBlockConverter) convertTopLevel(n *html.Node) ([]Node, error) {
	switch n.Type {
	case html.TextNode:
		t := strings.TrimSpace(n.Data)
		if t == "" {
			return nil, nil
		}
		return []Node{paragraphWithPlainText(t)}, nil
	case html.ElementNode:
		return c.convertBlockElement(n)
	default:
		return nil, nil
	}
}

func (c *htmlBlockConverter) convertBlockElement(n *html.Node) ([]Node, error) {
	switch n.Data {
	case "p":
		inlines, err := c.convertInlineFragments(n)
		if err != nil {
			return nil, err
		}
		return []Node{{Type: NodeParagraph, Content: inlines}}, nil
	case "h1", "h2", "h3", "h4", "h5", "h6":
		level := int(n.Data[1] - '0')
		inlines, err := c.convertInlineFragments(n)
		if err != nil {
			return nil, err
		}
		attrs, err := json.Marshal(HeadingAttrs{Level: level})
		if err != nil {
			return nil, fmt.Errorf("cannot marshal heading attrs: %w", err)
		}
		return []Node{{
			Type:    NodeHeading,
			Attrs:   attrs,
			Content: inlines,
		}}, nil
	case "blockquote":
		return c.convertBlockquote(n)
	case "pre":
		return c.convertPre(n)
	case "hr":
		return []Node{{Type: NodeHorizontalRule}}, nil
	case "ul":
		return c.convertList(n, false)
	case "ol":
		return c.convertList(n, true)
	case "table":
		return c.convertTable(n)
	case "br":
		return []Node{{
			Type:    NodeParagraph,
			Content: []Node{{Type: NodeHardBreak}},
		}}, nil
	case "img":
		img, err := c.convertImageElement(n)
		if err != nil {
			return nil, err
		}
		if img == nil {
			return nil, nil
		}
		return []Node{{
			Type:    NodeParagraph,
			Content: []Node{*img},
		}}, nil
	case "div", "section", "article", "aside", "main", "header", "footer", "nav",
		"center", "figure", "body", "html", "span":
		return c.unwrapBlockElement(n)
	default:
		return c.unwrapBlockElement(n)
	}
}

func (c *htmlBlockConverter) unwrapBlockElement(n *html.Node) ([]Node, error) {
	return c.convertBlockChildren(n)
}

func (c *htmlBlockConverter) convertBlockChildren(n *html.Node) ([]Node, error) {
	var out []Node
	for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
		nodes, err := c.convertTopLevel(ch)
		if err != nil {
			return nil, err
		}
		out = append(out, nodes...)
	}
	return out, nil
}

func (c *htmlBlockConverter) convertBlockquote(n *html.Node) ([]Node, error) {
	if hasBlockElementChild(n) {
		inner, err := c.convertBlockChildren(n)
		if err != nil {
			return nil, err
		}
		return []Node{{Type: NodeBlockquote, Content: inner}}, nil
	}
	inlines, err := c.convertInlineFragments(n)
	if err != nil {
		return nil, err
	}
	var content []Node
	if len(inlines) > 0 {
		content = []Node{{Type: NodeParagraph, Content: inlines}}
	}
	return []Node{{Type: NodeBlockquote, Content: content}}, nil
}

func (c *htmlBlockConverter) convertPre(n *html.Node) ([]Node, error) {
	var lang *string
	var textBuf strings.Builder
	for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
		if ch.Type == html.ElementNode && ch.Data == "code" {
			lang = codeLanguageFromClass(attrVal(ch, "class"))
			var walkText func(*html.Node)
			walkText = func(x *html.Node) {
				if x.Type == html.TextNode {
					textBuf.WriteString(x.Data)
				}
				for cc := x.FirstChild; cc != nil; cc = cc.NextSibling {
					walkText(cc)
				}
			}
			walkText(ch)
			break
		}
	}
	if textBuf.Len() == 0 {
		var walkText func(*html.Node)
		walkText = func(x *html.Node) {
			if x.Type == html.TextNode {
				textBuf.WriteString(x.Data)
			}
			for cc := x.FirstChild; cc != nil; cc = cc.NextSibling {
				walkText(cc)
			}
		}
		walkText(n)
	}

	content := textBuf.String()
	attrs, err := json.Marshal(CodeBlockAttrs{Language: lang})
	if err != nil {
		return nil, fmt.Errorf("cannot marshal code block attrs: %w", err)
	}

	var textNodes []Node
	if content != "" {
		textNodes = []Node{{Type: NodeText, Text: &content}}
	}
	return []Node{{
		Type:    NodeCodeBlock,
		Attrs:   attrs,
		Content: textNodes,
	}}, nil
}

func codeLanguageFromClass(class string) *string {
	const prefix = "language-"
	for _, part := range strings.Fields(class) {
		if strings.HasPrefix(part, prefix) {
			lang := strings.TrimPrefix(part, prefix)
			if lang != "" {
				return &lang
			}
		}
	}
	return nil
}

func (c *htmlBlockConverter) convertList(n *html.Node, ordered bool) ([]Node, error) {
	var items []Node
	for li := n.FirstChild; li != nil; li = li.NextSibling {
		if li.Type != html.ElementNode || li.Data != "li" {
			continue
		}
		body, err := c.convertListItem(li)
		if err != nil {
			return nil, err
		}
		if len(body) == 0 {
			continue
		}
		items = append(items, Node{Type: NodeListItem, Content: body})
	}
	if len(items) == 0 {
		return nil, nil
	}
	if ordered {
		start := parseOlStart(n)
		attrs, err := json.Marshal(OrderedListAttrs{Start: start})
		if err != nil {
			return nil, fmt.Errorf("cannot marshal ordered list attrs: %w", err)
		}
		return []Node{{
			Type:    NodeOrderedList,
			Attrs:   attrs,
			Content: items,
		}}, nil
	}
	return []Node{{
		Type:    NodeBulletList,
		Content: items,
	}}, nil
}

func parseOlStart(n *html.Node) int {
	s := attrVal(n, "start")
	if s == "" {
		return 1
	}
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || v < 1 {
		return 1
	}
	return v
}

func (c *htmlBlockConverter) convertListItem(li *html.Node) ([]Node, error) {
	if hasBlockElementChild(li) {
		return c.convertBlockChildren(li)
	}
	inlines, err := c.convertInlineFragments(li)
	if err != nil {
		return nil, err
	}
	if len(inlines) == 0 {
		return nil, nil
	}
	return []Node{{Type: NodeParagraph, Content: inlines}}, nil
}

func hasBlockElementChild(n *html.Node) bool {
	for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
		if ch.Type == html.ElementNode && blockTagName(ch.Data) {
			return true
		}
	}
	return false
}

func blockTagName(name string) bool {
	switch name {
	case "p", "div", "blockquote", "pre", "ul", "ol", "table",
		"h1", "h2", "h3", "h4", "h5", "h6", "hr":
		return true
	default:
		return false
	}
}

func (c *htmlBlockConverter) convertTable(n *html.Node) ([]Node, error) {
	var rows []*html.Node
	var collect func(*html.Node)
	collect = func(x *html.Node) {
		if x.Type == html.ElementNode && x.Data == "tr" {
			rows = append(rows, x)
			return
		}
		for ch := x.FirstChild; ch != nil; ch = ch.NextSibling {
			collect(ch)
		}
	}
	collect(n)

	var rowNodes []Node
	for _, tr := range rows {
		row, err := c.convertTableRow(tr)
		if err != nil {
			return nil, err
		}
		if row != nil {
			rowNodes = append(rowNodes, *row)
		}
	}
	if len(rowNodes) == 0 {
		return nil, nil
	}
	return []Node{{Type: NodeTable, Content: rowNodes}}, nil
}

func (c *htmlBlockConverter) convertTableRow(tr *html.Node) (*Node, error) {
	var cells []Node
	for ch := tr.FirstChild; ch != nil; ch = ch.NextSibling {
		if ch.Type != html.ElementNode {
			continue
		}
		var cell *Node
		var err error
		switch ch.Data {
		case "th":
			cell, err = c.convertTableCell(ch, NodeTableHeader)
		case "td":
			cell, err = c.convertTableCell(ch, NodeTableCell)
		default:
			continue
		}
		if err != nil {
			return nil, err
		}
		if cell != nil {
			cells = append(cells, *cell)
		}
	}
	if len(cells) == 0 {
		return nil, nil
	}
	return &Node{Type: NodeTableRow, Content: cells}, nil
}

func (c *htmlBlockConverter) convertTableCell(n *html.Node, typ NodeType) (*Node, error) {
	attrs, err := json.Marshal(TableCellAttrs{})
	if err != nil {
		return nil, fmt.Errorf("cannot marshal table cell attrs: %w", err)
	}
	inlines, err := c.convertInlineFragments(n)
	if err != nil {
		return nil, err
	}
	content := []Node{{Type: NodeParagraph, Content: inlines}}
	return &Node{Type: typ, Attrs: attrs, Content: content}, nil
}

func (c *htmlBlockConverter) convertInlineFragments(parent *html.Node) ([]Node, error) {
	var out []Node
	for ch := parent.FirstChild; ch != nil; ch = ch.NextSibling {
		nodes, err := c.convertInlineNode(ch)
		if err != nil {
			return nil, err
		}
		out = append(out, nodes...)
	}
	return out, nil
}

func (c *htmlBlockConverter) convertInlineNode(n *html.Node) ([]Node, error) {
	switch n.Type {
	case html.TextNode:
		if n.Data == "" {
			return nil, nil
		}
		t := n.Data
		return []Node{{
			Type:  NodeText,
			Text:  &t,
			Marks: copyMarks(c.marks),
		}}, nil
	case html.ElementNode:
		return c.convertInlineElement(n)
	default:
		return nil, nil
	}
}

func (c *htmlBlockConverter) convertInlineElement(n *html.Node) ([]Node, error) {
	switch n.Data {
	case "p", "div", "section", "article", "aside", "main", "header", "footer", "nav",
		"center", "figure", "h1", "h2", "h3", "h4", "h5", "h6":
		// Block-ish tags inside inline HTML: unwrap to children only.
		return c.convertInlineFragments(n)
	case "br":
		return []Node{{Type: NodeHardBreak}}, nil
	case "strong", "b":
		return c.withMark(Mark{Type: MarkStrong}, n)
	case "em", "i":
		return c.withMark(Mark{Type: MarkEm}, n)
	case "s", "strike", "del":
		return c.withMark(Mark{Type: MarkStrike}, n)
	case "u":
		return c.withMark(Mark{Type: MarkUnderline}, n)
	case "code":
		return c.withMark(Mark{Type: MarkCode}, n)
	case "a":
		return c.convertAnchor(n)
	case "img":
		img, err := c.convertImageElement(n)
		if err != nil || img == nil {
			return nil, err
		}
		return []Node{*img}, nil
	case "span":
		return c.convertInlineFragments(n)
	default:
		return c.convertInlineFragments(n)
	}
}

func (c *htmlBlockConverter) withMark(m Mark, n *html.Node) ([]Node, error) {
	c.marks = append(c.marks, m)
	nodes, err := c.convertInlineFragments(n)
	c.marks = c.marks[:len(c.marks)-1]
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func (c *htmlBlockConverter) convertAnchor(n *html.Node) ([]Node, error) {
	href := attrVal(n, "href")
	if href == "" {
		return c.convertInlineFragments(n)
	}
	var title *string
	if t := attrVal(n, "title"); t != "" {
		title = &t
	}
	attrs, err := json.Marshal(LinkAttrs{Href: href, Title: title})
	if err != nil {
		return nil, fmt.Errorf("cannot marshal link attrs: %w", err)
	}
	m := Mark{Type: MarkLink, Attrs: attrs}
	c.marks = append(c.marks, m)
	nodes, err := c.convertInlineFragments(n)
	c.marks = c.marks[:len(c.marks)-1]
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func (c *htmlBlockConverter) convertImageElement(n *html.Node) (*Node, error) {
	src := attrVal(n, "src")
	if src == "" {
		return nil, nil
	}
	var alt, title *string
	if a := attrVal(n, "alt"); a != "" {
		alt = &a
	}
	if t := attrVal(n, "title"); t != "" {
		title = &t
	}
	attrs, err := json.Marshal(ImageAttrs{Src: src, Alt: alt, Title: title})
	if err != nil {
		return nil, fmt.Errorf("cannot marshal image attrs: %w", err)
	}
	return &Node{Type: NodeImage, Attrs: attrs}, nil
}

func attrVal(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}
