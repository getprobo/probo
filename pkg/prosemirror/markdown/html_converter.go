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

import (
	"encoding/json"
	"fmt"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/prosemirror/internal/urlsafe"
	"regexp"
	"strconv"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"
)

// htmlBlockSanitizePolicy matches tags and attributes we can represent as
// ProseMirror/Tiptap JSON and strips scripts, event handlers, and unsafe URLs.
var htmlBlockSanitizePolicy = func() *bluemonday.Policy {
	policy := bluemonday.UGCPolicy()
	policy.AllowAttrs("width").OnElements("img")
	// Checkbox inputs are the HTML form of a task list. Only these attributes
	// are kept, and the converter stores them as a checked flag rather than HTML.
	policy.AllowAttrs("type", "checked", "disabled").OnElements("input")
	policy.AllowAttrs("data-type", "data-checked").OnElements("ul", "ol", "li")
	// An ordered list may start at zero. Keep that index so a later plain
	// group in a mixed task list keeps the source numbering.
	policy.AllowAttrs("start").Matching(bluemonday.Integer).OnElements("ol")

	return policy
}()

func sanitizeHTMLBlockContent(s string) string {
	return htmlBlockSanitizePolicy.Sanitize(s)
}

// convertProseMirrorFromInlineHTML sanitizes inline raw HTML and maps it to
// ProseMirror inline children (text, hardBreak, marks) and image blocks.
func convertProseMirrorFromInlineHTML(raw string) ([]prosemirror.Node, error) {
	sanitized := strings.TrimSpace(sanitizeHTMLBlockContent(raw))
	if sanitized == "" {
		return nil, nil
	}

	roots, err := parseHTMLFragmentRoots(sanitized)
	if err != nil {
		return nil, fmt.Errorf("cannot convert inline html: %w", err)
	}

	c := &htmlBlockConverter{}

	var out []prosemirror.Node

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

	plain := strings.TrimSpace(textFromHTMLRoots(roots))
	if plain == "" {
		return nil, nil
	}

	return []prosemirror.Node{{Type: prosemirror.NodeText, Text: plain}}, nil
}

func convertProseMirrorFromHTMLBlock(raw string) ([]prosemirror.Node, error) {
	sanitized := strings.TrimSpace(sanitizeHTMLBlockContent(raw))
	if sanitized == "" {
		return nil, nil
	}

	nodes, plain, err := htmlFragmentToProseMirrorBlocks(sanitized)
	if err != nil {
		return nil, fmt.Errorf("cannot convert html block to prosemirror: %w", err)
	}

	if len(nodes) > 0 {
		return nodes, nil
	}

	plain = strings.TrimSpace(plain)
	if plain == "" {
		return nil, nil
	}

	return []prosemirror.Node{paragraphWithPlainText(plain)}, nil
}

func paragraphWithPlainText(s string) prosemirror.Node {
	return prosemirror.Node{
		Type: prosemirror.NodeParagraph,
		Content: []prosemirror.Node{{
			Type: prosemirror.NodeText,
			Text: s,
		}},
	}
}

// The HTML fragment parser panics on input elements when the fragment has no
// context node. Checkbox inputs are rewritten to spans before parsing.
var checkboxInputTag = regexp.MustCompile(`(?i)<input\b[^>]*>`)

func rewriteCheckboxInputs(s string) string {
	return checkboxInputTag.ReplaceAllStringFunc(s, func(tag string) string {
		attrs := parseTagAttrs(tag)
		if !strings.EqualFold(attrs["type"], "checkbox") {
			return ""
		}

		if _, checked := attrs["checked"]; checked {
			return `<span data-task-checkbox="true"></span>`
		}

		return `<span data-task-checkbox="false"></span>`
	})
}

// parseTagAttrs reads the attributes of a single start tag, such as
// "<input type=\"checkbox\" checked>". It returns nil if tag is not a
// well-formed start tag.
func parseTagAttrs(tag string) map[string]string {
	tokenizer := html.NewTokenizer(strings.NewReader(tag))
	if tokenizer.Next() != html.StartTagToken {
		return nil
	}

	token := tokenizer.Token()

	attrs := make(map[string]string, len(token.Attr))
	for _, attr := range token.Attr {
		attrs[strings.ToLower(attr.Key)] = attr.Val
	}

	return attrs
}

func parseHTMLFragmentRoots(htmlStr string) ([]*html.Node, error) {
	return html.ParseFragmentWithOptions(
		strings.NewReader(rewriteCheckboxInputs(htmlStr)),
		nil,
		html.ParseOptionEnableScripting(false),
	)
}

