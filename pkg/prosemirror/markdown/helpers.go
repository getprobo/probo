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
	"bytes"
	"fmt"
	"html"
	"strings"

	"go.probo.inc/probo/pkg/prosemirror"
)

func writeAttr(buf *bytes.Buffer, name, value string) {
	buf.WriteByte(' ')
	buf.WriteString(name)
	buf.WriteString(`="`)
	buf.WriteString(html.EscapeString(value))
	buf.WriteByte('"')
}

func maxConsecutiveBackticks(s string) int {
	max, cur := 0, 0

	for i := 0; i < len(s); i++ {
		if s[i] == '`' {
			cur++
			if cur > max {
				max = cur
			}
		} else {
			cur = 0
		}
	}

	return max
}

func listTightness(n prosemirror.Node, itemType prosemirror.NodeType) (tight bool, err error) {
	tight = true

	for _, item := range n.Content {
		if item.Type != itemType {
			return false, fmt.Errorf(
				"invalid child type %q (expected %q)",
				item.Type,
				itemType,
			)
		}

		if len(item.Content) != 1 {
			tight = false
		}
	}

	return tight, nil
}

func collectText(nodes []prosemirror.Node) string {
	var buf strings.Builder

	for _, n := range nodes {
		if n.Text != "" {
			buf.WriteString(n.Text)
		}
	}

	return buf.String()
}

func chooseFence(code string) string {
	fence := "```"
	for strings.Contains(code, fence) {
		fence += "`"
	}

	return fence
}

func writeMarkdownEscaped(buf *bytes.Buffer, s string) {
	for _, c := range s {
		switch c {
		case '\\', '*', '_', '`', '[', ']', '~', '|', '<':
			buf.WriteByte('\\')
		}

		buf.WriteRune(c)
	}
}

func writeEscapingPipes(buf *bytes.Buffer, raw []byte) {
	for _, b := range raw {
		if b == '|' {
			buf.WriteString(`\|`)
			continue
		}

		buf.WriteByte(b)
	}
}

func escapeMarkdown(s string) string {
	var buf bytes.Buffer
	buf.Grow(len(s))
	writeMarkdownEscaped(&buf, s)

	return buf.String()
}
