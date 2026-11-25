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
	"net/http/httptest"
	"testing"

	"github.com/elimity-com/scim"
	scimfilter "github.com/scim2/filter-parser/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/iam/avatar"
)

func TestSelectPhotoURI(t *testing.T) {
	t.Parallel()

	t.Run("prefers the primary photo over a thumbnail", func(t *testing.T) {
		t.Parallel()

		uri := selectPhotoURI([]any{
			map[string]any{"value": "https://cdn.example.com/thumb.jpg", "type": "thumbnail"},
			map[string]any{"value": "https://cdn.example.com/face.jpg", "type": "photo", "primary": true},
		})

		assert.Equal(t, "https://cdn.example.com/face.jpg", uri)
	})

	t.Run("falls back to the first photo with a value", func(t *testing.T) {
		t.Parallel()

		uri := selectPhotoURI([]any{
			map[string]any{"value": " "},
			map[string]any{"value": " https://cdn.example.com/first.jpg "},
			map[string]any{"value": "https://cdn.example.com/second.jpg"},
		})

		assert.Equal(t, "https://cdn.example.com/first.jpg", uri)
	})

	t.Run("selects nothing from a value that is not a list", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, selectPhotoURI("https://cdn.example.com/face.jpg"))
		assert.Empty(t, selectPhotoURI(nil))
	})

	t.Run("selects nothing from an empty list", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, selectPhotoURI([]any{}))
	})
}

func TestDecodePhotoDataURI(t *testing.T) {
	t.Parallel()

	t.Run("decodes a base64 data uri", func(t *testing.T) {
		t.Parallel()

		value := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString([]byte("face"))

		body, err := decodePhotoDataURI(value)

		require.NoError(t, err)
		assert.Equal(t, []byte("face"), body)
	})

	t.Run("rejects a data uri that is not base64", func(t *testing.T) {
		t.Parallel()

		body, err := decodePhotoDataURI("data:image/jpeg,face")

		assert.Nil(t, body)
		assert.ErrorIs(t, err, errUnsupportedPhoto)
	})

	t.Run("accepts a payload of exactly the avatar size limit", func(t *testing.T) {
		t.Parallel()

		value := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(make([]byte, avatar.MaxFileSize))

		body, err := decodePhotoDataURI(value)

		require.NoError(t, err)
		assert.Len(t, body, avatar.MaxFileSize)
	})

	t.Run("rejects a payload above the avatar size limit", func(t *testing.T) {
		t.Parallel()

		value := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(make([]byte, avatar.MaxFileSize+1))

		body, err := decodePhotoDataURI(value)

		assert.Nil(t, body)
		assert.ErrorIs(t, err, errUnsupportedPhoto)
	})

	t.Run("rejects an empty payload", func(t *testing.T) {
		t.Parallel()

		body, err := decodePhotoDataURI("data:image/jpeg;base64,")

		assert.Nil(t, body)
		assert.ErrorIs(t, err, errUnsupportedPhoto)
	})
}

func TestReadPhoto(t *testing.T) {
	t.Parallel()

	t.Run("downloads an https url", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("face"))
		}))
		t.Cleanup(server.Close)

		svc := &Service{httpClient: server.Client()}

		body, err := svc.readPhoto(context.Background(), server.URL+"/face.jpg")

		require.NoError(t, err)
		assert.Equal(t, []byte("face"), body)
	})

	t.Run("rejects a url that is not https", func(t *testing.T) {
		t.Parallel()

		svc := &Service{}

		for _, value := range []string{
			"http://cdn.example.com/face.jpg",
			"https:///face.jpg",
			"cdn.example.com/face.jpg",
		} {
			body, err := svc.readPhoto(context.Background(), value)

			assert.Nil(t, body, value)
			assert.ErrorIs(t, err, errUnsupportedPhoto, value)
		}
	})

	t.Run("reports a failed download without the url", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "gone", http.StatusNotFound)
		}))
		t.Cleanup(server.Close)

		svc := &Service{httpClient: server.Client()}

		body, err := svc.readPhoto(context.Background(), server.URL+"/face.jpg")

		assert.Nil(t, body)
		require.ErrorIs(t, err, errPhotoFetch)
		assert.NotContains(t, err.Error(), server.URL)
	})
}

func TestIncomingPhoto_NoPhotoAttribute(t *testing.T) {
	t.Parallel()

	svc := &Service{}

	staged, err := svc.incomingPhoto(context.Background(), "")

	require.NoError(t, err)
	assert.Nil(t, staged)
}

func TestParseUserFromAttributes_Photos(t *testing.T) {
	t.Parallel()

	attrs := ParseUserFromAttributes(scim.ResourceAttributes{
		"userName": "ada@example.com",
		"photos": []any{
			map[string]any{"value": "https://cdn.example.com/ada.jpg", "type": "photo", "primary": true},
		},
	})

	assert.Equal(t, "https://cdn.example.com/ada.jpg", attrs.PhotoURI)
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

	assert.Equal(t, "https://cdn.example.com/ada.jpg", attrs.PhotoURI)
}
