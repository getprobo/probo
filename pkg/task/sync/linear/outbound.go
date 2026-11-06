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

package linear

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
	linear "go.probo.inc/probo/pkg/task/sync/linear/client"
	"go.probo.inc/probo/pkg/task/sync/shared"
)

const maxOutboundRevisionAttempts = 5

type (
	outboundHandler struct {
		svc   *Service
		pg    *pg.Client
		newer func(context.Context, *coredata.TaskSyncJob, gid.GID, coredata.Scoper) (bool, error)
		done  func(context.Context, *coredata.TaskSyncJob) error
	}

	OutboundHooks struct {
		Newer func(context.Context, *coredata.TaskSyncJob, gid.GID, coredata.Scoper) (bool, error)
		Done  func(context.Context, *coredata.TaskSyncJob) error
	}
)

func HandleOutbound(
	svc *Service,
	hooks OutboundHooks,
	ctx context.Context,
	item *coredata.TaskSyncJob,
	payload shared.JobPayload,
) error {
	handler := &outboundHandler{
		svc:   svc,
		pg:    svc.pg,
		newer: hooks.Newer,
		done:  hooks.Done,
	}

	return handler.handle(ctx, item, payload)
}

func (h *outboundHandler) handle(
	ctx context.Context,
	item *coredata.TaskSyncJob,
	payload shared.JobPayload,
) error {
	scope := coredata.NewScopeFromObjectID(item.ID)

	client, err := h.linearClient(ctx, scope, payload.ConnectorID)
	if err != nil {
		return err
	}

	switch payload.Action {
	case shared.SyncActionCommentUpsert:
		return h.processCommentUpsert(ctx, item, payload, scope, client)
	case shared.SyncActionCommentDelete:
		return h.processCommentDelete(ctx, item, payload, client)
	}

	states, err := client.ListWorkflowStates(ctx, payload.TeamID)
	if err != nil {
		return fmt.Errorf("cannot list Linear workflow states: %w", err)
	}

	switch payload.Action {
	case shared.SyncActionCancel:
		return h.processCancel(ctx, item, payload, scope, client, states)
	case shared.SyncActionUpdate:
		return h.processUpdate(ctx, item, payload, scope, client, states)
	default:
		return fmt.Errorf("cannot process task sync job: unknown action %q", payload.Action)
	}
}

