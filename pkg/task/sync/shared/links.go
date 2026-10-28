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

package shared

import (
	"context"
	"errors"
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func LoadCurrentOutboundLink(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	taskID gid.GID,
) (*coredata.TaskExternalLink, error) {
	link := &coredata.TaskExternalLink{}
	if err := link.LoadByTaskID(ctx, conn, scope, taskID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot load task external link: %w", err)
	}

	return link, nil
}

func OutboundLinkMatchesJob(link *coredata.TaskExternalLink, payload JobPayload) bool {
	return link != nil && link.ExternalID == payload.ExternalID
}

func CommentLinksByCommentID(
	ctx context.Context,
	db *pg.Client,
	scope coredata.Scoper,
	taskID gid.GID,
) (map[gid.GID]*coredata.TaskCommentExternalLink, error) {
	links, err := loadCommentLinks(ctx, db, scope, taskID)
	if err != nil {
		return nil, err
	}

	byID := make(map[gid.GID]*coredata.TaskCommentExternalLink, len(links))
	for _, link := range links {
		byID[link.TaskCommentID] = link
	}

	return byID, nil
}

func CommentLinksByExternalID(
	ctx context.Context,
	db *pg.Client,
	scope coredata.Scoper,
	taskID gid.GID,
) (map[string]*coredata.TaskCommentExternalLink, error) {
	links, err := loadCommentLinks(ctx, db, scope, taskID)
	if err != nil {
		return nil, err
	}

	byID := make(map[string]*coredata.TaskCommentExternalLink, len(links))
	for _, link := range links {
		byID[link.ExternalID] = link
	}

	return byID, nil
}

func LoadCommentUpsert(
	ctx context.Context,
	db *pg.Client,
	item *coredata.TaskSyncJob,
	payload JobPayload,
	scope coredata.Scoper,
	sent bool,
) (*coredata.TaskComment, *coredata.TaskCommentExternalLink, bool, error) {
	var (
		comment *coredata.TaskComment
		mapping *coredata.TaskCommentExternalLink
		skip    bool
	)

	err := db.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			comment = &coredata.TaskComment{}
			if err := comment.LoadByID(ctx, tx, scope, payload.CommentID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					skip = true

					return nil
				}

				return fmt.Errorf("cannot load task comment: %w", err)
			}

			link, err := LoadCurrentOutboundLink(ctx, tx, scope, payload.TaskID)
			if err != nil {
				return err
			}

			if !OutboundLinkMatchesJob(link, payload) {
				skip = true

				return nil
			}

			if !sent {
				newer, err := item.HasNewerOutboundCommentJob(
					ctx,
					tx,
					scope,
					payload.TaskID,
					payload.CommentID,
					CommentSyncActions(),
				)
				if err != nil {
					return err
				}

				if newer {
					skip = true

					return nil
				}
			}

			existing := &coredata.TaskCommentExternalLink{}

			err = existing.LoadByCommentID(ctx, tx, scope, comment.ID)
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return nil
			}

			if err != nil {
				return fmt.Errorf("cannot load task comment external link: %w", err)
			}

			mapping = existing

			return nil
		},
	)
	if err != nil {
		return nil, nil, false, err
	}

	return comment, mapping, skip, nil
}

func loadCommentLinks(
	ctx context.Context,
	db *pg.Client,
	scope coredata.Scoper,
	taskID gid.GID,
) (coredata.TaskCommentExternalLinks, error) {
	var links coredata.TaskCommentExternalLinks

	err := db.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			loaded, err := page.LoadAll(
				ctx,
				page.OrderBy[coredata.TaskCommentExternalLinkOrderField]{
					Field:     coredata.TaskCommentExternalLinkOrderFieldTaskCommentID,
					Direction: page.OrderDirectionAsc,
				},
				func(
					ctx context.Context,
					cursor *page.Cursor[coredata.TaskCommentExternalLinkOrderField],
				) ([]*coredata.TaskCommentExternalLink, error) {
					var batch coredata.TaskCommentExternalLinks
					if err := batch.LoadByTaskID(ctx, conn, scope, taskID, cursor); err != nil {
						return nil, err
					}

					return batch, nil
				},
			)
			if err != nil {
				return fmt.Errorf("cannot load task comment external links: %w", err)
			}

			links = loaded

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return links, nil
}
