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

package types

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func NewTaskPicture(picture *coredata.TaskPicture, file *File) *TaskPicture {
	return &TaskPicture{
		ID:             picture.ID,
		OrganizationID: picture.OrganizationID,
		TaskID:         picture.TaskID,
		FileID:         picture.FileID,
		File:           file,
		LinearAssetURL: picture.LinearAssetURL,
		CreatedAt:      picture.CreatedAt,
		UpdatedAt:      picture.UpdatedAt,
	}
}

func NewListTaskPicturesOutput(
	picturePage *page.Page[*coredata.TaskPicture, coredata.TaskPictureOrderField],
	files map[gid.GID]*File,
) ListTaskPicturesOutput {
	pictures := make([]*TaskPicture, 0, len(picturePage.Data))
	for _, picture := range picturePage.Data {
		pictures = append(pictures, NewTaskPicture(picture, files[picture.FileID]))
	}

	var nextCursor *page.CursorKey

	if picturePage.Info.HasNext && len(picturePage.Data) > 0 {
		cursorKey := picturePage.Data[len(picturePage.Data)-1].CursorKey(picturePage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListTaskPicturesOutput{
		NextCursor:   nextCursor,
		TaskPictures: pictures,
	}
}
