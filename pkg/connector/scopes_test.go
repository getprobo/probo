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
