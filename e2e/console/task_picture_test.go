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
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func tinyPNG(t *testing.T) []byte {
	t.Helper()

	raw, err := base64.StdEncoding.DecodeString(
		"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==",
	)
	require.NoError(t, err)

	return raw
}

func TestTaskPicture_UploadListAndDelete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	taskID := factory.NewTaskWithoutMeasure(owner).WithName("Task for pictures").Create()

	uploadQuery := `
		mutation UploadTaskPicture($input: UploadTaskPictureInput!) {
			uploadTaskPicture(input: $input) {
				taskPictureEdge {
					node {
						id
						file {
							fileName
							mimeType
						}
					}
				}
			}
		}
	`

	var uploaded struct {
		UploadTaskPicture struct {
			TaskPictureEdge struct {
				Node struct {
					ID   string `json:"id"`
					File struct {
						FileName string `json:"fileName"`
						MimeType string `json:"mimeType"`
					} `json:"file"`
				} `json:"node"`
			} `json:"taskPictureEdge"`
		} `json:"uploadTaskPicture"`
	}

	err := owner.ExecuteWithFile(
		uploadQuery,
		map[string]any{
			"input": map[string]any{
				"taskId": taskID,
				"file":   nil,
			},
		},
		"input.file",
		testutil.UploadFile{
			Filename:    "pixel.png",
			ContentType: "image/png",
			Content:     tinyPNG(t),
		},
		&uploaded,
	)
	require.NoError(t, err)

	picture := uploaded.UploadTaskPicture.TaskPictureEdge.Node
	assert.NotEmpty(t, picture.ID)
	assert.Equal(t, "pixel.png", picture.File.FileName)
	assert.Equal(t, "image/png", picture.File.MimeType)

	listQuery := `
		query($id: ID!) {
			node(id: $id) {
				... on Task {
					pictures(first: 10) {
						edges {
							node { id }
						}
					}
				}
			}
		}
	`

	var listed struct {
		Node struct {
			Pictures struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"pictures"`
		} `json:"node"`
	}

	err = owner.Execute(listQuery, map[string]any{"id": taskID}, &listed)
	require.NoError(t, err)
	require.Len(t, listed.Node.Pictures.Edges, 1)
	assert.Equal(t, picture.ID, listed.Node.Pictures.Edges[0].Node.ID)

	deleteQuery := `
		mutation DeleteTaskPicture($input: DeleteTaskPictureInput!) {
			deleteTaskPicture(input: $input) {
				deletedTaskPictureId
			}
		}
	`

	var deleted struct {
		DeleteTaskPicture struct {
			DeletedTaskPictureID string `json:"deletedTaskPictureId"`
		} `json:"deleteTaskPicture"`
	}

	err = owner.Execute(deleteQuery, map[string]any{
		"input": map[string]any{
			"taskPictureId": picture.ID,
		},
	}, &deleted)
	require.NoError(t, err)
	assert.Equal(t, picture.ID, deleted.DeleteTaskPicture.DeletedTaskPictureID)
}

func TestTaskPicture_AcceptsTextFile(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()

	query := `
		mutation UploadTaskPicture($input: UploadTaskPictureInput!) {
			uploadTaskPicture(input: $input) {
				taskPictureEdge {
					node {
						id
						file {
							fileName
							mimeType
						}
					}
				}
			}
		}
	`

	var uploaded struct {
		UploadTaskPicture struct {
			TaskPictureEdge struct {
				Node struct {
					ID   string `json:"id"`
					File struct {
						FileName string `json:"fileName"`
						MimeType string `json:"mimeType"`
					} `json:"file"`
				} `json:"node"`
			} `json:"taskPictureEdge"`
		} `json:"uploadTaskPicture"`
	}

	err := owner.ExecuteWithFile(
		query,
		map[string]any{
			"input": map[string]any{
				"taskId": taskID,
				"file":   nil,
			},
		},
		"input.file",
		testutil.UploadFile{
			Filename:    "notes.txt",
			ContentType: "text/plain",
			Content:     []byte("not a picture"),
		},
		&uploaded,
	)
	require.NoError(t, err)

	file := uploaded.UploadTaskPicture.TaskPictureEdge.Node.File
	assert.Equal(t, "notes.txt", file.FileName)
	assert.Equal(t, "text/plain", file.MimeType)
}

func TestTaskPicture_RejectsUnsupportedFile(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()

	query := `
		mutation UploadTaskPicture($input: UploadTaskPictureInput!) {
			uploadTaskPicture(input: $input) {
				taskPictureEdge {
					node { id }
				}
			}
		}
	`

	err := owner.ExecuteWithFile(
		query,
		map[string]any{
			"input": map[string]any{
				"taskId": taskID,
				"file":   nil,
			},
		},
		"input.file",
		testutil.UploadFile{
			Filename:    "malware.exe",
			ContentType: "application/octet-stream",
			Content:     []byte("not a supported file"),
		},
		nil,
	)
	testutil.RequireErrorCode(t, err, "INVALID", "unsupported file types must be rejected")
}

func TestTaskPicture_ViewerCannotUpload(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()

	query := `
		mutation UploadTaskPicture($input: UploadTaskPictureInput!) {
			uploadTaskPicture(input: $input) {
				taskPictureEdge {
					node { id }
				}
			}
		}
	`

	err := viewer.ExecuteWithFile(
		query,
		map[string]any{
			"input": map[string]any{
				"taskId": taskID,
				"file":   nil,
			},
		},
		"input.file",
		testutil.UploadFile{
			Filename:    "pixel.png",
			ContentType: "image/png",
			Content:     tinyPNG(t),
		},
		nil,
	)
	testutil.RequireForbiddenError(t, err, "viewer cannot upload task pictures")
}

func TestTaskPicture_ViewerCanList(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()

	uploadQuery := `
		mutation UploadTaskPicture($input: UploadTaskPictureInput!) {
			uploadTaskPicture(input: $input) {
				taskPictureEdge {
					node { id }
				}
			}
		}
	`

	var uploaded struct {
		UploadTaskPicture struct {
			TaskPictureEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"taskPictureEdge"`
		} `json:"uploadTaskPicture"`
	}

	err := owner.ExecuteWithFile(
		uploadQuery,
		map[string]any{
			"input": map[string]any{
				"taskId": taskID,
				"file":   nil,
			},
		},
		"input.file",
		testutil.UploadFile{
			Filename:    "pixel.png",
			ContentType: "image/png",
			Content:     tinyPNG(t),
		},
		&uploaded,
	)
	require.NoError(t, err)

	query := `
		query($id: ID!) {
			node(id: $id) {
				... on Task {
					pictures(first: 10) {
						edges {
							node { id }
						}
					}
				}
			}
		}
	`

	var listed struct {
		Node struct {
			Pictures struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"pictures"`
		} `json:"node"`
	}

	err = viewer.Execute(query, map[string]any{"id": taskID}, &listed)
	require.NoError(t, err)
	require.Len(t, listed.Node.Pictures.Edges, 1)
	assert.Equal(t, uploaded.UploadTaskPicture.TaskPictureEdge.Node.ID, listed.Node.Pictures.Edges[0].Node.ID)
}
