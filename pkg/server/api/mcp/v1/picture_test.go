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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/task"
	"go.probo.inc/probo/pkg/validator"
)

func TestDecodePictureContent(t *testing.T) {
	t.Parallel()

	const png = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="

	body, err := decodePictureContent("  " + png + "\n")
	require.NoError(t, err)
	assert.Equal(t, []byte{0x89, 0x50, 0x4e, 0x47}, body[:4])

	_, err = decodePictureContent("not base64!!!")
	var validationErrors validator.ValidationErrors
	require.ErrorAs(t, err, &validationErrors)
	assert.Equal(t, "content_base64", validationErrors[0].Field)

	oversized := strings.Repeat("A", base64.StdEncoding.EncodedLen(task.MaxPictureBytes+1)+4)
	_, err = decodePictureContent(oversized)
	require.ErrorAs(t, err, &validationErrors)
	assert.Equal(t, "content_base64", validationErrors[0].Field)
}
