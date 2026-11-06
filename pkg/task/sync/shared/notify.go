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
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/notification"
	"go.probo.inc/probo/pkg/prosemirror"
)

func NotifyMentionedProfiles(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	taskID gid.GID,
	taskName string,
	oldContent string,
	newContent string,
	actorProfileID *gid.GID,
) error {
	if oldContent == newContent {
		return nil
	}

	mentionedIDs, err := prosemirror.NewlyMentionedProfileIDs(oldContent, newContent)
	if err != nil {
		return fmt.Errorf("cannot collect new mentions: %w", err)
	}

	if len(mentionedIDs) == 0 {
		return nil
	}

	if err := notification.InsertTaskMentions(
		ctx,
		tx,
		scope,
		organizationID,
		taskID,
		taskName,
		mentionedIDs,
		actorProfileID,
	); err != nil {
		return fmt.Errorf("cannot insert task mention notifications: %w", err)
	}

	return nil
}

func NotifyCommentMentions(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	taskID gid.GID,
	oldContent string,
	newContent string,
	actorProfileID *gid.GID,
) error {
	task := &coredata.Task{}
	if err := task.LoadByID(ctx, tx, scope, taskID); err != nil {
		return fmt.Errorf("cannot load task for mention notification: %w", err)
	}

	if err := NotifyMentionedProfiles(
		ctx,
		tx,
		scope,
		task.OrganizationID,
		task.ID,
		task.Name,
		oldContent,
		newContent,
		actorProfileID,
	); err != nil {
		return fmt.Errorf("cannot notify synced mentions: %w", err)
	}

	return nil
}
