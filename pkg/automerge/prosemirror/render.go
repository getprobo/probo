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
	"errors"
	"fmt"
	"slices"
	"sort"

	"go.probo.inc/probo/pkg/automerge"
)

// errUnknownBlock signals that a block span carries a type the renderer does not
// recognize. It is handled gracefully rather than aborting the whole render so a
// single unfamiliar block (for example from a newer client) can never block
// persistence of an entire document.
var errUnknownBlock = errors.New("unknown Automerge block type")

type schemaMapping struct {
	Automerge   string
	ProseMirror string
}

// blockMappings and markMappings are the Go half of the shared ProseMirror <->
// Automerge schema ledger (testdata/schema-mapping.json). Keeping them as data
// lets a drift test assert that this renderer, the frontend adapter, and the
// ledger all agree on block-type strings, mark names, and mark order. markMappings
// is ordered by ProseMirror schema rank because ProseMirror stores marks in that
// order, so the renderer must emit them the same way for byte-identical documents.
var (
	blockMappings = []schemaMapping{
		{Automerge: blockTypeParagraph, ProseMirror: "paragraph"},
		{Automerge: blockTypeHeading, ProseMirror: "heading"},
		{Automerge: blockTypeCode, ProseMirror: "codeBlock"},
		{Automerge: blockTypeBlockquote, ProseMirror: "blockquote"},
		{Automerge: blockTypeOrderedListItem, ProseMirror: "listItem"},
		{Automerge: blockTypeUnorderedListItem, ProseMirror: "listItem"},
		{Automerge: blockTypeHorizontalRule, ProseMirror: "horizontalRule"},
		{Automerge: blockTypeHardBreak, ProseMirror: "hardBreak"},
		{Automerge: blockTypeTable, ProseMirror: "table"},
		{Automerge: blockTypeTableCell, ProseMirror: "tableCell"},
		{Automerge: blockTypeTableHeader, ProseMirror: "tableHeader"},
		{Automerge: blockTypeTableRow, ProseMirror: "tableRow"},
	}

	markMappings = []schemaMapping{
		{Automerge: "link", ProseMirror: "link"},
		{Automerge: "strong", ProseMirror: "bold"},
		{Automerge: "em", ProseMirror: "italic"},
		{Automerge: "strike", ProseMirror: "strike"},
		{Automerge: "underline", ProseMirror: "underline"},
		{Automerge: "code", ProseMirror: "code"},
	}

	blockNodeNames  = map[string]string{}
	markNodeNames   = map[string]string{}
	markRenderOrder = map[string]int{}
)

func init() {
	for _, mapping := range blockMappings {
		blockNodeNames[mapping.Automerge] = mapping.ProseMirror
	}

	for index, mapping := range markMappings {
		markNodeNames[mapping.Automerge] = mapping.ProseMirror
		markRenderOrder[mapping.Automerge] = index
	}
}

type (
	node struct {
		Type    string         `json:"type"`
		Attrs   map[string]any `json:"attrs,omitempty"`
		Content []node         `json:"content,omitempty"`
		Text    string         `json:"text,omitempty"`
		Marks   []mark         `json:"marks,omitempty"`
	}

	mark struct {
		Type  string         `json:"type"`
		Attrs map[string]any `json:"attrs,omitempty"`
	}

	block struct {
		Type    string
		Parents []string
		Attrs   map[string]any
		Content []node
	}
)

const (
	blockTypeBlockquote        = "blockquote"
	blockTypeCode              = "code-block"
	blockTypeHardBreak         = "hard-break"
	blockTypeHeading           = "heading"
	blockTypeHorizontalRule    = "horizontal-rule"
	blockTypeOrderedListItem   = "ordered-list-item"
	blockTypeParagraph         = "paragraph"
	blockTypeTable             = "table"
	blockTypeTableCell         = "table-cell"
	blockTypeTableHeader       = "table-header"
	blockTypeTableRow          = "table-row"
	blockTypeUnorderedListItem = "unordered-list-item"
)

func Render(spans []automerge.Span) (string, error) {
	blocks, err := collectBlocks(spans)
	if err != nil {
		return "", err
	}

	content, consumed, err := renderBlocks(blocks, nil)
	if err != nil {
		return "", err
	}

	if consumed != len(blocks) {
		return "", fmt.Errorf("cannot render Automerge blocks: consumed %d of %d", consumed, len(blocks))
	}

	if len(content) == 0 {
		content = []node{{Type: "paragraph"}}
	}

	data, err := json.Marshal(node{Type: "doc", Content: content})
	if err != nil {
		return "", fmt.Errorf("cannot marshal ProseMirror document: %w", err)
	}

	return string(data), nil
}

