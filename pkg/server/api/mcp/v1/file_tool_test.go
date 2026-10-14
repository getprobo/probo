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
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeFileContent(t *testing.T) {
	t.Parallel()

	t.Run("decodes a file at the size limit", func(t *testing.T) {
		t.Parallel()

		raw := bytes.Repeat([]byte("a"), 5)
		payload, err := decodeFileContent(base64.StdEncoding.EncodeToString(raw), int64(len(raw)))

		require.NoError(t, err)
		assert.Equal(t, raw, payload)
	})

	t.Run("rejects a payload larger than the limit", func(t *testing.T) {
		t.Parallel()

		raw := bytes.Repeat([]byte("a"), 8)
		payload, err := decodeFileContent(base64.StdEncoding.EncodeToString(raw), 5)

		require.Error(t, err)
		assert.Nil(t, payload)
		assert.EqualError(t, err, "file size exceeds maximum allowed size of 5 bytes")
	})

	t.Run("rejects invalid base64 without decoding the rest", func(t *testing.T) {
		t.Parallel()

		payload, err := decodeFileContent("!!!!", 50)

		require.Error(t, err)
		assert.Nil(t, payload)
		assert.EqualError(t, err, "content_base64 is not valid base64")
	})

	t.Run("rejects an oversized encoded string before reading it all", func(t *testing.T) {
		t.Parallel()

		// A valid prefix plus a huge tail. The reader stops once the decoded
		// size passes the limit, so the tail is never decoded.
		encoded := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("a"), 8)) + strings.Repeat("A", 1<<20)
		payload, err := decodeFileContent(encoded, 5)

		require.Error(t, err)
		assert.Nil(t, payload)
		assert.EqualError(t, err, "file size exceeds maximum allowed size of 5 bytes")
	})
}