func (h *outboundHandler) linearClient(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) (*linear.Client, error) {
	var dbConnector *coredata.Connector

	err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			dbConnector = &coredata.Connector{}
			if err := dbConnector.LoadByID(
				ctx,
				conn,
				scope,
				connectorID,
				h.svc.encryptionKey,
			); err != nil {
				return fmt.Errorf("cannot load Linear connector: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	client, _, err := h.svc.linearClientForConnector(ctx, scope, dbConnector)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (h *outboundHandler) processCancel(
	ctx context.Context,
	item *coredata.TaskSyncJob,
	payload shared.JobPayload,
	scope coredata.Scoper,
	client *linear.Client,
	states []linear.WorkflowState,
) error {
	skip, err := h.newer(ctx, item, payload.TaskID, scope)
	if err != nil {
		return err
	}

	if skip {
		return h.done(ctx, item)
	}

	stateID, err := PickWorkflowStateID(states, coredata.TaskStateCanceled)
	if err != nil {
		return err
	}

	_, err = client.UpdateIssue(
		ctx,
		payload.ExternalID,
		linear.IssueUpdateInput{StateID: &stateID},
	)
	if err != nil {
		return fmt.Errorf("cannot cancel Linear issue: %w", err)
	}

	return h.done(ctx, item)
}

func (h *outboundHandler) processUpdate(
	ctx context.Context,
	item *coredata.TaskSyncJob,
	payload shared.JobPayload,
	scope coredata.Scoper,
	client *linear.Client,
	states []linear.WorkflowState,
) error {
	sent := false

	for range maxOutboundRevisionAttempts {
		task, destination, skip, err := h.loadOutboundUpdate(ctx, item, payload, scope, sent)
		if err != nil {
			return err
		}

		if skip {
			return h.done(ctx, item)
		}

		_, sentHash, err := outboundTaskHash(task)
		if err != nil {
			return fmt.Errorf("cannot hash outbound task: %w", err)
		}

		markdown, err := h.svc.linearIssueDescriptionMarkdown(
			ctx,
			scope,
			client,
			task.OrganizationID,
			task.Content,
		)
		if err != nil {
			return fmt.Errorf("cannot convert task mentions: %w", err)
		}

		stateID, err := PickWorkflowStateID(states, task.State)
		if err != nil {
			return err
		}

		priority, err := outboundLinearPriority(task.Priority, destination)
		if err != nil {
			return err
		}

		description, _, _, embedErr := h.svc.embedLinearDescription(
			ctx,
			client,
			scope,
			payload.ConnectorID,
			payload.ExternalID,
			task.OrganizationID,
			task.ID,
			markdown,
			task.Content,
		)
		// A failed attachment stays in the text. The issue update still sends
		// the title, state, priority, and assignee, and the job retries the file.
		if embedErr != nil && !isPartialLinearEmbed(embedErr) {
			return fmt.Errorf("cannot embed Linear files: %w", embedErr)
		}

		dueDate := shared.DeadlineToDate(task.Deadline)

		issue, err := client.UpdateIssue(
			ctx,
			payload.ExternalID,
			linear.IssueUpdateInput{
				Title:       &task.Name,
				Description: &description,
				StateID:     &stateID,
				Priority:    priority,
				DueDate:     dueDate,
				DueDateSet:  true,
				AssigneeID:  h.svc.linearAssigneeID(ctx, scope, client, task.AssignedToID),
			},
		)
		if err != nil {
			return fmt.Errorf("cannot update Linear issue: %w", err)
		}

		if embedErr != nil {
			return fmt.Errorf("cannot embed Linear files: %w", embedErr)
		}

		sent = true

		stable, err := h.finishOutboundUpdate(
			ctx,
			payload,
			scope,
			sentHash,
			issue,
			priority != nil,
		)
		if err != nil {
			return err
		}

		if stable {
			return h.done(ctx, item)
		}
	}

	return fmt.Errorf("task changed during Linear update")
}

func (h *outboundHandler) loadOutboundUpdate(
	ctx context.Context,
	item *coredata.TaskSyncJob,
	payload shared.JobPayload,
	scope coredata.Scoper,
	sent bool,
) (*coredata.Task, json.RawMessage, bool, error) {
	var (
		task        *coredata.Task
		destination json.RawMessage
		skip        bool
	)

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			task = &coredata.Task{}
			if err := task.LoadByID(ctx, tx, scope, payload.TaskID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					skip = true

					return nil
				}

				return fmt.Errorf("cannot load task %q: %w", payload.TaskID, err)
			}

			link, err := shared.LoadCurrentOutboundLink(ctx, tx, scope, payload.TaskID)
			if err != nil {
				return err
			}

			if !shared.OutboundLinkMatchesJob(link, payload) {
				skip = true

				return nil
			}

			destination = link.Destination

			if sent {
				return nil
			}

			newer, err := item.HasNewerOutboundJob(ctx, tx, scope, payload.TaskID, shared.CommentSyncActions())
			if err != nil {
				return err
			}

			skip = newer

			return nil
		},
	)
	if err != nil {
		return nil, nil, false, err
	}

	return task, destination, skip, nil
}

