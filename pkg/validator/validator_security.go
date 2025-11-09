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
	"fmt"
	"regexp"
	"strings"
)

var (
	htmlTagRegex = regexp.MustCompile(`<[^>]*>`)
)

// NoHTML validates that a string does not contain HTML tags or angle brackets.
// It rejects:
// - HTML tags (e.g., <script>, <b>, <div>, etc.)
// - Angle brackets (< and >) even when not part of complete tags
//
// This helps prevent XSS attacks and ensures user input doesn't contain HTML markup.
// Combine with PrintableText() for comprehensive text field validation.
func NoHTML() ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		str, ok := actualValue.(string)
		if !ok {
			return newValidationError(ErrorCodeInvalidFormat, "value must be a string")
		}

		if str == "" {
			return nil
		}

		// Check for HTML tags first (more specific error message)
		if htmlTagRegex.MatchString(str) {
			return newValidationError(ErrorCodeInvalidFormat, "must not contain HTML tags")
		}

		// Check for angle brackets (even without complete tags)
		if strings.ContainsAny(str, "<>") {
			return newValidationError(ErrorCodeInvalidFormat, "must not contain angle brackets")
		}

		return nil
	}
}

// PrintableText validates that a string contains only printable UTF-8 characters.
// It rejects:
// - Control characters (0x00-0x1F and 0x7F-0x9F, including null bytes and tabs, but allows newlines and carriage returns)
// - Unicode direction override characters (RLO, LRO, PDF, etc.)
// - Zero-width characters (ZWSP, ZWNJ, ZWJ, etc.)
// - Other invisible or formatting characters
// - Private use area characters
// - Replacement characters
//
// This validator does NOT check for HTML tags - use NoHTML() for that.
// This validator allows line breaks (newline and carriage return) for multi-line text fields.
// Use NoNewLine() or SafeTextNoNewLine() for single-line fields that should reject line breaks.
func PrintableText() ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		str, ok := actualValue.(string)
		if !ok {
			return newValidationError(ErrorCodeInvalidFormat, "value must be a string")
		}

		if str == "" {
			return nil
		}

		// Check each rune for invisible or problematic characters
		for i, r := range str {
			// Allow normal space
			if r == ' ' {
				continue
			}

			// Allow newline (0x0A) and carriage return (0x0D) for multi-line text
			if r == '\n' || r == '\r' {
				continue
			}

			// Reject control characters (0x00-0x1F and 0x7F-0x9F), except newline and carriage return
			if r < 0x20 || (r >= 0x7F && r < 0xA0) {
				return newValidationError(ErrorCodeInvalidFormat, fmt.Sprintf("contains invalid control character at position %d", i))
			}

			// Reject Unicode direction override and formatting characters
			// U+200E LEFT-TO-RIGHT MARK (LRM)
			// U+200F RIGHT-TO-LEFT MARK (RLM)
			// U+202A LEFT-TO-RIGHT EMBEDDING (LRE)
			// U+202B RIGHT-TO-LEFT EMBEDDING (RLE)
			// U+202C POP DIRECTIONAL FORMATTING (PDF)
			// U+202D LEFT-TO-RIGHT OVERRIDE (LRO)
			// U+202E RIGHT-TO-LEFT OVERRIDE (RLO)
			// U+2066 LEFT-TO-RIGHT ISOLATE (LRI)
			// U+2067 RIGHT-TO-LEFT ISOLATE (RLI)
			// U+2068 FIRST STRONG ISOLATE (FSI)
			// U+2069 POP DIRECTIONAL ISOLATE (PDI)
			if r >= 0x200E && r <= 0x200F || r >= 0x202A && r <= 0x202E || r >= 0x2066 && r <= 0x2069 {
				return newValidationError(ErrorCodeInvalidFormat, fmt.Sprintf("contains bidirectional override character at position %d", i))
			}

			// Reject zero-width characters
			// U+200B ZERO WIDTH SPACE (ZWSP)
			// U+200C ZERO WIDTH NON-JOINER (ZWNJ)
			// U+200D ZERO WIDTH JOINER (ZWJ)
			// U+FEFF ZERO WIDTH NO-BREAK SPACE (BOM)
			if r == 0x200B || r == 0x200C || r == 0x200D || r == 0xFEFF {
				return newValidationError(ErrorCodeInvalidFormat, fmt.Sprintf("contains zero-width character at position %d", i))
			}

			// Reject other format characters (Cf category)
			// U+00AD SOFT HYPHEN
			// U+2060 WORD JOINER
			// U+180E MONGOLIAN VOWEL SEPARATOR (deprecated but still problematic)
			if r == 0x00AD || r == 0x2060 || r == 0x180E {
				return newValidationError(ErrorCodeInvalidFormat, fmt.Sprintf("contains invisible formatting character at position %d", i))
			}

			// Reject private use area characters (often used for exploits)
			// U+E000-U+F8FF Private Use Area
			// U+F0000-U+FFFFD Supplementary Private Use Area-A
			// U+100000-U+10FFFD Supplementary Private Use Area-B
			if (r >= 0xE000 && r <= 0xF8FF) || (r >= 0xF0000 && r <= 0xFFFFD) || (r >= 0x100000 && r <= 0x10FFFD) {
				return newValidationError(ErrorCodeInvalidFormat, fmt.Sprintf("contains private use character at position %d", i))
			}

			// Reject replacement character (often indicates encoding issues)
			if r == 0xFFFD {
				return newValidationError(ErrorCodeInvalidFormat, fmt.Sprintf("contains replacement character at position %d", i))
			}
		}

		return nil
	}
}

