// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
