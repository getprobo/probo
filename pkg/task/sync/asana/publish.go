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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	client "go.probo.inc/probo/pkg/task/sync/asana/client"
	"go.probo.inc/probo/pkg/task/sync/shared"
)

const (
	asanaAccountCursorPrefix = "asana-accounts:"

	// asanaNamedPageLimit matches the Asana client's default page size.
	asanaNamedPageLimit = 50
)

type (
	asanaAccountCursor struct {
		Accounts []asanaAccountCursorEntry `json:"accounts"`
	}

	asanaAccountCursorEntry struct {
		ConnectorID string `json:"c"`
		Offset      string `json:"o"`
	}
)

type AsanaPage struct {
	Items       []client.Named
	Tasks       []client.Task
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
		return AsanaPage{}, fmt.Errorf("cannot load Asana accounts: %w", err)
	}

	page, err := s.listAsanaNamed(
		ctx,
		accounts,
		limit,
		offset,
		func(ctx context.Context, account asanaAccount, limit int, offset string) (client.Page, error) {
			return account.client.ListTeams(ctx, account.settings.WorkspaceGID, query, limit, offset)
		},
	)
	if err != nil {
		return AsanaPage{}, fmt.Errorf("cannot list Asana teams: %w", err)
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
		return AsanaPage{}, fmt.Errorf("cannot load Asana accounts: %w", err)
	}

	page, err := s.listAsanaNamed(
		ctx,
		accounts,
		limit,
		offset,
		func(ctx context.Context, account asanaAccount, limit int, offset string) (client.Page, error) {
			return account.client.ListProjects(
				ctx,
				account.settings.WorkspaceGID,
				teamGID,
				query,
				limit,
				offset,
			)
		},
	)
	if err != nil {
		return AsanaPage{}, fmt.Errorf("cannot list Asana projects: %w", err)
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
	projectGID = strings.TrimSpace(projectGID)
	if projectGID == "" {
		return AsanaPage{}, ErrAsanaProjectIDRequired
	}

	accounts, err := s.asanaAccounts(ctx, scope, organizationID)
	if err != nil {
		return AsanaPage{}, fmt.Errorf("cannot load Asana accounts: %w", err)
	}

	account, _, err := s.asanaAccountForProject(ctx, accounts, projectGID)
	if err != nil {
		return AsanaPage{}, fmt.Errorf("cannot resolve Asana project: %w", err)
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

func (s *Service) listAsanaNamed(
	ctx context.Context,
	accounts []asanaAccount,
	limit int,
	offset string,
	list func(context.Context, asanaAccount, int, string) (client.Page, error),
) (AsanaPage, error) {
	ids := make([]string, 0, len(accounts))
	for i := range accounts {
		ids = append(ids, accounts[i].connector.ID.String())
	}

	offsets, err := asanaPageOffsets(ids, offset)
	if err != nil {
		return AsanaPage{}, fmt.Errorf("cannot read Asana page cursor: %w", err)
	}

	var (
		page          AsanaPage
		next          = map[string]string{}
		singleEnd     string
		singleHasNext bool
		remaining     = asanaPageBudget(limit)
	)

	for i := range accounts {
		id := accounts[i].connector.ID.String()

		accountOffset, ok := offsets[id]
		if !ok {
			continue
		}

		if len(accounts) > 1 && remaining <= 0 {
			next[id] = accountOffset

			continue
		}

		requestLimit := limit
		if len(accounts) > 1 {
			requestLimit = remaining
		}

		listed, err := list(ctx, accounts[i], requestLimit, accountOffset)
		if err != nil {
			return AsanaPage{}, fmt.Errorf("cannot list Asana page: %w", err)
		}

		page.Items = append(page.Items, listed.Items...)

		if len(accounts) <= 1 {
			if listed.HasNextPage && listed.EndCursor != "" {
				singleEnd = listed.EndCursor
				singleHasNext = true
			}

			continue
		}

		remaining -= len(listed.Items)
		if remaining < 0 {
			remaining = 0
		}

		if listed.HasNextPage && listed.EndCursor != "" {
			next[id] = listed.EndCursor
		}
	}

	if len(accounts) <= 1 {
		page.EndCursor = singleEnd
		page.HasNextPage = singleHasNext

		return page, nil
	}

	if len(next) == 0 {
		return page, nil
	}

	encoded, err := encodeAsanaAccountCursors(next)
	if err != nil {
		return AsanaPage{}, fmt.Errorf("cannot encode Asana page cursor: %w", err)
	}

	page.EndCursor = encoded
	page.HasNextPage = true

	return page, nil
}

// A non-positive limit uses the Asana client's default page size.
func asanaPageBudget(limit int) int {
	if limit <= 0 {
		return asanaNamedPageLimit
	}

	return limit
}

func asanaPageOffsets(accountIDs []string, offset string) (map[string]string, error) {
	decoded, composite, err := decodeAsanaAccountCursors(offset)
	if err != nil {
		return nil, fmt.Errorf("cannot decode Asana page cursor: %w", err)
	}

	if len(accountIDs) == 1 {
		id := accountIDs[0]
		if composite {
			next, present := decoded[id]
			if !present {
				return map[string]string{}, nil
			}

			return map[string]string{id: next}, nil
		}

		return map[string]string{id: offset}, nil
	}

	if strings.TrimSpace(offset) == "" {
		offsets := make(map[string]string, len(accountIDs))
		for _, id := range accountIDs {
			offsets[id] = ""
		}

		return offsets, nil
	}

	if !composite {
		return nil, fmt.Errorf("cannot use Asana page cursor")
	}

	offsets := make(map[string]string, len(accountIDs))
	for _, id := range accountIDs {
		if next, present := decoded[id]; present {
			offsets[id] = next
		}
	}

	return offsets, nil
}

func encodeAsanaAccountCursors(offsets map[string]string) (string, error) {
	ids := make([]string, 0, len(offsets))
	for id := range offsets {
		if id == "" {
			continue
		}

		ids = append(ids, id)
	}

	slices.Sort(ids)

	entries := make([]asanaAccountCursorEntry, 0, len(ids))
	for _, id := range ids {
		entries = append(entries, asanaAccountCursorEntry{ConnectorID: id, Offset: offsets[id]})
	}

	raw, err := json.Marshal(asanaAccountCursor{Accounts: entries})
	if err != nil {
		return "", fmt.Errorf("cannot marshal Asana page cursor: %w", err)
	}

	return asanaAccountCursorPrefix + base64.RawURLEncoding.EncodeToString(raw), nil
}

func decodeAsanaAccountCursors(offset string) (map[string]string, bool, error) {
	if !strings.HasPrefix(offset, asanaAccountCursorPrefix) {
		return nil, false, nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(offset, asanaAccountCursorPrefix))
	if err != nil {
		return nil, false, fmt.Errorf("cannot decode Asana page cursor encoding: %w", err)
	}

	var decoded asanaAccountCursor
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, false, fmt.Errorf("cannot unmarshal Asana page cursor: %w", err)
	}

	offsets := make(map[string]string, len(decoded.Accounts))
	for _, account := range decoded.Accounts {
		if account.ConnectorID == "" {
			continue
		}

		offsets[account.ConnectorID] = account.Offset
	}

	return offsets, true, nil
}

func (s *Service) ProjectForNewTask(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	explicitSet bool,
	explicitProjectID *string,
) (string, bool, error) {
	if explicitSet {
		projectID, publish := projectChoice(explicitSet, explicitProjectID, "")

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

	projectID, publish := projectChoice(explicitSet, explicitProjectID, defaultProjectID)

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
		return nil, fmt.Errorf("cannot load task: %w", err)
	}

	if err := s.requireUnlinked(ctx, scope, taskID); err != nil {
		return nil, fmt.Errorf("cannot check task link: %w", err)
	}

	accounts, err := s.asanaAccounts(ctx, scope, task.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("cannot load Asana accounts: %w", err)
	}

	account, _, err := s.asanaAccountForProject(ctx, accounts, projectGID)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve Asana project: %w", err)
	}

	created, err := s.createAsanaTask(ctx, scope, account, task, projectGID)
	if err != nil {
		return nil, fmt.Errorf("cannot create Asana task: %w", err)
	}

	remote, err := account.client.GetTask(ctx, created.GID)
	if err != nil {
		s.deletePublishedAsanaTask(ctx, account, created.GID)

		return nil, fmt.Errorf("cannot reload Asana task: %w", err)
	}

	link, err := s.finishAsanaLink(
		ctx,
		scope,
		account,
		task,
		remote,
		projectGID,
		coredata.TaskExternalLinkOriginProbo,
		false,
	)
	if err != nil {
		s.deletePublishedAsanaTask(ctx, account, created.GID)

		return nil, fmt.Errorf("cannot finish Asana link: %w", err)
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
		return nil, fmt.Errorf("cannot load task: %w", err)
	}

	if err := s.requireUnlinked(ctx, scope, taskID); err != nil {
		return nil, fmt.Errorf("cannot check task link: %w", err)
	}

	accounts, err := s.asanaAccounts(ctx, scope, task.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("cannot load Asana accounts: %w", err)
	}

	account, _, err := s.asanaAccountForProject(ctx, accounts, projectGID)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve Asana project: %w", err)
	}

	remote, err := account.client.GetTask(ctx, remoteTaskGID)
	if err != nil {
		if client.IsNotFound(err) {
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
		return nil, fmt.Errorf("cannot finish Asana link: %w", err)
	}

	s.ensureAsanaWebhook(ctx, scope, account.connector.ID, projectGID)

	return link, nil
}

