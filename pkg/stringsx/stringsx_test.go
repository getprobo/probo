// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package stringsx_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/stringsx"
)

func TestNormalizeAlnum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "letters lowercased", input: "Letaido", expected: "letaido"},
		{name: "strips punctuation and spaces", input: "letaido.com - Inc", expected: "letaidocominc"},
		{name: "keeps digits", input: "auth0", expected: "auth0"},
		{name: "empty", input: "", expected: ""},
		{name: "only punctuation", input: "-_.:", expected: ""},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tt.expected, stringsx.NormalizeAlnum(tt.input))
			},
		)
	}
}
