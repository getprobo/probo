// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

	"go.gearno.de/x/ref"
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
		{"valid pointer", ref.Ref("hello"), 3, false},
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
		{"valid pointer", ref.Ref("hi"), 5, false},
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

func TestPattern(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		pattern   string
		message   string
		wantError bool
	}{
		{"valid pattern", "abc123", `^[a-z0-9]+$`, "", false},
		{"invalid pattern", "ABC123", `^[a-z0-9]+$`, "", true},
		{"custom message", "invalid", `^valid$`, "must be 'valid'", true},
		{"nil pointer", (*string)(nil), `^test$`, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Pattern(tt.pattern, tt.message)(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("Pattern() error = %v, wantError %v", err, tt.wantError)
			}
			if err != nil && tt.message != "" && err.Message != tt.message {
				t.Errorf("Expected message '%s', got '%s'", tt.message, err.Message)
			}
		})
	}
}

func TestAlphaNumeric(t *testing.T) {
	t.Run("valid alphanumeric", func(t *testing.T) {
		str := "abc123"
		err := AlphaNumeric()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("only letters", func(t *testing.T) {
		str := "abcDEF"
		err := AlphaNumeric()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("only numbers", func(t *testing.T) {
		str := "123456"
		err := AlphaNumeric()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("contains spaces", func(t *testing.T) {
		str := "abc 123"
		err := AlphaNumeric()(&str)
		if err == nil {
			t.Error("expected validation error")
		}
	})

	t.Run("contains special characters", func(t *testing.T) {
		str := "abc-123"
		err := AlphaNumeric()(&str)
		if err == nil {
			t.Error("expected validation error")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		str := ""
		err := AlphaNumeric()(&str)
		if err != nil {
			t.Errorf("expected no error for empty string, got: %v", err)
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var str *string
		err := AlphaNumeric()(str)
		if err != nil {
			t.Errorf("expected no error for nil, got: %v", err)
		}
	})
}

func TestNoSpaces(t *testing.T) {
	t.Run("no spaces", func(t *testing.T) {
		str := "hello-world"
		err := NoSpaces()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("contains spaces", func(t *testing.T) {
		str := "hello world"
		err := NoSpaces()(&str)
		if err == nil {
			t.Error("expected validation error")
		}
	})

	t.Run("multiple spaces", func(t *testing.T) {
		str := "hello  world  test"
		err := NoSpaces()(&str)
		if err == nil {
			t.Error("expected validation error")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		str := ""
		err := NoSpaces()(&str)
		if err != nil {
			t.Errorf("expected no error for empty string, got: %v", err)
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var str *string
		err := NoSpaces()(str)
		if err != nil {
			t.Errorf("expected no error for nil, got: %v", err)
		}
	})
}

func TestSlug(t *testing.T) {
	t.Run("valid slug", func(t *testing.T) {
		str := "hello-world"
		err := Slug()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid slug with numbers", func(t *testing.T) {
		str := "hello-world-123"
		err := Slug()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("single word", func(t *testing.T) {
		str := "hello"
		err := Slug()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("contains uppercase", func(t *testing.T) {
		str := "Hello-World"
		err := Slug()(&str)
		if err == nil {
			t.Error("expected validation error for uppercase")
		}
	})

	t.Run("contains spaces", func(t *testing.T) {
		str := "hello world"
		err := Slug()(&str)
		if err == nil {
			t.Error("expected validation error for spaces")
		}
	})

	t.Run("contains underscores", func(t *testing.T) {
		str := "hello_world"
		err := Slug()(&str)
		if err == nil {
			t.Error("expected validation error for underscores")
		}
	})

	t.Run("starts with hyphen", func(t *testing.T) {
		str := "-hello"
		err := Slug()(&str)
		if err == nil {
			t.Error("expected validation error for leading hyphen")
		}
	})

	t.Run("ends with hyphen", func(t *testing.T) {
		str := "hello-"
		err := Slug()(&str)
		if err == nil {
			t.Error("expected validation error for trailing hyphen")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		str := ""
		err := Slug()(&str)
		if err != nil {
			t.Errorf("expected no error for empty string, got: %v", err)
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var str *string
		err := Slug()(str)
		if err != nil {
			t.Errorf("expected no error for nil, got: %v", err)
		}
	})
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
