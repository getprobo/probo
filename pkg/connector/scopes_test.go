// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package connector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseScopeString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "", []string{}},
		{"whitespace only", "   ", []string{}},
		{"single", "read:user", []string{"read:user"}},
		{"multi space", "read:user write:user", []string{"read:user", "write:user"}},
		{"multi comma github style", "repo,gist", []string{"gist", "repo"}},
		{"mixed separators", "read:user,write:user", []string{"read:user", "write:user"}},
		{"extra whitespace", "  read:user   write:user  ", []string{"read:user", "write:user"}},
		{"duplicates", "a a b", []string{"a", "b"}},
		{"sorted output", "z y a", []string{"a", "y", "z"}},
		{"github comma with space", "repo, gist", []string{"gist", "repo"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.want, ParseScopeString(c.in))
		})
	}
}

func TestUnionScopes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   [][]string
		want []string
	}{
		{"both empty", [][]string{{}, {}}, []string{}},
		{"first empty", [][]string{{}, {"a", "b"}}, []string{"a", "b"}},
		{"second empty", [][]string{{"a", "b"}, {}}, []string{"a", "b"}},
		{"disjoint", [][]string{{"a"}, {"b"}}, []string{"a", "b"}},
		{"overlap", [][]string{{"a", "b"}, {"b", "c"}}, []string{"a", "b", "c"}},
		{"three sets", [][]string{{"a"}, {"b"}, {"c"}}, []string{"a", "b", "c"}},
		{"deduplicates", [][]string{{"a", "a"}, {"a"}}, []string{"a"}},
		{"drops empty strings", [][]string{{"a", ""}, {""}}, []string{"a"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.want, UnionScopes(c.in...))
		})
	}
}

func TestFormatScopeString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   []string
		want string
	}{
		{"empty", []string{}, ""},
		{"single", []string{"a"}, "a"},
		{"multi sorted", []string{"b", "a"}, "a b"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.want, FormatScopeString(c.in))
		})
	}
}
