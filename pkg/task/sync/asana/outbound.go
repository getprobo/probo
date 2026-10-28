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

package asana

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	client "go.probo.inc/probo/pkg/task/sync/asana/client"
	"go.probo.inc/probo/pkg/task/sync/shared"
)

func (h *outboundHandler) handleAsana(ctx context.Context, item *coredata.TaskSyncJob, payload JobPayload) error {
	scope := coredata.NewScopeFromObjectID(item.ID)

	account, err := h.svc.asanaClientByConnectorID(ctx, scope, payload.ConnectorID)
	if err != nil {
		return fmt.Errorf("cannot load Asana connector: %w", err)
	}

	// A failed registration is only logged. The next outbound job retries it.
	if payload.TeamID != "" {
		h.svc.ensureAsanaWebhook(ctx, scope, payload.ConnectorID, payload.TeamID)
	}

	switch payload.Action {
	case shared.SyncActionCommentUpsert:
		if err := h.processAsanaCommentUpsert(ctx, item, payload, scope, account); err != nil {
			return fmt.Errorf("cannot sync Asana comment: %w", err)
		}
	case shared.SyncActionCommentDelete:
		if err := h.processAsanaCommentDelete(ctx, item, payload, scope, account); err != nil {
			return fmt.Errorf("cannot delete Asana comment: %w", err)
		}
	case shared.SyncActionCancel:
		if err := h.processAsanaCancel(ctx, item, payload, scope, account); err != nil {
			return fmt.Errorf("cannot cancel Asana task: %w", err)
		}
	case shared.SyncActionUpdate:
		if err := h.processAsanaUpdate(ctx, item, payload, scope, account); err != nil {
			return fmt.Errorf("cannot update Asana task: %w", err)
		}
	default:
		return fmt.Errorf("cannot process Asana sync job: unknown action %q", payload.Action)
	}

	return nil
}

func (h *outboundHandler) processAsanaUpdate(
	ctx context.Context,
	item *coredata.TaskSyncJob,
	payload JobPayload,
	scope coredata.Scoper,
	account asanaAccount,
) error {
	skip, err := h.newer(ctx, item, payload.TaskID, scope)
	if err != nil {
		return fmt.Errorf("cannot check newer Asana sync job: %w", err)
	}

	if skip {
		return h.done(ctx, item)
	}

	task := &coredata.Task{}

	var link *coredata.TaskExternalLink

	err = h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := task.LoadByID(ctx, conn, scope, payload.TaskID); err != nil {
				return fmt.Errorf("cannot load task: %w", err)
			}

			loaded, err := shared.LoadCurrentOutboundLink(ctx, conn, scope, payload.TaskID)
			if err != nil {
				return fmt.Errorf("cannot load task external link: %w", err)
			}

			link = loaded

			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot load task for Asana sync: %w", err)
	}

	if link == nil || link.Provider != coredata.ConnectorProviderAsana || link.ExternalID != payload.ExternalID {
		return h.done(ctx, item)
	}

	projectGID := payload.TeamID
	mapping := projectMapping(account.settings, projectGID)

	htmlNotes, err := h.svc.asanaHTMLNotes(ctx, scope, account, task.OrganizationID, task.Content)
	if err != nil {
		return fmt.Errorf("cannot render Asana description: %w", err)
	}

	completed := CompletedForState(task.State)
	assigneeGID, assigneeOK := h.svc.asanaAssigneeGID(ctx, scope, account, task.AssignedToID)

	update := client.TaskUpdate{
		Name:         &task.Name,
		HTMLNotesSet: true,
		HTMLNotes:    &htmlNotes,
		Completed:    &completed,
		DueOnSet:     true,
		DueOn:        shared.DeadlineToDate(task.Deadline),
		AssigneeSet:  assigneeOK,
		AssigneeGID:  assigneeGID,
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

	meta, err := readAsanaFiles(link.Metadata)
	if err != nil {
		return fmt.Errorf("cannot read Asana files: %w", err)
	}

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
		return fmt.Errorf("cannot push Asana files: %w", err)
	}

	meta.Files = files

	encoded, err := writeAsanaFiles(meta)
	if err != nil {
		return fmt.Errorf("cannot encode Asana files: %w", err)
	}

	if err := h.storeAsanaFileMeta(ctx, scope, payload, encoded); err != nil {
		return fmt.Errorf("cannot store Asana files: %w", err)
	}

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

	if err := h.storeAsanaSnapshotHash(ctx, scope, payload, encoded, hash, remote.ModifiedAt); err != nil {
		return fmt.Errorf("cannot store Asana sync hash: %w", err)
	}

	return h.done(ctx, item)
}

