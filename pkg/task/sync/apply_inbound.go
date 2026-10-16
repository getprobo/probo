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
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/notification"
	"go.probo.inc/probo/pkg/task/sync/linear"
	"go.probo.inc/probo/pkg/webhook"
)

func (s *Service) ApplyInboundIssue(
	ctx context.Context,
	envelope *linear.WebhookEnvelope,
) error {
	if envelope.Type != "Issue" {
		return nil
	}

	if envelope.Action == "remove" {
		return s.unlinkRemovedIssue(ctx, envelope)
	}

	if envelope.Action != "update" && envelope.Action != "create" {
		return nil
	}

	data, err := envelope.IssueData()
	if err != nil {
		return err
	}

	if data.ID == "" {
		return nil
	}

	update, creates, err := s.previewInboundIssue(ctx, envelope, data)
	if err != nil {
		return err
	}

	var (
		unusedUpdate  *linearImport
		unusedCreates []preparedImportedIssue
	)

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			links := coredata.TaskExternalLinks{}
			if err := links.LoadByExternalID(
				ctx,
				tx,
				coredata.NewNoScope(),
				coredata.ConnectorProviderLinearSync,
				data.ID,
			); err != nil {
				return fmt.Errorf("cannot load task external links: %w", err)
			}

			link, ok := PickLinkForLinearOrganization(links, envelope.OrganizationID)
			if !ok {
				if envelope.Action != "create" {
					unusedUpdate = update

					return nil
				}

				return s.importCreatedLinearIssue(ctx, tx, envelope, data, creates)
			}

			unusedCreates = creates

			if isAppActor(link.Metadata, envelope.Actor.ID) {
				unusedUpdate = update

				return nil
			}

			scope := coredata.NewScope(link.OrganizationID.TenantID())
			if err := link.LoadByTaskIDForUpdate(ctx, tx, scope, link.TaskID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					unusedUpdate = update

					return nil
				}

				return fmt.Errorf("cannot lock task external link: %w", err)
			}

			if link.ExternalID != data.ID {
				unusedUpdate = update

				return nil
			}

			if isAppActor(link.Metadata, envelope.Actor.ID) {
				unusedUpdate = update

				return nil
			}

			remoteUpdatedAt := linearDateTime(data.UpdatedAt)
			if inboundEventIsStale(remoteUpdatedAt, link.RemoteUpdatedAt) {
				unusedUpdate = update

				return nil
			}

			task := &coredata.Task{}
			if err := task.LoadByIDForUpdate(ctx, tx, scope, link.TaskID); err != nil {
				return fmt.Errorf("cannot load task %q: %w", link.TaskID, err)
			}

			var metas map[string]importedFileMeta

			if data.Has("description") && update != nil {
				data.Description = update.markdown
				metas = update.metas

				if next, changed, err := mergedLinearAttachments(
					link.Metadata,
					mergeAttachmentMaps(attachmentsFromMetadata(link.Metadata), update.files),
				); err != nil {
					return err
				} else if changed {
					link.Metadata = next
				}
			}

			mapped, err := mapInboundIssue(task, data, attachmentsFromMetadata(link.Metadata))
			if err != nil {
				return err
			}

			if len(metas) > 0 {
				content, markdown, err := applyImportedFiles(mapped.Content, metas)
				if err != nil {
					return fmt.Errorf("cannot apply imported Linear files: %w", err)
				}

				mapped.Content = content
				mapped.Markdown = markdown
			}

			oldTask := *task

			if _, err := applyInboundAssignee(
				ctx,
				tx,
				scope,
				link.OrganizationID,
				task,
				data,
			); err != nil {
				return fmt.Errorf("cannot apply Linear assignee: %w", err)
			}

			hash := ContentHash(
				mapped.Name,
				mapped.Markdown,
				mapped.State,
				mapped.Priority,
				mapped.Deadline,
				task.AssignedToID,
			)
			if link.ContentHash != nil && *link.ContentHash == hash {
				return touchInboundLink(ctx, tx, scope, link, data, remoteUpdatedAt, hash, mapped.Priority)
			}

			oldState := task.State
			oldPriority := task.Priority
			now := time.Now()

			task.Name = mapped.Name
			task.Content = mapped.Content
			task.State = mapped.State
			task.Priority = mapped.Priority
			task.Deadline = mapped.Deadline
			task.UpdatedAt = now

			if task.State != oldState || task.Priority != oldPriority {
				if err := task.NextRankForStatePriority(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot get next rank: %w", err)
				}
			}

			if err := task.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update task from Linear: %w", err)
			}

			actorID, err := inboundActivityActorID(ctx, tx, scope, link.OrganizationID, envelope.Actor)
			if err != nil {
				return fmt.Errorf("cannot resolve Linear activity actor: %w", err)
			}

			if err := s.recordInboundActivities(
				ctx,
				tx,
				scope,
				&oldTask,
				task,
				actorID,
				now,
			); err != nil {
				return err
			}

			assigneeChanged := task.AssignedToID != nil &&
				(oldTask.AssignedToID == nil || *oldTask.AssignedToID != *task.AssignedToID)
			if assigneeChanged {
				if err := notification.InsertTaskAssignment(
					ctx,
					tx,
					scope,
					task.OrganizationID,
					task.ID,
					task.Name,
					*task.AssignedToID,
					actorID,
				); err != nil {
					return err
				}
			}

			internalControlIDs, err := (coredata.InternalControlTask{}).LoadInternalControlIDsByTaskID(ctx, tx, scope, task.ID)
			if err != nil {
				return fmt.Errorf("cannot load task internal controls: %w", err)
			}

			if err := webhook.InsertTaskUpdated(ctx, tx, scope, &oldTask, task, internalControlIDs, internalControlIDs); err != nil {
				return fmt.Errorf("cannot emit task updated webhook: %w", err)
			}

			return touchInboundLink(ctx, tx, scope, link, data, remoteUpdatedAt, hash, mapped.Priority)
		},
	)
	if err != nil {
		return err
	}

	if err := s.discardInboundImport(ctx, unusedUpdate); err != nil {
		return err
	}

	return s.discardPreparedIssues(ctx, unusedCreates)
}

