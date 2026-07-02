// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

func TestContainsSubstring(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		substr    string
		wantError bool
	}{
		{"contains substring", "hello {{cookie_policy_link}} world", "{{cookie_policy_link}}", false},
		{"missing substring", "hello world", "{{cookie_policy_link}}", true},
		{"exact match", "{{cookie_policy_link}}", "{{cookie_policy_link}}", false},
		{"empty string", "", "{{cookie_policy_link}}", true},
		{"nil pointer", (*string)(nil), "{{cookie_policy_link}}", false},
		{"non-string", 123, "foo", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ContainsSubstring(tt.substr)(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ContainsSubstring() error = %v, wantError %v", err, tt.wantError)
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
