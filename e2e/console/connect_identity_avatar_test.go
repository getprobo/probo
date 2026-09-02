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

package console_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestConnectIdentity_AvatarUpload(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	const uploadMutation = `
		mutation UpdateIdentityAvatar($input: UpdateIdentityAvatarInput!) {
			updateIdentityAvatar(input: $input) {
				identity {
					id
					avatar {
						id
						fileName
						downloadUrl
					}
				}
			}
		}
	`

	pngContent := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}

	var uploadResult struct {
		UpdateIdentityAvatar struct {
			Identity struct {
				ID     string `json:"id"`
				Avatar *struct {
					ID          string `json:"id"`
					FileName    string `json:"fileName"`
					DownloadURL string `json:"downloadUrl"`
				} `json:"avatar"`
			} `json:"identity"`
		} `json:"updateIdentityAvatar"`
	}

	err := owner.ExecuteConnectWithFile(uploadMutation, map[string]any{
		"input": map[string]any{
			"file": nil,
		},
	}, "input.file", testutil.UploadFile{
		Filename:    "avatar.png",
		ContentType: "image/png",
		Content:     pngContent,
	}, &uploadResult)
	require.NoError(t, err)

	require.NotNil(t, uploadResult.UpdateIdentityAvatar.Identity.Avatar)
	assert.Equal(t, "avatar.png", uploadResult.UpdateIdentityAvatar.Identity.Avatar.FileName)
	assert.Contains(
		t,
		uploadResult.UpdateIdentityAvatar.Identity.Avatar.DownloadURL,
		"/api/files/v1/public/",
	)

	const deleteMutation = `
		mutation DeleteIdentityAvatar {
			deleteIdentityAvatar {
				identity {
					id
					avatar {
						id
					}
				}
			}
		}
	`

	var deleteResult struct {
		DeleteIdentityAvatar struct {
			Identity struct {
				Avatar *struct {
					ID string `json:"id"`
				} `json:"avatar"`
			} `json:"identity"`
		} `json:"deleteIdentityAvatar"`
	}

	err = owner.ExecuteConnect(deleteMutation, map[string]any{}, &deleteResult)
	require.NoError(t, err)
	assert.Nil(t, deleteResult.DeleteIdentityAvatar.Identity.Avatar)
}
