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
	"strings"
	"testing"
)

func TestNewValidator(t *testing.T) {
	tests := []struct {
		name           string
		categories     []string
		expectedMimes  []string
		expectedExts   []string
		unexpectedMime string
	}{
		{
			name:           "No categories (all types)",
			categories:     []string{},
			expectedMimes:  []string{},
			expectedExts:   []string{},
			unexpectedMime: "application/pdf",
		},
		{
			name:           "Only documents",
			categories:     []string{CategoryDocument},
			expectedMimes:  []string{"application/pdf", "application/msword"},
			expectedExts:   []string{".pdf", ".doc", ".docx"},
			unexpectedMime: "image/jpeg",
		},
		{
			name:           "Only images",
			categories:     []string{CategoryImage},
			expectedMimes:  []string{"image/jpeg", "image/png", "image/gif"},
			expectedExts:   []string{".jpg", ".png", ".gif"},
			unexpectedMime: "application/pdf",
		},
		{
			name:           "Multiple categories",
			categories:     []string{CategoryImage, CategoryDocument},
			expectedMimes:  []string{"image/jpeg", "application/pdf"},
			expectedExts:   []string{".jpg", ".pdf"},
			unexpectedMime: "video/mp4",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v := NewValidator(WithCategories(tc.categories...))

			// Test expected mime types are allowed
			for _, mime := range tc.expectedMimes {
				if !v.AllowedMimeTypes[mime] {
					t.Errorf("Expected mime type %q to be allowed, but it wasn't", mime)
				}
			}

			// Test expected extensions are allowed
			for _, ext := range tc.expectedExts {
				if _, ok := v.AllowedExtensions[ext]; !ok {
					t.Errorf("Expected extension %q to be allowed, but it wasn't", ext)
				}
			}

			// Test unexpected mime type is not allowed
			if v.AllowedMimeTypes[tc.unexpectedMime] {
				t.Errorf("Unexpected mime type %q should not be allowed, but it was", tc.unexpectedMime)
			}

			// Check categories were set correctly
			if len(tc.categories) != len(v.Categories) {
				t.Errorf("Expected %d categories, got %d", len(tc.categories), len(v.Categories))
			}

			// Default max size should be set
			if v.MaxFileSize != DefaultMaxFileSize {
				t.Errorf("Expected default max file size %d, got %d", DefaultMaxFileSize, v.MaxFileSize)
			}
		})
	}
}

func TestWithMaxFileSize(t *testing.T) {
	v := NewValidator()
	originalSize := v.MaxFileSize

	// Test changing the max file size
	newSize := int64(10 * 1024 * 1024) // 10MB
	v = v.WithMaxFileSize(newSize)

	if v.MaxFileSize != newSize {
		t.Errorf("Expected max file size to be %d after WithMaxFileSize, got %d", newSize, v.MaxFileSize)
	}

	if originalSize == v.MaxFileSize {
		t.Errorf("Max file size should have changed from original %d", originalSize)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		validator   *FileValidator
		filename    string
		contentType string
		fileSize    int64
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "Valid PDF file",
			validator:   NewValidator(WithCategories(CategoryDocument)),
			filename:    "test.pdf",
			contentType: "application/pdf",
			fileSize:    1024 * 1024, // 1MB
			shouldError: false,
		},
		{
			name:        "File too large",
			validator:   NewValidator(WithCategories(CategoryDocument)).WithMaxFileSize(1024 * 1024), // 1MB max
			filename:    "test.pdf",
			contentType: "application/pdf",
			fileSize:    2 * 1024 * 1024, // 2MB
			shouldError: true,
			errorMsg:    "file size exceeds maximum allowed size",
		},
		{
			name:        "Disallowed content type",
			validator:   NewValidator(WithCategories(CategoryDocument)),
			filename:    "test.jpg",
			contentType: "image/jpeg",
			fileSize:    1024 * 1024,
			shouldError: true,
			errorMsg:    "content type \"image/jpeg\" is not allowed",
		},
		{
			name:        "Missing file extension",
			validator:   NewValidator(WithCategories(CategoryDocument)),
			filename:    "testfile",
			contentType: "application/pdf",
			fileSize:    1024,
			shouldError: true,
			errorMsg:    "file has no extension",
		},
		{
			name:        "Disallowed file extension",
			validator:   NewValidator(WithCategories(CategoryDocument)),
			filename:    "test.exe",
			contentType: "application/octet-stream",
			fileSize:    1024,
			shouldError: true,
			errorMsg:    "content type \"application/octet-stream\" is not allowed",
		},
		{
			name:        "Content type doesn't match extension",
			validator:   NewValidator(WithCategories(CategoryImage, CategoryDocument)),
			filename:    "test.pdf",
			contentType: "image/jpeg",
			fileSize:    1024,
			shouldError: true,
			errorMsg:    "content type \"image/jpeg\" does not match extension \".pdf\"",
		},
		{
			name:        "Valid image file",
			validator:   NewValidator(WithCategories(CategoryImage)),
			filename:    "test.jpg",
			contentType: "image/jpeg",
			fileSize:    1024 * 1024,
			shouldError: false,
		},
		{
			name:        "Valid with multiple categories",
			validator:   NewValidator(WithCategories(CategoryImage, CategoryDocument)),
			filename:    "test.jpg",
			contentType: "image/jpeg",
			fileSize:    1024 * 1024,
			shouldError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.validator.Validate(tc.filename, tc.contentType, tc.fileSize)

			if tc.shouldError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tc.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tc.shouldError && err != nil && tc.errorMsg != "" {
				if !contains(err.Error(), tc.errorMsg) {
					t.Errorf("Error message '%s' does not contain expected text '%s'", err.Error(), tc.errorMsg)
				}
			}
		})
	}
}

func TestValidateEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		validator   *FileValidator
		filename    string
		contentType string
		fileSize    int64
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "Zero file size",
			validator:   NewValidator(WithCategories(CategoryText)),
			filename:    "empty.txt",
			contentType: "text/plain",
			fileSize:    0,
			shouldError: true,
			errorMsg:    "file can't be empty",
		},
		{
			name:        "Exact max file size",
			validator:   NewValidator(WithCategories(CategoryText)).WithMaxFileSize(1024),
			filename:    "exact.txt",
			contentType: "text/plain",
			fileSize:    1024,
			shouldError: false,
		},
		{
			name:        "File with uppercase extension",
			validator:   NewValidator(WithCategories(CategoryDocument)),
			filename:    "test.PDF",
			contentType: "application/pdf",
			fileSize:    1024,
			shouldError: false,
		},
		{
			name:        "File with multiple extensions",
			validator:   NewValidator(),
			filename:    "test.tar.gz",
			contentType: "application/gzip",
			fileSize:    1024,
			shouldError: true,
			errorMsg:    "content type \"application/gzip\" is not allowed",
		},
		{
			name:        "Empty filename",
			validator:   NewValidator(WithCategories(CategoryText)),
			filename:    "",
			contentType: "text/plain",
			fileSize:    1024,
			shouldError: true,
			errorMsg:    "file has no extension",
		},
		{
			name:        "Empty content type",
			validator:   NewValidator(WithCategories(CategoryText)),
			filename:    "test.txt",
			contentType: "",
			fileSize:    1024,
			shouldError: true,
			errorMsg:    "content type \"\" is not allowed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.validator.Validate(tc.filename, tc.contentType, tc.fileSize)

			if tc.shouldError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tc.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tc.shouldError && err != nil && tc.errorMsg != "" {
				if !contains(err.Error(), tc.errorMsg) {
					t.Errorf("Error message '%s' does not contain expected text '%s'", err.Error(), tc.errorMsg)
				}
			}
		})
	}
}

