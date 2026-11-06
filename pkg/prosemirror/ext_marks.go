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

import (
	"encoding/json"
	"fmt"
	"strings"
)

type boldMark struct{}

// Bold renders strong emphasis.
func Bold() MarkSpec {
	return boldMark{}
}

func (boldMark) Name() MarkType {
	return MarkStrong
}

func (boldMark) Check(_ *Mark, _ CheckMode) error {
	return nil
}

func (boldMark) OpenHTML(w *HTMLWriter, _ Mark) error {
	w.WriteString("<strong>")

	return nil
}

func (boldMark) CloseHTML(w *HTMLWriter, _ Mark) {
	w.WriteString("</strong>")
}

func (boldMark) OpenMarkdown(w *MarkdownWriter, _ Mark) error {
	w.buf.WriteString("**")

	return nil
}

func (boldMark) CloseMarkdown(w *MarkdownWriter, _ Mark) error {
	w.buf.WriteString("**")

	return nil
}

type italicMark struct{}

// Italic renders emphasis.
func Italic() MarkSpec {
	return italicMark{}
}

func (italicMark) Name() MarkType {
	return MarkEm
}

func (italicMark) Check(_ *Mark, _ CheckMode) error {
	return nil
}

func (italicMark) OpenHTML(w *HTMLWriter, _ Mark) error {
	w.WriteString("<em>")

	return nil
}

func (italicMark) CloseHTML(w *HTMLWriter, _ Mark) {
	w.WriteString("</em>")
}

func (italicMark) OpenMarkdown(w *MarkdownWriter, _ Mark) error {
	w.buf.WriteByte('*')

	return nil
}

func (italicMark) CloseMarkdown(w *MarkdownWriter, _ Mark) error {
	w.buf.WriteByte('*')

	return nil
}

type underlineMark struct{}

// Underline renders underlined text.
func Underline() MarkSpec {
	return underlineMark{}
}

func (underlineMark) Name() MarkType {
	return MarkUnderline
}

func (underlineMark) Check(_ *Mark, _ CheckMode) error {
	return nil
}

func (underlineMark) OpenHTML(w *HTMLWriter, _ Mark) error {
	w.WriteString("<u>")

	return nil
}

func (underlineMark) CloseHTML(w *HTMLWriter, _ Mark) {
	w.WriteString("</u>")
}

func (underlineMark) OpenMarkdown(w *MarkdownWriter, _ Mark) error {
	w.buf.WriteString("<u>")

	return nil
}

func (underlineMark) CloseMarkdown(w *MarkdownWriter, _ Mark) error {
	w.buf.WriteString("</u>")

	return nil
}

type strikeMark struct{}

// Strike renders struck-through text.
func Strike() MarkSpec {
	return strikeMark{}
}

func (strikeMark) Name() MarkType {
	return MarkStrike
}

func (strikeMark) Check(_ *Mark, _ CheckMode) error {
	return nil
}

func (strikeMark) OpenHTML(w *HTMLWriter, _ Mark) error {
	w.WriteString("<s>")

	return nil
}

func (strikeMark) CloseHTML(w *HTMLWriter, _ Mark) {
	w.WriteString("</s>")
}

func (strikeMark) OpenMarkdown(w *MarkdownWriter, _ Mark) error {
	w.buf.WriteString("~~")

	return nil
}

func (strikeMark) CloseMarkdown(w *MarkdownWriter, _ Mark) error {
	w.buf.WriteString("~~")

	return nil
}

type codeMark struct{}

// Code renders inline monospaced text.
func Code() MarkSpec {
	return codeMark{}
}

func (codeMark) Name() MarkType {
	return MarkCode
}

func (codeMark) Check(_ *Mark, _ CheckMode) error {
	return nil
}

func (codeMark) OpenHTML(w *HTMLWriter, _ Mark) error {
	w.WriteString("<code>")

	return nil
}

func (codeMark) CloseHTML(w *HTMLWriter, _ Mark) {
	w.WriteString("</code>")
}

// Markdown renders inline code through the writer's code-text path, so these
// hooks are never invoked.
func (codeMark) OpenMarkdown(_ *MarkdownWriter, _ Mark) error {
	return nil
}

func (codeMark) CloseMarkdown(_ *MarkdownWriter, _ Mark) error {
	return nil
}

type linkMark struct{}

// Link renders a hyperlink.
func Link() MarkSpec {
	return linkMark{}
}

func (linkMark) Name() MarkType {
	return MarkLink
}

func (linkMark) Check(m *Mark, mode CheckMode) error {
	if mode != CheckSanitize {
		return nil
	}

	attrs, err := m.LinkAttrs()
	if err != nil {
		m.Attrs = []byte(`{"href":"#"}`)
		return nil
	}

	attrs.Href = safeLinkHref(attrs.Href)

	raw, err := json.Marshal(attrs)
	if err != nil {
		m.Attrs = []byte(`{"href":"#"}`)
		return nil
	}

	m.Attrs = raw

	return nil
}

func (linkMark) OpenHTML(w *HTMLWriter, m Mark) error {
	attrs, err := m.LinkAttrs()
	if err != nil {
		return fmt.Errorf("cannot render link mark: %w", err)
	}

	w.WriteString("<a")
	w.WriteAttr("href", safeLinkHref(attrs.Href))

	if attrs.Target != nil {
		w.WriteAttr("target", *attrs.Target)
	}

	rel := linkRelToEmit(attrs)
	if rel != "" {
		w.WriteAttr("rel", rel)
	}

	if attrs.Class != nil {
		w.WriteAttr("class", *attrs.Class)
	}

	if attrs.Title != nil {
		w.WriteAttr("title", *attrs.Title)
	}

	w.WriteByte('>')

	return nil
}

func (linkMark) CloseHTML(w *HTMLWriter, _ Mark) {
	w.WriteString("</a>")
}

func (linkMark) OpenMarkdown(w *MarkdownWriter, _ Mark) error {
	w.buf.WriteByte('[')

	return nil
}

func (linkMark) CloseMarkdown(w *MarkdownWriter, m Mark) error {
	attrs, err := m.LinkAttrs()
	if err != nil {
		return fmt.Errorf("cannot render link mark: %w", err)
	}

	w.buf.WriteString("](")
	w.buf.WriteString(safeLinkHref(attrs.Href))

	if attrs.Title != nil {
		w.buf.WriteString(` "`)
		w.buf.WriteString(strings.ReplaceAll(*attrs.Title, `"`, `\"`))
		w.buf.WriteByte('"')
	}

	w.buf.WriteByte(')')

	return nil
}
