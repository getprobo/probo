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
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/task/sync/linear"
	"go.probo.inc/probo/pkg/webhook"
)

type (
	// RecordUpdateActivitiesFunc writes task activity rows for an inbound
	// Linear apply. Injected from pkg/task to avoid an import cycle.
	RecordUpdateActivitiesFunc func(
		ctx context.Context,
		tx pg.Tx,
		scope coredata.Scoper,
		oldTask *coredata.Task,
		task *coredata.Task,
		actorID *gid.GID,
		now time.Time,
	) error

	Service struct {
		pg                *pg.Client
		encryptionKey     cipher.EncryptionKey
		connectorRegistry *connector.Registry
		baseURL           string
		// linearAPIBaseURL is the LINEAR_SYNC provider registration's
		// Endpoints.APIBase, threaded in by probod so a deployment that
		// repoints the Linear Sync connector moves these calls too. See
		// linear.NewClient.
		linearAPIBaseURL       string
		logger                 *log.Logger
		recordUpdateActivities RecordUpdateActivitiesFunc
	}

	LinearTeam struct {
		ID   string
		Name string
		Key  string
	}
)

const linearPublishPendingExternalIDPrefix = "pending:"

func NewService(
	pgClient *pg.Client,
	encryptionKey cipher.EncryptionKey,
	connectorRegistry *connector.Registry,
	baseURL string,
	linearAPIBaseURL string,
	logger *log.Logger,
	recordUpdateActivities RecordUpdateActivitiesFunc,
) *Service {
	return &Service{
		pg:                     pgClient,
		encryptionKey:          encryptionKey,
		connectorRegistry:      connectorRegistry,
		baseURL:                baseURL,
		linearAPIBaseURL:       linearAPIBaseURL,
		logger:                 logger,
		recordUpdateActivities: recordUpdateActivities,
	}
}