func textFromHTMLRoots(roots []*html.Node) string {
	var (
		b    strings.Builder
		walk func(*html.Node)
	)

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

func htmlFragmentToProseMirrorBlocks(htmlStr string) ([]prosemirror.Node, string, error) {
	roots, err := parseHTMLFragmentRoots(htmlStr)
	if err != nil {
		return nil, "", err
	}

	c := &htmlBlockConverter{}

	var out []prosemirror.Node

	for _, root := range roots {
		nodes, err := c.convertTopLevel(root)
		if err != nil {
			return nil, "", err
		}

		out = append(out, nodes...)
	}

	return out, textFromHTMLRoots(roots), nil
}

type htmlBlockConverter struct {
	marks []prosemirror.Mark
}

func (c *htmlBlockConverter) convertTopLevel(n *html.Node) ([]prosemirror.Node, error) {
	switch n.Type {
	case html.TextNode:
		t := strings.TrimSpace(n.Data)
		if t == "" {
			return nil, nil
		}

		return []prosemirror.Node{paragraphWithPlainText(t)}, nil
	case html.ElementNode:
		return c.convertBlockElement(n)
	default:
		return nil, nil
	}
}

func (c *htmlBlockConverter) convertBlockElement(n *html.Node) ([]prosemirror.Node, error) {
	switch n.Data {
	case "p":
		inlines, err := c.convertInlineFragments(n)
		if err != nil {
			return nil, err
		}

		return blocksFromInlines(prosemirror.NodeParagraph, nil, inlines), nil
	case "h1", "h2", "h3", "h4", "h5", "h6":
		level := int(n.Data[1] - '0')

		inlines, err := c.convertInlineFragments(n)
		if err != nil {
			return nil, err
		}

		attrs, err := json.Marshal(prosemirror.HeadingAttrs{Level: level})
		if err != nil {
			return nil, fmt.Errorf("cannot marshal heading attrs: %w", err)
		}

		return blocksFromInlines(prosemirror.NodeHeading, attrs, inlines), nil
	case "blockquote":
		return c.convertBlockquote(n)
	case "pre":
		return c.convertPre(n)
	case "hr":
		return []prosemirror.Node{{Type: prosemirror.NodeHorizontalRule}}, nil
	case "ul":
		return c.convertList(n, false)
	case "ol":
		return c.convertList(n, true)
	case "table":
		return c.convertTable(n)
	case "br":
		return []prosemirror.Node{{
			Type:    prosemirror.NodeParagraph,
			Content: []prosemirror.Node{{Type: prosemirror.NodeHardBreak}},
		}}, nil
	case "img":
		img, err := c.convertImageElement(n)
		if err != nil {
			return nil, err
		}

		if img == nil {
			return nil, nil
		}

		return []prosemirror.Node{*img}, nil
	case "div", "section", "article", "aside", "main", "header", "footer", "nav",
		"center", "figure", "body", "html", "span":
		return c.unwrapBlockElement(n)
	default:
		return c.unwrapBlockElement(n)
	}
}

func (c *htmlBlockConverter) unwrapBlockElement(n *html.Node) ([]prosemirror.Node, error) {
	if !hasBlockElementChild(n) {
		inlines, err := c.convertInlineFragments(n)
		if err != nil {
			return nil, err
		}

		if len(inlines) == 0 {
			return nil, nil
		}

		return blocksFromInlines(prosemirror.NodeParagraph, nil, inlines), nil
	}

	return c.convertBlockChildren(n)
}

func (c *htmlBlockConverter) convertBlockChildren(n *html.Node) ([]prosemirror.Node, error) {
	var out []prosemirror.Node

	for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
		nodes, err := c.convertTopLevel(ch)
		if err != nil {
			return nil, err
		}

		out = append(out, nodes...)
	}

	return out, nil
}

func (c *htmlBlockConverter) convertBlockquote(n *html.Node) ([]prosemirror.Node, error) {
	if hasBlockElementChild(n) {
		inner, err := c.convertBlockChildren(n)
		if err != nil {
			return nil, err
		}

		return []prosemirror.Node{{Type: prosemirror.NodeBlockquote, Content: inner}}, nil
	}

	inlines, err := c.convertInlineFragments(n)
	if err != nil {
		return nil, err
	}

	var content []prosemirror.Node
	if len(inlines) > 0 {
		content = blocksFromInlines(prosemirror.NodeParagraph, nil, inlines)
	}

	return []prosemirror.Node{{Type: prosemirror.NodeBlockquote, Content: content}}, nil
}

