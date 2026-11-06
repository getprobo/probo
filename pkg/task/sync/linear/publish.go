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
	"net/url"
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	linear "go.probo.inc/probo/pkg/task/sync/linear/client"
	"go.probo.inc/probo/pkg/task/sync/shared"
	"go.probo.inc/probo/pkg/webhook"
)

const linearPublishPendingExternalIDPrefix = "pending:"

func (s *Service) PublishToLinear(
	ctx context.Context,
	scope coredata.Scoper,
	taskID gid.GID,
	teamID string,
) (*coredata.TaskExternalLink, error) {
	if teamID == "" {
		return nil, ErrLinearTeamIDRequired
	}

	var (
		task        *coredata.Task
		client      *linear.Client
		dbConnector *coredata.Connector
	)

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			task = &coredata.Task{}
			if err := task.LoadByID(ctx, conn, scope, taskID); err != nil {
				return fmt.Errorf("cannot load task %q: %w", taskID, err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load task for Linear publish: %w", err)
	}

	accounts, err := s.linearAccountsForOrganization(ctx, scope, task.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("cannot load Linear accounts: %w", err)
	}

	account, err := s.linearAccountForTeam(ctx, accounts, teamID)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve Linear team: %w", err)
	}

	client = account.client
	dbConnector = account.connector

	states, err := client.ListWorkflowStates(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("cannot list Linear workflow states: %w", err)
	}

	stateID, err := PickWorkflowStateID(states, task.State)
	if err != nil {
		return nil, fmt.Errorf("cannot pick Linear workflow state: %w", err)
	}

	canonical, err := ContentToMarkdown(task.Content)
	if err != nil {
		return nil, fmt.Errorf("cannot convert task content to markdown: %w", err)
	}

	viewerID, err := client.ViewerID(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot load Linear viewer: %w", err)
	}

	linearOrganizationID, err := client.OrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot load Linear organization: %w", err)
	}

	destination, err := json.Marshal(
		coredata.TaskExternalLinkDestination{
			TeamID:               teamID,
			LinearOrganizationID: linearOrganizationID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal destination: %w", err)
	}

	reused, archiveIssueID, archiveConnectorID, err := s.claimLinearLink(
		ctx,
		scope,
		task,
		dbConnector,
		destination,
		teamID,
		"",
	)
	if err != nil {
		return nil, fmt.Errorf("cannot claim pending Linear publish: %w", err)
	}

	s.archiveAbandonedLinearIssue(ctx, accounts, task.ID, archiveConnectorID, archiveIssueID)

	var (
		issue *linear.Issue
		files map[string]string
	)

	if reused != nil {
		issue = issueFromExternalLink(reused)
		files = attachmentsFromMetadata(reused.Metadata)
	} else {
		markdown, err := s.linearIssueDescriptionMarkdown(
			ctx,
			scope,
			client,
			task.OrganizationID,
			task.Content,
		)
		if err != nil {
			s.compensateFailedPublish(ctx, client, scope, taskID, "")

			return nil, fmt.Errorf("cannot convert task mentions: %w", err)
		}

		description, uploaded, _, err := s.embedLinearDescription(
			ctx,
			client,
			scope,
			dbConnector.ID,
			linearPublishPendingExternalID(task.ID),
			task.OrganizationID,
			task.ID,
			markdown,
			task.Content,
		)
		if err != nil {
			s.compensateFailedPublish(ctx, client, scope, taskID, "")

			return nil, fmt.Errorf("cannot embed Linear files: %w", err)
		}

		files = uploaded

		issue, err = client.CreateIssue(
			ctx,
			linear.IssueInput{
				TeamID:      teamID,
				Title:       task.Name,
				Description: description,
				StateID:     stateID,
				Priority:    TaskPriorityToLinear(task.Priority),
				DueDate:     shared.DeadlineToDate(task.Deadline),
				AssigneeID:  s.linearAssigneeID(ctx, scope, client, task.AssignedToID),
			},
		)
		if err != nil {
			s.compensateFailedPublish(ctx, client, scope, taskID, "")
			return nil, fmt.Errorf("cannot create Linear issue: %w", err)
		}
	}

	link, err := s.finishLinearPublish(
		ctx,
		scope,
		client,
		task,
		dbConnector,
		destination,
		viewerID,
		issue,
		canonical,
		files,
	)
	if err != nil {
		s.compensateFailedPublish(ctx, client, scope, taskID, issue.ID)
		return nil, fmt.Errorf("cannot finish Linear publish: %w", err)
	}

	createdIDs, err := s.copyTaskCommentsToLinear(ctx, scope, client, link)
	if err != nil {
		s.rollbackLinkedComments(ctx, scope, client, nil, createdIDs)
		s.compensateFailedPublish(ctx, client, scope, taskID, issue.ID)

		return nil, fmt.Errorf("cannot publish task comments to Linear: %w", err)
	}

	return link, nil
}

