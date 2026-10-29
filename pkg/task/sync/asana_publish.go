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
	"net/url"
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/task/sync/asana"
)

type AsanaPage struct {
	Items       []asana.Named
	Tasks       []asana.Task
	EndCursor   string
	HasNextPage bool
}

func (s *Service) ListAsanaTeams(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	query string,
	limit int,
	offset string,
) (AsanaPage, error) {
	accounts, err := s.asanaAccounts(ctx, scope, organizationID)
	if err != nil {
		return AsanaPage{}, err
	}

	var page AsanaPage

	for _, account := range accounts {
		listed, err := account.client.ListTeams(ctx, account.settings.WorkspaceGID, query, limit, offset)
		if err != nil {
			return AsanaPage{}, fmt.Errorf("cannot list Asana teams: %w", err)
		}

		page.Items = append(page.Items, listed.Items...)
		if listed.HasNextPage {
			page.HasNextPage = true
			page.EndCursor = listed.EndCursor
		}
	}

	return page, nil
}

func (s *Service) ListAsanaProjects(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	teamGID string,
	query string,
	limit int,
	offset string,
) (AsanaPage, error) {
	accounts, err := s.asanaAccounts(ctx, scope, organizationID)
	if err != nil {
		return AsanaPage{}, err
	}

	var page AsanaPage

	for _, account := range accounts {
		listed, err := account.client.ListProjects(
			ctx,
			account.settings.WorkspaceGID,
			teamGID,
			query,
			limit,
			offset,
		)
		if err != nil {
			return AsanaPage{}, fmt.Errorf("cannot list Asana projects: %w", err)
		}

		page.Items = append(page.Items, listed.Items...)
		if listed.HasNextPage {
			page.HasNextPage = true
			page.EndCursor = listed.EndCursor
		}
	}

	return page, nil
}

func (s *Service) ListAsanaTasks(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	projectGID string,
	query string,
	limit int,
	offset string,
) (AsanaPage, error) {
	if strings.TrimSpace(projectGID) == "" {
		return AsanaPage{}, ErrAsanaProjectIDRequired
	}

	accounts, err := s.asanaAccounts(ctx, scope, organizationID)
	if err != nil {
		return AsanaPage{}, err
	}

	account, _, err := s.asanaAccountForProject(ctx, accounts, projectGID)
	if err != nil {
		return AsanaPage{}, err
	}

	listed, err := account.client.ListTasks(
		ctx,
		account.settings.WorkspaceGID,
		projectGID,
		query,
		limit,
		offset,
	)
	if err != nil {
		return AsanaPage{}, fmt.Errorf("cannot list Asana tasks: %w", err)
	}

	return AsanaPage{
		Tasks:       listed.Tasks,
		EndCursor:   listed.EndCursor,
		HasNextPage: listed.HasNextPage,
	}, nil
}

func (s *Service) ProjectForNewTask(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	explicitSet bool,
	explicitProjectID *string,
) (string, bool, error) {
	if explicitSet {
		projectID, publish := LinearTeamForCreate(explicitSet, explicitProjectID, "")

		return projectID, publish, nil
	}

	settings, err := s.AsanaSettings(ctx, scope, organizationID)
	if err != nil {
		return "", false, err
	}

	defaultProjectID := ""
	if settings.DefaultProject != nil {
		defaultProjectID = settings.DefaultProject.GID
	}

	projectID, publish := LinearTeamForCreate(explicitSet, explicitProjectID, defaultProjectID)

	return projectID, publish, nil
}

func (s *Service) PublishToAsana(
	ctx context.Context,
	scope coredata.Scoper,
	taskID gid.GID,
	projectGID string,
) (*coredata.TaskExternalLink, error) {
	projectGID = strings.TrimSpace(projectGID)
	if projectGID == "" {
		return nil, ErrAsanaProjectIDRequired
	}

	task, err := s.loadTask(ctx, scope, taskID)
	if err != nil {
		return nil, err
	}

	if err := s.requireUnlinked(ctx, scope, taskID); err != nil {
		return nil, err
	}

	accounts, err := s.asanaAccounts(ctx, scope, task.OrganizationID)
	if err != nil {
		return nil, err
	}

	account, _, err := s.asanaAccountForProject(ctx, accounts, projectGID)
	if err != nil {
		return nil, err
	}

	created, err := s.createAsanaTask(ctx, scope, account, task, projectGID)
	if err != nil {
		return nil, err
	}

	link, err := s.finishAsanaLink(
		ctx,
		scope,
		account,
		task,
		created,
		projectGID,
		coredata.TaskExternalLinkOriginProbo,
		false,
	)
	if err != nil {
		if deleteErr := account.client.DeleteTask(ctx, created.GID); deleteErr != nil && s.logger != nil {
			s.logger.WarnCtx(ctx, "cannot delete Asana task after failed publish", log.Error(deleteErr))
		}

		return nil, err
	}

	s.ensureAsanaWebhook(ctx, scope, account.connector.ID, projectGID)

	return link, nil
}