func (c *htmlBlockConverter) convertPre(n *html.Node) ([]prosemirror.Node, error) {
	var (
		lang    *string
		textBuf strings.Builder
	)

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

	attrs, err := json.Marshal(prosemirror.CodeBlockAttrs{Language: lang})
	if err != nil {
		return nil, fmt.Errorf("cannot marshal code block attrs: %w", err)
	}

	var textNodes []prosemirror.Node
	if content != "" {
		textNodes = []prosemirror.Node{{Type: prosemirror.NodeText, Text: content}}
	}

	return []prosemirror.Node{{
		Type:    prosemirror.NodeCodeBlock,
		Attrs:   attrs,
		Content: textNodes,
	}}, nil
}

func codeLanguageFromClass(class string) *string {
	const prefix = "language-"
	for part := range strings.FieldsSeq(class) {
		if after, ok := strings.CutPrefix(part, prefix); ok {
			lang := after
			if lang != "" {
				return &lang
			}
		}
	}

	return nil
}

func (c *htmlBlockConverter) convertList(n *html.Node, ordered bool) ([]prosemirror.Node, error) {
	items, err := c.htmlListItems(n)
	if err != nil {
		return nil, err
	}

	groups := htmlListGroups(items, ordered)
	if len(groups) == 0 {
		return nil, nil
	}

	return htmlListsForGroups(n, ordered, groups)
}

type htmlListItem struct {
	node    prosemirror.Node
	ordinal int
}

func (c *htmlBlockConverter) htmlListItems(n *html.Node) ([]htmlListItem, error) {
	var items []htmlListItem

	ordinal := 0

	for li := n.FirstChild; li != nil; li = li.NextSibling {
		if li.Type != html.ElementNode || li.Data != "li" {
			continue
		}

		item, ok, err := c.htmlListItem(li)
		if err != nil {
			return nil, err
		}

		if !ok {
			ordinal++

			continue
		}

		items = append(items, htmlListItem{node: item, ordinal: ordinal})
		ordinal++
	}

	return items, nil
}

func htmlListGroups(items []htmlListItem, ordered bool) [][]htmlListItem {
	var groups [][]htmlListItem

	for _, item := range items {
		if len(groups) == 0 || startsHTMLListGroup(groups[len(groups)-1], item, ordered) {
			groups = append(groups, []htmlListItem{item})

			continue
		}

		last := len(groups) - 1
		groups[last] = append(groups[last], item)
	}

	return groups
}

func startsHTMLListGroup(group []htmlListItem, item htmlListItem, ordered bool) bool {
	last := group[len(group)-1]
	if last.node.Type != item.node.Type {
		return true
	}

	// An empty plain <li> is dropped. When it created a numbering gap in an
	// ordered list, start another plain run so its start attribute preserves
	// the source's later ordinal.
	return ordered && item.node.Type == prosemirror.NodeListItem && item.ordinal != last.ordinal+1
}

func (c *htmlBlockConverter) htmlListItem(li *html.Node) (prosemirror.Node, bool, error) {
	checked, task := taskItemState(li)
	stripLeadingCheckbox(li)

	body, err := c.convertListItem(li)
	if err != nil {
		return prosemirror.Node{}, false, err
	}

	if len(body) == 0 {
		// A task item that only held its checkbox has no body left
		// once stripLeadingCheckbox removes it, but its checked state
		// still needs a home: keep it as an empty paragraph instead of
		// dropping the item like a genuinely empty, checkbox-less <li>.
		if !task {
			return prosemirror.Node{}, false, nil
		}

		body = []prosemirror.Node{{Type: prosemirror.NodeParagraph}}
	}

	if task {
		trimLeadingTaskSpace(body)
	}

	item := prosemirror.Node{Type: prosemirror.NodeListItem, Content: body}
	if !task {
		return item, true, nil
	}

	attrs, err := json.Marshal(prosemirror.TaskItemAttrs{Checked: checked})
	if err != nil {
		return prosemirror.Node{}, false, fmt.Errorf("cannot marshal task item attrs: %w", err)
	}

	item.Type = prosemirror.NodeTaskItem
	item.Attrs = attrs

	return item, true, nil
}

