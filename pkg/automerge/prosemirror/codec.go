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
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/prosemirror"
)

// blockRole classifies how a node codec participates in the shared parse and
// render traversals. The role, not a per-node switch, decides whether a block
// holds inline content, folds a leading paragraph, nests block children, or
// stands alone as a leaf.
type blockRole int

const (
	// roleWrapper is a transparent list wrapper (bulletList, orderedList,
	// taskList). It owns no Automerge block of its own; only its items become
	// blocks, and the renderer recreates it from an item's Outer wrapper.
	roleWrapper blockRole = iota

	// roleTextblock holds inline content directly (paragraph, heading, code
	// block).
	roleTextblock

	// roleContainer folds a leading paragraph into its own inline content and
	// nests any remaining children as blocks (blockquote, list items, table
	// cells).
	roleContainer

	// roleStructural nests only block children (table, table row).
	roleStructural

	// roleEmbed is a block-level leaf with no inline content (horizontal rule,
	// image, file).
	roleEmbed

	// roleInlineEmbed is an inline leaf emitted inside a text block (hard
	// break).
	roleInlineEmbed
)

type (
	// nodeCodec converts one ProseMirror node type to and from its Automerge
	// block representation. Codecs live here, not on prosemirror.NodeSpec, so
	// the schema package never depends on the Automerge bridge.
	nodeCodec interface {
		// Node is the ProseMirror node type this codec handles. For list items
		// it is the item node (listItem, taskItem); the enclosing wrapper is
		// reported by Outer.
		Node() prosemirror.NodeType

		// Block is the Automerge block type. It is empty for wrappers, which
		// own no block of their own.
		Block() string

		// Role selects the shared parse and render behavior.
		Role() blockRole

		// Outer is the ProseMirror list wrapper for list and task items, and
		// empty for every other node.
		Outer() prosemirror.NodeType

		// AllowedUnder reports whether the block may appear directly beneath the
		// given ancestor block-type path.
		AllowedUnder(parents []string) bool

		// ToAttrs normalizes a ProseMirror node's attributes into the Automerge
		// block attributes.
		ToAttrs(attrs map[string]any) map[string]any

		// Render rebuilds the ProseMirror nodes for a block and its
		// already-rendered children. The returned string is the list wrapper
		// node type for list items, and empty otherwise.
		Render(source block, children []node) ([]node, string)
	}

	// markCodec converts one ProseMirror mark type to and from its Automerge
	// scalar value.
	markCodec interface {
		// Mark is the ProseMirror mark type.
		Mark() prosemirror.MarkType

		// Automerge is the Automerge mark name.
		Automerge() string

		// ToScalar encodes a ProseMirror mark as its Automerge scalar. It
		// returns false when the mark cannot be encoded and must be dropped.
		ToScalar(m pmMark) (automerge.Scalar, bool)

		// FromScalar decodes an Automerge mark value into a ProseMirror mark. It
		// returns false when the value is malformed and must be dropped.
		FromScalar(value any) (mark, bool)
	}
)

// nodeCodecs is the authoritative set of node codecs. markCodecs is ordered by
// ProseMirror schema rank because ProseMirror stores marks in that order and the
// renderer must emit them the same way for byte-identical documents.
var (
	nodeCodecs = []nodeCodec{
		paragraphCodec{},
		headingCodec{},
		codeBlockCodec{},
		blockquoteCodec{},
		orderedListItemCodec{},
		unorderedListItemCodec{},
		taskListItemCodec{},
		horizontalRuleCodec{},
		imageCodec{},
		fileCodec{},
		hardBreakCodec{},
		tableCodec{},
		tableRowCodec{},
		tableCellCodec{},
		tableHeaderCodec{},
	}

	markCodecs = []markCodec{
		linkCodec{},
		boolMarkCodec{mark: prosemirror.MarkStrong, automerge: "strong"},
		boolMarkCodec{mark: prosemirror.MarkEm, automerge: "em"},
		boolMarkCodec{mark: prosemirror.MarkStrike, automerge: "strike"},
		boolMarkCodec{mark: prosemirror.MarkUnderline, automerge: "underline"},
		boolMarkCodec{mark: prosemirror.MarkCode, automerge: "code"},
	}
)

// Lookup tables derived from the codec registry in init().
var (
	// parseCodecs maps a ProseMirror node type to its codec for the forward
	// traversal. List items are absent because they are reached through their
	// wrapper, not directly.
	parseCodecs = map[prosemirror.NodeType]nodeCodec{}

	// outerCodecs maps a list wrapper node type to the item codec it contains.
	outerCodecs = map[prosemirror.NodeType]nodeCodec{}

	// renderCodecs maps an Automerge block type to its codec for the reverse
	// traversal. Inline embeds and wrappers are absent: hard breaks are handled
	// while collecting blocks, and wrappers are recreated from item codecs.
	renderCodecs = map[string]nodeCodec{}

	markCodecsByPM        = map[prosemirror.MarkType]markCodec{}
	markCodecsByAutomerge = map[string]markCodec{}

	blockNodeNames  = map[string]string{}
	markNodeNames   = map[string]string{}
	markRenderOrder = map[string]int{}
)

func init() {
	for _, codec := range nodeCodecs {
		if codec.Block() != "" {
			blockNodeNames[codec.Block()] = string(codec.Node())
		}

		if codec.Outer() != "" {
			outerCodecs[codec.Outer()] = codec
		} else if codec.Role() != roleInlineEmbed {
			parseCodecs[codec.Node()] = codec
		}

		switch codec.Role() {
		case roleWrapper, roleInlineEmbed:
		default:
			if codec.Block() != "" {
				renderCodecs[codec.Block()] = codec
			}
		}
	}

	for index, codec := range markCodecs {
		markCodecsByPM[codec.Mark()] = codec
		markCodecsByAutomerge[codec.Automerge()] = codec
		markNodeNames[codec.Automerge()] = string(codec.Mark())
		markRenderOrder[codec.Automerge()] = index
	}
}

type baseCodec struct{}

func (baseCodec) Outer() prosemirror.NodeType { return "" }

func (baseCodec) AllowedUnder([]string) bool { return true }

func (baseCodec) ToAttrs(map[string]any) map[string]any { return map[string]any{} }

func acceptsInlineContent(blockType string) bool {
	codec, ok := renderCodecs[blockType]
	if !ok {
		return true
	}

	switch codec.Role() {
	case roleStructural, roleEmbed:
		return false
	default:
		return true
	}
}

func blockAllowedUnder(blockType string, parents []string) bool {
	codec, ok := renderCodecs[blockType]
	if !ok {
		return true
	}

	return codec.AllowedUnder(parents)
}