func (h *outboundHandler) finishOutboundUpdate(
	ctx context.Context,
	payload shared.JobPayload,
	scope coredata.Scoper,
	sentHash string,
	issue *linear.Issue,
	sentPriority bool,
) (bool, error) {
	var stable bool

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			task := &coredata.Task{}
			if err := task.LoadByID(ctx, tx, scope, payload.TaskID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					stable = true

					return nil
				}

				return fmt.Errorf("cannot load task %q: %w", payload.TaskID, err)
			}

			_, currentHash, err := outboundTaskHash(task)
			if err != nil {
				return err
			}

			if currentHash != sentHash {
				return nil
			}

			link, err := shared.LoadCurrentOutboundLink(ctx, tx, scope, payload.TaskID)
			if err != nil {
				return err
			}

			if !shared.OutboundLinkMatchesJob(link, payload) {
				stable = true

				return nil
			}

			link.ContentHash = &sentHash
			link.RemoteUpdatedAt = &issue.UpdatedAt
			link.ExternalIdentifier = issue.Identifier
			link.ExternalURL = issue.URL

			if sentPriority {
				destination, err := clearLinearNoneFromDestination(link.Destination)
				if err != nil {
					return err
				}

				link.Destination = destination
			}

			link.UpdatedAt = time.Now()

			if err := link.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update task external link: %w", err)
			}

			stable = true

			return nil
		},
	)
	if err != nil {
		return false, err
	}

	return stable, nil
}

func outboundTaskHash(task *coredata.Task) (string, string, error) {
	markdown, err := ContentToMarkdown(task.Content)
	if err != nil {
		return "", "", err
	}

	return markdown, ContentHash(
		task.Name,
		markdown,
		task.State,
		task.Priority,
		task.Deadline,
		task.AssignedToID,
	), nil
}

func (h *outboundHandler) processCommentDelete(
	ctx context.Context,
	item *coredata.TaskSyncJob,
	payload shared.JobPayload,
	client *linear.Client,
) error {
	if payload.ExternalCommentID == "" {
		return h.done(ctx, item)
	}

	if err := client.DeleteComment(ctx, payload.ExternalCommentID); err != nil {
		if errors.Is(err, linear.ErrCommentNotFound) {
			return h.done(ctx, item)
		}

		return fmt.Errorf("cannot delete Linear comment: %w", err)
	}

	return h.done(ctx, item)
}

func (h *outboundHandler) processCommentUpsert(
	ctx context.Context,
	item *coredata.TaskSyncJob,
	payload shared.JobPayload,
	scope coredata.Scoper,
	client *linear.Client,
) error {
	sent := false

	for range maxOutboundRevisionAttempts {
		comment, mapping, skip, err := shared.LoadCommentUpsert(ctx, h.pg, item, payload, scope, sent)
		if err != nil {
			return err
		}

		if skip {
			return h.done(ctx, item)
		}

		canonical, err := ContentToMarkdown(comment.Content)
		if err != nil {
			return fmt.Errorf("cannot convert task comment: %w", err)
		}

		if strings.TrimSpace(canonical) == "" {
			return h.done(ctx, item)
		}

		hash := shared.CommentContentHash(canonical)
		loaded := false

		var known map[string]string

		if mapping != nil && mapping.ContentHash != nil && *mapping.ContentHash == hash {
			ids, err := prosemirror.FileIDs(comment.Content)
			if err != nil {
				return fmt.Errorf("cannot read comment attachment file ids: %w", err)
			}

			if len(ids) == 0 {
				return h.done(ctx, item)
			}

			known, err = h.svc.linearAttachments(ctx, scope, payload.TaskID)
			if err != nil {
				return fmt.Errorf("cannot load Linear attachments: %w", err)
			}

			loaded = true

			missing, err := linearFilesMissing(known, comment.Content)
			if err != nil {
				return fmt.Errorf("cannot read comment attachment file ids: %w", err)
			}

			if !missing {
				return h.done(ctx, item)
			}
		}

		markdown, err := h.svc.linearCommentMarkdown(
			ctx,
			scope,
			client,
			comment.OrganizationID,
			comment.Content,
		)
		if err != nil {
			return fmt.Errorf("cannot convert task comment mentions: %w", err)
		}

		var description string

		if loaded {
			description, _, _, err = h.svc.embedKnownLinearDescription(
				ctx,
				client,
				scope,
				payload.ConnectorID,
				payload.ExternalID,
				comment.OrganizationID,
				payload.TaskID,
				markdown,
				comment.Content,
				known,
			)
		} else {
			description, _, _, err = h.svc.embedLinearDescription(
				ctx,
				client,
				scope,
				payload.ConnectorID,
				payload.ExternalID,
				comment.OrganizationID,
				payload.TaskID,
				markdown,
				comment.Content,
			)
		}

		if err != nil {
			return err
		}

		created := mapping == nil

		var remote *linear.Comment

		if created {
			author, err := h.svc.commentAuthor(ctx, scope, comment.OwnerID)
			if err != nil {
				return fmt.Errorf("cannot load Linear comment author: %w", err)
			}

			remote, err = client.CreateComment(ctx, payload.ExternalID, description, author)
			if err != nil {
				return fmt.Errorf("cannot create Linear comment: %w", err)
			}
		} else {
			remote, err = client.UpdateComment(ctx, mapping.ExternalID, description)
			if err != nil {
				return fmt.Errorf("cannot update Linear comment: %w", err)
			}
		}

		sent = true

		stable, discardID, err := h.finishCommentUpsert(
			ctx,
			payload,
			scope,
			comment.ID,
			remote,
			hash,
			created,
		)
		if err != nil {
			return err
		}

		if discardID != "" {
			if err := discardLinearComment(ctx, client, discardID); err != nil {
				return err
			}
		}

		if stable {
			return h.done(ctx, item)
		}
	}

	return fmt.Errorf("task comment changed during Linear update")
}

