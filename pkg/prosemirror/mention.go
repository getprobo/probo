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
	"fmt"
	"slices"
	"strings"

	"go.probo.inc/probo/pkg/gid"
)

// MentionedProfileIDs returns the distinct profile ids referenced by mention
// nodes in content, in document order.
func MentionedProfileIDs(content string) ([]gid.GID, error) {
	if strings.TrimSpace(content) == "" {
		return nil, nil
	}

	node, err := Parse(content)
	if err != nil {
		return nil, fmt.Errorf("cannot parse content: %w", err)
	}

	var ids []gid.GID
	if err := collectMentionedProfileIDs(&ids, node); err != nil {
		return nil, err
	}

	return ids, nil
}

// NewlyMentionedProfileIDs returns the profile ids mentioned in newContent
// that are not mentioned in oldContent.
func NewlyMentionedProfileIDs(oldContent string, newContent string) ([]gid.GID, error) {
	previous, err := MentionedProfileIDs(oldContent)
	if err != nil {
		return nil, fmt.Errorf("cannot collect previous mentions: %w", err)
	}

	current, err := MentionedProfileIDs(newContent)
	if err != nil {
		return nil, fmt.Errorf("cannot collect current mentions: %w", err)
	}

	added := make([]gid.GID, 0, len(current))
	for _, id := range current {
		if !slices.Contains(previous, id) {
			added = append(added, id)
		}
	}

	return added, nil
}

func collectMentionedProfileIDs(ids *[]gid.GID, n Node) error {
	if n.Type == NodeMention {
		attrs, err := n.MentionAttrs()
		if err != nil {
			return fmt.Errorf("cannot read mention: %w", err)
		}

		id, err := gid.ParseGID(attrs.ID)
		if err != nil {
			return fmt.Errorf("invalid mention id %q", attrs.ID)
		}

		if !slices.Contains(*ids, id) {
			*ids = append(*ids, id)
		}
	}

	for _, child := range n.Content {
		if err := collectMentionedProfileIDs(ids, child); err != nil {
			return err
		}
	}

	return nil
}
