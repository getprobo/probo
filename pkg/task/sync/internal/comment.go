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

package internal

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/hash"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/webhook"
	webhooktypes "go.probo.inc/probo/pkg/webhook/types"
)

func CommentContentHash(markdown string) string {
	return hash.SHA256HexString(markdown)
}

func DeleteSyncedComment(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	mapping *coredata.TaskCommentExternalLink,
) error {
	taskComment := &coredata.TaskComment{}
	if err := taskComment.LoadByID(ctx, tx, scope, mapping.TaskCommentID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			if err := mapping.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete task comment external link: %w", err)
			}

			if _, err := coredata.DeleteCommentedActivity(ctx, tx, scope, mapping.TaskCommentID); err != nil {
				return fmt.Errorf("cannot delete missing comment activity: %w", err)
			}

			return nil
		}

		return fmt.Errorf("cannot load task comment: %w", err)
	}

	if err := webhook.InsertData(
		ctx,
		tx,
		scope,
		taskComment.OrganizationID,
		coredata.WebhookEventTypeTaskCommentDeleted,
		webhooktypes.NewTaskComment(taskComment),
	); err != nil {
		return fmt.Errorf("cannot emit task comment deleted webhook: %w", err)
	}

	if err := taskComment.Delete(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot delete task comment: %w", err)
	}

	if _, err := coredata.DeleteCommentedActivity(ctx, tx, scope, taskComment.ID); err != nil {
		return fmt.Errorf("cannot delete synced comment activity: %w", err)
	}

	return nil
}

func UpdateImportedComment(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	mapping *coredata.TaskCommentExternalLink,
	content string,
	hash string,
	remoteUpdatedAt *time.Time,
	actorProfileID *gid.GID,
) error {
	taskComment := &coredata.TaskComment{}
	if err := taskComment.LoadByID(ctx, tx, scope, mapping.TaskCommentID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil
		}

		return fmt.Errorf("cannot load task comment: %w", err)
	}

	previous := *taskComment
	now := time.Now()
	taskComment.Content = content
	taskComment.UpdatedAt = now

	if err := taskComment.Update(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot update task comment: %w", err)
	}

	mapping.ContentHash = &hash
	mapping.RemoteUpdatedAt = remoteUpdatedAt
	mapping.UpdatedAt = now

	if err := mapping.Update(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot update task comment external link: %w", err)
	}

	if err := webhook.InsertUpdateData(
		ctx,
		tx,
		scope,
		taskComment.OrganizationID,
		coredata.WebhookEventTypeTaskCommentUpdated,
		webhooktypes.NewTaskComment(taskComment),
		webhooktypes.NewTaskComment(&previous),
	); err != nil {
		return fmt.Errorf("cannot emit task comment updated webhook: %w", err)
	}

	if err := NotifyCommentMentions(
		ctx,
		tx,
		scope,
		taskComment.TaskID,
		previous.Content,
		taskComment.Content,
		actorProfileID,
	); err != nil {
		return fmt.Errorf("cannot notify synced comment mentions: %w", err)
	}

	return nil
}

func DeleteCommentLinks(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	taskID gid.GID,
) error {
	var links coredata.TaskCommentExternalLinks
	if err := links.DeleteByTaskID(ctx, conn, scope, taskID); err != nil {
		return fmt.Errorf("cannot delete task comment external links: %w", err)
	}

	return nil
}

func ResolveCommentOwner(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	link *coredata.TaskExternalLink,
	email string,
) (*gid.GID, error) {
	email = strings.TrimSpace(email)
	if email != "" {
		addr, err := mail.ParseAddr(email)
		if err == nil {
			profile := &coredata.MembershipProfile{}

			err = profile.LoadByOrganizationIDAndEmail(ctx, conn, scope, link.OrganizationID, addr)
			if err == nil {
				return &profile.ID, nil
			}

			if !errors.Is(err, coredata.ErrResourceNotFound) {
				return nil, fmt.Errorf("cannot load comment owner: %w", err)
			}
		}
	}

	return nil, nil
}

func CommentAuthorName(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	ownerID *gid.GID,
) (string, error) {
	if ownerID == nil {
		return "", nil
	}

	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByID(ctx, conn, scope, *ownerID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return "", nil
		}

		return "", fmt.Errorf("cannot load comment author profile: %w", err)
	}

	return strings.TrimSpace(profile.FullName), nil
}
