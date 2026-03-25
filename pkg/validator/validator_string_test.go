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

package validator

import (
	"testing"
)

func TestMinLen(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		minLen    int
		wantError bool
	}{
		{"valid string", "hello", 3, false},
		{"exact length", "hello", 5, false},
		{"too short", "hi", 5, true},
		{"nil pointer", (*string)(nil), 5, false}, // Skip validation
		{"valid pointer", new("hello"), 3, false},
		{"non-string", 123, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MinLen(tt.minLen)(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("MinLen() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestMaxLen(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		maxLen    int
		wantError bool
	}{
		{"valid string", "hello", 10, false},
		{"exact length", "hello", 5, false},
		{"too long", "hello world", 5, true},
		{"nil pointer", (*string)(nil), 5, false}, // Skip validation
		{"valid pointer", new("hi"), 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MaxLen(tt.maxLen)(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("MaxLen() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestOneOf(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		allowed   []string
		wantError bool
	}{
		{"valid value", "apple", []string{"apple", "banana", "orange"}, false},
		{"invalid value", "grape", []string{"apple", "banana", "orange"}, true},
		{"nil pointer", (*string)(nil), []string{"apple"}, false},
		{"empty string", "", []string{"apple", ""}, false},
		{"non-string", 123, []string{"apple"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := OneOfSlice(tt.allowed)(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("OneOfSlice() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
