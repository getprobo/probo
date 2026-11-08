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

package probo

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filevalidation"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

func TestUploadInternalControlEvidenceRequest_ValidateRejectsEmptyFile(t *testing.T) {
	t.Parallel()

	req := UploadInternalControlEvidenceRequest{
		InternalControlID: gid.New(gid.NewTenantID(), coredata.InternalControlEntityType),
		File: FileUpload{
			Filename:    "empty.pdf",
			ContentType: "application/pdf",
		},
	}

	require.Error(t, req.Validate())
}

func TestFileUploadValidationErrors_PreservesCause(t *testing.T) {
	t.Parallel()

	oversize := filevalidation.NewValidator(
		filevalidation.WithCategories(filevalidation.CategoryDocument),
		filevalidation.WithMaxFileSize(8),
	).Validate("report.pdf", "application/pdf", 9)
	oversize = fmt.Errorf("cannot validate file: %w", oversize)

	errs, ok := fileUploadValidationErrors(oversize)
	require.True(t, ok)
	require.Len(t, errs, 1)
	assert.Equal(t, "file", errs[0].Field)
	assert.Equal(t, validator.ErrorCodeOutOfRange, errs[0].Code)

	rejectedType := filevalidation.NewValidator(
		filevalidation.WithCategories(filevalidation.CategoryDocument),
	).Validate("payload.exe", "application/octet-stream", 8)
	rejectedType = fmt.Errorf("cannot validate file: %w", rejectedType)

	errs, ok = fileUploadValidationErrors(rejectedType)
	require.True(t, ok)
	require.Len(t, errs, 1)
	assert.Equal(t, validator.ErrorCodeInvalidFormat, errs[0].Code)

	_, ok = fileUploadValidationErrors(fmt.Errorf("cannot upload file to S3"))
	assert.False(t, ok)
}

func TestEvidenceService_ContentTypeForFilename(t *testing.T) {
	t.Parallel()

	svc := EvidenceService{fileValidator: newEvidenceFileValidator()}

	t.Run(
		"infers the content type of accepted evidence files",
		func(t *testing.T) {
			t.Parallel()

			for filename, want := range map[string]string{
				"report.pdf":         "application/pdf",
				"aws-evidence.json":  "application/json",
				"screenshot.PNG":     "image/png",
				"access-review.xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			} {
				got, ok := svc.ContentTypeForFilename(filename)
				require.True(t, ok, filename)
				assert.Equal(t, want, got, filename)
			}
		},
	)

	t.Run(
		"rejects files without an accepted extension",
		func(t *testing.T) {
			t.Parallel()

			for _, filename := range []string{"payload.exe", "no-extension"} {
				_, ok := svc.ContentTypeForFilename(filename)
				assert.False(t, ok, filename)
			}
		},
	)
}

func TestEvidenceService_MaxFileSize(t *testing.T) {
	t.Parallel()

	svc := EvidenceService{fileValidator: newEvidenceFileValidator()}

	assert.Equal(t, int64(filevalidation.DefaultMaxFileSize), svc.MaxFileSize())
}
