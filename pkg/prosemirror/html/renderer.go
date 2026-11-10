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
	"bytes"
	"fmt"

	"go.probo.inc/probo/pkg/prosemirror"
)

type (
	// NodeRenderer renders one node type to HTML.
	NodeRenderer interface {
		Node() prosemirror.NodeType
		Render(w *Writer, n prosemirror.Node) error
	}

	// MarkRenderer opens and closes one mark type in HTML output.
	MarkRenderer interface {
		Mark() prosemirror.MarkType
		Open(w *Writer, m prosemirror.Mark) error
		Close(w *Writer, m prosemirror.Mark)
	}

	// Renderer renders a node tree for a compiled schema.
	Renderer struct {
		schema  *prosemirror.Schema
		nodes   []NodeRenderer
		marks   map[prosemirror.MarkType]MarkRenderer
		byIndex []NodeRenderer
	}
)

// NewRenderer compiles node and mark renderers for s. It panics when a schema
// node or mark has no renderer, or when the registration lists contain
// duplicates.
func NewRenderer(
	s *prosemirror.Schema,
	nodes []NodeRenderer,
	marks []MarkRenderer,
) *Renderer {
	byType := make(map[prosemirror.NodeType]NodeRenderer, len(nodes))
	for _, r := range nodes {
		t := r.Node()
		if _, dup := byType[t]; dup {
			panic(fmt.Sprintf("html: duplicate node renderer %q", t))
		}

		byType[t] = r
	}

	markByType := make(map[prosemirror.MarkType]MarkRenderer, len(marks))
	for _, r := range marks {
		t := r.Mark()
		if _, dup := markByType[t]; dup {
			panic(fmt.Sprintf("html: duplicate mark renderer %q", t))
		}

		markByType[t] = r
	}

	maxIndex := -1

	for _, t := range s.NodeTypes() {
		idx, ok := s.NodeIndex(t)
		if !ok {
			continue
		}

		if idx > maxIndex {
			maxIndex = idx
		}

		if _, ok := byType[t]; !ok {
			panic(fmt.Sprintf("html: schema node %q has no renderer", t))
		}
	}

	byIndex := make([]NodeRenderer, maxIndex+1)

	for _, t := range s.NodeTypes() {
		idx, ok := s.NodeIndex(t)
		if !ok {
			continue
		}

		byIndex[idx] = byType[t]
	}

	for _, t := range s.MarkTypes() {
		if _, ok := markByType[t]; !ok {
			panic(fmt.Sprintf("html: schema mark %q has no renderer", t))
		}
	}

	return &Renderer{
		schema:  s,
		nodes:   nodes,
		marks:   markByType,
		byIndex: byIndex,
	}
}

// Render renders n to an HTML string.
func (r *Renderer) Render(n prosemirror.Node) (string, error) {
	var buf bytes.Buffer

	w := &Writer{buf: &buf, renderer: r}
	if err := r.render(w, n); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (r *Renderer) render(w *Writer, n prosemirror.Node) error {
	idx, ok := r.schema.NodeIndex(n.Type)
	if !ok {
		return fmt.Errorf("cannot render node: unknown type %q", n.Type)
	}

	if idx < 0 || idx >= len(r.byIndex) || r.byIndex[idx] == nil {
		return fmt.Errorf("cannot render node: unknown type %q", n.Type)
	}

	return renderNode(w, r.byIndex[idx], n)
}

func (r *Renderer) openMark(w *Writer, m prosemirror.Mark) error {
	renderer, ok := r.marks[m.Type]
	if !ok {
		return fmt.Errorf("cannot render mark: unknown type %q", m.Type)
	}

	return openMark(w, renderer, m)
}

func (r *Renderer) closeMark(w *Writer, m prosemirror.Mark) {
	renderer, ok := r.marks[m.Type]
	if !ok {
		return
	}

	closeMark(w, renderer, m)
}