func htmlListsForGroups(n *html.Node, ordered bool, groups [][]htmlListItem) ([]prosemirror.Node, error) {
	var nodes []prosemirror.Node

	start := 1

	if ordered {
		start = parseOlStart(n)
	}

	for _, group := range groups {
		items := make([]prosemirror.Node, 0, len(group))
		for _, item := range group {
			items = append(items, item.node)
		}

		if group[0].node.Type == prosemirror.NodeTaskItem {
			nodes = append(nodes, prosemirror.Node{Type: prosemirror.NodeTaskList, Content: items})

			continue
		}

		if ordered {
			attrs, err := json.Marshal(prosemirror.OrderedListAttrs{Start: start + group[0].ordinal})
			if err != nil {
				return nil, fmt.Errorf("cannot marshal ordered list attrs: %w", err)
			}

			nodes = append(nodes, prosemirror.Node{
				Type:    prosemirror.NodeOrderedList,
				Attrs:   attrs,
				Content: items,
			})
		} else {
			nodes = append(nodes, prosemirror.Node{Type: prosemirror.NodeBulletList, Content: items})
		}
	}

	return nodes, nil
}

func trimLeadingTaskSpace(nodes []prosemirror.Node) {
	if len(nodes) == 0 || nodes[0].Type != prosemirror.NodeParagraph || len(nodes[0].Content) == 0 {
		return
	}

	first := &nodes[0].Content[0]
	if first.Type != prosemirror.NodeText || first.Text == "" {
		return
	}

	trimmed := strings.TrimPrefix(first.Text, " ")
	if trimmed == "" {
		nodes[0].Content = nodes[0].Content[1:]
		return
	}

	first.Text = trimmed
}

func taskItemState(li *html.Node) (bool, bool) {
	task := attrVal(li, "data-type") == "taskItem"
	checked := attrVal(li, "data-checked") == "true"

	if input := leadingCheckbox(li); input != nil {
		task = true

		if taskCheckboxChecked(input) {
			checked = true
		}
	}

	return checked, task
}

func stripLeadingCheckbox(li *html.Node) {
	taskItem := attrVal(li, "data-type") == "taskItem"

	for ch := li.FirstChild; ch != nil; ch = ch.NextSibling {
		if ch.Type == html.TextNode && strings.TrimSpace(ch.Data) == "" {
			continue
		}

		if isTaskCheckbox(ch) {
			li.RemoveChild(ch)
			return
		}

		if ch.Type == html.ElementNode && ch.Data == "label" && findCheckbox(ch) != nil {
			if taskItem || strings.TrimSpace(htmlText(ch)) == "" {
				li.RemoveChild(ch)
				return
			}

			removeCheckboxInputs(ch)

			return
		}

		return
	}
}

func leadingCheckbox(li *html.Node) *html.Node {
	for ch := li.FirstChild; ch != nil; ch = ch.NextSibling {
		if ch.Type == html.TextNode && strings.TrimSpace(ch.Data) == "" {
			continue
		}

		if isTaskCheckbox(ch) {
			return ch
		}

		if ch.Type == html.ElementNode && ch.Data == "label" {
			if input := findCheckbox(ch); input != nil {
				return input
			}
		}

		return nil
	}

	return nil
}

func findCheckbox(n *html.Node) *html.Node {
	if isTaskCheckbox(n) {
		return n
	}

	for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
		if input := findCheckbox(ch); input != nil {
			return input
		}
	}

	return nil
}

func removeCheckboxInputs(n *html.Node) {
	for ch := n.FirstChild; ch != nil; {
		next := ch.NextSibling
		if isTaskCheckbox(ch) {
			n.RemoveChild(ch)
		} else {
			removeCheckboxInputs(ch)
		}

		ch = next
	}
}

func isTaskCheckbox(n *html.Node) bool {
	if n == nil || n.Type != html.ElementNode {
		return false
	}

	if n.Data == "span" && attrVal(n, "data-task-checkbox") != "" {
		return true
	}

	return n.Data == "input" && strings.EqualFold(attrVal(n, "type"), "checkbox")
}

func taskCheckboxChecked(n *html.Node) bool {
	if n.Data == "span" {
		return attrVal(n, "data-task-checkbox") == "true"
	}

	return hasAttr(n, "checked")
}

func hasAttr(n *html.Node, key string) bool {
	for _, a := range n.Attr {
		if a.Key == key {
			return true
		}
	}

	return false
}

func htmlText(n *html.Node) string {
	var buf strings.Builder

	for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
		if ch.Type == html.TextNode {
			buf.WriteString(ch.Data)
			continue
		}

		buf.WriteString(htmlText(ch))
	}

	return buf.String()
}

