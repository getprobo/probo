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

package tasksync

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/task/sync/asana"
)

func (h *outboundHandler) handleAsana(ctx context.Context, item *coredata.TaskSyncJob, payload JobPayload) error {
	scope := coredata.NewScopeFromObjectID(item.ID)

	account, err := h.svc.asanaClientByConnectorID(ctx, scope, payload.ConnectorID)
	if err != nil {
		return err
	}

	switch payload.Action {
	case SyncActionCommentUpsert:
		return h.processAsanaCommentUpsert(ctx, item, payload, scope, account)
	case SyncActionCommentDelete:
		return h.processAsanaCommentDelete(ctx, item, payload, account)
	case SyncActionCancel:
		return h.processAsanaCancel(ctx, item, payload, account)
	case SyncActionUpdate:
		return h.processAsanaUpdate(ctx, item, payload, scope, account)
	default:
		return fmt.Errorf("cannot process Asana sync job: unknown action %q", payload.Action)
	}
}

func (h *outboundHandler) processAsanaUpdate(
	ctx context.Context,
	item *coredata.TaskSyncJob,
	payload JobPayload,
	scope coredata.Scoper,
	account asanaAccount,
) error {
	skip, err := h.hasNewerOutboundJob(ctx, item, payload.TaskID, scope)
	if err != nil {
		return err
	}

	if skip {
		return h.succeed(ctx, item)
	}

	task := &coredata.Task{}
	link := &coredata.TaskExternalLink{}

	err = h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := task.LoadByID(ctx, conn, scope, payload.TaskID); err != nil {
				return fmt.Errorf("cannot load task: %w", err)
			}

			if err := link.LoadByTaskID(ctx, conn, scope, payload.TaskID); err != nil {
				return fmt.Errorf("cannot load task external link: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	if link.Provider != coredata.ConnectorProviderAsana || link.ExternalID != payload.ExternalID {
		return h.succeed(ctx, item)
	}

	projectGID := payload.TeamID
	mapping := projectMapping(account.settings, projectGID)

	htmlNotes, err := h.svc.asanaHTMLNotes(ctx, scope, account, task.OrganizationID, task.Content)
	if err != nil {
		return err
	}

	completed := CompletedForState(task.State)

	update := asana.TaskUpdate{
		Name:         &task.Name,
		HTMLNotesSet: true,
		HTMLNotes:    &htmlNotes,
		Completed:    &completed,
		DueOnSet:     true,
		DueOn:        DeadlineToLinearDate(task.Deadline),
		AssigneeSet:  true,
		AssigneeGID:  h.svc.asanaAssigneeGID(ctx, scope, account, task.AssignedToID),
	}
	if option, ok := OutboundPriorityOption(task.Priority, mapping); ok {
		update.CustomFields = map[string]*string{mapping.PriorityFieldGID: &option}
	}

	if _, err := account.client.UpdateTask(ctx, link.ExternalID, update); err != nil {
		return fmt.Errorf("cannot update Asana task: %w", err)
	}

	if section := OutboundSection(task.State, mapping); section != "" {
		if err := account.client.AddTaskToSection(ctx, section, link.ExternalID); err != nil {
			return fmt.Errorf("cannot move Asana task: %w", err)
		}
	}

	meta := readAsanaFiles(link.Metadata)

	files, err := h.svc.pushAsanaFiles(
		ctx,
		scope,
		account,
		task.OrganizationID,
		link.ExternalID,
		task.Content,
		meta.Files,
	)
	if err != nil {
		return err
	}

	meta.Files = files

	remote, err := account.client.GetTask(ctx, link.ExternalID)
	if err != nil {
		return fmt.Errorf("cannot reload Asana task: %w", err)
	}

	sectionGID := sectionForProject(remote, projectGID)

	optionGID := ""
	if mapping != nil {
		optionGID = enumOption(remote, mapping.PriorityFieldGID)
	}

	hash := asanaSnapshotHash(remote, sectionGID, optionGID)

	encoded, err := writeAsanaFiles(meta)
	if err != nil {
		return err
	}

	err = h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			current := &coredata.TaskExternalLink{}
			if err := current.LoadByTaskID(ctx, tx, scope, payload.TaskID); err != nil {
				return fmt.Errorf("cannot load task external link: %w", err)
			}

			current.ContentHash = &hash
			current.Metadata = encoded
			now := time.Now()

			current.UpdatedAt = now
			if !remote.ModifiedAt.IsZero() {
				current.RemoteUpdatedAt = &remote.ModifiedAt
			}

			if err := current.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot store Asana sync hash: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	return h.succeed(ctx, item)
}

func (h *outboundHandler) processAsanaCancel(
	ctx context.Context,
	item *coredata.TaskSyncJob,
	payload JobPayload,
	account asanaAccount,
) error {
	completed := true
	if _, err := account.client.UpdateTask(ctx, payload.ExternalID, asana.TaskUpdate{Completed: &completed}); err != nil {
		if asana.IsNotFound(err) {
			return h.succeed(ctx, item)
		}

		return fmt.Errorf("cannot complete Asana task: %w", err)
	}

	mapping := projectMapping(account.settings, payload.TeamID)
	if section := OutboundSection(coredata.TaskStateCanceled, mapping); section != "" {
		if err := account.client.AddTaskToSection(ctx, section, payload.ExternalID); err != nil && !asana.IsNotFound(err) {
			return fmt.Errorf("cannot move canceled Asana task: %w", err)
		}
	}

	return h.succeed(ctx, item)
}

func (h *outboundHandler) processAsanaCommentUpsert(
	ctx context.Context,
	item *coredata.TaskSyncJob,
	payload JobPayload,
	scope coredata.Scoper,
	account asanaAccount,
) error {
	comment := &coredata.TaskComment{}
	mapping := &coredata.TaskCommentExternalLink{}

	err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := comment.LoadByID(ctx, conn, scope, payload.CommentID); err != nil {
				return fmt.Errorf("cannot load task comment: %w", err)
			}

			err := mapping.LoadByCommentID(ctx, conn, scope, payload.CommentID)
			if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
				return fmt.Errorf("cannot load task comment external link: %w", err)
			}

			if errors.Is(err, coredata.ErrResourceNotFound) {
				mapping.ExternalID = ""
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	htmlText, err := h.svc.asanaCommentHTML(
		ctx,
		scope,
		account,
		comment.OrganizationID,
		h.svc.asanaAuthorName(ctx, scope, comment.OwnerID),
		comment.Content,
	)
	if err != nil {
		return err
	}

	if htmlText == "" {
		return h.succeed(ctx, item)
	}

	story, err := upsertAsanaStory(ctx, account, payload.ExternalID, mapping.ExternalID, htmlText)
	if err != nil {
		return err
	}

	meta := readAsanaFiles(nil)

	err = h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			taskLink := &coredata.TaskExternalLink{}
			if err := taskLink.LoadByTaskID(ctx, conn, scope, comment.TaskID); err != nil {
				return fmt.Errorf("cannot load task external link: %w", err)
			}

			meta = readAsanaFiles(taskLink.Metadata)

			return nil
		},
	)
	if err != nil {
		return err
	}

	if meta.CommentFiles == nil {
		meta.CommentFiles = map[string]map[string]string{}
	}

	pushed, err := h.svc.pushAsanaFiles(
		ctx,
		scope,
		account,
		comment.OrganizationID,
		story.GID,
		comment.Content,
		meta.CommentFiles[story.GID],
	)
	if err != nil {
		return err
	}

	meta.CommentFiles[story.GID] = pushed

	encoded, err := writeAsanaFiles(meta)
	if err != nil {
		return err
	}

	hash := CommentContentHash(story.HTMLText)
	now := time.Now()

	createdAt := mapping.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}

	err = h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			link := &coredata.TaskCommentExternalLink{
				TaskCommentID:  comment.ID,
				OrganizationID: comment.OrganizationID,
				TaskID:         comment.TaskID,
				ConnectorID:    payload.ConnectorID,
				Provider:       coredata.ConnectorProviderAsana,
				ExternalID:     story.GID,
				ContentHash:    &hash,
				CreatedAt:      createdAt,
				UpdatedAt:      now,
			}
			if !story.CreatedAt.IsZero() {
				link.RemoteUpdatedAt = &story.CreatedAt
			}

			if mapping.ExternalID == "" {
				if err := link.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert Asana comment link: %w", err)
				}
			} else if err := link.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update Asana comment link: %w", err)
			}

			taskLink := &coredata.TaskExternalLink{}
			if err := taskLink.LoadByTaskID(ctx, tx, scope, comment.TaskID); err != nil {
				return fmt.Errorf("cannot load task external link: %w", err)
			}

			taskLink.Metadata = encoded

			taskLink.UpdatedAt = now
			if err := taskLink.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot store Asana comment files: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	return h.succeed(ctx, item)
}

