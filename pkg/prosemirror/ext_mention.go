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
	"strings"

	"go.probo.inc/probo/pkg/gid"
)

// MentionOptions configures a mention node. EntityTypes, when non-empty,
// restricts mentions to GIDs of those entity types; empty accepts any GID.
type MentionOptions struct {
	EntityTypes []uint16
}

type mentionNodeSpec struct {
	entityTypes []uint16
}

// Mention is a leaf node referencing an entity by GID with a display label.
func Mention(opts MentionOptions) NodeSpec {
	return mentionNodeSpec{entityTypes: opts.EntityTypes}
}

func (mentionNodeSpec) Name() NodeType {
	return NodeMention
}

func (mentionNodeSpec) Group() Group {
	return GroupInline
}

func (mentionNodeSpec) Content() ContentRule {
	return ContentRule{}
}

func (mentionNodeSpec) Visible() bool {
	return true
}

func (m mentionNodeSpec) Check(n *Node, mode CheckMode) error {
	if err := validateLeaf(n); err != nil {
		return fmt.Errorf("cannot validate mention node: %w", err)
	}

	attrs, err := n.MentionAttrs()
	if err != nil {
		return fmt.Errorf("cannot validate mention node: %w", err)
	}

	parsed, err := gid.ParseGID(attrs.ID)
	if err != nil {
		return fmt.Errorf("cannot validate mention node: invalid id")
	}

	if len(m.entityTypes) > 0 && !slices.Contains(m.entityTypes, parsed.EntityType()) {
		return fmt.Errorf("cannot validate mention node: invalid id")
	}

	if strings.TrimSpace(attrs.Label) == "" || len(attrs.Label) > MaxMentionLabelBytes {
		return fmt.Errorf("cannot validate mention node: invalid label")
	}

	if mode != CheckSanitize {
		return nil
	}

	attrs.Label = strings.TrimSpace(attrs.Label)

	raw, err := json.Marshal(attrs)
	if err != nil {
		return nil
	}

	n.Attrs = raw

	return nil
}

func (mentionNodeSpec) RenderHTML(w *HTMLWriter, n Node) error {
	attrs, err := n.MentionAttrs()
	if err != nil {
		return fmt.Errorf("cannot render mention node: %w", err)
	}

	w.WriteString(`<span class="mention" data-type="mention"`)
	w.WriteAttr("data-id", attrs.ID)
	w.WriteAttr("data-label", attrs.Label)
	w.WriteByte('>')
	w.WriteEscaped("@" + attrs.Label)
	w.WriteString("</span>")

	return nil
}

func (mentionNodeSpec) RenderMarkdown(w *MarkdownWriter, n Node) error {
	attrs, err := n.MentionAttrs()
	if err != nil {
		return fmt.Errorf("cannot render mention node: %w", err)
	}

	w.ensurePrefix()
	w.buf.WriteString(escapeMarkdown("@" + attrs.Label))

	return nil
}
