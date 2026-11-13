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

package scim

import (
	"context"
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/elimity-com/scim"
	scimfilter "github.com/scim2/filter-parser/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelectPhotoURI(t *testing.T) {
	t.Parallel()

	t.Run("prefers the primary photo over a thumbnail", func(t *testing.T) {
		t.Parallel()

		uri, ok := selectPhotoURI([]any{
			map[string]any{"value": "https://cdn.example.com/thumb.jpg", "type": "thumbnail"},
			map[string]any{"value": "https://cdn.example.com/face.jpg", "type": "photo", "primary": true},
		})

		require.True(t, ok)
		assert.Equal(t, "https://cdn.example.com/face.jpg", uri)
	})

	t.Run("prefers a photo type when none is primary", func(t *testing.T) {
		t.Parallel()

		uri, ok := selectPhotoURI([]any{
			map[string]any{"value": "https://cdn.example.com/thumb.jpg", "type": "thumbnail", "primary": "false"},
			map[string]any{"value": "https://cdn.example.com/face.jpg", "type": "photo"},
		})

		require.True(t, ok)
		assert.Equal(t, "https://cdn.example.com/face.jpg", uri)
	})

	t.Run("accepts a string primary flag", func(t *testing.T) {
		t.Parallel()

		uri, ok := selectPhotoURI([]any{
			map[string]any{"value": "https://cdn.example.com/face.jpg", "primary": "true"},
		})

		require.True(t, ok)
		assert.Equal(t, "https://cdn.example.com/face.jpg", uri)
	})

	t.Run("treats an empty list as a removed photo", func(t *testing.T) {
		t.Parallel()

		uri, ok := selectPhotoURI([]any{})

		require.True(t, ok)
		assert.Empty(t, uri)
	})
}

func TestPrepareSCIMPhoto(t *testing.T) {
	t.Parallel()

	svc := &Service{}

	t.Run("decodes a data uri without a network call", func(t *testing.T) {
		t.Parallel()

		payload := base64.StdEncoding.EncodeToString([]byte("face"))
		value := "data:image/jpeg;base64," + payload
		photo, reason := svc.prepareSCIMPhoto(context.Background(), &value, nil, "")

		require.Empty(t, reason)
		require.NotNil(t, photo)
		assert.Equal(t, []byte("face"), photo.Body)
		assert.NotEmpty(t, photo.ETag)
		assert.False(t, photo.Remove)
	})

	t.Run("keeps an explicit etag", func(t *testing.T) {
		t.Parallel()

		payload := base64.StdEncoding.EncodeToString([]byte("face"))
		value := "data:image/png;base64," + payload
		etag := "directory-etag"
		photo, reason := svc.prepareSCIMPhoto(context.Background(), &value, &etag, "")

		require.Empty(t, reason)
		require.NotNil(t, photo)
		assert.Equal(t, "directory-etag", photo.ETag)
	})

	t.Run("rejects an http photo url", func(t *testing.T) {
		t.Parallel()

		value := "http://cdn.example.com/face.jpg"
		photo, reason := svc.prepareSCIMPhoto(context.Background(), &value, nil, "")

		assert.Nil(t, photo)
		assert.Equal(t, "unsupported image", reason)
	})

	t.Run("rejects a photo url with user info", func(t *testing.T) {
		t.Parallel()

		value := "https://user:secret@cdn.example.com/face.jpg"
		photo, reason := svc.prepareSCIMPhoto(context.Background(), &value, nil, "")

		assert.Nil(t, photo)
		assert.Equal(t, "unsupported image", reason)
	})

	t.Run("clears the photo when the attribute is empty", func(t *testing.T) {
		t.Parallel()

		value := ""
		photo, reason := svc.prepareSCIMPhoto(context.Background(), &value, nil, "")

		require.Empty(t, reason)
		require.NotNil(t, photo)
		assert.True(t, photo.Remove)
	})

	t.Run("leaves the avatar alone when no photo attribute is present", func(t *testing.T) {
		t.Parallel()

		photo, reason := svc.prepareSCIMPhoto(context.Background(), nil, nil, "")

		assert.Nil(t, photo)
		assert.Empty(t, reason)
	})

	t.Run("skips the download when the etag is unchanged", func(t *testing.T) {
		t.Parallel()

		client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			t.Fatal("unchanged photo was downloaded")

			return nil, nil
		})}
		unchanged := &Service{httpClient: client}
		value := "https://cdn.example.com/face.jpg"
		etag := "directory-etag"
		photo, reason := unchanged.prepareSCIMPhoto(context.Background(), &value, &etag, etag)

		require.Empty(t, reason)
		require.NotNil(t, photo)
		assert.Empty(t, photo.Body)
		assert.Equal(t, etag, photo.ETag)
		assert.False(t, photo.Remove)
	})
}

func TestPhotoActionFor(t *testing.T) {
	t.Parallel()

	t.Run("applies a new picture", func(t *testing.T) {
		t.Parallel()

		action := photoActionFor("", &scimPhoto{ETag: "new", Body: []byte("face")})

		assert.Equal(t, photoActionApply, action)
	})

	t.Run("skips an unchanged etag", func(t *testing.T) {
		t.Parallel()

		action := photoActionFor("same", &scimPhoto{ETag: "same", Body: []byte("face")})

		assert.Equal(t, photoActionNone, action)
	})

	t.Run("clears a removed picture", func(t *testing.T) {
		t.Parallel()

		action := photoActionFor("old", &scimPhoto{Remove: true})

		assert.Equal(t, photoActionClear, action)
	})

	t.Run("does not clear an avatar that was never synced", func(t *testing.T) {
		t.Parallel()

		action := photoActionFor("", &scimPhoto{Remove: true})

		assert.Equal(t, photoActionNone, action)
	})
}

func TestParseUserFromAttributes_Photos(t *testing.T) {
	t.Parallel()

	attrs := ParseUserFromAttributes(scim.ResourceAttributes{
		"userName": "ada@example.com",
		"photos": []any{
			map[string]any{"value": "https://cdn.example.com/ada.jpg", "type": "photo", "primary": true},
		},
		"photoETag": "not-a-core-attribute",
		PhotoExtensionURN: map[string]any{
			"photoETag": "etag-1",
		},
	})

	require.NotNil(t, attrs.PhotoURI)
	assert.Equal(t, "https://cdn.example.com/ada.jpg", *attrs.PhotoURI)
	require.NotNil(t, attrs.PhotoETag)
	assert.Equal(t, "etag-1", *attrs.PhotoETag)
}

func TestParseUserFromPatchOperations_ReplacesPhoto(t *testing.T) {
	t.Parallel()

	path, err := scimfilter.ParsePath([]byte("photos"))
	require.NoError(t, err)

	attrs := ParseUserFromPatchOperations([]scim.PatchOperation{{
		Op:   "replace",
		Path: &path,
		Value: []any{
			map[string]any{
				"value":   "https://cdn.example.com/ada.jpg",
				"primary": true,
			},
		},
	}})

	require.NotNil(t, attrs.PhotoURI)
	assert.Equal(t, "https://cdn.example.com/ada.jpg", *attrs.PhotoURI)
}

func TestParseUserFromPatchOperations_RemovesPhoto(t *testing.T) {
	t.Parallel()

	path, err := scimfilter.ParsePath([]byte("photos"))
	require.NoError(t, err)

	attrs := ParseUserFromPatchOperations([]scim.PatchOperation{{
		Op:   "remove",
		Path: &path,
	}})

	require.NotNil(t, attrs.PhotoURI)
	assert.Empty(t, *attrs.PhotoURI)
	assert.Nil(t, attrs.PhotoETag)
}
