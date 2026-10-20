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

package filevalidation

import (
	"mime"
	"path/filepath"
	"slices"
	"strings"
)

// ContentTypeForFilename returns the canonical MIME type for filename's
// extension, or "" when that extension is not a supported file type.
func ContentTypeForFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return ""
	}

	for _, fileType := range FileTypes {
		if slices.Contains(fileType.Extensions, ext) {
			return fileType.MimeType
		}
	}

	return ""
}

// ResolveContentType returns the media type to validate and store.
// Charset and other parameters are dropped. An empty type or
// application/octet-stream is replaced with the type implied by filename
// when that extension is supported.
func ResolveContentType(contentType, filename string) string {
	mediaType := mediaTypeOf(contentType)
	if mediaType == "" || mediaType == "application/octet-stream" {
		if detected := ContentTypeForFilename(filename); detected != "" {
			return detected
		}
	}

	if mediaType == "" {
		return "application/octet-stream"
	}

	return mediaType
}

func mediaTypeOf(contentType string) string {
	contentType = strings.TrimSpace(contentType)
	if contentType == "" {
		return ""
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return contentType
	}

	return mediaType
}
