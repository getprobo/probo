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
	"strings"
	"testing"
)

func TestNoHTML(t *testing.T) {
	t.Run("valid text without HTML", func(t *testing.T) {
		str := "This is a normal text"
		err := NoHTML()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid text with special characters", func(t *testing.T) {
		str := "Price: $10.99 - 20% off!"
		err := NoHTML()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid UTF-8 text", func(t *testing.T) {
		str := "José García 张伟"
		err := NoHTML()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid text with emojis", func(t *testing.T) {
		str := "Hello World 🌍"
		err := NoHTML()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid - script tag XSS", func(t *testing.T) {
		str := "<script>alert('xss')</script>"
		err := NoHTML()(&str)
		if err == nil {
			t.Fatal("expected validation error for script tag")
		}
		if !strings.Contains(err.Message, "HTML tags") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - simple bold tag", func(t *testing.T) {
		str := "Hello <b>World</b>"
		err := NoHTML()(&str)
		if err == nil {
			t.Fatal("expected validation error for bold tag")
		}
		if !strings.Contains(err.Message, "HTML tags") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - div tag", func(t *testing.T) {
		str := "<div>Content</div>"
		err := NoHTML()(&str)
		if err == nil {
			t.Error("expected validation error for div tag")
		}
	})

	t.Run("invalid - self-closing tag", func(t *testing.T) {
		str := "Line break<br/>here"
		err := NoHTML()(&str)
		if err == nil {
			t.Error("expected validation error for self-closing tag")
		}
	})

	t.Run("invalid - img tag", func(t *testing.T) {
		str := `<img src="x" onerror="alert(1)">`
		err := NoHTML()(&str)
		if err == nil {
			t.Error("expected validation error for img tag")
		}
	})

	t.Run("invalid - anchor tag", func(t *testing.T) {
		str := `<a href="javascript:alert(1)">Click</a>`
		err := NoHTML()(&str)
		if err == nil {
			t.Error("expected validation error for anchor tag")
		}
	})

	t.Run("invalid - less than symbol", func(t *testing.T) {
		str := "5 < 10"
		err := NoHTML()(&str)
		if err == nil {
			t.Fatal("expected validation error for angle bracket")
		}
		if !strings.Contains(err.Message, "angle brackets") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - greater than symbol", func(t *testing.T) {
		str := "10 > 5"
		err := NoHTML()(&str)
		if err == nil {
			t.Error("expected validation error for angle bracket")
		}
	})

	t.Run("invalid - both angle brackets", func(t *testing.T) {
		str := "5 < x > 10"
		err := NoHTML()(&str)
		if err == nil {
			t.Error("expected validation error for angle brackets")
		}
	})

	t.Run("invalid - malformed tag", func(t *testing.T) {
		str := "text <incomplete"
		err := NoHTML()(&str)
		if err == nil {
			t.Error("expected validation error for incomplete tag")
		}
	})

	t.Run("invalid - encoded attempt", func(t *testing.T) {
		str := "<ScRiPt>alert(1)</ScRiPt>"
		err := NoHTML()(&str)
		if err == nil {
			t.Error("expected validation error for mixed case script tag")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		str := ""
		err := NoHTML()(&str)
		if err != nil {
			t.Errorf("expected no error for empty string, got: %v", err)
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var str *string
		err := NoHTML()(str)
		if err != nil {
			t.Errorf("expected no error for nil, got: %v", err)
		}
	})

	t.Run("not a string", func(t *testing.T) {
		num := 123
		err := NoHTML()(&num)
		if err == nil {
			t.Fatal("expected validation error for non-string")
		}
		if !strings.Contains(err.Message, "must be a string") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("combined with other validators", func(t *testing.T) {
		v := New()
		title := "Product Title 2024"
		v.Check(&title, "title", Required(), NoHTML(), MinLen(3), MaxLen(100))

		if v.HasErrors() {
			t.Errorf("expected no errors, got: %v", v.Errors())
		}
	})

	t.Run("combined with PrintableText", func(t *testing.T) {
		v := New()
		title := "José García-O'Brien"
		v.Check(&title, "title", Required(), NoHTML(), PrintableText(), MinLen(3), MaxLen(100))

		if v.HasErrors() {
			t.Errorf("expected no errors, got: %v", v.Errors())
		}
	})

	t.Run("combined validators catch XSS", func(t *testing.T) {
		v := New()
		malicious := "<script>alert('xss')</script>"
		v.Check(&malicious, "content", Required(), NoHTML(), PrintableText())

		if !v.HasErrors() {
			t.Error("expected validation errors")
		}

		// Should have error from NoHTML
		errors := v.Errors()
		found := false
		for _, err := range errors {
			if strings.Contains(err.Message, "HTML tags") || strings.Contains(err.Message, "angle brackets") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected error about HTML tags or angle brackets")
		}
	})

	t.Run("combined validators catch invisible chars and HTML", func(t *testing.T) {
		v := New()
		malicious := "<b>test\x00text</b>"
		v.Check(&malicious, "content", NoHTML(), PrintableText())

		if !v.HasErrors() {
			t.Error("expected validation errors")
		}

		// Should have at least one error (NoHTML will catch it first)
		if len(v.Errors()) < 1 {
			t.Error("expected at least one validation error")
		}
	})
}

func TestPrintableText(t *testing.T) {
	t.Run("valid UTF-8 text with accents", func(t *testing.T) {
		str := "José García"
		err := PrintableText()(&str)
		if err != nil {
			t.Errorf("expected no error for valid UTF-8 name, got: %v", err)
		}
	})

	t.Run("valid text with emojis", func(t *testing.T) {
		str := "Hello World 🌍"
		err := PrintableText()(&str)
		if err != nil {
			t.Errorf("expected no error for emojis, got: %v", err)
		}
	})

	t.Run("valid Chinese characters", func(t *testing.T) {
		str := "张伟"
		err := PrintableText()(&str)
		if err != nil {
			t.Errorf("expected no error for Chinese characters, got: %v", err)
		}
	})

	t.Run("valid Arabic text", func(t *testing.T) {
		str := "محمد"
		err := PrintableText()(&str)
		if err != nil {
			t.Errorf("expected no error for Arabic text, got: %v", err)
		}
	})

	t.Run("valid Cyrillic text", func(t *testing.T) {
		str := "Александр"
		err := PrintableText()(&str)
		if err != nil {
			t.Errorf("expected no error for Cyrillic text, got: %v", err)
		}
	})

	t.Run("valid text with apostrophe and hyphen", func(t *testing.T) {
		str := "O'Brien-Smith"
		err := PrintableText()(&str)
		if err != nil {
			t.Errorf("expected no error for apostrophe and hyphen, got: %v", err)
		}
	})

	t.Run("valid text with numbers", func(t *testing.T) {
		str := "Product 2024"
		err := PrintableText()(&str)
		if err != nil {
			t.Errorf("expected no error for text with numbers, got: %v", err)
		}
	})

	t.Run("valid text with punctuation", func(t *testing.T) {
		str := "Hello, World! How are you?"
		err := PrintableText()(&str)
		if err != nil {
			t.Errorf("expected no error for punctuation, got: %v", err)
		}
	})

	t.Run("valid text with angle brackets", func(t *testing.T) {
		str := "5 < 10 > 3"
		err := PrintableText()(&str)
		if err != nil {
			t.Errorf("expected no error for angle brackets (HTML checking is separate), got: %v", err)
		}
	})

	t.Run("invalid - RLO character", func(t *testing.T) {
		str := "test\u202Eexe.txt"
		err := PrintableText()(&str)
		if err == nil {
			t.Fatal("expected validation error for RLO character")
		}
		if !strings.Contains(err.Message, "bidirectional override") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - LRO character", func(t *testing.T) {
		str := "test\u202Dtext"
		err := PrintableText()(&str)
		if err == nil {
			t.Error("expected validation error for LRO character")
		}
	})

	t.Run("invalid - zero-width space", func(t *testing.T) {
		str := "test\u200Btext"
		err := PrintableText()(&str)
		if err == nil {
			t.Fatal("expected validation error for zero-width space")
		}
		if !strings.Contains(err.Message, "zero-width") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - zero-width non-joiner", func(t *testing.T) {
		str := "test\u200Ctext"
		err := PrintableText()(&str)
		if err == nil {
			t.Error("expected validation error for zero-width non-joiner")
		}
	})

	t.Run("invalid - zero-width joiner", func(t *testing.T) {
		str := "test\u200Dtext"
		err := PrintableText()(&str)
		if err == nil {
			t.Error("expected validation error for zero-width joiner")
		}
	})

	t.Run("invalid - BOM character", func(t *testing.T) {
		str := "\uFEFFtest"
		err := PrintableText()(&str)
		if err == nil {
			t.Error("expected validation error for BOM character")
		}
	})

	t.Run("invalid - null byte", func(t *testing.T) {
		str := "test\x00text"
		err := PrintableText()(&str)
		if err == nil {
			t.Fatal("expected validation error for null byte")
		}
		if !strings.Contains(err.Message, "control character") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - tab character", func(t *testing.T) {
		str := "test\ttext"
		err := PrintableText()(&str)
		if err == nil {
			t.Error("expected validation error for tab character")
		}
	})

	t.Run("valid - newline character", func(t *testing.T) {
		str := "test\ntext"
		err := PrintableText()(&str)
		if err != nil {
			t.Errorf("expected no error for newline character, got: %v", err)
		}
	})

	t.Run("valid - carriage return", func(t *testing.T) {
		str := "test\rtext"
		err := PrintableText()(&str)
		if err != nil {
			t.Errorf("expected no error for carriage return, got: %v", err)
		}
	})

	t.Run("valid - multiple newlines", func(t *testing.T) {
		str := "hello foo\nbar\n\njd"
		err := PrintableText()(&str)
		if err != nil {
			t.Errorf("expected no error for multiple newlines, got: %v", err)
		}
	})

	t.Run("invalid - soft hyphen", func(t *testing.T) {
		str := "test\u00ADtext"
		err := PrintableText()(&str)
		if err == nil {
			t.Fatal("expected validation error for soft hyphen")
		}
		if !strings.Contains(err.Message, "invisible formatting") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - word joiner", func(t *testing.T) {
		str := "test\u2060text"
		err := PrintableText()(&str)
		if err == nil {
			t.Error("expected validation error for word joiner")
		}
	})

	t.Run("invalid - private use area character", func(t *testing.T) {
		str := "test\uE000text"
		err := PrintableText()(&str)
		if err == nil {
			t.Fatal("expected validation error for private use area")
		}
		if !strings.Contains(err.Message, "private use") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - replacement character", func(t *testing.T) {
		str := "test\uFFFDtext"
		err := PrintableText()(&str)
		if err == nil {
			t.Fatal("expected validation error for replacement character")
		}
		if !strings.Contains(err.Message, "replacement character") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - DEL control character", func(t *testing.T) {
		str := "test\x7Ftext"
		err := PrintableText()(&str)
		if err == nil {
			t.Error("expected validation error for DEL control character")
		}
	})

	t.Run("invalid - C1 control character", func(t *testing.T) {
		str := "test\u0080text"
		err := PrintableText()(&str)
		if err == nil {
			t.Error("expected validation error for C1 control character")
		}
	})

	t.Run("invalid - LTR mark", func(t *testing.T) {
		str := "test\u200Etext"
		err := PrintableText()(&str)
		if err == nil {
			t.Error("expected validation error for LTR mark")
		}
	})

	t.Run("invalid - RTL mark", func(t *testing.T) {
		str := "test\u200Ftext"
		err := PrintableText()(&str)
		if err == nil {
			t.Error("expected validation error for RTL mark")
		}
	})

	t.Run("valid with pointer", func(t *testing.T) {
		str := "Valid Name"
		err := PrintableText()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		str := ""
		err := PrintableText()(&str)
		if err != nil {
			t.Errorf("expected no error for empty string, got: %v", err)
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var str *string
		err := PrintableText()(str)
		if err != nil {
			t.Errorf("expected no error for nil, got: %v", err)
		}
	})

	t.Run("not a string", func(t *testing.T) {
		num := 123
		err := PrintableText()(&num)
		if err == nil {
			t.Fatal("expected validation error for non-string")
		}
		if !strings.Contains(err.Message, "must be a string") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("combined with other validators", func(t *testing.T) {
		v := New()
		title := "Product Title 2024"
		v.Check(&title, "title", Required(), PrintableText(), MinLen(3), MaxLen(100))

		if v.HasErrors() {
			t.Errorf("expected no errors, got: %v", v.Errors())
		}
	})

	t.Run("position reported correctly", func(t *testing.T) {
		str := "abc\x00def"
		err := PrintableText()(&str)
		if err == nil {
			t.Fatal("expected validation error")
		}
		if !strings.Contains(err.Message, "position 3") {
			t.Errorf("expected position 3 in error message, got: %s", err.Message)
		}
	})

	t.Run("UTF-8 position counting", func(t *testing.T) {
		// Test that position is counted correctly with UTF-8 characters
		// The range loop in Go iterates by runes, so position will be rune index
		str := "abc\x00"
		err := PrintableText()(&str)
		if err == nil {
			t.Fatal("expected validation error")
		}
		// The null byte is at rune position 3 (after 'a', 'b', 'c')
		if !strings.Contains(err.Message, "position 3") {
			t.Errorf("expected position 3 in error message, got: %s", err.Message)
		}
	})
}

func TestSafeText(t *testing.T) {
	t.Run("valid text", func(t *testing.T) {
		str := "Product Name 2024"
		err := SafeText(100)(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid UTF-8 text", func(t *testing.T) {
		str := "José García"
		err := SafeText(50)(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid text with emoji", func(t *testing.T) {
		str := "Hello World 🌍"
		err := SafeText(50)(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid text with apostrophe and hyphen", func(t *testing.T) {
		str := "O'Brien-Smith"
		err := SafeText(50)(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid - empty string", func(t *testing.T) {
		str := ""
		err := SafeText(100)(&str)
		if err == nil {
			t.Fatal("expected validation error for empty string")
		}
		if !strings.Contains(err.Message, "empty") && !strings.Contains(err.Message, "required") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - exceeds max length", func(t *testing.T) {
		str := "This is a very long string that exceeds the maximum length"
		err := SafeText(10)(&str)
		if err == nil {
			t.Fatal("expected validation error for exceeding max length")
		}
		if !strings.Contains(err.Message, "at most") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - contains HTML tags", func(t *testing.T) {
		str := "Hello <b>World</b>"
		err := SafeText(100)(&str)
		if err == nil {
			t.Fatal("expected validation error for HTML tags")
		}
		if !strings.Contains(err.Message, "HTML tags") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - contains script tag", func(t *testing.T) {
		str := "<script>alert('xss')</script>"
		err := SafeText(100)(&str)
		if err == nil {
			t.Error("expected validation error for script tag")
		}
	})

	t.Run("invalid - contains angle brackets", func(t *testing.T) {
		str := "5 < 10"
		err := SafeText(100)(&str)
		if err == nil {
			t.Fatal("expected validation error for angle brackets")
		}
		if !strings.Contains(err.Message, "angle brackets") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - contains null byte", func(t *testing.T) {
		str := "test\x00text"
		err := SafeText(100)(&str)
		if err == nil {
			t.Fatal("expected validation error for null byte")
		}
		if !strings.Contains(err.Message, "control character") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - contains tab character", func(t *testing.T) {
		str := "test\ttext"
		err := SafeText(100)(&str)
		if err == nil {
			t.Error("expected validation error for tab character")
		}
	})

	t.Run("valid - contains newline", func(t *testing.T) {
		str := "test\ntext"
		err := SafeText(100)(&str)
		if err != nil {
			t.Errorf("expected no error for newline, got: %v", err)
		}
	})

	t.Run("valid - contains multiple newlines", func(t *testing.T) {
		str := "hello foo\nbar\n\njd"
		err := SafeText(100)(&str)
		if err != nil {
			t.Errorf("expected no error for multiple newlines, got: %v", err)
		}
	})

	t.Run("invalid - contains zero-width space", func(t *testing.T) {
		str := "test\u200Btext"
		err := SafeText(100)(&str)
		if err == nil {
			t.Fatal("expected validation error for zero-width space")
		}
		if !strings.Contains(err.Message, "zero-width") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - contains RLO character", func(t *testing.T) {
		str := "test\u202Eexe.txt"
		err := SafeText(100)(&str)
		if err == nil {
			t.Fatal("expected validation error for RLO character")
		}
		if !strings.Contains(err.Message, "bidirectional override") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - contains private use area character", func(t *testing.T) {
		str := "test\uE000text"
		err := SafeText(100)(&str)
		if err == nil {
			t.Fatal("expected validation error for private use area")
		}
		if !strings.Contains(err.Message, "private use") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var str *string
		err := SafeText(100)(str)
		if err != nil {
			t.Errorf("expected no error for nil pointer, got: %v", err)
		}
	})

	t.Run("not a string", func(t *testing.T) {
		num := 123
		err := SafeText(100)(&num)
		if err == nil {
			t.Fatal("expected validation error for non-string")
		}
		if !strings.Contains(err.Message, "must be a string") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("combined with validator struct", func(t *testing.T) {
		v := New()
		title := "Product Title 2024"
		v.Check(&title, "title", SafeText(100))

		if v.HasErrors() {
			t.Errorf("expected no errors, got: %v", v.Errors())
		}
	})

	t.Run("combined with validator struct - invalid", func(t *testing.T) {
		v := New()
		malicious := "<script>alert('xss')</script>"
		v.Check(&malicious, "content", SafeText(100))

		if !v.HasErrors() {
			t.Error("expected validation errors")
		}

		errors := v.Errors()
		found := false
		for _, err := range errors {
			if strings.Contains(err.Message, "HTML tags") || strings.Contains(err.Message, "angle brackets") {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected error about HTML tags or angle brackets")
		}
	})

	t.Run("edge case - exactly at max length", func(t *testing.T) {
		str := "12345"
		err := SafeText(5)(&str)
		if err != nil {
			t.Errorf("expected no error for string at max length, got: %v", err)
		}
	})

	t.Run("edge case - one character over max length", func(t *testing.T) {
		str := "123456"
		err := SafeText(5)(&str)
		if err == nil {
			t.Error("expected validation error for string over max length")
		}
	})
}

func TestNoNewLine(t *testing.T) {
	t.Run("valid text without newlines", func(t *testing.T) {
		str := "Product Name 2024"
		err := NoNewLine()(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid - contains newline", func(t *testing.T) {
		str := "Line 1\nLine 2"
		err := NoNewLine()(&str)
		if err == nil {
			t.Fatal("expected validation error for newline")
		}
		if !strings.Contains(err.Message, "newline") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - contains carriage return", func(t *testing.T) {
		str := "Line 1\rLine 2"
		err := NoNewLine()(&str)
		if err == nil {
			t.Fatal("expected validation error for carriage return")
		}
		if !strings.Contains(err.Message, "carriage return") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - contains both newline and carriage return", func(t *testing.T) {
		str := "Line 1\n\rLine 3"
		err := NoNewLine()(&str)
		if err == nil {
			t.Error("expected validation error for newline or carriage return")
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var str *string
		err := NoNewLine()(str)
		if err != nil {
			t.Errorf("expected no error for nil pointer, got: %v", err)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		str := ""
		err := NoNewLine()(&str)
		if err != nil {
			t.Errorf("expected no error for empty string, got: %v", err)
		}
	})
}

func TestSafeTextNoNewLine(t *testing.T) {
	t.Run("valid text", func(t *testing.T) {
		str := "Product Name 2024"
		err := SafeTextNoNewLine(100)(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid UTF-8 text", func(t *testing.T) {
		str := "José García"
		err := SafeTextNoNewLine(50)(&str)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid - contains newline", func(t *testing.T) {
		str := "Line 1\nLine 2"
		err := SafeTextNoNewLine(100)(&str)
		if err == nil {
			t.Fatal("expected validation error for newline")
		}
		if !strings.Contains(err.Message, "newline") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - contains carriage return", func(t *testing.T) {
		str := "Line 1\rLine 2"
		err := SafeTextNoNewLine(100)(&str)
		if err == nil {
			t.Fatal("expected validation error for carriage return")
		}
		if !strings.Contains(err.Message, "carriage return") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - empty string", func(t *testing.T) {
		str := ""
		err := SafeTextNoNewLine(100)(&str)
		if err == nil {
			t.Fatal("expected validation error for empty string")
		}
		if !strings.Contains(err.Message, "empty") && !strings.Contains(err.Message, "required") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - exceeds max length", func(t *testing.T) {
		str := "This is a very long string that exceeds the maximum length"
		err := SafeTextNoNewLine(10)(&str)
		if err == nil {
			t.Fatal("expected validation error for exceeding max length")
		}
		if !strings.Contains(err.Message, "at most") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - contains HTML tags", func(t *testing.T) {
		str := "Hello <b>World</b>"
		err := SafeTextNoNewLine(100)(&str)
		if err == nil {
			t.Fatal("expected validation error for HTML tags")
		}
		if !strings.Contains(err.Message, "HTML tags") {
			t.Errorf("unexpected error message: %s", err.Message)
		}
	})

	t.Run("invalid - contains tab character", func(t *testing.T) {
		str := "test\ttext"
		err := SafeTextNoNewLine(100)(&str)
		if err == nil {
			t.Error("expected validation error for tab character")
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var str *string
		err := SafeTextNoNewLine(100)(str)
		if err != nil {
			t.Errorf("expected no error for nil pointer, got: %v", err)
		}
	})

	t.Run("edge case - exactly at max length", func(t *testing.T) {
		str := "12345"
		err := SafeTextNoNewLine(5)(&str)
		if err != nil {
			t.Errorf("expected no error for string at max length, got: %v", err)
		}
	})
}

