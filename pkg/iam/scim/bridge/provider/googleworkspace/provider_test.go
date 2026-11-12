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

package googleworkspace

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admin "google.golang.org/api/admin/directory/v1"
)

func TestScimUserFromDirectory_IncludesPhotoETag(t *testing.T) {
	t.Parallel()

	user := scimUserFromDirectory(&admin.User{
		PrimaryEmail: "ada@example.com",
		Id:           "directory-id",
		Name: &admin.UserName{
			FullName:   "Ada Lovelace",
			GivenName:  "Ada",
			FamilyName: "Lovelace",
		},
		ThumbnailPhotoEtag: "photo-etag",
	})

	assert.Equal(t, "ada@example.com", user.UserName)
	assert.Equal(t, "Ada Lovelace", user.DisplayName)
	assert.Equal(t, "directory-id", user.ExternalID)
	assert.Equal(t, "photo-etag", user.PhotoETag)
	assert.True(t, user.PhotoETagKnown)
	assert.True(t, user.Active)
}

func TestDecodeGooglePhoto(t *testing.T) {
	t.Parallel()

	// 0xfb, 0xff encodes as "+/8=", which uses the alphabet and padding
	// Google rewrites before the photo is stored.
	raw := []byte{0xfb, 0xff}
	encoded := base64.StdEncoding.EncodeToString(raw)
	require.Contains(t, encoded, "+")
	require.Contains(t, encoded, "/")
	require.Contains(t, encoded, "=")

	googleEncoded := strings.NewReplacer("+", "-", "/", "_", "=", "*").Replace(encoded)
	decoded, err := decodeGooglePhoto(googleEncoded)

	require.NoError(t, err)
	assert.Equal(t, raw, decoded)

	dotPadding := strings.ReplaceAll(googleEncoded, "*", ".")
	decoded, err = decodeGooglePhoto(dotPadding)

	require.NoError(t, err)
	assert.Equal(t, raw, decoded)
}

func TestGooglePhotoMIME(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "image/png", googlePhotoMIME("PNG"))
	assert.Equal(t, "image/jpeg", googlePhotoMIME("JPEG"))
	assert.Equal(t, "image/jpeg", googlePhotoMIME(""))
}
