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

package attachment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
)

// Attach checks that every file id in content belongs to the parent
// organization and rewrites image and file nodes from the stored records.
// A file that is not yet attached to this parent gets one row. Rows for
// this parent whose file is no longer in content are deleted. Content with no
// file ids is returned unchanged.
func Attach(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	organizationID gid.GID,
	parentID gid.GID,
	content string,
) (string, error) {
	bound := content

	var ids []gid.GID

	if strings.TrimSpace(content) != "" {
		node, err := prosemirror.Parse(content)
		if err != nil {
			return "", fmt.Errorf("cannot parse attachments: %w", err)
		}

		ids, err = fileIDs(node)
		if err != nil {
			return "", fmt.Errorf("cannot read attachments: %w", err)
		}

		if len(ids) > 0 {
			var attachments coredata.Attachments
			if err := attachments.LoadByFileIDs(ctx, conn, scope, ids); err != nil {
				return "", fmt.Errorf("cannot load attachments: %w", err)
			}

			byFile := make(map[gid.GID][]*coredata.Attachment, len(ids))
			for _, row := range attachments {
				byFile[row.FileID] = append(byFile[row.FileID], row)
			}

			var files coredata.Files
			if err := files.LoadByIDs(ctx, conn, scope, ids); err != nil {
				return "", fmt.Errorf("cannot load attachments: %w", err)
			}

			storedFiles := make(map[gid.GID]*coredata.File, len(files))
			for _, file := range files {
				if file.OrganizationID != organizationID {
					return "", fmt.Errorf("attachment %s is not in this organization", file.ID)
				}

				storedFiles[file.ID] = file
			}

			for _, id := range ids {
				if attachmentFor(byFile[id], organizationID, parentID) != nil {
					continue
				}

				file := storedFiles[id]

				row := coredata.Attachment{
					FileID:         file.ID,
					OrganizationID: organizationID,
					ParentID:       parentID,
					CreatedAt:      time.Now(),
				}
				if err := row.Insert(ctx, conn, scope); err != nil {
					return "", fmt.Errorf("cannot insert attachment %s: %w", file.ID, err)
				}
			}

			stored := make([]prosemirror.StoredFile, 0, len(files))
			for _, file := range files {
				stored = append(stored, prosemirror.StoredFile{
					ID:       file.ID.String(),
					FileName: file.FileName,
					MimeType: file.MimeType,
					Size:     file.FileSize,
				})
			}

			bound, err = prosemirror.BindStoredFiles(content, stored)
			if err != nil {
				return "", fmt.Errorf("cannot bind attachments: %w", err)
			}
		}
	}

	if err := (coredata.Attachments{}).DeleteExcept(
		ctx,
		conn,
		scope,
		organizationID,
		parentID,
		ids,
	); err != nil {
		return "", fmt.Errorf("cannot delete attachments: %w", err)
	}

	return bound, nil
}

// FileIDs returns the file ids referenced by content.
func FileIDs(content string) ([]gid.GID, error) {
	if strings.TrimSpace(content) == "" {
		return nil, nil
	}

	node, err := prosemirror.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("cannot parse attachments: %w", err)
	}

	ids, err := fileIDs(node)
	if err != nil {
		return nil, fmt.Errorf("cannot read attachments: %w", err)
	}

	return ids, nil
}

// Copy duplicates attachments from one parent onto another for the given file ids.
// A file can stay attached to both parents.
func Copy(
	ctx context.Context,
	conn pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	fromParentID gid.GID,
	toParentID gid.GID,
	fileIDs []gid.GID,
) error {
	if len(fileIDs) == 0 || fromParentID == toParentID {
		return nil
	}

	var source coredata.Attachments
	if err := source.LoadByParentFileIDs(
		ctx,
		conn,
		scope,
		organizationID,
		fromParentID,
		fileIDs,
	); err != nil {
		return fmt.Errorf("cannot load attachments: %w", err)
	}

	var destination coredata.Attachments
	if err := destination.LoadByParentFileIDs(
		ctx,
		conn,
		scope,
		organizationID,
		toParentID,
		fileIDs,
	); err != nil {
		return fmt.Errorf("cannot load attachments: %w", err)
	}

	present := make(map[gid.GID]struct{}, len(destination))
	for _, row := range destination {
		present[row.FileID] = struct{}{}
	}

	for _, row := range source {
		if _, ok := present[row.FileID]; ok {
			continue
		}

		next := *row
		next.ParentID = toParentID
		next.OrganizationID = organizationID

		if err := next.Insert(ctx, conn, scope); err != nil {
			return fmt.Errorf("cannot copy attachment: %w", err)
		}
	}

	return nil
}

func fileIDs(node prosemirror.Node) ([]gid.GID, error) {
	refs := prosemirror.ReferencedFiles(node)
	ids := make([]gid.GID, 0, len(refs))
	seen := make(map[string]struct{}, len(refs))

	for _, ref := range refs {
		if _, ok := seen[ref.ID]; ok {
			continue
		}

		seen[ref.ID] = struct{}{}

		id, err := gid.ParseGID(ref.ID)
		if err != nil {
			return nil, fmt.Errorf("cannot parse attachment id: %w", err)
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func attachmentFor(
	rows []*coredata.Attachment,
	organizationID gid.GID,
	parentID gid.GID,
) *coredata.Attachment {
	for _, row := range rows {
		if row.OrganizationID == organizationID && row.ParentID == parentID {
			return row
		}
	}

	return nil
}
