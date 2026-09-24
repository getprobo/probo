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

package cmdutil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/filevalidation"
)

func TestFileContentType(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"/home/me/report.pdf": "application/pdf",
		"REPORT.PDF":          "application/pdf",
		"notes.md":            "text/markdown",
		"screenshot.png":      "image/png",
		"export.csv":          "text/csv",
		"archive.tar.gz":      "application/octet-stream",
		"no-extension":        "application/octet-stream",
	}

	for filename, want := range cases {
		t.Run(filename, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, want, cmdutil.FileContentType(filename))
		})
	}
}

// TestFileContentType_AcceptedByValidator guards the pairing the API enforces:
// the type the CLI declares must be the one the extension allows.
func TestFileContentType_AcceptedByValidator(t *testing.T) {
	t.Parallel()

	validator := filevalidation.NewValidator(
		filevalidation.WithCategories(
			filevalidation.CategoryDocument,
			filevalidation.CategorySpreadsheet,
			filevalidation.CategoryPresentation,
			filevalidation.CategoryText,
			filevalidation.CategoryImage,
			filevalidation.CategoryData,
			filevalidation.CategoryVideo,
		),
	)

	for _, fileType := range filevalidation.FileTypes {
		for _, ext := range fileType.Extensions {
			filename := "evidence" + ext

			assert.NoErrorf(
				t,
				validator.Validate(filename, cmdutil.FileContentType(filename), 1),
				"%s", filename,
			)
		}
	}
}
