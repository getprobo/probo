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

package filevalidation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/filevalidation"
)

func TestResolveContentType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		contentType string
		filename    string
		want        string
	}{
		{
			name:        "octet-stream pdf uses the extension",
			contentType: "application/octet-stream",
			filename:    "report.pdf",
			want:        "application/pdf",
		},
		{
			name:        "empty type uses the extension",
			contentType: "",
			filename:    "notes.md",
			want:        "text/markdown",
		},
		{
			name:        "charset parameter is dropped",
			contentType: "text/plain; charset=utf-8",
			filename:    "notes.txt",
			want:        "text/plain",
		},
		{
			name:        "avi does not use the stdlib vendor type",
			contentType: "application/octet-stream",
			filename:    "clip.avi",
			want:        "video/x-msvideo",
		},
		{
			name:        "uppercase extension",
			contentType: "application/octet-stream",
			filename:    "photo.PNG",
			want:        "image/png",
		},
		{
			name:        "explicit type is kept",
			contentType: "application/pdf",
			filename:    "report.pdf",
			want:        "application/pdf",
		},
		{
			name:        "unknown extension stays octet-stream",
			contentType: "application/octet-stream",
			filename:    "payload.exe",
			want:        "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, filevalidation.ResolveContentType(tt.contentType, tt.filename))
		})
	}
}

func TestValidate_ReturnsFileValidationError(t *testing.T) {
	t.Parallel()

	validator := filevalidation.NewValidator(filevalidation.WithCategories(filevalidation.CategoryDocument))
	err := validator.Validate("payload.exe", "application/octet-stream", 128)
	require.Error(t, err)

	var fileErr *filevalidation.Error
	require.ErrorAs(t, err, &fileErr)
	assert.Contains(t, fileErr.Error(), "application/octet-stream")
}

func TestResolveContentType_AcceptedByEvidenceCategories(t *testing.T) {
	t.Parallel()

	validator := filevalidation.NewValidator(
		filevalidation.WithCategories(
			filevalidation.CategoryDocument,
			filevalidation.CategorySpreadsheet,
			filevalidation.CategoryPresentation,
			filevalidation.CategoryData,
			filevalidation.CategoryText,
			filevalidation.CategoryImage,
			filevalidation.CategoryVideo,
		),
	)

	filenames := []string{
		"report.pdf",
		"notes.md",
		"clip.avi",
		"sheet.xlsx",
		"data.csv",
		"photo.PNG",
		"link.uri",
		"auth.log",
	}

	for _, filename := range filenames {
		t.Run(filename, func(t *testing.T) {
			t.Parallel()

			contentType := filevalidation.ResolveContentType("application/octet-stream", filename)
			err := validator.Validate(filename, contentType, 128)
			require.NoError(t, err)
		})
	}
}
