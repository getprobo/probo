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

package suggestion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDescribeChange(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		change suggestionChange
		want   string
	}{
		{
			name:   "text replaced",
			change: suggestionChange{Kind: "TEXT", Deleted: new("old"), Inserted: new("new")},
			want:   `replaced "old" with "new"`,
		},
		{
			name:   "text deleted",
			change: suggestionChange{Kind: "TEXT", Deleted: new("old"), Inserted: new("")},
			want:   `deleted "old"`,
		},
		{
			name:   "text inserted",
			change: suggestionChange{Kind: "TEXT", Deleted: new(""), Inserted: new("new")},
			want:   `added "new"`,
		},
		{
			name:   "format added",
			change: suggestionChange{Kind: "FORMAT", Mark: new("BOLD"), FormatAction: new("ADD")},
			want:   "added bold",
		},
		{
			name:   "format removed",
			change: suggestionChange{Kind: "FORMAT", Mark: new("ITALIC"), FormatAction: new("REMOVE")},
			want:   "removed italic",
		},
		{
			name: "link added",
			change: suggestionChange{
				Kind:         "FORMAT",
				Mark:         new("LINK"),
				FormatAction: new("ADD"),
				Href:         new("https://example.com"),
			},
			want: "added link to https://example.com",
		},
		{
			name: "link changed",
			change: suggestionChange{
				Kind:         "FORMAT",
				Mark:         new("LINK"),
				FormatAction: new("ADD"),
				Href:         new("https://new.example.com"),
				PreviousHref: new("https://old.example.com"),
			},
			want: "changed link from https://old.example.com to https://new.example.com",
		},
		{
			name: "block with a level",
			change: suggestionChange{
				Kind:   "BLOCK",
				Before: &blockLabel{Kind: "HEADING", Level: new(2)},
				After:  &blockLabel{Kind: "PARAGRAPH"},
			},
			want: "heading 2 -> paragraph",
		},
		{
			name: "block without a level",
			change: suggestionChange{
				Kind:   "BLOCK",
				Before: &blockLabel{Kind: "PARAGRAPH"},
				After:  &blockLabel{Kind: "BULLET_LIST"},
			},
			want: "paragraph -> bullet list",
		},
		{
			name:   "block with a missing side",
			change: suggestionChange{Kind: "BLOCK", After: &blockLabel{Kind: "QUOTE"}},
			want:   "block -> quote",
		},
		{
			name:   "node added",
			change: suggestionChange{Kind: "NODE", Node: new("ROW"), NodeAction: new("INSERT")},
			want:   "added row",
		},
		{
			name:   "node removed",
			change: suggestionChange{Kind: "NODE", Node: new("TABLE"), NodeAction: new("DELETE")},
			want:   "removed table",
		},
		{
			name:   "unknown kind",
			change: suggestionChange{Kind: "OTHER"},
			want:   "other",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, describeChange(tc.change))
		})
	}
}
