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

	"go.probo.inc/probo/pkg/gid"
)

const (
	// MaxFileNameBytes is the longest file name a file node accepts.
	MaxFileNameBytes = 1024

	// MaxMentionLabelBytes is the longest label a mention node accepts.
	MaxMentionLabelBytes = 255

	// maxSchemaNodes caps the node count so a ContentRule fits in a uint64
	// bitset of allowed child indices.
	maxSchemaNodes = 64
)

type (
	// CheckMode selects the behavior of a NodeSpec or MarkSpec Check. In
	// CheckSanitize mode a Check also normalizes attributes in place.
	CheckMode int

	// Group classifies where a node may appear in parent content rules.
	Group uint8
)

const (
	// CheckValidate verifies a node without mutating it.
	CheckValidate CheckMode = iota

	// CheckSanitize verifies a node and normalizes its attributes in place.
	CheckSanitize
)

const (
	// GroupNone is a node that no content rule references by group.
	GroupNone Group = iota

	// GroupBlock is a block-level node.
	GroupBlock

	// GroupInline is an inline node.
	GroupInline
)

type (
	// ContentRule describes the children a node may contain. Groups and Types
	// are the allowed child kinds; Min is the fewest children required; First
	// constrains the first child; RejectChildMarks forbids marks on children
	// (code block text).
	ContentRule struct {
		Groups           []Group
		Types            []NodeType
		Min              int
		First            NodeType
		RejectChildMarks bool
	}

	// NodeSpec describes one node type in a Schema.
	NodeSpec interface {
		Name() NodeType
		Group() Group
		Content() ContentRule
		Check(n *Node, mode CheckMode) error
		RenderHTML(w *HTMLWriter, n Node) error
		RenderMarkdown(w *MarkdownWriter, n Node) error
	}

	// MarkSpec describes one mark type in a Schema.
	MarkSpec interface {
		Name() MarkType
		Check(m *Mark, mode CheckMode) error
		OpenHTML(w *HTMLWriter, m Mark) error
		CloseHTML(w *HTMLWriter, m Mark)
		OpenMarkdown(w *MarkdownWriter, m Mark) error
		CloseMarkdown(w *MarkdownWriter, m Mark) error
	}

	// textLengthSpec is an optional NodeSpec capability contributing to
	// Schema.TextLength. Nodes without it contribute only through children.
	textLengthSpec interface {
		TextLength(n Node) int
	}

	// visibleSpec is an optional NodeSpec capability marking a node as visible
	// content for Schema.HasVisibleContent regardless of its text.
	visibleSpec interface {
		Visible() bool
	}
)

type (
	// Schema is a compiled set of node and mark specs. It drives validation,
	// sanitization, and rendering.
	Schema struct {
		index   map[NodeType]uint8
		entries []nodeEntry
		marks   map[MarkType]MarkSpec
	}

	nodeEntry struct {
		spec    NodeSpec
		group   Group
		content compiledContent
		textLen textLengthSpec
		visible bool
	}

	compiledContent struct {
		allowed          uint64
		min              int
		expected         string
		first            NodeType
		hasFirst         bool
		rejectChildMarks bool
	}
)

