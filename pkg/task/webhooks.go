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

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/webhook"
	webhooktypes "go.probo.inc/probo/pkg/webhook/types"
)

func emitTaskCreated(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	task *coredata.Task,
) error {
	if err := webhook.InsertData(
		ctx,
		tx,
		scope,
		task.OrganizationID,
		coredata.WebhookEventTypeTaskCreated,
		webhooktypes.NewTask(task),
	); err != nil {
		return fmt.Errorf("cannot insert task created webhook event: %w", err)
	}

	return nil
}

func emitTaskUpdated(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	previous *coredata.Task,
	task *coredata.Task,
) error {
	return webhook.InsertTaskUpdated(ctx, tx, scope, previous, task)
}

func emitTaskDeleted(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	task *coredata.Task,
) error {
	if err := webhook.InsertData(
		ctx,
		tx,
		scope,
		task.OrganizationID,
		coredata.WebhookEventTypeTaskDeleted,
		webhooktypes.NewTask(task),
	); err != nil {
		return fmt.Errorf("cannot insert task deleted webhook event: %w", err)
	}

	return nil
}

func emitTaskCommentCreated(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	comment *coredata.TaskComment,
) error {
	if err := webhook.InsertData(
		ctx,
		tx,
		scope,
		comment.OrganizationID,
		coredata.WebhookEventTypeTaskCommentCreated,
		webhooktypes.NewTaskComment(comment),
	); err != nil {
		return fmt.Errorf("cannot insert task comment created webhook event: %w", err)
	}

	return nil
}

func emitTaskCommentUpdated(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	previous *coredata.TaskComment,
	comment *coredata.TaskComment,
) error {
	if err := webhook.InsertUpdateData(
		ctx,
		tx,
		scope,
		comment.OrganizationID,
		coredata.WebhookEventTypeTaskCommentUpdated,
		webhooktypes.NewTaskComment(comment),
		webhooktypes.NewTaskComment(previous),
	); err != nil {
		return fmt.Errorf("cannot insert task comment updated webhook event: %w", err)
	}

	return nil
}

func emitTaskCommentDeleted(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	comment *coredata.TaskComment,
) error {
	if err := webhook.InsertData(
		ctx,
		tx,
		scope,
		comment.OrganizationID,
		coredata.WebhookEventTypeTaskCommentDeleted,
		webhooktypes.NewTaskComment(comment),
	); err != nil {
		return fmt.Errorf("cannot insert task comment deleted webhook event: %w", err)
	}

	return nil
}
