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

package asana

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"go.probo.inc/probo/pkg/prosemirror"
)

// Asana rejects unsupported tags, including p, and rejects attributes on
// every tag except a. Paragraphs are separate lines. Headings below h2
// become h2, which is the lowest heading a task accepts.
func renderAsanaBlocks(
	nodes []prosemirror.Node,
	gids map[string]string,
	origin string,
) (string, error) {
	parts := make([]string, 0, len(nodes))

	for _, node := range nodes {
		if node.Type == prosemirror.NodeImage || node.Type == prosemirror.NodeFile {
			continue
		}

		part, err := renderAsanaBlock(node, gids, origin)
		if err != nil {
			return "", err
		}

		parts = append(parts, part)
	}

	return strings.Join(parts, "\n"), nil
}

func renderAsanaBlock(
	node prosemirror.Node,
	gids map[string]string,
	origin string,
) (string, error) {
	switch node.Type {
	case prosemirror.NodeParagraph:
		return renderAsanaInlines(node.Content, gids, origin)
	case prosemirror.NodeHeading:
		attrs, err := node.HeadingAttrs()
		if err != nil {
			return "", fmt.Errorf("cannot render heading: %w", err)
		}

		level := 2
		if attrs.Level == 1 {
			level = 1
		}

		inner, err := renderAsanaInlines(node.Content, gids, origin)
		if err != nil {
			return "", fmt.Errorf("cannot render heading: %w", err)
		}

		tag := "h" + strconv.Itoa(level)

		return "<" + tag + ">" + inner + "</" + tag + ">", nil
	case prosemirror.NodeBlockquote:
		inner, err := renderAsanaBlocks(node.Content, gids, origin)
		if err != nil {
			return "", fmt.Errorf("cannot render blockquote: %w", err)
		}

		return "<blockquote>" + inner + "</blockquote>", nil
	case prosemirror.NodeCodeBlock:
		inner, err := renderAsanaInlines(node.Content, gids, origin)
		if err != nil {
			return "", fmt.Errorf("cannot render code block: %w", err)
		}

		return "<pre>" + inner + "</pre>", nil
	case prosemirror.NodeHorizontalRule:
		return "<hr/>", nil
	case prosemirror.NodeBulletList:
		return renderAsanaList("ul", node.Content, gids, origin, false)
	case prosemirror.NodeOrderedList:
		return renderAsanaOrderedList(node, gids, origin)
	case prosemirror.NodeTaskList:
		return renderAsanaList("ul", node.Content, gids, origin, true)
	case prosemirror.NodeListItem, prosemirror.NodeTaskItem:
		return renderAsanaListItem(node, gids, origin)
	case prosemirror.NodeTable:
		return renderAsanaTable(node, gids, origin)
	case prosemirror.NodeText, prosemirror.NodeHardBreak, prosemirror.NodeMention:
		return renderAsanaInline(node, gids, origin)
	default:
		if len(node.Content) == 0 {
			return "", nil
		}

		return renderAsanaBlocks(node.Content, gids, origin)
	}
}

func renderAsanaList(
	tag string,
	items []prosemirror.Node,
	gids map[string]string,
	origin string,
	tasks bool,
) (string, error) {
	var buf strings.Builder

	buf.WriteByte('<')
	buf.WriteString(tag)
	buf.WriteByte('>')

	for _, item := range items {
		if tasks && item.Type == prosemirror.NodeTaskItem {
			rendered, err := renderAsanaTaskItem(item, gids, origin)
			if err != nil {
				return "", fmt.Errorf("cannot render task item: %w", err)
			}

			buf.WriteString(rendered)

			continue
		}

		rendered, err := renderAsanaListItem(item, gids, origin)
		if err != nil {
			return "", fmt.Errorf("cannot render list item: %w", err)
		}

		buf.WriteString(rendered)
	}

	buf.WriteString("</")
	buf.WriteString(tag)
	buf.WriteByte('>')

	return buf.String(), nil
}

// Asana numbers every plain ol from 1 and rejects a start attribute.
// A non-default start is a ul whose items carry that numbering.
func renderAsanaOrderedList(
	node prosemirror.Node,
	gids map[string]string,
	origin string,
) (string, error) {
	attrs, err := node.OrderedListAttrs()
	if err != nil {
		return "", fmt.Errorf("cannot read ordered list: %w", err)
	}

	if attrs.Start == 1 {
		rendered, err := renderAsanaList("ol", node.Content, gids, origin, false)
		if err != nil {
			return "", fmt.Errorf("cannot render ordered list: %w", err)
		}

		return rendered, nil
	}

	rendered, err := renderAsanaNumberedList(attrs.Start, node.Content, gids, origin)
	if err != nil {
		return "", fmt.Errorf("cannot render numbered list: %w", err)
	}

	return rendered, nil
}

func renderAsanaNumberedList(
	start int,
	items []prosemirror.Node,
	gids map[string]string,
	origin string,
) (string, error) {
	var buf strings.Builder

	buf.WriteString("<ul>")

	for i, item := range items {
		rendered, err := renderAsanaNumberedItem(start+i, item, gids, origin)
		if err != nil {
			return "", err
		}

		buf.WriteString(rendered)
	}

	buf.WriteString("</ul>")

	return buf.String(), nil
}