func (s *Service) LinkToAsana(
	ctx context.Context,
	scope coredata.Scoper,
	taskID gid.GID,
	projectGID string,
	remoteTaskGID string,
) (*coredata.TaskExternalLink, error) {
	projectGID = strings.TrimSpace(projectGID)
	remoteTaskGID = strings.TrimSpace(remoteTaskGID)

	if projectGID == "" {
		return nil, ErrAsanaProjectIDRequired
	}

	if remoteTaskGID == "" {
		return nil, ErrAsanaTaskIDRequired
	}

	task, err := s.loadTask(ctx, scope, taskID)
	if err != nil {
		return nil, err
	}

	if err := s.requireUnlinked(ctx, scope, taskID); err != nil {
		return nil, err
	}

	accounts, err := s.asanaAccounts(ctx, scope, task.OrganizationID)
	if err != nil {
		return nil, err
	}

	account, _, err := s.asanaAccountForProject(ctx, accounts, projectGID)
	if err != nil {
		return nil, err
	}

	remote, err := account.client.GetTask(ctx, remoteTaskGID)
	if err != nil {
		if asana.IsNotFound(err) {
			return nil, ErrAsanaTaskNotFound
		}

		return nil, fmt.Errorf("cannot load Asana task: %w", err)
	}

	if sectionForProject(remote, projectGID) == "" && !taskInProject(remote, projectGID) {
		return nil, ErrAsanaTaskNotFound
	}

	link, err := s.finishAsanaLink(
		ctx,
		scope,
		account,
		task,
		remote,
		projectGID,
		coredata.TaskExternalLinkOriginProbo,
		true,
	)
	if err != nil {
		return nil, err
	}

	s.ensureAsanaWebhook(ctx, scope, account.connector.ID, projectGID)

	return link, nil
}

func taskInProject(task asana.Task, projectGID string) bool {
	for _, membership := range task.Memberships {
		if membership.ProjectGID == projectGID {
			return true
		}
	}

	return false
}

