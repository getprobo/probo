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
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
)

func TestInternalControlLinkEvidenceUpload(t *testing.T) {
	t.Parallel()

	internalControlID := gid.New(gid.NewTenantID(), coredata.InternalControlEntityType)

	t.Run("builds a uri file from a trimmed url and name", func(t *testing.T) {
		t.Parallel()

		upload, err := internalControlLinkEvidenceUpload(&types.AddInternalControlLinkEvidenceInput{
			InternalControlID: internalControlID,
			Name:              "  ACC-16 policy  ",
			URL:               "  https://example.com/acc-16  ",
		})
		require.NoError(t, err)
		assert.Equal(t, internalControlID, upload.InternalControlID)
		assert.Equal(t, "ACC-16 policy.uri", upload.File.Filename)
		assert.Equal(t, "text/uri-list", upload.File.ContentType)
		assert.Equal(t, int64(len("https://example.com/acc-16")), upload.File.Size)

		body, err := io.ReadAll(upload.File.Content)
		require.NoError(t, err)
		assert.Equal(t, "https://example.com/acc-16", string(body))
	})

	t.Run("rejects a non-http url", func(t *testing.T) {
		t.Parallel()

		_, err := internalControlLinkEvidenceUpload(&types.AddInternalControlLinkEvidenceInput{
			InternalControlID: internalControlID,
			Name:              "Policy",
			URL:               "javascript:alert(1)",
		})
		require.Error(t, err)
	})

	t.Run("rejects a blank name", func(t *testing.T) {
		t.Parallel()

		_, err := internalControlLinkEvidenceUpload(&types.AddInternalControlLinkEvidenceInput{
			InternalControlID: internalControlID,
			Name:              "   ",
			URL:               "https://example.com/acc-16",
		})
		require.Error(t, err)
	})
}
