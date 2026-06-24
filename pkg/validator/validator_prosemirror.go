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

package validator

import (
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/prosemirror"
)

// ProseMirrorDocumentContent requires non-empty string values to be valid
// ProseMirror/Tiptap JSON with root type "doc". Empty and whitespace-only
// strings are allowed.
func ProseMirrorDocumentContent() ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		s, ok := actualValue.(string)
		if !ok {
			return newValidationError(ErrorCodeInvalidFormat, "value must be a string")
		}

		if strings.TrimSpace(s) == "" {
			return nil
		}

		if err := prosemirror.ValidateDocumentContentJSON(s); err != nil {
			return newValidationError(ErrorCodeInvalidFormat, err.Error())
		}

		return nil
	}
}

// ProseMirrorDocumentMaxTextLength validates that the total text content
// within a ProseMirror/Tiptap JSON document does not exceed maxLength bytes.
// Only user-visible text is counted; structural markup is excluded.
// Nil, empty, and whitespace-only values pass. Invalid JSON is skipped
// (let ProseMirrorDocumentContent handle format errors).
func ProseMirrorDocumentMaxTextLength(maxLength int) ValidatorFunc {
	return func(value any) *ValidationError {
		actualValue, isNil := dereferenceValue(value)
		if isNil {
			return nil
		}

		s, ok := actualValue.(string)
		if !ok {
			return newValidationError(ErrorCodeInvalidFormat, "value must be a string")
		}

		if strings.TrimSpace(s) == "" {
			return nil
		}

		n, err := prosemirror.Parse(s)
		if err != nil {
			return nil
		}

		if n.TextLength() > maxLength {
			return newValidationError(
				ErrorCodeTooLong,
				fmt.Sprintf("text content must be at most %d characters", maxLength),
			)
		}

		return nil
	}
}