func (s *Service) GetLinkByTaskID(
	ctx context.Context,
	scope coredata.Scoper,
	taskID gid.GID,
) (*coredata.TaskExternalLink, error) {
	link := &coredata.TaskExternalLink{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := link.LoadByTaskID(ctx, conn, scope, taskID); err != nil {
				return fmt.Errorf("cannot load task external link: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get task external link: %w", err)
	}

	return link, nil
}

func (s *Service) GetLinksByTaskIDs(
	ctx context.Context,
	scope coredata.Scoper,
	taskIDs []gid.GID,
) (map[gid.GID]*coredata.TaskExternalLink, error) {
	links := coredata.TaskExternalLinks{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := links.LoadByTaskIDs(ctx, conn, scope, taskIDs); err != nil {
				return fmt.Errorf("cannot load task external links: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get task external links: %w", err)
	}

	byTaskID := make(map[gid.GID]*coredata.TaskExternalLink, len(links))
	for _, link := range links {
		byTaskID[link.TaskID] = link
	}

	return byTaskID, nil
}

func (s *Service) ListLinearTeams(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) ([]LinearTeam, error) {
	var (
		teams []LinearTeam
		after *string
	)

	for range linearListMaxPages {
		page, err := s.SearchLinearTeams(ctx, scope, organizationID, "", linearPickerPageSizeMax, after)
		if err != nil {
			return nil, fmt.Errorf("cannot list Linear teams: %w", err)
		}

		teams = append(teams, page.Teams...)
		if !page.HasNextPage || page.EndCursor == "" {
			return teams, nil
		}

		next := page.EndCursor
		after = &next
	}

	return nil, fmt.Errorf("cannot list Linear teams: pagination limit reached")
}

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

	markdown, err := ContentToMarkdown(task.Content)
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

	var issue *linear.Issue

	if reused != nil {
		issue = issueFromExternalLink(reused)
	} else {
		issue, err = client.CreateIssue(
			ctx,
			linear.IssueInput{
				TeamID:      teamID,
				Title:       task.Name,
				Description: markdown,
				StateID:     stateID,
				Priority:    TaskPriorityToLinear(task.Priority),
				DueDate:     DeadlineToLinearDate(task.Deadline),
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
		markdown,
	)
	if err != nil {
		s.compensateFailedPublish(ctx, client, scope, taskID, issue.ID)
		return nil, fmt.Errorf("cannot finish Linear publish: %w", err)
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

	content, markdown, err := linearIssueContent(issue)
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

	link, err := s.finishLinearLink(
		ctx,
		scope,
		task,
		account.connector,
		destination,
		viewerID,
		issue,
		content,
		markdown,
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

		if !errors.Is(err, ErrTaskAlreadyLinked) {
			s.deletePendingLinearLink(ctx, scope, taskID, issue.ID)
		}

		return nil, fmt.Errorf("cannot finish Linear link: %w", err)
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
		Provider:       coredata.ConnectorProviderLinear,
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
						return ErrTaskAlreadyLinked
					}

					return fmt.Errorf("cannot insert pending task external link: %w", err)
				}

				return nil
			}

			if err != nil {
				return fmt.Errorf("cannot load task external link: %w", err)
			}

			if isLinearPublishComplete(existing) {
				return ErrTaskAlreadyLinked
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

			if err := pending.Insert(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return ErrTaskAlreadyLinked
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

func linearIssueContent(issue *linear.Issue) (string, string, error) {
	content, err := MarkdownToContent(issue.Description)
	if err != nil {
		return "", "", err
	}

	markdown, err := ContentToMarkdown(content)
	if err != nil {
		return "", "", err
	}

	return content, markdown, nil
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
	attachmentID string,
	identityID *gid.GID,
) (*coredata.TaskExternalLink, error) {
	metadata, err := json.Marshal(
		map[string]string{
			"attachment_id": attachmentID,
			"app_actor_id":  viewerID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal metadata: %w", err)
	}

	now := time.Now()
	link := &coredata.TaskExternalLink{
		OrganizationID:     task.OrganizationID,
		TaskID:             task.ID,
		ConnectorID:        dbConnector.ID,
		Provider:           coredata.ConnectorProviderLinear,
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
				return ErrTaskAlreadyLinked
			}

			oldTask := *current
			if err := applyLinearIssueFields(ctx, tx, scope, current, issue, content, now); err != nil {
				return fmt.Errorf("cannot apply Linear issue to task: %w", err)
			}

			actorID, err := linkActivityActorID(ctx, tx, scope, identityID, current.OrganizationID)
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

			if err := webhook.InsertTaskUpdated(ctx, tx, scope, &oldTask, current); err != nil {
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
					return ErrTaskAlreadyLinked
				}

				return fmt.Errorf("cannot update task external link: %w", err)
			}

			needsSync, err := taskNeedsOutboundReconcile(current, hash)
			if err != nil {
				return fmt.Errorf("cannot compare linked task snapshot: %w", err)
			}

			if needsSync {
				if err := s.enqueueOutboundTx(ctx, tx, scope, link, SyncActionUpdate); err != nil {
					return fmt.Errorf("cannot enqueue outbound task sync: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return link, nil
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
	current.Deadline = LinearDateToDeadline(issue.DueDate)
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

func linkActivityActorID(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	identityID *gid.GID,
	organizationID gid.GID,
) (*gid.GID, error) {
	if identityID == nil {
		return nil, nil
	}

	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByIdentityIDAndOrganizationID(
		ctx,
		tx,
		scope,
		*identityID,
		organizationID,
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot load actor profile: %w", err)
	}

	return &profile.ID, nil
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

	metadata, err := json.Marshal(
		map[string]string{
			"attachment_id": attachmentID,
			"app_actor_id":  viewerID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal metadata: %w", err)
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
		Provider:           coredata.ConnectorProviderLinear,
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
				return ErrTaskAlreadyLinked
			}

			link.CreatedAt = existing.CreatedAt
			if err := link.Update(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return ErrTaskAlreadyLinked
				}

				return fmt.Errorf("cannot update task external link: %w", err)
			}

			needsSync, err := taskNeedsOutboundReconcile(current, hash)
			if err != nil {
				return fmt.Errorf("cannot compare published task snapshot: %w", err)
			}

			if needsSync {
				if err := s.enqueueOutboundTx(ctx, tx, scope, link, SyncActionUpdate); err != nil {
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
				return ErrTaskAlreadyLinked
			}

			link.ExternalID = issue.ID
			link.ExternalIdentifier = issue.Identifier
			link.ExternalURL = issue.URL
			link.Destination = destination
			link.RemoteUpdatedAt = &issue.UpdatedAt
			link.UpdatedAt = time.Now()

			if err := link.Update(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return ErrTaskAlreadyLinked
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

			return nil
		},
	)
}

func (s *Service) Unlink(
	ctx context.Context,
	scope coredata.Scoper,
	taskID gid.GID,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			link := &coredata.TaskExternalLink{}
			if err := link.LoadByTaskID(ctx, tx, scope, taskID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrTaskNotLinked
				}

				return fmt.Errorf("cannot load task external link: %w", err)
			}

			if err := link.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete task external link: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) EnqueueOutbound(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	taskID gid.GID,
	action SyncAction,
) error {
	link := &coredata.TaskExternalLink{}
	if err := link.LoadByTaskID(ctx, tx, scope, taskID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil
		}

		return fmt.Errorf("cannot load task external link: %w", err)
	}

	if err := s.enqueueOutboundTx(ctx, tx, scope, link, action); err != nil {
		return fmt.Errorf("cannot enqueue outbound task sync: %w", err)
	}

	return nil
}

func (s *Service) enqueueOutboundTx(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	link *coredata.TaskExternalLink,
	action SyncAction,
) error {
	teamID, err := link.DestinationTeamID()
	if err != nil {
		return fmt.Errorf("cannot read Linear destination team: %w", err)
	}

	payload, err := json.Marshal(
		JobPayload{
			Action:             action,
			TaskID:             link.TaskID,
			ExternalID:         link.ExternalID,
			ExternalIdentifier: link.ExternalIdentifier,
			TeamID:             teamID,
			ConnectorID:        link.ConnectorID,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot marshal task sync payload: %w", err)
	}

	now := time.Now()
	if err := coredata.CompletePendingOutboundJobsForTask(
		ctx,
		tx,
		scope,
		link.TaskID,
		now,
	); err != nil {
		return fmt.Errorf("cannot complete pending outbound task sync jobs: %w", err)
	}

	job := &coredata.TaskSyncJob{
		ID:             gid.New(scope.GetTenantID(), coredata.TaskSyncJobEntityType),
		OrganizationID: link.OrganizationID,
		Direction:      coredata.TaskSyncJobDirectionOutbound,
		Status:         coredata.TaskSyncJobStatusPending,
		Payload:        payload,
		CreatedAt:      now,
		UpdatedAt:      now,
		AttemptCount:   0,
	}

	if err := job.Insert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot insert task sync job: %w", err)
	}

	return nil
}

func (s *Service) linearClientForConnector(
	ctx context.Context,
	scope coredata.Scoper,
	dbConnector *coredata.Connector,
) (*linear.Client, *coredata.Connector, error) {
	oauthConn, ok := dbConnector.Connection.(*connector.OAuth2Connection)
	if !ok {
		return nil, nil, fmt.Errorf("cannot use Linear connector: unsupported credential")
	}

	if missing := missingTaskSyncScopes(oauthConn.Scopes()); len(missing) > 0 {
		return nil, nil, NewLinearReconnectRequiredError(missing)
	}

	if err := s.connectorRegistry.ConfigureConnection(string(dbConnector.Provider), oauthConn); err != nil {
		return nil, nil, fmt.Errorf("cannot configure Linear connector: %w", err)
	}

	tokenBefore := oauthConn.AccessToken

	httpClient, err := s.oauthHTTPClient(ctx, oauthConn, dbConnector.Provider)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create Linear HTTP client: %w", err)
	}

	if oauthConn.AccessToken != tokenBefore {
		dbConnector.UpdatedAt = time.Now()

		err := s.pg.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				if err := dbConnector.Update(ctx, tx, scope, s.encryptionKey); err != nil {
					return fmt.Errorf("cannot persist refreshed Linear token: %w", err)
				}

				return nil
			},
		)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot persist refreshed Linear token: %w", err)
		}
	}

	return linear.NewClient(httpClient, s.linearAPIBaseURL), dbConnector, nil
}

func (s *Service) oauthHTTPClient(
	ctx context.Context,
	conn *connector.OAuth2Connection,
	provider coredata.ConnectorProvider,
) (*http.Client, error) {
	refreshCfg := s.connectorRegistry.GetOAuth2RefreshConfig(string(provider))
	if refreshCfg != nil {
		return conn.RefreshableClient(ctx, *refreshCfg)
	}

	return conn.Client(ctx)
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