func (s *Service) loadTask(ctx context.Context, scope coredata.Scoper, taskID gid.GID) (*coredata.Task, error) {
	task := &coredata.Task{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := task.LoadByID(ctx, conn, scope, taskID); err != nil {
				return fmt.Errorf("cannot load task %q: %w", taskID, err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *Service) requireUnlinked(ctx context.Context, scope coredata.Scoper, taskID gid.GID) error {
	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			link := &coredata.TaskExternalLink{}
			if err := link.LoadByTaskID(ctx, conn, scope, taskID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return err
			}

			return ErrTaskAlreadyLinked
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) createAsanaTask(
	ctx context.Context,
	scope coredata.Scoper,
	account asanaAccount,
	task *coredata.Task,
	projectGID string,
) (asana.Task, error) {
	htmlNotes, err := s.asanaHTMLNotes(ctx, scope, account, task.OrganizationID, task.Content)
	if err != nil {
		return asana.Task{}, err
	}

	mapping := projectMapping(account.settings, projectGID)
	input := asana.TaskCreate{
		Name:        task.Name,
		HTMLNotes:   htmlNotes,
		ProjectGID:  projectGID,
		Completed:   CompletedForState(task.State),
		DueOn:       DeadlineToLinearDate(task.Deadline),
		AssigneeGID: s.asanaAssigneeGID(ctx, scope, account, task.AssignedToID),
	}

	if option, ok := OutboundPriorityOption(task.Priority, mapping); ok {
		input.CustomFields = map[string]string{mapping.PriorityFieldGID: option}
	}

	created, err := account.client.CreateTask(ctx, input)
	if err != nil {
		return asana.Task{}, fmt.Errorf("cannot create Asana task: %w", err)
	}

	if section := OutboundSection(task.State, mapping); section != "" {
		if err := account.client.AddTaskToSection(ctx, section, created.GID); err != nil {
			_ = account.client.DeleteTask(ctx, created.GID)

			return asana.Task{}, fmt.Errorf("cannot move Asana task to section: %w", err)
		}
	}

	return created, nil
}

func (s *Service) finishAsanaLink(
	ctx context.Context,
	scope coredata.Scoper,
	account asanaAccount,
	task *coredata.Task,
	remote asana.Task,
	projectGID string,
	origin coredata.TaskExternalLinkOrigin,
	applyRemote bool,
) (*coredata.TaskExternalLink, error) {
	mapping := projectMapping(account.settings, projectGID)
	sectionGID := sectionForProject(remote, projectGID)

	optionGID := ""
	if mapping != nil {
		optionGID = enumOption(remote, mapping.PriorityFieldGID)
	}

	hash := asanaSnapshotHash(remote, sectionGID, optionGID)
	now := time.Now()

	content := task.Content

	var err error

	if applyRemote {
		content, err = s.contentFromAsanaHTML(ctx, scope, account, task.OrganizationID, remote.HTMLNotes)
		if err != nil {
			return nil, err
		}
	}

	destination, err := json.Marshal(coredata.TaskExternalLinkDestination{TeamID: projectGID, ProjectID: &projectGID})
	if err != nil {
		return nil, fmt.Errorf("cannot marshal Asana destination: %w", err)
	}

	taskURL := remote.Permalink
	if taskURL == "" {
		taskURL, err = asanaTaskURL(s.asanaAPIBaseURL, projectGID, remote.GID)
		if err != nil {
			return nil, err
		}
	}

	link := &coredata.TaskExternalLink{
		OrganizationID:     task.OrganizationID,
		TaskID:             task.ID,
		ConnectorID:        account.connector.ID,
		Provider:           coredata.ConnectorProviderAsana,
		ExternalID:         remote.GID,
		ExternalIdentifier: remote.GID,
		ExternalURL:        taskURL,
		Destination:        destination,
		Origin:             origin,
		ContentHash:        &hash,
		Metadata:           json.RawMessage(`{}`),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if !remote.ModifiedAt.IsZero() {
		link.RemoteUpdatedAt = &remote.ModifiedAt
	}

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if applyRemote {
				if err := s.applyRemoteTask(ctx, tx, scope, task, remote, content, projectGID, mapping); err != nil {
					return err
				}
			}

			if err := link.Insert(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return ErrTaskAlreadyLinked
				}

				return fmt.Errorf("cannot insert Asana task link: %w", err)
			}

			if err := s.enqueueOutboundTx(ctx, tx, scope, link, SyncActionUpdate); err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return link, nil
}

func (s *Service) applyRemoteTask(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	task *coredata.Task,
	remote asana.Task,
	content string,
	projectGID string,
	mapping *AsanaProjectMapping,
) error {
	old := *task

	task.Name = remote.Name
	task.Content = content
	task.State = InboundState(sectionForProject(remote, projectGID), remote.Completed, task.State, mapping)
	task.Deadline = LinearDateToDeadline(firstNonEmpty(remote.DueAt, remote.DueOn))

	if mapping != nil {
		if priority, ok := InboundPriority(enumOption(remote, mapping.PriorityFieldGID), task.Priority, mapping); ok {
			task.Priority = priority
		}
	}

	if _, err := s.applyAsanaAssignee(ctx, tx, scope, task, remote.AssigneeEmail); err != nil {
		return err
	}

	if task.State != old.State || task.Priority != old.Priority {
		if err := task.NextRankForStatePriority(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot rank imported Asana task: %w", err)
		}
	}

	task.UpdatedAt = time.Now()
	if err := task.Update(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot update task from Asana: %w", err)
	}

	if s.recordUpdateActivities != nil {
		if err := s.recordUpdateActivities(ctx, tx, scope, &old, task, nil, task.UpdatedAt); err != nil {
			return fmt.Errorf("cannot record Asana task activity: %w", err)
		}
	}

	if err := s.notifySyncedMentions(
		ctx,
		tx,
		scope,
		task.OrganizationID,
		task.ID,
		task.Name,
		old.Content,
		task.Content,
		nil,
	); err != nil {
		return fmt.Errorf("cannot notify synced mentions: %w", err)
	}

	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}

	return ""
}

func asanaOrigin(apiBaseURL string) (string, error) {
	base, err := url.Parse(apiBaseURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		if err == nil {
			err = fmt.Errorf("missing scheme or host")
		}

		return "", fmt.Errorf("cannot parse Asana API base url: %w", err)
	}

	return (&url.URL{Scheme: base.Scheme, Host: base.Host}).String(), nil
}

func asanaTaskURL(apiBaseURL, projectGID, taskGID string) (string, error) {
	origin, err := asanaOrigin(apiBaseURL)
	if err != nil {
		return "", err
	}

	taskURL, err := url.JoinPath(
		origin,
		"0",
		url.PathEscape(projectGID),
		url.PathEscape(taskGID),
	)
	if err != nil {
		return "", fmt.Errorf("cannot build Asana task url: %w", err)
	}

	return taskURL, nil
}