func (s *Service) LinkToLinear(
	ctx context.Context,
	scope coredata.Scoper,
	taskID gid.GID,
	teamID string,
	issueID string,
	identityID *gid.GID,
) (*coredata.TaskExternalLink, error) {
	if teamID == "" {
		return nil, ErrLinearTeamIDRequired
	}

	if issueID == "" {
		return nil, ErrLinearIssueIDRequired
	}

	var task *coredata.Task

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			task = &coredata.Task{}
			if err := task.LoadByID(ctx, conn, scope, taskID); err != nil {
				return fmt.Errorf("cannot load task %q: %w", taskID, err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load task for Linear link: %w", err)
	}

	accounts, err := s.linearAccountsForOrganization(ctx, scope, task.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("cannot load Linear accounts: %w", err)
	}

	account, err := s.linearAccountForTeam(ctx, accounts, teamID)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve Linear team: %w", err)
	}

	issue, err := account.client.GetIssue(ctx, issueID)
	if err != nil {
		if linearIssueNotFound(err) {
			return nil, ErrLinearIssueNotFound
		}

		return nil, fmt.Errorf("cannot load Linear issue: %w", err)
	}

	if issue.TeamID != teamID {
		return nil, ErrLinearIssueNotFound
	}

	linearOrganizationID, err := account.client.OrganizationID(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot load Linear organization: %w", err)
	}

	viewerID, err := account.client.ViewerID(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot load Linear viewer: %w", err)
	}

	destination, err := json.Marshal(
		coredata.TaskExternalLinkDestination{
			TeamID:               teamID,
			LinearOrganizationID: linearOrganizationID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal destination: %w", err)
	}

	_, archiveIssueID, archiveConnectorID, err := s.claimLinearLink(
		ctx,
		scope,
		task,
		account.connector,
		destination,
		teamID,
		issue.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot claim pending Linear link: %w", err)
	}

	s.archiveAbandonedLinearIssue(ctx, accounts, task.ID, archiveConnectorID, archiveIssueID)

	content, markdown, files, err := s.linearIssueContent(
		ctx,
		account.client,
		scope,
		task.OrganizationID,
		account.connector.ID,
		issue.Description,
	)
	if err != nil {
		s.deletePendingLinearLink(ctx, scope, taskID, issue.ID)

		return nil, fmt.Errorf("cannot convert Linear description: %w", err)
	}

	taskURL, err := s.taskURL(task.OrganizationID, task.ID)
	if err != nil {
		s.deletePendingLinearLink(ctx, scope, taskID, issue.ID)

		return nil, fmt.Errorf("cannot build task URL: %w", err)
	}

	attachmentID, err := account.client.LinkAttachment(ctx, issue.ID, taskURL, "Probo task")
	if err != nil {
		s.deletePendingLinearLink(ctx, scope, taskID, issue.ID)

		return nil, fmt.Errorf("cannot link Linear attachment: %w", err)
	}

	link, previous, linked, err := s.finishLinearLink(
		ctx,
		scope,
		task,
		account.connector,
		destination,
		viewerID,
		issue,
		content,
		markdown,
		files,
		attachmentID,
		identityID,
	)
	if err != nil {
		if deleteErr := account.client.DeleteAttachment(ctx, attachmentID); deleteErr != nil && s.logger != nil {
			s.logger.WarnCtx(
				ctx,
				"cannot delete Linear attachment after failed link",
				log.String("task_id", taskID.String()),
				log.Error(deleteErr),
			)
		}

		if !errors.Is(err, shared.ErrTaskAlreadyLinked) {
			s.deletePendingLinearLink(ctx, scope, taskID, issue.ID)
		}

		return nil, fmt.Errorf("cannot finish Linear link: %w", err)
	}

	if err := s.syncLinkedComments(ctx, scope, account.client, link); err != nil {
		restoreErr := s.restoreLinkedTask(ctx, scope, previous, linked, identityID)

		if deleteErr := account.client.DeleteAttachment(ctx, attachmentID); deleteErr != nil && s.logger != nil {
			s.logger.WarnCtx(
				ctx,
				"cannot delete Linear attachment after failed comment sync",
				log.String("task_id", taskID.String()),
				log.Error(deleteErr),
			)
		}

		s.deletePendingLinearLink(ctx, scope, taskID, issue.ID)

		if restoreErr != nil {
			return nil, fmt.Errorf("cannot restore task after failed Linear comment sync: %w", restoreErr)
		}

		return nil, fmt.Errorf("cannot sync Linear comments: %w", err)
	}

	if err := s.notifyLinkedMentions(ctx, scope, previous, linked, identityID); err != nil && s.logger != nil {
		s.logger.WarnCtx(
			ctx,
			"cannot notify newly linked Linear mentions",
			log.String("task_id", taskID.String()),
			log.Error(err),
		)
	}

	return link, nil
}

func linearPublishPendingExternalID(taskID gid.GID) string {
	return linearPublishPendingExternalIDPrefix + taskID.String()
}

func isLinearPublishPending(link *coredata.TaskExternalLink) bool {
	return link != nil && strings.HasPrefix(link.ExternalID, linearPublishPendingExternalIDPrefix)
}

func isLinearPublishComplete(link *coredata.TaskExternalLink) bool {
	if link == nil || isLinearPublishPending(link) {
		return false
	}

	var payload struct {
		AttachmentID string `json:"attachment_id"`
	}

	if err := json.Unmarshal(link.Metadata, &payload); err != nil {
		return false
	}

	return payload.AttachmentID != ""
}

func PublishComplete(link *coredata.TaskExternalLink) bool {
	if link == nil || link.ExternalID == "" || isLinearPublishPending(link) {
		return false
	}

	if link.Provider == coredata.ConnectorProviderAsana {
		return true
	}

	return isLinearPublishComplete(link)
}

func reuseExistingLinearIssue(link *coredata.TaskExternalLink, teamID string) (bool, error) {
	if link == nil || isLinearPublishPending(link) {
		return false, nil
	}

	existingTeamID, err := link.DestinationTeamID()
	if err != nil {
		return false, fmt.Errorf("cannot read existing Linear destination: %w", err)
	}

	return existingTeamID == teamID, nil
}

func clientForConnector(accounts []linearAccount, connectorID gid.GID) *linear.Client {
	for i := range accounts {
		if accounts[i].connector.ID == connectorID {
			return accounts[i].client
		}
	}

	return nil
}

func (s *Service) archiveAbandonedLinearIssue(
	ctx context.Context,
	accounts []linearAccount,
	taskID gid.GID,
	connectorID gid.GID,
	issueID string,
) {
	if issueID == "" {
		return
	}

	archiveClient := clientForConnector(accounts, connectorID)
	if archiveClient == nil {
		return
	}

	if err := archiveClient.ArchiveIssue(ctx, issueID); err != nil && s.logger != nil {
		s.logger.WarnCtx(
			ctx,
			"cannot archive Linear issue after team change",
			log.String("task_id", taskID.String()),
			log.Error(err),
		)
	}
}

func keepLinearLink(link *coredata.TaskExternalLink, teamID, keepIssueID string) (bool, error) {
	if isLinearPublishPending(link) {
		existingTeamID, err := link.DestinationTeamID()
		if err != nil {
			return false, fmt.Errorf("cannot read existing Linear destination: %w", err)
		}

		return existingTeamID == teamID, nil
	}

	if keepIssueID != "" {
		return link.ExternalID == keepIssueID, nil
	}

	return reuseExistingLinearIssue(link, teamID)
}

func issueFromExternalLink(link *coredata.TaskExternalLink) *linear.Issue {
	issue := &linear.Issue{
		ID:         link.ExternalID,
		Identifier: link.ExternalIdentifier,
		URL:        link.ExternalURL,
	}

	if link.RemoteUpdatedAt != nil {
		issue.UpdatedAt = *link.RemoteUpdatedAt
	}

	return issue
}

func (s *Service) claimLinearLink(
	ctx context.Context,
	scope coredata.Scoper,
	task *coredata.Task,
	dbConnector *coredata.Connector,
	destination json.RawMessage,
	teamID string,
	keepIssueID string,
) (*coredata.TaskExternalLink, string, gid.GID, error) {
	now := time.Now()
	pending := &coredata.TaskExternalLink{
		OrganizationID: task.OrganizationID,
		TaskID:         task.ID,
		ConnectorID:    dbConnector.ID,
		Provider:       coredata.ConnectorProviderLinearSync,
		ExternalID:     linearPublishPendingExternalID(task.ID),
		Destination:    destination,
		Origin:         coredata.TaskExternalLinkOriginProbo,
		Metadata:       json.RawMessage(`{}`),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	var (
		reused             *coredata.TaskExternalLink
		archiveIssueID     string
		archiveConnectorID gid.GID
	)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			existing := &coredata.TaskExternalLink{}

			err := existing.LoadByTaskIDForUpdate(ctx, tx, scope, task.ID)
			if errors.Is(err, coredata.ErrResourceNotFound) {
				if err := pending.Insert(ctx, tx, scope); err != nil {
					if errors.Is(err, coredata.ErrResourceAlreadyExists) {
						return shared.ErrTaskAlreadyLinked
					}

					return fmt.Errorf("cannot insert pending task external link: %w", err)
				}

				return nil
			}

			if err != nil {
				return fmt.Errorf("cannot load task external link: %w", err)
			}

			if PublishComplete(existing) {
				return shared.ErrTaskAlreadyLinked
			}

			keep, err := keepLinearLink(existing, teamID, keepIssueID)
			if err != nil {
				return err
			}

			if keep {
				if keepIssueID == "" && !isLinearPublishPending(existing) {
					copied := *existing
					reused = &copied
				}

				return nil
			}

			if !isLinearPublishPending(existing) {
				archiveIssueID = existing.ExternalID
				archiveConnectorID = existing.ConnectorID
			}

			if err := existing.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete task external link: %w", err)
			}

			if err := shared.DeleteCommentLinks(ctx, tx, scope, task.ID); err != nil {
				return err
			}

			if err := pending.Insert(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return shared.ErrTaskAlreadyLinked
				}

				return fmt.Errorf("cannot insert pending task external link: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, "", gid.GID{}, fmt.Errorf("cannot claim Linear link: %w", err)
	}

	return reused, archiveIssueID, archiveConnectorID, nil
}

func (s *Service) linearIssueContent(
	ctx context.Context,
	client *linear.Client,
	scope coredata.Scoper,
	organizationID gid.GID,
	connectorID gid.GID,
	description string,
) (string, string, map[string]string, error) {
	markdown, files, metas, err := s.importLinearMarkdown(
		ctx,
		client,
		scope,
		connectorID,
		organizationID,
		description,
		nil,
	)
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot import Linear description files: %w", err)
	}

	content, canonical, err := s.contentFromLinearMarkdown(
		ctx,
		client,
		scope,
		organizationID,
		markdown,
		metas,
	)
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot convert Linear description: %w", err)
	}

	return content, canonical, files, nil
}