func parseOlStart(n *html.Node) int {
	s := attrVal(n, "start")
	if s == "" {
		return 1
	}

	v, err := strconv.Atoi(strings.TrimSpace(s))
	// HTML numbers a list from zero when start="0". A missing, non-numeric,
	// or negative value keeps the default of 1.
	if err != nil || v < 0 {
		return 1
	}

	return v
}

func (c *htmlBlockConverter) convertListItem(li *html.Node) ([]prosemirror.Node, error) {
	var nodes []prosemirror.Node

	if hasBlockElementChild(li) {
		children, err := c.convertBlockChildren(li)
		if err != nil {
			return nil, fmt.Errorf("cannot convert list item blocks: %w", err)
		}

		nodes = children
	} else {
		inlines, err := c.convertInlineFragments(li)
		if err != nil {
			return nil, fmt.Errorf("cannot convert list item text: %w", err)
		}

		if len(inlines) == 0 {
			return nil, nil
		}

		nodes = blocksFromInlines(prosemirror.NodeParagraph, nil, inlines)
	}

	if len(nodes) == 0 {
		return nil, nil
	}

	return ensureLeadingParagraph(nodes), nil
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
		"h1", "h2", "h3", "h4", "h5", "h6", "hr",
		"section", "article", "aside", "main", "header", "footer", "nav",
		"figure", "center",
		"html", "body", "head":
		return true
	default:
		return false
	}
}

// collectTableRows returns <tr> nodes that belong to this table only: direct
// children of thead/tbody/tfoot, or direct <tr> children of the table element
// (implicit tbody). It does not descend into cells or nested tables, so inner
// tables cannot contribute rows to the outer table.
func collectTableRows(table *html.Node) []*html.Node {
	var rows []*html.Node

	for ch := table.FirstChild; ch != nil; ch = ch.NextSibling {
		if ch.Type != html.ElementNode {
			continue
		}

		switch ch.Data {
		case "thead", "tbody", "tfoot":
			for tr := ch.FirstChild; tr != nil; tr = tr.NextSibling {
				if tr.Type == html.ElementNode && tr.Data == "tr" {
					rows = append(rows, tr)
				}
			}
		case "tr":
			rows = append(rows, ch)
		case "caption", "colgroup", "col":
			// Table metadata / columns; not row containers.
		default:
			// Ignore other direct children (e.g. invalid markup).
		}
	}

	return rows
}

func (c *htmlBlockConverter) convertTable(n *html.Node) ([]prosemirror.Node, error) {
	rows := collectTableRows(n)

	var rowNodes []prosemirror.Node

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

	return []prosemirror.Node{{Type: prosemirror.NodeTable, Content: rowNodes}}, nil
}

