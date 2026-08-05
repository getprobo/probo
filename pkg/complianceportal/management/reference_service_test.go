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
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReferenceLogoFileForCreate(t *testing.T) {
	t.Parallel()

	t.Run(
		"zero file uses default logo",
		func(t *testing.T) {
			t.Parallel()

			got := referenceLogoFileForCreate(File{})

			require.NotNil(t, got.Content)
			assert.Equal(t, "reference-logo.png", got.Filename)
			assert.Equal(t, "image/png", got.ContentType)
			assert.Positive(t, got.Size)
		},
	)

	t.Run(
		"nil content with filename and size still uses default",
		func(t *testing.T) {
			t.Parallel()

			got := referenceLogoFileForCreate(File{
				Filename:    "client-logo.png",
				Size:        4096,
				ContentType: "image/png",
			})

			require.NotNil(t, got.Content)
			assert.Equal(t, "reference-logo.png", got.Filename)
			assert.Equal(t, "image/png", got.ContentType)
		},
	)

	t.Run(
		"non-nil content is preserved",
		func(t *testing.T) {
			t.Parallel()

			payload := []byte("not-a-real-png")
			input := File{
				Content:     bytes.NewReader(payload),
				Filename:    "upload.png",
				Size:        int64(len(payload)),
				ContentType: "image/png",
			}

			got := referenceLogoFileForCreate(input)

			assert.Equal(t, input.Filename, got.Filename)
			assert.Equal(t, input.Size, got.Size)
			assert.Equal(t, input.ContentType, got.ContentType)
			require.NotNil(t, got.Content)

			read, err := io.ReadAll(got.Content)
			require.NoError(t, err)
			assert.Equal(t, payload, read)
		},
	)
}