func collectBlocks(spans []automerge.Span) ([]block, error) {
	blocks := make([]block, 0)

	for i, span := range spans {
		switch span.Type {
		case automerge.SpanTypeBlock:
			blockType, ok := span.Block["type"].(string)
			if !ok || blockType == "" {
				return nil, fmt.Errorf("cannot collect Automerge block span %d: missing type", i)
			}

			if blockType == blockTypeHardBreak {
				if len(blocks) == 0 {
					return nil, fmt.Errorf("cannot collect Automerge hard break span %d without a parent block", i)
				}

				blocks[len(blocks)-1].Content = append(
					blocks[len(blocks)-1].Content,
					node{Type: blockNodeNames[blockTypeHardBreak]},
				)

				continue
			}

			parents, err := stringSlice(span.Block["parents"])
			if err != nil {
				return nil, fmt.Errorf("cannot collect Automerge block span %d parents: %w", i, err)
			}

			attrs, _ := span.Block["attrs"].(map[string]any)
			blocks = append(
				blocks,
				block{
					Type:    blockType,
					Parents: parents,
					Attrs:   attrs,
				},
			)
		case automerge.SpanTypeText:
			if len(blocks) == 0 || blocks[len(blocks)-1].Type == blockTypeHorizontalRule {
				var parents []string
				if len(blocks) > 0 {
					parents = slices.Clone(blocks[len(blocks)-1].Parents)
				}

				blocks = append(
					blocks,
					block{
						Type:    blockTypeParagraph,
						Parents: parents,
					},
				)
			}

			marks, err := renderMarks(span.Marks)
			if err != nil {
				return nil, fmt.Errorf("cannot collect Automerge text span %d marks: %w", i, err)
			}

			if span.Text != "" {
				blocks[len(blocks)-1].Content = append(
					blocks[len(blocks)-1].Content,
					node{
						Type:  "text",
						Text:  span.Text,
						Marks: marks,
					},
				)
			}
		default:
			return nil, fmt.Errorf("cannot collect Automerge span %d: unknown type %q", i, span.Type)
		}
	}

	return blocks, nil
}

func renderBlocks(blocks []block, parents []string) ([]node, int, error) {
	content := make([]node, 0)
	consumed := 0

	for consumed < len(blocks) {
		current := blocks[consumed]
		if !slices.Equal(current.Parents, parents) {
			break
		}

		childParents := append(append([]string(nil), parents...), current.Type)

		childEnd := consumed + 1
		for childEnd < len(blocks) && hasPrefix(blocks[childEnd].Parents, childParents) {
			childEnd++
		}

		children, childConsumed, err := renderBlocks(blocks[consumed+1:childEnd], childParents)
		if err != nil {
			return nil, 0, err
		}

		if childConsumed != childEnd-consumed-1 {
			return nil, 0, fmt.Errorf("cannot render children of Automerge block %q", current.Type)
		}

		rendered, listType, err := renderBlock(current, children)
		if err != nil {
			if errors.Is(err, errUnknownBlock) {
				if len(current.Content) > 0 {
					content = append(content, node{Type: "paragraph", Content: current.Content})
				}

				content = append(content, children...)
				consumed = childEnd

				continue
			}

			return nil, 0, err
		}

		if listType != "" {
			if len(content) > 0 && content[len(content)-1].Type == listType {
				content[len(content)-1].Content = append(content[len(content)-1].Content, rendered)
			} else {
				content = append(
					content,
					node{
						Type:    listType,
						Content: []node{rendered},
					},
				)
			}
		} else {
			content = append(content, rendered)
		}

		consumed = childEnd
	}

	return content, consumed, nil
}