func (s *Service) unlinkRemovedIssue(
	ctx context.Context,
	envelope *linear.WebhookEnvelope,
) error {
	data, err := envelope.IssueData()
	if err != nil {
		return err
	}

	if data.ID == "" {
		return nil
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			link, scope, ok, err := lockedLinearLink(ctx, tx, data.ID, envelope.OrganizationID)
			if err != nil || !ok {
				return err
			}

			if err := link.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete task external link: %w", err)
			}

			if err := deleteCommentLinks(ctx, tx, scope, link.TaskID); err != nil {
				return err
			}

			return nil
		},
	)
}

type inboundMappedFields struct {
	Name     string
	Content  string
	Markdown string
	State    coredata.TaskState
	Priority coredata.TaskPriority
	Deadline *time.Time
}

func mapInboundIssue(
	task *coredata.Task,
	data *linear.IssueWebhookData,
	files map[string]string,
) (inboundMappedFields, error) {
	markdown, err := ContentToMarkdown(task.Content)
	if err != nil {
		return inboundMappedFields{}, err
	}

	mapped := inboundMappedFields{
		Name:     task.Name,
		Content:  task.Content,
		Markdown: markdown,
		State:    task.State,
		Priority: task.Priority,
		Deadline: task.Deadline,
	}

	if data.Has("title") {
		mapped.Name = data.Title
	}

	if data.Has("description") {
		content, err := MarkdownToContent(restoreLinearFiles(data.Description, files))
		if err != nil {
			return inboundMappedFields{}, err
		}

		canonical, err := ContentToMarkdown(content)
		if err != nil {
			return inboundMappedFields{}, err
		}

		mapped.Content = content
		mapped.Markdown = canonical
	}

	if data.Has("state") {
		mapped.State = LinearTypeToTaskState(data.State.Type)
	}

	if data.Has("priority") {
		if priority, ok := LinearPriorityToTask(data.Priority); ok {
			mapped.Priority = priority
		}
	}

	if data.Has("dueDate") {
		mapped.Deadline = LinearDateToDeadline(data.DueDate)
	}

	return mapped, nil
}

func (s *Service) previewInboundIssue(
	ctx context.Context,
	envelope *linear.WebhookEnvelope,
	data *linear.IssueWebhookData,
) (*linearImport, []preparedImportedIssue, error) {
	link, _, ok, err := s.loadLinearLink(ctx, data.ID, envelope.OrganizationID)
	if err != nil {
		return nil, nil, err
	}

	if ok {
		if isAppActor(link.Metadata, envelope.Actor.ID) || !data.Has("description") {
			return nil, nil, nil
		}

		imported, err := s.prepareLinearImport(
			ctx,
			coredata.NewScope(link.OrganizationID.TenantID()),
			link.ConnectorID,
			link.OrganizationID,
			data.Description,
			attachmentsFromMetadata(link.Metadata),
			false,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot import Linear description files: %w", err)
		}

		return imported, nil, nil
	}

	if envelope.Action != "create" {
		return nil, nil, nil
	}

	creates, err := s.prepareImportedIssues(ctx, envelope, data)
	if err != nil {
		return nil, nil, err
	}

	return nil, creates, nil
}