func (s *Service) finishLinearLink(
	ctx context.Context,
	scope coredata.Scoper,
	task *coredata.Task,
	dbConnector *coredata.Connector,
	destination json.RawMessage,
	viewerID string,
	issue *linear.Issue,
	content string,
	markdown string,
	files map[string]string,
	attachmentID string,
	identityID *gid.GID,
) (*coredata.TaskExternalLink, *coredata.Task, *coredata.Task, error) {
	metadata, err := linearLinkMetadata(attachmentID, viewerID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot build Linear link metadata: %w", err)
	}

	if len(files) > 0 {
		metadata, err = mergeAttachments(metadata, files)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("cannot merge Linear attachments: %w", err)
		}
	}

	now := time.Now()
	link := &coredata.TaskExternalLink{
		OrganizationID:     task.OrganizationID,
		TaskID:             task.ID,
		ConnectorID:        dbConnector.ID,
		Provider:           coredata.ConnectorProviderLinearSync,
		ExternalID:         issue.ID,
		ExternalIdentifier: issue.Identifier,
		ExternalURL:        issue.URL,
		Destination:        destination,
		Origin:             coredata.TaskExternalLinkOriginProbo,
		RemoteUpdatedAt:    &issue.UpdatedAt,
		Metadata:           metadata,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	var (
		previous *coredata.Task
		linked   *coredata.Task
	)

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			current := &coredata.Task{}
			if err := current.LoadByIDForUpdate(ctx, tx, scope, task.ID); err != nil {
				return fmt.Errorf("cannot load task %q: %w", task.ID, err)
			}

			existing := &coredata.TaskExternalLink{}
			if err := existing.LoadByTaskIDForUpdate(ctx, tx, scope, task.ID); err != nil {
				return fmt.Errorf("cannot load task external link: %w", err)
			}

			if !isLinearPublishPending(existing) && existing.ExternalID != issue.ID {
				return shared.ErrTaskAlreadyLinked
			}

			snapshot := *current
			previous = &snapshot
			oldTask := snapshot

			if err := applyLinearIssueFields(ctx, tx, scope, current, issue, content, now); err != nil {
				return fmt.Errorf("cannot apply Linear issue to task: %w", err)
			}

			applied := *current
			linked = &applied

			actorID, err := s.actorProfile(ctx, tx, scope, identityID, current.OrganizationID)
			if err != nil {
				return fmt.Errorf("cannot resolve task activity actor: %w", err)
			}

			if s.recordUpdateActivities != nil {
				if err := s.recordUpdateActivities(
					ctx,
					tx,
					scope,
					&oldTask,
					current,
					actorID,
					now,
				); err != nil {
					return fmt.Errorf("cannot record task update events: %w", err)
				}
			}

			internalControlIDs, err := (coredata.InternalControlTask{}).LoadInternalControlIDsByTaskID(ctx, tx, scope, current.ID)
			if err != nil {
				return fmt.Errorf("cannot load task internal controls: %w", err)
			}

			if err := webhook.InsertTaskUpdated(ctx, tx, scope, &oldTask, current, internalControlIDs, internalControlIDs); err != nil {
				return fmt.Errorf("cannot emit task updated webhook: %w", err)
			}

			hash := ContentHash(
				current.Name,
				markdown,
				current.State,
				current.Priority,
				current.Deadline,
				current.AssignedToID,
			)
			link.ContentHash = &hash

			link.CreatedAt = existing.CreatedAt
			if err := link.Update(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return shared.ErrTaskAlreadyLinked
				}

				return fmt.Errorf("cannot update task external link: %w", err)
			}

			needsSync, err := taskNeedsOutboundReconcile(current, hash)
			if err != nil {
				return fmt.Errorf("cannot compare linked task snapshot: %w", err)
			}

			if needsSync {
				if err := s.enqueueOutbound(ctx, tx, scope, link, shared.SyncActionUpdate); err != nil {
					return fmt.Errorf("cannot enqueue outbound task sync: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot save Linear task link: %w", err)
	}

	return link, previous, linked, nil
}

