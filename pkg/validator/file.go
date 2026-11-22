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
	"fmt"
	"path/filepath"
	"slices"
	"strings"

)

const (
	// DefaultMaxFileSize defines the default maximum allowed file size (50MB)
	DefaultMaxFileSize = 50 * 1024 * 1024
)

// File type categories
const (
	FileCategoryDocument     = "document"
	FileCategorySpreadsheet  = "spreadsheet"
	FileCategoryPresentation = "presentation"
	FileCategoryText         = "text"
	FileCategoryImage        = "image"
	FileCategoryData         = "data"
	FileCategoryVideo        = "video"
)

// FileType defines a supported file type with its MIME type and extensions
type FileType struct {
	MimeType   string
	Extensions []string
	Category   string
}

// FileTypes is a list of all supported file types
var FileTypes = []FileType{
	// Document types
	{MimeType: "application/pdf", Extensions: []string{".pdf"}, Category: FileCategoryDocument},
	{MimeType: "application/msword", Extensions: []string{".doc"}, Category: FileCategoryDocument},
	{MimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document", Extensions: []string{".docx"}, Category: FileCategoryDocument},
	{MimeType: "application/vnd.oasis.opendocument.text", Extensions: []string{".odt"}, Category: FileCategoryDocument},

	// Spreadsheet types
	{MimeType: "application/vnd.ms-excel", Extensions: []string{".xls"}, Category: FileCategorySpreadsheet},
	{MimeType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", Extensions: []string{".xlsx"}, Category: FileCategorySpreadsheet},
	{MimeType: "application/vnd.oasis.opendocument.spreadsheet", Extensions: []string{".ods"}, Category: FileCategorySpreadsheet},

	// Presentation types
	{MimeType: "application/vnd.ms-powerpoint", Extensions: []string{".ppt"}, Category: FileCategoryPresentation},
	{MimeType: "application/vnd.openxmlformats-officedocument.presentationml.presentation", Extensions: []string{".pptx"}, Category: FileCategoryPresentation},
	{MimeType: "application/vnd.oasis.opendocument.presentation", Extensions: []string{".odp"}, Category: FileCategoryPresentation},

	// Text types
	{MimeType: "text/markdown", Extensions: []string{".md"}, Category: FileCategoryText},
	{MimeType: "text/plain", Extensions: []string{".txt", ".log"}, Category: FileCategoryText},
	{MimeType: "text/x-log", Extensions: []string{".log"}, Category: FileCategoryText},
	{MimeType: "text/uri-list", Extensions: []string{".uri"}, Category: FileCategoryText},
	{MimeType: "text/uri-list; charset=utf-8", Extensions: []string{".uri"}, Category: FileCategoryText},

	// Image types
	{MimeType: "image/jpeg", Extensions: []string{".jpg", ".jpeg"}, Category: FileCategoryImage},
	{MimeType: "image/png", Extensions: []string{".png"}, Category: FileCategoryImage},
	{MimeType: "image/svg+xml", Extensions: []string{".svg"}, Category: FileCategoryImage},
	{MimeType: "image/webp", Extensions: []string{".webp"}, Category: FileCategoryImage},

	// Data types
	{MimeType: "application/yaml", Extensions: []string{".yaml", ".yml"}, Category: FileCategoryData},
	{MimeType: "application/json", Extensions: []string{".json"}, Category: FileCategoryData},
	{MimeType: "text/yaml", Extensions: []string{".yaml", ".yml"}, Category: FileCategoryData},
	{MimeType: "text/json", Extensions: []string{".json"}, Category: FileCategoryData},
	{MimeType: "text/csv", Extensions: []string{".csv"}, Category: FileCategoryData},
	{MimeType: "application/csv", Extensions: []string{".csv"}, Category: FileCategoryData},
	// Windows browsers report .csv as the legacy Excel type.
	{MimeType: "application/vnd.ms-excel", Extensions: []string{".csv"}, Category: FileCategoryData},

	// Video types
	{MimeType: "video/mp4", Extensions: []string{".mp4"}, Category: FileCategoryVideo},
	{MimeType: "video/mpeg", Extensions: []string{".mpeg", ".mpg"}, Category: FileCategoryVideo},
	{MimeType: "video/quicktime", Extensions: []string{".mov"}, Category: FileCategoryVideo},
	{MimeType: "video/x-msvideo", Extensions: []string{".avi"}, Category: FileCategoryVideo},
	{MimeType: "video/webm", Extensions: []string{".webm"}, Category: FileCategoryVideo},
}

// FileError is a file rejected by Validate. Callers should report it to the
// client; it is not an internal failure. Code is the validation cause.
type FileError struct {
	message string
	code    ErrorCode
}

func (e *FileError) Error() string {
	return e.message
}

// Code returns the validation cause for this rejection.
func (e *FileError) Code() ErrorCode {
	return e.code
}

func fileValidationError(code ErrorCode, format string, args ...any) error {
	return &FileError{
		message: fmt.Sprintf(format, args...),
		code:    code,
	}
}

// FileValidator is a configurable file validator
type FileValidator struct {
	// MaxFileSize specifies the maximum file size in bytes
	MaxFileSize int64

	// AllowedMimeTypes is a map of allowed MIME types
	AllowedMimeTypes map[string]bool

	// AllowedExtensions maps file extensions to their expected MIME types
	AllowedExtensions map[string][]string

	// Categories is a list of categories to include
	Categories []string
}

type FileValidatorOption func(v *FileValidator) *FileValidator

func addFileType(v *FileValidator, fileType FileType) {
	v.AllowedMimeTypes[fileType.MimeType] = true

	for _, ext := range fileType.Extensions {
		if v.AllowedExtensions[ext] == nil {
			v.AllowedExtensions[ext] = []string{}
		}

		v.AllowedExtensions[ext] = append(v.AllowedExtensions[ext], fileType.MimeType)
	}
}

func WithFileCategories(categories ...string) FileValidatorOption {
	return func(v *FileValidator) *FileValidator {
		v.Categories = categories

		categoryMap := make(map[string]bool)
		for _, category := range categories {
			categoryMap[category] = true
		}

		for _, fileType := range FileTypes {
			if categoryMap[fileType.Category] {
				addFileType(v, fileType)
			}
		}

		return v
	}
}

func WithFileMimeTypes(mimeTypes ...string) FileValidatorOption {
	return func(v *FileValidator) *FileValidator {
		wanted := make(map[string]bool, len(mimeTypes))
		for _, mimeType := range mimeTypes {
			wanted[mimeType] = true
		}

		found := make(map[string]bool, len(wanted))

		for _, fileType := range FileTypes {
			if !wanted[fileType.MimeType] {
				continue
			}

			found[fileType.MimeType] = true
			addFileType(v, fileType)
		}

		unknown := make([]string, 0, len(wanted))
		for mimeType := range wanted {
			if !found[mimeType] {
				unknown = append(unknown, mimeType)
			}
		}

		if len(unknown) > 0 {
			slices.Sort(unknown)
			panic("validator: unknown MIME types: " + strings.Join(unknown, ", "))
		}

		return v
	}
}

func WithMaxFileSize(maxFileSize int64) FileValidatorOption {
	return func(v *FileValidator) *FileValidator {
		v.MaxFileSize = maxFileSize

		return v
	}
}

func NewFileValidator(opts ...FileValidatorOption) *FileValidator {
	v := &FileValidator{
		MaxFileSize:       DefaultMaxFileSize,
		AllowedMimeTypes:  make(map[string]bool),
		AllowedExtensions: make(map[string][]string),
	}

	for _, opt := range opts {
		v = opt(v)
	}

	return v
}

// NewAttachmentValidator returns the validator used for attachment uploads.
func NewAttachmentValidator() *FileValidator {
	return NewFileValidator(
		WithFileCategories(
			FileCategoryDocument,
			FileCategorySpreadsheet,
			FileCategoryPresentation,
			FileCategoryText,
			FileCategoryImage,
			FileCategoryData,
		),
	)
}

// ContentTypeForAttachment returns the MIME type accepted for filename.
func ContentTypeForAttachment(filename string) (string, bool) {
	return NewAttachmentValidator().ContentTypeForFilename(filename)
}

// WithMaxFileSize sets the maximum file size and returns the validator
func (v *FileValidator) WithMaxFileSize(maxSize int64) *FileValidator {
	v.MaxFileSize = maxSize
	return v
}

// ContentTypeForFilename returns the first allowed MIME type for filename's extension.
func (v *FileValidator) ContentTypeForFilename(filename string) (string, bool) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return "", false
	}

	contentTypes := v.AllowedExtensions[ext]
	if len(contentTypes) == 0 {
		return "", false
	}

	return contentTypes[0], true
}

// Validate validates that the file meets the configured requirements
func (v *FileValidator) Validate(filename string, contentType string, size int64) error {
	if size == 0 {
		return fileValidationError(ErrorCodeOutOfRange, "file can't be empty")
	}

	if size > v.MaxFileSize {
		return fileValidationError(
			ErrorCodeOutOfRange,
			"file size exceeds maximum allowed size of %d bytes",
			v.MaxFileSize,
		)
	}

	if !v.AllowedMimeTypes[contentType] {
		return fileValidationError(ErrorCodeInvalidFormat, "content type %q is not allowed", contentType)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return fileValidationError(ErrorCodeInvalidFormat, "file has no extension")
	}

	allowedTypes, extAllowed := v.AllowedExtensions[ext]
	if !extAllowed {
		return fileValidationError(ErrorCodeInvalidFormat, "file extension %q is not allowed", ext)
	}

	validType := slices.Contains(allowedTypes, contentType)

	if !validType {
		return fileValidationError(
			ErrorCodeInvalidFormat,
			"content type %q does not match extension %q",
			contentType,
			ext,
		)
	}

	return nil
}