func (c *htmlBlockConverter) convertTableRow(tr *html.Node) (*prosemirror.Node, error) {
	var cells []prosemirror.Node

	for ch := tr.FirstChild; ch != nil; ch = ch.NextSibling {
		if ch.Type != html.ElementNode {
			continue
		}

		var (
			cell *prosemirror.Node
			err  error
		)

		switch ch.Data {
		case "th":
			cell, err = c.convertTableCell(ch, prosemirror.NodeTableHeader)
		case "td":
			cell, err = c.convertTableCell(ch, prosemirror.NodeTableCell)
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

	return &prosemirror.Node{Type: prosemirror.NodeTableRow, Content: cells}, nil
}

func (c *htmlBlockConverter) convertTableCell(n *html.Node, typ prosemirror.NodeType) (*prosemirror.Node, error) {
	cellAttrs := prosemirror.TableCellAttrs{
		Colspan: tableSpanFromHTML(n, "colspan"),
		Rowspan: tableSpanFromHTML(n, "rowspan"),
	}

	attrs, err := json.Marshal(cellAttrs)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal table cell attrs: %w", err)
	}

	inlines, err := c.convertInlineFragments(n)
	if err != nil {
		return nil, err
	}

	content := blocksFromInlines(prosemirror.NodeParagraph, nil, inlines)

	return &prosemirror.Node{Type: typ, Attrs: attrs, Content: content}, nil
}

func (c *htmlBlockConverter) convertInlineFragments(parent *html.Node) ([]prosemirror.Node, error) {
	var out []prosemirror.Node

	for ch := parent.FirstChild; ch != nil; ch = ch.NextSibling {
		nodes, err := c.convertInlineNode(ch)
		if err != nil {
			return nil, err
		}

		out = append(out, nodes...)
	}

	return out, nil
}

func (c *htmlBlockConverter) convertInlineNode(n *html.Node) ([]prosemirror.Node, error) {
	switch n.Type {
	case html.TextNode:
		if n.Data == "" {
			return nil, nil
		}

		t := n.Data

		return []prosemirror.Node{{
			Type:  prosemirror.NodeText,
			Text:  t,
			Marks: copyMarks(c.marks),
		}}, nil
	case html.ElementNode:
		return c.convertInlineElement(n)
	default:
		return nil, nil
	}
}

func (c *htmlBlockConverter) convertInlineElement(n *html.Node) ([]prosemirror.Node, error) {
	switch n.Data {
	case "p", "div", "section", "article", "aside", "main", "header", "footer", "nav",
		"center", "figure", "h1", "h2", "h3", "h4", "h5", "h6":
		// Block-ish tags inside inline HTML: unwrap to children only.
		return c.convertInlineFragments(n)
	case "br":
		return []prosemirror.Node{{Type: prosemirror.NodeHardBreak}}, nil
	case "strong", "b":
		return c.withMark(prosemirror.Mark{Type: prosemirror.MarkStrong}, n)
	case "em", "i":
		return c.withMark(prosemirror.Mark{Type: prosemirror.MarkEm}, n)
	case "s", "strike", "del":
		return c.withMark(prosemirror.Mark{Type: prosemirror.MarkStrike}, n)
	case "u":
		return c.withMark(prosemirror.Mark{Type: prosemirror.MarkUnderline}, n)
	case "code":
		return c.withMark(prosemirror.Mark{Type: prosemirror.MarkCode}, n)
	case "a":
		return c.convertAnchor(n)
	case "img":
		img, err := c.convertImageElement(n)
		if err != nil || img == nil {
			return nil, err
		}

		return []prosemirror.Node{*img}, nil
	case "span":
		return c.convertInlineFragments(n)
	default:
		return c.convertInlineFragments(n)
	}
}

func (c *htmlBlockConverter) withMark(m prosemirror.Mark, n *html.Node) ([]prosemirror.Node, error) {
	c.marks = append(c.marks, m)
	nodes, err := c.convertInlineFragments(n)
	c.marks = c.marks[:len(c.marks)-1]

	if err != nil {
		return nil, err
	}

	return nodes, nil
}

func (c *htmlBlockConverter) convertAnchor(n *html.Node) ([]prosemirror.Node, error) {
	href := attrVal(n, "href")
	if href == "" {
		return c.convertInlineFragments(n)
	}

	var title *string
	if t := attrVal(n, "title"); t != "" {
		title = &t
	}

	attrs, err := json.Marshal(prosemirror.LinkAttrs{Href: urlsafe.LinkHref(href), Title: title})
	if err != nil {
		return nil, fmt.Errorf("cannot marshal link attrs: %w", err)
	}

	m := prosemirror.Mark{Type: prosemirror.MarkLink, Attrs: attrs}
	c.marks = append(c.marks, m)
	nodes, err := c.convertInlineFragments(n)
	c.marks = c.marks[:len(c.marks)-1]

	if err != nil {
		return nil, err
	}

	return nodes, nil
}

func (c *htmlBlockConverter) convertImageElement(n *html.Node) (*prosemirror.Node, error) {
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

	imgAttrs := prosemirror.ImageAttrs{Src: src, Alt: alt, Title: title, Width: imageWidthPercent(attrVal(n, "width"))}

	attrs, err := json.Marshal(imgAttrs)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal image attrs: %w", err)
	}

	return &prosemirror.Node{Type: prosemirror.NodeImage, Attrs: attrs}, nil
}

// imageWidthPercent reads a column percentage from 1 through 100.
func imageWidthPercent(raw string) *int {
	raw = strings.TrimSpace(raw)
	if !strings.HasSuffix(raw, "%") {
		return nil
	}

	width, err := strconv.Atoi(strings.TrimSpace(strings.TrimSuffix(raw, "%")))
	if err != nil || width < 1 || width > 100 {
		return nil
	}

	return &width
}

func attrVal(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}

	return ""
}

// tableSpanFromHTML reads colspan or rowspan from a table cell element.
// Missing, invalid, or non-positive values yield 1, matching HTML defaults
// and ProseMirror/Tiptap cell attrs.
func tableSpanFromHTML(n *html.Node, key string) int {
	s := strings.TrimSpace(attrVal(n, key))
	if s == "" {
		return 1
	}

	v, err := strconv.Atoi(s)
	if err != nil || v < 1 {
		return 1
	}

	return v
}
