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

	"go.probo.inc/probo/pkg/automerge"
)

// ToSpans converts a ProseMirror document (the JSON the editor stores) into the
// Automerge rich-text spans that seed a collaboration document. It is the
// forward counterpart of Render: writing the returned spans with
// Text.UpdateSpans and then rendering the document reproduces the input, which
// is what lets the server seed a document version's CRDT from its stored
// ProseMirror content without a JavaScript build step.
//
// The traversal is the inverse of render.go: every block node emits a block
// marker carrying its full ancestor path of Automerge block types; a container
// block (blockquote, list item, table cell) folds its first paragraph's inline
// content into itself, matching how @automerge/prosemirror flattens the tree;
// list wrappers are transparent and only their items become blocks; and marks
// map by the shared schema ledger. Unknown nodes and marks are dropped, mirroring
// the renderer's tolerance for schema drift.
func ToSpans(documentJSON string) ([]automerge.SpanInput, error) {
	var document pmNode
	if err := json.Unmarshal([]byte(documentJSON), &document); err != nil {
		return nil, fmt.Errorf("cannot parse ProseMirror document: %w", err)
	}

	if document.Type != "doc" {
		return nil, fmt.Errorf("expected a ProseMirror doc node, got %q", document.Type)
	}

	spans := make([]automerge.SpanInput, 0)
	for _, child := range document.Content {
		spans = emitBlock(spans, child, nil)
	}

	return spans, nil
}

// UpdateSpansConfig returns the span-writing configuration the seeder uses. Marks
// expand after their range by default, matching the frontend adapter, so text
// typed at a mark's trailing edge inherits it.
func UpdateSpansConfig() automerge.UpdateSpansConfig {
	return automerge.UpdateSpansConfig{DefaultExpand: automerge.MarkExpandAfter}
}

type (
	pmNode struct {
		Type    string         `json:"type"`
		Attrs   map[string]any `json:"attrs,omitempty"`
		Content []pmNode       `json:"content,omitempty"`
		Text    string         `json:"text,omitempty"`
		Marks   []pmMark       `json:"marks,omitempty"`
	}

	pmMark struct {
		Type  string         `json:"type"`
		Attrs map[string]any `json:"attrs,omitempty"`
	}
)

// pmNodeToBlock maps a ProseMirror node name to its Automerge block type. List
// items are absent because their type depends on the enclosing list wrapper.
var pmNodeToBlock = map[string]string{
	"paragraph":      blockTypeParagraph,
	"heading":        blockTypeHeading,
	"codeBlock":      blockTypeCode,
	"blockquote":     blockTypeBlockquote,
	"horizontalRule": blockTypeHorizontalRule,
	"table":          blockTypeTable,
	"tableRow":       blockTypeTableRow,
	"tableCell":      blockTypeTableCell,
	"tableHeader":    blockTypeTableHeader,
}

var pmMarkToAutomerge = map[string]string{}

func init() {
	for _, mapping := range markMappings {
		pmMarkToAutomerge[mapping.ProseMirror] = mapping.Automerge
	}
}

func emitBlock(spans []automerge.SpanInput, n pmNode, parents []string) []automerge.SpanInput {
	switch n.Type {
	case "bulletList", "orderedList":
		itemType := blockTypeUnorderedListItem
		if n.Type == "orderedList" {
			itemType = blockTypeOrderedListItem
		}

		for _, item := range n.Content {
			if item.Type != "listItem" {
				continue
			}

			spans = append(spans, blockSpan(itemType, parents, nil))
			// A list item always renders exactly one leading paragraph from its
			// own content, so its first paragraph is always folded in.
			spans = emitContainerChildren(spans, item.Content, childPath(parents, itemType), true)
		}

		return spans
	case "paragraph", "heading", "codeBlock":
		blockType := pmNodeToBlock[n.Type]
		spans = append(spans, blockSpan(blockType, parents, blockAttributes(blockType, n.Attrs)))

		return emitInline(spans, n.Content, childPath(parents, blockType))
	case "blockquote", "tableCell", "tableHeader":
		blockType := pmNodeToBlock[n.Type]
		spans = append(spans, blockSpan(blockType, parents, blockAttributes(blockType, n.Attrs)))

		// These containers only render a leading paragraph from their own content
		// when it is non-empty (or they have no other children), so an empty
		// first paragraph with siblings must stay an explicit block.
		return emitContainerChildren(spans, n.Content, childPath(parents, blockType), false)
	case "table", "tableRow":
		blockType := pmNodeToBlock[n.Type]
		spans = append(spans, blockSpan(blockType, parents, blockAttributes(blockType, n.Attrs)))

		childParents := childPath(parents, blockType)
		for _, child := range n.Content {
			spans = emitBlock(spans, child, childParents)
		}

		return spans
	case "horizontalRule":
		return append(spans, blockSpan(blockTypeHorizontalRule, parents, map[string]any{}))
	default:
		// Unknown block node: drop it rather than abort, mirroring the renderer.
		return spans
	}
}

