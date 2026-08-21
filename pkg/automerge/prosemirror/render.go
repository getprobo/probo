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
	"slices"
	"sort"

	"go.probo.inc/probo/pkg/automerge"
)

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
		content = append(content, flattenBlocks(blocks[consumed:])...)
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

	for _, span := range spans {
		switch span.Type {
		case automerge.SpanTypeBlock:
			blockType, ok := span.Block["type"].(string)
			if !ok || blockType == "" {
				blockType = "__unknown__"
			}

			if blockType == blockTypeHardBreak {
				if len(blocks) == 0 || !acceptsInlineContent(blocks[len(blocks)-1].Type) {
					parents := tolerantStringSlice(span.Block["parents"])
					blocks = append(
						blocks,
						block{
							Type:    blockTypeParagraph,
							Parents: inlineFallbackParents(parents),
						},
					)
				}

				blocks[len(blocks)-1].Content = append(
					blocks[len(blocks)-1].Content,
					node{Type: blockNodeNames[blockTypeHardBreak]},
				)

				continue
			}

			attrs, _ := span.Block["attrs"].(map[string]any)
			blocks = appendNormalizedBlock(
				blocks,
				block{
					Type:    blockType,
					Parents: tolerantStringSlice(span.Block["parents"]),
					Attrs:   attrs,
				},
			)
		case automerge.SpanTypeText:
			if len(blocks) == 0 || !acceptsInlineContent(blocks[len(blocks)-1].Type) {
				var parents []string
				if len(blocks) > 0 {
					parents = inlineFallbackParents(blocks[len(blocks)-1].Parents)
				}

				blocks = appendNormalizedBlock(
					blocks,
					block{
						Type:    blockTypeParagraph,
						Parents: parents,
					},
				)
			}

			marks, err := renderMarks(span.Marks)
			if err != nil {
				marks = nil
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
			continue
		}
	}

	return blocks, nil
}

func appendNormalizedBlock(blocks []block, next block) []block {
	if len(blocks) == 0 {
		next.Parents = nil

		return append(blocks, next)
	}

	previous := blocks[len(blocks)-1]
	previousPath := append(slices.Clone(previous.Parents), previous.Type)
	prefixLength := 0

	for prefixLength < len(next.Parents) &&
		prefixLength < len(previousPath) &&
		next.Parents[prefixLength] == previousPath[prefixLength] {
		prefixLength++
	}

	next.Parents = slices.Clone(next.Parents[:prefixLength])

	return append(blocks, next)
}

// hoistNodes prepares already-rendered children for a context that is not the one
// they were rendered under. Table parts only mean anything inside a table, so they
// are unwrapped into their own content rather than lifted verbatim.
func hoistNodes(values []node) []node {
	result := make([]node, 0, len(values))

	for _, value := range values {
		switch value.Type {
		case blockNodeNames[blockTypeTableRow],
			blockNodeNames[blockTypeTableCell],
			blockNodeNames[blockTypeTableHeader]:
			result = append(result, hoistNodes(value.Content)...)
		default:
			result = append(result, value)
		}
	}

	return result
}

// blockAllowedUnder reports whether a block type may appear directly beneath the
// given ancestor path. Table parts only mean anything inside their container, and
// the Markdown and HTML renderers reject them anywhere else, so a stray cell or row
// degrades to its inline content instead of poisoning the whole document.
func blockAllowedUnder(blockType string, parents []string) bool {
	var parent string
	if len(parents) > 0 {
		parent = parents[len(parents)-1]
	}

	switch blockType {
	case blockTypeTableCell, blockTypeTableHeader:
		return parent == blockTypeTableRow
	case blockTypeTableRow:
		return parent == blockTypeTable
	default:
		return true
	}
}

func acceptsInlineContent(blockType string) bool {
	switch blockType {
	case blockTypeHorizontalRule, blockTypeTable, blockTypeTableRow:
		return false
	default:
		return true
	}
}

func inlineFallbackParents(parents []string) []string {
	parents = slices.Clone(parents)

	for len(parents) > 0 {
		last := parents[len(parents)-1]
		if last != blockTypeTable && last != blockTypeTableRow {
			break
		}

		parents = parents[:len(parents)-1]
	}

	return parents
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
			children = append(children, flattenBlocks(blocks[consumed+1+childConsumed:childEnd])...)
		}

		if !blockAllowedUnder(current.Type, parents) {
			if len(current.Content) > 0 {
				content = append(
					content,
					node{Type: blockNodeNames[blockTypeParagraph], Content: current.Content},
				)
			}

			content = append(content, hoistNodes(children)...)
			consumed = childEnd

			continue
		}

		rendered, listType := renderBlock(current, children)

		if listType != "" {
			if len(content) > 0 && content[len(content)-1].Type == listType {
				content[len(content)-1].Content = append(content[len(content)-1].Content, rendered[0])
			} else {
				content = append(
					content,
					node{
						Type:    listType,
						Content: []node{rendered[0]},
					},
				)
			}

			content = append(content, rendered[1:]...)
		} else {
			content = append(content, rendered...)
		}

		consumed = childEnd
	}

	return content, consumed, nil
}