func (s *Service) restoreLinkedTask(
	ctx context.Context,
	scope coredata.Scoper,
	previous *coredata.Task,
	linked *coredata.Task,
	identityID *gid.GID,
) error {
	if previous == nil || linked == nil {
		return nil
	}

	now := time.Now()

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			current := &coredata.Task{}
			if err := current.LoadByIDForUpdate(ctx, tx, scope, previous.ID); err != nil {
				return fmt.Errorf("cannot load task %q: %w", previous.ID, err)
			}

			oldTask := *current

			if !current.UpdatedAt.Truncate(time.Microsecond).Equal(linked.UpdatedAt.Truncate(time.Microsecond)) {
				return nil
			}

			if err := coredata.CompletePendingOutboundJobsForTask(
				ctx,
				tx,
				scope,
				previous.ID,
				shared.CommentSyncActions(),
				now,
			); err != nil {
				return fmt.Errorf("cannot complete pending task sync jobs: %w", err)
			}

			current.Name = previous.Name
			current.Content = previous.Content
			current.State = previous.State
			current.Priority = previous.Priority
			current.Deadline = previous.Deadline
			current.AssignedToID = previous.AssignedToID
			current.Rank = previous.Rank
			current.UpdatedAt = now

			if err := current.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot restore task after failed Linear link: %w", err)
			}

			actorID, err := s.actorProfile(ctx, tx, scope, identityID, current.OrganizationID)
			if err != nil {
				return fmt.Errorf("cannot resolve task activity actor: %w", err)
			}

			if s.recordUpdateActivities != nil {
				if err := s.recordUpdateActivities(
					ctx,
					tx,
					scope,
					&oldTask,
					current,
					actorID,
					now,
				); err != nil {
					return fmt.Errorf("cannot record task update events: %w", err)
				}
			}

			internalControlIDs, err := (coredata.InternalControlTask{}).LoadInternalControlIDsByTaskID(ctx, tx, scope, current.ID)
			if err != nil {
				return fmt.Errorf("cannot load task internal controls: %w", err)
			}

			if err := webhook.InsertTaskUpdated(ctx, tx, scope, &oldTask, current, internalControlIDs, internalControlIDs); err != nil {
				return fmt.Errorf("cannot emit task updated webhook: %w", err)
			}

			return nil
		},
	)
}