func (h *outboundHandler) processAsanaCommentDelete(
	ctx context.Context,
	item *coredata.TaskSyncJob,
	payload JobPayload,
	account asanaAccount,
) error {
	if payload.ExternalCommentID != "" {
		if err := account.client.DeleteStory(ctx, payload.ExternalCommentID); err != nil && !asana.IsNotFound(err) {
			return fmt.Errorf("cannot delete Asana comment: %w", err)
		}
	}

	return h.succeed(ctx, item)
}

func upsertAsanaStory(
	ctx context.Context,
	account asanaAccount,
	taskGID string,
	storyGID string,
	htmlText string,
) (asana.Story, error) {
	if storyGID == "" {
		story, err := account.client.CreateStory(ctx, taskGID, htmlText)
		if err != nil {
			return asana.Story{}, fmt.Errorf("cannot create Asana comment: %w", err)
		}

		return story, nil
	}

	story, err := account.client.UpdateStory(ctx, storyGID, htmlText)
	if err == nil {
		return story, nil
	}

	if !asana.IsNotFound(err) {
		return asana.Story{}, fmt.Errorf("cannot update Asana comment: %w", err)
	}

	story, err = account.client.CreateStory(ctx, taskGID, htmlText)
	if err != nil {
		return asana.Story{}, fmt.Errorf("cannot recreate Asana comment: %w", err)
	}

	return story, nil
}
