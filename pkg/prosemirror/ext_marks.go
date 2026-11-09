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

// Markdown renders inline code through the writer's code-text path, so these
// hooks are never invoked.

type linkMark struct{}

// LinkAttrs contains attributes for link marks.
type LinkAttrs struct {
	Href   string  `json:"href"`
	Target *string `json:"target,omitempty"`
	Rel    *string `json:"rel,omitempty"`
	Class  *string `json:"class,omitempty"`
	Title  *string `json:"title,omitempty"`
}

// LinkAttrs parses and returns the link attributes from a link mark.
func (m Mark) LinkAttrs() (LinkAttrs, error) {
	return decodeAttrs(m.Attrs, LinkAttrs{}, "link")
}

// Link renders a hyperlink.
func Link() MarkSpec {
	return linkMark{}
}

func (linkMark) Name() MarkType {
	return MarkLink
}

func (linkMark) Check(m *Mark, mode CheckMode) error {
	attrs, err := m.LinkAttrs()
	if err != nil {
		if mode == CheckSanitize {
			m.Attrs = []byte(`{"href":"#"}`)
			return nil
		}

		return fmt.Errorf("cannot validate link mark: %w", err)
	}

	if mode != CheckSanitize {
		return nil
	}

	attrs.Href = SafeLinkHref(attrs.Href)

	if err := writeAttrs(&m.Attrs, attrs); err != nil {
		return fmt.Errorf("cannot validate link mark: %w", err)
	}

	return nil
}
