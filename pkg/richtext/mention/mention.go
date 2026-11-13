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

package mention

import (
	"bytes"
	"fmt"
	"strings"
	"unicode/utf8"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
)

type nodeSpec struct {
	entityType *uint16
}

// Document is a mention that accepts any GID.
func Document() prosemirror.NodeSpec {
	return nodeSpec{}
}

// Profile is a mention that accepts only membership-profile GIDs.
func Profile() prosemirror.NodeSpec {
	t := coredata.MembershipProfileEntityType

	return nodeSpec{entityType: &t}
}

func (nodeSpec) Name() prosemirror.NodeType {
	return prosemirror.NodeMention
}

func (nodeSpec) Group() prosemirror.Group {
	return prosemirror.GroupInline
}

func (nodeSpec) Content() prosemirror.ContentRule {
	return prosemirror.ContentRule{}
}

func (nodeSpec) Visible() bool {
	return true
}

// TextLength counts a mention as "@" plus its label, matching the editor's
// content-length accounting for comments and notes.
func (nodeSpec) TextLength(n prosemirror.Node) int {
	if n, ok := labelRunes(n.Attrs); ok {
		return 1 + n
	}

	attrs, err := n.MentionAttrs()
	if err != nil {
		return 0
	}

	return 1 + utf8.RuneCountInString(attrs.Label)
}

var labelKey = []byte(`"label"`)

// labelRunes counts the label without allocating. It reports false when the
// JSON uses escapes, so the caller can fall back to Unmarshal.
func labelRunes(raw []byte) (int, bool) {
	rest := raw
	for {
		i := bytes.Index(rest, labelKey)
		if i < 0 {
			return 0, false
		}

		rest = rest[i+len(labelKey):]

		rest = bytes.TrimLeft(rest, " \t\r\n")
		if len(rest) == 0 || rest[0] != ':' {
			continue
		}

		rest = bytes.TrimLeft(rest[1:], " \t\r\n")
		if len(rest) == 0 || rest[0] != '"' {
			continue
		}

		rest = rest[1:]
		n := 0

		for len(rest) > 0 {
			if rest[0] == '\\' {
				return 0, false
			}

			if rest[0] == '"' {
				return n, true
			}

			r, size := utf8.DecodeRune(rest)
			if r == utf8.RuneError && size == 1 {
				return 0, false
			}

			n++
			rest = rest[size:]
		}

		return 0, false
	}
}

func (m nodeSpec) Check(n *prosemirror.Node, mode prosemirror.CheckMode) error {
	if err := prosemirror.ValidateLeaf(n); err != nil {
		return fmt.Errorf("cannot validate mention node: %w", err)
	}

	attrs, err := n.MentionAttrs()
	if err != nil {
		return fmt.Errorf("cannot validate mention node: %w", err)
	}

	if strings.TrimSpace(attrs.ID) == "" {
		return fmt.Errorf("cannot validate mention node: invalid id")
	}

	if err := m.validateID(attrs.ID); err != nil {
		return fmt.Errorf("cannot validate mention node: invalid id")
	}

	if strings.TrimSpace(attrs.Label) == "" || len(attrs.Label) > prosemirror.MaxMentionLabelBytes {
		return fmt.Errorf("cannot validate mention node: invalid label")
	}

	if mode != prosemirror.CheckSanitize {
		return nil
	}

	attrs.Label = strings.TrimSpace(attrs.Label)

	if err := n.SetAttrs(attrs); err != nil {
		return fmt.Errorf("cannot validate mention node: %w", err)
	}

	return nil
}

func (m nodeSpec) validateID(id string) error {
	parsed, err := gid.ParseGID(id)
	if err != nil {
		return err
	}

	if m.entityType != nil && parsed.EntityType() != *m.entityType {
		return fmt.Errorf("invalid mention id")
	}

	return nil
}