// NewSchema compiles nodes and marks into a Schema. It panics on
// misconfiguration: more than 64 nodes, a missing doc or text node, a content
// rule that references an unknown node, or a duplicate node or mark name.
func NewSchema(nodes []NodeSpec, marks []MarkSpec) *Schema {
	if len(nodes) == 0 {
		panic("prosemirror: schema requires at least one node")
	}

	if len(nodes) > maxSchemaNodes {
		panic(fmt.Sprintf("prosemirror: schema supports at most %d nodes", maxSchemaNodes))
	}

	index := make(map[NodeType]uint8, len(nodes))
	groupBits := make(map[Group]uint64, len(nodes))

	for i, spec := range nodes {
		name := spec.Name()
		if _, dup := index[name]; dup {
			panic(fmt.Sprintf("prosemirror: duplicate node %q", name))
		}

		index[name] = uint8(i)
		groupBits[spec.Group()] |= 1 << uint(i)
	}

	if _, ok := index[NodeDoc]; !ok {
		panic("prosemirror: schema requires a doc node")
	}

	if _, ok := index[NodeText]; !ok {
		panic("prosemirror: schema requires a text node")
	}

	entries := make([]nodeEntry, len(nodes))

	for i, spec := range nodes {
		content, err := compileContent(spec.Content(), index, groupBits)
		if err != nil {
			panic(fmt.Sprintf("prosemirror: node %q: %v", spec.Name(), err))
		}

		entry := nodeEntry{
			spec:    spec,
			group:   spec.Group(),
			content: content,
		}

		if tl, ok := spec.(textLengthSpec); ok {
			entry.textLen = tl
		}

		if v, ok := spec.(visibleSpec); ok {
			entry.visible = v.Visible()
		}

		entries[i] = entry
	}

	markIndex := make(map[MarkType]MarkSpec, len(marks))

	for _, spec := range marks {
		name := spec.Name()
		if _, dup := markIndex[name]; dup {
			panic(fmt.Sprintf("prosemirror: duplicate mark %q", name))
		}

		markIndex[name] = spec
	}

	return &Schema{
		index:   index,
		entries: entries,
		marks:   markIndex,
	}
}

func compileContent(
	rule ContentRule,
	index map[NodeType]uint8,
	groupBits map[Group]uint64,
) (compiledContent, error) {
	var allowed uint64

	for _, g := range rule.Groups {
		allowed |= groupBits[g]
	}

	for _, t := range rule.Types {
		idx, ok := index[t]
		if !ok {
			return compiledContent{}, fmt.Errorf("content references unknown node %q", t)
		}

		allowed |= 1 << uint(idx)
	}

	hasFirst := rule.First != ""
	if hasFirst {
		if _, ok := index[rule.First]; !ok {
			return compiledContent{}, fmt.Errorf("content references unknown node %q", rule.First)
		}
	}

	expected := ""
	switch {
	case len(rule.Types) > 0:
		expected = string(rule.Types[0])
	case len(rule.Groups) > 0:
		expected = groupName(rule.Groups[0])
	}

	return compiledContent{
		allowed:          allowed,
		min:              rule.Min,
		expected:         expected,
		first:            rule.First,
		hasFirst:         hasFirst,
		rejectChildMarks: rule.RejectChildMarks,
	}, nil
}

func groupName(g Group) string {
	switch g {
	case GroupBlock:
		return "block"
	case GroupInline:
		return "inline"
	default:
		return ""
	}
}

// knownNodeTypes lists every node type the package defines, so validation can
// tell a type excluded from a schema from a genuinely unknown one.
var knownNodeTypes = map[NodeType]struct{}{
	NodeDoc:            {},
	NodeParagraph:      {},
	NodeBlockquote:     {},
	NodeHeading:        {},
	NodeCodeBlock:      {},
	NodeHorizontalRule: {},
	NodeHardBreak:      {},
	NodeText:           {},
	NodeImage:          {},
	NodeFile:           {},
	NodeBulletList:     {},
	NodeOrderedList:    {},
	NodeListItem:       {},
	NodeTaskList:       {},
	NodeTaskItem:       {},
	NodeTable:          {},
	NodeTableRow:       {},
	NodeTableCell:      {},
	NodeTableHeader:    {},
	NodeMention:        {},
}

// Validate reports whether n is a valid document for the schema.
func (s *Schema) Validate(n Node) error {
	if n.Type != NodeDoc {
		return fmt.Errorf("document content root must be type %q", NodeDoc)
	}

	return s.validateNode(&n, CheckValidate)
}

func (s *Schema) validateNode(n *Node, mode CheckMode) error {
	idx, ok := s.index[n.Type]
	if !ok {
		return s.unknownNode(n.Type)
	}

	entry := &s.entries[idx]
	if err := entry.spec.Check(n, mode); err != nil {
		return err
	}

	if n.Type == NodeText {
		for i := range n.Marks {
			mark, ok := s.marks[n.Marks[i].Type]
			if !ok {
				return fmt.Errorf("cannot validate mark: unknown type %q", n.Marks[i].Type)
			}

			if err := mark.Check(&n.Marks[i], mode); err != nil {
				return err
			}
		}
	}

	return s.checkContent(n, entry.content, mode)
}