func applyLinearIssueFields(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	current *coredata.Task,
	issue *linear.Issue,
	content string,
	now time.Time,
) error {
	deadline, err := shared.DateToDeadline(issue.DueDate)
	if err != nil {
		return fmt.Errorf("cannot parse Linear deadline: %w", err)
	}

	name := strings.TrimSpace(issue.Title)
	if name == "" {
		name = current.Name
	}

	state := current.State
	if issue.StateType != "" {
		state = LinearTypeToTaskState(issue.StateType)
	}

	priority := current.Priority
	if mapped, ok := LinearPriorityToTask(issue.Priority); ok {
		priority = mapped
	}

	oldState := current.State
	oldPriority := current.Priority
	current.Name = name
	current.Content = content
	current.State = state
	current.Priority = priority
	current.Deadline = deadline
	current.UpdatedAt = now

	if err := applyLinearAssignee(ctx, tx, scope, current, issue.AssigneeEmail); err != nil {
		return fmt.Errorf("cannot apply Linear assignee: %w", err)
	}

	if current.State != oldState || current.Priority != oldPriority {
		if err := current.NextRankForStatePriority(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot get next rank: %w", err)
		}
	}

	if err := current.Update(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot update task from Linear issue: %w", err)
	}

	return nil
}

func applyLinearAssignee(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	task *coredata.Task,
	email string,
) error {
	email = strings.TrimSpace(email)
	if email == "" {
		task.AssignedToID = nil

		return nil
	}

	addr, err := mail.ParseAddr(email)
	if err != nil {
		task.AssignedToID = nil

		return nil
	}

	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByOrganizationIDAndEmail(ctx, tx, scope, task.OrganizationID, addr); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			task.AssignedToID = nil

			return nil
		}

		return fmt.Errorf("cannot load assignee: %w", err)
	}

	task.AssignedToID = &profile.ID

	return nil
}

