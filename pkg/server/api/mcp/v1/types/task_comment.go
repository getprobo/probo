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
	"fmt"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

func NewTaskComment(c *coredata.TaskComment) (*TaskComment, error) {
	content, err := richTextToMarkdown(c.Content)
	if err != nil {
		return nil, fmt.Errorf("cannot convert task comment content to markdown: %w", err)
	}

	return &TaskComment{
		ID:             c.ID,
		OrganizationID: c.OrganizationID,
		TaskID:         c.TaskID,
		OwnerID:        c.OwnerID,
		Content:        content,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}, nil
}

func NewListTaskCommentsOutput(
	commentPage *page.Page[*coredata.TaskComment, coredata.TaskCommentOrderField],
) (ListTaskCommentsOutput, error) {
	comments := make([]*TaskComment, 0, len(commentPage.Data))
	for _, v := range commentPage.Data {
		comment, err := NewTaskComment(v)
		if err != nil {
			return ListTaskCommentsOutput{}, err
		}

		comments = append(comments, comment)
	}

	var nextCursor *page.CursorKey

	if commentPage.Info.HasNext && len(commentPage.Data) > 0 {
		cursorKey := commentPage.Data[len(commentPage.Data)-1].CursorKey(commentPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListTaskCommentsOutput{
		NextCursor:   nextCursor,
		TaskComments: comments,
	}, nil
}
