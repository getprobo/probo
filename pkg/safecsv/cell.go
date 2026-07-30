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

package safecsv

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// SanitizeRecord returns a copy of record with spreadsheet-safe cell values.
func SanitizeRecord(record []string) []string {
	if len(record) == 0 {
		return record
	}

	out := make([]string, len(record))
	for i, field := range record {
		out[i] = SanitizeCell(field)
	}

	return out
}

// SanitizeCell prefixes values that spreadsheet tools may interpret as formulas.
// Leading spaces are ignored for formula detection only; tab and carriage return
// at the start are always escaped. The written cell keeps the original text with
// a leading single-quote when sanitization applies.
func SanitizeCell(value string) string {
	if value == "" {
		return value
	}

	if spreadsheetLeadingControl(value[0]) {
		return "'" + value
	}

	trimmed := strings.TrimLeftFunc(value, unicode.IsSpace)
	if trimmed == "" {
		return value
	}

	r, _ := utf8.DecodeRuneInString(trimmed)
	if formulaLeadingRune(r) {
		return "'" + value
	}

	return value
}

func spreadsheetLeadingControl(b byte) bool {
	switch b {
	case '\t', '\r':
		return true
	default:
		return false
	}
}

func formulaLeadingRune(r rune) bool {
	switch r {
	case '=', '+', '-', '@', '\\', '|', '%':
		return true
	default:
		return false
	}
}