func (s *Service) loadLinearLink(
	ctx context.Context,
	issueID string,
	linearOrganizationID string,
) (*coredata.TaskExternalLink, coredata.Scoper, bool, error) {
	if s.pg == nil {
		return nil, nil, false, fmt.Errorf("cannot load task external link: database is not configured")
	}

	links := coredata.TaskExternalLinks{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := links.LoadByExternalID(
				ctx,
				conn,
				coredata.NewNoScope(),
				coredata.ConnectorProviderLinearSync,
				issueID,
			); err != nil {
				return fmt.Errorf("cannot load task external links: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, false, err
	}

	link, ok := PickLinkForLinearOrganization(links, linearOrganizationID)
	if !ok {
		return nil, nil, false, nil
	}

	return link, coredata.NewScope(link.OrganizationID.TenantID()), true, nil
}

func lockedLinearLink(
	ctx context.Context,
	tx pg.Tx,
	issueID string,
	linearOrganizationID string,
) (*coredata.TaskExternalLink, coredata.Scoper, bool, error) {
	links := coredata.TaskExternalLinks{}
	if err := links.LoadByExternalID(
		ctx,
		tx,
		coredata.NewNoScope(),
		coredata.ConnectorProviderLinearSync,
		issueID,
	); err != nil {
		return nil, nil, false, fmt.Errorf("cannot load task external links: %w", err)
	}

	link, ok := PickLinkForLinearOrganization(links, linearOrganizationID)
	if !ok {
		return nil, nil, false, nil
	}

	scope := coredata.NewScope(link.OrganizationID.TenantID())
	if err := link.LoadByTaskIDForUpdate(ctx, tx, scope, link.TaskID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil, false, nil
		}

		return nil, nil, false, fmt.Errorf("cannot lock task external link: %w", err)
	}

	if link.ExternalID != issueID {
		return nil, nil, false, nil
	}

	return link, scope, true, nil
}

func PickLinkForLinearOrganization(
	links coredata.TaskExternalLinks,
	linearOrganizationID string,
) (*coredata.TaskExternalLink, bool) {
	if linearOrganizationID == "" {
		return nil, false
	}

	var match *coredata.TaskExternalLink

	for _, link := range links {
		if link.DestinationLinearOrganizationID() != linearOrganizationID {
			continue
		}

		if match != nil {
			return nil, false
		}

		match = link
	}

	if match == nil {
		return nil, false
	}

	return match, true
}

func touchInboundLink(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	link *coredata.TaskExternalLink,
	data *linear.IssueWebhookData,
	remoteUpdatedAt *time.Time,
	hash string,
	mappedPriority coredata.TaskPriority,
) error {
	link.ContentHash = &hash

	if remoteUpdatedAt != nil {
		link.RemoteUpdatedAt = remoteUpdatedAt
	}

	if data.Has("identifier") && data.Identifier != "" {
		link.ExternalIdentifier = data.Identifier
	}

	if data.Has("url") && data.URL != "" {
		link.ExternalURL = data.URL
	}

	if data.Has("priority") {
		destination, err := applyLinearPriorityToDestination(
			link.Destination,
			data.Priority,
			mappedPriority,
		)
		if err != nil {
			return err
		}

		link.Destination = destination
	}

	link.UpdatedAt = time.Now()

	if err := link.Update(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot update task external link: %w", err)
	}

	return nil
}

func isAppActor(metadata json.RawMessage, actorID string) bool {
	if actorID == "" || len(metadata) == 0 {
		return false
	}

	var payload struct {
		AppActorID string `json:"app_actor_id"`
	}

	if err := json.Unmarshal(metadata, &payload); err != nil {
		return false
	}

	return payload.AppActorID != "" && payload.AppActorID == actorID
}

func inboundEventIsStale(remoteUpdatedAt, stored *time.Time) bool {
	if stored == nil {
		return false
	}

	if remoteUpdatedAt == nil {
		return true
	}

	return !remoteUpdatedAt.After(*stored)
}

func (s *Service) recordInboundActivities(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	oldTask *coredata.Task,
	task *coredata.Task,
	actorID *gid.GID,
	now time.Time,
) error {
	if s == nil || s.recordUpdateActivities == nil {
		return nil
	}

	if err := s.recordUpdateActivities(ctx, tx, scope, oldTask, task, actorID, now); err != nil {
		return fmt.Errorf("cannot record task update events: %w", err)
	}

	return nil
}

func inboundActivityActorID(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	actor linear.WebhookActor,
) (*gid.GID, error) {
	email := strings.TrimSpace(actor.Email)
	if email == "" {
		return nil, nil
	}

	addr, err := mail.ParseAddr(email)
	if err != nil {
		return nil, nil
	}

	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByOrganizationIDAndEmail(
		ctx,
		tx,
		scope,
		organizationID,
		addr,
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &profile.ID, nil
}

func linearDateTime(value string) *time.Time {
	if value == "" {
		return nil
	}

	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return &t
	}

	if t, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return &t
	}

	return nil
}