func renderBlock(source block, children []node) (node, string, error) {
	switch source.Type {
	case blockTypeParagraph:
		if len(children) > 0 {
			return node{}, "", fmt.Errorf("paragraph Automerge block cannot contain child blocks")
		}

		return node{Type: blockNodeNames[blockTypeParagraph], Content: source.Content}, "", nil
	case blockTypeHeading:
		if len(children) > 0 {
			return node{}, "", fmt.Errorf("heading Automerge block cannot contain child blocks")
		}

		level := intAttribute(source.Attrs, "level", 1)
		if level < 1 || level > 6 {
			level = 1
		}

		return node{
			Type:    blockNodeNames[blockTypeHeading],
			Attrs:   map[string]any{"level": level},
			Content: source.Content,
		}, "", nil
	case blockTypeCode:
		if len(children) > 0 {
			return node{}, "", fmt.Errorf("code Automerge block cannot contain child blocks")
		}

		attrs := map[string]any{"language": nil}
		if language, ok := source.Attrs["language"].(string); ok {
			attrs["language"] = language
		}

		return node{Type: blockNodeNames[blockTypeCode], Attrs: attrs, Content: source.Content}, "", nil
	case blockTypeHorizontalRule:
		if len(source.Content) > 0 || len(children) > 0 {
			return node{}, "", fmt.Errorf("horizontal rule Automerge block cannot contain content")
		}

		return node{Type: blockNodeNames[blockTypeHorizontalRule]}, "", nil
	case blockTypeBlockquote:
		paragraph := node{Type: blockNodeNames[blockTypeParagraph], Content: source.Content}

		return node{
			Type:    blockNodeNames[blockTypeBlockquote],
			Content: append([]node{paragraph}, children...),
		}, "", nil
	case blockTypeOrderedListItem:
		paragraph := node{Type: blockNodeNames[blockTypeParagraph], Content: source.Content}

		return node{
			Type:    blockNodeNames[blockTypeOrderedListItem],
			Content: append([]node{paragraph}, children...),
		}, "orderedList", nil
	case blockTypeUnorderedListItem:
		paragraph := node{Type: blockNodeNames[blockTypeParagraph], Content: source.Content}

		return node{
			Type:    blockNodeNames[blockTypeUnorderedListItem],
			Content: append([]node{paragraph}, children...),
		}, "bulletList", nil
	case blockTypeTable:
		if len(source.Content) > 0 {
			return node{}, "", fmt.Errorf("table Automerge block cannot contain inline content")
		}

		return node{Type: blockNodeNames[blockTypeTable], Content: children}, "", nil
	case blockTypeTableRow:
		if len(source.Content) > 0 {
			return node{}, "", fmt.Errorf("table row Automerge block cannot contain inline content")
		}

		return node{Type: blockNodeNames[blockTypeTableRow], Content: children}, "", nil
	case blockTypeTableCell, blockTypeTableHeader:
		content := children
		if len(source.Content) > 0 || len(children) == 0 {
			paragraph := node{Type: blockNodeNames[blockTypeParagraph], Content: source.Content}
			content = append([]node{paragraph}, children...)
		}

		attrs := map[string]any{
			"colspan":  intAttribute(source.Attrs, "colspan", 1),
			"rowspan":  intAttribute(source.Attrs, "rowspan", 1),
			"colwidth": intSliceAttribute(source.Attrs, "colwidth"),
		}

		nodeType := blockNodeNames[blockTypeTableCell]
		if source.Type == blockTypeTableHeader {
			nodeType = blockNodeNames[blockTypeTableHeader]
		}

		return node{
			Type:    nodeType,
			Attrs:   attrs,
			Content: content,
		}, "", nil
	default:
		return node{}, "", errUnknownBlock
	}
}

func renderMarks(values map[string]any) ([]mark, error) {
	if len(values) == 0 {
		return nil, nil
	}

	names := make([]string, 0, len(values))

	for name := range values {
		if _, known := markRenderOrder[name]; !known {
			// Unknown marks are dropped rather than aborting the render, mirroring
			// the frontend's tolerance for schema drift: the formatting is lost but
			// the text survives and downstream renderers only see known marks.
			continue
		}

		names = append(names, name)
	}

	sort.SliceStable(names, func(i, j int) bool {
		return markRenderOrder[names[i]] < markRenderOrder[names[j]]
	})

	marks := make([]mark, 0, len(names))

	for _, name := range names {
		if name == "link" {
			raw, ok := values[name].(string)
			if !ok {
				continue
			}

			var attrs map[string]any
			if err := json.Unmarshal([]byte(raw), &attrs); err != nil {
				continue
			}

			marks = append(marks, mark{Type: markNodeNames[name], Attrs: attrs})

			continue
		}

		marks = append(marks, mark{Type: markNodeNames[name]})
	}

	if len(marks) == 0 {
		return nil, nil
	}

	return marks, nil
}

func stringSlice(value any) ([]string, error) {
	if value == nil {
		return nil, nil
	}

	values, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("expected an array")
	}

	result := make([]string, len(values))
	for i, value := range values {
		result[i], ok = value.(string)
		if !ok {
			return nil, fmt.Errorf("value %d is not a string", i)
		}
	}

	return result, nil
}

func intAttribute(attrs map[string]any, name string, fallback int) int {
	value, ok := attrs[name].(float64)
	if !ok {
		return fallback
	}

	return int(value)
}

func intSliceAttribute(attrs map[string]any, name string) []int {
	values, ok := attrs[name].([]any)
	if !ok {
		return nil
	}

	result := make([]int, 0, len(values))
	for _, value := range values {
		number, ok := value.(float64)
		if !ok {
			return nil
		}

		result = append(result, int(number))
	}

	return result
}

func hasPrefix(values, prefix []string) bool {
	return len(values) >= len(prefix) && slices.Equal(values[:len(prefix)], prefix)
}