// Test each file category individually
func TestFileCategories(t *testing.T) {
	categoryTests := []struct {
		category      string
		validExt      string
		validMimeType string
	}{
		{CategoryDocument, ".pdf", "application/pdf"},
		{CategorySpreadsheet, ".xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		{CategoryPresentation, ".pptx", "application/vnd.openxmlformats-officedocument.presentationml.presentation"},
		{CategoryText, ".txt", "text/plain"},
		{CategoryImage, ".png", "image/png"},
		{CategoryData, ".json", "application/json"},
		{CategoryVideo, ".mp4", "video/mp4"},
	}

	for _, tc := range categoryTests {
		t.Run(tc.category, func(t *testing.T) {
			v := NewValidator(WithCategories(tc.category))

			// Test that the validator accepts files of this category
			err := v.Validate("test"+tc.validExt, tc.validMimeType, 1024)
			if err != nil {
				t.Errorf("Expected no error for %s file but got: %v", tc.category, err)
			}

			// Test that all other categories are rejected
			for _, otherTC := range categoryTests {
				if otherTC.category == tc.category {
					continue
				}

				err := v.Validate("test"+otherTC.validExt, otherTC.validMimeType, 1024)
				if err == nil {
					t.Errorf("Expected error when validating %s file with %s validator, but got none",
						otherTC.category, tc.category)
				}
			}
		})
	}
}

func TestFileTypes(t *testing.T) {
	// Check that FileTypes has entries for all defined categories
	categories := map[string]bool{
		CategoryDocument:     false,
		CategorySpreadsheet:  false,
		CategoryPresentation: false,
		CategoryText:         false,
		CategoryImage:        false,
		CategoryData:         false,
		CategoryVideo:        false,
	}

	for _, ft := range FileTypes {
		categories[ft.Category] = true

		// Verify each file type has required fields
		if ft.MimeType == "" {
			t.Errorf("FileType with category %q has empty MimeType", ft.Category)
		}

		if len(ft.Extensions) == 0 {
			t.Errorf("FileType with MimeType %q has no extensions", ft.MimeType)
		}
	}

	// Check if all categories have at least one file type
	for category, found := range categories {
		if !found {
			t.Errorf("No file types defined for category %q", category)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestMultipleExtensionsPerMimeType(t *testing.T) {
	// Test MIME types that have multiple allowed extensions
	for _, fileType := range FileTypes {
		if len(fileType.Extensions) > 1 {
			v := NewValidator(WithCategories(fileType.Category))

			// All extensions for this MIME type should be valid
			for _, ext := range fileType.Extensions {
				err := v.Validate("test"+ext, fileType.MimeType, 1024)
				if err != nil {
					t.Errorf("Extension %q should be valid for MIME type %q but got error: %v",
						ext, fileType.MimeType, err)
				}
			}
		}
	}
}

func TestExtensionsWithMultipleMimeTypes(t *testing.T) {
	// Create a map of extensions to MIME types
	extToMimes := make(map[string][]string)
	for _, fileType := range FileTypes {
		for _, ext := range fileType.Extensions {
			extToMimes[ext] = append(extToMimes[ext], fileType.MimeType)
		}
	}

	// Find extensions that have multiple MIME types
	for ext, mimeTypes := range extToMimes {
		if len(mimeTypes) > 1 {
			v := NewValidator(
				WithCategories(
					CategoryData,
					CategoryDocument,
					CategoryImage,
					CategoryPresentation,
					CategorySpreadsheet,
					CategoryText,
					CategoryVideo,
				),
			)

			// All MIME types for this extension should be valid
			for _, mimeType := range mimeTypes {
				err := v.Validate("test"+ext, mimeType, 1024)
				if err != nil {
					t.Errorf("MIME type %q should be valid for extension %q but got error: %v",
						mimeType, ext, err)
				}
			}
		}
	}
}

// BenchmarkValidate benchmarks the Validate function
func BenchmarkValidate(b *testing.B) {
	v := NewValidator(WithCategories(CategoryDocument))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = v.Validate("test.pdf", "application/pdf", 1024)
	}
}

// BenchmarkNewValidator benchmarks the NewValidator function
func BenchmarkNewValidator(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewValidator(WithCategories(CategoryDocument, CategoryImage))
	}
}

// TestUtilFunctions tests utility functions
func TestUtilFunctions(t *testing.T) {
	// Test contains function
	testCases := []struct {
		str      string
		substr   string
		expected bool
	}{
		{"hello world", "hello", true},
		{"hello world", "world", true},
		{"hello world", "goodbye", false},
		{"hello world", "", true}, // Empty string is always contained
		{"", "hello", false},      // Empty string contains nothing except empty string
		{"", "", true},            // Empty string contains empty string
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("contains(%q,%q)", tc.str, tc.substr), func(t *testing.T) {
			result := contains(tc.str, tc.substr)
			if result != tc.expected {
				t.Errorf("contains(%q,%q) = %v; want %v", tc.str, tc.substr, result, tc.expected)
			}
		})
	}
}

func TestValidateCustomAllowedExtensions(t *testing.T) {
	// Test the file extension not allowed path
	v := &FileValidator{
		MaxFileSize:      DefaultMaxFileSize,
		AllowedMimeTypes: map[string]bool{"application/pdf": true, "image/jpeg": true},
		AllowedExtensions: map[string][]string{
			".pdf": {"application/pdf"},
			// No .jpg extension here
		},
	}

	// This will trigger the "file extension is not allowed" path
	err := v.Validate("test.jpg", "image/jpeg", 1024)
	if err == nil {
		t.Error("Expected error for unregistered extension, got none")
	}
	if !contains(err.Error(), "file extension \".jpg\" is not allowed") {
		t.Errorf("Unexpected error message: %s", err.Error())
	}

	// Test the content type doesn't match extension path
	v2 := &FileValidator{
		MaxFileSize:      DefaultMaxFileSize,
		AllowedMimeTypes: map[string]bool{"application/pdf": true, "image/jpeg": true, "image/png": true},
		AllowedExtensions: map[string][]string{
			".pdf": {"application/pdf"},
			".jpg": {"image/jpeg"},
			".png": {"image/png"},
		},
	}

	// This will trigger the content type mismatch path
	err = v2.Validate("test.jpg", "image/png", 1024)
	if err == nil {
		t.Error("Expected error for content type not matching extension, got none")
	}
	if !contains(err.Error(), "content type \"image/png\" does not match extension \".jpg\"") {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}