// emitContainerChildren writes the children of a container block. A container's
// first paragraph may be folded into the container's own inline content (no
// separate block marker), which is how list items, blockquotes, and table cells
// carry their first line; any remaining children become nested blocks.
//
// Whether the fold happens matches how the renderer reconstructs the container.
// A list item always renders one leading paragraph from its own content, so its
// first paragraph is always folded (alwaysFoldFirstParagraph). A blockquote or
// table cell only renders a leading paragraph when its own content is non-empty
// or it has no other children, so an empty first paragraph that has siblings
// must remain an explicit block or it would be lost on the round trip.
func emitContainerChildren(
	spans []automerge.SpanInput,
	children []pmNode,
	childParents []string,
	alwaysFoldFirstParagraph bool,
) []automerge.SpanInput {
	if len(children) > 0 && children[0].Type == "paragraph" {
		fold := alwaysFoldFirstParagraph ||
			len(children) == 1 ||
			len(children[0].Content) > 0
		if fold {
			spans = emitInline(spans, children[0].Content, childParents)

			for _, child := range children[1:] {
				spans = emitBlock(spans, child, childParents)
			}

			return spans
		}
	}

	for _, child := range children {
		spans = emitBlock(spans, child, childParents)
	}

	return spans
}

func emitInline(spans []automerge.SpanInput, inline []pmNode, blockParents []string) []automerge.SpanInput {
	for _, child := range inline {
		switch child.Type {
		case "text":
			if child.Text == "" {
				continue
			}

			span := automerge.SpanInput{Text: child.Text}
			if marks := convertMarks(child.Marks); len(marks) > 0 {
				span.Marks = marks
			}

			spans = append(spans, span)
		case "hardBreak":
			spans = append(spans, hardBreakSpan(blockParents))
		default:
			continue
		}
	}

	return spans
}

func convertMarks(marks []pmMark) map[string]automerge.Scalar {
	if len(marks) == 0 {
		return nil
	}

	result := make(map[string]automerge.Scalar, len(marks))

	for _, mark := range marks {
		name, ok := pmMarkToAutomerge[mark.Type]
		if !ok {
			continue
		}

		if name == "link" {
			payload, err := json.Marshal(
				map[string]any{
					"href":  stringAttribute(mark.Attrs, "href"),
					"title": nullableStringAttribute(mark.Attrs, "title"),
				},
			)
			if err != nil {
				continue
			}

			result[name] = automerge.StringScalar(string(payload))

			continue
		}

		result[name] = automerge.BoolScalar(true)
	}

	return result
}

func blockSpan(blockType string, parents []string, attrs map[string]any) automerge.SpanInput {
	if attrs == nil {
		attrs = map[string]any{}
	}

	return automerge.SpanInput{
		Block: map[string]any{
			"type":    blockType,
			"parents": toAnySlice(parents),
			"attrs":   attrs,
			"isEmbed": false,
		},
	}
}

func hardBreakSpan(parents []string) automerge.SpanInput {
	return automerge.SpanInput{
		Block: map[string]any{
			"type":    blockTypeHardBreak,
			"parents": toAnySlice(parents),
			"attrs":   map[string]any{},
			"isEmbed": true,
		},
	}
}

func blockAttributes(blockType string, attrs map[string]any) map[string]any {
	switch blockType {
	case blockTypeHeading:
		level := intAttribute(attrs, "level", 1)
		if level < 1 || level > 6 {
			level = 1
		}

		return map[string]any{"level": level}
	case blockTypeCode:
		if language, ok := attrs["language"].(string); ok {
			return map[string]any{"language": language}
		}

		return map[string]any{}
	case blockTypeTableCell, blockTypeTableHeader:
		cellAttrs := map[string]any{
			"colspan": intAttribute(attrs, "colspan", 1),
			"rowspan": intAttribute(attrs, "rowspan", 1),
		}

		if colwidth := numberSliceAttribute(attrs, "colwidth"); colwidth != nil {
			cellAttrs["colwidth"] = colwidth
		}

		return cellAttrs
	default:
		return map[string]any{}
	}
}

func childPath(parents []string, blockType string) []string {
	return append(slices.Clone(parents), blockType)
}

func toAnySlice(values []string) []any {
	result := make([]any, len(values))
	for i, value := range values {
		result[i] = value
	}

	return result
}

func numberSliceAttribute(attrs map[string]any, name string) []any {
	values, ok := attrs[name].([]any)
	if !ok {
		return nil
	}

	result := make([]any, 0, len(values))
	for _, value := range values {
		number, ok := value.(float64)
		if !ok {
			return nil
		}

		result = append(result, number)
	}

	return result
}

func stringAttribute(attrs map[string]any, name string) string {
	if value, ok := attrs[name].(string); ok {
		return value
	}

	return ""
}

func nullableStringAttribute(attrs map[string]any, name string) any {
	if value, ok := attrs[name].(string); ok {
		return value
	}

	return nil
}
