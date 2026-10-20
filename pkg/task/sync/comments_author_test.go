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

package tasksync

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestPublicAvatarURL(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.NilTenant, coredata.FileEntityType)
	want := "https://app.example.com/api/files/v1/public/" + fileID.String()

	t.Run(
		"public file path",
		func(t *testing.T) {
			t.Parallel()

			avatarURL, err := publicAvatarURL("https://app.example.com", fileID)

			require.NoError(t, err)
			assert.Equal(t, want, avatarURL)
		},
	)

	t.Run(
		"replaces a base path",
		func(t *testing.T) {
			t.Parallel()

			avatarURL, err := publicAvatarURL("https://app.example.com/console", fileID)

			require.NoError(t, err)
			assert.Equal(t, want, avatarURL)
		},
	)

	t.Run(
		"empty base",
		func(t *testing.T) {
			t.Parallel()

			avatarURL, err := publicAvatarURL("  ", fileID)

			require.NoError(t, err)
			assert.Empty(t, avatarURL)
		},
	)

	t.Run(
		"invalid base",
		func(t *testing.T) {
			t.Parallel()

			_, err := publicAvatarURL("not a url", fileID)

			require.Error(t, err)
		},
	)
}
