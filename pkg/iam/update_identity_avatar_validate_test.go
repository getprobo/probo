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

package iam

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/validator"
)

func TestUpdateIdentityAvatarRequest_Validate_RejectsSVG(t *testing.T) {
	t.Parallel()

	req := &UpdateIdentityAvatarRequest{
		File: UploadedFile{
			Filename:    "avatar.svg",
			ContentType: "image/svg+xml",
			Size:        128,
		},
	}
	err := req.Validate()
	require.Error(t, err)

	validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
	require.True(t, ok)
	assert.NotEmpty(t, validationErrors.ByField("file"))
	assert.Equal(t, validator.ErrorCodeUnsafeContent, validationErrors.ByField("file")[0].Code)
}

func TestUpdateIdentityAvatarRequest_Validate_RejectsEmptyFile(t *testing.T) {
	t.Parallel()

	req := &UpdateIdentityAvatarRequest{
		File: UploadedFile{
			Filename:    "avatar.png",
			ContentType: "image/png",
			Size:        0,
		},
	}
	err := req.Validate()
	require.Error(t, err)

	validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
	require.True(t, ok)
	assert.NotEmpty(t, validationErrors.ByField("file"))
	assert.Equal(t, validator.ErrorCodeInvalidFormat, validationErrors.ByField("file")[0].Code)
}