func (s *Service) finishLinearPublish(
	ctx context.Context,
	scope coredata.Scoper,
	client *linear.Client,
	task *coredata.Task,
	dbConnector *coredata.Connector,
	destination json.RawMessage,
	viewerID string,
	issue *linear.Issue,
	markdown string,
	files map[string]string,
) (*coredata.TaskExternalLink, error) {
	if err := s.persistPublishedIssueIdentity(ctx, scope, task.ID, issue, destination); err != nil {
		return nil, fmt.Errorf("cannot persist Linear issue identity: %w", err)
	}

	taskURL, err := s.taskURL(task.OrganizationID, task.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot build task URL: %w", err)
	}

	attachmentID, err := client.LinkAttachment(ctx, issue.ID, taskURL, "Probo task")
	if err != nil {
		return nil, fmt.Errorf("cannot link Linear attachment: %w", err)
	}

	metadata, err := linearLinkMetadata(attachmentID, viewerID)
	if err != nil {
		return nil, fmt.Errorf("cannot build Linear link metadata: %w", err)
	}

	if len(files) > 0 {
		metadata, err = mergeAttachments(metadata, files)
		if err != nil {
			return nil, fmt.Errorf("cannot merge Linear attachments: %w", err)
		}
	}

	hash := ContentHash(
		task.Name,
		markdown,
		task.State,
		task.Priority,
		task.Deadline,
		task.AssignedToID,
	)
	now := time.Now()
	link := &coredata.TaskExternalLink{
		OrganizationID:     task.OrganizationID,
		TaskID:             task.ID,
		ConnectorID:        dbConnector.ID,
		Provider:           coredata.ConnectorProviderLinearSync,
		ExternalID:         issue.ID,
		ExternalIdentifier: issue.Identifier,
		ExternalURL:        issue.URL,
		Destination:        destination,
		Origin:             coredata.TaskExternalLinkOriginProbo,
		RemoteUpdatedAt:    &issue.UpdatedAt,
		ContentHash:        &hash,
		Metadata:           metadata,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			current := &coredata.Task{}
			if err := current.LoadByIDForUpdate(ctx, tx, scope, task.ID); err != nil {
				return fmt.Errorf("cannot load task %q: %w", task.ID, err)
			}

			existing := &coredata.TaskExternalLink{}
			if err := existing.LoadByTaskIDForUpdate(ctx, tx, scope, task.ID); err != nil {
				return fmt.Errorf("cannot load task external link: %w", err)
			}

			if !isLinearPublishPending(existing) && existing.ExternalID != issue.ID {
				return shared.ErrTaskAlreadyLinked
			}

			link.CreatedAt = existing.CreatedAt
			if err := link.Update(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return shared.ErrTaskAlreadyLinked
				}

				return fmt.Errorf("cannot update task external link: %w", err)
			}

			needsSync, err := taskNeedsOutboundReconcile(current, hash)
			if err != nil {
				return fmt.Errorf("cannot compare published task snapshot: %w", err)
			}

			if needsSync {
				if err := s.enqueueOutbound(ctx, tx, scope, link, shared.SyncActionUpdate); err != nil {
					return fmt.Errorf("cannot enqueue outbound task sync: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot save Linear publish: %w", err)
	}

	return link, nil
}

func (s *Service) persistPublishedIssueIdentity(
	ctx context.Context,
	scope coredata.Scoper,
	taskID gid.GID,
	issue *linear.Issue,
	destination json.RawMessage,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			link := &coredata.TaskExternalLink{}
			if err := link.LoadByTaskIDForUpdate(ctx, tx, scope, taskID); err != nil {
				return fmt.Errorf("cannot load pending task external link: %w", err)
			}

			if !isLinearPublishPending(link) && link.ExternalID != issue.ID {
				return shared.ErrTaskAlreadyLinked
			}

			link.ExternalID = issue.ID
			link.ExternalIdentifier = issue.Identifier
			link.ExternalURL = issue.URL
			link.Destination = destination
			link.RemoteUpdatedAt = &issue.UpdatedAt
			link.UpdatedAt = time.Now()

			if err := link.Update(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return shared.ErrTaskAlreadyLinked
				}

				return fmt.Errorf("cannot persist Linear issue identity: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) compensateFailedPublish(
	ctx context.Context,
	client *linear.Client,
	scope coredata.Scoper,
	taskID gid.GID,
	issueID string,
) {
	if issueID != "" {
		if err := client.ArchiveIssue(ctx, issueID); err != nil && s.logger != nil {
			s.logger.WarnCtx(
				ctx,
				"cannot archive Linear issue after failed publish",
				log.String("task_id", taskID.String()),
				log.Error(err),
			)
		}
	}

	if err := s.deleteOwnedTaskLink(ctx, scope, taskID, issueID); err != nil && s.logger != nil {
		s.logger.WarnCtx(
			ctx,
			"cannot delete pending task external link after failed publish",
			log.String("task_id", taskID.String()),
			log.Error(err),
		)
	}
}

func (s *Service) deletePendingLinearLink(ctx context.Context, scope coredata.Scoper, taskID gid.GID, issueID string) {
	if err := s.deleteOwnedTaskLink(ctx, scope, taskID, issueID); err != nil && s.logger != nil {
		s.logger.WarnCtx(
			ctx,
			"cannot delete pending task external link after failed link",
			log.String("task_id", taskID.String()),
			log.Error(err),
		)
	}
}

func (s *Service) deleteOwnedTaskLink(
	ctx context.Context,
	scope coredata.Scoper,
	taskID gid.GID,
	issueID string,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			link := &coredata.TaskExternalLink{}
			if err := link.LoadByTaskIDForUpdate(ctx, tx, scope, taskID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot load task external link: %w", err)
			}

			pendingID := linearPublishPendingExternalID(taskID)
			if link.ExternalID != pendingID && (issueID == "" || link.ExternalID != issueID) {
				return nil
			}

			if err := link.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete task external link: %w", err)
			}

			if err := shared.DeleteCommentLinks(ctx, tx, scope, taskID); err != nil {
				return err
			}

			return nil
		},
	)
}

func linearLinkMetadata(attachmentID, viewerID string) (json.RawMessage, error) {
	metadata := map[string]string{
		"attachment_id": attachmentID,
		"app_actor_id":  viewerID,
	}

	encoded, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal metadata: %w", err)
	}

	return encoded, nil
}

func (s *Service) taskURL(organizationID, taskID gid.GID) (string, error) {
	u, err := url.JoinPath(
		s.baseURL,
		"organizations",
		url.PathEscape(organizationID.String()),
		"governance",
		"tasks",
		url.PathEscape(taskID.String()),
	)
	if err != nil {
		return "", fmt.Errorf("cannot build task URL: %w", err)
	}

	return u, nil
}