// NoNewLine validates that a string does not contain newline or carriage return characters.
// It rejects:
// - Newline characters (\n, 0x0A)
// - Carriage return characters (\r, 0x0D)
//
// This is useful for validating single-line fields like names and titles where line breaks
// should not be allowed.
func NoNewLine() ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		str, ok := actualValue.(string)
		if !ok {
			return newValidationError(ErrorCodeInvalidFormat, "value must be a string")
		}

		if str == "" {
			return nil
		}

		for i, r := range str {
			if r == '\n' {
				return newValidationError(ErrorCodeInvalidFormat, fmt.Sprintf("contains newline character at position %d", i))
			}
			if r == '\r' {
				return newValidationError(ErrorCodeInvalidFormat, fmt.Sprintf("contains carriage return character at position %d", i))
			}
		}

		return nil
	}
}

// SafeText validates that a string is non-empty, bounded, and contains only safe content.
// It combines NotEmpty, MaxLen, NoHTML, and PrintableText validators.
// This allows newlines and carriage returns for multi-line text fields.
// Use SafeTextNoNewLine for single-line field validation that should reject line breaks.
func SafeText(maxLen int) ValidatorFunc {
	validators := []ValidatorFunc{
		NotEmpty(),
		MaxLen(maxLen),
		NoHTML(),
		PrintableText(),
	}

	return func(value any) *ValidationError {
		for _, validator := range validators {
			if err := validator(value); err != nil {
				return err
			}
		}
		return nil
	}
}

// SafeTextNoNewLine validates that a string is non-empty, bounded, and contains only safe content
// without newlines or carriage returns. It combines NotEmpty, MaxLen, NoHTML, PrintableText, and NoNewLine validators.
// This is ideal for validating single-line fields like names, titles, and display names.
func SafeTextNoNewLine(maxLen int) ValidatorFunc {
	validators := []ValidatorFunc{
		NotEmpty(),
		MaxLen(maxLen),
		NoHTML(),
		PrintableText(),
		NoNewLine(),
	}

	return func(value any) *ValidationError {
		for _, validator := range validators {
			if err := validator(value); err != nil {
				return err
			}
		}
		return nil
	}
}