func (h *outboundHandler) processAsanaCancel(
	ctx context.Context,
	item *coredata.TaskSyncJob,
	payload JobPayload,
	scope coredata.Scoper,
	account asanaAccount,
) error {
	skip, err := h.newer(ctx, item, payload.TaskID, scope)
	if err != nil {
		return fmt.Errorf("cannot check newer Asana sync job: %w", err)
	}

	if skip {
		return h.done(ctx, item)
	}

	completed := true
	if _, err := account.client.UpdateTask(ctx, payload.ExternalID, client.TaskUpdate{Completed: &completed}); err != nil {
		if client.IsNotFound(err) {
			return h.done(ctx, item)
		}

		return fmt.Errorf("cannot complete Asana task: %w", err)
	}

	mapping := projectMapping(account.settings, payload.TeamID)
	if section := OutboundSection(coredata.TaskStateCanceled, mapping); section != "" {
		if err := account.client.AddTaskToSection(ctx, section, payload.ExternalID); err != nil && !client.IsNotFound(err) {
			return fmt.Errorf("cannot move canceled Asana task: %w", err)
		}
	}

	remote, err := account.client.GetTask(ctx, payload.ExternalID)
	if err != nil {
		if client.IsNotFound(err) {
			return h.done(ctx, item)
		}

		return fmt.Errorf("cannot reload Asana task: %w", err)
	}

	if err := h.storeAsanaSnapshot(ctx, payload, scope, account, remote); err != nil {
		return fmt.Errorf("cannot store Asana sync hash: %w", err)
	}

	return h.done(ctx, item)
}

func (h *outboundHandler) storeAsanaSnapshot(
	ctx context.Context,
	payload JobPayload,
	scope coredata.Scoper,
	account asanaAccount,
	remote client.Task,
) error {
	projectGID := payload.TeamID
	mapping := projectMapping(account.settings, projectGID)
	sectionGID := sectionForProject(remote, projectGID)

	optionGID := ""
	if mapping != nil {
		optionGID = enumOption(remote, mapping.PriorityFieldGID)
	}

	hash := asanaSnapshotHash(remote, sectionGID, optionGID)

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			current, err := shared.LoadCurrentOutboundLink(ctx, tx, scope, payload.TaskID)
			if err != nil {
				return fmt.Errorf("cannot load task external link: %w", err)
			}

			if current == nil || current.Provider != coredata.ConnectorProviderAsana || current.ExternalID != payload.ExternalID {
				return nil
			}

			current.ContentHash = &hash
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
}

func (h *outboundHandler) storeAsanaFileMeta(
	ctx context.Context,
	scope coredata.Scoper,
	payload JobPayload,
	metadata []byte,
) error {
	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			current, err := shared.LoadCurrentOutboundLink(ctx, tx, scope, payload.TaskID)
			if err != nil {
				return fmt.Errorf("cannot load task external link: %w", err)
			}

			if current == nil || current.Provider != coredata.ConnectorProviderAsana || current.ExternalID != payload.ExternalID {
				return nil
			}

			current.Metadata = metadata

			current.UpdatedAt = time.Now()
			if err := current.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update Asana files: %w", err)
			}

			return nil
		},
	)
}

func (h *outboundHandler) storeAsanaSnapshotHash(
	ctx context.Context,
	scope coredata.Scoper,
	payload JobPayload,
	metadata []byte,
	hash string,
	remoteUpdatedAt time.Time,
) error {
	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			current, err := shared.LoadCurrentOutboundLink(ctx, tx, scope, payload.TaskID)
			if err != nil {
				return fmt.Errorf("cannot load task external link: %w", err)
			}

			if current == nil || current.Provider != coredata.ConnectorProviderAsana || current.ExternalID != payload.ExternalID {
				return nil
			}

			current.ContentHash = new(hash)
			current.Metadata = metadata

			current.UpdatedAt = time.Now()
			if !remoteUpdatedAt.IsZero() {
				current.RemoteUpdatedAt = new(remoteUpdatedAt)
			}

			if err := current.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update Asana sync hash: %w", err)
			}

			return nil
		},
	)
}

func (h *outboundHandler) processAsanaCommentUpsert(
	ctx context.Context,
	item *coredata.TaskSyncJob,
	payload JobPayload,
	scope coredata.Scoper,
	account asanaAccount,
) error {
	comment, mapping, skip, err := shared.LoadCommentUpsert(ctx, h.pg, item, payload, scope, false)
	if err != nil {
		return fmt.Errorf("cannot load Asana comment: %w", err)
	}

	if skip {
		return h.done(ctx, item)
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
		return fmt.Errorf("cannot render Asana comment: %w", err)
	}

	if htmlText == "" {
		return h.done(ctx, item)
	}

	previousGID := ""
	if mapping != nil {
		previousGID = mapping.ExternalID
	}

	story, err := upsertAsanaStory(ctx, account, payload.ExternalID, previousGID, htmlText)
	if err != nil {
		return fmt.Errorf("cannot upsert Asana comment: %w", err)
	}

	err = h.saveAsanaCommentStory(ctx, scope, comment, payload, mapping, story)
	if err != nil {
		if story.GID != "" && story.GID != previousGID {
			return deleteCreatedAsanaStory(ctx, account, story.GID, err)
		}

		return fmt.Errorf("cannot save Asana comment link: %w", err)
	}

	meta, err := readAsanaFiles(nil)
	if err != nil {
		return fmt.Errorf("cannot read Asana comment files: %w", err)
	}

	err = h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			taskLink := &coredata.TaskExternalLink{}
			if err := taskLink.LoadByTaskID(ctx, conn, scope, comment.TaskID); err != nil {
				return fmt.Errorf("cannot load task external link: %w", err)
			}

			loaded, err := readAsanaFiles(taskLink.Metadata)
			if err != nil {
				return fmt.Errorf("cannot read Asana file metadata: %w", err)
			}

			meta = loaded

			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot load task external link: %w", err)
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
		return fmt.Errorf("cannot push Asana comment files: %w", err)
	}

	meta.CommentFiles[story.GID] = pushed

	encoded, err := writeAsanaFiles(meta)
	if err != nil {
		return fmt.Errorf("cannot encode Asana comment files: %w", err)
	}

	now := time.Now()

	err = h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
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
		return fmt.Errorf("cannot store Asana comment files: %w", err)
	}

	return h.done(ctx, item)
}