func (s *Schema) checkContent(n *Node, c compiledContent, mode CheckMode) error {
	if c.hasFirst {
		if len(n.Content) == 0 || n.Content[0].Type != c.first {
			return fmt.Errorf("%s must start with a %s", n.Type, c.first)
		}

		for i := range n.Content {
			child := &n.Content[i]
			if i > 0 && !s.childAllowed(c, child.Type) {
				return fmt.Errorf("invalid child type %q in %s (expected %s)", child.Type, n.Type, c.expected)
			}

			if err := s.validateNode(child, mode); err != nil {
				return err
			}
		}

		return nil
	}

	if len(n.Content) < c.min {
		return fmt.Errorf("%s must contain at least %d %s", n.Type, c.min, c.expected)
	}

	for i := range n.Content {
		child := &n.Content[i]
		if !s.childAllowed(c, child.Type) {
			return fmt.Errorf("invalid child type %q in %s (expected %s)", child.Type, n.Type, c.expected)
		}

		if c.rejectChildMarks && len(child.Marks) > 0 {
			return fmt.Errorf("%s text cannot have marks", n.Type)
		}

		if err := s.validateNode(child, mode); err != nil {
			return err
		}
	}

	return nil
}

func (s *Schema) childAllowed(c compiledContent, t NodeType) bool {
	idx, ok := s.index[t]
	if !ok {
		return false
	}

	return c.allowed&(1<<uint(idx)) != 0
}

func (s *Schema) unknownNode(t NodeType) error {
	if _, known := knownNodeTypes[t]; known {
		return fmt.Errorf("%s node is not allowed in this schema", t)
	}

	return fmt.Errorf("cannot validate node: unknown type %q", t)
}

// Parse unmarshals raw ProseMirror JSON and requires a doc root.
func (s *Schema) Parse(raw string) (Node, error) {
	n, err := Parse(raw)
	if err != nil {
		return Node{}, fmt.Errorf("cannot parse document content as ProseMirror JSON: %w", err)
	}

	if n.Type != NodeDoc {
		return Node{}, fmt.Errorf("document content root must be type %q", NodeDoc)
	}

	return n, nil
}

// Sanitize parses, validates, and normalizes raw ProseMirror JSON, returning
// the re-serialized document.
func (s *Schema) Sanitize(raw string) (string, error) {
	n, err := s.Parse(raw)
	if err != nil {
		return "", err
	}

	if err := s.validateNode(&n, CheckSanitize); err != nil {
		return "", fmt.Errorf("cannot sanitize document content: %w", err)
	}

	out, err := json.Marshal(n)
	if err != nil {
		return "", fmt.Errorf("cannot marshal sanitized document: %w", err)
	}

	return string(out), nil
}

// TextLength returns the number of text characters across the node tree.
func (s *Schema) TextLength(n Node) int {
	length := 0

	if idx, ok := s.index[n.Type]; ok {
		if tl := s.entries[idx].textLen; tl != nil {
			length += tl.TextLength(n)
		}
	}

	for _, child := range n.Content {
		length += s.TextLength(child)
	}

	return length
}

// HasVisibleContent reports whether the node tree renders any visible content:
// a node marked visible, or non-blank text.
func (s *Schema) HasVisibleContent(n Node) bool {
	if idx, ok := s.index[n.Type]; ok {
		if s.entries[idx].visible {
			return true
		}
	}

	if n.Text != nil && strings.TrimSpace(*n.Text) != "" {
		return true
	}

	for _, child := range n.Content {
		if s.HasVisibleContent(child) {
			return true
		}
	}

	return false
}

func validFileID(id string) bool {
	_, err := gid.ParseGID(id)

	return err == nil
}

func validateLeaf(n *Node) error {
	if len(n.Content) > 0 {
		return fmt.Errorf("%s cannot have children", n.Type)
	}

	if err := rejectNonTextPayload(n); err != nil {
		return fmt.Errorf("cannot validate %s node: %w", n.Type, err)
	}

	return nil
}

func rejectNonTextPayload(n *Node) error {
	if n.Text != nil {
		return fmt.Errorf("%s cannot have text", n.Type)
	}

	if len(n.Marks) > 0 {
		return fmt.Errorf("%s cannot have marks", n.Type)
	}

	return nil
}