func (h *outboundHandler) finishCommentUpsert(
	ctx context.Context,
	payload shared.JobPayload,
	scope coredata.Scoper,
	commentID gid.GID,
	remote *linear.Comment,
	sentHash string,
	created bool,
) (bool, string, error) {
	var (
		stable    bool
		discardID string
	)

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			comment := &coredata.TaskComment{}
			if err := comment.LoadByIDForUpdate(ctx, tx, scope, commentID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					stable = true

					if created {
						discardID = remote.ID
					}

					return nil
				}

				return fmt.Errorf("cannot load task comment: %w", err)
			}

			rendered, err := ContentToMarkdown(comment.Content)
			if err != nil {
				return fmt.Errorf("cannot convert task comment: %w", err)
			}

			if strings.TrimSpace(rendered) == "" {
				stable = true

				if created {
					discardID = remote.ID
				}

				return nil
			}

			if shared.CommentContentHash(rendered) != sentHash {
				if created {
					discardID = remote.ID
				}

				return nil
			}

			link, err := shared.LoadCurrentOutboundLink(ctx, tx, scope, payload.TaskID)
			if err != nil {
				return err
			}

			if !shared.OutboundLinkMatchesJob(link, payload) {
				stable = true

				if created {
					discardID = remote.ID
				}

				return nil
			}

			existing := &coredata.TaskCommentExternalLink{}

			err = existing.LoadByCommentID(ctx, tx, scope, commentID)
			if errors.Is(err, coredata.ErrResourceNotFound) {
				if !created {
					stable = true

					return nil
				}

				createdLink := newCommentLink(link, commentID, remote, sentHash, time.Now())
				if err := createdLink.Insert(ctx, tx, scope); err != nil {
					if errors.Is(err, coredata.ErrResourceAlreadyExists) {
						discardID = remote.ID

						return nil
					}

					return fmt.Errorf("cannot insert task comment external link: %w", err)
				}

				stable = true

				return nil
			}

			if err != nil {
				return fmt.Errorf("cannot load task comment external link: %w", err)
			}

			if created {
				discardID = remote.ID
				stable = existing.ContentHash != nil && *existing.ContentHash == sentHash

				return nil
			}

			if err := saveCommentLinkTx(ctx, tx, scope, link, commentID, remote, sentHash); err != nil {
				return err
			}

			stable = true

			return nil
		},
	)
	if err != nil {
		return false, "", err
	}

	return stable, discardID, nil
}

func discardLinearComment(ctx context.Context, client *linear.Client, commentID string) error {
	if err := client.DeleteComment(ctx, commentID); err != nil && !errors.Is(err, linear.ErrCommentNotFound) {
		return fmt.Errorf("cannot delete stale Linear comment: %w", err)
	}

	return nil
}
