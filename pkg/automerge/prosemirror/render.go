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
					node{Type: "hardBreak"},
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
			if len(blocks) == 0 {
				blocks = append(blocks, block{Type: blockTypeParagraph})
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

		return node{Type: "paragraph", Content: source.Content}, "", nil
	case blockTypeHeading:
		if len(children) > 0 {
			return node{}, "", fmt.Errorf("heading Automerge block cannot contain child blocks")
		}

		level := intAttribute(source.Attrs, "level", 1)
		if level < 1 || level > 6 {
			level = 1
		}

		return node{
			Type:    "heading",
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

		return node{Type: "codeBlock", Attrs: attrs, Content: source.Content}, "", nil
	case blockTypeHorizontalRule:
		if len(source.Content) > 0 || len(children) > 0 {
			return node{}, "", fmt.Errorf("horizontal rule Automerge block cannot contain content")
		}

		return node{Type: "horizontalRule"}, "", nil
	case blockTypeBlockquote:
		paragraph := node{Type: "paragraph", Content: source.Content}

		return node{
			Type:    "blockquote",
			Content: append([]node{paragraph}, children...),
		}, "", nil
	case blockTypeOrderedListItem:
		paragraph := node{Type: "paragraph", Content: source.Content}

		return node{
			Type:    "listItem",
			Content: append([]node{paragraph}, children...),
		}, "orderedList", nil
	case blockTypeUnorderedListItem:
		paragraph := node{Type: "paragraph", Content: source.Content}

		return node{
			Type:    "listItem",
			Content: append([]node{paragraph}, children...),
		}, "bulletList", nil
	default:
		return node{}, "", fmt.Errorf("unsupported Automerge block type %q", source.Type)
	}
}

func renderMarks(values map[string]any) ([]mark, error) {
	if len(values) == 0 {
		return nil, nil
	}

	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}

	sort.Strings(names)

	marks := make([]mark, 0, len(names))
	for _, name := range names {
		switch name {
		case "strong":
			marks = append(marks, mark{Type: "bold"})
		case "em":
			marks = append(marks, mark{Type: "italic"})
		case "strike", "underline", "code":
			marks = append(marks, mark{Type: name})
		case "link":
			raw, ok := values[name].(string)
			if !ok {
				return nil, fmt.Errorf("link mark value must be a string")
			}

			var attrs map[string]any
			if err := json.Unmarshal([]byte(raw), &attrs); err != nil {
				return nil, fmt.Errorf("cannot decode link mark: %w", err)
			}

			marks = append(marks, mark{Type: "link", Attrs: attrs})
		default:
			return nil, fmt.Errorf("unsupported Automerge mark %q", name)
		}
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

func hasPrefix(values, prefix []string) bool {
	return len(values) >= len(prefix) && slices.Equal(values[:len(prefix)], prefix)
}