func flattenBlocks(blocks []block) []node {
	content := make([]node, 0, len(blocks))

	for _, source := range blocks {
		if len(source.Content) > 0 {
			content = append(
				content,
				node{Type: blockNodeNames[blockTypeParagraph], Content: source.Content},
			)
		}
	}

	return content
}

func renderBlock(source block, children []node) ([]node, string) {
	switch source.Type {
	case blockTypeParagraph:
		rendered := node{Type: blockNodeNames[blockTypeParagraph], Content: source.Content}

		return append([]node{rendered}, children...), ""
	case blockTypeHeading:
		level := intAttribute(source.Attrs, "level", 1)
		if level < 1 || level > 6 {
			level = 1
		}

		rendered := node{
			Type:    blockNodeNames[blockTypeHeading],
			Attrs:   map[string]any{"level": level},
			Content: source.Content,
		}

		return append([]node{rendered}, children...), ""
	case blockTypeCode:
		attrs := map[string]any{"language": nil}
		if language, ok := source.Attrs["language"].(string); ok {
			attrs["language"] = language
		}

		rendered := node{
			Type:    blockNodeNames[blockTypeCode],
			Attrs:   attrs,
			Content: source.Content,
		}

		return append([]node{rendered}, children...), ""
	case blockTypeHorizontalRule:
		rendered := []node{{Type: blockNodeNames[blockTypeHorizontalRule]}}
		if len(source.Content) > 0 {
			rendered = append(
				rendered,
				node{Type: blockNodeNames[blockTypeParagraph], Content: source.Content},
			)
		}

		return append(rendered, children...), ""
	case blockTypeBlockquote:
		content := children
		if len(source.Content) > 0 || len(children) == 0 {
			paragraph := node{Type: blockNodeNames[blockTypeParagraph], Content: source.Content}
			content = append([]node{paragraph}, children...)
		}

		return []node{{
			Type:    blockNodeNames[blockTypeBlockquote],
			Content: content,
		}}, ""
	case blockTypeOrderedListItem:
		paragraph := node{Type: blockNodeNames[blockTypeParagraph], Content: source.Content}

		return []node{{
			Type:    blockNodeNames[blockTypeOrderedListItem],
			Content: append([]node{paragraph}, children...),
		}}, "orderedList"
	case blockTypeUnorderedListItem:
		paragraph := node{Type: blockNodeNames[blockTypeParagraph], Content: source.Content}

		return []node{{
			Type:    blockNodeNames[blockTypeUnorderedListItem],
			Content: append([]node{paragraph}, children...),
		}}, "bulletList"
	case blockTypeTable:
		rows, spills := partitionNodes(children, blockNodeNames[blockTypeTableRow])
		rendered := make([]node, 0, 2+len(spills))

		if len(rows) > 0 {
			rendered = append(
				rendered,
				node{Type: blockNodeNames[blockTypeTable], Content: rows},
			)
		}

		if len(source.Content) > 0 {
			rendered = append(
				rendered,
				node{Type: blockNodeNames[blockTypeParagraph], Content: source.Content},
			)
		}

		return append(rendered, hoistNodes(spills)...), ""
	case blockTypeTableRow:
		cells, spills := partitionNodes(
			children,
			blockNodeNames[blockTypeTableCell],
			blockNodeNames[blockTypeTableHeader],
		)
		rendered := make([]node, 0, 2+len(spills))

		if len(cells) > 0 {
			rendered = append(
				rendered,
				node{Type: blockNodeNames[blockTypeTableRow], Content: cells},
			)
		}

		if len(source.Content) > 0 {
			rendered = append(
				rendered,
				node{Type: blockNodeNames[blockTypeParagraph], Content: source.Content},
			)
		}

		return append(rendered, hoistNodes(spills)...), ""
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

		return []node{{
			Type:    nodeType,
			Attrs:   attrs,
			Content: content,
		}}, ""
	default:
		rendered := make([]node, 0, 1+len(children))
		if len(source.Content) > 0 {
			rendered = append(
				rendered,
				node{Type: blockNodeNames[blockTypeParagraph], Content: source.Content},
			)
		}

		return append(rendered, hoistNodes(children)...), ""
	}
}

func partitionNodes(values []node, allowed ...string) ([]node, []node) {
	matches := make([]node, 0, len(values))
	spills := make([]node, 0)

	for _, value := range values {
		if slices.Contains(allowed, value.Type) {
			matches = append(matches, value)
		} else {
			spills = append(spills, value)
		}
	}

	return matches, spills
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

func tolerantStringSlice(value any) []string {
	values, ok := value.([]any)
	if !ok {
		return nil
	}

	result := make([]string, 0, len(values))

	for _, value := range values {
		if item, ok := value.(string); ok && item != "" {
			result = append(result, item)
		}
	}

	return result
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
