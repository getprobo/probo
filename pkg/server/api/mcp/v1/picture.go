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

package mcp_v1

import (
	"encoding/base64"
	"strings"

	"go.probo.inc/probo/pkg/task"
	"go.probo.inc/probo/pkg/validator"
)

// decodePictureContent rejects payloads larger than one byte past the picture
// limit before decoding, so an oversized base64 string cannot allocate the
// decoded image.
func decodePictureContent(encoded string) ([]byte, error) {
	encoded = strings.TrimSpace(encoded)
	maxLen := base64.StdEncoding.EncodedLen(task.MaxPictureBytes + 1)

	if len(encoded) > maxLen {
		v := validator.New()
		v.Check(len(encoded), "content_base64", validator.Max(maxLen))

		return nil, v.Error()
	}

	body, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		v := validator.New()
		v.Check(
			encoded,
			"content_base64",
			func(any) *validator.ValidationError {
				return &validator.ValidationError{
					Code:    validator.ErrorCodeInvalidFormat,
					Message: "must be standard base64",
				}
			},
		)

		return nil, v.Error()
	}

	return body, nil
}