func (h *outboundHandler) processAsanaCommentDelete(
	ctx context.Context,
	item *coredata.TaskSyncJob,
	payload JobPayload,
	scope coredata.Scoper,
	account asanaAccount,
) error {
	if payload.ExternalCommentID != "" {
		if err := account.client.DeleteStory(ctx, payload.ExternalCommentID); err != nil && !client.IsNotFound(err) {
			return fmt.Errorf("cannot delete Asana comment: %w", err)
		}

		if err := h.removeAsanaCommentFiles(ctx, scope, payload); err != nil {
			return fmt.Errorf("cannot store Asana comment files: %w", err)
		}
	}

	return h.done(ctx, item)
}

func (h *outboundHandler) removeAsanaCommentFiles(
	ctx context.Context,
	scope coredata.Scoper,
	payload JobPayload,
) error {
	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			current, err := shared.LoadCurrentOutboundLink(ctx, tx, scope, payload.TaskID)
			if err != nil {
				return fmt.Errorf("cannot load task external link: %w", err)
			}

			if current == nil || current.Provider != coredata.ConnectorProviderAsana || current.ExternalID != payload.ExternalID {
				return nil
			}

			meta, err := readAsanaFiles(current.Metadata)
			if err != nil {
				return fmt.Errorf("cannot read Asana comment files: %w", err)
			}

			if _, ok := meta.CommentFiles[payload.ExternalCommentID]; !ok {
				return nil
			}

			delete(meta.CommentFiles, payload.ExternalCommentID)

			encoded, err := writeAsanaFiles(meta)
			if err != nil {
				return fmt.Errorf("cannot encode Asana comment files: %w", err)
			}

			current.Metadata = encoded

			current.UpdatedAt = time.Now()
			if err := current.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update Asana comment files: %w", err)
			}

			return nil
		},
	)
}

func (h *outboundHandler) saveAsanaCommentStory(
	ctx context.Context,
	scope coredata.Scoper,
	comment *coredata.TaskComment,
	payload JobPayload,
	mapping *coredata.TaskCommentExternalLink,
	story client.Story,
) error {
	now := time.Now()
	hash := shared.CommentContentHash(story.HTMLText)

	createdAt := now
	if mapping != nil && !mapping.CreatedAt.IsZero() {
		createdAt = mapping.CreatedAt
	}

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

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if mapping == nil || mapping.ExternalID == "" {
				if err := link.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert Asana comment link: %w", err)
				}

				return nil
			}

			if err := link.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update Asana comment link: %w", err)
			}

			return nil
		},
	)
}

func deleteCreatedAsanaStory(
	ctx context.Context,
	account asanaAccount,
	storyGID string,
	cause error,
) error {
	err := account.client.DeleteStory(ctx, storyGID)
	if err == nil || client.IsNotFound(err) {
		return fmt.Errorf("cannot save Asana comment link: %w", cause)
	}

	return fmt.Errorf(
		"cannot save Asana comment link: %w",
		errors.Join(
			cause,
			fmt.Errorf("cannot delete Asana comment: %w", err),
		),
	)
}

func upsertAsanaStory(
	ctx context.Context,
	account asanaAccount,
	taskGID string,
	storyGID string,
	htmlText string,
) (client.Story, error) {
	if storyGID == "" {
		story, err := account.client.CreateStory(ctx, taskGID, htmlText)
		if err != nil {
			return client.Story{}, fmt.Errorf("cannot create Asana comment: %w", err)
		}

		return story, nil
	}

	story, err := account.client.UpdateStory(ctx, storyGID, htmlText)
	if err == nil {
		return story, nil
	}

	if !client.IsNotFound(err) {
		return client.Story{}, fmt.Errorf("cannot update Asana comment: %w", err)
	}

	story, err = account.client.CreateStory(ctx, taskGID, htmlText)
	if err != nil {
		return client.Story{}, fmt.Errorf("cannot recreate Asana comment: %w", err)
	}

	return story, nil
}