func taskInProject(task client.Task, projectGID string) bool {
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

				return fmt.Errorf("cannot load task external link: %w", err)
			}

			return shared.ErrTaskAlreadyLinked
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
) (client.Task, error) {
	htmlNotes, err := s.asanaHTMLNotes(ctx, scope, account, task.OrganizationID, task.Content)
	if err != nil {
		return client.Task{}, fmt.Errorf("cannot render Asana description: %w", err)
	}

	mapping := projectMapping(account.settings, projectGID)

	input := client.TaskCreate{
		Name:       task.Name,
		HTMLNotes:  htmlNotes,
		ProjectGID: projectGID,
		Completed:  CompletedForState(task.State),
		DueOn:      shared.DeadlineToDate(task.Deadline),
	}

	assigneeGID, assigneeOK := s.asanaAssigneeGID(ctx, scope, account, task.AssignedToID)
	if assigneeOK {
		input.AssigneeGID = assigneeGID
	}

	if option, ok := OutboundPriorityOption(task.Priority, mapping); ok {
		input.CustomFields = map[string]string{mapping.PriorityFieldGID: option}
	}

	created, err := account.client.CreateTask(ctx, input)
	if err != nil {
		return client.Task{}, fmt.Errorf("cannot create Asana task: %w", err)
	}

	if section := OutboundSection(task.State, mapping); section != "" {
		if err := account.client.AddTaskToSection(ctx, section, created.GID); err != nil {
			s.deletePublishedAsanaTask(ctx, account, created.GID)

			return client.Task{}, fmt.Errorf("cannot move Asana task to section: %w", err)
		}
	}

	return created, nil
}

