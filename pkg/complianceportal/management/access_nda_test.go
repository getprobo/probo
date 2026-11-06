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

package management

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

func TestUploadRecordedNDARequest_Validate(t *testing.T) {
	t.Parallel()

	accessID := gid.New(gid.NewTenantID(), coredata.CompliancePortalAccessEntityType)

	t.Run("accepts a pdf signed in the past", func(t *testing.T) {
		t.Parallel()

		req := &UploadRecordedNDARequest{
			AccessID:    accessID,
			FileName:    "signed-nda.pdf",
			ContentType: "application/pdf",
			Size:        128,
			SignedAt:    time.Now().Add(-24 * time.Hour),
		}

		require.NoError(t, req.Validate())
	})

	t.Run("infers pdf when the content type is missing", func(t *testing.T) {
		t.Parallel()

		req := &UploadRecordedNDARequest{
			AccessID: accessID,
			FileName: "signed-nda.pdf",
			Size:     128,
			SignedAt: time.Now().Add(-time.Hour),
		}

		require.NoError(t, req.Validate())
		assert.Equal(t, "application/pdf", req.ContentType)
	})

	t.Run("rejects a non-pdf", func(t *testing.T) {
		t.Parallel()

		req := &UploadRecordedNDARequest{
			AccessID:    accessID,
			FileName:    "signed-nda.txt",
			ContentType: "text/plain",
			Size:        128,
			SignedAt:    time.Now().Add(-time.Hour),
		}

		err := req.Validate()
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "not allowed"))
	})

	t.Run("rejects a missing signature date", func(t *testing.T) {
		t.Parallel()

		req := &UploadRecordedNDARequest{
			AccessID:    accessID,
			FileName:    "signed-nda.pdf",
			ContentType: "application/pdf",
			Size:        128,
		}

		err := req.Validate()
		validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
		require.True(t, ok)
		assert.NotEmpty(t, validationErrors.ByField("signed_at"))
	})
}

func TestEnsureRecordedNDAPDF(t *testing.T) {
	t.Parallel()

	t.Run("accepts a pdf header and preserves the body", func(t *testing.T) {
		t.Parallel()

		body := []byte("%PDF-1.7\nhello")
		reader, err := ensureRecordedNDAPDF(bytes.NewReader(body))
		require.NoError(t, err)

		got, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, body, got)
	})

	t.Run("rejects content without a pdf signature", func(t *testing.T) {
		t.Parallel()

		_, err := ensureRecordedNDAPDF(bytes.NewReader([]byte("not a pdf")))
		validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
		require.True(t, ok)
		assert.NotEmpty(t, validationErrors.ByField("file"))
	})
}
