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

package task

import (
	"context"
	"fmt"
	"strings"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	tasksync "go.probo.inc/probo/pkg/task/sync"
)

type attachmentReader struct {
	pg    *pg.Client
	files *filemanager.Service
}

func (r *attachmentReader) ReadAttachment(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	fileID gid.GID,
) (tasksync.Attachment, error) {
	if r == nil || r.files == nil {
		return tasksync.Attachment{}, fmt.Errorf("cannot read attachment %q: storage is not configured", fileID)
	}

	file := &coredata.File{}

	err := r.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := file.LoadByID(ctx, conn, scope, fileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return tasksync.Attachment{}, fmt.Errorf("cannot read attachment %q: %w", fileID, err)
	}

	if file.OrganizationID != organizationID {
		return tasksync.Attachment{}, fmt.Errorf("cannot read attachment %q: file belongs to another organization", fileID)
	}

	if file.DeletedAt != nil {
		return tasksync.Attachment{}, fmt.Errorf("cannot read attachment %q: file is deleted", fileID)
	}

	body, err := r.files.GetFileBytes(ctx, file)
	if err != nil {
		return tasksync.Attachment{}, fmt.Errorf("cannot read attachment %q: %w", fileID, err)
	}

	name := strings.TrimSpace(file.FileName)
	if name == "" {
		name = "file"
	}

	mimeType := strings.TrimSpace(file.MimeType)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return tasksync.Attachment{
		Name:     name,
		MimeType: mimeType,
		Body:     body,
	}, nil
}