func renderAsanaNumberedItem(
	number int,
	node prosemirror.Node,
	gids map[string]string,
	origin string,
) (string, error) {
	inner, err := renderAsanaBlocks(node.Content, gids, origin)
	if err != nil {
		return "", err
	}

	return "<li>" + strconv.Itoa(number) + ". " + inner + "</li>", nil
}

func renderAsanaListItem(
	node prosemirror.Node,
	gids map[string]string,
	origin string,
) (string, error) {
	inner, err := renderAsanaBlocks(node.Content, gids, origin)
	if err != nil {
		return "", err
	}

	return "<li>" + inner + "</li>", nil
}

func renderAsanaTaskItem(
	node prosemirror.Node,
	gids map[string]string,
	origin string,
) (string, error) {
	attrs, err := node.TaskItemAttrs()
	if err != nil {
		return "", fmt.Errorf("cannot render task item: %w", err)
	}

	inner, err := renderAsanaBlocks(node.Content, gids, origin)
	if err != nil {
		return "", fmt.Errorf("cannot render task item body: %w", err)
	}

	if attrs.Checked {
		inner = "[x] " + inner
	} else {
		inner = "[ ] " + inner
	}

	return "<li>" + inner + "</li>", nil
}

func renderAsanaTable(
	node prosemirror.Node,
	gids map[string]string,
	origin string,
) (string, error) {
	rows := make([]string, 0, len(node.Content))

	for _, row := range node.Content {
		cells := make([]string, 0, len(row.Content))

		for _, cell := range row.Content {
			text, err := renderAsanaBlocks(cell.Content, gids, origin)
			if err != nil {
				return "", err
			}

			cells = append(cells, text)
		}

		rows = append(rows, strings.Join(cells, " | "))
	}

	return strings.Join(rows, "\n"), nil
}

func renderAsanaInlines(
	nodes []prosemirror.Node,
	gids map[string]string,
	origin string,
) (string, error) {
	var buf strings.Builder

	for _, node := range nodes {
		rendered, err := renderAsanaInline(node, gids, origin)
		if err != nil {
			return "", err
		}

		buf.WriteString(rendered)
	}

	return buf.String(), nil
}

func renderAsanaInline(
	node prosemirror.Node,
	gids map[string]string,
	origin string,
) (string, error) {
	switch node.Type {
	case prosemirror.NodeHardBreak:
		return "\n", nil
	case prosemirror.NodeMention:
		return renderAsanaMention(node, gids, origin)
	case prosemirror.NodeText:
		return renderAsanaText(node)
	default:
		if len(node.Content) == 0 {
			return "", nil
		}

		return renderAsanaInlines(node.Content, gids, origin)
	}
}

func renderAsanaMention(
	node prosemirror.Node,
	gids map[string]string,
	origin string,
) (string, error) {
	attrs, err := node.MentionAttrs()
	if err != nil {
		return "", fmt.Errorf("cannot render mention: %w", err)
	}

	label := htmlEscape("@" + attrs.Label)

	userGID := ""
	if gids != nil {
		userGID = gids[attrs.ID]
	}

	if !asanaNumericGID.MatchString(userGID) {
		return label, nil
	}

	profileURL, err := asanaProfileURL(origin, userGID)
	if err != nil {
		return label, nil
	}

	return `<a data-asana-gid="` + userGID + `" href="` + htmlEscape(profileURL) + `">` + label + `</a>`, nil
}

func renderAsanaText(node prosemirror.Node) (string, error) {
	if node.Text == nil {
		return "", fmt.Errorf("cannot render text node: text is nil")
	}

	opened := make([]string, 0, len(node.Marks))

	var buf strings.Builder

	for _, mark := range node.Marks {
		switch mark.Type {
		case prosemirror.MarkStrong:
			buf.WriteString("<strong>")

			opened = append(opened, "</strong>")
		case prosemirror.MarkEm:
			buf.WriteString("<em>")

			opened = append(opened, "</em>")
		case prosemirror.MarkUnderline:
			buf.WriteString("<u>")

			opened = append(opened, "</u>")
		case prosemirror.MarkStrike:
			buf.WriteString("<s>")

			opened = append(opened, "</s>")
		case prosemirror.MarkCode:
			buf.WriteString("<code>")

			opened = append(opened, "</code>")
		case prosemirror.MarkLink:
			attrs, err := mark.LinkAttrs()
			if err != nil {
				return "", fmt.Errorf("cannot render link: %w", err)
			}

			buf.WriteString(`<a href="`)
			buf.WriteString(htmlEscape(prosemirror.SafeLinkHref(attrs.Href)))
			buf.WriteString(`">`)

			opened = append(opened, "</a>")
		}
	}

	buf.WriteString(htmlEscape(*node.Text))

	for _, o := range slices.Backward(opened) {
		buf.WriteString(o)
	}

	return buf.String(), nil
}
