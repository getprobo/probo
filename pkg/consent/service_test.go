// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package consent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnonymizeIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ipv4 zeroes last octet",
			input:    "192.168.1.42",
			expected: "192.168.1.0",
		},
		{
			name:     "ipv4 already anonymized",
			input:    "10.0.0.0",
			expected: "10.0.0.0",
		},
		{
			name:     "ipv4 loopback",
			input:    "127.0.0.1",
			expected: "127.0.0.0",
		},
		{
			name:     "ipv6 zeroes last 80 bits",
			input:    "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			expected: "2001:db8:85a3::",
		},
		{
			name:     "ipv6 loopback",
			input:    "::1",
			expected: "::",
		},
		{
			name:     "invalid ip returns empty string",
			input:    "not-an-ip",
			expected: "",
		},
		{
			name:     "empty string returns empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				result := anonymizeIP(tt.input)
				assert.Equal(t, tt.expected, result)
			},
		)
	}
}