func (s *Service) deletePublishedAsanaTask(ctx context.Context, account asanaAccount, taskGID string) {
	if deleteErr := account.client.DeleteTask(context.WithoutCancel(ctx), taskGID); deleteErr != nil && s.logger != nil {
		s.logger.WarnCtx(ctx, "cannot delete Asana task after failed publish", log.Error(deleteErr))
	}
}

func (s *Service) finishAsanaLink(
	ctx context.Context,
	scope coredata.Scoper,
	account asanaAccount,
	task *coredata.Task,
	remote client.Task,
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
			return nil, fmt.Errorf("cannot convert Asana description: %w", err)
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
			return nil, fmt.Errorf("cannot build Asana task URL: %w", err)
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
				if err := s.applyRemoteTask(ctx, tx, scope, task, remote, content, projectGID, mapping, true); err != nil {
					return fmt.Errorf("cannot apply Asana task: %w", err)
				}
			}

			if err := link.Insert(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return shared.ErrTaskAlreadyLinked
				}

				return fmt.Errorf("cannot insert Asana task link: %w", err)
			}

			if err := s.enqueueOutbound(ctx, tx, scope, link, shared.SyncActionUpdate); err != nil {
				return fmt.Errorf("cannot enqueue outbound task sync: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot save Asana task link: %w", err)
	}

	return link, nil
}

func (s *Service) applyRemoteTask(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	task *coredata.Task,
	remote client.Task,
	content string,
	projectGID string,
	mapping *AsanaProjectMapping,
	replace bool,
) error {
	deadline, err := shared.DateToDeadline(firstNonEmpty(remote.DueAt, remote.DueOn))
	if err != nil {
		return fmt.Errorf("cannot parse Asana deadline: %w", err)
	}

	old := *task

	sectionGID := sectionForProject(remote, projectGID)
	task.Name = remote.Name
	task.Content = content
	task.Deadline = deadline

	// A link replaces the Probo task. A later sync keeps an unmapped local state.
	if replace {
		task.State = inboundStateFromSection(sectionGID, remote.Completed, task.State, mapping)
	} else {
		task.State = InboundState(sectionGID, remote.Completed, task.State, mapping)
	}

	if mapping != nil {
		optionGID := enumOption(remote, mapping.PriorityFieldGID)

		var (
			priority coredata.TaskPriority
			ok       bool
		)

		if replace {
			priority, ok = inboundPriorityFromOption(optionGID, task.Priority, mapping)
		} else {
			priority, ok = InboundPriority(optionGID, task.Priority, mapping)
		}

		if ok {
			task.Priority = priority
		}
	}

	if _, err := s.applyAsanaAssignee(ctx, tx, scope, task, remote.AssigneeEmail); err != nil {
		return fmt.Errorf("cannot apply Asana assignee: %w", err)
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

	if err := shared.NotifyMentionedProfiles(
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
		return "", fmt.Errorf("cannot resolve Asana origin: %w", err)
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
