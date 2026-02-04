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

package filevalidation

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
	CategoryDocument     = "document"
	CategorySpreadsheet  = "spreadsheet"
	CategoryPresentation = "presentation"
	CategoryText         = "text"
	CategoryImage        = "image"
	CategoryData         = "data"
	CategoryVideo        = "video"
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
	{MimeType: "application/pdf", Extensions: []string{".pdf"}, Category: CategoryDocument},
	{MimeType: "application/msword", Extensions: []string{".doc"}, Category: CategoryDocument},
	{MimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document", Extensions: []string{".docx"}, Category: CategoryDocument},
	{MimeType: "application/vnd.oasis.opendocument.text", Extensions: []string{".odt"}, Category: CategoryDocument},

	// Spreadsheet types
	{MimeType: "application/vnd.ms-excel", Extensions: []string{".xls"}, Category: CategorySpreadsheet},
	{MimeType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", Extensions: []string{".xlsx"}, Category: CategorySpreadsheet},
	{MimeType: "application/vnd.oasis.opendocument.spreadsheet", Extensions: []string{".ods"}, Category: CategorySpreadsheet},

	// Presentation types
	{MimeType: "application/vnd.ms-powerpoint", Extensions: []string{".ppt"}, Category: CategoryPresentation},
	{MimeType: "application/vnd.openxmlformats-officedocument.presentationml.presentation", Extensions: []string{".pptx"}, Category: CategoryPresentation},
	{MimeType: "application/vnd.oasis.opendocument.presentation", Extensions: []string{".odp"}, Category: CategoryPresentation},

	// Text types
	{MimeType: "text/plain", Extensions: []string{".txt"}, Category: CategoryText},
	{MimeType: "text/x-log", Extensions: []string{".log"}, Category: CategoryText},
	{MimeType: "text/uri-list", Extensions: []string{".uri"}, Category: CategoryText},
	{MimeType: "text/uri-list; charset=utf-8", Extensions: []string{".uri"}, Category: CategoryText},

	// Image types
	{MimeType: "image/jpeg", Extensions: []string{".jpg", ".jpeg"}, Category: CategoryImage},
	{MimeType: "image/png", Extensions: []string{".png"}, Category: CategoryImage},
	{MimeType: "image/gif", Extensions: []string{".gif"}, Category: CategoryImage},
	{MimeType: "image/svg+xml", Extensions: []string{".svg"}, Category: CategoryImage},
	{MimeType: "image/webp", Extensions: []string{".webp"}, Category: CategoryImage},

	// Data types
	{MimeType: "application/yaml", Extensions: []string{".yaml", ".yml"}, Category: CategoryData},
	{MimeType: "application/json", Extensions: []string{".json"}, Category: CategoryData},
	{MimeType: "text/yaml", Extensions: []string{".yaml", ".yml"}, Category: CategoryData},
	{MimeType: "text/json", Extensions: []string{".json"}, Category: CategoryData},
	{MimeType: "text/csv", Extensions: []string{".csv"}, Category: CategoryData},
	{MimeType: "application/csv", Extensions: []string{".csv"}, Category: CategoryData},

	// Video types
	{MimeType: "video/mp4", Extensions: []string{".mp4"}, Category: CategoryVideo},
	{MimeType: "video/mpeg", Extensions: []string{".mpeg", ".mpg"}, Category: CategoryVideo},
	{MimeType: "video/quicktime", Extensions: []string{".mov"}, Category: CategoryVideo},
	{MimeType: "video/x-msvideo", Extensions: []string{".avi"}, Category: CategoryVideo},
	{MimeType: "video/webm", Extensions: []string{".webm"}, Category: CategoryVideo},
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

type Option func(v *FileValidator) *FileValidator

func WithCategories(categories ...string) Option {
	return func(v *FileValidator) *FileValidator {
		v.Categories = categories

		categoryMap := make(map[string]bool)
		for _, category := range categories {
			categoryMap[category] = true
		}

		for _, fileType := range FileTypes {
			if categoryMap[fileType.Category] {
				v.AllowedMimeTypes[fileType.MimeType] = true
				for _, ext := range fileType.Extensions {
					if v.AllowedExtensions[ext] == nil {
						v.AllowedExtensions[ext] = []string{}
					}
					v.AllowedExtensions[ext] = append(v.AllowedExtensions[ext], fileType.MimeType)
				}
			}
		}

		return v
	}
}

func WithMaxFileSize(maxFileSize int64) Option {
	return func(v *FileValidator) *FileValidator {
		v.MaxFileSize = maxFileSize

		return v
	}
}

func NewValidator(opts ...Option) *FileValidator {
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

// WithMaxFileSize sets the maximum file size and returns the validator
func (v *FileValidator) WithMaxFileSize(maxSize int64) *FileValidator {
	v.MaxFileSize = maxSize
	return v
}

// Validate validates that the file meets the configured requirements
func (v *FileValidator) Validate(filename string, contentType string, size int64) error {
	if size == 0 {
		return fmt.Errorf("file can't be empty")
	}

	if size > v.MaxFileSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", v.MaxFileSize)
	}

	if !v.AllowedMimeTypes[contentType] {
		return fmt.Errorf("content type %q is not allowed", contentType)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return fmt.Errorf("file has no extension")
	}

	allowedTypes, extAllowed := v.AllowedExtensions[ext]
	if !extAllowed {
		return fmt.Errorf("file extension %q is not allowed", ext)
	}

	validType := slices.Contains(allowedTypes, contentType)

	if !validType {
		return fmt.Errorf("content type %q does not match extension %q", contentType, ext)
	}

	return nil
}
