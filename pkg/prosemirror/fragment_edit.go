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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrFragmentEditNotFound  = errors.New("prosemirror fragment edit target not found")
	ErrFragmentEditAmbiguous = errors.New("prosemirror fragment edit target matches more than once")
)

type (
	// FragmentEdit replaces one contiguous sibling sequence with another.
	FragmentEdit struct {
		Target      []Node
		Replacement []Node
	}

	fragmentMatch struct {
		parent *Node
		index  int
		length int
	}
)

// ParseFragment unmarshals a ProseMirror node or a JSON array of sibling nodes.
func ParseFragment(raw string) ([]Node, error) {
	trimmed := bytes.TrimSpace([]byte(raw))
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("cannot parse prosemirror fragment: empty")
	}

	if trimmed[0] == '[' {
		var nodes []Node
		if err := json.Unmarshal(trimmed, &nodes); err != nil {
			return nil, fmt.Errorf("cannot parse prosemirror fragment array: %w", err)
		}

		return nodes, nil
	}

	var node Node
	if err := json.Unmarshal(trimmed, &node); err != nil {
		return nil, fmt.Errorf("cannot parse prosemirror fragment node: %w", err)
	}

	return []Node{node}, nil
}

// ApplyFragmentEdits applies each edit in order to a document tree. If any edit
// fails to find exactly one match, none of the later edits run and the original
// document is left unchanged by returning an error before the caller persists.
func ApplyFragmentEdits(doc Node, edits []FragmentEdit) (Node, error) {
	if doc.Type != NodeDoc {
		return Node{}, fmt.Errorf("fragment edits require a doc root, got %q", doc.Type)
	}

	current := cloneNode(doc)

	for i, edit := range edits {
		if len(edit.Target) == 0 {
			return Node{}, fmt.Errorf("fragment edit %d has an empty target", i)
		}

		matches := findFragmentMatches(&current, edit.Target)
		switch len(matches) {
		case 0:
			return Node{}, fmt.Errorf("fragment edit %d: %w", i, ErrFragmentEditNotFound)
		case 1:
			applyFragmentMatch(matches[0], edit.Replacement)
		default:
			return Node{}, fmt.Errorf("fragment edit %d: %w", i, ErrFragmentEditAmbiguous)
		}
	}

	return current, nil
}

func findFragmentMatches(root *Node, target []Node) []fragmentMatch {
	var matches []fragmentMatch
	collectFragmentMatches(root, target, &matches)

	return matches
}

func collectFragmentMatches(parent *Node, target []Node, matches *[]fragmentMatch) {
	if parent == nil {
		return
	}

	if len(parent.Content) >= len(target) {
		for i := 0; i <= len(parent.Content)-len(target); i++ {
			if nodesEqual(parent.Content[i:i+len(target)], target) {
				*matches = append(
					*matches,
					fragmentMatch{
						parent: parent,
						index:  i,
						length: len(target),
					},
				)
			}
		}
	}

	for i := range parent.Content {
		collectFragmentMatches(&parent.Content[i], target, matches)
	}
}

func applyFragmentMatch(match fragmentMatch, replacement []Node) {
	content := match.parent.Content
	next := make([]Node, 0, len(content)-match.length+len(replacement))
	next = append(next, content[:match.index]...)
	next = append(next, replacement...)
	next = append(next, content[match.index+match.length:]...)
	match.parent.Content = next
}

func nodesEqual(a, b []Node) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !nodeEqual(a[i], b[i]) {
			return false
		}
	}

	return true
}

func nodeEqual(a, b Node) bool {
	left, err := json.Marshal(a)
	if err != nil {
		return false
	}

	right, err := json.Marshal(b)
	if err != nil {
		return false
	}

	return bytes.Equal(left, right)
}

func cloneNode(n Node) Node {
	raw, err := json.Marshal(n)
	if err != nil {
		panic(fmt.Errorf("cannot clone prosemirror node: %w", err))
	}

	var out Node
	if err := json.Unmarshal(raw, &out); err != nil {
		panic(fmt.Errorf("cannot clone prosemirror node: %w", err))
	}

	return out
}
